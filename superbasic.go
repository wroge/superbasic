//nolint:exhaustivestruct,exhaustruct,wrapcheck,funlen,cyclop
package superbasic

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

// Sqlizer returns a prepared statement.
type Sqlizer interface {
	ToSQL() (string, []any, error)
}

// SQL takes a template with placeholders into which expressions can be compiled.
// Expressions can be of type Sqlizer or []Sqlizer. []Sqlizer gets joined to an
// Expression with ", " as separator. All other values will be put into the arguments
// of a prepared statement.
// You can escape '?' by using '??'.
func SQL(template string, expressions ...any) Expression {
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

		if exprIndex >= len(expressions) {
			return Expression{Err: ErrInvalidNumberOfArguments}
		}

		builder.WriteString(template[:index])
		template = template[index+1:]

		var (
			sql  string
			args []any
			err  error
		)

		switch expr := expressions[exprIndex].(type) {
		case Sqlizer:
			sql, args, err = expr.ToSQL()
		case []Sqlizer:
			sql, args, err = Join(", ", expr...).ToSQL()
		case []Expression:
			sql, args, err = Join(", ", toSqlizerSlice(expr)...).ToSQL()
		default:
			sql = "?"
			args = []any{expr}
		}

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

// Append builds an Expression by appending Sqlizer's.
func Append(expr ...Sqlizer) Expression {
	return Join("", expr...)
}

// Join builds an Expression by joining Sqlizer's with a separator.
func Join(sep string, expr ...Sqlizer) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expr))

	isFirst := true

	for _, e := range expr {
		if e == nil {
			continue
		}

		sql, args, err := e.ToSQL()
		if err != nil {
			return Expression{Err: err}
		}

		if sql == "" {
			continue
		}

		if isFirst {
			isFirst = false

			builder.WriteString(sql)
		} else {
			builder.WriteString(sep + sql)
		}

		arguments = append(arguments, args...)
	}

	return Expression{SQL: builder.String(), Args: arguments}
}

// If returns then if condition is true.
// If the condition is false, an empty Expression is returned.
func If(condition bool, then Sqlizer) Expression {
	if condition {
		return ToExpression(then)
	}

	return Expression{}
}

// IfElse returns then Sqlizer on true and els on false as Expression.
func IfElse(condition bool, then, els Sqlizer) Expression {
	if condition {
		return ToExpression(then)
	}

	return ToExpression(els)
}

// ToExpression returns a valid Expression from a Sqlizer.
func ToExpression(expr Sqlizer) Expression {
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

// Columns in a Sqlizer that joins a list of identifiers with ", ".
type Columns []string

// ToSQL is the implementation of the Sqlizer interface.
func (c Columns) ToSQL() (string, []any, error) {
	return strings.Join(c, ", "), nil, nil
}

// Values is a Sqlizer that takes a list of values.
type Values [][]any

// ToSQL is the implementation of the Sqlizer interface.
func (v Values) ToSQL() (string, []any, error) {
	if len(v) == 0 {
		return "", nil, nil
	}

	placeholders := make([]string, len(v))

	var args []any

	for i, values := range v {
		placeholders[i] = "(" + strings.Repeat(", ?", len(values))[2:] + ")"

		if len(args) == 0 {
			args = make([]any, 0, len(v)*len(values))
		}

		args = append(args, values...)
	}

	return strings.Join(placeholders, ", "), args, nil
}

// ToPostgres transforms a Sqlizer to a valid postgres statement.
func ToPostgres(expr Sqlizer) (string, []any, error) {
	sql, args, err := expr.ToSQL()
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

func toSqlizerSlice(s []Expression) []Sqlizer {
	out := make([]Sqlizer, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}
