# bookstore

## 初始化工具

```shell
# 安装 mysql
docker pull mysql

docker run -d \
--name mysql5.7 \
-v $DOCKERPLACE/mysql-5.7/conf:/etc/mysql/conf.d \
-v $DOCKERPLACE/mysql-5.7/logs:/logs \
-v $DOCKERPLACE/mysql-5.7/data:/var/lib/mysql \
-e MYSQL_ROOT_PASSWORD=root \
-p 3307:3306 \
--net fix-net \
--ip 192.168.0.3 \
mysql:5.7

docker exec -it mysql5.7 bash -c 'mysql -h192.168.0.3 -uroot -proot -e "select 1 +1 as a"'
# +---+
# | a |
# +---+
# | 2 |
# +---+
```

```shell
# 安装 redis
docker pull redis

docker run -itd \
--name redis \
-p 6379:6379 \
redis

docker exec -it redis bash -c 'redis-cli dbsize'
# (integer) 1
```
```shell
# 安装 etcd
docker pull quay.io/coreos/etcd

docker run -itd \
-p 2379:2379 \
--name etcd quay.io/coreos/etcd /usr/local/bin/etcd \
--advertise-client-urls http://0.0.0.0:2379 \
--listen-client-urls http://0.0.0.0:2379

docker exec -it etcd sh -c 'etcdctl set /testdir/testkey "Hello world"'
docker exec -it etcd sh -c 'etcdctl get /testdir/testkey'
# Hello world
```

```shell
# 安装protoc
## 下载并解压
https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protobuf-all-3.14.0.tar.gz
## 源码安装
cd protobuf-3.7.0/
./configure --prefix=/usr/local/protobuf
sudo make 
sudo make install

## 配置环境
vi ~/.bash_profile
export PROTOBUF=/usr/local/protobuf
export PATH=$PROTOBUF/bin:$PATH

source ~/.bash_profile

## 验收
protoc --version
# libprotoc 3.14.0
```

```shell
# 安装 protoc-gen-go 跨语言交互中间层
GOPROXY=https://goproxy.cn/,direct go get -u github.com/golang/protobuf/protoc-gen-go
# 安装 goctl 命令行工具
GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u github.com/tal-tech/go-zero/tools/goctl
```

## 初始化项目

```shell
# 创建项目目录
mkdir bookstore && cd bookstroe

# 初始化 go.mod 文件
go mod init bookstore

# 分目录组织模块
mkdir -p api rpc/{add,check,model}

# bookstore
# 	|--go.mod
# 	|--api
# 	|--rpc
# 		  |--add
# 		  |--check
# 		  |--model
```

## HTTP 网关（壳）

### 初始化 API 壳

```shell
# 切入api目录
cd api

# 生成 Api Gateway 的壳
goctl api -o bookstore.api

# 填充 api 逻辑，为 go-zero 的 type 语法，与 golang 基本一致
vi bookstore.api
```

### 编写 API 代码

```go
// 添加（请求与响应）
type (
    addReq {
        book  string `form:"book"` -- 表单参数
        price int64  `form:"price"`
    }

    addResp {
        ok bool `json:"ok"` -- json
    }
)

// 检查是否存在（请求与响应）
type (
    checkReq {
        book string `form:"book"`  -- 表单参数
    }

    checkResp {
        found bool  `json:"found"`  -- json
        price int64 `json:"price"`
    }
)

// 网关 api 
service bookstore-api {  -- 服务名称
    @handler AddHandler  -- 处理器
    get /add(addReq) returns(addResp) -- 请求方法、请求 URI 、请求参数、返回值

    @handler CheckHandler
    get /check(checkReq) returns(checkResp)
}
```

### 自动生成业务

```shell
# goctl api 语言 -api API描述文件 -dir 代码存储路径
goctl api go -api bookstore.api -dir .

# api
# ├── bookstore.api                  // api定义
# ├── bookstore.go                   // main入口定义
# ├── etc
# │   └── bookstore-api.yaml         // 配置文件(URI、host、port)
# └── internal
#     ├── config
#     │   └── config.go              // 定义配置
#     ├── handler
#     │   ├── addhandler.go          // 实现addHandler
#     │   ├── checkhandler.go        // 实现checkHandler
#     │   └── routes.go              // 定义路由处理
#     ├── logic											 // TODo 需要自己实现业务逻辑
#     │   ├── addlogic.go            // 实现AddLogic
#     │   └── checklogic.go          // 实现CheckLogic
#     ├── svc
#     │   └── servicecontext.go      // 定义ServiceContext
#     └── types
#         └── types.go               // 定义请求、返回结构体
```

### http 调用测试

