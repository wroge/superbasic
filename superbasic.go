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

type Expression interface {
	ToSQL() (string, []any, error)
}

func compile(sep string, expression any) (string, []any, error) {
	switch expr := expression.(type) {
	case Expression:
		return expr.ToSQL()
	case []Expression:
		return Join(sep, expr...).ToSQL()
	case [][]any:
		arr := make([]Expression, len(expr))

		for i, d := range expr {
			arr[i] = Values(d...)
		}

		return Join(", ", arr...).ToSQL()
	case []any:
		return Values(expr...).ToSQL()
	default:
		return "?", []any{expr}, nil
	}
}

func Values(values ...any) Expression {
	return expression{sql: fmt.Sprintf("(%s)", strings.Repeat(", ?", len(values))[2:]), args: values}
}

// SQL takes a template with placeholders into which expressions can be compiled.
// []Expression is compiled to Join(", ", expr...).
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

		if exprIndex >= len(expressions) {
			return NumberOfArgumentsError{}
		}

		if expressions[exprIndex] == nil {
			return ExpressionError{}
		}

		builder.WriteString(sql[:index])
		sql = sql[index+1:]

		sql, args, err := compile(", ", expressions[exprIndex])
		if err != nil {
			return expression{err: err}
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex != len(expressions)-1 {
		return NumberOfArgumentsError{}
	}

	return expression{sql: builder.String(), args: arguments}
}

func Append(expressions ...Expression) Expression {
	return Join("", expressions...)
}

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

func If(condition bool, then any) Expression {
	if condition {
		return SQL("?", then)
	}

	return expression{}
}

func IfElse(condition bool, then, els any) Expression {
	if condition {
		return SQL("?", then)
	}

	return SQL("?", els)
}

func And(expr ...Expression) Expression {
	return Join(" AND ", expr...)
}

func Or(left, right Expression) Expression {
	return SQL("(? OR ?)", left, right)
}

func Not(expr Expression) Expression {
	return SQL("NOT (?)", expr)
}

func Equals(left, right any) Expression {
	return SQL("? = ?", left, right)
}

func EqualsIdent(ident string, value any) Expression {
	return Equals(SQL(ident), value)
}

func NotEquals(left, right any) Expression {
	return SQL("? <> ?", left, right)
}

func NotEqualsIdent(ident string, value any) Expression {
	return NotEquals(SQL(ident), value)
}

func Greater(left, right any) Expression {
	return SQL("? > ?", left, right)
}

func GreaterIdent(ident string, value any) Expression {
	return Greater(SQL(ident), value)
}

func GreaterOrEquals(left, right any) Expression {
	return SQL("? >= ?", left, right)
}

func GreaterOrEqualsIdent(ident string, value any) Expression {
	return GreaterOrEquals(SQL(ident), value)
}

func Less(left, right any) Expression {
	return SQL("? < ?", left, right)
}

func LessIdent(ident string, value any) Expression {
	return Less(SQL(ident), value)
}

func LessOrEquals(left, right any) Expression {
	return SQL("? <= ?", left, right)
}

func LessOrEqualsIdent(ident string, value any) Expression {
	return LessOrEquals(SQL(ident), value)
}

func In(left, right any) Expression {
	return SQL("? IN ?", left, right)
}

func InIdent(ident string, value any) Expression {
	return In(SQL(ident), value)
}

func NotIn(left, right any) Expression {
	return SQL("? NOT IN ?", left, right)
}

func NotInIdent(ident string, value any) Expression {
	return NotIn(SQL(ident), value)
}

func IsNull(expr any) Expression {
	return SQL("? IS NULL", expr)
}

func IsNullIdent(ident string) Expression {
	return IsNull(SQL(ident))
}

func IsNotNull(expr any) Expression {
	return SQL("? IS NOT NULL", expr)
}

func IsNotNullIdent(ident string) Expression {
	return IsNotNull(SQL(ident))
}

func Between(expr, lower, higher any) Expression {
	return SQL("? BETWEEN ? AND ?", expr, lower, higher)
}

func BetweenIdent(ident string, lower, higher any) Expression {
	return Between(SQL(ident), lower, higher)
}

func NotBetween(expr, lower, higher any) Expression {
	return SQL("? NOT BETWEEN ? AND ?", expr, lower, higher)
}

func NotBetweenIdent(ident string, lower, higher any) Expression {
	return NotBetween(SQL(ident), lower, higher)
}

func Like(left, right any) Expression {
	return SQL("? LIKE ?", left, right)
}

func LikeIdent(ident string, value any) Expression {
	return Like(SQL(ident), value)
}

func NotLike(left, right any) Expression {
	return SQL("? NOT LIKE ?", left, right)
}

func NotLikeIdent(ident string, value any) Expression {
	return NotLike(SQL(ident), value)
}

func Cast(expr any, as string) Expression {
	return SQL(fmt.Sprintf("CAST(? AS %s)", as), expr)
}

func CastIdent(ident string, as string) Expression {
	return Cast(SQL(ident), as)
}

type Query struct {
	With    Expression
	Select  Expression
	From    Expression
	Where   Expression
	GroupBy Expression
	Having  Expression
	Window  Expression
	OrderBy Expression
	Limit   uint64
	Offset  uint64
}

func (q Query) ToSQL() (string, []any, error) {
	return Join(" ",
		If(q.With != nil, SQL("WITH ?", q.With)),
		SQL("SELECT"),
		IfElse(q.Select != nil, q.Select, SQL("*")),
		If(q.From != nil, SQL("FROM ?", q.From)),
		If(q.Where != nil, SQL("WHERE ?", q.Where)),
		If(q.GroupBy != nil, SQL("GROUP BY ?", q.GroupBy)),
		If(q.Having != nil, SQL("HAVING ?", q.Having)),
		If(q.Having != nil, SQL("WINDOW ?", q.Window)),
		If(q.OrderBy != nil, SQL("ORDER BY ?", q.OrderBy)),
		If(q.Limit > 0, SQL(fmt.Sprintf("LIMIT %d", q.Limit))),
		If(q.Offset > 0, SQL(fmt.Sprintf("OFFSET %d", q.Offset))),
	).ToSQL()
}

type Insert struct {
	Into    string
	Columns []string
	Data    [][]any
}

func (i Insert) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("INSERT INTO %s", i.Into)),
		If(len(i.Columns) > 0, SQL(fmt.Sprintf("(%s)", strings.Join(i.Columns, ", ")))),
		SQL("VALUES ?", i.Data),
	).ToSQL()
}

type Update struct {
	Table string
	Sets  []Expression
	Where Expression
}

func (u Update) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("UPDATE %s SET ?", u.Table), Join(", ", u.Sets...)),
		If(u.Where != nil, SQL("WHERE ?", u.Where)),
	).ToSQL()
}

type Delete struct {
	From  string
	Where Expression
}

func (d Delete) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("DELETE FROM %s", d.From)),
		If(d.Where != nil, SQL("WHERE ?", d.Where)),
	).ToSQL()
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
