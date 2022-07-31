# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

This package compiles expressions and value-lists into SQL strings and thus offers an alternative to conventional query builders.

### Compile Values into SQL

In this example, a list of values is compiled into an SQL string.

https://github.com/wroge/superbasic/blob/4bc62cc1dd5773214783b8fa50647958038de2ab/example/main.go#L11-L26

### Compile Expressions into SQL

Similarly, expressions can be compiled in place of placeholders, which offers many new possibilities to create prepared statements. Some helper functions can be used, but my favorite is to write raw sql at any time.

https://github.com/wroge/superbasic/blob/4bc62cc1dd5773214783b8fa50647958038de2ab/example/main.go#L28-L38

### Queries

With this library it is particularly easy to create dynamic queries based on conditions. In this example, the WHERE-clause is only included if a corresponding expression exists.

https://github.com/wroge/superbasic/blob/4bc62cc1dd5773214783b8fa50647958038de2ab/example/main.go#L40-L58

Of course you can do the same with an ordinary Select Builder.

https://github.com/wroge/superbasic/blob/4bc62cc1dd5773214783b8fa50647958038de2ab/example/main.go#L60-L76

### Insert Builder

Additionally, there are other query builders (Insert, Update, Delete) that can be used to create prepared statements.


https://github.com/wroge/superbasic/blob/4bc62cc1dd5773214783b8fa50647958038de2ab/example/main.go#L78-L90
