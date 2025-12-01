## 任务派发系统开发规格文档（基于glue v0.8.4）

## 1. 项目概述
本系统是一个任务派发系统，主要包含以下核心功能模块：
- 后台管理系统：用于对任务进行统一的管理和配置。
- 分布式任务执行器：独立运行的任务执行模块，支持分布式部署和执行。
- 通知机制：任务完成后，根据预设配置向指定接收者发送通知。
- 日志追踪：提供任务执行过程和结果的详细日志查看功能。
基于自研`glue v0.8.4`实现，遵循Golang最佳实践与框架规范。

## 2. 技术约束
### 2.1 框架与语言
- 语言版本：Go 1.24+
- 前端开发：reactjs 6+, antd-pro
- 核心框架：glue v0.8.4（模块路径：`github.com/zhiyunliu/glue`）
- 禁止使用其他Web框架（如Gin、Echo）或ORM工具（如GORM）

### 2.2 依赖管理
```go
// go.mod必须包含
require github.com/zhiyunliu/glue v0.8.4
```

## 3. 代码结构规范
### 3.1 目录结构（强制遵循）
```
根目录/
├── constants/                # HTTP接口层
│   ├── enums/                # 枚举定义
│   │   └── userstatus/       # 具体枚举定义
│   │       └── userstatus.go  # 用户状态枚举
│   ├── cachekey/              # 缓存定义
│   │   └── cachekey.go        # 缓存key定义
│   └── queuekey/                # 消息队列定义
│       └── queuekey.go        # 消息队列key定义
├── services/                # HTTP接口层
│   ├── user/                # 平台服务
│   │   ├──user.go           # 用户服务
│   │   └──xxx.go      # xxx服务
│   └── services.go      # 服务注册
├── modules/            # 业务逻辑层
│   ├── sqls/           
│   │   └──user.go      # 数据库sql脚本
│   └── user.go
├── models/             # 请求/响应结构体 
│   └──user.go # 请求/响应结构体
├── config/
│   └── config.go       # 配置定义
├── frontend/                # 前端代码
├── go.mod
├── .gitlab-ci.yml       # GitLab CI/CD配置
├── main.go     # 项目入口
└── go.sum


```

### 3.2 命名规则
- 包名：小写单数（如`api`、`service`）
- 文件命名：`xxx.go`/`xxx.yyy.go`
- 结构体命名：大驼峰（如`UserHandler`、`UserService`）
- 函数命名：大驼峰（如`GetUserInfo`、`CreateUser`）

## 4. 核心功能规格
### 4.1 接口列表
| 接口路径       | HTTP方法 | 功能描述       | 请求参数          | 响应数据          |
|----------------|----------|----------------|-------------------|-------------------|
| `/users/info?id={id}`  | GET      | 查询单个用户   | query参数`id`      | `UserInfoResp`    |
| `/users/list`  | GET     | 查询用户列表    | `UserListReqParam` | 切片数据`UserListResp`    |
| `/users/create`  | POST      | 创建用户       |`CreateUserReqParam` | `UserInfoResp` |
| `/users/update`  | POST      | 更新用户       |`UpdateUserReqParam` | `UserInfoResp` |
| `/users/delete?id={id}`  | POST  | 删除用户       |  query参数`id`      | resp.Success() |

#### 4.1.1 接口参数
 - 请求参数: 请求参数都必须有结构体
 - 响应参数: 响应有两个方式进行返回 1. resp.Success() ,2. resp.Error() 。 如：resp.Success(UserInfoResp)



### 4.2 静态资源定义
#### 4.2.1 静态资源定义（constants/cachekey/cachekey.go）
```go
package cachekey

type CacheItem struct {
	Key     string
	Timeout int
}

var (
	UserInfo = &CacheItem{Key: "example:userinfo:@{user_id}", Timeout: 60}//timeout:秒
)

```

#### 4.2.2 枚举定义（constants/enums/userstatus/userstatus.go）
```go
package userstatus

type UserStatus int
const (
	Normal = 0
	Disable = 1
	Deleted = 2
)

```

#### 4.2.3 消息队列定义（constants/queuekey/queuekey.go）
```go
package queuekey

const (
	UserInfoCheckStatus              = "example:userinfo:check-status"
)

```


