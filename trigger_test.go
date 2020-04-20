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

func TestAsyncReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(v interface{}) {
		count += 1
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

func TestConcurrentReadCallback(t *testing.T) {
	var trigger Trigger
	var count int
	wait := make(chan int)
	callback := func(v interface{}) {
		count += 1
		wait <- count
	}

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
