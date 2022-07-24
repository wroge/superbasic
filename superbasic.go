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

func toSQL(expr any, sep string) (string, []any, error) {
	switch t := expr.(type) {
	case Expression:
		return t.ToSQL()
	case Sqlizer:
		return t.ToSQL()
	case Expr:
		return t.Expression().ToSQL()
	case []Expression:
		a := make([]any, len(t))
		for i := range t {
			a[i] = t[i]
		}

		return Join(sep, a...).ToSQL()
	case []Sqlizer:
		a := make([]any, len(t))
		for i := range t {
			a[i] = t[i]
		}

		return Join(sep, a...).ToSQL()
	case []Expr:
		a := make([]any, len(t))
		for i := range t {
			a[i] = t[i]
		}

		return Join(sep, a...).ToSQL()
	default:
		return "?", []any{t}, nil
	}
}

func SQL(sql string, args ...any) Expression {
	return Expression{
		SQL:  sql,
		Args: args,
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

	b := &strings.Builder{}
	args := make([]any, 0, len(e.Args))

	i := -1

	for {
		index := strings.IndexRune(e.SQL, '?')
		if index < 0 {
			b.WriteString(e.SQL)
			break
		}

		if index < len(e.SQL)-1 && e.SQL[index+1] == '?' {
			b.WriteString(e.SQL[:index+2])
			e.SQL = e.SQL[index+2:]
			continue
		}

		i++

		if i >= len(e.Args) {
			return "", nil, ErrInvalidNumberOfArguments
		}

		b.WriteString(e.SQL[:index])
		e.SQL = e.SQL[index+1:]

		s, a, err := toSQL(e.Args[i], ", ")
		if err != nil {
			return "", nil, err
		}

		if s == "" {
			continue
		}

		b.WriteString(s)
		args = append(args, a...)
	}

	if i != len(e.Args)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return b.String(), args, nil
}

func ToPostgres(expr any) (string, []any, error) {
	s, a, err := toSQL(expr, ", ")
	if err != nil {
		return "", nil, err
	}

	b := &strings.Builder{}
	i := -1

	for {
		index := strings.IndexRune(s, '?')
		if index < 0 {
			b.WriteString(s)
			break
		}

		if index < len(s)-1 && s[index+1] == '?' {
			b.WriteString(s[:index+1])
			s = s[index+2:]
			continue
		}

		i++

		b.WriteString(fmt.Sprintf("%s$%d", s[:index], i+1))
		s = s[index+1:]
	}

	if i != len(a)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return b.String(), a, nil
}

func Join(sep string, expr ...any) Expression {
	if sep == "" {
		return Append(expr...)
	}

	b := &strings.Builder{}
	args := make([]any, 0, len(expr))

	i := 0

	for _, e := range expr {
		s, a, err := toSQL(e, sep)
		if err != nil {
			return Expression{Err: err}
		}

		if s == "" {
			continue
		}

		if i != 0 {
			b.WriteString(sep)
		}

		b.WriteString(s)
		args = append(args, a...)

		i++
	}

	return SQL(b.String(), args...)
}

func Append(expr ...any) Expression {
	b := &strings.Builder{}
	args := make([]any, 0, len(expr))

	for _, e := range expr {
		s, a, err := toSQL(e, ", ")
		if err != nil {
			return Expression{Err: err}
		}

		if s == "" {
			continue
		}

		b.WriteString(s)
		args = append(args, a...)
	}

	return SQL(b.String(), args...)
}

func If(condition bool, then any, els any) Expression {
	if condition {
		s, a, err := toSQL(then, ", ")
		if err != nil {
			return Expression{Err: err}
		}

		return SQL(s, a...)
	}

	s, a, err := toSQL(els, ", ")
	if err != nil {
		return Expression{Err: err}
	}

	return SQL(s, a...)
}

func Skip() Expression {
	return Expression{}
}
