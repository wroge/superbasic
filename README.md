# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/wgs84)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/wgs84)](https://goreportcard.com/report/github.com/wroge/wgs84)

```go get github.com/wroge/superbasic``` and use a super-simple SQL builder consisting of the 4 basic functions ```SQL(sql, expr...)```, ```Append(expr...)```, ```Join(sep, expr...)``` and ```If(condition, then, else)```.

```go
insert := superbasic.SQL(
    "INSERT INTO presidents (first, last) VALUES ?",
    superbasic.Join(", ",
        superbasic.SQL("(?)", superbasic.Join(", ", "Joe", "Biden")),
        superbasic.SQL("(?)", superbasic.Join(", ", "Donald", "Trump")),
        superbasic.SQL("(?)", superbasic.Join(", ", "Barack", "Obama")),
        superbasic.SQL("(?)", superbasic.Join(", ", "George W.", "Bush")),
        superbasic.SQL("(?)", superbasic.Join(", ", "Bill", "Clinton")),
        superbasic.SQL("(?)", superbasic.Join(", ", "George H. W.", "Bush")),
    ),
)

fmt.Println(insert.ToSQL())
// INSERT INTO presidents (first, last) VALUES (?, ?), (?, ?), (?, ?), (?, ?), (?, ?), (?, ?) 
// [Joe Biden Donald Trump Barack Obama George W. Bush Bill Clinton George H. W. Bush]
```

## Dynamic Queries

This package is written to be able to build dynamic queries in the simplest way possible.
Each ```Expression``` is automatically compiled to the correct place.
There is a ```ToPostgres``` function to translate expressions for Postgres databases.

```go
type Select struct {
	Columns []superbasic.Sqlizer
	From    superbasic.Sqlizer
	Where   []superbasic.Sqlizer
}

func (s Select) Expression() superbasic.Expression {
	return superbasic.Append(
		superbasic.SQL("SELECT "),
		superbasic.If(len(s.Columns) > 0, s.Columns, superbasic.SQL("*")),
		superbasic.If(s.From != nil, superbasic.SQL(" FROM ?", s.From), superbasic.Skip()),
		superbasic.If(len(s.Where) > 0,
			superbasic.SQL(" WHERE ?",
				superbasic.If(len(s.Where) > 1,
					superbasic.Join(" AND ", s.Where),
					s.Where,
				),
			),
			superbasic.Skip(),
		),
	)
}

func main() {
	query := Select{
		Columns: []superbasic.Sqlizer{
			superbasic.SQL("id"),
			superbasic.SQL("first"),
			superbasic.SQL("last"),
		},
		From: superbasic.SQL("presidents"),
		Where: []superbasic.Sqlizer{
			superbasic.SQL("last = ?", "Bush"),
		},
	}

	fmt.Println(superbasic.ToPostgres(query))
	// SELECT id, first, last FROM presidents WHERE last = $1 [Bush]

	query = Select{
		Columns: []superbasic.Sqlizer{},
		From:    superbasic.SQL("presidents"),
		Where: []superbasic.Sqlizer{
			superbasic.SQL("last = ?", "Bush"),
			superbasic.SQL("first = ?", "Joe"),
		},
	}

	fmt.Println(superbasic.ToPostgres(query))
	// SELECT * FROM presidents WHERE last = $1 AND first = $2 [Bush Joe]

	query = Select{
		From: superbasic.SQL("presidents"),
	}

	fmt.Println(superbasic.ToPostgres(query))
	// SELECT * FROM presidents []
}
```

## Types and Interfaces

```superbasic.Expression```, ```superbasic.Sqlizer```, ```superbasic.Expr```, 
```[]superbasic.Expression```, ```[]superbasic.Sqlizer``` and ```[]superbasic.Expr```
can be passed to any function and are handled accordingly. 
The slice expressions are converted to ```Join(sep, expr...)``` within ```Join``` expressions and to ```Join(", ", expr...)``` in all other cases.

```go 
type Expr interface {
	Expression() Expression
}

type Sqlizer interface {
	ToSQL() (string, []any, error)
}
```
