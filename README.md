# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.Compile``` compiles expressions into an SQL template and thus offers an alternative to conventional query builders.

```go
type Expression interface { 
    ToSQL() (string, []any, error) 
}
```

## Compile Expressions into SQL

You can compile a list of values into an SQL template...

```go
expr := superbasic.Compile("INSERT INTO presidents (nr, first, last) VALUES ? RETURNING id", 
    []superbasic.Values{ 
        {46, "Joe", "Biden"}, 
        {45, "Donald", "trump"}, 
        {44, "Barack", "Obama"}, 
        {43, "George W.", "Bush"}, 
        {42, "Bill", "Clinton"}, 
        {41, "George H. W.", "Bush"}, 
    }, 
) 

fmt.Println(superbasic.ToPositional("$", expr)) 
// INSERT INTO presidents (nr, first, last) VALUES 
// 		($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18) 
//		RETURNING id 
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush] 
```

or any other expression. Lists of expressions are always joined by ", ".

```go
expr := superbasic.Compile("UPDATE presidents SET ? WHERE ?",
    []superbasic.Expression{
        superbasic.SQL("first = ?", "Donald"),
        superbasic.SQL("last = ?", "Trump"),
    },
    superbasic.SQL("nr = ?", 45),
)

fmt.Println(expr.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE nr = ?
// [Donald Trump 45]
```

## Query

Additionally, there are Query, Insert, Update and Delete helpers that can be used to create expressions.

```go
query := superbasic.Query{
    From: superbasic.SQL("presidents"),
    Where: superbasic.Compile("(? OR ?)",
        superbasic.SQL("last = ?", "Bush"),
        superbasic.SQL("nr >= ?", 45),
    ),
    OrderBy: superbasic.SQL("nr"),
    Limit:   3,
}

fmt.Println(query.ToSQL())
// SELECT * FROM presidents WHERE (last = ? OR nr >= ?) ORDER BY nr LIMIT 3
// [Bush 44]
```

The Query helper can be used as a reference to build your own expressions...

```go
type Query struct {
	With    Expression
	Select  Expression
	From    Expression
	Where   Expression
	GroupBy Expression
	Having  Expression
	Window  Expression
	OrderBy Expression
	Limit   uint64
	Offset  uint64
}

func (q Query) ToSQL() (string, []any, error) {
	return Join(" ",
		If(q.With != nil, Compile("WITH ?", q.With)),
		SQL("SELECT"),
		IfElse(q.Select != nil, q.Select, SQL("*")),
		If(q.From != nil, Compile("FROM ?", q.From)),
		If(q.Where != nil, Compile("WHERE ?", q.Where)),
		If(q.GroupBy != nil, Compile("GROUP BY ?", q.GroupBy)),
		If(q.Having != nil, Compile("HAVING ?", q.Having)),
		If(q.Having != nil, Compile("WINDOW ?", q.Window)),
		If(q.OrderBy != nil, Compile("ORDER BY ?", q.OrderBy)),
		If(q.Limit > 0, SQL(fmt.Sprintf("LIMIT %d", q.Limit))),
		If(q.Offset > 0, SQL(fmt.Sprintf("OFFSET %d", q.Offset))),
	).ToSQL()
}
```
