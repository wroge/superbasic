package superbasic_test

import (
	"testing"

	"github.com/wroge/superbasic"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	insert := superbasic.SQL("INSERT INTO presidents (?) VALUES ? RETURNING id",
		superbasic.Columns{"id", "first", "last"},
		superbasic.Values{
			{46, "Joe", "Biden"},
			{45, "Donald", "trump"},
			{44, "Barack", "Obama"},
			{43, "George W.", "Bush"},
			{42, "Bill", "Clinton"},
			{41, "George H. W.", "Bush"},
		},
	)

	sql, args, err := superbasic.ToPostgres(insert)
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

	update := superbasic.SQL("UPDATE presidents SET ? WHERE ?",
		[]superbasic.Sqlizer{
			superbasic.SQL("first = ?", "Donald"),
			superbasic.SQL("last = ?", "Trump"),
		},
		superbasic.SQL("id = ?", 45),
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

	columns := []string{"id", "first", "last"}
	lastnames := []any{"Bush", "Clinton"}
	sort := "first"

	query := superbasic.Append(
		superbasic.SQL("SELECT ? FROM presidents", superbasic.Columns(columns)),
		superbasic.If(len(lastnames) > 0, superbasic.SQL(" WHERE last IN ?", superbasic.Values{lastnames})),
		superbasic.If(sort != "", superbasic.SQL(" ORDER BY ?", superbasic.SQL(sort))),
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
