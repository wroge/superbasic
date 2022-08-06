# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.SQL``` compiles expressions into SQL strings and thus offers an alternative to conventional query builders.

### Compile Expressions into SQL

You can compile a list of values into an SQL string...

https://github.com/wroge/superbasic/blob/5daab8f309f59becbe89872659894345e5976221/example/main.go#L11-L26

or any other expression. Lists of expressions are always joined by a comma.

https://github.com/wroge/superbasic/blob/5daab8f309f59becbe89872659894345e5976221/example/main.go#L28-L38

### Queries

With this library it is particularly easy to create dynamic queries based on conditions. In this example, the WHERE-clause is only included if a corresponding expression exists.

https://github.com/wroge/superbasic/blob/5daab8f309f59becbe89872659894345e5976221/example/main.go#L40-L52

Of course you can do the same with this Query.

https://github.com/wroge/superbasic/blob/5daab8f309f59becbe89872659894345e5976221/example/main.go#L54-L62

### Insert, Update, Delete

Additionally, there are Insert, Update and Delete helpers that can be used to create prepared statements.

https://github.com/wroge/superbasic/blob/5daab8f309f59becbe89872659894345e5976221/example/main.go#L64-L80
