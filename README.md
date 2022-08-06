# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.SQL``` compiles expressions into SQL strings and thus offers an alternative to conventional query builders.

https://github.com/wroge/superbasic/blob/828653a4af80c2af574af64bebe761ac650672fd/superbasic.go#L40-L42

You can compile a list of values into an SQL string...

https://github.com/wroge/superbasic/blob/828653a4af80c2af574af64bebe761ac650672fd/example/main.go#L11-L26

or any other expression. Lists of expressions are always joined by a comma.

https://github.com/wroge/superbasic/blob/828653a4af80c2af574af64bebe761ac650672fd/example/main.go#L28-L38

Additionally, there are Query, Insert, Update and Delete helpers that can be used to create prepared statements.

https://github.com/wroge/superbasic/blob/828653a4af80c2af574af64bebe761ac650672fd/example/main.go#L40-L51

The Query helper can be used as a reference to build your own expressions...

https://github.com/wroge/superbasic/blob/91bc5df478e8ba0de9910628cac8ae8d54241fe4/superbasic.go#L165-L195
