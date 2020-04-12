package reactor

type Comparable interface {
	Compare() int // returns -1 for less than, 0 for equal, 1 for greater than
	Value() interface{}
}

type Callback func(Watcher, Comparable, Comparable)

type Watcher interface {
	SetValue()
	Value() interface{}
	AddCallback(Callback)
	AddAsyncCallback(Callback)
	AddConcurrentCallback(Callback)
	AddConditionalCallback(Callback, func(Comparable, Comparable)bool)
}

type Reactor inteface {
	Watcher
	AddBinding(*Watcher, Callback)
	AddSimpleBinding(*Watcher, Callback)
	AddDelayedBinding(*Watcher, Callback)
	AddConcurrentBinding(*Watcher, Callback)
}
