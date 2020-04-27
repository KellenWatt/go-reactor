package reactor

import (
	"errors"
)

var CallbackFormatError = errors.New("Improper callback type")

// Callback <- func(...interface{})


// Async decorates a Callback as asynchronous.
func (c Callback) Async() Callback {
	return func(v ...interface{}) {
		go c(v...)
	}
}

// Concurrent decorates a Callback as concurrent.
func (c Callback) Concurrent() Callback {
	conEventLock.Lock()
	if conEvent == nil {
		conEvent = make(chan conEventState, 100)
		go runConcurrentEvent()
	}
	conEventLock.Unlock()
	return func(v ...interface{}) {
		conEvent <- conEventState{v, c}
	}
}

// Conditional decorates a Callback as conditional. w will be executed
// only if f returns true.
func (c Callback) Conditional(f func(...interface{})bool) Callback {
	return func(v ...interface{}) {
		if f(v...) {
			c(v...)
		}
	}
}

// ExtractRead extracts the value passed to a Callback if that value is 
// intended for a read callback. Returns an error if there is more than one 
// value passed to the function.
func ExtractRead(v ...interface{}) (interface{}, error) {
	if len(v) != 1 {
		return nil, CallbackFormatError
	}
	return v[0], nil
}

// ExtractWrite extracts the values passed to a Callback if those values are 
// intended for a write callback. Returns an error if there aren't exactly two 
// values passed to the function.
func ExtractWrite(v ...interface{}) (interface{}, interface{}, error) {
	if len(v) != 2 {
		return nil, nil, CallbackFormatError
	}
	return v[0], v[1], nil
}



