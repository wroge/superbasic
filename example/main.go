//nolint:gomnd,exhaustivestruct,forbidigo,funlen,wsl,exhaustruct
package main

import (
	"fmt"

	"github.com/wroge/superbasic"
)

func main() {
	expr := superbasic.SQL("INSERT INTO presidents (nr, first, last) VALUES ? RETURNING id",
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

	expr = superbasic.SQL("UPDATE presidents SET ? WHERE ?",
		[]superbasic.Expression{
			superbasic.SQL("first = ?", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		},
		superbasic.SQL("nr = ?", 45),
	)

	fmt.Println(expr.ToSQL())
	// UPDATE presidents SET first = ?, last = ? WHERE nr = ?
	// [Donald Trump 45]

	where := superbasic.SQL("(? OR ?)",
		superbasic.SQL("last = ?", "Bush"),
		superbasic.SQL("nr >= ?", 45),
	)

	builder := superbasic.Build().
		SQL("SELECT * FROM presidents").
		If(where.SQL != "", superbasic.SQL(" WHERE ?", where)).
		SQL(" ORDER BY nr LIMIT 3")

	fmt.Println(builder.ToSQL())
	// SELECT * FROM presidents WHERE (last = ? OR nr >= ?) ORDER BY nr LIMIT 3
	// [Bush 44]

	query := superbasic.Query{
		From:    superbasic.SQL("presidents"),
		Where:   where,
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

	fmt.Println(superbasic.SQL("? RETURNING id", insert).ToSQL())
	// INSERT INTO presidents (nr, first, last) VALUES
	// 		(?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?) RETURNING id
	// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]
}
