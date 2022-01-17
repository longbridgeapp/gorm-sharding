# Keygen

English | [简体中文](./README.zh-CN.md)

A Distributed Primary Key generator.

## Differences with the snowflake algorithm

Keygen id contains a table index in it's bit sequence, so we can locate the sharding table whitout the sharding key.
For example, use Keygen id, we could use:

```
select * from orders where id = 76362673717182593
```

Use snowflake, you should use:

```
select * from orders where id = 76362673717182593 and user_id = 100
```

The former is commonly used in CRUD codes so we made this improvement.

## License

This project under MIT license.
