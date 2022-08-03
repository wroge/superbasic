package main

import (
	"fmt"
	"strings"

	"github.com/wroge/superbasic"
)

func main() {
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

	query = superbasic.Query{
		Select:  superbasic.SQL(strings.Join(columns, ", ")),
		From:    superbasic.SQL("presidents"),
		Where:   where,
		OrderBy: superbasic.SQL("nr"),
	}
	fmt.Println(query.ToSQL())
	// SELECT nr, first, last FROM presidents WHERE (last = ? OR nr >= ?) ORDER BY nr LIMIT 3
	// [Bush 44]

	insert = superbasic.Insert{
		Into:    "presidents",
		Columns: []string{"nr", "first", "last"},
		Data: [][]any{
			{46, "Joe", "Biden"},
			{45, "Donald", "trump"},
			{44, "Barack", "Obama"},
			{43, "George W.", "Bush"},
			{42, "Bill", "Clinton"},
			{41, "George H. W.", "Bush"},
		},
	}

	fmt.Println(superbasic.Join(" ", insert, superbasic.SQL("RETURNING id")).ToSQL())
	// INSERT INTO presidents (nr, first, last) VALUES
	// 		(?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?) RETURNING id
	// [46 Joe Biden 45 Donald trump 44 Barack Obama 43 George W. Bush 42 Bill Clinton 41 George H. W. Bush]
}
