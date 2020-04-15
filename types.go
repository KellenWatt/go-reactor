// Package reactor provides an API for a basic event system. To facilitate this 
// cleanly, the interfaces provide several different types of responsiveness.
//
// For basic information on thread safety, see 
// https://github.com/KellenWatt/reactor. For more specific information
// see the specific type implementations.
package reactor

// A Comparable represents a value that can be ordered
type Comparable interface {
	// Compare returns -1 for <, 0 for ==, 1 for >
	Compare(Comparable) int
	// Retruns the value of the Comparable
	Value() interface{}
}

// variable defines the basic data unit of data for package reactor
type Variable interface {
	// Sets the internal Comparable value, with possible additional checks
	SetValue(Comparable)
	// Returns the internal Comparable value
	Value() Comparable
}

// ReadCallback defines the basic function used in all read callbacks.
// A ReadCallback's argument is the value when the callback was invoked. This 
// is supplied instaed of required a call to Reactor.Value(), to prevent 
// accidental infinite recursion
type ReadCallback func(Comparable)

// WriteCallback defines the basic function used in all write callbacks.
// A WriteCallback's arguments are as follows: the previous value and the new 
// value when the callback has been invoked. This is supplied instead of 
// requiring a call to Reactor.Value() to prevent accidental indirect recursion.
type WriteCallback func(Comparable, Comparable)

// ReadReactor is an interface that defines the basic methods that enable 
// various form of callbacks that run when the reactor is read by it's Value()
// method.
type ReadReactor interface {
	Variable
	// Adds a callback that runs instantly
	AddReadCallback(ReadCallback)
	// Adds a callback that runs in a goroutine
	AddAsyncReadCallback(ReadCallback)
	// Adds a callback that runs at some point in the future
	AddConcurrentReadCallback(ReadCallback)
	// Adds a callback that only runs if the comparison function returns true
	AddConditionalReadCallback(ReadCallback, func(Comparable) bool)
}

// WriteReactor is an interface that defines the basic methods that enable 
// various form of callbacks that run when the reactor is written to by it's 
// SetValue() method.
type WriteReactor interface {
	Variable
	// Adds a callback that runs instantly
	AddWriteCallback(WriteCallback)
	// Adds a callback that runs in a goroutine
	AddAsyncWriteCallback(WriteCallback)
	// Adds a callback that runs at some point in the future
	AddConcurrentWriteCallback(WriteCallback)
	// Adds a callback that only runs if the comparison function returns true.
	// The value will still change.
	AddConditionalWriteCallback(WriteCallback, func(Comparable, Comparable) bool)
}

// ReadWriteReactor is an interface that groups WriteReactor and ReadReactor 
// methods. This is the intended base interface, and it's the one that is used 
// for Binder.
type ReadWriteReactor interface {
	ReadReactor
	WriteReactor
	addWatcher(binding)
}

// BindingFunc represents a function that takes a value, and returns a value 
// based on that function
type BindingFunc func(Comparable) Comparable

// Binder is an interface that represents a value that is dependent on another 
// value, and changes when that value changes. The value of a Binder is 
// guaranteed to be eventually correct, no matter the binding type.
type Binder interface {
	ReadWriteReactor
	// Adds a binding that sets the value based on the result of the function
	AddBinding(ReadWriteReactor, BindingFunc)
	// Adds a binding that performs basic, unconditional assignment
	AddSimpleBinding(ReadWriteReactor)
	// Adds a binding that only runs when the value is requested
	AddDelayedBinding(ReadWriteReactor, BindingFunc)
	// Adds a binding that runs eventually
	AddConcurrentBinding(ReadWriteReactor, BindingFunc)
}

type binding struct {
	// This needs to be a pointer
	bound ReadWriteReactor
	bindingFunc BindingFunc
	concurrent bool
}


