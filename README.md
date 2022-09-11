# The superbasic SQL-Builder

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wroge/superbasic)
[![Go Report Card](https://goreportcard.com/badge/github.com/wroge/superbasic)](https://goreportcard.com/report/github.com/wroge/superbasic)
![golangci-lint](https://github.com/wroge/superbasic/workflows/golangci-lint/badge.svg)
[![codecov](https://codecov.io/gh/wroge/superbasic/branch/main/graph/badge.svg?token=SBSedMOGHR)](https://codecov.io/gh/wroge/superbasic)
[![tippin.me](https://badgen.net/badge/%E2%9A%A1%EF%B8%8Ftippin.me/@_wroge/F0918E)](https://tippin.me/@_wroge)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/wroge/superbasic.svg?style=social)](https://github.com/wroge/superbasic/tags)

```superbasic.Compile``` compiles expressions into an SQL template and thus offers an alternative to conventional query builders.

If you need support for multiple SQL dialects, take a look at [wroge/esperanto](https://github.com/wroge/esperanto).
To scan rows to types, i recommend [wroge/scan](https://github.com/wroge/scan).

```go
package main

import (
	"fmt"

	"github.com/wroge/superbasic"
)

func main() {
	// 1. CREATE SCHEMA

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

	// 2. INSERT

	insert := superbasic.Join(" ",
		superbasic.SQL("INSERT INTO presidents (first, last)"),
		superbasic.Compile("VALUES ?",
			superbasic.Join(", ",
				superbasic.Map(presidents,
					func(president President) superbasic.Expression {
						return superbasic.Values{president.First, president.Last}
					})...,
			),
		),
		superbasic.SQL("RETURNING nr"),
	)

	fmt.Println(superbasic.Finalize("$%d", insert))
	// INSERT INTO presidents (first, last) VALUES ($1, $2), ($3, $4) RETURNING nr [George Washington John Adams] nil
}

type President struct {
	First string
	Last  string
}

var presidents = []President{
	{"George", "Washington"},
	{"John", "Adams"},
	// {"Thomas", "Jefferson"},
	// {"James", "Madison"},
	// {"James", "Monroe"},
	// {"John Quincy", "Adams"},
	// {"Andrew", "Jackson"},
	// {"Martin", "Van Buren"},
	// {"William Henry", "Harrison"},
	// {"John", "Tyler"},
	// {"James K.", "Polk"},
	// {"Zachary", "Taylor"},
	// {"Millard", "Fillmore"},
	// {"Franklin", "Pierce"},
	// {"James", "Buchanan"},
	// {"Abraham", "Lincoln"},
	// {"Andrew", "Johnson"},
	// {"Ulysses S.", "Grant"},
	// {"Rutherford B.", "Hayes"},
	// {"James A.", "Garfield"},
	// {"Chester A.", "Arthur"},
	// {"Grover", "Cleveland"},
	// {"Benjamin", "Harrison"},
	// {"Grover", "Cleveland"},
	// {"William", "McKinley"},
	// {"Theodore", "Roosevelt"},
	// {"William Howard", "Taft"},
	// {"Woodrow", "Wilson"},
	// {"Warren G.", "Harding"},
	// {"Calvin", "Coolidge"},
	// {"Herbert", "Hoover"},
	// {"Franklin D.", "Roosevelt"},
	// {"Harry S.", "Truman"},
	// {"Dwight D.", "Eisenhower"},
	// {"John F.", "Kennedy"},
	// {"Lyndon B.", "Johnson"},
	// {"Richard", "Nixon"},
	// {"Gerald", "Ford"},
	// {"Jimmy", "Carter"},
	// {"Ronald", "Reagan"},
	// {"George H. W.", "Bush"},
	// {"Bill", "Clinton"},
	// {"George W.", "Bush"},
	// {"Barack", "Obama"},
	// {"Donald", "Trump"},
	// {"Joe", "Biden"},
}
```
