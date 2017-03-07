//Dumps all errors into a CSV for easy digestion
package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/log"
)

func printHeader(f *os.File) {
	// Print out header
	fmt.Println("Error Group,Error Code, Error Description")
	f.Write([]byte("Error Group,Error Code, Error Description"))
}

func printGroup(f *os.File, groupName string, group interface{}) {
	v := reflect.ValueOf(group)
	for i := 0; i < v.NumField(); i++ {
		t := fmt.Sprintf("%s, %d, %s\n", groupName, v.Field(i).Interface().(consts.ErrCodes).Code, v.Field(i).Interface().(consts.ErrCodes).Description)
		if v.Field(i).Interface().(consts.ErrCodes).Code != 0 {
			fmt.Printf(t)
			f.Write([]byte(t))
		}
	}
}

func main() {
	// Open file
	f, err := os.Create("errors.txt")
	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
		return
	}
	defer f.Close()

	printHeader(f)

	// Range through each error group
	printGroup(f, "Generic Errors", consts.GenericErrors)
	printGroup(f, "Counterparty Errors", consts.CounterpartyErrors)
	printGroup(f, "Ripple Errors", consts.RippleErrors)
}
