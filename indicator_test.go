package reactor

import (
	"testing"
	"strconv"
)

func TestIndicatorSetValue(t *testing.T) {
	var trigger Indicator
	want := 10

	trigger.SetValue(want)

	got := trigger.value
	resgot := trigger.Value()

	if got != want {
		t.Fatalf("internal value = %d; want %d", got, want)
	}

	if got != resgot {
		t.Fatalf("Value() = %d; want %d", resgot, want)
	}
}

func TestIndicatorAddBinder(t *testing.T) {
	var trigger Indicator
	var ind Indicator

	trigger.AddBinder(&ind, func(v interface{})interface{}{return v.(int)+1}, false)

	if len(trigger.bindings) != 1 {
		t.Fatal("No binding added")
	}
}

func TestIndicatorReadCallback(t *testing.T) {
	var trigger Indicator
	var count int
	callback := func(v interface{}) {
		count += 1
	}
	
	trigger.AddReadCallback(callback)
	trigger.Value()
	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=count; i<countWant; i++ {
		trigger.Value()
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d; got %d", countWant, count)
	}
}

func TestIndicatorNilReadCallback(t *testing.T) {
	var trigger Indicator
	var got interface{}
	callback := func(v interface{}) {
		got = v
	}

	trigger.AddReadCallback(callback)
	trigger.Value()

	if got != nil {
		t.Errorf("Unitialized trigger should have Value of nil. Got %v", got)
	}
}

// Async callbacks make no guarantee of order or execution time/priority
// nor should they be strictly expected to do so.
func TestIndicatorAsyncReadCallback(t *testing.T) {
	var trigger Indicator
	var count int
	wait := make(chan int)
	callback := ReadCallback(func(v interface{}) {
		trigger.Lock.Lock()
		count += 1
		trigger.Lock.Unlock()
		wait <- count
	}).Async()

	trigger.AddReadCallback(callback)
	trigger.Value()

	<-wait
	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=1; i<countWant; i++ {
		trigger.Value()
	}

	for i:=1; i<countWant; {
		select {
		case <-wait:
			i++
		}
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d; got %d", countWant, count)
	}
}

// Concurrent callbacks are guaranteed to be executed in the order received
// in a non-parallel fashion. Parallelism added by the developer is beyond 
// the scope of this module and is not accounted for.
func TestIndicatorConcurrentReadCallback(t *testing.T) {
	var trigger Indicator
	var count int
	wait := make(chan int)
	callback := ReadCallback(func(v interface{}) {
		trigger.Lock.Lock()
			count += 1
		trigger.Lock.Unlock()
		wait <- count
	}).Concurrent()
	// stops read concurrency mechanism, to ensure test isolation

	trigger.AddReadCallback(callback)
	defer killRead()
	trigger.Value()

	<-wait
	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=count; i<countWant; i++ {
		trigger.Value()
	}

	for i:=count; i<countWant; {
		select {
		case <-wait:
			i++
		}
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d; got %d", countWant, count)
	}

}

func TestIndicatorConditionalReadCallback(t *testing.T) {
	var trigger Indicator
	var count int
	maxCount := 5
	condition := func(v interface{}) bool {
		return count < maxCount
	}
	callback := ReadCallback(func(v interface{}) {
		count += 1
	}).Conditional(condition)

	trigger.AddReadCallback(callback)
	trigger.Value()

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.Value()
	}

	if count != maxCount {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", maxCount, iters, count)
	}
}

func TestIndicatorWriteCallback(t *testing.T) {
	var trigger Indicator
	var count int
	var prevValue int
	callback := func(prev, v interface{}) {
		if prev != nil {
			prevValue = prev.(int)
		}
		count += 1
	}

	trigger.AddWriteCallback(callback)
	trigger.SetValue(1)

	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=count; i<countWant; i++ {
		trigger.SetValue(i+1)
		if prevValue != i {
			t.Fatalf("Previous not supplied correctly. Want %d; got %d", i, prevValue)
		}
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d, got %d", countWant, count)
	}
}

func TestIndicatorNilWriteCallback(t *testing.T) {
	var trigger Indicator
	var got interface{}
	callback := func(prev, v interface{}) {
		got = prev
	}

	trigger.AddWriteCallback(callback)
	trigger.SetValue(0)

	if got != nil {
		t.Errorf("First call to SetValue should have value of nil. Got %v", got)
	}
}

// Async callbacks make no guarantee of order or execution time/priority
// nor should they be strictly expected to do so.
func TestIndicatorAsyncWriteCallback(t *testing.T) {
	var trigger Indicator
	var count int
	wait := make(chan int)
	callback := WriteCallback(func(prev, v interface{}) {
		trigger.Lock.Lock()
		count += 1
		trigger.Lock.Unlock()
		
		if prev != nil && prev.(int) != v.(int)-1 {
			t.Fatalf("Previous value for %v should be %v; got %v", v, v.(int)-1, prev)
		}
		wait <- count
	}).Async()

	trigger.AddWriteCallback(callback)
	trigger.SetValue(1)

	<-wait
	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=1; i<countWant; i++ {
		trigger.SetValue(i+1)
	}

	for i:=1; i<countWant; i++{
		<-wait
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d; got %d", countWant, count)
	}
}

