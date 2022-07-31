# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

This package compiles expressions and value-lists into SQL strings and thus offers an alternative to conventional query builders.

### Compile Values into SQL

In this example, a list of values is compiled into an SQL string.

[embedmd]:# (example/main.go /[ \t]insert :=/ /Bush]/)
```go
	insert := superbasic.SQL("INSERT INTO presidents (nr, first, last) VALUES ? RETURNING id",
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
	// INSERT INTO presidents (nr, first, last) VALUES
	// 		($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12), ($13, $14, $15), ($16, $17, $18)
	//		RETURNING id
	// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]
```

### Compile Expressions into SQL

Similarly, expressions can be compiled in place of placeholders, which offers many new possibilities to create prepared statements. Some helper functions can be used, but my favorite is to write raw sql at any time.

[embedmd]:# (example/main.go /[ \t]update :=/ /45]/)
```go
	update := superbasic.SQL("UPDATE presidents SET ? WHERE ?",
		superbasic.Join(", ",
			superbasic.EqualsIdent("first", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		),
		superbasic.SQL("nr = ?", 45),
	)

	fmt.Println(update.ToSQL())
	// UPDATE presidents SET first = ?, last = ? WHERE nr = ?
	// [Donald Trump 45]
```

### Queries

With this library it is particularly easy to create dynamic queries based on conditions. In this example, the WHERE-clause is only included if a corresponding expression exists.

[embedmd]:# (example/main.go /[ \t]columns :=/ /44]/)
```go
	columns := []string{"nr", "first", "last"}

	where := superbasic.Or(
		superbasic.EqualsIdent("last", "Bush"),
		superbasic.SQL("nr >= ?", 45),
	)

	query := superbasic.Join(" ",
		superbasic.SQL("SELECT"),
		superbasic.IfElse(len(columns) > 0, superbasic.SQL(strings.Join(columns, ", ")), superbasic.SQL("*")),
		superbasic.SQL("FROM presidents"),
		superbasic.If(where != nil, superbasic.SQL("WHERE ?", where)),
		superbasic.SQL(fmt.Sprintf("ORDER BY %s", "nr")),
		superbasic.SQL(fmt.Sprintf("LIMIT %d", 3)),
	)

	fmt.Println(query.ToSQL())
	// SELECT nr, first, last FROM presidents WHERE (last = ? OR nr >= ?) ORDER BY nr LIMIT 3
	// [Bush 44]
```

Of course you can do the same with an ordinary Select Builder.

[embedmd]:# (example/main.go /[ \t]select_builder :=/ /44]/)
```go
	select_builder := &superbasic.SelectBuilder{}

	if len(columns) > 0 {
		select_builder.Select(strings.Join(columns, ", "))
	}

	select_builder.From("presidents")

	if where != nil {
		select_builder.WhereExpr(where)
	}

	select_builder.OrderBy("nr").Limit(3)

	fmt.Println(select_builder.ToSQL())
	// SELECT nr, first, last FROM presidents WHERE (last = ? OR nr >= ?) ORDER BY nr LIMIT 3
	// [Bush 44]
```

### Insert Builder

Additionally, there are other query builders (Insert, Update, Delete) that can be used to create prepared statements.

[embedmd]:# (example/main.go /[ \t]insert_builder :=/ /Bush]/)
```go
	insert_builder := superbasic.Insert("presidents").
		Columns("nr", "first", "last").
		Values(46, "Joe", "Biden").
		Values(45, "Donald", "trump").
		Values(44, "Barack", "Obama").
		Values(43, "George W.", "Bush").
		Values(42, "Bill", "Clinton").
		Values(41, "George H. W.", "Bush")

	fmt.Println(superbasic.Join(" ", insert_builder, superbasic.SQL("RETURNING id")).ToSQL())
	// INSERT INTO presidents (nr, first, last) VALUES
	// 		(?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?) RETURNING id
	// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]
```
