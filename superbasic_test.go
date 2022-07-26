//nolint:goconst,exhaustivestruct,exhaustruct
package superbasic_test

import (
	"testing"

	"github.com/wroge/superbasic"
)

func TestTable(t *testing.T) {
	t.Parallel()

	table := superbasic.Table{
		IfNotExists: true,
		Name:        "presidents",
		Columns: []superbasic.Sqlizer{
			superbasic.Column{
				Name: "id",
				Type: "SERIAL",
				Constraints: []superbasic.Sqlizer{
					superbasic.SQL("PRIMARY KEY"),
				},
			},
			superbasic.Column{
				Name: "first",
				Type: "TEXT",
				Constraints: []superbasic.Sqlizer{
					superbasic.SQL("NOT NULL"),
				},
			},
			superbasic.Column{
				Name: "last",
				Type: "TEXT",
				Constraints: []superbasic.Sqlizer{
					superbasic.SQL("NOT NULL"),
				},
			},
		},
		Constraints: []superbasic.Sqlizer{
			superbasic.SQL("UNIQUE (first, last)"),
		},
	}

	sql, err := superbasic.ToDDL(table)
	if err != nil {
		t.Error(err)
	}

	if sql != "CREATE TABLE IF NOT EXISTS presidents (id SERIAL PRIMARY KEY, first TEXT NOT NULL, "+
		"last TEXT NOT NULL, UNIQUE (first, last))" {
		t.Fatal(sql)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	expr := superbasic.Delete{
		From:  "presidents",
		Where: superbasic.SQL("last = ?", "Bush"),
	}

	sql, args, err := superbasic.ToPostgres(expr)
	if err != nil {
		t.Error(err)
	}

	if sql != "DELETE FROM presidents WHERE last = $1" ||
		len(args) != 1 || args[0] != "Bush" {
		t.Fatal(sql, args)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	joe := "Joe"
	biden := "Biden"

	update := superbasic.Update{
		Table: "presidents",
		Set: []superbasic.Sqlizer{
			superbasic.SQL("last = ?", biden),
		},
		Where: superbasic.SQL("first = ?", joe),
	}

	sql, args, err := superbasic.ToPostgres(update)
	if err != nil {
		t.Error(err)
	}

	if sql != "UPDATE presidents SET last = $1 WHERE first = $2" ||
		len(args) != 2 || args[0] != biden || args[1] != joe {
		t.Fatal(sql, args)
	}
}

func TestInsert(t *testing.T) {
	t.Parallel()

	insert := superbasic.Append(
		superbasic.Insert{
			Into:    "presidents",
			Columns: []string{"first", "last"},
			Values: [][]any{
				{"Joe", "Bden"},
				{"Donald", "Trump"},
				{"Barack", "Obama"},
				{"George W.", "Bush"},
				{"Bill", "Clinton"},
				{"George H. W.", "Bush"},
			},
		},
		superbasic.SQL(" RETURNING id"),
	)

	sql, args, err := superbasic.ToPostgres(insert)
	if err != nil {
		t.Error(err)
	}

	if sql != "INSERT INTO presidents (first, last) VALUES ($1, $2), ($3, $4), ($5, $6),"+
		" ($7, $8), ($9, $10), ($11, $12) RETURNING id" ||
		len(args) != 12 || args[0] != "Joe" || args[1] != "Bden" || args[10] != "George H. W." || args[11] != "Bush" {
		t.Fatal(sql, args)
	}
}

func TestSelect(t *testing.T) {
	t.Parallel()

	joe := "Joe"

	query := superbasic.Select{
		Columns: []superbasic.Sqlizer{
			superbasic.SQL("id"),
			superbasic.SQL("first"),
			superbasic.SQL("last"),
		},
		From: []superbasic.Sqlizer{
			superbasic.SQL("presidents"),
		},
		Where: superbasic.SQL("? OR ?",
			superbasic.SQL("last = ?", "Bush"), superbasic.SQL("first = ?", joe)),
		OrderBy: []superbasic.Sqlizer{
			superbasic.SQL("last"),
		},
		Limit: 3,
	}

	sql, args, err := superbasic.ToPostgres(query)
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT id, first, last FROM presidents WHERE last = $1 OR first = $2 ORDER BY last LIMIT 3" ||
		len(args) != 2 || args[0] != "Bush" || args[1] != joe {
		t.Fatal(sql, args)
	}
}

func TestBuilder(t *testing.T) {
	b := superbasic.NewBuilder()

	b.WriteSQL("SELECT ").WriteSQL("first, last")
	b.WriteSQL(" FROM presidents")
	b.WriteSQL(" WHERE ")
	b.Write(superbasic.Join(" OR ", superbasic.SQL("last = ?", "Bush"), superbasic.SQL("first = ?", "Joe")))

	sql, args, err := b.ToSQL()
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT first, last FROM presidents WHERE last = ? OR first = ?" || len(args) != 2 {
		t.Fatal(sql, args)
	}
}
