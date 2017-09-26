package v2alpha1

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/appscode/kutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	batchv2alpha1 "k8s.io/client-go/pkg/apis/batch/v2alpha1"
)

func GetGroupVersionKind(v interface{}) schema.GroupVersionKind {
	return batchv2alpha1.SchemeGroupVersion.WithKind(kutil.GetKind(v))
}

func AssignTypeKind(v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("%v must be a pointer", v)
	}

	switch u := v.(type) {
	case *batchv2alpha1.CronJob:
		u.APIVersion = batchv2alpha1.SchemeGroupVersion.String()
		u.Kind = kutil.GetKind(v)
		return nil
	}
	return errors.New("unknown api object type")
}
