package reactor



type ReadCallback func(interface{}) interface{}
type WriteCallback func(interface{}, interface{}) interface{}
type BindingFunc func(interface{}) interface{}

type Reactor interface {
	Value() interface{}
	SetValue(interface{}) 
	AddBinder(Binder, BindingFunc)
}

type ReadReactor interface {
	Reactor
	AddReadCallback(ReadCallback)
	AddAsyncReadCallback(ReadCallback)
	AddConcurrentReadCallback(ReadCallback)
	AddConditionalReadCallback(ReadCallback, func(interface{}) bool)
}

type WriteReactor interface {
	Reactor
	AddWriteCallback(WriteCallback)
	AddAsyncWriteCallback(WriteCallback)
	AddConcurrentWriteCallback(WriteCallback)
	AddConditionalWriteCallback(WriteCallback, func(interface{}) bool)
}

type ReadWriteReactor interface {
	ReadReactor
	WriteReactor
}

type Binder interface {
	Reactor
	AddBinding(Reactor, BindingFunc)
	AddSimpleBinding(Reactor)
	AddDelayedBinding(Reactor, BindingFunc)
	AddConcurrentBinding(Reactor, BindingFunc)
}

type ReadBinder interface {
	ReadReactor
	Binder
}

type WriteBinder interface {
	WriteReactor
	Binder
}

type ReadWriteBinder interface {
	ReadBinder
	WriteBinder
}


type binding struct {
	source, binder Reactor
	f BindingFunc
}
