#services
a drop-in services discovery library based on etcd

# pb.go产生方式
protoc ./*.proto --go_out=plugins=grpc:./

#etcd目录结构
etcd目录结构采用 http://gliderlabs.com/registrator/latest/ 提供的结构:          

>    /backends/service_xxx/service_id ---> ip:port

- 调用 Init(..) 将服务发现限定在给定范围
- 添加环境变量 SERVICE_NAME 指定第二栏中的service名称
- 添加环境变量 SERVICE_ID 指定第三栏中的service_id