```shell
go run bookstore.go -f etc/bookstore-api.yaml
# Starting server at 0.0.0.0:8888...

curl -i "http://localhost:8888/check?book=go-zero"
# HTTP/1.1 200 OK
# Content-Type: application/json
# Date: Mon, 21 Dec 2020 02:24:06 GMT
# Content-Length: 25
# 
# {"found":false,"price":0}% 
```

## RPC 服务（壳）

### Add RPC

```shell
# 切入 rpc/add 目录
cd rpc/add

# 通过命令生成 rpc proto 模板
goctl rpc template -o add.proto

# 填充proto 逻辑
vi add.proto
```

```protobuf
syntax = "proto3";

package add;

message addReq {
    string book = 1;
    int64 price = 2;
}

message addResp {
    bool ok = 1;
}

service adder {
    rpc add(addReq) returns(addResp);
}
```

```shell
# 基于 proto 生成 rpc 壳
goctl rpc proto -src add.proto -dir .

#rpc/add
# ├── add
# │   └── add.pb.go
# ├── add.go                      // rpc服务main函数
# ├── add.proto                   // rpc接口定义
# ├── adder
# │   └── adder.go                // 提供了外部调用方法，无需修改
# ├── etc
# │   └── add.yaml                // 配置文件
# └── internal
#     ├── config
#     │   └── config.go           // 配置定义
#     ├── logic
#     │   └── addlogic.go         // add业务逻辑在这里实现
#     ├── server
#     │   └── adderserver.go      // 调用入口, 不需要修改
#     └── svc
#         └── servicecontext.go   // 定义ServiceContext，传递依赖
```

```shell
# 尝试运行 add rpc
go run add.go -f etc/add.yaml
# Starting rpc server at 127.0.0.1:8080...
```

### Check RPC

```shell
# 切入 rpc/check 目录
cd rpc/check

# 通过命令生成 rpc proto 模板
goctl rpc template -o check.proto

# 填充proto 逻辑
vi check.proto
```

```protobuf
syntax = "proto3";

package check;

message checkReq {
  string book = 1;
}

message checkResp {
  bool found = 1;
  int64 price = 2;
}

service checker {
  rpc check(checkReq) returns(checkResp);
}
```

```shell
# 基于 proto 生成 rpc 壳
goctl rpc proto -src check.proto -dir .

# rpc/check
# ├── check
# │   └── check.pb.go
# ├── check.go                            // rpc服务main函数
# ├── check.proto                         // rpc接口定义
# ├── checker
# │   └── checker.go                      // 提供了外部调用方法，无需修改
# ├── etc
# │   └── check.yaml                      // 配置文件
# └── internal
#     ├── config
#     │   └── config.go                   // 配置定义
#     ├── logic
#     │   └── checklogic.go               // check业务逻辑在这里实现
#     ├── server
#     │   └── checkerserver.go            // 调用入口, 不需要修改
#     └── svc
#         └── servicecontext.go           // 定义ServiceContext，传递依赖
```

```shell
# check rpc 端口修改为 8081 与 add rpc 错开
vi etc/check.yaml
```

```shell
# 尝试运行 
go run check.go -f etc/check.yaml
# Starting rpc server at 127.0.0.1:8081...
```

## Model 模型

```shell
cd rpc/model

# 编写 sql
vi book.sql
```

```sql
CREATE TABLE `book`
(
  `book` varchar(255) NOT NULL COMMENT 'book name',
  `price` int NOT NULL COMMENT 'book price',
  PRIMARY KEY(`book`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

```sql
-- sql 命令行建库
create database gozero;

-- 建表
source book.sql;
```

```shell
# 基于 ddl 语句生成模型
goctl model mysql ddl -c -src book.sql -dir .

# 或 基于 url 生成模型层
# goctl model mysql datasource -url="user:password@tcp(127.0.0.1:3306)/database" -table="table1,table2"  -dir="./model"
goctl model mysql datasource -url="root:root@tcp(127.0.0.1:3306)/gozero" -table="book"  -dir="."

# rpc/model
# ├── book.sql
# ├── bookmodel.go     // CRUD+cache代码
# └── vars.go               // 定义常量和变量
```

## RPC结合DB、Cache（核）

### yaml配置

```shell
# 分别在 check.yaml 和 add.yaml 中增加下面配置
vi rpc/check/etc/check.yaml
vi rpc/add/etc/add.yaml
```

```yaml
DataSource: root:root@tcp(localhost:3307)/gozero
Table: book
Cache:
  - Host: localhost:6379
