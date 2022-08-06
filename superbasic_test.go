//nolint:gofumpt,exhaustivestruct,exhaustruct,staticcheck,gosimple
package superbasic_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/wroge/superbasic"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	insert := superbasic.SQL("INSERT INTO presidents (?) VALUES ? RETURNING id",
		superbasic.SQL("nr, first, last"),
		[]superbasic.Values{
			{46, "Joe", "Biden"},
			{45, "Donald", "trump"},
			{44, "Barack", "Obama"},
			{43, "George W.", "Bush"},
			{42, "Bill", "Clinton"},
			{41, "George H. W.", "Bush"},
		},
	)

	sql, args, err := superbasic.ToPositional("$", insert)
	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (nr, first, last) VALUES"+
		" ($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12),"+
		" ($13, $14, $15), ($16, $17, $18) RETURNING id" || len(args) != 18 {
		t.Fatal(sql, args)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	update := superbasic.SQL("UPDATE presidents SET ? WHERE ?",
		[]superbasic.Expression{
			superbasic.SQL("first = ?", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		},
		superbasic.SQL("nr = ?", 45),
	)

	sql, args, err := update.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET first = ?, last = ? WHERE nr = ?" || len(args) != 3 {
		t.Fatal(sql, args)
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	var search superbasic.Expression

	search = superbasic.Join{
		Sep: " AND ",
		Expressions: []superbasic.Expression{
			superbasic.SQL("last IN ?", superbasic.Values{"Bush", "Clinton"}),
			superbasic.SQL("NOT (nr > ?)", 42),
		},
	}
	sort := "first"

	query := superbasic.Append{
		superbasic.SQL("SELECT nr, first, last FROM presidents"),
		superbasic.If(search != nil, superbasic.SQL(" WHERE ?", search)),
		superbasic.If(sort != "", superbasic.SQL(fmt.Sprintf(" ORDER BY %s", sort))),
	}

	sql, args, err := query.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT nr, first, last FROM presidents WHERE last IN (?, ?) AND NOT (nr > ?) ORDER BY first" ||
		len(args) != 3 {
		t.Fatal(sql, args)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	del := superbasic.SQL("DELETE FROM presidents WHERE ?",
		superbasic.Join{
			Sep: " AND ",
			Expressions: []superbasic.Expression{
				superbasic.SQL("last = ?", "Bush"),
				superbasic.SQL("first = ?", "Joe"),
			},
		},
	)

	sql, args, err := superbasic.ToPositional("$", del)
	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = $1 AND first = $2" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}

func TestEscape(t *testing.T) {
	t.Parallel()

	expr := superbasic.SQL("?? hello ? ??", "world")

	sql, args, err := superbasic.ToPositional("$", expr)
	if err != nil {
		t.Error(err)
	}

	if sql != "? hello $1 ?" || len(args) != 1 {
		t.Fatal(sql, args)
	}
}

func TestExpressionSlice(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.SQL("?", []superbasic.Expression{
		superbasic.SQL("hello"),
		superbasic.SQL("world"),
		superbasic.IfElse(true, superbasic.SQL("welcome"), superbasic.SQL("moin")),
		superbasic.IfElse(false, superbasic.SQL("welcome"), superbasic.SQL("moin")),
	}).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello, world, welcome, moin" || len(args) != 0 {
		t.Fatal(sql, args)
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Join{
		Sep: ", ",
		Expressions: []superbasic.Expression{
			superbasic.SQL(""),
			superbasic.SQL("? ?", "hello"),
		},
	}.ToSQL()
	if err.Error() != "invalid number of arguments: got '1' want '1'" {
		t.Fatal("NumberOfArgumentsError", sql, args, err)
	}
}

func TestIf(t *testing.T) {
	t.Parallel()

	sql, _, err := superbasic.If(false, nil).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "" {
		t.Fail()
	}
}

func TestSQL(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.SQL("?", nil).ToSQL()
	if err.Error() != "invalid expression" {
		t.Fatal("ExpressionError", sql, args, err)
	}

	sql, args, err = superbasic.SQL("?", []superbasic.Expression{
		superbasic.SQL("hello"), superbasic.SQL("world"),
	}).ToSQL()

	if sql != "hello, world" {
		t.Fatal(sql, args, err)
	}

	sql, args, err = superbasic.SQL("?", []superbasic.Expression{
		superbasic.SQL("hello"), superbasic.SQL(""), superbasic.SQL("?"),
	}).ToSQL()
	if (!errors.Is(err, superbasic.NumberOfArgumentsError{})) {
		t.Fatal("NumberOfArgumentsError 2", sql, args, err)
	}

	sql, args, err = superbasic.SQL("?", []superbasic.Expression{
		superbasic.SQL("", []any{"hello"}),
	}).ToSQL()
	if sql != "" {
		t.Fatal(sql, args, err)
	}
}

func TestOr(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.SQL("(? OR ?)", superbasic.SQL("hello"), superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "(hello OR moin)" {
		t.Fatal(sql, args, err)
	}
}

func TestNotEquals(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.SQL("hello != ?", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello != moin" {
		t.Fatal(sql, args, err)
	}
}

func TestPositional(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.ToPositional("$", nil)
	if !errors.Is(err, superbasic.ExpressionError{}) {
		t.Fatal(sql, args, err)
	}

	sql, args, err = superbasic.ToPositional("$", superbasic.SQL("?"))
	if !errors.Is(err, superbasic.NumberOfArgumentsError{}) {
		t.Fatal(sql, args, err)
	}

	sql, args, err = superbasic.ToPositional("$", superbasic.SQL("?"))
	if !errors.Is(err, superbasic.NumberOfArgumentsError{}) {
		t.Fatal(sql, args, err)
	}
}

func TestQuery2(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Query{
		With:    superbasic.SQL("with"),
		Select:  superbasic.SQL("column"),
		From:    superbasic.SQL("from"),
		Where:   superbasic.SQL("where"),
		GroupBy: superbasic.SQL("group"),
		Having:  superbasic.SQL("having"),
		Window:  superbasic.SQL("window"),
		OrderBy: superbasic.SQL("order"),
		Limit:   1,
		Offset:  1,
	}.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "WITH with SELECT column FROM from WHERE where GROUP BY group HAVING having"+
		" WINDOW window ORDER BY order LIMIT 1 OFFSET 1" {
		t.Fatal(sql, args)
	}
}

func TestInsert2(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Insert{
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
	}.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (nr, first, last) VALUES"+
		" (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?)" {
		t.Fatal(sql, args, err)
	}
}

func TestUpdate2(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Update{
		Table: "presidents",
		Sets: []superbasic.Expression{
			superbasic.SQL("first = ?", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		},
		Where: superbasic.SQL("nr = ?", 45),
	}.ToSQL()

	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET first = ?, last = ? WHERE nr = ?" {
		t.Fatal(sql, args)
	}
}

func TestDelete2(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Delete{
		From:  "presidents",
		Where: superbasic.SQL("last = ?", "Bush"),
	}.ToSQL()

	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = ?" {
		t.Fatal(sql, args)
	}
}
