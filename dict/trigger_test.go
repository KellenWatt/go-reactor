package dict

import (
	"testing"
	"reflect"
	
	"github.com/KellenWatt/reactor"
)

func initMap() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	m[1] = "one"
	m[2] = "two"
	m[3] = "three"
	m[4] = "four"
	m[5] = "five"
	m[6] = "six"
	m[7] = "seven"
	m[8] = "eight"
	m[9] = "nine"
	m[11] = "eleven"

	return m
}

func TestTriggerSetValue(t *testing.T) {
	var trigger Trigger
	want := initMap()

	trigger.SetValue(want)

	got := trigger.value
	resgot := trigger.Value()

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Direct value comparison: Expected %v; got %v", want, got)
	}

	if !reflect.DeepEqual(want, resgot) {
		t.Fatalf("Value() comparison: Expected %v; got %v", want, resgot)
	}
}

func TestTriggerSetValueTypeCheck(t *testing.T) {
	var trigger Trigger
	
	defer func() {
		if recover() == nil {
			t.Fatal("Expected value when assigning non-[]interface{} value")
		}
	}()

	trigger.SetValue(10)
}

func TestTriggerSetValueMapTypeCheck(t *testing.T) {
	var trigger Trigger
	
	defer func() {
		if recover() == nil {
			t.Fatal("Expected value when assigning non-interface{} map value")
		}
	}()

	m := make(map[interface{}]int)
	trigger.SetValue(m)
}


func TestTriggerIndependentSetValue(t *testing.T) {
	var trigger Trigger
	src := initMap()
	want := initMap()

	trigger.SetValue(src)

	src[3] = 0

	if reflect.DeepEqual(src, trigger.value) {
		t.Fatalf("Changes to source should not affect internal: Expected %v; got %v", want, trigger.value)
	}
}

func TestTriggerIndependentValueResult(t *testing.T) {
	var trigger Trigger
	want := initMap()

	trigger.SetValue(want)
	val := trigger.Value().(map[interface{}]interface{})

	val[3] = 0

	if !reflect.DeepEqual(trigger.value, want) {
		t.Fatalf("Changes to result of value should not affect internal: Expected %v; got %v", 
				 want, trigger.value)
	}
}


func TestTriggerAddBinder(t *testing.T) {
	var trigger Trigger
	var ind reactor.Indicator

	trigger.AddBinder(&ind, reactor.TrivialBinding, false)

	if len(trigger.bindings) != 1 {
		t.Fatalf("Binding not added")
	}
}

func TestTriggerReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	callback := func(v interface{}) {
		count += 1
	}

	trigger.AddReadCallback(callback)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.Value()
	}

	if count != iters {
		t.Fatalf("ReadCallbacks not being run. Of %d iterations, ran %d times; expected %d", iters, count, iters)
	}
}


// Fix
func TestTriggerWriteCallback(t *testing.T) {
	var trigger Trigger
	var count int
	callback := func(prev, v interface{}) {
		prevVal := prev.(map[interface{}]interface{})
		val := v.(map[interface{}]interface{})
		count += 1

		// Poor test, not terribly generic, but it does validate.
		if prevVal != nil && prevVal[count-2] != val[count-1].(int) - 1 { 
			t.Fatalf("prev not being set correctly. Expected %v, got %v", 
			         map[interface{}]interface{}{count-2: count-2}, prevVal)
		}
	}

	trigger.AddWriteCallback(callback)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue(map[interface{}]interface{}{i: i})
	}

	if count != iters {
		t.Fatalf("ReadCallbacks not being run. Of %d iterations, ran %d times; expected %d", iters, count, iters)
	}
}

