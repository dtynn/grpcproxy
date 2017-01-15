#### GrpcProxy
只为反向代理 grpc 请求  

##### usage
安装  
```
go get -v github.com/dtynn/grpcproxy
go install github.com/dtynn/grpcproxy/...
```


启动  
```
grpcproxy run -c path/to/config/file
```

后端示例  
```
gproxy service foo 51001
gproxy service foo 51002
gproxy service bar 51003
gproxy service bar 51004
gproxy service bar 51005
```

测试  
```
gproxy call "localhost:8000" for test
gproxy call "localhost:8000" for testerror
gproxy call "localhost:8000" bar test
gproxy call "localhost:8000" bar testerror
```

##### TODO
- 日志


