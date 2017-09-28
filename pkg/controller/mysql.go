package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	"github.com/k8sdb/mysql/pkg/validator"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (c *Controller) create(mysql *tapi.MySQL) error {
	_, err := kutildb.TryPatchMySQL(c.ExtClient, mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = tapi.DatabasePhaseCreating
		return in
	})
	if err != nil {
		c.recorder.Eventf(mysql.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidateMySQL(c.Client, mysql); err != nil {
		c.recorder.Event(mysql.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		mysql.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate MySQL",
	)
	// Check DormantDatabase
	matched, err := c.matchDormantDatabase(mysql)
	if err != nil {
		return err
	}
	if matched {
		//TODO: Use Annotation Key
		mysql.Annotations = map[string]string{
			"kubedb.com/ignore": "",
		}
		if err := c.ExtClient.MySQLs(mysql.Namespace).Delete(mysql.Name, &metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf(
				`Failed to resume MySQL "%v" from DormantDatabase "%v". Error: %v`,
				mysql.Name,
				mysql.Name,
				err,
			)
		}

		_, err := kutildb.TryPatchDormantDatabase(c.ExtClient, mysql.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(mysql.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(mysql.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, mysql.Namespace); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	if err := c.ensureService(mysql); err != nil {
		return err
	}

	// ensure database StatefulSet
	if err := c.ensureStatefulSet(mysql); err != nil {
		return err
	}

	c.recorder.Event(
		mysql.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created MySQL",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(mysql)
	if mysql.Spec.Monitor != nil {
		if err := c.addMonitor(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(mysql *tapi.MySQL) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(mysql.Namespace).Get(mysql.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				mysql.Name,
				err,
			)
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string) (bool, error) {
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	if dormantDb.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindMySQL {
		return sendEvent(fmt.Sprintf(`Invalid MySQL: "%v". Exists DormantDatabase "%v" of different Kind`,
			mysql.Name, dormantDb.Name))
	}

	// Check InitSpec
	initSpecAnnotationStr := dormantDb.Annotations[tapi.MySQLInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *tapi.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if mysql.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, mysql.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.MySQL
	originalSpec := mysql.Spec
	originalSpec.Init = nil

	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &apiv1.SecretVolumeSource{
			SecretName: mysql.Name + "-admin-auth",
		}
	}

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("MySQL spec mismatches with OriginSpec in DormantDatabases")
	}

	return true, nil
}

func (c *Controller) ensureService(mysql *tapi.MySQL) error {
	// Check if service name exists
	found, err := c.findService(mysql)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(mysql); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(mysql *tapi.MySQL) error {
	found, err := c.findStatefulSet(mysql)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Create statefulSet for MySQL database
	statefulSet, err := c.createStatefulSet(mysql)
	if err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	if mysql.Spec.Init != nil && mysql.Spec.Init.SnapshotSource != nil {
		_, err := kutildb.TryPatchMySQL(c.ExtClient, mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
			in.Status.Phase = tapi.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(mysql, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		if err := c.initialize(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	_, err = kutildb.TryPatchMySQL(c.ExtClient, mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
		in.Status.Phase = tapi.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(mysql, apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	return nil
}

func (c *Controller) ensureBackupScheduler(mysql *tapi.MySQL) {
	// Setup Schedule backup
	if mysql.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(mysql, mysql.ObjectMeta, mysql.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(mysql.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(mysql *tapi.MySQL) error {
	snapshotSource := mysql.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		mysql.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = mysql.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		return err
	}

	job, err := c.createRestoreJob(mysql, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(job, mysql, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(mysql *tapi.MySQL) error {
	if mysql.Annotations != nil {
		if val, found := mysql.Annotations["kubedb.com/ignore"]; found {
			c.recorder.Event(mysql.ObjectReference(), apiv1.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(mysql.ObjectReference(), apiv1.EventTypeNormal, eventer.EventReasonPausing, "Pausing MySQL")

	if mysql.Spec.DoNotPause {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`MySQL "%v" is locked.`,
			mysql.Name,
		)

		if err := c.reCreateMySQL(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate MySQL: "%v". Reason: %v`,
				mysql.Name,
				err,
			)
			return err
		}
		return nil
	}

	if _, err := c.createDormantDatabase(mysql); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			apiv1.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			mysql.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		mysql.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		mysql.Name,
	)

	c.cronController.StopBackupScheduling(mysql.ObjectMeta)

	if mysql.Spec.Monitor != nil {
		if err := c.deleteMonitor(mysql); err != nil {
			c.recorder.Eventf(
				mysql.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			mysql.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldMySQL, updatedMySQL *tapi.MySQL) error {
	if err := validator.ValidateMySQL(c.Client, updatedMySQL); err != nil {
		c.recorder.Event(updatedMySQL.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedMySQL.ObjectReference(),
		apiv1.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate MySQL",
	)

	if err := c.ensureService(updatedMySQL); err != nil {
		return err
	}
	if err := c.ensureStatefulSet(updatedMySQL); err != nil {
		return err
	}

	if !reflect.DeepEqual(updatedMySQL.Spec.BackupSchedule, oldMySQL.Spec.BackupSchedule) {
		c.ensureBackupScheduler(updatedMySQL)
	}

	if !reflect.DeepEqual(oldMySQL.Spec.Monitor, updatedMySQL.Spec.Monitor) {
		if err := c.updateMonitor(oldMySQL, updatedMySQL); err != nil {
			c.recorder.Eventf(
				updatedMySQL.ObjectReference(),
				apiv1.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedMySQL.ObjectReference(),
			apiv1.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}
