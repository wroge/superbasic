# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/wgs84)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/wgs84)](https://goreportcard.com/report/github.com/wroge/wgs84)

```go get github.com/wroge/superbasic``` and use a super-simple SQL builder consisting of the three basic functions ```Append(expr...)```, ```Join(sep, expr...)``` and ```SQL(sql, expr...)```.

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

This library is written to be able to build dynamic queries in the simplest way possible.
Each ```Expression``` is automatically compiled to the correct place. ```[]Expression``` is converted to ```Join(sep, expr...)``` within ```Join``` expressions and to ```Join(", ", expr...)``` in all other cases.
There is a ```ToPostgres``` function to translate expressions for Postgres databases.

```go
columns := []superbasic.Expression{
    superbasic.SQL("id"),
    superbasic.SQL("first"),
    superbasic.SQL("last"),
}

if len(columns) == 0 {
    columns = []superbasic.Expression{
        superbasic.SQL("*"),
    }
}

or := [2]superbasic.Expression{
    superbasic.SQL("last = ?", "Bush"),
    superbasic.SQL("first = ?", "Joe"),
}

where := superbasic.SQL(
    "WHERE ?",
    superbasic.SQL("? OR ?", or[0], or[1]),
)

query := superbasic.SQL(
    "SELECT ? FROM presidents ?",
    columns, // []superbasic.Expression gets translated to Join(", ", expr...)
    where,
)

fmt.Println(query.ToPostgres())
// SELECT id, first, last FROM presidents WHERE last = $1 OR first = $2
// [Bush Joe]
```
