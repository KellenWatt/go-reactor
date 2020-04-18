package reactor

type readConState struct {
    value interface{}
    f ReadCallback
}

type writeConState struct {
    prev, value interface{}
    f WriteCallback
}

type conRead chan readConState
type conWrite chan writeConState

func runConcurrentRead() {
    for _,c := range conRead {
        c.f(c.value)
    }
}

func runConcurrentWrite() {
    for _,c := range conWrite {
        c.f(c.value)
    }
}
func makeAsyncWrite(w WriteCallback) WriteCallback {
    return func(p, v interface{}) {
        go w(p,v)
    }
}

func makeConcurrentRead(r ReadCallback) ReadCallback {
    return func(v interface{}) {
        conRead <- readConState{v, r}
    }
}

func makeConcurrentWrite(w WriteCallback) WriteCallback {
    return func(p, v interface{}) {
        conWrite <- writeConState{p,v,w}
    }
}

func makeConditionalRead(r ReadCallback, f func(interface{})bool) ReadCallbcak {
    return func(v interface{}) {
        if f(v) {
            r(v)
        }
    }
}

func makeConditionalWrite(w WriteCallback, f func(interface{}, interface{})bool) WriteCallback {
    return func(p, v interface{}) {
        if f(p,v) {
            w(p,v)
        }
    }
}
