package slice

import(
	"testing"
)


func TestSliceReadCallbackWrapper(t *testing.T) {
    var trigger Trigger
    var run bool
    callback := SliceReadCallback(func(v []interface{}) {
        run = true
    })

    trigger.AddReadCallback(callback)
    trigger.Value()

    if !run {
        t.Fatal("Wrapped callback not run")
    }
}

func TestSliceWriteCallbackWrapper(t *testing.T) {
    var trigger Trigger
    var run bool
    callback := SliceWriteCallback(func(prev, v []interface{}) {
        run = true
    })
        
    trigger.AddWriteCallback(callback)
    trigger.SetValue([]interface{}{1})
    
    if !run {
        t.Fatal("Wrapped callback not run")
    }
}

func TestSliceIndexReadCallbackWrapper(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	var count int
	callback := IndexReadCallback(func(i int, v interface{}) {
		count += 1
		if arr[i] != v {
			t.Fatalf("Expected value at index %d to be %v; got %v", i, arr[i], v)
		}
	})

	trigger.AddIndexReadCallback(callback)
	trigger.SetValue(arr)

	iters := len(arr)
	for i:=0; i<iters; i++ {
		trigger.At(i)
	}

	if count != iters {
		t.Fatalf("Expected count to be %d after %d iterations; got %d", iters, iters, count)
	}
}

func TestSliceIndexWriteCallbackWrapper(t *testing.T) {
	var trigger Trigger
	arr := []interface{}{1,2,3,4,5}
	var count int
	callback := IndexWriteCallback(func(prevDex int, prev interface{}, i int, v interface{}) {
		if prev != arr[prevDex] {
			t.Fatalf("Expected previous value at index %d to be %v; got %v", prevDex, arr[prevDex], prev)
		}
		if v != 0 {
			t.Fatalf("Expected index %d to be 0; got %v", i, v)
		}
		count += 1
	})

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



