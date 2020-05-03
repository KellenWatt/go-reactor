package dict

import (
	"testing"
)

func TestMapReadCallbackWrapper(t *testing.T) {
	var trigger Trigger
	var run bool
	callback := MapReadCallback(func(map[interface{}]interface{}) {
		run = true
	})

	trigger.AddReadCallback(callback)
	trigger.Value()

	if !run {
		t.Fatal("Wrapped callback not run")
	}
}

func TestMapWriteCallbackWrapper(t *testing.T) {
	var trigger Trigger
	var run bool
	callback := MapWriteCallback(func(map[interface{}]interface{}, map[interface{}]interface{}) {
		run = true
	})

	trigger.AddWriteCallback(callback)
	trigger.SetValue(map[interface{}]interface{}{})

	if !run {
		t.Fatal("Wrapped callback not run")
	}
}

func TestKeyReadCallbackWrapper(t *testing.T) {
	var trigger Trigger
	var run bool
	callback := KeyReadCallback(func(interface{}, interface{}) {
		run = true
	})

	trigger.AddKeyReadCallback(callback)
	trigger.Get("Doesn't matter")

	if !run {
		t.Fatal("Wrapped callback not run")
	}
}

func TestKeyWriteCallbackWrapper(t *testing.T) {
	var trigger Trigger
	var run bool
	callback := KeyWriteCallback(func(pKey, pVal, key, val interface{}) {
		run = true
	})

	trigger.AddKeyWriteCallback(callback)
	trigger.Set("Doesn't matter", "Who cares")

	if !run {
		t.Fatal("Wrapped callback not run")
	}
}
