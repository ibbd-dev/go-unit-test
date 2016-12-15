# go-unit-test
golang单元测试，使用goconvey对各个项目的单元测试的结果进行可视化

注：同一时间只允许一个项目在做单元测试。

## Install

```sh
# 获取
go get -u github.com/ibbd-dev/go-unit-test

# 配置
# 主要配置host, 端口号, 项目, 启动单元测试项目的有效期（过期会自动关闭）等
# 默认主程序端口号为8180，单元测试项目的端口号为8181
cd /path/to/go-unit-test
cp env.go.example env.go
vim env.go

# 编译并启动
# 如果启动不成功，可能是端口号已经被占用了
go build
./go-unit-test
```

# 使用

- 浏览器访问：http://localhost:8180/index ，根据配置`env.go`中的配置信息，这里会显示相应的项目的启动方式。
- 点击其中其中一个项目，会先新窗口打开页面，并会自动跳转到目标页面

