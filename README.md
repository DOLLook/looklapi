# micro-webapi
micro-webapi是一个基于kataras/iris构建的轻量级微服务开发脚手架项目，旨在帮助快速搭建基于http的微服务项目。
* 项目集成了mongodb，redis，并实现了redis基本操作和分布式锁操作。
* 使用rabbitmq作为消息中间件，实现了rabbitmq连接池，以及消息重试策略，可以根据业务快速定义消费者。
* 使用consul作为服务发现，实现了服务注册、健康检查和健康服务缓存刷新，通过简单配置即可进行服务注册和获取。

项目不限制web框架，开发者也可以方便地接入gin等web框架处理http请求。