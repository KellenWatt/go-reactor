package reactor

import (
	"testing"
)

func TestSetValue(t *testing.T) {
	var trigger Trigger
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

func TestAddBinder(t *testing.T) {
	var trigger Trigger
	var ind Indicator

	trigger.AddBinder(&ind, func(v interface{})interface{}{return v.(int)+1}, false)

	if len(trigger.bindings) != 1 {
		t.Fatal("No binding added")
	}
}

func TestReadCallback(t *testing.T) {
	var trigger Trigger
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

func TestNilReadCallback(t *testing.T) {
	var trigger Trigger
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
func TestAsyncReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(v interface{}) {
		trigger.Lock.Lock()
		count += 1
		trigger.Lock.Unlock()
		wait <- count
	}

	trigger.AddAsyncReadCallback(callback)
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
func TestConcurrentReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(v interface{}) {
		trigger.Lock.Lock()
			count += 1
		trigger.Lock.Unlock()
		wait <- count
	}
	// stops read concurrency mechanism, to ensure test isolation
	defer killRead()

	trigger.AddConcurrentReadCallback(callback)
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

func TestConditionalReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	callback := func(v interface{}) {
		count += 1
	}
	maxCount := 5
	condition := func(v interface{}) bool {
		return count < maxCount
	}

	trigger.AddConditionalReadCallback(callback, condition)
	trigger.Value()

	iters := 10
	for i:=0; i<iters; i++ {
		trigger.Value()
	}

	if count != maxCount {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", maxCount, iters, count)
	}
}

func TestWriteCallback(t *testing.T) {
	var trigger Trigger
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

func TestNilWriteCallback(t *testing.T) {
	var trigger Trigger
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
func TestAsyncWriteCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(prev, v interface{}) {
		trigger.Lock.Lock()
		count += 1
		trigger.Lock.Unlock()
		
		if prev != nil && prev.(int) != v.(int)-1 {
			t.Fatalf("Previous value for %v should be %v; got %v", v, v.(int)-1, prev)
		}
		wait <- count
	}

	trigger.AddAsyncWriteCallback(callback)
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
func TestConcurrentWriteCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(prev, v interface{}) {
		trigger.Lock.Lock()
			count += 1
		trigger.Lock.Unlock()
		wait <- count
	}

	trigger.AddConcurrentWriteCallback(callback)
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

func TestConditionalWriteCallback(t *testing.T) {
	var trigger Trigger
	var count int
	callback := func(prev,v interface{}) {
		count += 1
	}
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

	trigger.AddConditionalWriteCallback(callback, condition)
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
func TestMultipleConcurrentRead(t *testing.T) {
	var t1,t2 Trigger
	var out1,out2 int
	wait1,wait2 := make(chan bool), make(chan bool)
	c1 := func(v interface{}) {
		out1 += 1
		wait1 <- true
	}
	c2 := func(v interface{}) {
		out2 += 1
		wait2 <- true
	}

	t1.AddConcurrentReadCallback(c1)
	t2.AddConcurrentReadCallback(c2)
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
func TestMultipleConcurrentWrite(t *testing.T) {
	var t1,t2 Trigger
	wait1,wait2 := make(chan bool), make(chan bool)
	c1 := func(prev, v interface{}) {
		wait1 <- true
	}
	c2 := func(prev, v interface{}) {
		wait2 <- true
	}

	t1.AddConcurrentWriteCallback(c1)
	t2.AddConcurrentWriteCallback(c2)
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

func TestCombinedCallbacks(t *testing.T) {
	var trigger Trigger

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

