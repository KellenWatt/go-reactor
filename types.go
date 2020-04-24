// Package reactor implements callbacks and bindings, similar to that of an 
// event loop-based program.
//
// Callbacks are defined as functions that are run once a triggering condition 
// has been met, and do not return any values. There are four types of 
// callbacks provided by this package, and are as follows: basic, asynchronous, 
// concurrent, and conditional. A basic callback is one that is called 
// immediately upon the triggering condition. Asynchronous callbacks are run 
// immediately using a goroutine. Concurrent callbacks will run eventually, 
// but there's no promise of when they're run. Finally, conditional callbacks 
// are similar to basic callbacks, but are only run when a given condition is 
// met. For more information on any of these, see the specific methods.
//
// Bindings are defined as tying two variables together, with one variable 
// depending on the value of the other. A variable that is bound to another can
// have its value set independently of the variable its bound to, but it will 
// be updated when the other variable changes. There are three bindings offered
// by this package: basic, delayed, and concurrent. Basic bindings are 
// executed immediately upon the bound-to variable changing. A delayed binding
// is only executed when the bound variable calls its Value method. Finally, 
// a concurrent binding is executed such that the bound variable will be 
// updated eventually, but not necessarily immediately. For more information 
// on any of these, see the specific methods.
//
// Since both callbacks and bindings offer methods that are or can be executed
// asynchronously, each Trigger and Indicator instance offers a mutex or 
// similar functionality to ensure thread safety on operations that affect its 
// internal values. This is provided as a convenience. No methods of Trigger 
// or Indicator are thread-safe and should not be treated as such.
//
// Aside from ReadWriteBinder, ReadWriteInitiator, and Initiator, all 
// interfaces provided by this package are provided as a convenience.
package reactor


// ReadCallback is the function type used in all read callbacks
type ReadCallback func(interface{})
// WriteCallback is the function type used in all write callbacks
type WriteCallback func(interface{}, interface{})
// BindingFunc is the function type used in all bindings
type BindingFunc func(interface{}) interface{}


// Initiator is the interface that defines the minimum functions required 
// to have a functional and still (mostly) generic callback system.
// 
// Value returns the value respresented by the Initiator.
//
// SetValue sets the value of the Initiator.
//
// AddBinder registers a binding with the Initiator instance. This is 
// largely a convenience method used by Binders to keep interfaces fully public, 
// and while its use is permitted, it is heavily discouraged for the purposes 
// of readability and predicatability.
type Initiator interface {
	Value() interface{}
	SetValue(interface{}) 
	AddBinder(Binder, BindingFunc, bool)
}

// ReadInitiator is the interface that defines various callback methods to 
// respond to calls to Value. Each method registers a callback with the 
// ReadInitiator. For more information, see documentations for implementations.
type ReadInitiator interface {
	Initiator
	AddReadCallback(ReadCallback)
}

// WriteInitiator is the interface the defines various callback methods to 
// respond to calls to SetValue. Each method registers a callback with the 
// WriteInitiator. For more information, see documentation for implementations.
type WriteInitiator interface {
	Initiator
	AddWriteCallback(WriteCallback)
}

// ReadWriteInitiator is the interface that groups methods from ReadInitiator 
// and WriteInitiator for convenience.
type ReadWriteInitiator interface {
	ReadInitiator
	WriteInitiator
}

// Binder is the interface that defines various bindings that determine the 
// value of the Binder, based on the value of an Initiator. Each method 
// registers a binding with the necessary party. For more information, see 
// documention for implementations.
type Binder interface {
	Initiator
	AddBinding(Initiator, BindingFunc)
	AddTrivialBinding(Initiator)
	AddDelayedBinding(Initiator, BindingFunc)
	AddConcurrentBinding(Initiator, BindingFunc)
}

// ReadBinder is the interface that combines ReadInitiator and Binder methods 
// for convenience.
type ReadBinder interface {
	ReadInitiator
	Binder
}

// WriteBinder is the interface that combines WriteInitiator and Binder methods 
// for convenience.
type WriteBinder interface {
	WriteInitiator
	Binder
}

// ReadWriteBinder is the interface that combines ReadBinder and WriteBinder 
// methods for convenience. This is the most comprehensive interface and 
// represents the combined functionality of all other interfaces in the 
// reactor package.
type ReadWriteBinder interface {
	ReadBinder
	WriteBinder
}


type binding struct {
	source, binder Initiator
	f BindingFunc
	concurrent bool
}