### 4.3 数据模型定义
#### 4.3.1 业务数据实体（models/user.go）
```go
package models

// CreateUserReqParam 创建用户请求
type CreateUserReqParam struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
}

// UserListReqParam 用户列表请求
type UserListReqParam struct {
    Status int    `json:"status" validate:"oneof=0 1"`
}

// UpdateUserReqParam 更新用户请求
type UpdateUserReqParam struct {
    Email  string `json:"email" validate:"email"`
    Status int    `json:"status" validate:"oneof=0 1"`
}

// UserInfoResp 用户响应
type UserInfoResp struct {
    ID        uint64 `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Status    int    `json:"status"`
    CreatedAt string `json:"created_at"`
}
```


### 4.4 配置文件定义
#### 4.4.1 配置对象（config/config.go）
```go
package config

import (
	"192.168.1.27/micro-basic/modules/config"
	"github.com/zhiyunliu/glue/engine"
)

type Config struct {
	config.Config
	RpcName      string          `json:"rpc_name"`
}

var Sys = &Config{
	Config: config.Sys,
	RpcName:"default",
}

```
#### 4.4.2 配置文件（config.json）

```json
{
	"app":{                /*app程序说明*/
		"mode":"release", /*启动模式：release,debug*/
		"ip_mask":"192.168" /*ip地址段*/
	},
	"registry":"nacos://default",	 /*注册中心*/
	"dbs":{ /*数据库配置*/
		"microsql":{"proto":"sqlserver","conn":"server=mssqldb.example.com;database=demo;user id=@{DB_MSSQL_ACCOUNT};password=@{DB_MSSQL_PWD};","max_open":10,"max_idle":10,"life_time":100}
	},
	"rpcs":{ /*rpc配置*/
		"default":{"proto":"grpc" }
	},
	"xhttp":{ /*http配置*/
		"default":{"balancer":"wrr","conn_timeout":10 }
	},
	"caches":{ /*缓存配置*/
		"default":{"proto":"redis","addr":"redis://localhost"}
	},
	"queues":{ /*消息队列配置*/ 
		"default":{"proto":"redis","addr":"redis://localhost"},
		"streamredis":{"proto":"streamredis","addr":"redis://localhost"}
	},
	"redis":{ /*redis配置*/
		"localhost":{
			"addrs":["localhost:6379"],
 			"db":14,
			"dial_timeout":10,
			"read_timeout":10,
			"write_timeout":10,
			"pool_size":20
		}
	},
	"nacos":{ /*nacos配置*/
		"default":{
			"encrypt":false,
			"client":{"namespace_id":""},
			"server":[{"ipaddr":"localhost","port":8848}],
			"options":{"prefix":"","group":"","cluster":"","weight":100}
		}
	},
	"servers":{ /*app服务列表配置*/
		"apiserver":{  /*app服务配置   对应：api.New("apiserver") */
			"config":{"addr":":8084","status":"start","read_timeout":10,"write_timeout":10,"read_header_timeout":10,"max_header_bytes":65535}
		},
		"mqcserver":{ /*app服务配置   对应：mqc.New("mqcserver") */
			"config":{"addr":"queues://streamredis","status":"start"},
			"tasks":[
				{"queue":"example:userinfo:check-status","concurrency":10} /*消息队列处理配置  queue:消息队列名词， concurrency：并发数*/
			]
		},
		"rpcserver":{ /*app服务配置   对应rpc.New("rpcserver") */
			"config":{"addr":":8081","status":"start"}
		}
	}
}

```
配置文件格式说明



* apiserver 对应api.New("apiserver")
* mqcserver 对应mqc.New("mqcserver")
* rpcserver 对应rpc.New("rpcserver")

#### 4.4.2 CI文件（.gitlab-ci.yml）

```yaml
stages:
  - prepare
  - test
  - lint
  - build
  - deploy
  - recover

variables:
  # 项目名称 ---整个需要修改为对应的项目
  PROJECT_NAME: 'example-ms-manager-xxx'
  # 要编译的项目相对目录, 根目录填写: '.', 注意不要带后面的 /
  SRC_DIR: '.'
  # 邮件备注消息
  EMAIL_REMARK: '无备注'
  EMAILSENT: 'dev.list@example.com test.list@example.com ops.list@example.com'
  PROJECT_TITLE: '${PROJECT_NAME}.(${CI_COMMIT_REF_NAME})'
  TMP_PROJECT_DIR: "/tmp/gitlab/${CI_PIPELINE_ID}/${PROJECT_NAME}"
  GIT_STRATEGY: none

