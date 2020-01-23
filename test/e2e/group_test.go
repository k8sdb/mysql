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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/mysql/test/e2e/framework"
	"kubedb.dev/mysql/test/e2e/matcher"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
)

var _ = Describe("MySQL Group Replication Tests", func() {
	var (
		err          error
		f            *framework.Invocation
		mysql        *api.MySQL
		garbageMySQL *api.MySQLList
		//skipMessage string
		dbName       string
		dbNameKubedb string
	)

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
	var writeOnPrimary = func(primaryPodIndex int) {
		By(fmt.Sprintf("Write on primary '%s-%d'", mysql.Name, primaryPodIndex))
		f.EventuallyCreateDatabase(mysql.ObjectMeta, dbName).Should(BeTrue())
		f.EventuallyCreateTable(mysql.ObjectMeta, dbNameKubedb).Should(BeTrue())
		rowCnt := 1
		f.EventuallyInsertRow(mysql.ObjectMeta, dbNameKubedb, primaryPodIndex, rowCnt).Should(BeTrue())
		f.EventuallyCountRow(mysql.ObjectMeta, dbNameKubedb, primaryPodIndex).Should(Equal(rowCnt))
	}
	var CheckDBVersionForGroupReplication = func() {
		if framework.DBCatalogName != "5.7.25" && framework.DBCatalogName != "5.7-v2" {
			Skip("For group replication CheckDBVersionForGroupReplication, DB version must be one of '5.7.25' or '5.7-v2'")
		}
	}

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQLGroup()
		garbageMySQL = new(api.MySQLList)
		//skipMessage = ""
		dbName = "mysql"
		dbNameKubedb = "kubedb"

		CheckDBVersionForGroupReplication()
	})

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			f.PrintDebugHelpers(mysql.Name, int(*mysql.Spec.Replicas))
		}
	})

	Context("Behaviour tests", func() {
		BeforeEach(func() {
			createAndWaitForRunning()
		})

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

		It("should be possible to create a basic 3 member group", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			writeOnPrimary(0)
			rowCnt := 1
			primaryPodIndex := 0
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				if i == primaryPodIndex {
					continue
				}

				By(fmt.Sprintf("Write on secondary '%s-%d'", mysql.Name, i))
				f.InsertRowFromSecondary(mysql.ObjectMeta, dbNameKubedb, i).Should(BeFalse())

				By(fmt.Sprintf("Read from secondary '%s-%d'", mysql.Name, i))
				f.EventuallyCountRow(mysql.ObjectMeta, dbNameKubedb, i).Should(Equal(rowCnt))
			}
		})

		It("should failover successfully", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			writeOnPrimary(0)

			By(fmt.Sprintf("Taking down the primary '%s-%d'", mysql.Name, 0))
			err = f.RemoverPrimaryToFailover(mysql.ObjectMeta, 0)
			Expect(err).NotTo(HaveOccurred())

			By("Checking status after failover")
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(
					Or(
						Equal(1),
						Equal(2),
					),
				)
			}

			By("Checking for data after failover")
			rowCnt := 1
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Read from '%s-%d'", mysql.Name, i))
				f.EventuallyCountRow(mysql.ObjectMeta, dbNameKubedb, i).Should(Equal(rowCnt))
			}
		})

		It("should be possible to scale up", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			By("Scaling up")
			mysql, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
				in.Spec.Replicas = types.Int32P(api.MySQLDefaultGroupSize + 1)

				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait for new member to be ready")
			Expect(f.WaitUntilPodRunningBySelector(mysql)).NotTo(HaveOccurred())

			By("Checking status after scaling up")
			for i := 0; i < api.MySQLDefaultGroupSize+1; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize + 1))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			writeOnPrimary(0)

			primaryPodIndex := 0
			rowCnt := 1
			for i := 0; i < api.MySQLDefaultGroupSize+1; i++ {
				if i == primaryPodIndex {
					continue
				}

				By(fmt.Sprintf("Write on secondary '%s-%d'", mysql.Name, i))
				f.InsertRowFromSecondary(mysql.ObjectMeta, dbNameKubedb, i).Should(BeFalse())

				By(fmt.Sprintf("Read from secondary '%s-%d'", mysql.Name, i))
				f.EventuallyCountRow(mysql.ObjectMeta, dbNameKubedb, i).Should(Equal(rowCnt))
			}
		})

		It("Should be possible to scale down", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			By("Scaling down")
			mysql, err = f.PatchMySQL(mysql.ObjectMeta, func(in *api.MySQL) *api.MySQL {
				in.Spec.Replicas = types.Int32P(api.MySQLDefaultGroupSize - 1)

				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for all member to be ready")
			Expect(f.WaitUntilPodRunningBySelector(mysql)).NotTo(HaveOccurred())

			By("Checking status after scaling down")
			for i := 0; i < api.MySQLDefaultGroupSize-1; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, dbName, i).Should(Equal(api.MySQLDefaultGroupSize - 1))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, dbName, i).Should(Equal(0))
			}

			writeOnPrimary(0)

			primaryPodIndex := 0
			rowCnt := 1
			for i := 0; i < api.MySQLDefaultGroupSize-1; i++ {
				if i == primaryPodIndex {
					continue
				}

				By(fmt.Sprintf("Write on secondary '%s-%d'", mysql.Name, i))
				f.InsertRowFromSecondary(mysql.ObjectMeta, dbNameKubedb, i).Should(BeFalse())

				By(fmt.Sprintf("Read from secondary '%s-%d'", mysql.Name, i))
				f.EventuallyCountRow(mysql.ObjectMeta, dbNameKubedb, i).Should(Equal(rowCnt))
			}
		})
	})

	Context("PDB", func() {

		It("should run evictions successfully", func() {
			// Create MySQL
			By("Create and run MySQL Group with three replicas")
			createAndWaitForRunning()
			//Evict MySQL pods
			By("Try to evict pods")
			err := f.EvictPodsFromStatefulSet(mysql.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
