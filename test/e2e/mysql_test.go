/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package e2e_test

import (
	"fmt"
	"os"
	"strconv"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/mysql/test/e2e/framework"
	"kubedb.dev/mysql/test/e2e/matcher"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	meta_util "kmodules.xyz/client-go/meta"
	store "kmodules.xyz/objectstore-api/api/v1"
	stashV1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashV1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
)

//rbac "k8s.io/api/rbac/v1"
//"k8s.io/apimachinery/pkg/api/resource"

const (
	S3_BUCKET_NAME       = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME      = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME = "SWIFT_CONTAINER_NAME"
	MYSQL_DATABASE       = "MYSQL_DATABASE"
	MYSQL_ROOT_PASSWORD  = "MYSQL_ROOT_PASSWORD"
)

var _ = Describe("MySQL", func() {
	var (
		err          error
		f            *framework.Invocation
		mysql        *api.MySQL
		garbageMySQL *api.MySQLList
		//snapshot         *api.Snapshot
		secret      *core.Secret
		skipMessage string
		//skipDataChecking bool
		dbName string
	)

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQL()
		garbageMySQL = new(api.MySQLList)
		//snapshot = f.Snapshot()
		skipMessage = ""
		//skipDataChecking = true
		dbName = "mysql"

		By("Ensure the apiservices are ready")
		f.EnsureAPIServiceReady().Should(Succeed())
	})

	var createAndWaitForRunning = func() {
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(mysql.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(mysql.ObjectMeta, dbName).Should(BeTrue())
	}

	var testGeneralBehaviour = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}
		// Create MySQL
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

		By("Inserting Rows")
		f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for mysql to be paused")
		f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

		// Create MySQL object again to resume it
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for DormantDatabase to be deleted")
		f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

	}

	//var shouldTakeSnapshot = func() {
	//	// Create and wait for running MySQL
	//	createAndWaitForRunning()
	//
	//	By("Create Secret")
	//	err := f.CreateSecret(secret)
	//	Expect(err).NotTo(HaveOccurred())
	//
	//	By("Create Snapshot")
	//	err = f.CreateSnapshot(snapshot)
	//	Expect(err).NotTo(HaveOccurred())
	//
	//	By("Check for Succeed snapshot")
	//	f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
	//
	//	if !skipDataChecking {
	//		By("Check for snapshot data")
	//		f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
	//	}
	//}

	var shouldInsertData = func() {
		// Create and wait for running MySQL
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

		By("Inserting Row")
		f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		//By("Create Snapshot")
		//err = f.CreateSnapshot(snapshot)
		//Expect(err).NotTo(HaveOccurred())
		//
		//By("Check for Succeed snapshot")
		//f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
		//
		//if !skipDataChecking {
		//	By("Check for snapshot data")
		//	f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//}
	}

	var deleteTestResource = func() {
		if mysql == nil {
			log.Infoln("Skipping cleanup. Reason: mysql is nil")
			return
		}

		By("Check if mysql " + mysql.Name + " exists.")
		my, err := f.GetMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// MySQL was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: MySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if my.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for mysql to be paused")
			f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut mysql")
			_, err := f.PatchDormantDatabase(mysql.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(mysql.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for mysql resources to be wipedOut")
		f.EventuallyWipedOut(mysql.ObjectMeta).Should(Succeed())
	}

	//var deleteSnapshot = func() {
	//
	//	By("Deleting Snapshot: " + snapshot.Name)
	//	err = f.DeleteSnapshot(snapshot.ObjectMeta)
	//	if err != nil && !kerr.IsNotFound(err) {
	//		Expect(err).NotTo(HaveOccurred())
	//	}
	//
	//	if !skipDataChecking {
	//		// do not try to check snapshot data if secret does not exist
	//		_, err = f.GetSecret(secret.ObjectMeta)
	//		if err != nil && kerr.IsNotFound(err) {
	//			log.Infof("Skipping checking snapshot data. Reason: secret %s not found", secret.Name)
	//			return
	//		}
	//		Expect(err).NotTo(HaveOccurred())
	//
	//		By("Checking Snapshot's data wiped out from backend")
	//		f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
	//	}
	//}

	AfterEach(func() {
		// delete resources for current MySQL
		deleteTestResource()

		// old MySQL are in garbageMySQL list. delete their resources.
		for _, my := range garbageMySQL.Items {
			*mysql = my
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers()
	})

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			if mysql.Spec.Replicas == nil {
				mysql.Spec.Replicas = types.Int32P(1)
			}
			f.PrintDebugHelpers(mysql.Name, int(*mysql.Spec.Replicas))
		}
	})

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", testGeneralBehaviour)
			})

			Context("with custom SA Name", func() {
				var customSecret *core.Secret

				BeforeEach(func() {
					customSecret = f.SecretForDatabaseAuthentication(mysql.ObjectMeta, false)
					mysql.Spec.DatabaseSecret = &core.SecretVolumeSource{
						SecretName: customSecret.Name,
					}
					err := f.CreateSecret(customSecret)
					Expect(err).NotTo(HaveOccurred())
					mysql.Spec.PodTemplate.Spec.ServiceAccountName = "my-custom-sa"
					mysql.Spec.TerminationPolicy = api.TerminationPolicyPause
				})

				It("should start and resume successfully", func() {
					//shouldTakeSnapshot()
					createAndWaitForRunning()
					if mysql == nil {
						Skip("Skipping")
					}
					By("Check if MySQL " + mysql.Name + " exists.")
					_, err := f.GetMySQL(mysql.ObjectMeta)
					if err != nil {
						if kerr.IsNotFound(err) {
							// Postgres was not created. Hence, rest of cleanup is not necessary.
							return
						}
						Expect(err).NotTo(HaveOccurred())
					}

					By("Delete mysql: " + mysql.Name)
					err = f.DeleteMySQL(mysql.ObjectMeta)
					if err != nil {
						if kerr.IsNotFound(err) {
							// Postgres was not created. Hence, rest of cleanup is not necessary.
							log.Infof("Skipping rest of cleanup. Reason: Postgres %s is not found.", mysql.Name)
							return
						}
						Expect(err).NotTo(HaveOccurred())
					}

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					By("Resume DB")
					createAndWaitForRunning()
				})
			})

			Context("PDB", func() {

				It("should run eviction successfully", func() {
					mysql.Spec.Replicas = types.Int32P(3)
					// Create MySQL
					By("Create and run MySQL with three replicas")
					createAndWaitForRunning()
					//Evict MySQL pods
					By("Try to evict pods")
					err := f.EvictPodsFromStatefulSet(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		//Context("Snapshot", func() {
		//
		//	BeforeEach(func() {
		//		skipDataChecking = false
		//		snapshot.Spec.DatabaseName = mysql.Name
		//	})
		//
		//	AfterEach(func() {
		//		// delete snapshot and check for data wipeOut
		//		deleteSnapshot()
		//
		//		By("Deleting secret: " + secret.Name)
		//		err := f.DeleteSecret(secret.ObjectMeta)
		//		if err != nil && !kerr.IsNotFound(err) {
		//			Expect(err).NotTo(HaveOccurred())
		//		}
		//	})
		//
		//	Context("For Custom Resources", func() {
		//
		//		BeforeEach(func() {
		//			snapshot.Spec.DatabaseName = mysql.Name
		//			secret = f.SecretForGCSBackend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.GCS = &store.GCSSpec{
		//				Bucket: os.Getenv(GCS_BUCKET_NAME),
		//			}
		//		})
		//
		//		Context("with custom SA", func() {
		//			var customSAForDB *core.ServiceAccount
		//			var customRoleForDB *rbac.Role
		//			var customRoleBindingForDB *rbac.RoleBinding
		//			var customSAForSnapshot *core.ServiceAccount
		//			var customRoleForSnapshot *rbac.Role
		//			var customRoleBindingForSnapshot *rbac.RoleBinding
		//			BeforeEach(func() {
		//				mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
		//				customSAForDB = f.ServiceAccount()
		//				mysql.Spec.PodTemplate.Spec.ServiceAccountName = customSAForDB.Name
		//				customRoleForDB = f.RoleForMySQL(mysql.ObjectMeta)
		//				customRoleBindingForDB = f.RoleBinding(customSAForDB.Name, customRoleForDB.Name)
		//
		//				customSAForSnapshot = f.ServiceAccount()
		//				snapshot.Spec.PodTemplate.Spec.ServiceAccountName = customSAForSnapshot.Name
		//				customRoleForSnapshot = f.RoleForSnapshot(mysql.ObjectMeta)
		//				customRoleBindingForSnapshot = f.RoleBinding(customSAForSnapshot.Name, customRoleForSnapshot.Name)
		//
		//				By("Create Database SA")
		//				err = f.CreateServiceAccount(customSAForDB)
		//				Expect(err).NotTo(HaveOccurred())
		//				By("Create Database Role")
		//				err = f.CreateRole(customRoleForDB)
		//				Expect(err).NotTo(HaveOccurred())
		//				By("Create Database RoleBinding")
		//				err = f.CreateRoleBinding(customRoleBindingForDB)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Create Snapshot SA")
		//				err = f.CreateServiceAccount(customSAForSnapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//				By("Create Snapshot Role")
		//				err = f.CreateRole(customRoleForSnapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//				By("Create Snapshot RoleBinding")
		//				err = f.CreateRoleBinding(customRoleBindingForSnapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//			})
		//
		//			It("should take snapshot successfully", func() {
		//				shouldInsertData()
		//			})
		//
		//			It("should initialize from snapshot successfully", func() {
		//				// Create MySQL and take Snapshot
		//				shouldInsertData()
		//
		//				oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//				garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)
		//
		//				By("Create mysql from snapshot")
		//				mysql = f.MySQL()
		//				mysql.Spec.Init = &api.InitSpec{
		//					SnapshotSource: &api.SnapshotSourceSpec{
		//						Namespace: snapshot.Namespace,
		//						Name:      snapshot.Name,
		//					},
		//				}
		//
		//				By("Creating init Snapshot Mysql without secret name" + mysql.Name)
		//				err = f.CreateMySQL(mysql)
		//				Expect(err).Should(HaveOccurred())
		//
		//				// for snapshot init, user have to use older secret,
		//				// because the username & password  will be replaced to
		//				mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
		//
		//				//Create New role and bind it to existing SA
		//				By("Get new Role and RB")
		//				customRoleForReplayDB := f.RoleForMySQL(mysql.ObjectMeta)
		//				customRoleBindingForReplayDB := f.RoleBinding(customSAForDB.Name, customRoleForReplayDB.Name)
		//
		//				By("Create Database Role")
		//				err = f.CreateRole(customRoleForReplayDB)
		//				Expect(err).NotTo(HaveOccurred())
		//				By("Create Database RoleBinding")
		//				err = f.CreateRoleBinding(customRoleBindingForReplayDB)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				mysql.Spec.PodTemplate.Spec.ServiceAccountName = customSAForDB.Name
		//
		//				// Create and wait for running MySQL
		//				By("Create DB with snapshot data")
		//				createAndWaitForRunning()
		//
		//				By("Checking Row Count of Table")
		//				f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
		//			})
		//		})
		//
		//		Context("with custom Secret", func() {
		//			var customSecret *core.Secret
		//			BeforeEach(func() {
		//				customSecret = f.SecretForDatabaseAuthentication(mysql.ObjectMeta, false)
		//				mysql.Spec.DatabaseSecret = &core.SecretVolumeSource{
		//					SecretName: customSecret.Name,
		//				}
		//
		//			})
		//
		//			It("should not delete database secret", func() {
		//				By("Create Database Secret")
		//				err := f.CreateSecret(customSecret)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Take Snapshot")
		//				shouldTakeSnapshot()
		//				By("Delete test resources")
		//				deleteTestResource()
		//				By("Confirm Database Secret exists")
		//				err = f.CheckSecret(customSecret)
		//				Expect(err).NotTo(HaveOccurred())
		//			})
		//
		//			It("should delete database secret", func() {
		//				kubedbSecret := f.SecretForDatabaseAuthentication(mysql.ObjectMeta, true)
		//				mysql.Spec.DatabaseSecret = &core.SecretVolumeSource{
		//					SecretName: kubedbSecret.Name,
		//				}
		//				By("Create Database Secret")
		//				err = f.CreateSecret(kubedbSecret)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Take Snapshot")
		//				shouldTakeSnapshot()
		//				By("Delete test resources")
		//				deleteTestResource()
		//				By("Confirm Database Secret doesnt exist")
		//				err = f.CheckSecret(kubedbSecret)
		//				Expect(err).To(HaveOccurred())
		//			})
		//		})
		//	})
		//
		//	Context("In Local", func() {
		//
		//		BeforeEach(func() {
		//			skipDataChecking = true
		//			secret = f.SecretForLocalBackend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//		})
		//
		//		Context("With EmptyDir as Snapshot's backend", func() {
		//			BeforeEach(func() {
		//				snapshot.Spec.Local = &store.LocalSpec{
		//					MountPath: "/repo",
		//					VolumeSource: core.VolumeSource{
		//						EmptyDir: &core.EmptyDirVolumeSource{},
		//					},
		//				}
		//			})
		//
		//			It("should take Snapshot successfully", shouldTakeSnapshot)
		//		})
		//
		//		Context("With PVC as Snapshot's backend", func() {
		//			var snapPVC *core.PersistentVolumeClaim
		//
		//			BeforeEach(func() {
		//				snapPVC = f.GetPersistentVolumeClaim()
		//				err := f.CreatePersistentVolumeClaim(snapPVC)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				snapshot.Spec.Local = &store.LocalSpec{
		//					MountPath: "/repo",
		//					VolumeSource: core.VolumeSource{
		//						PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
		//							ClaimName: snapPVC.Name,
		//						},
		//					},
		//				}
		//			})
		//
		//			AfterEach(func() {
		//				err := f.DeletePersistentVolumeClaim(snapPVC.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//			})
		//
		//			It("should delete Snapshot successfully", func() {
		//				shouldTakeSnapshot()
		//
		//				By("Deleting Snapshot")
		//				err := f.DeleteSnapshot(snapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Waiting for Snapshot to be deleted")
		//				f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())
		//			})
		//		})
		//	})
		//
		//	Context("In S3", func() {
		//		BeforeEach(func() {
		//			secret = f.SecretForS3Backend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.S3 = &store.S3Spec{
		//				Bucket: os.Getenv(S3_BUCKET_NAME),
		//			}
		//		})
		//
		//		It("should take Snapshot successfully", shouldInsertData)
		//
		//		Context("faulty snapshot", func() {
		//			BeforeEach(func() {
		//				skipDataChecking = true
		//				snapshot.Spec.StorageSecretName = secret.Name
		//				snapshot.Spec.S3 = &store.S3Spec{
		//					Bucket: "nonexisting",
		//				}
		//			})
		//
		//			It("snapshot should fail", func() {
		//				// Create and wait for running db
		//				createAndWaitForRunning()
		//
		//				By("Create Secret")
		//				err := f.CreateSecret(secret)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Create Snapshot")
		//				err = f.CreateSnapshot(snapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Check for failed snapshot")
		//				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseFailed))
		//			})
		//		})
		//
		//		Context("Delete One Snapshot keeping others", func() {
		//			var initScriptConfigmap *core.ConfigMap
		//
		//			BeforeEach(func() {
		//				initScriptConfigmap, err = f.InitScriptConfigMap()
		//				Expect(err).ShouldNot(HaveOccurred())
		//				By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
		//				Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())
		//
		//				mysql.Spec.Init = &api.InitSpec{
		//					ScriptSource: &api.ScriptSourceSpec{
		//						VolumeSource: core.VolumeSource{
		//							ConfigMap: &core.ConfigMapVolumeSource{
		//								LocalObjectReference: core.LocalObjectReference{
		//									Name: initScriptConfigmap.Name,
		//								},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			AfterEach(func() {
		//				By("Deleting configMap: " + initScriptConfigmap.Name)
		//				Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
		//			})
		//
		//			It("Delete One Snapshot keeping others", func() {
		//				// Create MySQL and take Snapshot
		//				shouldTakeSnapshot()
		//
		//				oldSnapshot := snapshot.DeepCopy()
		//
		//				// New snapshot that has old snapshot's name in prefix
		//				snapshot.Name += "-2"
		//
		//				By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
		//				err = f.CreateSnapshot(snapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Check for Succeeded snapshot")
		//				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
		//
		//				if !skipDataChecking {
		//					By("Check for snapshot data")
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//
		//				// delete old snapshot
		//				By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
		//				err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Waiting for old Snapshot to be deleted")
		//				f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
		//					f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
		//				}
		//
		//				// check remaining snapshot
		//				By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
		//				_, err = f.GetSnapshot(snapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//			})
		//		})
		//	})
		//
		//	Context("In GCS", func() {
		//		BeforeEach(func() {
		//			secret = f.SecretForGCSBackend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.GCS = &store.GCSSpec{
		//				Bucket: os.Getenv(GCS_BUCKET_NAME),
		//			}
		//		})
		//
		//		Context("Without Init", func() {
		//			It("should take Snapshot successfully", shouldInsertData)
		//		})
		//
		//		Context("With Init", func() {
		//			var initScriptConfigmap *core.ConfigMap
		//
		//			BeforeEach(func() {
		//				initScriptConfigmap, err = f.InitScriptConfigMap()
		//				Expect(err).ShouldNot(HaveOccurred())
		//				By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
		//				Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())
		//
		//				mysql.Spec.Init = &api.InitSpec{
		//					ScriptSource: &api.ScriptSourceSpec{
		//						VolumeSource: core.VolumeSource{
		//							ConfigMap: &core.ConfigMapVolumeSource{
		//								LocalObjectReference: core.LocalObjectReference{
		//									Name: initScriptConfigmap.Name,
		//								},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			AfterEach(func() {
		//				By("Deleting configMap: " + initScriptConfigmap.Name)
		//				Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
		//			})
		//
		//			It("should take Snapshot successfully", shouldTakeSnapshot)
		//		})
		//
		//		Context("Delete One Snapshot keeping others", func() {
		//			var initScriptConfigmap *core.ConfigMap
		//
		//			BeforeEach(func() {
		//				initScriptConfigmap, err = f.InitScriptConfigMap()
		//				Expect(err).ShouldNot(HaveOccurred())
		//				By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
		//				Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())
		//
		//				mysql.Spec.Init = &api.InitSpec{
		//					ScriptSource: &api.ScriptSourceSpec{
		//						VolumeSource: core.VolumeSource{
		//							ConfigMap: &core.ConfigMapVolumeSource{
		//								LocalObjectReference: core.LocalObjectReference{
		//									Name: initScriptConfigmap.Name,
		//								},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			AfterEach(func() {
		//				By("Deleting configMap: " + initScriptConfigmap.Name)
		//				Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
		//			})
		//
		//			It("Delete One Snapshot keeping others", func() {
		//				// Create MySQL and take Snapshot
		//				shouldTakeSnapshot()
		//
		//				oldSnapshot := snapshot.DeepCopy()
		//
		//				// New snapshot that has old snapshot's name in prefix
		//				snapshot.Name += "-2"
		//
		//				By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
		//				err = f.CreateSnapshot(snapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Check for Succeeded snapshot")
		//				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
		//
		//				if !skipDataChecking {
		//					By("Check for snapshot data")
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//
		//				// delete old snapshot
		//				By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
		//				err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Waiting for old Snapshot to be deleted")
		//				f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
		//					f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
		//				}
		//
		//				// check remaining snapshot
		//				By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
		//				_, err = f.GetSnapshot(snapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//			})
		//		})
		//
		//	})
		//
		//	Context("In Azure", func() {
		//		BeforeEach(func() {
		//			secret = f.SecretForAzureBackend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.Azure = &store.AzureSpec{
		//				Container: os.Getenv(AZURE_CONTAINER_NAME),
		//			}
		//		})
		//
		//		It("should take Snapshot successfully", shouldInsertData)
		//
		//		Context("Delete One Snapshot keeping others", func() {
		//			var initScriptConfigmap *core.ConfigMap
		//
		//			BeforeEach(func() {
		//				initScriptConfigmap, err = f.InitScriptConfigMap()
		//				Expect(err).ShouldNot(HaveOccurred())
		//				By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
		//				Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())
		//
		//				mysql.Spec.Init = &api.InitSpec{
		//					ScriptSource: &api.ScriptSourceSpec{
		//						VolumeSource: core.VolumeSource{
		//							ConfigMap: &core.ConfigMapVolumeSource{
		//								LocalObjectReference: core.LocalObjectReference{
		//									Name: initScriptConfigmap.Name,
		//								},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			AfterEach(func() {
		//				By("Deleting configMap: " + initScriptConfigmap.Name)
		//				Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
		//			})
		//
		//			It("Delete One Snapshot keeping others", func() {
		//				// Create MySQL and take Snapshot
		//				shouldTakeSnapshot()
		//
		//				oldSnapshot := snapshot.DeepCopy()
		//
		//				// New snapshot that has old snapshot's name in prefix
		//				snapshot.Name += "-2"
		//
		//				By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
		//				err = f.CreateSnapshot(snapshot)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Check for Succeeded snapshot")
		//				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
		//
		//				if !skipDataChecking {
		//					By("Check for snapshot data")
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//
		//				// delete old snapshot
		//				By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
		//				err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				By("Waiting for old Snapshot to be deleted")
		//				f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
		//					f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
		//				}
		//
		//				// check remaining snapshot
		//				By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
		//				_, err = f.GetSnapshot(snapshot.ObjectMeta)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				if !skipDataChecking {
		//					By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
		//					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//				}
		//			})
		//		})
		//	})
		//
		//	Context("In Swift", func() {
		//		BeforeEach(func() {
		//			secret = f.SecretForSwiftBackend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.Swift = &store.SwiftSpec{
		//				Container: os.Getenv(SWIFT_CONTAINER_NAME),
		//			}
		//		})
		//
		//		It("should take Snapshot successfully", shouldInsertData)
		//	})
		//
		//	Context("Snapshot PodVolume Template - In S3", func() {
		//
		//		BeforeEach(func() {
		//			secret = f.SecretForS3Backend()
		//			snapshot.Spec.StorageSecretName = secret.Name
		//			snapshot.Spec.S3 = &store.S3Spec{
		//				Bucket: os.Getenv(S3_BUCKET_NAME),
		//			}
		//		})
		//
		//		var shouldHandleJobVolumeSuccessfully = func() {
		//			// Create and wait for running MySQL
		//			createAndWaitForRunning()
		//
		//			By("Get MySQL")
		//			es, err := f.GetMySQL(mysql.ObjectMeta)
		//			Expect(err).NotTo(HaveOccurred())
		//			mysql.Spec = es.Spec
		//
		//			By("Create Secret")
		//			err = f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			// determine pvcSpec and storageType for job
		//			// start
		//			pvcSpec := snapshot.Spec.PodVolumeClaimSpec
		//			if pvcSpec == nil {
		//				pvcSpec = mysql.Spec.Storage
		//			}
		//			st := snapshot.Spec.StorageType
		//			if st == nil {
		//				st = &mysql.Spec.StorageType
		//			}
		//			Expect(st).NotTo(BeNil())
		//			// end
		//
		//			By("Create Snapshot")
		//			err = f.CreateSnapshot(snapshot)
		//			if *st == api.StorageTypeDurable && pvcSpec == nil {
		//				By("Create Snapshot should have failed")
		//				Expect(err).Should(HaveOccurred())
		//				return
		//			} else {
		//				Expect(err).NotTo(HaveOccurred())
		//			}
		//
		//			By("Get Snapshot")
		//			snap, err := f.GetSnapshot(snapshot.ObjectMeta)
		//			Expect(err).NotTo(HaveOccurred())
		//			snapshot.Spec = snap.Spec
		//
		//			if *st == api.StorageTypeEphemeral {
		//				storageSize := "0"
		//				if pvcSpec != nil {
		//					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
		//						storageSize = sz.String()
		//					}
		//				}
		//				By(fmt.Sprintf("Check for Job Empty volume size: %v", storageSize))
		//				f.EventuallyJobVolumeEmptyDirSize(snapshot.ObjectMeta).Should(Equal(storageSize))
		//			} else if *st == api.StorageTypeDurable {
		//				sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]
		//				Expect(found).NotTo(BeFalse())
		//
		//				By("Check for Job PVC Volume size from snapshot")
		//				f.EventuallyJobPVCSize(snapshot.ObjectMeta).Should(Equal(sz.String()))
		//			}
		//
		//			By("Check for succeeded snapshot")
		//			f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))
		//
		//			if !skipDataChecking {
		//				By("Check for snapshot data")
		//				f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		//			}
		//		}
		//
		//		// db StorageType Scenarios
		//		// ==============> Start
		//		var dbStorageTypeScenarios = func() {
		//			Context("DBStorageType - Durable", func() {
		//				BeforeEach(func() {
		//					mysql.Spec.StorageType = api.StorageTypeDurable
		//					mysql.Spec.Storage = &core.PersistentVolumeClaimSpec{
		//						Resources: core.ResourceRequirements{
		//							Requests: core.ResourceList{
		//								core.ResourceStorage: resource.MustParse(framework.DBPvcStorageSize),
		//							},
		//						},
		//						StorageClassName: types.StringP(root.StorageClass),
		//					}
		//
		//				})
		//
		//				It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
		//			})
		//
		//			Context("DBStorageType - Ephemeral", func() {
		//				BeforeEach(func() {
		//					mysql.Spec.StorageType = api.StorageTypeEphemeral
		//					mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
		//				})
		//
		//				Context("DBPvcSpec is nil", func() {
		//					BeforeEach(func() {
		//						mysql.Spec.Storage = nil
		//					})
		//
		//					It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
		//				})
		//
		//				Context("DBPvcSpec is given [not nil]", func() {
		//					BeforeEach(func() {
		//						mysql.Spec.Storage = &core.PersistentVolumeClaimSpec{
		//							Resources: core.ResourceRequirements{
		//								Requests: core.ResourceList{
		//									core.ResourceStorage: resource.MustParse(framework.DBPvcStorageSize),
		//								},
		//							},
		//							StorageClassName: types.StringP(root.StorageClass),
		//						}
		//					})
		//
		//					It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
		//				})
		//			})
		//		}
		//		// End <==============
		//
		//		// Snapshot PVC Scenarios
		//		// ==============> Start
		//		var snapshotPvcScenarios = func() {
		//			Context("Snapshot PVC is given [not nil]", func() {
		//				BeforeEach(func() {
		//					snapshot.Spec.PodVolumeClaimSpec = &core.PersistentVolumeClaimSpec{
		//						Resources: core.ResourceRequirements{
		//							Requests: core.ResourceList{
		//								core.ResourceStorage: resource.MustParse(framework.JobPvcStorageSize),
		//							},
		//						},
		//						StorageClassName: types.StringP(root.StorageClass),
		//					}
		//				})
		//
		//				dbStorageTypeScenarios()
		//			})
		//
		//			Context("Snapshot PVC is nil", func() {
		//				BeforeEach(func() {
		//					snapshot.Spec.PodVolumeClaimSpec = nil
		//				})
		//
		//				dbStorageTypeScenarios()
		//			})
		//		}
		//		// End <==============
		//
		//		Context("Snapshot StorageType is nil", func() {
		//			BeforeEach(func() {
		//				snapshot.Spec.StorageType = nil
		//			})
		//
		//			snapshotPvcScenarios()
		//		})
		//
		//		Context("Snapshot StorageType is Ephemeral", func() {
		//			BeforeEach(func() {
		//				ephemeral := api.StorageTypeEphemeral
		//				snapshot.Spec.StorageType = &ephemeral
		//			})
		//
		//			snapshotPvcScenarios()
		//		})
		//
		//		Context("Snapshot StorageType is Durable", func() {
		//			BeforeEach(func() {
		//				durable := api.StorageTypeDurable
		//				snapshot.Spec.StorageType = &durable
		//			})
		//
		//			snapshotPvcScenarios()
		//		})
		//	})
		//})

		Context("Initialize", func() {

			Context("With Script", func() {
				var initScriptConfigmap *core.ConfigMap

				BeforeEach(func() {
					initScriptConfigmap, err = f.InitScriptConfigMap()
					Expect(err).ShouldNot(HaveOccurred())
					By("Create init Script ConfigMap: " + initScriptConfigmap.Name + "\n" + initScriptConfigmap.Data["init.sql"])
					Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())

					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: initScriptConfigmap.Name,
									},
								},
							},
						},
					}
				})

				AfterEach(func() {
					By("Deleting configMap: " + initScriptConfigmap.Name)
					Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
				})

				It("should run successfully", func() {
					// Create MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				})
			})

			//Context("With Snapshot", func() {
			//
			//	AfterEach(func() {
			//		// delete snapshot and check for data wipeOut
			//		deleteSnapshot()
			//
			//		By("Deleting secret: " + secret.Name)
			//		err := f.DeleteSecret(secret.ObjectMeta)
			//		if err != nil && !kerr.IsNotFound(err) {
			//			Expect(err).NotTo(HaveOccurred())
			//		}
			//	})
			//
			//	var shouldInitializeFromSnapshot = func() {
			//		// Create MySQL and take Snapshot
			//		shouldInsertData()
			//
			//		oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
			//		Expect(err).NotTo(HaveOccurred())
			//		garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)
			//
			//		By("Create mysql from snapshot")
			//		mysql = f.MySQL()
			//		mysql.Spec.Init = &api.InitSpec{
			//			SnapshotSource: &api.SnapshotSourceSpec{
			//				Namespace: snapshot.Namespace,
			//				Name:      snapshot.Name,
			//			},
			//		}
			//
			//		By("Creating init Snapshot Mysql without secret name" + mysql.Name)
			//		err = f.CreateMySQL(mysql)
			//		Expect(err).Should(HaveOccurred())
			//
			//		// for snapshot init, user have to use older secret,
			//		// because the username & password  will be replaced to
			//		mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
			//
			//		// Create and wait for running MySQL
			//		createAndWaitForRunning()
			//
			//		By("Checking Row Count of Table")
			//		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
			//	}
			//
			//	Context("From Local backend", func() {
			//		var snapPVC *core.PersistentVolumeClaim
			//
			//		BeforeEach(func() {
			//
			//			skipDataChecking = true
			//			snapPVC = f.GetPersistentVolumeClaim()
			//			err := f.CreatePersistentVolumeClaim(snapPVC)
			//			Expect(err).NotTo(HaveOccurred())
			//
			//			secret = f.SecretForLocalBackend()
			//			snapshot.Spec.DatabaseName = mysql.Name
			//			snapshot.Spec.StorageSecretName = secret.Name
			//
			//			snapshot.Spec.Local = &store.LocalSpec{
			//				MountPath: "/repo",
			//				VolumeSource: core.VolumeSource{
			//					PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
			//						ClaimName: snapPVC.Name,
			//					},
			//				},
			//			}
			//		})
			//
			//		AfterEach(func() {
			//			err := f.DeletePersistentVolumeClaim(snapPVC.ObjectMeta)
			//			Expect(err).NotTo(HaveOccurred())
			//		})
			//
			//		It("should initialize successfully", shouldInitializeFromSnapshot)
			//	})
			//
			//	Context("From GCS backend", func() {
			//
			//		BeforeEach(func() {
			//
			//			skipDataChecking = false
			//			secret = f.SecretForGCSBackend()
			//			snapshot.Spec.StorageSecretName = secret.Name
			//			snapshot.Spec.DatabaseName = mysql.Name
			//
			//			snapshot.Spec.GCS = &store.GCSSpec{
			//				Bucket: os.Getenv(GCS_BUCKET_NAME),
			//			}
			//		})
			//
			//		It("should initialize successfully", shouldInitializeFromSnapshot)
			//	})
			//})

			// To run this test,
			// 1st: Deploy stash latest operator
			// 2nd: create mysql related tasks and functions either
			// 		from `kubedb.dev/mysql/hack/dev/examples/stash01_config.yaml`
			//	 or	from helm chart in `stash.appscode.dev/mysql/chart/mysql-stash`
			Context("With Stash/Restic", func() {
				var bc *stashV1beta1.BackupConfiguration
				var bs *stashV1beta1.BackupSession
				var rs *stashV1beta1.RestoreSession
				var repo *stashV1alpha1.Repository

				BeforeEach(func() {
					//skipDataChecking = true
					if !f.FoundStashCRDs() {
						Skip("Skipping tests for stash integration. reason: stash operator is not running.")
					}
				})

				AfterEach(func() {
					By("Deleting BackupConfiguration")
					err := f.DeleteBackupConfiguration(bc.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Deleting RestoreSession")
					err = f.DeleteRestoreSession(rs.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Deleting Repository")
					err = f.DeleteRepository(repo.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})

				var createAndWaitForInitializing = func() {
					By("Creating MySQL: " + mysql.Name + " with replicas " + strconv.Itoa(int(*mysql.Spec.Replicas)))
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Initializing mysql")
					f.EventuallyMySQLPhase(mysql.ObjectMeta).Should(Equal(api.DatabasePhaseInitializing))

					By("Wait for AppBinding to create")
					f.EventuallyAppBinding(mysql.ObjectMeta).Should(BeTrue())

					By("Check valid AppBinding Specs")
					err = f.CheckAppBindingSpec(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Waiting for database to be ready")
					f.EventuallyDatabaseReady(mysql.ObjectMeta, dbName).Should(BeTrue())
				}

				var shouldInitializeFromStash = func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Rows")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					By("Create Secret")
					err = f.CreateSecret(secret)
					Expect(err).NotTo(HaveOccurred())

					By("Create Repositories")
					err = f.CreateRepository(repo)
					Expect(err).NotTo(HaveOccurred())

					By("Create BackupConfiguration")
					err = f.CreateBackupConfiguration(bc)
					Expect(err).NotTo(HaveOccurred())

					By("Wait until BackupSession be created")
					bs, err = f.WaitUntilBackkupSessionBeCreated(bc.ObjectMeta)

					// eventually backupsession succeeded
					By("Check for Succeeded backupsession")
					f.EventuallyBackupSessionPhase(bs.ObjectMeta).Should(Equal(stashV1beta1.BackupSessionSucceeded))

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)

					By("Create mysql from stash")
					*mysql = *f.MySQL()
					rs = f.RestoreSession(mysql.ObjectMeta, oldMySQL.ObjectMeta)
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
					mysql.Spec.Init = &api.InitSpec{
						StashRestoreSession: &core.LocalObjectReference{
							Name: rs.Name,
						},
					}

					// Create and wait for running MySQL
					createAndWaitForInitializing()

					By("Create RestoreSession")
					err = f.CreateRestoreSession(rs)
					Expect(err).NotTo(HaveOccurred())

					// eventually backupsession succeeded
					By("Check for Succeeded restoreSession")
					f.EventuallyRestoreSessionPhase(rs.ObjectMeta).Should(Equal(stashV1beta1.RestoreSessionSucceeded))

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				}

				Context("From GCS backend", func() {

					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						secret = f.PatchSecretForRestic(secret)
						bc = f.BackupConfiguration(mysql.ObjectMeta)
						repo = f.Repository(mysql.ObjectMeta, secret.Name)

						repo.Spec.Backend = store.Backend{
							GCS: &store.GCSSpec{
								Bucket: os.Getenv("GCS_BUCKET_NAME"),
								Prefix: fmt.Sprintf("stash/%v/%v", mysql.Namespace, mysql.Name),
							},
							StorageSecretName: secret.Name,
						}
					})

					It("should run successfully", shouldInitializeFromStash)
				})

			})

		})

		Context("Resume", func() {

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					// Delete without caring if DB is resumed
					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for MySQL to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				})
			})

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				})
			})

			Context("with init Script", func() {
				var initScriptConfigmap *core.ConfigMap

				BeforeEach(func() {
					initScriptConfigmap, err = f.InitScriptConfigMap()
					Expect(err).ShouldNot(HaveOccurred())
					By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
					Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())

					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: initScriptConfigmap.Name,
									},
								},
							},
						},
					}
				})

				AfterEach(func() {
					By("Deleting configMap: " + initScriptConfigmap.Name)
					Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					mysql, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(mysql.Spec.Init).NotTo(BeNil())

					By("Checking MySQL crd does not have kubedb.com/initialized annotation")
					_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
					Expect(err).To(HaveOccurred())
				})
			})

			//Context("With Snapshot Init", func() {
			//
			//	AfterEach(func() {
			//		// delete snapshot and check for data wipeOut
			//		deleteSnapshot()
			//
			//		By("Deleting secret: " + secret.Name)
			//		err := f.DeleteSecret(secret.ObjectMeta)
			//		if err != nil && !kerr.IsNotFound(err) {
			//			Expect(err).NotTo(HaveOccurred())
			//		}
			//	})
			//
			//	BeforeEach(func() {
			//		skipDataChecking = false
			//		secret = f.SecretForGCSBackend()
			//		snapshot.Spec.StorageSecretName = secret.Name
			//		snapshot.Spec.GCS = &store.GCSSpec{
			//			Bucket: os.Getenv(GCS_BUCKET_NAME),
			//		}
			//		snapshot.Spec.DatabaseName = mysql.Name
			//	})
			//
			//	It("should resume successfully", func() {
			//		// Create MySQL and take Snapshot
			//		shouldInsertData()
			//
			//		oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)
			//
			//		By("Create mysql from snapshot")
			//		mysql = f.MySQL()
			//		mysql.Spec.Init = &api.InitSpec{
			//			SnapshotSource: &api.SnapshotSourceSpec{
			//				Namespace: snapshot.Namespace,
			//				Name:      snapshot.Name,
			//			},
			//		}
			//
			//		By("Creating MySQL without secret name to init from Snapshot: " + mysql.Name)
			//		err = f.CreateMySQL(mysql)
			//		Expect(err).Should(HaveOccurred())
			//
			//		// for snapshot init, user have to use older secret,
			//		// because the username & password  will be replaced to
			//		mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
			//
			//		// Create and wait for running MySQL
			//		createAndWaitForRunning()
			//
			//		By("Checking Row Count of Table")
			//		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
			//
			//		By("Delete mysql")
			//		err = f.DeleteMySQL(mysql.ObjectMeta)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		By("Wait for mysql to be paused")
			//		f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())
			//
			//		// Create MySQL object again to resume it
			//		By("Create MySQL: " + mysql.Name)
			//		err = f.CreateMySQL(mysql)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		By("Wait for DormantDatabase to be deleted")
			//		f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())
			//
			//		By("Wait for Running mysql")
			//		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())
			//
			//		By("Checking Row Count of Table")
			//		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
			//
			//		mysql, err = f.GetMySQL(mysql.ObjectMeta)
			//		Expect(err).NotTo(HaveOccurred())
			//		Expect(mysql.Spec.Init).ShouldNot(BeNil())
			//
			//		By("Checking MySQL has kubedb.com/initialized annotation")
			//		_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
			//		Expect(err).NotTo(HaveOccurred())
			//	})
			//})

			Context("Multiple times with init", func() {
				var initScriptConfigmap *core.ConfigMap

				BeforeEach(func() {
					initScriptConfigmap, err = f.InitScriptConfigMap()
					Expect(err).ShouldNot(HaveOccurred())
					By("Create init Script ConfigMap: " + initScriptConfigmap.Name)
					Expect(f.CreateConfigMap(initScriptConfigmap)).ShouldNot(HaveOccurred())

					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: initScriptConfigmap.Name,
									},
								},
							},
						},
					}
				})

				AfterEach(func() {
					By("Deleting configMap: " + initScriptConfigmap.Name)
					Expect(f.DeleteConfigMap(initScriptConfigmap.ObjectMeta)).NotTo(HaveOccurred())
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")

						By("Delete mysql")
						err = f.DeleteMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for mysql to be paused")
						f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

						// Create MySQL object again to resume it
						By("Create MySQL: " + mysql.Name)
						err = f.CreateMySQL(mysql)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

						By("Wait for Running mysql")
						f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

						By("Checking Row Count of Table")
						f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

						mysql, err := f.GetMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
						Expect(mysql.Spec.Init).ShouldNot(BeNil())

						By("Checking MySQL crd does not have kubedb.com/initialized annotation")
						_, err = meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
						Expect(err).To(HaveOccurred())
					}
				})
			})
		})

		//Context("SnapshotScheduler", func() {
		//
		//	BeforeEach(func() {
		//		skipDataChecking = false
		//	})
		//
		//	AfterEach(func() {
		//		snapshotList, err := f.GetSnapshotList(mysql.ObjectMeta)
		//		Expect(err).NotTo(HaveOccurred())
		//
		//		for _, snap := range snapshotList.Items {
		//			snapshot = &snap
		//
		//			// delete snapshot and check for data wipeOut
		//			deleteSnapshot()
		//		}
		//
		//		By("Deleting secret: " + secret.Name)
		//		err = f.DeleteSecret(secret.ObjectMeta)
		//		if err != nil && !kerr.IsNotFound(err) {
		//			Expect(err).NotTo(HaveOccurred())
		//		}
		//	})
		//
		//	Context("With Startup", func() {
		//
		//		var shouldStartupSchedular = func() {
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			// Create and wait for running MySQL
		//			createAndWaitForRunning()
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Remove Backup Scheduler from MySQL")
		//			_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
		//		}
		//
		//		Context("with local", func() {
		//			BeforeEach(func() {
		//				skipDataChecking = true
		//				secret = f.SecretForLocalBackend()
		//				mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			It("should run scheduler successfully", shouldStartupSchedular)
		//		})
		//
		//		Context("with GCS", func() {
		//			BeforeEach(func() {
		//				secret = f.SecretForGCSBackend()
		//				mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 1m",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						GCS: &store.GCSSpec{
		//							Bucket: os.Getenv(GCS_BUCKET_NAME),
		//						},
		//					},
		//				}
		//			})
		//
		//			It("should run scheduler successfully", shouldStartupSchedular)
		//		})
		//	})
		//
		//	Context("With Update - with Local", func() {
		//
		//		BeforeEach(func() {
		//			skipDataChecking = true
		//			secret = f.SecretForLocalBackend()
		//		})
		//
		//		It("should run scheduler successfully", func() {
		//			// Create and wait for running MySQL
		//			createAndWaitForRunning()
		//
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Update mysql")
		//			_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
		//				in.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Remove Backup Scheduler from MySQL")
		//			_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
		//		})
		//	})
		//
		//	Context("Re-Use DormantDatabase's scheduler", func() {
		//
		//		BeforeEach(func() {
		//			skipDataChecking = true
		//			secret = f.SecretForLocalBackend()
		//		})
		//
		//		It("should re-use scheduler successfully", func() {
		//			// Create and wait for running MySQL
		//			createAndWaitForRunning()
		//
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Update mysql")
		//			_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
		//				in.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Creating Table")
		//			f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())
		//
		//			By("Inserting Row")
		//			f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())
		//
		//			By("Checking Row Count of Table")
		//			f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
		//
		//			By("Delete mysql")
		//			err = f.DeleteMySQL(mysql.ObjectMeta)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Wait for mysql to be paused")
		//			f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())
		//
		//			// Create MySQL object again to resume it
		//			By("Create MySQL: " + mysql.Name)
		//			err = f.CreateMySQL(mysql)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Wait for DormantDatabase to be deleted")
		//			f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())
		//
		//			By("Wait for Running mysql")
		//			f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())
		//
		//			By("Checking Row Count of Table")
		//			f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(5))
		//
		//			By("Remove Backup Scheduler from MySQL")
		//			_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())
		//		})
		//	})
		//})

		Context("Termination Policy", func() {

			BeforeEach(func() {
				//skipDataChecking = false
				secret = f.SecretForGCSBackend()
				//snapshot.Spec.StorageSecretName = secret.Name
				//snapshot.Spec.GCS = &store.GCSSpec{
				//	Bucket: os.Getenv(GCS_BUCKET_NAME),
				//}
				//snapshot.Spec.DatabaseName = mysql.Name
			})

			Context("with TerminationDoNotTerminate", func() {
				BeforeEach(func() {
					//skipDataChecking = true
					mysql.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("MySQL is not paused. Check for mysql")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeTrue())

					By("Check for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Update mysql to set spec.terminationPolicy = Pause")
					_, err := f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("with TerminationPolicyPause (default)", func() {

				AfterEach(func() {
					//// delete snapshot and check for data wipeOut
					//deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should create DormantDatabase and resume from it", func() {
					// Run MySQL and take snapshot
					shouldInsertData()

					By("Deleting MySQL crd")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means mysql object is deleted
					By("Waiting for mysql to be paused")
					f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

					By("Checking PVC hasn't been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(1))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))

					//By("Checking snapshot hasn't been deleted")
					//f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())
					//
					//if !skipDataChecking {
					//	By("Check for snapshot data")
					//	f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					//}

					// Create MySQL object again to resume it
					By("Create (resume) MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking row count of table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				})
			})

			Context("with TerminationPolicyDelete", func() {

				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				AfterEach(func() {
					//// delete snapshot and check for data wipeOut
					//deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should not create DormantDatabase and should not delete secret", func() {
					// Run MySQL
					shouldInsertData()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until mysql is deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVC has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))

					//By("Checking Snapshot hasn't been deleted")
					//f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())
					//
					//if !skipDataChecking {
					//	By("Check for intact snapshot data")
					//	f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					//}
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should not create DormantDatabase and should wipeOut all", func() {
					// Run MySQL
					shouldInsertData()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until mysql is deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVCs has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

					//By("Checking Snapshots has been deleted")
					//f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())

					By("Checking Secrets has been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(0))
				})
			})
		})

		Context("EnvVars", func() {

			Context("Database Name as EnvVar", func() {

				It("should create DB with name provided in EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()
				})
			})

			Context("Root Password as EnvVar", func() {

				It("should reject to create MySQL CRD", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_ROOT_PASSWORD,
							Value: "not@secret",
						},
					}
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("Update EnvVar", func() {

				It("should not reject to update EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					mysql.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()

					By("Patching EnvVar")
					_, _, err = util.PatchMySQL(f.ExtClient().KubedbV1alpha1(), mysql, func(in *api.MySQL) *api.MySQL {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  MYSQL_DATABASE,
								Value: "patched-db",
							},
						}
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("Custom config", func() {

			customConfigs := []string{
				"max_connections=200",
				"read_buffer_size=1048576", // 1MB
			}

			Context("from configMap", func() {
				var userConfig *core.ConfigMap

				BeforeEach(func() {
					userConfig = f.GetCustomConfig(customConfigs)
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					err := f.DeleteConfigMap(userConfig.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

				})

				It("should set configuration provided in configMap", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					By("Creating configMap: " + userConfig.Name)
					err := f.CreateConfigMap(userConfig)
					Expect(err).NotTo(HaveOccurred())

					mysql.Spec.ConfigSource = &core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: userConfig.Name,
							},
						},
					}

					// Create MySQL
					createAndWaitForRunning()

					By("Checking mysql configured from provided custom configuration")
					for _, cfg := range customConfigs {
						f.EventuallyMySQLVariable(mysql.ObjectMeta, dbName, cfg).Should(matcher.UseCustomConfig(cfg))
					}
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create MySQL
				createAndWaitForRunning()

				By("Creating Table")
				f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

				By("Inserting Rows")
				f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

				By("Checking Row Count of Table")
				f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
			}

			Context("Ephemeral", func() {

				Context("General Behaviour", func() {

					BeforeEach(func() {
						mysql.Spec.StorageType = api.StorageTypeEphemeral
						mysql.Spec.Storage = nil
						mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						mysql.Spec.StorageType = api.StorageTypeEphemeral
						mysql.Spec.Storage = nil
						mysql.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create MySQL object", func() {

						By("Creating MySQL: " + mysql.Name)
						err := f.CreateMySQL(mysql)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
