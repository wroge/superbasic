//nolint:gofumpt
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
		[][]any{
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
		superbasic.Join(", ",
			superbasic.EqualsIdent("first", "Donald"),
			superbasic.EqualsIdent("last", "Trump"),
		),
		superbasic.EqualsIdent("nr", 45),
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

	search := superbasic.And(
		superbasic.InIdent("last", []any{"Bush", "Clinton"}),
		superbasic.Not(superbasic.GreaterIdent("nr", 42)),
	)
	sort := "first"

	query := superbasic.Append(
		superbasic.SQL("SELECT nr, first, last FROM presidents"),
		superbasic.If(search != nil, superbasic.SQL(" WHERE ?", search)),
		superbasic.If(sort != "", superbasic.SQL(fmt.Sprintf(" ORDER BY %s", sort))),
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

	del := superbasic.SQL("DELETE FROM presidents WHERE ?",
		superbasic.Join(" AND ",
			superbasic.EqualsIdent("last", "Bush"),
			superbasic.Equals(superbasic.SQL("first"), "Joe"),
		),
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

	sql, args, err := superbasic.SQL("?", superbasic.Join(", ",
		superbasic.SQL("hello"),
		superbasic.SQL("world"),
		superbasic.IfElse(true, superbasic.SQL("welcome"), superbasic.SQL("moin")),
		superbasic.IfElse(false, superbasic.SQL("welcome"), superbasic.SQL("moin")),
	)).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello, world, welcome, moin" || len(args) != 0 {
		t.Fatal(sql, args)
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Join(", ", nil).ToSQL()
	if (!errors.Is(err, superbasic.ExpressionError{})) {
		t.Fatal("ExpressionError", sql, args, err)
	}

	sql, args, err = superbasic.Join(", ", superbasic.SQL(""), superbasic.SQL("?")).ToSQL()
	if err.Error() != "invalid number of arguments" {
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

	sql, args, err := superbasic.Or(superbasic.SQL("hello"), superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "(hello OR moin)" {
		t.Fatal(sql, args, err)
	}
}

func TestNotEquals(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.NotEqualsIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello <> moin" {
		t.Fatal(sql, args, err)
	}
}

func TestGreaterOrEqualsIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.GreaterOrEqualsIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello >= moin" {
		t.Fatal(sql, args, err)
	}
}

func TestLessIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.LessIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello < moin" {
		t.Fatal(sql, args, err)
	}
}

func TestLessOrEqualsIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.LessOrEqualsIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello <= moin" {
		t.Fatal(sql, args, err)
	}
}

func TestNotInIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.NotInIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello NOT IN moin" {
		t.Fatal(sql, args, err)
	}
}

func TestIsNullIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.IsNullIdent("hello").ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello IS NULL" {
		t.Fatal(sql, args, err)
	}
}

func TestIsNotNullIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.IsNotNullIdent("hello").ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello IS NOT NULL" {
		t.Fatal(sql, args, err)
	}
}

func TestBetweenIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.BetweenIdent("hello", 1, 2).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello BETWEEN ? AND ?" {
		t.Fatal(sql, args, err)
	}
}

func TestNotBetweenIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.NotBetweenIdent("hello", 1, 2).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello NOT BETWEEN ? AND ?" {
		t.Fatal(sql, args, err)
	}
}

func TestLikeIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.LikeIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello LIKE moin" {
		t.Fatal(sql, args, err)
	}
}

func TestNotLikeIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.NotLikeIdent("hello", superbasic.SQL("moin")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "hello NOT LIKE moin" {
		t.Fatal(sql, args, err)
	}
}

func TestCastIdent(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.CastIdent("hello", "TEXT").ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "CAST(hello AS TEXT)" {
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

	sql, args, err = superbasic.ToPositional("$", superbasic.SQL("?", "hello", "world"))
	if !errors.Is(err, superbasic.NumberOfArgumentsError{}) {
		t.Fatal(sql, args, err)
	}
}

func TestSelectBuilder(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Select("select").
		From("from").
		Where("where").
		GroupBy("group").
		Having("having").
		OrderBy("order").
		Limit(1).
		Offset(1).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT select FROM from WHERE where GROUP BY group HAVING having ORDER BY order LIMIT 1 OFFSET 1" {
		t.Fatal(sql, args)
	}
}

func TestInsertBuilder(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Insert("presidents").
		Columns("nr", "first", "last").
		Values(46, "Joe", "Biden").
		Values(45, "Donald", "trump").
		Values(44, "Barack", "Obama").
		Values(43, "George W.", "Bush").
		Values(42, "Bill", "Clinton").
		Values(41, "George H. W.", "Bush").
		ToSQL()

	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (nr, first, last) VALUES"+
		" (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?)" {
		t.Fatal(sql, args, err)
	}
}

func TestUpdateBuilder(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Update("presidents").
		SetExpr(superbasic.EqualsIdent("first", "Donald")).
		Set("last = ?", "Trump").
		WhereExpr(superbasic.EqualsIdent("nr", 45)).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET first = ?, last = ? WHERE nr = ?" {
		t.Fatal(sql, args)
	}
}

func TestDeleteBuilder(t *testing.T) {
	t.Parallel()

	sql, args, err := superbasic.Delete("presidents").
		WhereExpr(superbasic.EqualsIdent("last", "Bush")).ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = ?" {
		t.Fatal(sql, args)
	}
}
