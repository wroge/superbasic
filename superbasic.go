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

func compile(sep string, expression any) (string, []any, error) {
	switch expr := expression.(type) {
	case Expression:
		return expr.ToSQL()
	case []Expression:
		return Join(sep, anySlice(expr)...).ToSQL()
	case [][]any:
		return Join(sep, anySlice(expr)...).ToSQL()
	case []any:
		return fmt.Sprintf("(%s)", strings.Repeat(", ?", len(expr))[2:]), expr, nil
	default:
		return "?", []any{expr}, nil
	}
}

// SQL takes a template with placeholders into which expressions can be compiled.
// Expression []Expression is compiled to Join(", ", expr...).
// Expression []any is compiled to (?, ?).
// Expression [][]any is compiled to (?, ?), (?, ?).
// Escape '?' by using '??'.
func SQL(sql string, expressions ...any) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expressions))

	exprIndex := -1

	for {
		index := strings.IndexRune(sql, '?')
		if index < 0 {
			builder.WriteString(sql)

			break
		}

		if index < len(sql)-1 && sql[index+1] == '?' {
			builder.WriteString(sql[:index+2])
			sql = sql[index+2:]

			continue
		}

		exprIndex++

		if expressions[exprIndex] == nil {
			return ExpressionError{}
		}

		if exprIndex >= len(expressions) {
			return NumberOfArgumentsError{}
		}

		builder.WriteString(sql[:index])
		sql = sql[index+1:]

		sql, args, err := compile(", ", expressions[exprIndex])
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

// Append expressions.
func Append(expressions ...any) Expression {
	return Join("", expressions...)
}

// Join joins expressions by a separator.
func Join(sep string, expressions ...any) Expression {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(expressions))

	isFirst := true

	for _, expr := range expressions {
		if expr == nil {
			return ExpressionError{}
		}

		sql, args, err := compile(sep, expr)
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
func If(condition bool, then any) Expression {
	if condition {
		return SQL("?", then)
	}

	return expression{}
}

// IfElse returns an expression based on a condition.
func IfElse(condition bool, then, els any) Expression {
	if condition {
		return SQL("?", then)
	}

	return SQL("?", els)
}

// And returns a AND expression.
func And(expr ...any) Expression {
	return Join(" AND ", expr...)
}

// Or returns a OR expression.
func Or(left, right any) Expression {
	return SQL("(? OR ?)", left, right)
}

// Not returns a NOT expression.
func Not(expr any) Expression {
	return SQL("NOT (?)", expr)
}

// Equals returns an expression with an '=' sign.
func Equals(left, right any) Expression {
	return SQL("? = ?", left, right)
}

// EqualsIdent returns an expression with an '=' sign.
func EqualsIdent(ident string, value any) Expression {
	return Equals(SQL(ident), value)
}

// NotEquals returns an expression with an '<>' sign.
func NotEquals(left, right any) Expression {
	return SQL("? <> ?", left, right)
}

// NotEqualsIdent returns an expression with an '<>' sign.
func NotEqualsIdent(ident string, value any) Expression {
	return NotEquals(SQL(ident), value)
}

// Greater returns an expression with an '>' sign.
func Greater(left, right any) Expression {
	return SQL("? > ?", left, right)
}

// GreaterIdent returns an expression with an '>' sign.
func GreaterIdent(ident string, value any) Expression {
	return Greater(SQL(ident), value)
}

// GreaterOrEquals returns an expression with an '>=' sign.
func GreaterOrEquals(left, right any) Expression {
	return SQL("? >= ?", left, right)
}

// GreaterOrEqualsIdent returns an expression with an '>=' sign.
func GreaterOrEqualsIdent(ident string, value any) Expression {
	return GreaterOrEquals(SQL(ident), value)
}

// Less returns an expression with an '<' sign.
func Less(left, right any) Expression {
	return SQL("? < ?", left, right)
}

// LessIdent returns an expression with an '<' sign.
func LessIdent(ident string, value any) Expression {
	return Less(SQL(ident), value)
}

// LessOrEquals returns an expression with an '<=' sign.
func LessOrEquals(left, right any) Expression {
	return SQL("? <= ?", left, right)
}

// LessOrEqualsIdent returns an expression with an '<=' sign.
func LessOrEqualsIdent(ident string, value any) Expression {
	return LessOrEquals(SQL(ident), value)
}

// In returns a IN expression.
func In(left, right any) Expression {
	return SQL("? IN ?", left, right)
}

// InIdent returns a IN expression.
func InIdent(ident string, value any) Expression {
	return In(SQL(ident), value)
}

// NotIn returns a NOT IN expression.
func NotIn(left, right any) Expression {
	return SQL("? NOT IN ?", left, right)
}

// NotInIdent returns a NOT IN expression.
func NotInIdent(ident string, value any) Expression {
	return NotIn(SQL(ident), value)
}

// IsNull returns a IS NULL expression.
func IsNull(expr any) Expression {
	return SQL("? IS NULL", expr)
}

// IsNullIdent returns a IS NULL expression.
func IsNullIdent(ident string) Expression {
	return IsNull(SQL(ident))
}

// IsNotNull returns a IS NOT NULL expression.
func IsNotNull(expr any) Expression {
	return SQL("? IS NOT NULL", expr)
}

// IsNotNullIdent returns a IS NOT NULL expression.
func IsNotNullIdent(ident string) Expression {
	return IsNotNull(SQL(ident))
}

// Between returns a BETWEEN expression.
func Between(expr, lower, higher any) Expression {
	return SQL("? BETWEEN ? ANY ?", expr, lower, higher)
}

// BetweenIdent returns a BETWEEN expression.
func BetweenIdent(ident string, lower, higher any) Expression {
	return Between(SQL(ident), lower, higher)
}

// NotBetween returns a NOT BETWEEN expression.
func NotBetween(expr, lower, higher any) Expression {
	return SQL("? NOT BETWEEN ? ANY ?", expr, lower, higher)
}

// NotBetweenIdent returns a NOT BETWEEN expression.
func NotBetweenIdent(ident string, lower, higher any) Expression {
	return NotBetween(SQL(ident), lower, higher)
}

// Like returns a LIKE expression.
func Like(left, right any) Expression {
	return SQL("? LIKE ?", left, right)
}

// LikeIdent returns a LIKE expression.
func LikeIdent(ident string, value any) Expression {
	return Like(SQL(ident), value)
}

// NotLike returns a NOT LIKE expression.
func NotLike(left, right any) Expression {
	return SQL("? NOT LIKE ?", left, right)
}

// NotLikeIdent returns a LIKE expression.
func NotLikeIdent(ident string, value any) Expression {
	return NotLike(SQL(ident), value)
}

// Cast returns a CAST expression.
func Cast(expr any, as string) Expression {
	return SQL(fmt.Sprintf("CAST(? AS %s)", as), expr)
}

// CastIdent returns a CAST expression.
func CastIdent(ident string, as string) Expression {
	return Cast(SQL(ident), as)
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

func anySlice[T any](s []T) []any {
	out := make([]any, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}
