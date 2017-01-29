#### grpcproxy
an http2 and http2 only reverse proxy, basically for grpc usage.

##### usage
installation

```
go get -v github.com/dtynn/grpcproxy
go install github.com/dtynn/grpcproxy/...
```


start  
```
grpcproxy run -c path/to/config/file
```

backend examples    
```
gproxy service foo 51001
gproxy service foo 51002
gproxy service bar 51003
gproxy service bar 51004
gproxy service bar 51005
```

request testing  
```
gproxy call "localhost:8000" for test
gproxy call "localhost:8000" for testerror
gproxy call "localhost:8000" bar test
gproxy call "localhost:8000" bar testerror
```

##### TODO
- TLS support for both frontend and backend ✅
- load balance policies ✅
- failure policies
- log options
- test grpc streaming request
- ...


