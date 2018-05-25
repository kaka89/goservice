# goservice
Discovery and register components implements with go and etcd.
基于 go 和 etcd 实现的服务注册、发现组件，可应用于微服务的开发（如果要自个儿实现这块的东西的话）

# register
服务注册，当服务启动的时候，应该同时启动注册功能。之后，注册服务会不断的刷新自己的注册内容。
示例：
```
register.Register("TestService2", "192.168.0.2", 8080, endpoints, 10*time.Second, 11)
```
参数说明：
- serviceName：服务名，用于服务发现用
- 服务本地地址：可以是内网地址，也可以只外网地址，甚至可以是域名，取决于服务的部署以及调用情况。一般而言，只有内网地址
- 服务本地端口：服务监听的端口
- etcd endpoints：etcd 服务的地址
- refreshInterval： 注册内容的刷新频率
- keyTTL： 内容在etcd上保留的时间，注册时，会保证 ttl 大于 refresh 的时间，否则会出现空置



# discovery
调用一个服务前，从 etcd 处获取相应的服务地址。负载均衡采用本地随机。本地会缓存服务的地址，并监听服务注册内容的变化，优先从本地缓存中读取地址.


