# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.SQL``` compiles expressions into a template.  
In addition, this package provides a set of functions that can be used to create expressions.  

```go
go get github.com/wroge/superbasic

// []Expression is compiled to Join(sep, expr...). (default sep is ", ")
// []any is compiled to (?, ?).  
// [][]any is compiled to (?, ?), (?, ?).  
// Escape '?' by using '??'.

insert := superbasic.SQL("INSERT INTO presidents (?) VALUES ? RETURNING id",
	superbasic.SQL("nr, first, last"),
	[][]any{
		{46, "Joe", "Biden"},
		{45, "Donald", "trump"},
		{44, "Barack", "Obama"},
		{43, "George W.", "Bush"},
		{42, "Bill", "Clinton"},
		{41, "George H. W.", "Bush"},
	},
)

fmt.Println(superbasic.ToPositional("$", insert))
// INSERT INTO presidents (nr, first, last) VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18) RETURNING id
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]


update := superbasic.SQL("UPDATE presidents SET ? WHERE ?",
	superbasic.Join(", ",
		superbasic.EqualsIdent("first", "Donald"),
		superbasic.EqualsIdent("last", "Trump"),
	),
	superbasic.EqualsIdent("nr", 45),
)

fmt.Println(update.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE nr = ?
// [Donald Trump 45]


search := superbasic.And(
	superbasic.InIdent("last", []any{"Bush", "Clinton"}),
	superbasic.Not(superbasic.GreaterIdent("nr", 42)),
)
sort := "first"

query := superbasic.Append(
	superbasic.SQL("SELECT nr, first, last FROM presidents"),
	superbasic.If(search != nil, superbasic.SQL(" WHERE ?", search)),
	superbasic.If(sort != "", superbasic.SQL(fmt.Sprintf(" ORDER BY %s", sort))),
)

fmt.Println(query.ToSQL())
// SELECT nr, first, last FROM presidents WHERE last IN (?, ?) AND NOT (nr > ?) ORDER BY first
// [Bush Clinton 42]

```

## Query Builders

These query builders are a few examples of how you can create your own expressions. See the documentation for that.

```go

query := superbasic.SelectSQL("nr, first, last").
	FromSQL("presidents").
	Where(superbasic.Or(
		superbasic.EqualsIdent("last", "Bush"),
		superbasic.GreaterIdent("nr", 44),
	)).
	Limit(3)

fmt.Println(query.ToSQL())
// SELECT nr, first, last FROM presidents WHERE (last = ? OR nr > ?)
// [Bush 44]


insert := superbasic.Insert("presidents").
	Columns("nr", "first", "last").
	AddRow(46, "Joe", "Biden").
	AddRow(45, "Donald", "trump").
	AddRow(44, "Barack", "Obama").
	AddRow(43, "George W.", "Bush").
	AddRow(42, "Bill", "Clinton").
	AddRow(41, "George H. W.", "Bush")

fmt.Println(insert.ToSQL())
// INSERT INTO presidents (nr, first, last) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?) 
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]


update := superbasic.Update("presidents").
	AddSet(superbasic.EqualsIdent("first", "Donald")).
	AddSetSQL("last = ?", "Trump").
	Where(superbasic.EqualsIdent("nr", 45))

fmt.Println(update.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE nr = ? 
// [Donald Trump 45]


expr := superbasic.Delete("presidents").
	Where(superbasic.EqualsIdent("last", "Bush"))

fmt.Println(expr.ToSQL())
// DELETE FROM ? WHERE last = ? 
// [presidents Bush]
```