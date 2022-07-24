package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

type Expr interface {
	ToSQL() (string, []any, error)
}

type ExprDDL interface {
	ToDDL() (string, error)
}

var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

//nolint:cyclop
func toSQL(expr any, sep string) (string, []any, error) {
	switch exp := expr.(type) {
	case Expression:
		return exp.ToSQL()
	case Expr:
		s, a, err := exp.ToSQL()
		if err != nil {
			return "", nil, fmt.Errorf("cannot resolve superbasic.Expr: %w", err)
		}

		return s, a, nil
	case []Expression:
		a := make([]any, len(exp))
		for i := range exp {
			a[i] = exp[i]
		}

		return Join(sep, a...).ToSQL()
	case []Expr:
		a := make([]any, len(exp))
		for i := range exp {
			a[i] = exp[i]
		}

		return Join(sep, a...).ToSQL()
	case ExprDDL:
		s, err := exp.ToDDL()
		if err != nil {
			return "", nil, fmt.Errorf("cannot resolve superbasic.ExprDDL: %w", err)
		}

		return s, nil, nil
	case []ExprDDL:
		a := make([]any, len(exp))
		for i := range exp {
			a[i] = exp[i]
		}

		return Join(sep, a...).ToSQL()
	default:
		return "?", []any{exp}, nil
	}
}

func SQL(sql string, args ...any) Expression {
	return Expression{
		SQL:  sql,
		Args: args,
		Err:  nil,
	}
}

type Expression struct {
	SQL  string
	Args []any
	Err  error
}

func (e Expression) ToDDL() (string, error) {
	sql, args, err := e.ToSQL()
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		return "", ErrInvalidNumberOfArguments
	}

	return sql, nil
}

