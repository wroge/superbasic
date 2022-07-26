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

type Columns []string

func (i Columns) ToSQL() (string, []any, error) {
	return strings.Join(i, ", "), nil, nil
}

type Values [][]any

func (v Values) ToSQL() (string, []any, error) {
	builder := NewBuilder()

	for i, d := range v {
		if i != 0 {
			builder.WriteSQL(", ")
		}

		builder.Write(SQL("(?)", Join(", ", d...)))
	}

	return builder.ToSQL()
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