// Concurrent callbacks are guaranteed to be executed in the order received
// in a non-parallel fashion. Parallelism added by the developer is beyond 
// the scope of this module and is not accounted for.
func TestIndicatorConcurrentWriteCallback(t *testing.T) {
	var trigger Indicator
	var count int
	wait := make(chan int)
	callback := WriteCallback(func(prev, v interface{}) {
		trigger.Lock.Lock()
			count += 1
		trigger.Lock.Unlock()
		wait <- count
	}).Concurrent()

	trigger.AddWriteCallback(callback)
	defer killWrite()
	trigger.SetValue(1)

	<-wait
	if count != 1 {
		t.Fatalf("Expected count to be 1; got %d", count)
	}

	countWant := 10
	for i:=1; i<countWant; i++ {
		trigger.SetValue(i+1)
	}

	for i:=1; i<countWant; i++{
		<-wait
	}

	if count != countWant {
		t.Fatalf("Expected count to be %d; got %d", countWant, count)
	}
}

func TestIndicatorConditionalWriteCallback(t *testing.T) {
	var trigger Indicator
	var count int
	var prevVal interface{}
	condition := func(prev,v interface{}) bool {
		if prev == nil {
			prevVal = 0
		} else if v.(int) - prevVal.(int) == 2 {
			prevVal = v
			return true
		}

		return false
	}
	callback := WriteCallback(func(prev,v interface{}) {
		count += 1
	}).Conditional(condition)

	trigger.AddWriteCallback(callback)
	trigger.SetValue(1)

	iters := 10
	for i:=1; i<iters; i++ {
		trigger.SetValue(i+1)
	}

	maxCount := 5
	if count != maxCount {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", maxCount, iters, count)
	}
}

// check if two triggers can share the concurrency queue peacefully
func TestIndicatorMultipleConcurrentRead(t *testing.T) {
	var t1,t2 Indicator
	var out1,out2 int
	wait1,wait2 := make(chan bool), make(chan bool)
	c1 := ReadCallback(func(v interface{}) {
		out1 += 1
		wait1 <- true
	}).Concurrent()
	c2 := ReadCallback(func(v interface{}) {
		out2 += 1
		wait2 <- true
	}).Concurrent()

	t1.AddReadCallback(c1)
	t2.AddReadCallback(c2)
	defer killRead()

	maxCount := 10
	for i:=0; i<maxCount; i++ {
		t1.Value()
		t2.Value()
	}

	var seen1,seen2 int
	for seen1<maxCount || seen2<maxCount {
		select {
		case <-wait1:
			seen1 += 1
		case <-wait2:
			seen2 += 1
		}
	}

	if seen1 != maxCount {
		t.Errorf("First trigger counted %d times; expected %d", seen1, maxCount)
	}
	if seen2 != maxCount {
		t.Errorf("Second trigger counted %d times; expected %d", seen2, maxCount)
	}
}

// check if two triggers can share the concurrency queue peacefully
func TestIndicatorMultipleConcurrentWrite(t *testing.T) {
	var t1,t2 Indicator
	wait1,wait2 := make(chan bool), make(chan bool)
	c1 := WriteCallback(func(prev, v interface{}) {
		wait1 <- true
	}).Concurrent()
	c2 := WriteCallback(func(prev, v interface{}) {
		wait2 <- true
	}).Concurrent()

	t1.AddWriteCallback(c1)
	t2.AddWriteCallback(c2)
	defer killWrite()

	maxCount := 10
	for i:=0; i<maxCount; i++ {
		t1.SetValue(i)
		t2.SetValue(-i)
	}

	var seen1,seen2 int
	for seen1<maxCount || seen2<maxCount {
		select {
		case <-wait1:
			seen1 += 1
		case <-wait2:
			seen2 += 1
		}
	}

	if seen1 != maxCount {
		t.Errorf("First trigger counted %d times; expected %d", seen1, maxCount)
	}
	if seen2 != maxCount {
		t.Errorf("Second trigger counted %d times; expected %d", seen2, maxCount)
	}
}

func TestIndicatorCombinedCallbacks(t *testing.T) {
	var trigger Indicator

	var nums []int
	var nextNums []int
	readCallback := func(v interface{}) {
		nums = append(nums, v.(int))
	}
	writeCallback := func(prev, v interface{}) {
		nextNums = append(nextNums, v.(int))
	}

	trigger.SetValue(0)
	
	trigger.AddReadCallback(readCallback)
	trigger.AddWriteCallback(writeCallback)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue(trigger.Value().(int) + 1)
	}

	if len(nums) != len(nextNums) {
		t.Fatalf("Expected arrays to be equal length; got %d and %d", len(nums), len(nextNums))
	}

	for i,n := range nums {
		if n+1 != nextNums[i] {
			t.Fatalf("Expected arrays to be offset by one; got %v, and %v", nums, nextNums)
		}
	}
}


