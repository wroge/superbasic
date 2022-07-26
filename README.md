# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

This package provides utilities to build complex and dynamic but secure SQL queries.

```go
go get github.com/wroge/superbasic
```

Simply put, superbasic helps you write prepared statements, and is also a kind of template engine that compiles expressions.

```go
insert := superbasic.SQL("INSERT INTO presidents (?) VALUES ? RETURNING id",
	superbasic.Columns{"first", "last"},
	superbasic.Values{
		{"Joe", "Biden"},
		{"Donald", "trump"},
		{"Barack", "Obama"},
		{"George W.", "Bush"},
		{"Bill", "Clinton"},
		{"George H. W.", "Bush"},
	},
)

fmt.Println(superbasic.ToPostgres(insert))
// INSERT INTO presidents (first, last) VALUES ($1, $2), ($3, $4), ($5, $6), ($7, $8), ($9, $10), ($11, $12) RETURNING id 
// [Joe Biden Donald trump Barack Obama George W. Bush Bill Clinton George H. W. Bush]

update := superbasic.SQL("UPDATE presidents SET last = ? WHERE ?",
	"Trump",
	superbasic.SQL("first = ?", "Donald"),
)

fmt.Println(update.ToSQL())
// UPDATE presidents SET last = ? WHERE first = ? [Trump Donald]

columns := []string{"id", "first", "last"}
lastname := "Bush"
sort := "first"

query := superbasic.Append(
	superbasic.SQL("SELECT ? FROM presidents", superbasic.Columns(columns)),
	superbasic.If(lastname != "", superbasic.SQL(" WHERE last = ?", lastname)),
	superbasic.If(sort != "", superbasic.SQL(" ORDER BY ?", superbasic.SQL(sort))),
)

fmt.Println(query.ToSQL())
// SELECT id, first, last FROM presidents WHERE last = ? ORDER BY first [Bush]

delete := superbasic.SQL("DELETE FROM presidents WHERE ?",
	superbasic.Join(" AND ",
		superbasic.SQL("last = ?", "Bush"),
		superbasic.SQL("first = ?", "Joe"),
	),
)

fmt.Println(superbasic.ToPostgres(delete))
// DELETE FROM presidents WHERE last = $1 AND first = $2 [Bush Joe]

builder := superbasic.NewBuilder().WriteSQL("SELECT ")

builder.Write(superbasic.Columns(columns)).WriteSQL(" FROM presidents")

builder.Write(superbasic.SQL(" WHERE first IN (?, ?)", "Barack", "Donald"))

fmt.Println(builder.ToSQL())
// SELECT id, first, last FROM presidents WHERE first IN (?, ?) [Barack Donald]
```