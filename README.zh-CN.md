# Gorm Sharding

[English](./README.md) | 简体中文

Gorm Sharding 将大表拆分成多个小表来加速访问。

假设有一个一亿行的订单表，我们可以将它拆分成一千个小表，这样读写操作就会比操作原始表快很多。

![场景示例](./images/example-scenario.svg)

## 特色

- 非侵入式设计。加载插件，指定配置，完成。
- 快。没有网络中间件，跟 Go 一样快。
- 多数据库支持。PostgreSQL 已测试，MySQL 和 SQLite 进行中。

## 安装

```bash
go get -u github.com/longbridgeapp/gorm-sharding
```

## 用法

数据库连接打开后，使用分表插件注册需要拆分的表。

`Register` 函数接收一个 map，键是**原始表名**，值是一个 **resolver** ，由可配置的字段组成。

具体配置信息见 [Godoc](https://pkg.go.dev/github.com/longbridge/gorm-sharding)。

## 用法示例

```go
middleware := sharding.Register(map[string]sharding.Resolver{
	"orders": {
		EnableFullTable: true,
		ShardingColumn: "user_id",
		ShardingAlgorithm: func(value interface{}) (suffix string, err error) {
			switch user_id := value.(type) {
			case int64:
				return fmt.Sprintf("_%02d", user_id % 64), nil
			default:
				return "", errors.New("invalid user_id")
			}
		},
		ShardingAlgorithmByPrimaryKey: func(id int64) (suffix string) {
			return fmt.Sprintf("_%02d", keygen.TableIdx(id))
		},
		PrimaryKeyGenerate: func(tableIdx int64) int64 {
			return keygen.Next(tableIdx)
		}
	},
})
db.Use(middleware)
```

## 许可证

本项目使用 MIT 许可证。
