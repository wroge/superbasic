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

	insert := superbasic.Compile("INSERT INTO presidents (?) VALUES ? RETURNING id",
		superbasic.Compile("nr, first, last"),
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

	update := superbasic.Compile("UPDATE presidents SET ? WHERE ?",
		[]superbasic.Expression{
			superbasic.Compile("first = ?", "Donald"),
			superbasic.Compile("last = ?", "Trump"),
		},
		superbasic.Compile("nr = ?", 45),
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

	search = superbasic.Joiner{
		Sep: " AND ",
		Expressions: []superbasic.Expression{
			superbasic.Compile("last IN ?", superbasic.Values{"Bush", "Clinton"}),
			superbasic.Compile("NOT (nr > ?)", 42),
		},
	}
	sort := "first"

	query := superbasic.Append(
		superbasic.Compile("SELECT nr, first, last FROM presidents"),
		superbasic.If(search != nil, superbasic.Compile(" WHERE ?", search)),
		superbasic.If(sort != "", superbasic.Compile(fmt.Sprintf(" ORDER BY %s", sort))),
	)

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

	del := superbasic.Compile("DELETE FROM presidents WHERE ?",
		superbasic.Joiner{
			Sep: " AND ",
			Expressions: []superbasic.Expression{
				superbasic.Compile("last = ?", "Bush"),
				superbasic.Compile("first = ?", "Joe"),
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

	expr := superbasic.Compile("?? hello ? ??", "world")

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

	sql, args, err := superbasic.Compile("?", []superbasic.Expression{
		superbasic.Compile("hello"),
		superbasic.Compile("world"),
		superbasic.IfElse(true, superbasic.Compile("welcome"), superbasic.Compile("moin")),
		superbasic.IfElse(false, superbasic.Compile("welcome"), superbasic.Compile("moin")),
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

	sql, args, err := superbasic.Join(", ",
		superbasic.Compile(""),
		superbasic.Compile("? ?", "hello"),
	).ToSQL()
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

	sql, args, err := superbasic.Compile("?", nil).ToSQL()
	if err.Error() != "invalid expression" {
		t.Fatal("ExpressionError", sql, args, err)
	}

	sql, args, err = superbasic.Compile("?", []superbasic.Expression{
		superbasic.Compile("hello"), superbasic.Compile("world"),
	}).ToSQL()

	if sql != "hello, world" {
		t.Fatal(sql, args, err)
	}

	sql, args, err = superbasic.Compile("?", []superbasic.Expression{
		superbasic.Compile("hello"), superbasic.Compile(""), superbasic.Compile("?"),
	}).ToSQL()
	if (!errors.Is(err, superbasic.NumberOfArgumentsError{})) {
		t.Fatal("NumberOfArgumentsError 2", sql, args, err)
	}

	sql, args, err = superbasic.Compile("?", []superbasic.Expression{
		superbasic.Compile("", []any{"hello"}),
	}).ToSQL()
	if sql != "" {
		t.Fatal(sql, args, err)
	}
}

func TestOr(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Compile("(? OR ?)", superbasic.Compile("hello"), superbasic.Compile("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "(hello OR moin)" {
		t.Fatal(sql, args, err)
	}
}

func TestNotEquals(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Compile("hello != ?", superbasic.Compile("moin")).ToSQL()
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

	sql, args, err = superbasic.ToPositional("$", superbasic.Compile("?"))
	if !errors.Is(err, superbasic.NumberOfArgumentsError{}) {
		t.Fatal(sql, args, err)
	}

	sql, args, err = superbasic.ToPositional("$", superbasic.Compile("?"))
	if !errors.Is(err, superbasic.NumberOfArgumentsError{}) {
		t.Fatal(sql, args, err)
	}
}

func TestQuery2(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Query{
		With:    superbasic.Compile("with"),
		Select:  superbasic.Compile("column"),
		From:    superbasic.Compile("from"),
		Where:   superbasic.Compile("where"),
		GroupBy: superbasic.Compile("group"),
		Having:  superbasic.Compile("having"),
		Window:  superbasic.Compile("window"),
		OrderBy: superbasic.Compile("order"),
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
			superbasic.Compile("first = ?", "Donald"),
			superbasic.Compile("last = ?", "Trump"),
		},
		Where: superbasic.Compile("nr = ?", 45),
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
		Where: superbasic.Compile("last = ?", "Bush"),
	}.ToSQL()

	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = ?" {
		t.Fatal(sql, args)
	}
}
