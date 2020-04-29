package slice

import (
	"testing"
	"reflect"
	
	"github.com/KellenWatt/reactor"
)

func TestTriggerSetValue(t *testing.T) {
	var trigger Trigger
	want := []interface{}{1,2,3,4,5}

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

func TestTriggerSetValueSliceTypeCheck(t *testing.T) {
	var trigger Trigger
	
	defer func() {
		if recover() == nil {
			t.Fatal("Expected value when assigning non-interface{} slice value")
		}
	}()

	trigger.SetValue([]int{10})
}

func TestTriggerIndependentSetValue(t *testing.T) {
	var trigger Trigger
	src := []interface{}{1,2,3,4,5}
	want := []interface{}{1,2,3,4,5}

	trigger.SetValue(src)

	src[3] = 0

	if reflect.DeepEqual(src, trigger.value) {
		t.Fatalf("Changes to source should not affect internal: Expected %v; got %v", want, trigger.value)
	}
}

func TestTriggerIndependentValueResult(t *testing.T) {
	var trigger Trigger
	want := []interface{}{1,2,3,4,5}

	trigger.SetValue(want)
	val := trigger.Value().([]interface{})

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
		prevVal := prev.([]interface{})
		val := v.([]interface{})
		count += 1
		if len(prevVal) > 0 && prevVal[0] != val[0].(int)-1 {
			t.Fatalf("prev not being set correctly. Expected %v, got %v", []interface{}{val[0].(int)-1}, prevVal)
		}
	}

	trigger.AddWriteCallback(callback)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue([]interface{}{i})
	}

	if count != iters {
		t.Fatalf("ReadCallbacks not being run. Of %d iterations, ran %d times; expected %d", iters, count, iters)
	}
}

func TestTriggerAt(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	index := 2
	want := arr[index]

	trigger.SetValue(arr)
	got,_ := trigger.At(index)

	if got != want {
		t.Fatalf("Expected %v; got %v", want, got)
	}
}

func TestTriggerAtFail(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}

	trigger.SetValue(arr)
	// len(arr) guaranteed to be out of range
	_,err := trigger.At(len(arr))

	if err == nil {
		t.Fatalf("Expected OutOfBoundsError; got nothing")
	}
}

func TestTriggerSetAt(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	index := 3
	change := arr[index]

	num := 100
	trigger.SetAt(index, num)
	got,_ := trigger.At(index)

	if change == got {
		t.Fatalf("Expected index %d to change to %v; got %v", index, num, got)
	}	
}

func TestTriggerSetAtError(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	
	trigger.SetValue(arr)
	// len(Arr) guaranteed to be out of range
	err := trigger.SetAt(len(arr), 0)

	if err == nil {
		t.Fatalf("Expected OutOfBoundsError; got nothing")
	}
}

func TestTriggerAppend(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	add := 6

	trigger.SetValue(arr)
	trigger.Append(6)

	got := trigger.Value()

	if !reflect.DeepEqual(append(arr, add), got) {
		t.Fatalf("Expected %v to be appended to %v; got %v", add, arr, got)
	}
}

func TestTriggerPop(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	wantArr := arr[:len(arr)-1]
	want := arr[len(arr)-1]

	trigger.SetValue(arr)
	got,err := trigger.Pop()


	if err != nil {
		t.Fatalf("Expected success; got error: %v", err)
	}

	if !reflect.DeepEqual(wantArr, trigger.value) {
		t.Fatalf("Expected %v; got %v", wantArr, trigger.value)
	}

	if got != want {
		t.Fatalf("Pop returned %v; wanted %v", got, want)
	}
}

func TestTriggerPopEmpty(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{}

	trigger.SetValue(arr)
	_,err := trigger.Pop()

	if err == nil {
		t.Fatalf("Expected an error; got nil")
	}
}

func TestTriggerSlice(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	min,max := 2,4
	want := arr[min:max]

	trigger.SetValue(arr)
	got,err := trigger.Slice(min, max)

	if err != nil {
		t.Fatalf("Didn't expect error; got error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Expected slice %v; got %v", want, got)
	}
}

func TestTriggerIndependentSlice(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	min,max := 1,4
	want := arr[min:max]

	trigger.SetValue(arr)
	tmp,_ := trigger.Slice(min, max)
	tmp[0] = 0

	got,_ := trigger.Slice(min, max)

	if reflect.DeepEqual(tmp, got) {
		t.Fatalf("Expected constant slice. Expected %v; got %v", want, got)
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Expected equality to initial slice. Expected %v; got %v", want, got)
	}
}

func TestTriggerSliceEmpty(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	bound := 3

	trigger.SetValue(arr)
	got,err := trigger.Slice(bound, bound)

	if err != nil {
		t.Fatalf("Didn't expect error; got error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("Expected empty slice; got %v", got)
	}
}

func TestTriggerSliceError(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}

	trigger.SetValue(arr)
	_,err := trigger.Slice(1,0)
	if err == nil {
		t.Fatalf("Lower bound greater than upper, got no error")
	}

	_,err = trigger.Slice(-1,0)
	if err == nil {
		t.Fatalf("Lower bound below 0, got no error")
	}

	_,err = trigger.Slice(0, len(arr)+1)
	if err == nil {
		t.Fatalf("Upper bound above length, got no error")
	}

	_,err = trigger.Slice(len(arr)+1, len(arr)+1)
	if err == nil {
		t.Fatalf("Lower bound above length, got no error")
	}

	got,err := trigger.Slice(len(arr), len(arr))
	if err != nil {
		t.Fatalf("Lower bound at maximum, expected no error, got error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Expected empty slice, got %v", got)
	}
}

func TestTriggerSize(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	want := len(arr)
	trigger.SetValue(arr)

	got := trigger.Size()

	if want != got {
		t.Fatalf("Expected size to be %d; got %d", want, got)
	}
}

func TestTriggerIndexReadCallback(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	var count int
	callback := func(v interface{}) {
		index := v.(Index)
		if arr[index.Key] != index.Value {
			t.Fatalf("Expected index %d to be %v; got %v", index.Key, arr[index.Key], index.Value)
		}
		count += 1
	}

	trigger.AddIndexReadCallback(callback)
	trigger.SetValue(arr)

	iters := len(arr)
	for i:=0; i<len(arr); i++ {
		trigger.At(i)
	}

	if count != iters {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", iters, iters, count)
	}	
}

func TestTriggerIndexWriteCallback(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	var count int
	callback := func(prev, v interface{}) {
		index := v.(Index)
		p := prev.(Index)
		if p.Value != arr[p.Key] {
			t.Fatalf("Expected previous value at index %d to be %v; got %v", p.Key, arr[p.Key], p.Value)
		}
		if index.Value != 0 {
			t.Fatalf("Expected index %d to be 0; got %v", index.Key, index.Value)
		}
		count += 1
	}

	trigger.AddIndexWriteCallback(callback)
	trigger.SetValue(arr)

	iters := len(arr)
	for i:=0; i<len(arr); i++ {
		trigger.SetAt(i, 0)
	}

	if count != iters {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", iters, iters, count)
	}	
}
