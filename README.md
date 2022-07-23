# The superbasic SQL-Builder

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

This library was written to be able to build dynamic queries in the simplest way possible. There is a ```ToPostgres``` function to translate expressions with ```$``` placeholders.

```go
var where superbasic.Expression

where = superbasic.SQL(
    "WHERE ? OR ?",
    superbasic.SQL("last = ?", "Bush"),
    superbasic.SQL("first = ?", "Joe"),
)

query := superbasic.SQL(
    "SELECT ? FROM presidents ?",
    superbasic.Join(", ",
        superbasic.SQL("id"),
        superbasic.SQL("first"),
        superbasic.SQL("last"),
    ),
    where,
)

fmt.Println(query.ToPostgres())
// SELECT id, first, last FROM presidents WHERE last = $1 OR first = $2
// [Bush Joe]
```
