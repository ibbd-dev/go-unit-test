# go-unit-test
golang单元测试，使用goconvey对各个项目的单元测试的结果进行可视化

## Install

```sh
# 获取
go get -u github.com/ibbd-dev/go-unit-test

# 配置
cd /path/to/go-unit-test
cp env.go.example env.go
vim env.go

# 启动
go build
./go-unit-test
```

# Example

- 浏览器访问：http://localhost:8188/tools-float/show　，其中`tools-float`是配置中的项目名
- 上面会输出一个地址，访问该地址即可