func (e Expression) ToSQL() (string, []any, error) {
	if e.Err != nil {
		return "", nil, e.Err
	}

	build := &strings.Builder{}
	arguments := make([]any, 0, len(e.Args))

	argIndex := -1

	for {
		index := strings.IndexRune(e.SQL, '?')
		if index < 0 {
			build.WriteString(e.SQL)

			break
		}

		if index < len(e.SQL)-1 && e.SQL[index+1] == '?' {
			build.WriteString(e.SQL[:index+2])
			e.SQL = e.SQL[index+2:]

			continue
		}

		argIndex++

		if argIndex >= len(e.Args) {
			return "", nil, ErrInvalidNumberOfArguments
		}

		build.WriteString(e.SQL[:index])
		e.SQL = e.SQL[index+1:]

		sql, args, err := toSQL(e.Args[argIndex], ", ")
		if err != nil {
			return "", nil, err
		}

		if sql == "" {
			continue
		}

		build.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if argIndex != len(e.Args)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return build.String(), arguments, nil
}

func ToPostgres(expr any) (string, []any, error) {
	sql, args, err := toSQL(expr, ", ")
	if err != nil {
		return "", nil, err
	}

	build := &strings.Builder{}
	argIndex := -1

	for {
		index := strings.IndexRune(sql, '?')
		if index < 0 {
			build.WriteString(sql)

			break
		}

		if index < len(sql)-1 && sql[index+1] == '?' {
			build.WriteString(sql[:index+1])
			sql = sql[index+2:]

			continue
		}

		argIndex++

		build.WriteString(fmt.Sprintf("%s$%d", sql[:index], argIndex+1))
		sql = sql[index+1:]
	}

	if argIndex != len(args)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return build.String(), args, nil
}

func Join(sep string, expr ...any) Expression {
	if sep == "" {
		return Append(expr...)
	}

	build := &strings.Builder{}
	arguments := make([]any, 0, len(expr))

	argIndex := 0

	for _, e := range expr {
		sql, args, err := toSQL(e, sep)
		if err != nil {
			return Error(err)
		}

		if sql == "" {
			continue
		}

		if argIndex != 0 {
			build.WriteString(sep)
		}

		build.WriteString(sql)

		arguments = append(arguments, args...)

		argIndex++
	}

	return SQL(build.String(), arguments...)
}

func Append(expr ...any) Expression {
	build := &strings.Builder{}
	arguments := make([]any, 0, len(expr))

	for _, e := range expr {
		sql, args, err := toSQL(e, ", ")
		if err != nil {
			return Error(err)
		}

		if sql == "" {
			continue
		}

		build.WriteString(sql)

		arguments = append(arguments, args...)
	}

	return SQL(build.String(), arguments...)
}

func If(condition bool, then any) Expression {
	if condition {
		s, a, err := toSQL(then, ", ")
		if err != nil {
			return Error(err)
		}

		return SQL(s, a...)
	}

	return SQL("")
}

func IfElse(condition bool, then any, els any) Expression {
	if condition {
		s, a, err := toSQL(then, ", ")
		if err != nil {
			return Error(err)
		}

		return SQL(s, a...)
	}

	s, a, err := toSQL(els, ", ")
	if err != nil {
		return Error(err)
	}

	return SQL(s, a...)
}

func Error(err error) Expression {
	return Expression{SQL: "", Args: nil, Err: err}
}

type Select struct {
	Distinct bool
	Columns  []Expr
	From     []Expr
	Joins    []Expr
	Where    Expr
	GroupBy  []Expr
	Having   Expr
	OrderBy  []Expr
	Limit    uint64
	Offset   uint64
}

func (s Select) ToSQL() (string, []any, error) {
	return Append(
		SQL("SELECT "),
		If(s.Distinct, SQL("DISTINCT ")),
		IfElse(len(s.Columns) > 0, s.Columns, SQL("*")),
		If(len(s.From) > 0, SQL(" FROM ?", s.From)),
		If(len(s.Joins) > 0, SQL(" ?", s.Joins)),
		If(s.Where != nil, SQL(" WHERE ?", s.Where)),
		If(len(s.GroupBy) > 0, SQL(" GROUP BY ?", s.GroupBy)),
		If(s.Having != nil, SQL(" HAVING ?", s.Having)),
		If(len(s.OrderBy) > 0, SQL(" ORDER BY ?", s.OrderBy)),
		If(s.Limit > 0, SQL(fmt.Sprintf(" LIMIT %d", s.Limit))),
		If(s.Offset > 0, SQL(fmt.Sprintf(" OFFSET %d", s.Offset))),
	).ToSQL()
}

func Values(data ...[]any) Expression {
	values := make([]Expression, len(data))

	for j, d := range data {
		values[j] = SQL("(?)", Join(", ", d...))
	}

	return SQL("?", values)
}

type Insert struct {
	Into    string
	Columns []string
	Values  [][]any
}

func (i Insert) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("INSERT INTO %s ", i.Into)),
		If(len(i.Columns) > 0, SQL(fmt.Sprintf("(%s) ", strings.Join(i.Columns, ", ")))),
		SQL("VALUES ?", Values(i.Values...)),
	).ToSQL()
}

type Update struct {
	Table string
	Set   []Expr
	Where Expr
}

func (u Update) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("UPDATE %s SET ?", u.Table), u.Set),
		If(u.Where != nil, SQL(" WHERE ?", u.Where)),
	).ToSQL()
}

type Delete struct {
	From  string
	Where Expr
}

func (d Delete) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("DELETE FROM %s", d.From)),
		If(d.Where != nil, SQL(" WHERE ?", d.Where)),
	).ToSQL()
}

type Table struct {
	IfNotExists bool
	Name        string
	Columns     []ExprDDL
	Constraints []ExprDDL
}

func (ct Table) ToDDL() (string, error) {
	return Append(
		SQL("CREATE TABLE"),
		If(ct.IfNotExists, SQL(" IF NOT EXISTS")),
		SQL(fmt.Sprintf(" %s (?)", ct.Name), Join(", ", append(ct.Columns, ct.Constraints...))),
	).ToDDL()
}

type Column struct {
	Name        string
	Type        string
	Constraints []ExprDDL
}

func (cs Column) ToDDL() (string, error) {
	return Append(
		SQL(fmt.Sprintf("%s %s", cs.Name, cs.Type)),
		If(len(cs.Constraints) > 0, SQL(" ?", cs.Constraints)),
	).ToDDL()
}
