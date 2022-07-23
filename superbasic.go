package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

func toSQL(expr any) (string, []any, error) {
	switch t := expr.(type) {
	case Expression:
		return t.ToSQL()
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

		i++

		if i >= len(e.Args) {
			return "", nil, ErrInvalidNumberOfArguments
		}

		b.WriteString(e.SQL[:index])
		e.SQL = e.SQL[index+1:]

		s, a, err := toSQL(e.Args[i])
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

func (e Expression) ToPostgres() (string, []any, error) {
	s, a, err := e.ToSQL()
	if err != nil {
		return "", nil, err
	}

	b := &strings.Builder{}
	i := 0

	for {
		index := strings.IndexRune(s, '?')
		if index < 0 {
			b.WriteString(s)
			break
		}

		i++

		b.WriteString(fmt.Sprintf("%s$%d", s[:index], i))
		s = s[index+1:]
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
		s, a, err := toSQL(e)
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
	return Expression{
		SQL:  strings.Repeat("?", len(expr)),
		Args: expr,
	}
}
