package reactor

import (
	"fmt"
)

func ExampleTrigger() {
	var trigger Trigger

	trigger.AddReadCallback(func(v ...interface{}) {
		fmt.Printf("Value read: %v\n", v[0])
	})
	trigger.AddWriteCallback(func(v ...interface{}) {
		fmt.Printf("Value written: %v, Previous value: %v\n", v[1], v[0])
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
