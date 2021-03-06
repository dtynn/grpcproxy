package client

var callers = map[string]func(host, arg, cert, hostname string) error{
	"foo":   Foo,
	"bar":   Bar,
	"pulse": Pulse,
}

func Call(name, host, arg, cert, hostname string) error {
	return callers[name](host, arg, cert, hostname)
}
