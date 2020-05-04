package dict

import (
	"github.com/KellenWatt/reactor"
)

// MapReadCallback takes a function with a map[interface{}]interface{} 
// parameter and wraps it in a reactor.ReadCallback. This is to simplify 
// writing read callbacks for Trigger in this package.
func MapReadCallback(f func(map[interface{}]interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.(map[interface{}]interface{})
		f(val)
	}
}

// MapWriteCallback takes a function with two map[interface{}]interface{} 
// parameters and wraps it in a reactor.WriteCallback. This is to simplify 
// writing write callbacks for Trigger in this package.
func MapWriteCallback(f func(map[interface{}]interface{}, map[interface{}]interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		p := prev.(map[interface{}]interface{})
		val := v.(map[interface{}]interface{})
		f(p, val)
	}
}

// KeyReadCallback takes a function with two interface{} parameters and wraps 
// it in a ReadCallback. This function decomposes a Pair struct and passes its 
// members as arguments to f.
func KeyReadCallback(f func(interface{}, interface{})) reactor.ReadCallback {
	return func(v interface{}) {
		val := v.(Pair)
		f(val.Key, val.Value)
	}
}

// KeyWriteCallback takes a function with two sets of interface{} parameter 
// pairs and wraps it in a WriteCallback. This function decomposes previous 
// and new value Pair structs and passes their members as arguments to f in 
// a similar order to standard WriteCallbacks.
func KeyWriteCallback(f func(interface{}, interface{}, interface{}, interface{})) reactor.WriteCallback {
	return func(prev, v interface{}) {
		p := prev.(Pair)
		val := v.(Pair)
		f(p.Key, p.Value, val.Key, val.Value)
	}
}


