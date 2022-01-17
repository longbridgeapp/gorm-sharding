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

The `Register` function takes a map, the key is the **original table name** and the value is a **resolver** which is composed by five configurable fields that described in [Config description](#config-description).

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

## Config description

- [EnableFullTable](#EnableFullTable)
- [ShardingColumn](#ShardingColumn)
- [ShardingAlgorithm](#ShardingAlgorithm)
- [ShardingAlgorithmByPrimaryKey](#ShardingAlgorithmByPrimaryKey)
- [PrimaryKeyGenerate](#PrimaryKeyGenerate)

#### EnableFullTable

Whether to enable a full table.

When full table enabled, sharding plugin will double write data to both main table and sharding table.

#### ShardingColumn

Which table column you want to used for sharding the table rows.

For example, for a product order table, you may want to split the rows by `user_id`.

### ShardingAlgorithm

A function to generate the sharding table's suffix by the column value.

It's signature is `func(columnValue interface{}) (suffix string, err error)`.

For an example, see the [usage example](#usage-example) above.

#### ShardingAlgorithmByPrimaryKey

A function to generate the sharding table's suffix by the primary key.

It's signature is `func(id int64) (suffix string)`.

For an example, see the [usage example](#usage-example) above.

Note, when the record contains an id field, ShardingAlgorithmByPrimaryKey will preferred than ShardingAlgorithm.

#### PrimaryKeyGenerate

A function to generate the primary key.

Used only when insert and the record does not contains an id field.

It's signature is `func(tableIdx int64) int64`.

For an example, see the [usage example](#usage-example) above.

We recommend you use the [keygen](https://github.com/longbridgeapp/gorm-sharding/tree/main/keygen) component, it is a distributed primary key generator.

When use auto-increment like generator, the tableIdx argument could ignored.

## References

https://gorm.io/docs/write_plugins.html#Plugin

## License

This project under MIT license.
