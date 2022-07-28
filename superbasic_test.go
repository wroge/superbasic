package superbasic_test

import (
	"fmt"
	"testing"

	"github.com/wroge/superbasic"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	insert := superbasic.Compile("INSERT INTO presidents (?) VALUES ? RETURNING id",
		superbasic.Idents("id", "first", "last"),
		superbasic.Join(", ",
			superbasic.Values(46, "Joe", "Biden"),
			superbasic.Values(45, "Donald", "trump"),
			superbasic.Values(44, "Barack", "Obama"),
			superbasic.Values(43, "George W.", "Bush"),
			superbasic.Values(42, "Bill", "Clinton"),
			superbasic.Values(41, "George H. W.", "Bush"),
		),
	)

	sql, args, err := insert.ToPositional("$")
	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (id, first, last) VALUES"+
		" ($1, $2, $3), ($4, $5, $6), ($7, $8, $9), ($10, $11, $12),"+
		" ($13, $14, $15), ($16, $17, $18) RETURNING id" || len(args) != 18 {
		t.Fatal(sql, args)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	update := superbasic.Compile("UPDATE presidents SET ? WHERE ?",
		superbasic.Join(", ",
			superbasic.Equals("first", "Donald"),
			superbasic.Equals("last", "Trump"),
		),
		superbasic.Equals("id", 45),
	)

	sql, args, err := update.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET first = ?, last = ? WHERE id = ?" || len(args) != 3 {
		t.Fatal(sql, args)
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	search := superbasic.In("last", "Bush", "Clinton")
	sort := "first"

	query := superbasic.Append(
		superbasic.SQL("SELECT id, first, last FROM presidents"),
		superbasic.If(search.SQL != "", superbasic.Compile(" WHERE ?", search)),
		superbasic.If(sort != "", superbasic.SQL(fmt.Sprintf(" ORDER BY %s", sort))),
	)

	sql, args, err := query.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT id, first, last FROM presidents WHERE last IN (?, ?) ORDER BY first" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	del := superbasic.Compile("DELETE FROM presidents WHERE ?",
		superbasic.Join(" AND ",
			superbasic.Equals("last", "Bush"),
			superbasic.Equals("first", "Joe"),
		),
	)

	sql, args, err := del.ToPositional("$")
	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = $1 AND first = $2" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}

func TestEscape(t *testing.T) {
	t.Parallel()

	expr := superbasic.Compile("?? hello ? ??", superbasic.Value("world"))
	if expr.Err != nil {
		t.Error(expr.Err)
	}

	sql, args, err := expr.ToPositional("$")
	if err != nil {
		t.Error(err)
	}

	if sql != "? hello $1 ?" || len(args) != 1 {
		t.Fatal(sql, args)
	}
}

func TestExpressionSlice(t *testing.T) {
	t.Parallel()

	expr := superbasic.Compile("?", superbasic.Join(", ",
		superbasic.SQL("hello"),
		superbasic.SQL("world"),
		superbasic.IfElse(true, superbasic.SQL("welcome"), superbasic.SQL("moin")),
		superbasic.IfElse(false, superbasic.SQL("welcome"), superbasic.SQL("moin")),
	))
	if expr.Err != nil {
		t.Error(expr.Err)
	}

	if expr.SQL != "hello, world, welcome, moin" || len(expr.Args) != 0 {
		t.Fatal(expr.SQL, expr.Err)
	}
}
