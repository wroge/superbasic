package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

type Expr interface {
	Expression() Expression
}

type Sqlizer interface {
	ToSQL() (string, []any, error)
}

var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

//nolint:cyclop
func toSQL(expr any, sep string) (string, []any, error) {
	switch exp := expr.(type) {
	case Expression:
		return exp.ToSQL()
	case Sqlizer:
		s, a, err := exp.ToSQL()
		if err != nil {
			return "", nil, fmt.Errorf("cannot resolve superbasic.Sqlizer: %w", err)
		}

		return s, a, nil
	case Expr:
		return exp.Expression().ToSQL()
	case []Expression:
		a := make([]any, len(exp))
		for i := range exp {
			a[i] = exp[i]
		}

		return Join(sep, a...).ToSQL()
	case []Sqlizer:
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

func If(condition bool, then any, els any) Expression {
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

func Skip() Expression {
	return SQL("")
}

func Error(err error) Expression {
	return Expression{SQL: "", Args: nil, Err: err}
}