```

### config.go

```shell
# 分别在 check 和 add 中增加下面配置
vi rpc/check/internal/config.go
vi rpc/add/internal/config.go
```

```go
type Config struct {
    zrpc.RpcServerConf
    DataSource string             // 手动代码
    Table      string             // 手动代码
    Cache      cache.CacheConf    // 手动代码
}
```

### servicecontext.go

```shell
# 分别在 check 和 add 中增加下面代码
vi rpc/check/internal/svc/servicecontext.go
vi rpc/add/internal/svc/servicecontext.go
```

```go
type ServiceContext struct {
	c config.Config
	Model model.BookModel   // 手动代码
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		c: c,
		Model: model.NewBookModel(sqlx.NewMysql(c.DataSource), c.Cache, c.Table), // 手动代码
	}
}
```

### logic

#### checklogic.go

```go
func (l *CheckLogic) Check(in *check.CheckReq) (*check.CheckResp, error) {
  // 以下为新增
	resp, err := l.svcCtx.Model.FindOne(in.Book)
	if err != nil {
		return nil, err
	}

	return &check.CheckResp{
		Found: true,
		Price: resp.Price,
	}, nil
    // 以上为新增
}
```

#### addlogic.go

```go
func (l *AddLogic) Add(in *add.AddReq) (*add.AddResp, error) {
	// todo: add your logic here and delete this line
	_, err := l.svcCtx.Model.Insert(model.Book{
		Book:  in.Book,
		Price: in.Price,
	})
	if err != nil {
		return nil, err
	}

	return &add.AddResp{
		Ok: true,
	}, nil
}
```

## HTTP 网关结合 RPC 服务（核）

#### yaml 配置

```shell
# 添加如下代码（基于 etcd 注册发现 add、check 服务）
vi api/etc/bookstore-api.yaml
```

```yaml
Add:
  Etcd:
    Hosts:
      - localhost:2379
    Key: add.rpc
Check:
  Etcd:
    Hosts:
      - localhost:2379
    Key: check.rpc
```

#### config.go

```shell
vi internal/config/config.go
```

```go
type Config struct {
    rest.RestConf
    Add   zrpc.RpcClientConf     // 手动代码
    Check zrpc.RpcClientConf     // 手动代码
}
```

#### servicecontext.go

```shell
vi internal/svc/servicecontext.go
```

```go
type ServiceContext struct {
	Config config.Config
	Adder   adder.Adder          // 手动代码
	Checker checker.Checker      // 手动代码
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Adder:   adder.NewAdder(zrpc.MustNewClient(c.Add)),         // 手动代码
		Checker: checker.NewChecker(zrpc.MustNewClient(c.Check)),   // 手动代码
	}
}
```

#### logic

##### addlogic.go

```shell
vi internal/logic/addlogic.go
```

```go
func (l *AddLogic) Add(req types.AddReq) (*types.AddResp, error) {
  // 以下为新增
	resp, err := l.svcCtx.Adder.Add(l.ctx, &adder.AddReq{
		Book:  req.Book,
		Price: req.Price,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResp{
		Ok: resp.Ok,
	}, nil
  // 以上为新增
}
```

##### checklogic.go

```shell
vi internal/logic/checklogic.go
```

```go
func (l *CheckLogic) Check(req types.CheckReq) (*types.CheckResp, error) {
  // 以下为新增
	resp, err := l.svcCtx.Checker.Check(l.ctx, &checker.CheckReq{
		Book:  req.Book,
	})
	if err != nil {
		return nil, err
	}

	return &types.CheckResp{
		Found: resp.Found,
		Price: resp.Price,
	}, nil
  // 以上为新增
}
```

## 调用测试

```shell
# 启动 rpc 服务
go run check.go -f etc/check.yaml
go run add.go -f etc/add.yaml

# 启动 http 网关
go run bookstore.go -f etc/bookstore-api.yaml

# 插入
curl -i "http://localhost:8888/add?book=go-zero&price=10"
# {"ok":true}

# 查询
curl -i "http://localhost:8888/check?book=go-zero"
# {"found":true,"price":10}
```

## 压测

```shell
brew install wrk

# 10个线程，创建1000个连接，持续压测时间 30 秒
wrk -t10 -c1000 -d30s --latency "http://localhost:8888/check?book=go-zero"
# Running 30s test @ http://localhost:8888/check?book=go-zero
#   10 threads and 1000 connections
#   Thread Stats   Avg      Stdev     Max   +/- Stdev
#     Latency    51.61ms   25.08ms 218.08ms   68.55%
#     Req/Sec     1.95k   351.74     5.39k    74.38%
#   Latency Distribution
#      50%   52.60ms
#      75%   67.62ms
#      90%   81.64ms
#      99%  115.06ms
#   580103 requests in 30.10s, 73.58MB read
#   Socket errors: connect 0, read 3294, write 1, timeout 0
# Requests/sec:  19272.12
# Transfer/sec:      2.44MB

580103/30.10 约等于 19272 ，及 QPS 为 19272
```

