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
	"context"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
	store "kmodules.xyz/objectstore-api/api/v1"
	stashV1alpha1 "stash.appscode.dev/apimachinery/apis/stash/v1alpha1"
	stashV1beta1 "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

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
		secret       *core.Secret
		skipMessage  string
		dbName       string
	)

	BeforeEach(func() {
		secret = nil
		f = root.Invoke()
		mysql = f.MySQL()
		garbageMySQL = new(api.MySQLList)
		skipMessage = ""
		dbName = "mysql"
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

		By("Wait for mysql to be deleted")
		f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

		// Create MySQL object again to resume it
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

	}

	var shouldInsertData = func() {
		// Create and wait for running MySQL
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

		By("Inserting Row")
		f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 0, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
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

		By("Update mysql to set spec.terminationPolicy = WipeOut")
		_, err = f.PatchMySQL(my.ObjectMeta, func(in *api.MySQL) *api.MySQL {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		})
		Expect(err).NotTo(HaveOccurred())

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: MySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for mysql to be deleted")
		f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

		By("Wait for mysql resources to be wipedOut")
		f.EventuallyWipedOut(mysql.ObjectMeta).Should(Succeed())
	}

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
					mysql.Spec.TerminationPolicy = api.TerminationPolicyHalt
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
							// MySQL was not created. Hence, rest of cleanup is not necessary.
							return
						}
						Expect(err).NotTo(HaveOccurred())
					}

					By("Delete mysql: " + mysql.Name)
					err = f.DeleteMySQL(mysql.ObjectMeta)
					if err != nil {
						if kerr.IsNotFound(err) {
							// MySQL was not created. Hence, rest of cleanup is not necessary.
							log.Infof("Skipping rest of cleanup. Reason: MySQL %s is not found.", mysql.Name)
							return
						}
						Expect(err).NotTo(HaveOccurred())
					}

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

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

			// To run this test,
			// 1st: Deploy stash latest operator
			// 2nd: create mysql related tasks and functions either
			// 		from `kubedb.dev/mysql/hack/dev/examples/stash01_config.yaml`
			//	 or	from helm chart in `stash.appscode.dev/mysql/chart/mysql-stash`
			Context("With Stash/Restic", func() {
				var bc *stashV1beta1.BackupConfiguration
				var rs *stashV1beta1.RestoreSession
				var repo *stashV1alpha1.Repository

				BeforeEach(func() {
					if !f.FoundStashCRDs() {
						Skip("Skipping tests for stash integration. reason: stash operator is not running.")
					}
				})

				AfterEach(func() {
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

					By("Check for snapshot count in stash-repository")
					f.EventuallySnapshotInRepository(repo.ObjectMeta).Should(matcher.MoreThan(2))

					By("Delete BackupConfiguration to stop backup scheduling")
					err = f.DeleteBackupConfiguration(bc.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbageMySQL.Items = append(garbageMySQL.Items, *oldMySQL)

					By("Create mysql from stash")
					*mysql = *f.MySQL()
					rs = f.RestoreSession(mysql.ObjectMeta, repo)
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
					mysql.Spec.Init = &api.InitSpec{
						StashRestoreSession: &core.LocalObjectReference{
							Name: rs.Name,
						},
					}

					// Create and wait for running MySQL
					createAndWaitForInitializing()

					By("Waiting for database to be ready")
					f.EventuallyDatabaseReady(mysql.ObjectMeta, dbName).Should(BeTrue())

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
						repo = f.Repository(mysql.ObjectMeta)
						bc = f.BackupConfiguration(mysql.ObjectMeta, repo)

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
				It("should resume database successfully", func() {
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

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

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

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))
				})
			})

			Context("Without Init", func() {
				It("should resume database successfully", func() {
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

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

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

				It("should resume database successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					// Create MySQL object again to resume it
					By("Create MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					my, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(my.Spec.Init).NotTo(BeNil())

					By("Checking MySQL crd does not have kubedb.com/initialized annotation")
					_, err = meta_util.GetString(my.Annotations, api.AnnotationInitialized)
					Expect(err).To(HaveOccurred())
				})
			})

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

				It("should resume database successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")

						By("Delete mysql")
						err = f.DeleteMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for mysql to be deleted")
						f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

						// Create MySQL object again to resume it
						By("Create MySQL: " + mysql.Name)
						err = f.CreateMySQL(mysql)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for Running mysql")
						f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

						By("Checking Row Count of Table")
						f.EventuallyCountRow(mysql.ObjectMeta, dbName, 0).Should(Equal(3))

						my, err := f.GetMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
						Expect(my.Spec.Init).ShouldNot(BeNil())

						By("Checking MySQL crd does not have kubedb.com/initialized annotation")
						_, err = meta_util.GetString(my.Annotations, api.AnnotationInitialized)
						Expect(err).To(HaveOccurred())
					}
				})
			})
		})

		Context("Termination Policy", func() {

			Context("with TerminationDoNotTerminate", func() {
				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("MySQL is not halted. Check for mysql")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeTrue())

					By("Check for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Update mysql to set spec.terminationPolicy = Halt")
					_, err := f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.TerminationPolicy = api.TerminationPolicyHalt
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("with TerminationPolicyHalt", func() {

				It("should run successfully", func() {
					// Run MySQL and take snapshot
					shouldInsertData()

					By("Halt MySQL: Update mysql to set spec.halted = true")
					_, err := f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.Halted = true
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Wait for halted mysql")
					f.EventuallyMySQLPhase(mysql.ObjectMeta).Should(Equal(api.DatabasePhaseHalted))

					By("Resume MySQL: Update mysql to set spec.halted = false")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.Halted = false
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Running mysql")
					f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

					By("Deleting MySQL crd")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVC hasn't been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(1))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))

					// Create MySQL object again to resume it
					By("Create (resume) MySQL: " + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).NotTo(HaveOccurred())

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

				It("should run successfully", func() {
					// Run MySQL and take snapshot
					shouldInsertData()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for mysql to be deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVC has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(mysql.ObjectMeta).Should(Equal(1))
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					mysql.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should not create database and should wipeOut all", func() {
					// Run MySQL
					shouldInsertData()

					By("Delete mysql")
					err = f.DeleteMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until mysql is deleted")
					f.EventuallyMySQL(mysql.ObjectMeta).Should(BeFalse())

					By("Checking PVCs has been deleted")
					f.EventuallyPVCCount(mysql.ObjectMeta).Should(Equal(0))

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
					_, _, err = util.PatchMySQL(context.TODO(), f.ExtClient().KubedbV1alpha1(), mysql, func(in *api.MySQL) *api.MySQL {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  MYSQL_DATABASE,
								Value: "patched-db",
							},
						}
						return in
					}, metav1.PatchOptions{})
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

				Context("With TerminationPolicyHalt", func() {

					BeforeEach(func() {
						mysql.Spec.StorageType = api.StorageTypeEphemeral
						mysql.Spec.Storage = nil
						mysql.Spec.TerminationPolicy = api.TerminationPolicyHalt
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
