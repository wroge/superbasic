# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.Compile``` compiles expressions into a template.

In addition, this package provides a set of functions that can be used to create expressions.

```go
go get github.com/wroge/superbasic

// Sqlizer represents a prepared statement.
type Sqlizer interface {
	ToSQL() (string, []any, error)
}


insert := superbasic.Compile("INSERT INTO presidents (?) VALUES ? RETURNING id",
	superbasic.Idents("id", "first", "last"),
	superbasic.Join(", ",
		superbasic.Values(46, "Joe", "Biden"),
		superbasic.Values(45, "Donald", "trump"),
		superbasic.Values(44, "Barack", "Obama"),
		superbasic.Values(43, "George W.", "Bush"),
		superbasic.Values(42, "Bill", "Clinton"),
		superbasic.Values(41, "George H. W.", "Bush"),
	),
)

fmt.Println(insert.ToPositional("$"))
// INSERT INTO presidents (id, first, last) VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18) RETURNING id
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]


update := superbasic.Compile("UPDATE presidents SET ? WHERE ?",
	superbasic.Join(", ",
		superbasic.Equals("first", "Donald"),
		superbasic.Equals("last", "Trump"),
	),
	superbasic.Equals("id", 45),
)

fmt.Println(update.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE id = ?
// [Donald Trump 45]


search := superbasic.In("last", "Bush", "Clinton")
sort := "first"

query := superbasic.Append(
	superbasic.SQL("SELECT id, first, last FROM presidents"),
	superbasic.If(search.SQL != "", superbasic.Compile(" WHERE ?", search)),
	superbasic.If(sort != "", superbasic.SQL(fmt.Sprintf(" ORDER BY %s", sort))),
)

fmt.Println(query.ToSQL())
// SELECT id, first, last FROM presidents WHERE last IN (?, ?) ORDER BY first
// [Bush Clinton]
```