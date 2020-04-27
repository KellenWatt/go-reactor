package reactor

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

