package client

var callers = map[string]func(host, arg string) error{
	"foo": Foo,
	"bar": Bar,
}

func Call(name, host, arg string) error {
	return callers[name](host, arg)
}
