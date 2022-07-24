# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

This package provides utilities to build secure SQL queries. 

```go
go get github.com/wroge/superbasic
```

## Base Building-Blocks

The base building-blocks are
```superbasic.SQL(sql, expr...)```,
```superbasic.Append(expr...)```,
```superbasic.Join(sep, expr...)``` and
```superbasic.If(condition, then, else)```.
All further expressions are based on these four functions.
As the name suggests ```superbasic.Skip``` can be used to skip expressions, 
which is useful in ```superbasic.If``` conditions.

```go
search := "biden"

sql, args, err := superbasic.Append(
	superbasic.SQL("SELECT id, first, last FROM presidents"),
	superbasic.If(search != "", superbasic.SQL(" WHERE last = ?", search), superbasic.Skip()),
).ToSQL()
// SELECT id, first, last FROM presidents WHERE last = ? [Biden] <nil>
```

## Types and Interfaces

Expressions within the arguments, like ```superbasic.Expression``` or ```superbasic.Expr```, are recognized and handled accordingly. Unknown arguments are built into the expression using placeholders. Slices of expressions, like ```[]superbasic.Expression```, are joined by the separator in the ```superbasic.Join``` expression or by ```", "``` in all other cases.

```go 
sql, args, err = superbasic.SQL("SELECT ? FROM presidents", []superbasic.Expression{superbasic.SQL("first"), superbasic.SQL("last")}).ToSQL()
// SELECT first, last FROM presidents
```

## Build-in Queries

In addition, there are expressions, such as ```superbasic.Select``` or ```superbasic.Insert```, that allow you to build the queries even more easily. This example shows a query where the resulting sql expression is translated into Postgres format.

```go
query := superbasic.Select{
	Columns: []superbasic.Expr{
		superbasic.SQL("id"),
		superbasic.SQL("first"),
		superbasic.SQL("last"),
	},
	From: []superbasic.Expr{
		superbasic.SQL("presidents"),
	},
	Where: superbasic.SQL("? OR ?",
		superbasic.SQL("last = ?", "Bush"), superbasic.SQL("first = ?", "Joe")),
	OrderBy: []superbasic.Expr{
		superbasic.SQL("last"),
	},
	Limit: 3,
}

sql, args, err = superbasic.ToPostgres(query)
// SELECT id, first, last FROM presidents WHERE last = $1 OR first = $2 ORDER BY last LIMIT 3
// [Bush Joe]
```

## Customized Expressions 

The next section shows the ```superbasic.Select``` expression. Here you can see how easy it is to create your own expressions.

```go
type Select struct {
	Distinct bool
	Columns  []Expr
	From     []Expr
	Joins    []Expr
	Where    Expr
	GroupBy  []Expr
	Having   Expr
	OrderBy  []Expr
	Limit    uint64
	Offset   uint64
}

func (s Select) ToSQL() (string, []any, error) {
	return Append(
		SQL("SELECT "),
		If(s.Distinct, SQL("DISTINCT "), Skip()),
		If(len(s.Columns) > 0, s.Columns, SQL("*")),
		If(len(s.From) > 0, SQL(" FROM ?", s.From), Skip()),
		If(len(s.Joins) > 0, SQL(" ?", s.Joins), Skip()),
		If(s.Where != nil, SQL(" WHERE ?", s.Where), Skip()),
		If(len(s.GroupBy) > 0, SQL(" GROUP BY ?", s.GroupBy), Skip()),
		If(s.Having != nil, SQL(" HAVING ?", s.Having), Skip()),
		If(len(s.OrderBy) > 0, SQL(" ORDER BY ?", s.OrderBy), Skip()),
		If(s.Limit > 0, SQL(fmt.Sprintf(" LIMIT %d", s.Limit)), Skip()),
		If(s.Offset > 0, SQL(fmt.Sprintf(" OFFSET %d", s.Offset)), Skip()),
	).ToSQL()
}
```

## DDL expressions

It is also possible to create DDL queries, for example using the predefined expression ```superbasic.Table```.
```superbasic.ExprDDL``` is supported by ```superbasic.Expression``` and it makes sure that the expression doesn't contain arguments.

```go
table := superbasic.Table{
	IfNotExists: true,
	Name:        "presidents",
	Columns: []superbasic.ExprDDL{
		superbasic.Column{
			Name: "id",
			Type: "SERIAL",
			Constraints: []superbasic.ExprDDL{
				superbasic.SQL("PRIMARY KEY"),
			},
		},
		superbasic.Column{
			Name: "first",
			Type: "TEXT",
			Constraints: []superbasic.ExprDDL{
				superbasic.SQL("NOT NULL"),
			},
		},
		superbasic.Column{
			Name: "last",
			Type: "TEXT",
			Constraints: []superbasic.ExprDDL{
				superbasic.SQL("NOT NULL"),
			},
		},
	},
	Constraints: []superbasic.ExprDDL{
		superbasic.SQL("UNIQUE (first, last)"),
	},
}

sql, err = table.ToDDL())
// CREATE TABLE presidents (id SERIAL PRIMARY KEY, first TEXT NOT NULL, last TEXT NOT NULL, UNIQUE (first, last))
```