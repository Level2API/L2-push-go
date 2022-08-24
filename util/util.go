package util

import "fmt"

func AssertNilErr(err error,format string,args ...interface{}){
	if err!=nil{
		panic(fmt.Errorf(format+`->%v`,append(args,err)...))
	}
}
