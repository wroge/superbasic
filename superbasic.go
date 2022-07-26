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

		sql, args, err := toSQL(e.Args[argsIndex], ", ")
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
	if len(expr) == 0 {
		return SQL("")
	}

	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expr))

	isFirst := true

	for _, e := range expr {
		sql, args, err := toSQL(e, sep)
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

type Query struct {
	Select  any
	From    any
	Where   any
	GroupBy any
	Having  any
	OrderBy any
	Limit   uint64
	Offset  uint64
}

func (q Query) ToSQL() (string, []any, error) {
	return Append(
		SQL("SELECT ?", IfElse(q.Select != nil, q.Select, SQL("*"))),
		If(q.From != nil, SQL(" FROM ?", q.From)),
		If(q.Where != nil, SQL(" WHERE ?", q.Where)),
		If(q.GroupBy != nil, SQL(" GROUP BY ?", q.GroupBy)),
		If(q.Having != nil, SQL(" HAVING ?", q.Having)),
		If(q.OrderBy != nil, SQL(" ORDER BY ?", q.OrderBy)),
		If(q.Limit > 0, SQL(fmt.Sprintf(" LIMIT %d", q.Limit))),
		If(q.Offset > 0, SQL(fmt.Sprintf(" OFFSET %d", q.Offset))),
	).ToSQL()
}

func Or(left, right Sqlizer) Expression {
	return SQL("(? OR ?)", left, right)
}

func And(expr ...Sqlizer) Expression {
	return Join(" AND ", expr)
}

func Not(expr Sqlizer) Expression {
	return SQL("NOT (?)", expr)
}

func Equals(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s = ?", ident), value)
}

func NotEquals(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s <> ?", ident), value)
}

func Greater(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s > ?", ident), value)
}

func GreaterOrEquals(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s >= ?", ident), value)
}

func Less(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s < ?", ident), value)
}

func LessOrEquals(ident string, value any) Expression {
	return SQL(fmt.Sprintf("%s <= ?", ident), value)
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
	Set   any
	Where any
}

func (u Update) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("UPDATE %s SET ?", u.Table), u.Set),
		If(u.Where != nil, SQL(" WHERE ?", u.Where)),
	).ToSQL()
}

type Delete struct {
	From  string
	Where any
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
	Columns     any
	Constraints any
}

func (ct Table) ToSQL() (string, []any, error) {
	return Append(
		SQL("CREATE TABLE"),
		If(ct.IfNotExists, SQL(" IF NOT EXISTS")),
		SQL(fmt.Sprintf(" %s (", ct.Name)),
		Join(", ", ct.Columns),
		If(ct.Constraints != nil, SQL(", ?", ct.Constraints)),
		SQL(")"),
	).ToSQL()
}

type Column struct {
	Name        string
	Type        string
	Constraints any
}

func (cs Column) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("%s %s", cs.Name, cs.Type)),
		If(cs.Constraints != nil, SQL(" ?", Join(" ", cs.Constraints))),
	).ToSQL()
}

func toExpression(expr any, sep string) (Expression, bool) {
	switch expression := expr.(type) {
	case Expression:
		return expression, true
	case Sqlizer:
		sql, args, err := expression.ToSQL()

		return Expression{SQL: sql, Args: args, Err: err}, true
	case []Expression:
		if len(expression) == 0 {
			return SQL(""), true
		}

		return Join(sep, anySlice(expression)...), true
	case []Sqlizer:
		if len(expression) == 0 {
			return SQL(""), true
		}

		return Join(sep, anySlice(expression)...), true
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

func toSQL(expr any, sep string) (string, []any, error) {
	ex, ok := toExpression(expr, sep)
	if !ok {
		return "?", []any{expr}, nil
	}

	return ex.ToSQL()
}

func ToDDL(expr any) (string, error) {
	sql, args, err := toSQL(expr, ", ")
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

	sql, args, err := toSQL(expr, ", ")
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
