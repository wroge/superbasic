//nolint:exhaustivestruct,exhaustruct
package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

type Sqlizer interface {
	ToSQL() (string, []any, error)
}

type Expression struct {
	SQL  string
	Args []any
	Err  error
}

func (e Expression) ToSQL() (string, []any, error) {
	if e.Err != nil {
		return "", nil, e.Err
	}

	builder := &strings.Builder{}
	arguments := make([]any, 0, len(e.Args))

	argsIndex := -1

	for {
		index := strings.IndexRune(e.SQL, '?')
		if index < 0 {
			builder.WriteString(e.SQL)

			break
		}

		if index < len(e.SQL)-1 && e.SQL[index+1] == '?' {
			builder.WriteString(e.SQL[:index+2])
			e.SQL = e.SQL[index+2:]

			continue
		}

		argsIndex++

		if argsIndex >= len(e.Args) {
			return "", nil, ErrInvalidNumberOfArguments
		}

		builder.WriteString(e.SQL[:index])
		e.SQL = e.SQL[index+1:]

		sql, args, err := ToSQL(e.Args[argsIndex])
		if err != nil {
			return "", nil, err
		}

		if sql == "" {
			continue
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if argsIndex != len(e.Args)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return builder.String(), arguments, nil
}

func SQL(sql string, args ...any) Expression {
	return Expression{SQL: sql, Args: args}
}

func Error(err error) Expression {
	return Expression{SQL: "", Args: nil, Err: err}
}

func Join(sep string, expr ...any) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expr))

	isFirst := true

	for _, e := range expr {
		sql, args, err := ToSQL(e)
		if err != nil {
			return Expression{Err: err}
		}

		if sql == "" {
			continue
		}

		if !isFirst {
			builder.WriteString(sep)
		}

		isFirst = false

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	return Expression{SQL: builder.String(), Args: arguments}
}

func Append(expr ...any) Expression {
	return Join("", expr...)
}

func If(cond bool, then any) Expression {
	if !cond {
		return SQL("")
	}

	return SQL("?", then)
}

func IfElse(cond bool, then any, els any) Expression {
	if cond {
		return SQL("?", then)
	}

	return SQL("?", els)
}

type Select struct {
	Distinct bool
	Columns  []Sqlizer
	From     []Sqlizer
	Joins    []Sqlizer
	Where    Sqlizer
	GroupBy  []Sqlizer
	Having   Sqlizer
	OrderBy  []Sqlizer
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
	Set   []Sqlizer
	Where Sqlizer
}

func (u Update) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("UPDATE %s SET ?", u.Table), u.Set),
		If(u.Where != nil, SQL(" WHERE ?", u.Where)),
	).ToSQL()
}

type Delete struct {
	From  string
	Where Sqlizer
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
	Columns     []Sqlizer
	Constraints []Sqlizer
}

func (ct Table) ToSQL() (string, []any, error) {
	return Append(
		SQL("CREATE TABLE"),
		If(ct.IfNotExists, SQL(" IF NOT EXISTS")),
		SQL(fmt.Sprintf(" %s (?)", ct.Name), Join(", ", append(ct.Columns, ct.Constraints...))),
	).ToSQL()
}

type Column struct {
	Name        string
	Type        string
	Constraints []Sqlizer
}

func (cs Column) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("%s %s", cs.Name, cs.Type)),
		If(len(cs.Constraints) > 0, SQL(" ?", cs.Constraints)),
	).ToSQL()
}

func ToExpression(expr any) (Expression, bool) {
	switch expression := expr.(type) {
	case Expression:
		return expression, true
	case Sqlizer:
		sql, args, err := expression.ToSQL()

		return Expression{SQL: sql, Args: args, Err: err}, true
	case []Expression:
		return Join(", ", anySlice(expression)...), true
	case []Sqlizer:
		return Join(", ", anySlice(expression)...), true
	default:
		return Expression{}, false
	}
}

func anySlice[T any](s []T) []any {
	out := make([]any, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}

func ToSQL(expr any) (string, []any, error) {
	ex, ok := ToExpression(expr)
	if !ok {
		return "?", []any{expr}, nil
	}

	return ex.ToSQL()
}

func ToDDL(expr any) (string, error) {
	sql, args, err := ToSQL(expr)
	if err != nil {
		return "", err
	}

	if len(args) > 0 {
		return "", ErrInvalidNumberOfArguments
	}

	return sql, nil
}

func NewBuilder() *Builder {
	return &Builder{
		builder: &strings.Builder{},
		args:    make([]any, 0, 1),
	}
}

type Builder struct {
	builder *strings.Builder
	args    []any
	err     error
}

func (b *Builder) ToSQL() (string, []any, error) {
	if b.err != nil {
		return "", nil, b.err
	}

	if b.builder == nil {
		return "", nil, nil
	}

	return b.builder.String(), b.args, nil
}

func (b *Builder) Write(expr any) *Builder {
	if b.err != nil {
		return b
	}

	if b.builder == nil {
		b.builder = &strings.Builder{}
	}

	sql, args, err := ToSQL(expr)
	if err != nil {
		b.err = err

		return b
	}

	b.builder.WriteString(sql)
	b.args = append(b.args, args...)

	return b
}

func (b *Builder) WriteSQL(sql string, args ...any) *Builder {
	return b.Write(SQL(sql, args...))
}

func ToPostgres(expr any) (string, []any, error) {
	sql, args, err := ToSQL(expr)
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
