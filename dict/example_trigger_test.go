package dict

import (
	"fmt"
)

func ExampleTrigger() {
	var trigger Trigger

	readCallback := KeyReadCallback(func(k, v interface{}) {
		fmt.Printf("Value read for key %v: %v\n", k, v)
	})

	writeCallback := func(pKey, pVal, k, v interface{}) {
		fmt.Printf("Value written for key %v: %v, Previous value: %v\n", k, v, pVal)
	}

	trigger.AddKeyReadCallback(readCallback)
	trigger.AddKeyWriteCallback(KeyWriteCallback(writeCallback))

	trigger.SetValue(initMap())

	trigger.Get(1)
	trigger.Set(1, "something else")
	trigger.Get(1)
	got,ok := trigger.GetCheck("Something that doesn't exist")
	fmt.Println(got, ok)
	// Output:
	// Value read for key 1: one
	// Value written for key 1: something else, Previous value: one
	// Value read for key 1: something else
	// Value read for key Something that doesn't exist: <nil>
	// <nil> false
}
