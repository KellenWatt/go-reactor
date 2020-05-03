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

// Get
// No create

func TestTriggerGet(t *testing.T) {
	var trigger Trigger
	m := initMap()
	index := 3
	want := m[index]

	trigger.SetValue(m)
	got := trigger.Get(index)
	
	if got != want {
		t.Fatalf("Expected Get to return %v for key %v; got %v", want, index, got)
	}
}

func TestTriggerGetFromNil(t *testing.T) {
	var trigger Trigger

	trigger.Get("who cares?")
	if trigger.value == nil {
		t.Fatal("Expected unitialized trigger to exist after Get; got <nil>")
	}
}

func TestTriggerGetNil(t *testing.T) {
	var trigger Trigger

	trigger.SetValue(initMap())
	got := trigger.Get("who cares?")

	if got != nil {
		t.Fatalf("Expected invalid read to return zero value of interface{}; got %v", got)
	}
}

// GetCheck
// No create

func TestTriggerGetCheck(t *testing.T) {
	var trigger Trigger
	m := initMap()
	index := 5
	want := m[index]

	trigger.SetValue(m)
	got,ok := trigger.GetCheck(index)

	if !ok {
		t.Fatal("Expected valid read to return true; got false")
	}
	if got != want {
		t.Fatalf("Expected GetCheck to return %v for key %v; got %v", want, index, got)
	}

}

func TestTriggerGetCheckNil(t *testing.T) {
	var trigger Trigger
	m := initMap()
	index := "junk"
	want := m[index]

	got,ok := trigger.GetCheck(index)

	if ok {
		t.Fatal("Expected invalid read to return false; got true")
	}
	if got != want {
		t.Fatalf("Expected GetCheck to return <nil> for key %v; got %v", index, got)
	}
}

func TestTriggerGetCheckFromNil(t *testing.T) {
	var trigger Trigger

	_,ok := trigger.GetCheck("who cares?")
	if trigger.value == nil {
		t.Fatal("Expected unitialized trigger to exist after Get; got <nil>")
	}
	if ok {
		t.Fatal("Expected invalid read to return false; got true")
	}
}

// Set

func TestTriggerSet(t *testing.T) {
	var trigger Trigger
	m := initMap()
	key := 8
	want := "fifty"

	trigger.SetValue(m)
	trigger.Set(key, want)

	got := trigger.value[key]

	if got != want {
		t.Fatalf("Expected value at key %v to change to %v; got %v", key, want, got)
	}
}

func TestTriggerSetCreate(t *testing.T) {
	var trigger Trigger
	m := make(map[interface{}]interface{})
	key := 12
	want := "twelve"

	trigger.SetValue(m)
	trigger.Set(key, want)
	got,ok := trigger.value[key]

	if !ok {
		t.Errorf("Expected creation of key %v. Wasn't created", key)
	}

	if got != want {
		t.Fatalf("Expected creation of key %v with value %v; got %v", key, want, got)
	}
}

func TestTriggerSetIndependent(t *testing.T) {
	var trigger Trigger
	m := initMap()

	trigger.SetValue(m)
	trigger.Set(7, "something new")

	if reflect.DeepEqual(m, trigger.value) {
		t.Fatalf("Expected source and trigger maps to be different")
	}
}

func TestTriggerSetFromNil(t *testing.T) {
	var trigger Trigger

	trigger.Set("who cares?", "doesn't matter")
	if trigger.value == nil {
		t.Fatal("Expected unitialized trigger to exist after Get; got <nil>")
	}
}

// Delete

func TestTriggerDelete(t *testing.T) {
	var trigger Trigger
	m := initMap()
	key := 5

	trigger.SetValue(m)
	trigger.Delete(key)

	_,ok := trigger.value[key]

	if ok {
		t.Fatalf("Expected key %v to be deleted but wasn't", key)
	}
}

func TestTriggerDeleteNil(t *testing.T) {
	var trigger Trigger
	m := initMap()
	key := 100

	trigger.SetValue(m)

	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Deleting non-existant key should be a no-op; Delete panicked")
		}
	}()
	trigger.Delete(key)

	if _,ok := trigger.value[key]; ok {
		t.Fatalf("Key %v should have been deleted", key)
	}
}

func TestTriggerDeleteIndependent(t *testing.T) {
	var trigger Trigger
	m := initMap()
	key := 9

	trigger.SetValue(m)
	trigger.Delete(key)

	if reflect.DeepEqual(m, trigger.value) {
		if _,ok := m[key]; !ok {
			t.Fatalf("Key %v deleted from source map", key)
		}
		t.Fatal("Delete should not modify source map")
	}
}

