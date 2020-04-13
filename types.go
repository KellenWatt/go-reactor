// Package reactor gives an API for a basic event system. To facilitate this 
// cleanly, the interfaces provide several different types of responsiveness.
//
// For basic information on thread safety, see 
// https://github.com/KellenWatt/reactor. For more specific information
// see the specific type implementations.
package reactor

type Comparable interface {
	Compare(Comparable) int // returns -1 for less than, 0 for equal, 1 for greater than
	Value() interface{}
}

type Variable interface {
	SetValue()
	Value() Comparable
}

type Callback func(Reactor, Comparable, Comparable)

type Reactor interface {
	Variable
	AddCallback(Callback)
	AddAsyncCallback(Callback)
	AddConcurrentCallback(Callback)
	AddConditionalCallback(Callback, func(Comparable, Comparable) bool)
}

// type Reactor interface {
//     Watcher
//     AddBinding(Watcher, Callback)
//     AddSimpleBinding(Watcher, Callback)
//     AddDelayedBinding(Watcher, Callback)
//     AddAsyncBinding(Watcher, Callback)
//     AddConcurrentBinding(Watcher, Callback)
// }





