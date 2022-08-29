# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.Compile``` compiles expressions into an SQL template and thus offers an alternative to conventional query builders.

If you need support for multiple SQL dialects, take a look at [wroge/esperanto](https://github.com/wroge/esperanto).

```go
package main

import (
	"fmt"

	"github.com/wroge/superbasic"
)

func main() {
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

	fmt.Println(superbasic.ToPositional("$", expr))
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
}
```
