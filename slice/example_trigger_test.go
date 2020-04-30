package slice

import (
	"fmt"
)

func ExampleTrigger() {
	var trigger Trigger

	readCallback := IndexReadCallback(func(i int, v interface{}) {
		fmt.Printf("Value read at index %d: %v\n", i, v)
	})

	writeCallback := func(prevI int, prev interface{}, i int, v interface{}) {
		fmt.Printf("Value written at index %d: %v, Previous value: %v\n", i, v, prev)
	}

	trigger.AddIndexReadCallback(readCallback)
	trigger.AddIndexWriteCallback(IndexWriteCallback(writeCallback))

	trigger.SetValue([]interface{}{1,2,3,4,5})

	trigger.At(1)
	trigger.SetAt(1, 10)
	trigger.At(1)
	_,err := trigger.At(5)
	fmt.Println(err)
	// Output:
	// Value read at index 1: 2
	// Value written at index 1: 10, Previous value: 2
	// Value read at index 1: 10
	// Index out of bounds
}
