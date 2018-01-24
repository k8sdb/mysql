package framework

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/appscode/kutil/tools/portforward"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubedbTable struct {
	Id   int64
	Name string
}

func (f *Framework) GetMySQLClient(meta metav1.ObjectMeta) (*xorm.Engine, error) {
	mysql, err := f.GetMySQL(meta)
	if err != nil {
		return nil, err
	}
	clientPodName := fmt.Sprintf("%v-0", mysql.Name)
	tunnel := portforward.NewTunnel(
		f.kubeClient.CoreV1().RESTClient(),
		f.restConfig,
		mysql.Namespace,
		clientPodName,
		3306,
	)

	if err := tunnel.ForwardPort(); err != nil {
		return nil, err
	}
	pass, err := f.GetMySQLRootPassword(mysql)

	cnnstr := fmt.Sprintf("root:%v@tcp(127.0.0.1:%v)/mysql", pass, tunnel.Local)
	return xorm.NewEngine("mysql", cnnstr)
}

var randChars = []rune("abcdefghijklmnopqrstuvwxyzabcdef")

// Use this for generating random pat of a ID. Do not use this for generating short passwords or secrets.
func characters(len int) string {
	bytes := make([]byte, len)
	rand.Read(bytes)
	r := make([]rune, len)
	for i, b := range bytes {
		r[i] = randChars[b>>3]
	}
	return string(r)
}

func (f *Framework) EventuallyCreateTable(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			en, err := f.GetMySQLClient(meta)
			if err != nil {
				return false
			}

			if err := en.Ping(); err != nil {
				return false
			}

			err = en.Sync(new(KubedbTable))
			if err != nil {
				fmt.Println("creation error", err)
				return false
			}
			return true
		},
		time.Minute*15,
		time.Second*10,
	)
	return nil
}

func (f *Framework) EventuallyInsertRow(meta metav1.ObjectMeta, total int) GomegaAsyncAssertion {
	count := 0
	return Eventually(
		func() bool {
			en, err := f.GetMySQLClient(meta)
			if err != nil {
				return false
			}

			if err := en.Ping(); err != nil {
				return false
			}

			for i := count; i < total; i++ {
				if _, err := en.Insert(&KubedbTable{
					Name: characters(5),
				}); err != nil {
					return false
				}
				count++
			}
			return true
		},
		time.Minute*15,
		time.Second*10,
	)
	return nil
}

func (f *Framework) EventuallyCountRow(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() int {
			en, err := f.GetMySQLClient(meta)
			if err != nil {
				return -1
			}

			if err := en.Ping(); err != nil {
				return -1
			}

			kubedb := new(KubedbTable)
			total, err := en.Count(kubedb)
			if err != nil {
				return -1
			}

			return int(total)
		},
		time.Minute*15,
		time.Second*10,
	)
}
