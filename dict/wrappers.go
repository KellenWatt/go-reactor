package dict

import (
	"github.com/KellenWatt/reactor"
)

func MapReadCallback(f func(map[interface{}]interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.(map[interface{}]interface{})
		f(val)
	}
}

func MapWriteCallback(f func(map[interface{}]interface{}, map[interface{}]interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		p := prev.(map[interface{}]interface{})
		val := v.(map[interface{}]interface{})
		f(p, val)
	}
}

func KeyReadCallback(f func(interface{}, interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.(Pair)
		f(val.Key, val.Value)
	}
}

func KeyWriteCallback(f func(interface{}, interface{}, interface{}, interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		p := prev.(Pair)
		val := v.(Pair)
		f(p.Key, p.Value, val.Key, val.Value)
	}
}


