package reactor

import (
	"fmt"
)


func ExampleIndicator() {
	var trigger Trigger
	var ind Indicator

	ind.AddBinding(&trigger, func(v interface{}) interface{} {
		return v.(int) * 2
	})

	trigger.AddReadCallback(func(v ...interface{}) {
		fmt.Printf("Trigger has value: %v\n", v[0])
	})

	ind.AddReadCallback(func(v ...interface{}) {
		fmt.Printf("Indicator has value: %v\n", v[0])
	})

	trigger.Value()
	ind.Value()
	trigger.SetValue(10)
	trigger.Value()
	ind.Value()
	trigger.SetValue(3)
	trigger.Value()
	ind.Value()
	// Output:
	// Trigger has value: <nil>
	// Indicator has value: <nil>
	// Trigger has value: 10
	// Indicator has value: 20
	// Trigger has value: 3
	// Indicator has value: 6
}

func ExampleIndicator_intToString() {
	var trigger Trigger
	var ind Indicator

	ind.AddBinding(&trigger, func(v interface{}) interface{} {
		return fmt.Sprint(v)
	})

	trigger.SetValue(10)

	fmt.Printf("%v: %T; %v %T", trigger.Value(), trigger.Value(), ind.Value(), ind.Value())
	// Output:
	// 10: int; 10 string
}

func ExampleIndicator_AddDelayedBinding() {
	var trigger Trigger
	var ind Indicator

	ind.AddDelayedBinding(&trigger, func(v interface{}) interface{} {
		return v
	})

	ind.AddWriteCallback(func(v ...interface{}) {
		fmt.Printf("Indicator: Previous value: %v, New value: %v\n", v[0], v[1])
	})

	trigger.AddWriteCallback(func(v ...interface{}) {
		fmt.Printf("Trigger: Previous value: %v, New value: %v\n", v[0], v[1])
	})

	for i:=0; i<5; i++ {
		trigger.SetValue(i)
	}
	
	ind.Value()
	// Output:
	// Trigger: Previous value: <nil>, New value: 0
	// Trigger: Previous value: 0, New value: 1
	// Trigger: Previous value: 1, New value: 2
	// Trigger: Previous value: 2, New value: 3
	// Trigger: Previous value: 3, New value: 4
	// Indicator: Previous value: <nil>, New value: 4
}

