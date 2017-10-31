package e2e_test

import (
	"os"

	"github.com/appscode/go/types"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/mysql/test/e2e/framework"
	"github.com/k8sdb/mysql/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"fmt"
)

const (
	S3_BUCKET_NAME       = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME      = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME = "SWIFT_CONTAINER_NAME"
)

var _ = Describe("MySQL", func() {
	var (
		err         error
		f           *framework.Invocation
		mysql       *tapi.MySQL
		snapshot    *tapi.Snapshot
		secret      *apiv1.Secret
		skipMessage string
	)

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQL()
		snapshot = f.Snapshot()
		skipMessage = ""
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
		_, err := f.TryPatchDormantDatabase(mysql.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
			in.Spec.WipeOut = true
			return in
		})
		Expect(err).NotTo(HaveOccurred())

		By("Wait for mysql to be wipedOut")
		f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HaveWipedOut())

		err = f.DeleteDormantDatabase(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())
	}

	var shouldSuccessfullyRunning = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}

		// Create MySQL
		createAndWaitForRunning()

		// Delete test resource
		deleteTestResource()
	}

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", shouldSuccessfullyRunning)
			})

			Context("With PVC", func() {
				BeforeEach(func() {
					// set f.storage from cli flag. Example:
					// ginkgo test/e2e/ -- -storageclass="standard"
					if f.StorageClass == "" {
						skipMessage = "Missing StorageClassName. Provide as flag to test this."
					}
					mysql.Spec.Storage = &apiv1.PersistentVolumeClaimSpec{
						Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{
								apiv1.ResourceStorage: resource.MustParse("50Mi"),
							},
						},
						StorageClassName: types.StringP(f.StorageClass),
					}
				})
				It("should run successfully", shouldSuccessfullyRunning)
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
				Expect(err).NotTo(HaveOccurred())

				By("MySQL is not paused. Check for mysql")
				f.EventuallyMySQL(mysql.ObjectMeta).Should(BeTrue())

				By("Check for Running mysql")
				f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

				By("Update mysql to set DoNotPause=false")
				f.TryPatchMySQL(mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
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
				f.CreateSecret(secret)

				By("Create Snapshot")
				f.CreateSnapshot(snapshot)

				By("Check for Successed snapshot")
				f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(tapi.SnapshotPhaseSuccessed))

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
					snapshot.Spec.Local = &tapi.LocalSpec{
						Path: "/repo",
						VolumeSource: apiv1.VolumeSource{
							HostPath: &apiv1.HostPathVolumeSource{
								Path: "/repo",
							},
						},
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)

				// Additional
				Context("With PVC", func() {
					BeforeEach(func() {
						// set f.storage from cli flag. Example:
						// ginkgo test/e2e/ -- -storageclass="standard"
						if f.StorageClass == "" {
							skipMessage = "Missing StorageClassName. Provide as flag to test this."
						}
						mysql.Spec.Storage = &apiv1.PersistentVolumeClaimSpec{
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceStorage: resource.MustParse("5Gi"),
								},
							},
							StorageClassName: types.StringP(f.StorageClass),
						}
					})
					It("should run successfully", shouldTakeSnapshot)
				})
			})

			Context("In S3", func() {
				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &tapi.S3Spec{
						Bucket: os.Getenv(S3_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &tapi.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Azure", func() {
				BeforeEach(func() {
					secret = f.SecretForAzureBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Azure = &tapi.AzureSpec{
						Container: os.Getenv(AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &tapi.SwiftSpec{
						Container: os.Getenv(SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldTakeSnapshot)
			})
		})

		Context("Initialize", func() {
			Context("With Script", func() {
				BeforeEach(func() {
					mysql.Spec.Init = &tapi.InitSpec{
						ScriptSource: &tapi.ScriptSourceSpec{
							VolumeSource: apiv1.VolumeSource{
								GitRepo: &apiv1.GitRepoVolumeSource{
									Repository: "https://github.com/the-redback/mysql-init-script.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", shouldSuccessfullyRunning)
			})

			Context("With Snapshot", func() {
				AfterEach(func() {
					f.DeleteSecret(secret.ObjectMeta)
				})

				var shouldRestoreSnapshot = func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Create Snapshot")
					f.CreateSnapshot(snapshot)

					By("Check for Successed snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(tapi.SnapshotPhaseSuccessed))

					By("Check for snapshot data")
					f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())

					oldMySQL, err := f.GetMySQL(mysql.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Create mysql from snapshot")
					mysql = f.MySQL()
					mysql.Spec.Init = &tapi.InitSpec{
						SnapshotSource: &tapi.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					// Create and wait for running MySQL
					createAndWaitForRunning()

					// Delete test resource
					deleteTestResource()
					mysql = oldMySQL
					// Delete test resource
					deleteTestResource()
				}

				Context("with S3", func() {
					BeforeEach(func() {
						secret = f.SecretForS3Backend()
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.S3 = &tapi.S3Spec{
							Bucket: os.Getenv(S3_BUCKET_NAME),
						}
						snapshot.Spec.DatabaseName = mysql.Name
					})

					It("should run successfully", shouldRestoreSnapshot)
				})

				Context("with GCS", func() {
					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.GCS = &tapi.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}
						snapshot.Spec.DatabaseName = mysql.Name
					})

					It("should run successfully", shouldRestoreSnapshot)
				})
			})

		})

		Context("Resume", func() {
			var usedInitSpec bool
			BeforeEach(func() {
				usedInitSpec = false
			})

			var shouldResumeSuccessfully = func() {
				// Create and wait for running MySQL
				createAndWaitForRunning()

				By("Delete mysql")
				f.DeleteMySQL(mysql.ObjectMeta)

				By("Wait for mysql to be paused")
				f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

				_, err = f.TryPatchDormantDatabase(mysql.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
					in.Spec.Resume = true
					return in
				})
				Expect(err).NotTo(HaveOccurred())

				By("Wait for DormantDatabase to be deleted")
				f.EventuallyDormantDatabase(mysql.ObjectMeta).Should(BeFalse())

				By("Wait for Running mysql")
				f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

				mysql, err = f.GetMySQL(mysql.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				if usedInitSpec {
					Expect(mysql.Spec.Init).Should(BeNil())
					Expect(mysql.Annotations[tapi.MySQLInitSpec]).ShouldNot(BeEmpty())
				}

				// Delete test resource
				deleteTestResource()
			}

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("With Init", func() {
				BeforeEach(func() {
					usedInitSpec = true
					mysql.Spec.Init = &tapi.InitSpec{
						ScriptSource: &tapi.ScriptSourceSpec{
							VolumeSource: apiv1.VolumeSource{
								GitRepo: &apiv1.GitRepoVolumeSource{
									Repository: "https://github.com/the-redback/mysql-init-script.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("With original MySQL", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Delete mysql")
					f.DeleteMySQL(mysql.ObjectMeta)

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

					// Delete test resource
					deleteTestResource()
				})
				Context("with init", func() {
					BeforeEach(func() {
						usedInitSpec = true
						mysql.Spec.Init = &tapi.InitSpec{
							ScriptSource: &tapi.ScriptSourceSpec{
								VolumeSource: apiv1.VolumeSource{
									GitRepo: &apiv1.GitRepoVolumeSource{
										Repository: "https://github.com/the-redback/mysql-init-script.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					FIt("should resume DormantDatabase successfully", func() {
						// Create and wait for running MySQL
						createAndWaitForRunning()

						for i := 0; i < 3; i++ {
							By(">>>>>>>>>>>>>> "+fmt.Sprintf("%v", i) + " times running <<<<<<<<<<<")
							By("Delete mysql")
							f.DeleteMySQL(mysql.ObjectMeta)

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
						}

						// Delete test resource
						deleteTestResource()
					})
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

					By("Count multiple Snapshot")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					deleteTestResource()
				}

				Context("with local", func() {
					BeforeEach(func() {
						secret = f.SecretForLocalBackend()
						mysql.Spec.BackupSchedule = &tapi.BackupScheduleSpec{
							CronExpression: "@every 1m",
							SnapshotStorageSpec: tapi.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								Local: &tapi.LocalSpec{
									Path: "/repo",
									VolumeSource: apiv1.VolumeSource{
										HostPath: &apiv1.HostPathVolumeSource{
											Path: "/repo",
										},
									},
								},
							},
						}
					})

					It("should run schedular successfully", shouldStartupSchedular)
				})

				Context("with GCS and PVC", func() {
					BeforeEach(func() {
						secret = f.SecretForGCSBackend()
						mysql.Spec.BackupSchedule = &tapi.BackupScheduleSpec{
							CronExpression: "@every 1m",
							SnapshotStorageSpec: tapi.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								GCS: &tapi.GCSSpec{
									Bucket: os.Getenv(GCS_BUCKET_NAME),
								},
							},
						}
						if f.StorageClass == "" {
							skipMessage = "Missing StorageClassName. Provide as flag to test this."
						}
						mysql.Spec.Storage = &apiv1.PersistentVolumeClaimSpec{
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceStorage: resource.MustParse("50Mi"),
								},
							},
							StorageClassName: types.StringP(f.StorageClass),
						}
					})

					It("should run schedular successfully", shouldStartupSchedular)
				})
			})

			Context("With Update", func() {
				BeforeEach(func() {
					secret = f.SecretForLocalBackend()
				})
				It("should run schedular successfully", func() {
					// Create and wait for running MySQL
					createAndWaitForRunning()

					By("Create Secret")
					f.CreateSecret(secret)

					By("Update mysql")
					_, err = f.TryPatchMySQL(mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
						in.Spec.BackupSchedule = &tapi.BackupScheduleSpec{
							CronExpression: "@every 1m",
							SnapshotStorageSpec: tapi.SnapshotStorageSpec{
								StorageSecretName: secret.Name,
								Local: &tapi.LocalSpec{
									Path: "/repo",
									VolumeSource: apiv1.VolumeSource{
										HostPath: &apiv1.HostPathVolumeSource{
											Path: "/repo",
										},
									},
								},
							},
						}
						return in
					})
					Expect(err).NotTo(HaveOccurred())

					By("Count multiple Snapshot")
					f.EventuallySnapshotCount(mysql.ObjectMeta).Should(matcher.MoreThan(3))

					deleteTestResource()
				})
			})
		})

	})
})