before_script:
  - ci-ms-v2 before

after_script:
  - ci-ms-v2 after

prepare:
  stage: prepare
  when: manual
  allow_failure: false
  variables:
    GIT_STRATEGY: clone
  script:
    - ci-ms-v2 prepare

test:
  stage: test
  script:
    - ci-ms-v2 test


lint:
  stage: lint
  script:
    - ci-ms-v2 lint    


build:
  stage: build
  script:
    - ci-ms-v2 make
  artifacts:
    name: "${PROJECT_TITLE}"
    paths:
      - release/

deploy:
  stage: deploy
  dependencies:
    - build
  script:
    - ci-ms-v2 deploy

recover:
  stage: recover
  when: on_failure
  script:
    - sudo chown -R gitlab-runner:gitlab-runner ${CI_PROJECT_DIR}/*

```



## 5. 框架使用规范
### 5.1 启动流程（main.go）
```go
package main

import (
	"context"

	"192.168.1.27/micro-basic/modules/dbconn"
	"192.168.1.27/micro-manager/xxx/config"
	"192.168.1.27/micro-manager/xxx/services"

	"github.com/zhiyunliu/glue"
	_ "github.com/zhiyunliu/glue/contrib/cache/redis"
	_ "github.com/zhiyunliu/glue/contrib/dlocker/redis"
	_ "github.com/zhiyunliu/glue/contrib/metrics/prometheus"
	_ "github.com/zhiyunliu/queue-redis"
	_ "github.com/zhiyunliu/xdb-mssql"
	_ "github.com/zhiyunliu/xdb-mongodb"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"
	_ "github.com/zhiyunliu/glue/contrib/xhttp/http"
	_ "github.com/zhiyunliu/glue/contrib/xrpc/grpc"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server/api"
	"github.com/zhiyunliu/glue/server/rpc"
	"github.com/zhiyunliu/glue/xdb"
)

func main() {
	global.AppName = "xxx"
	//用于处理http的请求
	apiSrv := api.New("apiserver", api.WithServiceName(global.AppName), api.Log(log.WithRequest()))
	services.BindAPI(apiSrv)

	//用于处理rpc的请求
	rpcSrv := rpc.New("rpcserver", rpc.WithServiceName(global.AppName), rpc.Log(log.WithRequest()))
	services.BindRPC(rpcSrv)

	//用于处理mqc请求
	mqcSrv := mqc.New("mqcserver", mqc.WithServiceName(global.AppName), mqc.Log(log.WithRequest()))
	services.BindMQC(mqcSrv)


	opts := make([]glue.Option, 0)
	opts = append(opts, glue.Server(apiSrv, rpcSrv), glue.StartingHook(func(ctx context.Context) error {
		err := global.Config.ScanTo(config.Sys)
		if err != nil {
			return err
		}

		err = dbconn.Refactor( )
		return err
	}), glue.WithConfigSource(config.Sys.GetConfigSource()...))
	app := glue.NewApp(opts...)
	_ = app.Start()
}

```



## 6. 示例代码
### 6.1 services实现（services/user/user.go）
```go
package user

import (
	"192.168.1.27/micro-manager/xxx/models"
	"192.168.1.27/micro-manager/xxx/modules/user"
	"192.168.1.27/microservice/common/resp"
	"github.com/zhiyunliu/glue/context"
)


func GetUserInfo(ctx context.Context) interface{} {
	userId := ctx.Request().Query().Get("id")

	result, err := user.GetUserList(ctx.Context(),userId)
	if err != nil {
		ctx.Log().Errorf("GetUserList.err:%+v", err)
		return resp.Error(err)
	}
	return resp.Success(result)
}



func GetUserList(ctx context.Context) interface{} {
	param := &models.UserListReqParam{}
	if err :=ctx.Bind(&param);err!=nil{
		return resp.Error(err)
	}

	result, err := user.GetUserList(ctx.Context(),param)
	if err != nil {
		ctx.Log().Errorf("GetUserList.err:%+v", err)
		return resp.Error(err)
	}
	return resp.Success(result)
}


func CreateUser(ctx context.Context) interface{} {
	param := &models.CreateUserReqParam{}
	if err :=ctx.Bind(&param);err!=nil{
		return resp.Error(err)
	}

	result, err := user.GetUserList(ctx.Context(),param)
	if err != nil {
		ctx.Log().Errorf("GetUserList.err:%+v", err)
		return resp.Error(err)
	}
	return resp.Success(result)
}
 
func CheckStatus(ctx context.Context) interface{} {
	return resp.Success()
}
 
 
```
* 192.168.1.27/microservice/common/resp 是现有的代码包,直接引用即可不需要生成
* services 中代码的context 均为github.com/zhiyunliu/glue/context 不能和默认的context混淆


### 6.1.1 路由绑定（services/services.go）
```go
package services

import (
	"192.168.1.27/micro-manager/xxx/services/user"
  	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/server/api"
	"github.com/zhiyunliu/glue/server/rpc"
	"github.com/zhiyunliu/glue/server/mqc"
)

func BindAPI(server *api.Server) {
	userGroup := server.Group("/api/manager-xxx/user")
	userGroup.Handle("/get-user-info", user.GetUserInfo,engine.MethodGet)
	userGroup.Handle("/get-user-list", user.GetUserList,engine.MethodGet)

}

func BindRPC(server *rpc.Server) {
	userGroup := server.Group("/rpc/user")
	userGroup.Handle("/create", user.CreateUser,engine.MethodPost)

}
func BindMQC(server *mqc.Server) {
	userGroup := server.Group("/mqc/user")
	userGroup.Handle("/check-status", user.CheckStatus)
}


```






### 6.2 modules实现（modules/user.go）
```go
package user

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"192.168.1.27/micro-manager/xxx/config"
	"192.168.1.27/micro-manager/xxx/modules/sqls"
 	"192.168.1.27/micro-manager/models"
	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/golibs/xtransform"
	"github.com/zhiyunliu/golibs/xtypes"
)

func GetUserInfo(ctx context.Context,userId uint) (info models.UserInfo, err error) {
    info := models.UserInfo{} 
	dbObj := glue.DB(config.Sys.DBexample0725Back)
 	err = dbObj.FirstAs(ctx, sqls.GetUserInfo, map[string]interface{}{
        "id": userId,
    }, &info)
	if err != nil {
		return
	} 
	return info, nil

}


func  GetUserList(ctx context.Context, param  *models.UserListReqParam) (list  []models.UserInfo ,err error) {
	dbObj := glue.DB(config.Sys.DBexample0725Back)
	list = make([]models.UserInfo, 0)
	err = dbObj.QueryAs(ctx, sqls.GetUserList, param, &list)
	if err != nil {
		return
	} 
	return list, nil
}

func  CreateUser(ctx context.Context, param  *models.CreateUserReqParam) (info models.UserInfo ,err  error) {
	dbObj := glue.DB(config.Sys.DBexample0725)
 	result,err := dbObj.Exec(ctx, sqls.CreateUser, param)
	if err != nil {
		return
	} 
	if result.RowsAffected() <= 0 {
		return info, errors.New("创建用户失败")
	}
	info, err = GetUserInfo(ctx,result.LastInsertId())
	return info, nil
}


```
*  modules 中的context 均为系统的context ,不是框架的github.com/zhiyunliu/glue/context


### 6.3 sqls实现（modules/sqls/user.go）

 ```go
package sqls

 const GetUserInfo = `select * from usr_user_info(nolock) t where t.id= @{id}`

 const GetUserList = `
 select 
	id ,
	username ,
	email ,
	status 
 from usr_user_info(nolock) t
 where t.status= @{status}
 `


 const CreateUser = `
 insert into usr_user_info (username,email,status) values (@{username}, @{email}, @{status})
 
 `

```

### 6.4 sqls语句模板处理规则

 #### 6.4.1 参数化支持
```sql
@{field} 
如：
select * from table t where t.name = @{name}  
select * from table t where t.name = @{t.name}  
解析结果：
select * from table t where t.name = @p_name

```

#### 6.4.2 &符合链接

将参数进行and链接，如果参数值不存在或者为空将不会生成and条件
```sql
&{field} ， &{t.field}

如： 
select * from table t where t.id = @{id} &{name} 
select * from table t where t.id = @{id} &{t.name} 

解析结果：
select * from table t where t.id = @p_id and name = @p_name --参数存在
select * from table t where t.id = @p_id and t.name = @p_name --参数存在
或者
select * from table t where t.id = @p_id --参数不存在或者为空,空字符

```


#### 6.4.3 |符合链接

将参数进行or链接，如果参数值不存在或者为空将不会生成or条件

```sql
|{field} 

如：  
select * from table t where t.id = @{id} |{name} 
select * from table t where t.id = @{id} |{t.name} 

解析结果：
 select * from table t where t.id = @p_id or name = @p_name --参数存在
 select * from table t where t.id = @p_id or t.name = @p_name --参数存在
 或者
select * from table t where t.id = @p_id --参数不存在或者为空,空字符

```

#### 6.4.4  like / not like 支持

```sql
&{like field} ，&{like %field}， &{like field%} ，&{like %field%}
&{notlike field} ，&{notlike %field}， &{notlike field%} ，&{notlike %field%}
&{t.field like property} ，&{t.field like %property}， &{t.field like property%} ，&{t.field like %property%}
&{t.field notlike property} ，&{t.field notlike %property}， &{t.field notlike property%} ，&{t.field notlike %property%}
----(|符号类似)

样例： 
select * from table t where t.id = @{id} &{like name} 
select * from table t where t.id = @{id} &{like %name}
select * from table t where t.id = @{id} &{like name%}
select * from table t where t.id = @{id} &{like %name%}
select * from table t where t.id = @{id} &{t.field like %newname%}

解析结果：
select * from table t where t.id = @p_id and name like @p_name
select * from table t where t.id = @p_id and name like '%'+@p_name
select * from table t where t.id = @p_id and name like @p_name+'%'
select * from table t where t.id = @p_id and name like '%'+@p_name+'%'
select * from table t where t.id = @p_id and t.field like '%'+@p_newname+'%'

```

#### 6.4.5  in / not in 支持


```sql
---注意：in表达式只接受数组切片数据,其他类似直接返回空
&{in field} ,&{in t.field}
&{notin field} ,&{notin t.field} ,{not in field},|{not in t.field}
&{t.field in property}
&{t.field notin property},&{t.field not in property}
----(|符号类似)


样例： 
select * from table t where t.id = @{id} &{in t.name} 
select * from table t where t.id = @{id} &{t.name in myinputname} 

解析结果：
select * from table t where t.id = @p_id and t.name in (1,2,3)  --name:[1,2,3]
select * from table t where t.id = @p_id and t.name in ('1','2','3') --myinputname:["1","2","3"]
select * from table t where t.id = @p_id or t.name in ('1','2','3') --myinputname:["1","2","3"]
``` 

#### 6.4.7 运算符(>,>=,=,<>,<,<=), 支持符号&,|连接

```sql
&{> field} ,&{>= t.field}
&{t.field > property},&{t.field>=property}
----(|符号类似)

样例： 
select * from table t where t.id = @{id} &{> t.name} 
select * from table t where t.id = @{id} &{t.name = myinputname} 

解析结果：
select * from table t where t.id = @p_id and t.name > @p_name
select * from table t where t.id = @p_id and t.name = @p_myinputname
``` 

### 6.5 servers



## 7. 验收标准
### 7.1 功能验收
1. 所有接口正常响应，GET用于请求查询类接口，POST用于请求创建类/更新类接口,不再使用其他Method
2. 参数校验生效，错误提示清晰
3. 业务规则（如用户名唯一）正确执行
4. 异常场景（如查询不存在用户）返回正确状态码

### 7.2 框架规范验收
 
1. `golangci-lint run`无警告/错误
2. `go mod tidy`无未使用依赖
3. 无循环导入问题
4. 符合Golang代码风格（gofmt格式化）

## 8. 禁止项l
1. 禁止在services中编写数据访问逻辑中编写业务逻辑
2. 禁止使用`init()`函数初始化依赖
3. 禁止忽略错误返回
4. 禁止硬编码配置信息



