package framework

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) EventuallyONLINEMembersCount(meta metav1.ObjectMeta, dbName string, clientPodIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() int {
			tunnel, err := f.forwardPort(meta, clientPodIndex)
			if err != nil {
				return -1
			}
			defer tunnel.Close()

			en, err := f.getMySQLClient(meta, tunnel, dbName)
			if err != nil {
				return -1
			}
			defer en.Close()

			if err := en.Ping(); err != nil {
				return -1
			}

			var cnt int
			_, err = en.SQL("select count(MEMBER_STATE) from performance_schema.replication_group_members where MEMBER_STATE = ?", "ONLINE").Get(&cnt)
			if err != nil {
				return -1
			}
			return cnt
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) EventuallyCreateDatabase(meta metav1.ObjectMeta, dbName string) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			tunnel, err := f.forwardPort(meta, 0)
			if err != nil {
				return false
			}
			defer tunnel.Close()

			en, err := f.getMySQLClient(meta, tunnel, dbName)
			if err != nil {
				return false
			}
			defer en.Close()

			if err := en.Ping(); err != nil {
				return false
			}

			_, err = en.Exec("CREATE DATABASE kubedb")
			if err != nil {
				return false
			}
			return true
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) InsertRowFromSecondary(meta metav1.ObjectMeta, dbName string, clientPodIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			tunnel, err := f.forwardPort(meta, 1)
			if err != nil {
				return true
			}
			defer tunnel.Close()

			en, err := f.getMySQLClient(meta, tunnel, dbName)
			if err != nil {
				return true
			}
			defer func() {
				err = en.Close()
				return
			}()
			if err != nil {
				return true
			}

			if err := en.Ping(); err != nil {
				return true
			}

			if _, err := en.Insert(&KubedbTable{
				Name: fmt.Sprintf("%s-%d", meta.Name, clientPodIndex),
			}); err != nil {
				return false
			}

			return true
		},
		time.Minute*10,
		time.Second*10,
	)
}

func (f *Framework) GetPrimaryHostIndex(meta metav1.ObjectMeta, dbName string, clientPodIndex int) int {
	tunnel, err := f.forwardPort(meta, clientPodIndex)
	if err != nil {
		return -1
	}
	defer tunnel.Close()

	en, err := f.getMySQLClient(meta, tunnel, dbName)
	if err != nil {
		return -1
	}
	defer en.Close()

	if err := en.Ping(); err != nil {
		return -1
	}

	var row struct {
		Variable_name string
		Value         string
	}
	_, err = en.SQL("show status like \"%%primary%%\"").Get(&row)
	if err != nil {
		return -1
	}

	r, err2 := en.QueryString("select MEMBER_HOST from performance_schema.replication_group_members where MEMBER_ID = ?", row.Value)
	if err2 != nil || len(r) == 0 {
		return -1
	}

	idx, _ := strconv.Atoi(string(r[0]["MEMBER_HOST"][len(meta.Name)+1]))

	return idx
}

func (f *Framework) EventuallyGetPrimaryHostIndex(meta metav1.ObjectMeta, dbName string, clientPodIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() int {
			return f.GetPrimaryHostIndex(meta, dbName, clientPodIndex)
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) RemoverPrimaryToFailover(meta metav1.ObjectMeta, primaryPodIndex int) error {
	if _, err := f.kubeClient.CoreV1().Pods(meta.Namespace).Get(
		fmt.Sprintf("%s-%d", meta.Name, primaryPodIndex),
		metav1.GetOptions{},
	); err != nil {
		return err
	}

	return f.kubeClient.CoreV1().Pods(meta.Namespace).Delete(
		fmt.Sprintf("%s-%d", meta.Name, primaryPodIndex),
		&metav1.DeleteOptions{},
	)
}
