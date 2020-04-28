package reactor

import (
	"fmt"
)

func ExampleTrigger() {
	var trigger Trigger

	trigger.AddReadCallback(func(v interface{}) {
		fmt.Printf("Value read: %v\n", v)
	})
	trigger.AddWriteCallback(func(prev, v interface{}) {
		fmt.Printf("Value written: %v, Previous value: %v\n", v, prev)
	})

	trigger.Value()
	trigger.SetValue("world")
	trigger.SetValue("hello")
	trigger.Value()
	// Output:
	// Value read: <nil>
	// Value written: world, Previous value: <nil>
	// Value written: hello, Previous value: world
	// Value read: hello
}
