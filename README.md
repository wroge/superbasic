# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![tippin.me](https://badgen.net/badge/%E2%9A%A1%EF%B8%8Ftippin.me/@_wroge/yellow)](https://tippin.me/@_wroge)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.Compile``` compiles expressions into an SQL template and thus offers an alternative to conventional query builders.

- ```Compile``` replaces placeholders ```?``` with expressions.
- ```Join``` joins expressions by a separator.

```go
create := superbasic.Compile("CREATE TABLE presidents (\n\t?\n)",
	superbasic.Join(",\n\t",
		superbasic.SQL("nr SERIAL PRIMARY KEY"),
		superbasic.SQL("first TEXT NOT NULL"),
		superbasic.SQL("last TEXT NOT NULL"),
	),
)

fmt.Println(create.ToSQL())
// CREATE TABLE presidents (
//	nr SERIAL PRIMARY KEY,
//	first TEXT NOT NULL,
//	last TEXT NOT NULL
// )
```

- ```Map``` is a generic mapper function particularly helpful in the context of Join.
- ```Finalize``` replaces ```?``` placeholders with a string that can contain a positional part by ```%d```.

```go
presidents := []President{
	{"George", "Washington"},
	{"John", "Adams"},
}

insert := superbasic.Join(" ",
	superbasic.SQL("INSERT INTO presidents (first, last)"),
	superbasic.Compile("VALUES ?",
		superbasic.Join(", ",
			superbasic.Map(presidents,
				func(_ int, president President) superbasic.Expression {
					return superbasic.Values{president.First, president.Last}
				})...,
		),
	),
	superbasic.SQL("RETURNING nr"),
)

fmt.Println(superbasic.Finalize("$%d", insert))
// INSERT INTO presidents (first, last) VALUES ($1, $2), ($3, $4) RETURNING nr [George Washington John Adams]
```

- ```If``` condition is true, return expression, else skip.
- ```Switch``` returns an expression matching a value.

```go
dialect := "sqlite"
contains := "Joe"

query := superbasic.Join(" ", superbasic.SQL("SELECT * FROM presidents"),
	superbasic.If(contains != "", superbasic.Compile("WHERE ?",
		superbasic.Switch(dialect,
			superbasic.Case("postgres", superbasic.SQL("POSITION(? IN presidents.first) > 0", contains)),
			superbasic.Case("sqlite", superbasic.SQL("INSTR(presidents.first, ?) > 0", contains)),
		))))

fmt.Println(superbasic.Finalize("?", query))
// SELECT * FROM presidents WHERE INSTR(presidents.first, ?) > 0 [Joe] <nil>
```

To scan rows to types, i recommend [wroge/scan](https://github.com/wroge/scan).