func TestAddBinding(t *testing.T) {
	var trigger Trigger
	var ind Indicator
	bindFunc := func(v interface{}) interface{} {
		return v.(int) * 2
	}

	ind.AddBinding(&trigger, bindFunc)
	if len(trigger.bindings) != 1 {
		t.Fatal("Binding not registered")
	}

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue(i)
		if ind.Value() != 2 * trigger.Value().(int) {
			t.Fatalf("Binding inconsistent. Got %v for indicator from trigger of %v; want %d", 
					 ind.Value(), trigger.Value(), 2*i)
		}
	}
}

// This is done solely for an interface{} implementation, and may not be 
// feasible in a generic one.
func TestMultitypeBinding(t *testing.T) {
	var trigger Trigger
	var ind Indicator
	bindFunc := func(v interface{}) interface{} {
		var str []rune
		n := v.(int)
		for n != 0 {
			str = append([]rune{rune((n%10)+'0')}, str...)
			n /= 10
		}
		return string(str)
	}

	ind.AddBinding(&trigger, bindFunc)

	for i:=1; i<=100; i++ {
		trigger.SetValue(i)
		if ind.Value().(string) != strconv.Itoa(i) {
			t.Fatalf("Got %v; expected %s", ind.Value(), strconv.Itoa(i))
		}
	}

}

func TestAddTrivialBinding(t *testing.T) {
	var trigger Trigger
	var ind Indicator

	ind.AddTrivialBinding(&trigger)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue(i)
		if ind.Value() != trigger.Value() {
			t.Fatalf("Expected equality. Got %v for trigger of %v", ind.Value(), trigger.Value())
		}
	}
}

func TestAddDelayedBinding(t *testing.T) {
	var trigger Trigger
	var ind Indicator
	var count int
	bindFunc := func(v interface{}) interface{} {
		count += 1
		return v
	}

	ind.AddDelayedBinding(&trigger, bindFunc)

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.SetValue(i)
	}

	if count != 0 {
		t.Fatal("Delayed binding executed without call to Value")
	}
	
	n := ind.Value()
	if count != 1 || n.(int) != iters-1 {
		t.Fatalf("Expected count to be 1; got %d\nExpected Value to be %d; got %v", count, iters-1, n)
	}

	ind.Value()
	if count != 2 || n.(int) != iters-1 {
		t.Fatalf("Expected count to be 2; got %d\nValue inconsistent across calls to be %d; got %v", 
		         count, iters-1, n)
	}
}

func TestConcurrentBinding(t *testing.T) {
	var trigger Trigger
	var ind Indicator
	var count int
	bindFunc := func(v interface{}) interface{} {
		count += 1
		return v.(int)+1
	}

	ind.AddConcurrentBinding(&trigger, bindFunc)
	defer killBind()

	trigger.SetValue(0)

	// this is bad, but I can't think of any other way to do it
	for ind.Value() == nil || ind.Value().(int) != trigger.Value().(int) + 1 {
		// There should be a Fatal in here, but I can't think of how to do that reliably
		t.Log("Indicator not updated yet")
	}
	// Failure state should be here, but haven't figured out reliable logic for that
}


func TestChainBinding(t *testing.T) {
	var trigger Trigger
	var ind1,ind2 Indicator 

	bindFunc1 := func(v interface{}) interface{} {
		return v.(int) + 1
	}
	bindFunc2 := func(v interface{}) interface{} {
		return v.(int) * 2
	}

	ind1.AddBinding(&trigger, bindFunc1)
	ind2.AddBinding(&ind1, bindFunc2)

	trigger.SetValue(1)

	if ind1.Value() != 2 {
		t.Errorf("Expected first Indicator to be %d, with trigger %v; got %v",
	             trigger.Value().(int) + 1, trigger.Value(), ind1.Value())
	}

	if ind2.Value() != 4 {
		t.Fatalf("Expected second Indicator to be %d, with trigger %v; got %v",
				 (trigger.Value().(int)+1)*2, trigger.Value(), ind2.Value())
	}
}


func TestMultipleBindings(t *testing.T) {
	var trigger Trigger
	var ind1,ind2 Indicator 

	bindFunc1 := func(v interface{}) interface{} {
		return v.(int) + 1
	}
	bindFunc2 := func(v interface{}) interface{} {
		return v.(int) * 2
	}
	
	ind1.AddBinding(&trigger, bindFunc1)
	ind2.AddBinding(&trigger, bindFunc2)

	
	trigger.SetValue(2)

	if ind1.Value() != 3 {
		t.Errorf("Expected first Indicator to be %d, with trigger %v; got %v",
	             trigger.Value().(int) + 1, trigger.Value(), ind1.Value())
	}

	if ind2.Value() != 4 {
		t.Fatalf("Expected second Indicator to be %d, with trigger %v; got %v",
				 trigger.Value().(int) * 2, trigger.Value(), ind2.Value())
	}
}
