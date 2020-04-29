package slice

import (
	"github.com/KellenWatt/reactor"
)

// SliceReadCallback takes a function with an []interface{} parameter and wraps 
// it in a reactor.ReadCallback. This is to simplify writing read callbacks for 
// Trigger in this package.
func SliceReadCallback(f func([]interface{})) reactor.ReadCallback {
    return func(v interface{}) {
        val := v.([]interface{})
        f(val)
    }
}

// SliceWriteCallback takes a function with two []interface{} parameters and 
// wraps it in a reactor.WriteCallback. This is to simplify writing write 
// callbacks for Trigger in this package.
func SliceWriteCallback(f func([]interface{}, []interface{})) reactor.WriteCallback {
    return func(prev, v interface{}) {
        prevVal := prev.([]interface{})
        val := v.([]interface{})
        f(prevVal, val)
    }
}

// IndexReadCallback takes a function with a int and interface{} paramters and 
// wraps it in a ReadCallback. This function decomposes an Index struct and 
// passes its members as arguments to f.
func IndexReadCallback(f func(int,interface{})) reactor.ReadCallback {
    return func(v interface{}) {
        val := v.(Index)
        f(val.Key, val.Value)
    }
}

// IndexWriteCallback takes a function with two int and interface{} paramter 
// pairs and wraps it in a WriteCallback. This function decomposes previous 
// and new value Index structs and passes their members as arguments to f in 
// a similar order to standard WriteCallbacks.
func IndexWriteCallback(f func(int,interface{},int,interface{})) reactor.WriteCallback {
    return func(prev, v interface{}) {
        prevVal := prev.(Index)
        val := v.(Index)

        f(prevVal.Key, prevVal.Value, val.Key, val.Value)
    }
}
