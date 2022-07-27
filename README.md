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
	superbasic.Columns{"id", "first", "last"},
	[][]any{
		{46, "Joe", "Biden"},
		{45, "Donald", "trump"},
		{44, "Barack", "Obama"},
		{43, "George W.", "Bush"},
		{42, "Bill", "Clinton"},
		{41, "George H. W.", "Bush"},
	},
)

fmt.Println(superbasic.ToPostgres(insert))
// INSERT INTO presidents (id, first, last) VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18) RETURNING id 
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]


update := superbasic.SQL("UPDATE presidents SET ? WHERE ?",
		[]superbasic.Sqlizer{
			superbasic.SQL("first = ?", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		},
		superbasic.SQL("id = ?", 45),
	)

fmt.Println(update.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE id = ? 
// [Trump Donald 45]


columns := []string{"id", "first", "last"}
sort := "first"

query := superbasic.Append(
	superbasic.SQL("SELECT ? FROM presidents", superbasic.Columns(columns)),
	superbasic.SQL(" WHERE last IN ?", []any{"Bush", "Clinton"}),
	superbasic.If(sort != "", superbasic.SQL(" ORDER BY ?", superbasic.SQL(sort))),
)

fmt.Println(query.ToSQL())
// SELECT id, first, last FROM presidents WHERE last IN (?, ?) ORDER BY first 
// [Bush Clinton]


delete := superbasic.SQL("DELETE FROM presidents WHERE ?",
	superbasic.Join(" AND ",
		superbasic.SQL("last = ?", "Bush"),
		superbasic.SQL("first = ?", "Joe"),
	),
)

fmt.Println(superbasic.ToPostgres(delete))
// DELETE FROM presidents WHERE last = $1 AND first = $2 [Bush Joe]
```