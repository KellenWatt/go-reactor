package slice

import (
	"github.com/KellenWatt/reactor"
)

func SliceReadCallback(f func([]interface{})) reactor.ReadCallback {
    return func(v interface{}) {
        val := v.([]interface{})
        f(val)
    }
}

func SliceWriteCallback(f func([]interface{}, []interface{})) reactor.WriteCallback {
    return func(prev, v interface{}) {
        prevVal := prev.([]interface{})
        val := v.([]interface{})
        f(prevVal, val)
    }
}

// Helper methods for creating index callbacks, encapsulates more complex functions
func IndexReadCallback(f func(int,interface{})) reactor.ReadCallback {
    return func(v interface{}) {
        val := v.(Index)
        f(val.Key, val.Value)
    }
}

func IndexWriteCallback(f func(int,interface{},int,interface{})) reactor.WriteCallback {
    return func(prev, v interface{}) {
        prevVal := prev.(Index)
        val := v.(Index)

        f(prevVal.Key, prevVal.Value, val.Key, val.Value)
    }
}
