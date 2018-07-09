package e2e_test

import (
	"fmt"
	"os"

	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/mysql/test/e2e/framework"
	"github.com/kubedb/mysql/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
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
		err         error
		f           *framework.Invocation
		mysql       *api.MySQL
		snapshot    *api.Snapshot
		secret      *core.Secret
		skipMessage string
		dbName      string
	)

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQL()
		snapshot = f.Snapshot()
		skipMessage = ""
		dbName = "mysql"
	})

	var createAndWaitForRunning = func() {
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())
	}

	var deleteTestResource = func() {
		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

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

		By("Wait for mysql resources to be wipedOut")
		f.EventuallyWipedOut(mysql.ObjectMeta).Should(Succeed())
	}

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}
					// Create MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Rows")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					deleteTestResource()
				})
			})
		})

		Context("DoNotPause", func() {
			BeforeEach(func() {
				mysql.Spec.DoNotPause = true
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

				By("Update mysql to set DoNotPause=false")
				f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
					in.Spec.DoNotPause = false
					return in
				})

				// Delete test resource
				deleteTestResource()
			})
		})

		Context("Snapshot", func() {
			var skipDataCheck bool

			AfterEach(func() {
				f.DeleteSecret(secret.ObjectMeta)
			})

			BeforeEach(func() {
				skipDataCheck = false
				snapshot.Spec.DatabaseName = mysql.Name
			})

			var shouldTakeSnapshot = func() {
				// Create and wait for running MySQL
				createAndWaitForRunning()

				By("Create Secret")
				err := f.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())

				By("Create Snapshot")
				err = f.CreateSnapshot(snapshot)
				Expect(err).NotTo(HaveOccurred())

				By("Check for Successed snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

				if !skipDataCheck {
					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
				}

				// Delete test resource
				deleteTestResource()

				if !skipDataCheck {
					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
				}
			}

			Context("In Local", func() {
				BeforeEach(func() {
					skipDataCheck = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Local = &api.LocalSpec{
						MountPath: "/repo",
						VolumeSource: core.VolumeSource{
							EmptyDir: &core.EmptyDirVolumeSource{},
						},
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In S3", func() {
				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &api.S3Spec{
						Bucket: os.Getenv(S3_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &api.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
				})

				Context("Without Init", func() {
					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("With Init", func() {
					BeforeEach(func() {
						mysql.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/mysql-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("Delete One Snapshot keeping others", func() {
					BeforeEach(func() {
						mysql.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://github.com/kubedb/mysql-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create and wait for running MySQL
						createAndWaitForRunning()

						By("Create Secret")
						err := f.CreateSecret(secret)
						Expect(err).NotTo(HaveOccurred())

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						oldSnapshot := snapshot

						// create new Snapshot
						snapshot := f.Snapshot()
						snapshot.Spec.DatabaseName = mysql.Name
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.GCS = &api.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						By(fmt.Sprintf("Delete Snapshot %v", snapshot.Name))
						err = f.DeleteSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for Deleting Snapshot")
						f.EventuallySnapshot(mysql.ObjectMeta).Should(BeFalse())
						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
						}

						snapshot = oldSnapshot

						By(fmt.Sprintf("Old Snapshot %v Still Exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataCheck {
							By(fmt.Sprintf("Check for old snapshot %v data", snapshot.Name))
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						// Delete test resource
						deleteTestResource()

						if !skipDataCheck {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
						}
					})
				})

			})

			Context("In Azure", func() {
				BeforeEach(func() {
					secret = f.SecretForAzureBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Azure = &api.AzureSpec{
						Container: os.Getenv(AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &api.SwiftSpec{
						Container: os.Getenv(SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})
		})

		Context("Initialize", func() {
			Context("With Script", func() {
				BeforeEach(func() {
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", func() {
					// Create MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					// Delete test resource
					deleteTestResource()
				})

			})

			Context("With Snapshot", func() {
				AfterEach(func() {
					f.DeleteSecret(secret.ObjectMeta)
				})

				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &api.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = mysql.Name
				})

				It("should run successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Create Secret")
					f.CreateSecret(secret)

					By("Create Snapshot")
					f.CreateSnapshot(snapshot)

					By("Check for Successed snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Create mysql from snapshot")
					mysql = f.MySQL()
					mysql.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating init Snapshot Mysql without secret name" + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret

					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					// Delete test resource
					deleteTestResource()
					mysql = oldMySQL
					// Delete test resource
					deleteTestResource()
				})
			})
		})

		Context("Resume", func() {
			var usedInitScript bool
			var usedInitSnapshot bool
			BeforeEach(func() {
				usedInitScript = false
				usedInitSnapshot = false
			})

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					_, err = f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// Delete test resource
					deleteTestResource()
				})
			})

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					mysql, err = f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// Delete test resource
					deleteTestResource()
				})
			})

			Context("with init Script", func() {
				BeforeEach(func() {
					usedInitScript = true
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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

					_, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					// Delete test resource
					deleteTestResource()
					if usedInitScript {
						Expect(mysql.Spec.Init).ShouldNot(BeNil())
						if usedInitScript {
							Expect(mysql.Spec.Init).ShouldNot(BeNil())
							_, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
							Expect(err).To(HaveOccurred())
						}
					}
				})
			})

			Context("With Snapshot Init", func() {
				var skipDataCheck bool
				AfterEach(func() {
					f.DeleteSecret(secret.ObjectMeta)
				})
				BeforeEach(func() {
					skipDataCheck = false
					usedInitSnapshot = true
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &api.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = mysql.Name
				})
				It("should resume successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Create Secret")
					f.CreateSecret(secret)

					By("Create Snapshot")
					f.CreateSnapshot(snapshot)

					By("Check for Successed snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Create mysql from snapshot")
					mysql = f.MySQL()
					mysql.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating init Snapshot Mysql without secret name" + mysql.Name)
					err = f.CreateMySQL(mysql)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					mysql.Spec.DatabaseSecret = oldMySQL.Spec.DatabaseSecret
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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

					mysql, err = f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					if usedInitSnapshot {
						Expect(mysql.Spec.Init).ShouldNot(BeNil())
						_, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
						Expect(err).NotTo(HaveOccurred())
					}

					// Delete test resource
					deleteTestResource()
					mysql = oldMySQL
					// Delete test resource
					deleteTestResource()
					if !skipDataCheck {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
					}
				})
			})

			Context("Multiple times with init", func() {
				BeforeEach(func() {
					usedInitScript = true
					mysql.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://github.com/kubedb/mysql-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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

						_, err := f.GetMySQL(mysql.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Checking Row Count of Table")
						f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

						if usedInitScript {
							Expect(mysql.Spec.Init).ShouldNot(BeNil())
							_, err := meta_util.GetString(mysql.Annotations, api.AnnotationInitialized)
							Expect(err).To(HaveOccurred())
						}
					}

					// Delete test resource
					deleteTestResource()
				})
			})
		})

		Context("SnapshotScheduler", func() {
			AfterEach(func() {
				f.DeleteSecret(secret.ObjectMeta)
			})

			Context("With Startup", func() {

				var shouldStartupSchedular = func() {
					By("Create Secret")
					f.CreateSecret(secret)

					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())

					deleteTestResource()
				}

				Context("with local", func() {
					BeforeEach(func() {
						secret = f.SecretForLocalBackend()
						mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
							SnapshotStorageSpec: api.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								Local: &api.LocalSpec{
									MountPath: "/repo",
									VolumeSource: core.VolumeSource{
										EmptyDir: &core.EmptyDirVolumeSource{},
									},
								},
							},
						}
					})

					It("should run schedular successfully", shouldStartupSchedular)
				})

				Context("with GCS", func() {
					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						mysql.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 1m",
							SnapshotStorageSpec: api.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								GCS: &api.GCSSpec{
									Bucket: os.Getenv(GCS_BUCKET_NAME),
								},
							},
						}
					})

					It("should run schedular successfully", shouldStartupSchedular)
				})
			})

			Context("With Update - with Local", func() {
				BeforeEach(func() {
					secret = f.SecretForLocalBackend()
				})
				It("should run schedular successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update mysql")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
							SnapshotStorageSpec: api.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								Local: &api.LocalSpec{
									MountPath: "/repo",
									VolumeSource: core.VolumeSource{
										EmptyDir: &core.EmptyDirVolumeSource{},
									},
								},
							},
						}
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())

					deleteTestResource()
				})
			})

			Context("Re-Use DormantDatabase's scheduler", func() {
				BeforeEach(func() {
					secret = f.SecretForLocalBackend()
				})
				It("should re-use scheduler successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update mysql")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = &api.BackupScheduleSpec{
							CronExpression: "@every 20s",
							SnapshotStorageSpec: api.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								Local: &api.LocalSpec{
									MountPath: "/repo",
									VolumeSource: core.VolumeSource{
										EmptyDir: &core.EmptyDirVolumeSource{},
									},
								},
							},
						}
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Count multiple Snapshot Object")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(5))

					By("Remove Backup Scheduler from MySQL")
					_, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
						in.Spec.BackupSchedule = nil
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Verify multiple Succeeded Snapshot")
					f.EventuallyMultipleSnapshotFinishedProcessing(mysql.ObjectMeta).Should(Succeed())

					deleteTestResource()
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
					mysql.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					// Create MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Rows")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					deleteTestResource()
				})
			})

			Context("Root Password as EnvVar", func() {
				It("should reject to create MySQL CRD", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					mysql.Spec.Env = []core.EnvVar{
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
				It("should reject to update EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					mysql.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					// Create MySQL
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(mysql.ObjectMeta, dbName).Should(BeTrue())

					By("Inserting Rows")
					f.EventuallyInsertRow(mysql.ObjectMeta, dbName, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

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
					f.EventuallyCountRow(mysql.ObjectMeta, dbName).Should(Equal(3))

					By("Patching EnvVar")
					_, _, err = util.PatchMySQL(f.ExtClient(), mysql, func(in *api.MySQL) *api.MySQL {
						in.Spec.Env = []core.EnvVar{
							{
								Name:  MYSQL_DATABASE,
								Value: "patched-db",
							},
						}
						return in
					})
					Expect(err).To(HaveOccurred())

					deleteTestResource()
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
					f.DeleteConfigMap(userConfig.ObjectMeta)
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
	})
})
