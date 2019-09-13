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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MySQL Group Replication Tests", func() {
	var (
		err          error
		f            *framework.Invocation
		mysql        *api.MySQL
		garbageMySQL *api.MySQLList
		dbName       string
		dbNameKubedb string

		proxysql bool
		psql     *api.ProxySQL
	)

	var createAndWaitForRunningMySQL = func() {
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
		f.EventuallyDatabaseReady(mysql.ObjectMeta, proxysql, dbName, 0).Should(BeTrue())
	}

	var deleteMySQLResource = func() {
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

	var deleteProxySQLResource = func() {
		if psql == nil {
			log.Infoln("Skipping cleanup. Reason: ProxySQL object is nil")
			return
		}
		By("Check if ProxySQL " + psql.Name + " exists.")
		_, err = f.GetProxySQL(psql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// ProxySQL was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}
		By("Delete ProxySQL")
		err = f.DeleteProxySQL(psql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: ProxySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}
	}

	var deleteTestResource = func() {
		deleteMySQLResource()
		deleteProxySQLResource()
	}

	var deleteLeftOverStuffs = func() {
		// old MySQL are in garbageMySQL list. delete their resources.
		for _, my := range garbageMySQL.Items {
			*mysql = my
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers()
	}

	var createAndWaitForRunningProxySQL = func() {
		By("Create ProxySQL: " + psql.Name)
		err = f.CreateProxySQL(psql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running ProxySQL")
		f.EventuallyProxySQLPhase(psql.ObjectMeta).Should(Equal(api.DatabasePhaseRunning))
	}

	var countRows = func(meta metav1.ObjectMeta, podIndex, expectedRowCnt int) {
		By(fmt.Sprintf("Read row from member '%s-%d'", meta.Name, podIndex))
		f.EventuallyCountRow(meta, proxysql, dbNameKubedb, podIndex).Should(Equal(expectedRowCnt))
	}

	var insertRows = func(meta metav1.ObjectMeta, podIndex, rowCntToInsert int, expected bool) {
		By(fmt.Sprintf("Insert row on member '%s-%d' should be %v", meta.Name, podIndex, expected))
		if expected {
			f.EventuallyInsertRow(meta, proxysql, dbNameKubedb, podIndex, rowCntToInsert).Should(BeTrue())
		} else {
			f.EventuallyInsertRow(meta, proxysql, dbNameKubedb, podIndex, rowCntToInsert).Should(BeFalse())
		}
	}

	var create_Database_N_Table = func(meta metav1.ObjectMeta, podIndex int) {
		By("Create Database")
		f.EventuallyCreateDatabase(meta, proxysql, dbName, podIndex).Should(BeTrue())

		By("Create Table")
		f.EventuallyCreateTable(meta, proxysql, dbNameKubedb, podIndex).Should(BeTrue())
	}

	var writeToPrimary = func(meta metav1.ObjectMeta, podIndex int) {
		By(fmt.Sprintf("Write on '%s-%d'", meta.Name, podIndex))
		insertRows(meta, podIndex, 1, true)
	}

	var readFromEachMember = func(meta metav1.ObjectMeta, clusterSize, rowCnt int) {
		for j := 0; j < clusterSize; j += 1 {
			countRows(meta, j, rowCnt)
		}
	}

	var writeTo_Primary_N_ReadFrom_EachMember = func(meta metav1.ObjectMeta, primaryPodIndex, clusterSize int) {
		writeToPrimary(meta, primaryPodIndex)
		readFromEachMember(meta, clusterSize, 1)
	}

	var replicationCheck = func(meta metav1.ObjectMeta, primaryPodIndex, clusterSize int) {
		By("Checking replication")
		create_Database_N_Table(meta, primaryPodIndex)
		writeTo_Primary_N_ReadFrom_EachMember(meta, primaryPodIndex, clusterSize)
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
		dbName = "mysql"
		dbNameKubedb = "kubedb"
		proxysql = false

		CheckDBVersionForGroupReplication()
	})

	Context("Behaviour tests", func() {
		BeforeEach(func() {
			createAndWaitForRunningMySQL()
		})

		AfterEach(func() {
			// delete resources for current MySQL
			deleteTestResource()
			deleteLeftOverStuffs()
		})

		It("should be possible to create a basic 3 member group", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}

			primaryPodIndex := f.GetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, 0)
			replicationCheck(mysql.ObjectMeta, primaryPodIndex, api.MySQLDefaultGroupSize)
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				if i == primaryPodIndex {
					continue
				}

				insertRows(mysql.ObjectMeta, i, 1, false)
				countRows(mysql.ObjectMeta, i, 1)
			}
		})

		It("should failover successfully", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}

			replicationCheck(mysql.ObjectMeta, f.GetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, 0), api.MySQLDefaultGroupSize)

			By(fmt.Sprintf("Taking down the primary '%s-%d'", mysql.Name, 0))
			err = f.RemovePrimaryToFailover(mysql.ObjectMeta, f.GetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, 0))
			Expect(err).NotTo(HaveOccurred())

			By("Checking status after failover")
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(
					Or(
						Equal(1),
						Equal(2),
					),
				)
			}

			By("Checking for data after failover")
			readFromEachMember(mysql.ObjectMeta, api.MySQLDefaultGroupSize, 1)
		})

		It("should be possible to scale up", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
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
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize + 1))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}

			primaryPodIndex := f.GetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, 0)
			replicationCheck(mysql.ObjectMeta, primaryPodIndex, api.MySQLDefaultGroupSize+1)
			for i := 0; i < api.MySQLDefaultGroupSize+1; i++ {
				if i == primaryPodIndex {
					continue
				}

				insertRows(mysql.ObjectMeta, i, 1, false)
				countRows(mysql.ObjectMeta, i, 1)
			}
		})

		It("Should be possible to scale down", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
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
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize - 1))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}

			primaryPodIndex := f.GetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, 0)
			replicationCheck(mysql.ObjectMeta, primaryPodIndex, api.MySQLDefaultGroupSize-1)
			for i := 0; i < api.MySQLDefaultGroupSize-1; i++ {
				if i == primaryPodIndex {
					continue
				}

				insertRows(mysql.ObjectMeta, i, 1, false)
				countRows(mysql.ObjectMeta, i, 1)
			}
		})
	})

	Context("ProxySQL", func() {
		BeforeEach(func() {
			if !framework.ProxySQLTest {
				Skip("For ProxySQL test, the value of '--proxysql' flag must be 'true' while running e2e-tests command")
			}
			createAndWaitForRunningMySQL()

			psql = f.ProxySQL(mysql.Name)
			createAndWaitForRunningProxySQL()
		})

		AfterEach(func() {
			// delete resources for current MySQL
			deleteTestResource()
			deleteLeftOverStuffs()
		})

		It("should configure poxysql for backend servers", func() {
			for i := 0; i < api.MySQLDefaultGroupSize; i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyONLINEMembersCount(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", mysql.Name, i))
				f.EventuallyGetPrimaryHostIndex(mysql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}
			proxysql = true
			for i := 0; i < int(*psql.Spec.Replicas); i++ {
				By(fmt.Sprintf("Checking ONLINE member count from Pod '%s-%d'", psql.Name, i))
				f.EventuallyONLINEMembersCount(psql.ObjectMeta, proxysql, dbName, i).Should(Equal(api.MySQLDefaultGroupSize))

				By(fmt.Sprintf("Checking primary Pod index from Pod '%s-%d'", psql.Name, i))
				f.EventuallyGetPrimaryHostIndex(psql.ObjectMeta, proxysql, dbName, i).Should(Equal(0))
			}

			replicationCheck(psql.ObjectMeta, 0, int(*psql.Spec.Replicas))
			proxysql = false
			readFromEachMember(mysql.ObjectMeta, api.MySQLDefaultGroupSize, int(*psql.Spec.Replicas))
		})
	})

	Context("PDB", func() {

		It("should run evictions successfully", func() {
			// Create MySQL
			By("Create and run MySQL Group with three replicas")
			createAndWaitForRunningMySQL()
			//Evict MySQL pods
			By("Try to evict pods")
			err := f.EvictPodsFromStatefulSet(mysql.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