func TestTriggerDeleteFromNil(t *testing.T) {
	var trigger Trigger

	trigger.Get("who cares?")
	if trigger.value == nil {
		t.Fatal("Expected unitialized trigger to exist after Get; got <nil>")
	}
}

// Immutability tests not needed, differing output type forces copying of some type
func TestTriggerKeys(t *testing.T) {
	var trigger Trigger
	m := initMap()

	trigger.SetValue(m)
	got := trigger.Keys()

	if len(got) != len(m) {
		t.Fatalf("Expected %d keys to be in slice; got %d", len(m), len(got))
	}
	for _,k := range got {
		if _,ok := m[k]; !ok {
			t.Fatalf("Key %v in slice; not found in source map", k)
		}
	}
}

func TestTriggerKeysFromNil(t *testing.T) {
	var trigger Trigger
	
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Keys panicked. Why did it panic?")
		}
	}()
	got := trigger.Keys()

	if len(got) != 0 {
		t.Fatal("Expected keys to be empty, but wasn't")
	}
}

func TestTriggerValues(t *testing.T) {
	var trigger Trigger
	m := initMap()
	var values []interface{}
	for _,v := range values {
		values = append(values, v)
	}


	trigger.SetValue(m)
	got := trigger.Values()

	if len(got) != len(m) {
		t.Fatalf("Expected %d values to be in slice; got %d", len(m), len(got))
	}

	for _,v := range got {
		for i:=0; i<len(values); i++ {
			if v == values[i] {
				values = append(values[:i], values[i+1:]...)
				break
			}
		}
	}
	if len(values) > 0 {
		t.Fatalf("Expected slice to be equal to source; different values: %v", values)
	}
}

func TestTriggerValuesFromNil(t *testing.T) {
	var trigger Trigger

	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Values panicked. Why did it panic?")
		}
	}()
	got := trigger.Values()

	if len(got) != 0 {
		t.Fatal("Expected values to be empty, but wasn't")
	}
}

func TestTriggerKeysValuesEqual(t *testing.T) {
	var trigger Trigger
	m := initMap()

	trigger.SetValue(m)

	gotKeys := trigger.Keys()
	gotValues := trigger.Values()

	if len(gotKeys) != len(gotValues) {
		t.Fatal("Expected Keys and Values to return equal length slices")
	}

	if len(gotKeys) != len(m) {
		t.Fatal("Expected Keys and Values to be same length as source")
	}
}

func TestTriggerSize(t *testing.T) {
	var trigger Trigger
	m := initMap()

	trigger.SetValue(m)

	got := trigger.Size()

	if got != len(m) {
		t.Fatalf("Expected size to be equal to source (%d); got %d", len(m), got)
	}
}

func TestTriggerKeyReadCallback(t *testing.T) {
	var trigger Trigger
    m := initMap()
    var count int
    callback := func(v interface{}) {
        index := v.(Pair)
        if m[index.Key] != index.Value {
            t.Fatalf("Expected index %d to be %v; got %v", index.Key, m[index.Key], index.Value)
        }
        count += 1
    }

    trigger.AddKeyReadCallback(callback)
    trigger.SetValue(m)

	for k,_ := range m {
		trigger.Get(k)
	}

    if count != trigger.Size() {
        t.Fatalf("Expected count to be %d after %d iterations; got %d", 
		         trigger.Size(), trigger.Size(), count)
    }
}

func TestTriggerKeyWriteCallback(t *testing.T) {
	var trigger Trigger
    m := initMap()
    var count int
    callback := func(prev, v interface{}) {
        index := v.(Pair)
        p := prev.(Pair)
        if p.Value != m[p.Key] {
            t.Fatalf("Expected previous value at index %d to be %v; got %v", p.Key, m[p.Key], p.Value)
        }
        if index.Value != "eggs!" {
            t.Fatalf("Expected index %d to be \"eggs!\"; got %v", index.Key, index.Value)
        }
        count += 1
    }

    trigger.AddKeyWriteCallback(callback)
    trigger.SetValue(m)

	for k,_ := range m {
		trigger.Set(k, "eggs!")
	}

    if count != len(m) {
        t.Fatalf("Expected count to be %d after %d iterations; got %d", len(m), len(m), count)
    }
}
