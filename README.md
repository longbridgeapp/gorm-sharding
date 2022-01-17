# Gorm Sharding

English | [简体中文](./README.zh-CN.md)

Gorm Sharding splits large tables into smaller ones to speed up access.

Assume we have an order table with 100 million rows, we can divide it into 1000 smaller ones, so the read and write operation will faster than operate the original table.

![Example Scenario](./images/example-scenario.svg)

## Features

- Non-intrusive design. Load the plugin, specify the config, and all done.
- Lighting-fast. No network based middlewares, as fast as Go.
- Multiple database support. PostgreSQL tested, MySQL and SQLite is coming.

## Install

```bash
go get -u github.com/longbridgeapp/gorm-sharding
```

## Usage

After the database connection opened, use the sharding plugin that registered the tables which you want to shard.

The `Register` function takes a map, the key is the **original table name** and the value is a **resolver** which is composed by five configurable fields.

For config detail info, see [Godoc](https://pkg.go.dev/github.com/longbridge/gorm-sharding).

## Usage Example

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

## License

This project under MIT license.
