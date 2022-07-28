//nolint:exhaustivestruct,exhaustruct
package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidNumberOfArguments is returned if arguments doesn't match the number of placeholders.
var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

// ErrInvalidExpression is returned if expressions are nil.
var ErrInvalidExpression = errors.New("invalid expression")

// Sqlizer represents a prepared statement.
type Sqlizer interface {
	ToSQL() (string, []any, error)
}

// Compile takes a template with placeholders into which expressions can be compiled.
// You can escape '?' by using '??'.
func Compile(template string, expressions ...Sqlizer) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expressions))

	exprIndex := -1

	for {
		index := strings.IndexRune(template, '?')
		if index < 0 {
			builder.WriteString(template)

			break
		}

		if index < len(template)-1 && template[index+1] == '?' {
			builder.WriteString(template[:index+2])
			template = template[index+2:]

			continue
		}

		exprIndex++

		if expressions[exprIndex] == nil {
			return Expression{Err: ErrInvalidExpression}
		}

		if exprIndex >= len(expressions) {
			return Expression{Err: ErrInvalidNumberOfArguments}
		}

		builder.WriteString(template[:index])
		template = template[index+1:]

		sql, args, err := expressions[exprIndex].ToSQL()
		if err != nil {
			return Expression{Err: err}
		}

		if sql == "" {
			continue
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex != len(expressions)-1 {
		return Expression{Err: ErrInvalidNumberOfArguments}
	}

	return Expression{SQL: builder.String(), Args: arguments}
}

// SQL returns an expression.
func SQL(sql string, args ...any) Expression {
	return Expression{SQL: sql, Args: args}
}

// Append builds an Expression by appending Sqlizer's.
func Append(expressions ...Sqlizer) Expression {
	return Join("", expressions...)
}

// Join builds an Expression by joining Sqlizer's with a separator.
func Join(sep string, expressions ...Sqlizer) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expressions))

	isFirst := true

	for _, expr := range expressions {
		if expr == nil {
			return Expression{Err: ErrInvalidExpression}
		}

		sql, args, err := expr.ToSQL()
		if err != nil {
			return Expression{Err: err}
		}

		if sql == "" {
			continue
		}

		if isFirst {
			builder.WriteString(sql)

			isFirst = false
		} else {
			builder.WriteString(sep + sql)
		}

		arguments = append(arguments, args...)
	}

	return Expression{SQL: builder.String(), Args: arguments}
}

// If returns an expression based on a condition.
// If false an empty expression is returned.
func If(condition bool, then Sqlizer) Expression {
	if condition {
		return toExpression(then)
	}

	return Expression{}
}

// IfElse returns an expression based on a condition.
func IfElse(condition bool, then, els Sqlizer) Expression {
	if condition {
		return toExpression(then)
	}

	return toExpression(els)
}

// Idents joins idents with ", " to an expression.
func Idents(idents ...string) Expression {
	return Expression{SQL: strings.Join(idents, ", ")}
}

// Value returns an expression with a placeholder.
func Value(a any) Expression {
	return Expression{SQL: "?", Args: []any{a}}
}

// Values returns an expression with placeholders.
func Values(a ...any) Expression {
	return Expression{SQL: fmt.Sprintf("(%s)", strings.Repeat(", ?", len(a))[2:]), Args: a}
}

// Equals returns an expression with an '=' sign.
func Equals(ident string, value any) Expression {
	return Expression{SQL: ident + " = ?", Args: []any{value}}
}

// NotEquals returns an expression with an '<>' sign.
func NotEquals(ident string, value any) Expression {
	return Expression{SQL: ident + " <> ?", Args: []any{value}}
}

// Greater returns an expression with an '>' sign.
func Greater(ident string, value any) Expression {
	return Expression{SQL: ident + " > ?", Args: []any{value}}
}

// GreaterOrEquals returns an expression with an '>=' sign.
func GreaterOrEquals(ident string, value any) Expression {
	return Expression{SQL: ident + " >= ?", Args: []any{value}}
}

// Less returns an expression with an '<' sign.
func Less(ident string, value any) Expression {
	return Expression{SQL: ident + " < ?", Args: []any{value}}
}

// LessOrEquals returns an expression with an '<=' sign.
func LessOrEquals(ident string, value any) Expression {
	return Expression{SQL: ident + " <= ?", Args: []any{value}}
}

// In returns an expression with an 'IN' sign.
func In(ident string, values ...any) Expression {
	return Expression{SQL: fmt.Sprintf("%s IN (%s)", ident, strings.Repeat(", ?", len(values))[2:]), Args: values}
}

func toExpression(expr Sqlizer) Expression {
	if expr == nil {
		return Expression{}
	}

	sql, args, err := expr.ToSQL()
	if err != nil {
		return Expression{Err: err}
	}

	return Expression{SQL: sql, Args: args}
}

// Expression represents a prepared statement.
type Expression struct {
	SQL  string
	Args []any
	Err  error
}

// ToSQL is the implementation of the Sqlizer interface.
func (e Expression) ToSQL() (string, []any, error) {
	if e.Err != nil {
		return "", nil, e.Err
	}

	return e.SQL, e.Args, e.Err
}

func (e Expression) ToPositional(placeholder string) (string, []any, error) {
	build := &strings.Builder{}
	argIndex := -1

	for {
		index := strings.IndexRune(e.SQL, '?')
		if index < 0 {
			build.WriteString(e.SQL)

			break
		}

		if index < len(e.SQL)-1 && e.SQL[index+1] == '?' {
			build.WriteString(e.SQL[:index+1])
			e.SQL = e.SQL[index+2:]

			continue
		}

		argIndex++

		build.WriteString(fmt.Sprintf("%s%s%d", e.SQL[:index], placeholder, argIndex+1))
		e.SQL = e.SQL[index+1:]
	}

	if argIndex != len(e.Args)-1 {
		return "", nil, ErrInvalidNumberOfArguments
	}

	return build.String(), e.Args, nil
}
