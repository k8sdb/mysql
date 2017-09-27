package validator

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	amv "github.com/k8sdb/apimachinery/pkg/validator"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// TODO: Change method name. ValidateMySQL -> Validate<--->
func ValidateMySQL(client clientset.Interface, mysql *tapi.MySQL) error {
	if mysql.Spec.Version == "" {
		fmt.Println("Error in validation", "`Object 'Version' is missing in '%v'`, mysql.Spec")
		return fmt.Errorf(`Object 'Version' is missing in '%v'`, mysql.Spec)
	}

	// Set Database Image version
	version := fmt.Sprintf("%v", mysql.Spec.Version) // #Later , add -db with %v, ex: "%v-db"
	fmt.Println("Validate:", "database Image version", version)
	// TODO: docker.ImageMySQL should hold correct image name
	if err := docker.CheckDockerImageVersion("library/mysql", version); err != nil { // #LATER , docker image
		return fmt.Errorf(`Image %v:%v not found`, "mysql", version) // #LATER, Docker image
	}
	fmt.Println("Database image exists")

	fmt.Println("Validate:", "Storage", version)
	if mysql.Spec.Storage != nil {
		var err error
		if err = amv.ValidateStorage(client, mysql.Spec.Storage); err != nil {
			return err
		}
	}

	fmt.Println("Validate:", "Storage validation done")

	//// ---> Start
	//// TODO: Use following if database needs/supports authentication secret
	//// otherwise, delete
	//databaseSecret := mysql.Spec.DatabaseSecret
	//if databaseSecret != nil {
	//	if _, err := client.CoreV1().Secrets(mysql.Namespace).Get(databaseSecret.SecretName, metav1.GetOptions{}); err != nil {
	//		return err
	//	}
	//}
	//// ---> End

	backupScheduleSpec := mysql.Spec.BackupSchedule
	if backupScheduleSpec != nil {
		if err := amv.ValidateBackupSchedule(client, backupScheduleSpec, mysql.Namespace); err != nil {
			return err
		}
	}

	monitorSpec := mysql.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}

	}
	return nil
}
