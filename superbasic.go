//nolint:exhaustivestruct,exhaustruct,wrapcheck,ireturn
package superbasic

import (
	"fmt"
	"strings"
)

// NumberOfArgumentsError is returned if arguments doesn't match the number of placeholders.
type NumberOfArgumentsError struct{}

func (e NumberOfArgumentsError) Error() string {
	return "invalid number of arguments"
}

func (e NumberOfArgumentsError) ToSQL() (string, []any, error) {
	return "", nil, e
}

// ExpressionError is returned if expressions are nil.
type ExpressionError struct{}

func (e ExpressionError) Error() string {
	return "invalid expression"
}

func (e ExpressionError) ToSQL() (string, []any, error) {
	return "", nil, e
}

// Expression represents a prepared statement.
type Expression interface {
	ToSQL() (string, []any, error)
}

// Compile takes a template with placeholders into which expressions can be compiled.
// You can escape '?' by using '??'.
func Compile(template string, expressions ...Expression) Expression {
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
			return ExpressionError{}
		}

		if exprIndex >= len(expressions) {
			return NumberOfArgumentsError{}
		}

		builder.WriteString(template[:index])
		template = template[index+1:]

		sql, args, err := expressions[exprIndex].ToSQL()
		if err != nil {
			return expression{err: err}
		}

		if sql == "" {
			continue
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex != len(expressions)-1 {
		return NumberOfArgumentsError{}
	}

	return expression{sql: builder.String(), args: arguments}
}

// SQL returns an expression.
func SQL(sql string, args ...any) Expression {
	return expression{sql: sql, args: args}
}

// Append expressions.
func Append(expressions ...Expression) Expression {
	return Join("", expressions...)
}

// Join joins expressions by a separator.
func Join(sep string, expressions ...Expression) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expressions))

	isFirst := true

	for _, expr := range expressions {
		if expr == nil {
			return ExpressionError{}
		}

		sql, args, err := expr.ToSQL()
		if err != nil {
			return expression{err: err}
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

	return expression{sql: builder.String(), args: arguments}
}

// If returns an expression based on a condition.
// If false an empty expression is returned.
func If(condition bool, then Expression) Expression {
	if condition {
		return then
	}

	return expression{}
}

// IfElse returns an expression based on a condition.
func IfElse(condition bool, then, els Expression) Expression {
	if condition {
		return then
	}

	return els
}

// Idents joins idents with ", " to an expression.
func Idents(idents ...string) Expression {
	return expression{sql: strings.Join(idents, ", ")}
}

// Value returns an expression with a placeholder.
func Value(a any) Expression {
	return expression{sql: "?", args: []any{a}}
}

// Values returns an expression with placeholders.
func Values(a ...any) Expression {
	return expression{sql: fmt.Sprintf("(%s)", strings.Repeat(", ?", len(a))[2:]), args: a}
}

// And returns a AND expression.
func And(expr ...Expression) Expression {
	return Join(" AND ", expr...)
}

// Or returns a OR expression.
func Or(left, right Expression) Expression {
	return Compile("(? OR ?)", left, right)
}

// Not returns a NOT expression.
func Not(expr Expression) Expression {
	return Compile("NOT (?)", expr)
}

// Equals returns an expression with an '=' sign.
func Equals(ident string, value any) Expression {
	return expression{sql: ident + " = ?", args: []any{value}}
}

// NotEquals returns an expression with an '<>' sign.
func NotEquals(ident string, value any) Expression {
	return expression{sql: ident + " <> ?", args: []any{value}}
}

// Greater returns an expression with an '>' sign.
func Greater(ident string, value any) Expression {
	return expression{sql: ident + " > ?", args: []any{value}}
}

// GreaterOrEquals returns an expression with an '>=' sign.
func GreaterOrEquals(ident string, value any) Expression {
	return expression{sql: ident + " >= ?", args: []any{value}}
}

// Less returns an expression with an '<' sign.
func Less(ident string, value any) Expression {
	return expression{sql: ident + " < ?", args: []any{value}}
}

// LessOrEquals returns an expression with an '<=' sign.
func LessOrEquals(ident string, value any) Expression {
	return expression{sql: ident + " <= ?", args: []any{value}}
}

// In returns a IN expression.
func In(ident string, values ...any) Expression {
	return expression{sql: fmt.Sprintf("%s IN (%s)", ident, strings.Repeat(", ?", len(values))[2:]), args: values}
}

// NotIn returns a NOT IN expression.
func NotIn(ident string, values ...any) Expression {
	return expression{sql: fmt.Sprintf("%s NOT IN (%s)", ident, strings.Repeat(", ?", len(values))[2:]), args: values}
}

// IsNull returns a IS NULL expression.
func IsNull(ident string) Expression {
	return expression{sql: fmt.Sprintf("%s IS NULL", ident)}
}

// IsNotNull returns a IS NOT NULL expression.
func IsNotNull(ident string) Expression {
	return expression{sql: fmt.Sprintf("%s IS NOT NULL", ident)}
}

// Between returns a BETWEEN expression.
func Between(ident string, lower, higher any) Expression {
	return expression{sql: fmt.Sprintf("%s BETWEEN ? AND ?", ident), args: []any{lower, higher}}
}

// NotBetween returns a NOT BETWEEN expression.
func NotBetween(ident string, lower, higher any) Expression {
	return expression{sql: fmt.Sprintf("%s NOT BETWEEN ? AND ?", ident), args: []any{lower, higher}}
}

// Like returns a LIKE expression.
func Like(ident string, value any) Expression {
	return expression{sql: fmt.Sprintf("%s LIKE ?", ident), args: []any{value}}
}

// NotLike returns a NOT LIKE expression.
func NotLike(ident string, value any) Expression {
	return expression{sql: fmt.Sprintf("%s NOT LIKE ?", ident), args: []any{value}}
}

// Cast returns a CAST expression.
func Cast(value any, as string) Expression {
	return expression{sql: fmt.Sprintf("CAST(? AS %s)", as), args: []any{value}}
}

type expression struct {
	sql  string
	args []any
	err  error
}

func (e expression) ToSQL() (string, []any, error) {
	if e.err != nil {
		return "", nil, e.err
	}

	return e.sql, e.args, e.err
}

func ToPositional(placeholder string, expr Expression) (string, []any, error) {
	if expr == nil {
		return "", nil, ExpressionError{}
	}

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

		build.WriteString(fmt.Sprintf("%s%s%d", sql[:index], placeholder, argIndex+1))
		sql = sql[index+1:]
	}

	if argIndex != len(args)-1 {
		return "", nil, NumberOfArgumentsError{}
	}

	return build.String(), args, nil
}
