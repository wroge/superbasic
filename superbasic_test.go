package superbasic_test

import (
	"testing"

	"github.com/wroge/superbasic"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	insert := superbasic.SQL("INSERT INTO presidents (?) VALUES ? RETURNING id",
		superbasic.Columns{"first", "last"},
		superbasic.Values{
			{"Joe", "Biden"},
			{"Donald", "trump"},
			{"Barack", "Obama"},
			{"George W.", "Bush"},
			{"Bill", "Clinton"},
			{"George H. W.", "Bush"},
		},
	)

	sql, args, err := superbasic.ToPostgres(insert)
	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (first, last) "+
		"VALUES ($1, $2), ($3, $4), ($5, $6), ($7, $8), ($9, $10), ($11, $12) RETURNING id" || len(args) != 12 {
		t.Fatal(sql, args)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	update := superbasic.SQL("UPDATE presidents SET last = ? WHERE ?",
		"Trump",
		superbasic.SQL("first = ?", "Donald"),
	)

	sql, args, err := update.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET last = ? WHERE first = ?" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	columns := []string{"id", "first", "last"}
	lastname := "Bush"
	sort := "first"

	query := superbasic.Append(
		superbasic.SQL("SELECT ? FROM presidents", superbasic.Columns(columns)),
		superbasic.If(lastname != "", superbasic.SQL(" WHERE last = ?", lastname)),
		superbasic.If(sort != "", superbasic.SQL(" ORDER BY ?", superbasic.SQL(sort))),
	)

	sql, args, err := query.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT id, first, last FROM presidents WHERE last = ? ORDER BY first" || len(args) != 1 {
		t.Fatal(sql, args)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	del := superbasic.SQL("DELETE FROM presidents WHERE ?",
		superbasic.Join(" AND ",
			superbasic.SQL("last = ?", "Bush"),
			superbasic.SQL("first = ?", "Joe"),
		),
	)

	sql, args, err := superbasic.ToPostgres(del)
	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = $1 AND first = $2" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}
