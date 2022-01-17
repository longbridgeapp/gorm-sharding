# Gorm Sharding

Gorm 分表插件。

基于 SQL 解析器，执行前捕获原始 SQL，根据分表规则替换表名。

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

`Register` 函数接收一个 map，键是**原始表名**，值是一个 **resolver** ，由可配置的字段组成，见配置描述。

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

## 配置描述

- [EnableFullTable](#EnableFullTable)
- [ShardingColumn](#ShardingColumn)
- [ShardingAlgorithm](#ShardingAlgorithm)
- [ShardingAlgorithmByPrimaryKey](#ShardingAlgorithmByPrimaryKey)
- [PrimaryKeyGenerate](#PrimaryKeyGenerate)

#### EnableFullTable

Whether to enable a full table.
是否开启全表。

开启后，分表插件将双写数据到主表和分表。

#### ShardingColumn

用来分表的表字段。

如，对于一个订单表，你可能想用 `user_id` 拆分数据。

### ShardingAlgorithm

使用列值生成分表后缀的函数。

签名是 `func(columnValue interface{}) (suffix string, err error)`。

可以参考上面的用法示例。

#### ShardingAlgorithmByPrimaryKey

使用主键生成分表后缀的函数。

签名是 `func(id int64) (suffix string)`。

可以参考上面的用法示例。

注意，当记录包含 id 字段时，ShardingAlgorithmByPrimaryKey 优先于 ShardingAlgorithm。

#### PrimaryKeyGenerate

生成主键的函数。

只有当插入数据并且记录不包含 id 字段时使用。

签名是 `func(tableIdx int64) int64`。

可以参考上面的用法示例。

推荐使用 [keygen](https://github.com/longbridgeapp/gorm-sharding/tree/main/keygen) 组件，它是一个分布式的主键生成器。

当使用自增类的生成器时，tableIdx 参数可以忽略。

## 参考

https://gorm.io/docs/write_plugins.html#Plugin

## 许可证

本项目使用 MIT 许可证。
