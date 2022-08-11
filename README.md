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

You can compile any expression into an SQL template. Raw expressions can be created by ```superbasic.SQL```.

```go
expr := superbasic.Compile("INSERT INTO presidents (nr, first, last) VALUES ? RETURNING id",
	superbasic.Join(", ",
		superbasic.Values{46, "Joe", "Biden"},
		superbasic.Values{45, "Donald", "trump"},
		superbasic.Values{44, "Barack", "Obama"},
		superbasic.Values{43, "George W.", "Bush"},
		superbasic.Values{42, "Bill", "Clinton"},
		superbasic.Values{41, "George H. W.", "Bush"},
	),
)

fmt.Println(superbasic.Finalize("$", expr))
// INSERT INTO presidents (nr, first, last) VALUES
// 		($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18)
//		RETURNING id
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]

expr = superbasic.Compile("UPDATE presidents SET ? WHERE ?",
	superbasic.Join(", ",
		superbasic.SQL("first = ?", "Donald"),
		superbasic.SQL("last = ?", "Trump"),
	),
	superbasic.SQL("nr = ?", 45),
)

fmt.Println(expr.ToSQL())
// UPDATE presidents SET first = ?, last = ? WHERE nr = ?
// [Donald Trump 45]
```

## Query

Additionally, there are Query, Insert, Update and Delete helpers to create expressions.

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

insert := superbasic.Insert{
	Into:    "presidents",
	Columns: []string{"nr", "first", "last"},
	Data: []superbasic.Values{
		{46, "Joe", "Biden"},
		{45, "Donald", "trump"},
		{44, "Barack", "Obama"},
		{43, "George W.", "Bush"},
		{42, "Bill", "Clinton"},
		{41, "George H. W.", "Bush"},
	},
}
fmt.Println(superbasic.Finalize("$", superbasic.Compile("? RETURNING id", insert)))
// INSERT INTO presidents (nr, first, last) VALUES
// 		($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18)
//		RETURNING id
// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]
```
