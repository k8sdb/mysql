package oneliners

import (
	"fmt"
	"log"
	"runtime"
	"encoding/json"
	"strings"
	"reflect"
)

func FILE(a ...interface{}) {
	_, file, ln, ok := runtime.Caller(1)
	if ok {
		fmt.Println("__FILE__", file, "__LINE__", ln)
		if len(a) > 0 {
			fmt.Println(a...)
		}
	} else {
		log.Fatal("Failed to detect runtime caller info.")
	}
}


func PrettyJson(a interface{},msg ...string) {
	_, file, ln, ok := runtime.Caller(1)
	if ok {
		fmt.Println("__FILE__", file, "__LINE__", ln)
		if a!=nil {
			if reflect.TypeOf(a).String() == "[]uint8" {
				var js interface{}
				if err := json.Unmarshal(a.([]byte), &js); err == nil {
					a = js
				}
			}
			data,_:=json.MarshalIndent(a,"","   ")
			str:=""
			if len(msg) >0 {
				str = strings.Trim(fmt.Sprintf("%v",msg), "[]")
				str = fmt.Sprintf("[ %s ]",str)
			}
			fmt.Printf("============================%s============================\n",str)
			fmt.Println(string(data))
		}
	} else {
		log.Fatal("Failed to detect runtime caller info.")
	}
}