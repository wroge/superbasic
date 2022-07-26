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

## Base Building-Blocks

The base building-blocks are
```superbasic.SQL(sql, expr...)```,
```superbasic.Append(expr...)```,
```superbasic.Join(sep, expr...)```,
```superbasic.If(condition, then)``` and
```superbasic.IfElse(condition, then, else)```
All further expressions are based on these functions.

```go
search := "biden"

sql, args, err := superbasic.Append(
	superbasic.SQL("SELECT id, first, last FROM presidents"),
	superbasic.If(search != "", superbasic.SQL(" WHERE last = ?", search)),
).ToSQL()
// SELECT id, first, last FROM presidents WHERE last = ? [Biden] <nil>
```

## Types and Interfaces

Expressions within the arguments, like ```superbasic.Expression``` or ```superbasic.Sqlizer```, are recognized and handled accordingly. Unknown arguments are built into the expression using placeholders. Slices of expressions, like ```[]superbasic.Expression```, are joined by the separator in the ```superbasic.Join``` expression or by ```", "``` in all other cases.

```go 
sql, args, err = superbasic.SQL("SELECT ? FROM presidents", 
	[]superbasic.Expression{superbasic.SQL("first"), superbasic.SQL("last")}).ToSQL()
// SELECT first, last FROM presidents
```

## Build-in Queries

In addition, there are expressions, such as ```superbasic.Query``` or ```superbasic.Insert```, that allow you to build the queries even more easily. Sometimes it is more readable to use the helper functions like ```superbasic.And```and
```superbasic.Equals```, therefore this library also offers an increasing amount of utility-functions.

This example shows a query where the resulting sql expression is translated into Postgres format.

```go
query := superbasic.Query{
	Select: superbasic.SQL("p.id, p.first, p.last"),
	From:   superbasic.SQL("presidents AS p"),
	Where: superbasic.And(
		superbasic.Equals("p.last", "Bush"),
		superbasic.NotEquals("p.first", "George W."),
	),
	// Could also be written as:
	// Where: superbasic.Join(" AND ",
	// 		superbasic.SQL("p.last = ?", "Bush"),
	// 		superbasic.SQL("p.first <> ?", "George W."),
	// ),
	OrderBy: superbasic.SQL("p.last"),
	Limit:   3,
}

sql, args, err = superbasic.ToPostgres(query)
// SELECT p.id, p.first, p.last FROM presidents AS p WHERE p.last = $1 AND p.first <> $2 ORDER BY p.last LIMIT 3
// [Bush George W.]
```

All expressions can, of course, be nested within each other to produce new expressions. 
The ```superbasic.Insert``` example shows how to append the Postgres-specific expression ```RETURNING id```.

```go
insert := superbasic.Append(
	superbasic.Insert{
		Into:    "presidents",
		Columns: []string{"first", "last"},
		Values: [][]any{
			{"Joe", "Bden"},
			{"Donald", "Trump"},
			{"Barack", "Obama"},
			{"George W.", "Bush"},
			{"Bill", "Clinton"},
			{"George H. W.", "Bush"},
		},
	},
	superbasic.SQL(" RETURNING id"),
)

sql, args, err := superbasic.ToPostgres(insert)
// INSERT INTO presidents (first, last) VALUES ($1, $2), ($3, $4), ($5, $6), ($7, $8), ($9, $10), ($11, $12) RETURNING id 
// [Joe Bden Donald Trump Barack Obama George W. Bush Bill Clinton George H. W. Bush]
```

If you prefer to write the SQL queries yourself, you are of course welcome to do so.

```go
sql, args, err := superbasic.SQL("INSERT INTO presidents (first, last) VALUES ? RETURNING id", superbasic.Values(
	[]any{"Joe", "Bden"},
	[]any{"Donald", "Trump"},
	[]any{"Barack", "Obama"},
	[]any{"George W.", "Bush"},
	[]any{"Bill", "Clinton"},
	[]any{"George H. W.", "Bush"},
)).ToSQL()
// INSERT INTO presidents (first, last) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?), (?, ?) RETURNING id 
// [Joe Bden Donald Trump Barack Obama George W. Bush Bill Clinton George H. W. Bush]
```

## Customized Expressions 

The next section shows the ```superbasic.Query``` expression. Here you can see how easy it is to create your own expressions.

```go
type Query struct {
	Select  any
	From    any
	Where   any
	GroupBy any
	Having  any
	OrderBy any
	Limit   uint64
	Offset  uint64
}

func (q Query) ToSQL() (string, []any, error) {
	return Append(
		SQL("SELECT ?", IfElse(q.Select != nil, q.Select, SQL("*"))),
		If(q.From != nil, SQL(" FROM ?", q.From)),
		If(q.Where != nil, SQL(" WHERE ?", q.Where)),
		If(q.GroupBy != nil, SQL(" GROUP BY ?", q.GroupBy)),
		If(q.Having != nil, SQL(" HAVING ?", q.Having)),
		If(q.OrderBy != nil, SQL(" ORDER BY ?", q.OrderBy)),
		If(q.Limit > 0, SQL(fmt.Sprintf(" LIMIT %d", q.Limit))),
		If(q.Offset > 0, SQL(fmt.Sprintf(" OFFSET %d", q.Offset))),
	).ToSQL()
}
```

## DDL expressions

It is also possible to create DDL queries, for example using the predefined expression ```superbasic.Table```.

```go
table := superbasic.Table{
	IfNotExists: true,
	Name:        "presidents",
	Columns: []superbasic.Sqlizer{
		superbasic.Column{
			Name: "id",
			Type: "SERIAL",
			Constraints: []superbasic.Sqlizer{
				superbasic.SQL("PRIMARY KEY"),
			},
		},
		superbasic.SQL("first TEXT NOT NULL"),
		superbasic.SQL("last TEXT NOT NULL"),
	},
	Constraints: superbasic.SQL("UNIQUE (first, last)"),
}

sql, err := superbasic.ToDDL(table)
// CREATE TABLE IF NOT EXISTS presidents (id SERIAL PRIMARY KEY, first TEXT NOT NULL, last TEXT NOT NULL, UNIQUE (first, last))
```

## SQL Builder

Of course, there is also a builder that can be used to create the SQL query. Here is an example:

```go
b := superbasic.NewBuilder()

b.WriteSQL("SELECT ").WriteSQL("first, last")
b.WriteSQL(" FROM presidents")
b.WriteSQL(" WHERE ")
b.Write(superbasic.Join(" OR ", superbasic.SQL("last = ?", "Bush"), superbasic.SQL("first = ?", "Joe")))

sql, args, err := b.ToSQL()
// SELECT first, last FROM presidents WHERE last = ? OR first = ? 
// [Bush Joe]
```