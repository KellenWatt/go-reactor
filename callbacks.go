package reactor

// ReadCallback <- func(interface{})
// WriteCallback <- func(interface{}, interface{})
// BindingFunc <- func(interface{}) interface{}

// Async decorates a ReadCallback as asynchronous.
func (r ReadCallback) Async() ReadCallback {
	return func(v interface{}) {
		go r(v)
	}
}

// Concurrent decorates a ReadCallback as concurrent.
func (r ReadCallback) Concurrent() ReadCallback {
	conReadLock.Lock()
		if conRead == nil {
			conRead = make(chan conReadState, 100)
			go runConcurrentRead()
		}
	conReadLock.Unlock()
	return func(v interface{}) {
		conRead <- conReadState{v, r}
	}
}

// Conditional decorates a ReadCallback as conditional. r will be executed 
// only if f returns true.
func (r ReadCallback) Conditional(f func(interface{}) bool) ReadCallback {
	return func(v interface{}) {
		if f(v) {
			r(v)
		}
	}
}

// Async decorates a WriteCallback as asynchronous.
func (w WriteCallback) Async() WriteCallback {
	return func(prev, v interface{}) {
		go w(prev, v)
	}
}

// Concurrent decorates a WriteCallback as concurrent.
func (w WriteCallback) Concurrent() WriteCallback {
	conWriteLock.Lock()
		if conWrite == nil {
			conWrite = make(chan conWriteState, 100)
			go runConcurrentWrite()
		}
	conWriteLock.Unlock()
	return func(prev, v interface{}) {
		conWrite <- conWriteState{prev, v, w}
	}
}

// Conditional decorates a WriteCallback as conditional. w will be executed
// only if f returns true.
func (w WriteCallback) Conditional(f func(interface{},interface{}) bool) WriteCallback {
	return func(prev, v interface{}) {
		if f(prev, v) {
			w(prev, v)
		}
	}
}


func (c Callback) Async() Callback {
	return func(v ...interface{}) {
		go c(v...)
	}
}

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

func (c Callback) Conditional(f func(...interface{})bool) Callback {
	return func(v ...interface{}) {
		if f(v...) {
			c(v...)
		}
	}
}

