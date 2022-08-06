# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

This package compiles expressions into SQL strings and thus offers an alternative to conventional query builders.

### Compile Values into SQL

In this example, a list of values is compiled into an SQL string.

https://github.com/wroge/superbasic/blob/94b8f01875c84aadd95ab1e9c55f13106779cfac/example/main.go#L12-27

### Compile Expressions into SQL

Similarly, expressions can be compiled in place of placeholders, which offers many new possibilities to create prepared statements.

https://github.com/wroge/superbasic/blob/94b8f01875c84aadd95ab1e9c55f13106779cfac/example/main.go#L29-L39

### Queries

With this library it is particularly easy to create dynamic queries based on conditions. In this example, the WHERE-clause is only included if a corresponding expression exists.

https://github.com/wroge/superbasic/blob/94b8f01875c84aadd95ab1e9c55f13106779cfac/example/main.go#L41-L59

Of course you can do the same with this Query.

https://github.com/wroge/superbasic/blob/94b8f01875c84aadd95ab1e9c55f13106779cfac/example/main.go#L61-L70

### Insert, Update, Delete

Additionally, there are Insert, Update and Delete helpers that can be used to create prepared statements.


https://github.com/wroge/superbasic/blob/94b8f01875c84aadd95ab1e9c55f13106779cfac/example/main.go#L72-L88
