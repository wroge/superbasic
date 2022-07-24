package superbasic_test

import (
	"testing"

	"github.com/wroge/superbasic"
)

type Select struct {
	Columns []superbasic.Sqlizer
	From    superbasic.Sqlizer
	Where   []superbasic.Sqlizer
}

func (s Select) Expression() superbasic.Expression {
	return superbasic.Append(
		superbasic.SQL("SELECT "),
		superbasic.If(len(s.Columns) > 0, s.Columns, superbasic.SQL("*")),
		superbasic.If(s.From != nil, superbasic.SQL(" FROM ?", s.From), superbasic.Skip()),
		superbasic.If(len(s.Where) > 0,
			superbasic.SQL(" WHERE ?",
				superbasic.If(len(s.Where) > 1,
					superbasic.Join(" AND ", s.Where),
					s.Where,
				),
			),
			superbasic.Skip(),
		),
	)
}

func TestSelect(t *testing.T) {
	query := Select{
		Columns: []superbasic.Sqlizer{
			superbasic.SQL("id"),
			superbasic.SQL("first"),
			superbasic.SQL("last"),
		},
		From: superbasic.SQL("presidents"),
		Where: []superbasic.Sqlizer{
			superbasic.SQL("last = ?", "Bush"),
		},
	}

	sql, args, err := superbasic.ToPostgres(query)
	if err != nil {
		t.Error(err)
	}

	if sql != "SELECT id, first, last FROM presidents WHERE last = $1" && len(args) != 1 && args[0] != "Bush" {
		t.Fail()
	}
}
