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

func Select(sql string, args ...any) *SelectBuilder {
	return SelectExpr(SQL(sql, args...))
}

func SelectExpr(expr Expression) *SelectBuilder {
	return &SelectBuilder{
		sel: expr,
	}
}

type SelectBuilder struct {
	sel     Expression
	from    Expression
	where   Expression
	groupBy Expression
	having  Expression
	orderBy Expression
	limit   uint64
	offset  uint64
}

func (sb *SelectBuilder) ToSQL() (string, []any, error) {
	return Join(" ",
		IfElse(sb.sel != nil, SQL("SELECT ?", sb.sel), SQL("SELECT *")),
		If(sb.from != nil, SQL("FROM ?", sb.from)),
		If(sb.where != nil, SQL("WHERE ?", sb.where)),
		If(sb.groupBy != nil, SQL("GROUP BY ?", sb.groupBy)),
		If(sb.having != nil, SQL("HAVING ?", sb.having)),
		If(sb.orderBy != nil, SQL("ORDER BY ?", sb.orderBy)),
		If(sb.limit > 0, SQL(fmt.Sprintf("LIMIT %d", sb.limit))),
		If(sb.offset > 0, SQL(fmt.Sprintf("OFFSET %d", sb.offset))),
	).ToSQL()
}

func (sb *SelectBuilder) Select(sql string, args ...any) *SelectBuilder {
	return sb.SelectExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) SelectExpr(expr Expression) *SelectBuilder {
	sb.sel = expr

	return sb
}

func (sb *SelectBuilder) From(sql string, args ...any) *SelectBuilder {
	return sb.FromExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) FromExpr(expr Expression) *SelectBuilder {
	sb.from = expr

	return sb
}

func (sb *SelectBuilder) Where(sql string, args ...any) *SelectBuilder {
	return sb.WhereExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) WhereExpr(expr Expression) *SelectBuilder {
	sb.where = expr

	return sb
}

func (sb *SelectBuilder) GroupBy(sql string, args ...any) *SelectBuilder {
	return sb.GroupByExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) GroupByExpr(expr Expression) *SelectBuilder {
	sb.groupBy = expr

	return sb
}

func (sb *SelectBuilder) Having(sql string, args ...any) *SelectBuilder {
	return sb.HavingExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) HavingExpr(expr Expression) *SelectBuilder {
	sb.having = expr

	return sb
}

func (sb *SelectBuilder) OrderBy(sql string, args ...any) *SelectBuilder {
	return sb.OrderByExpr(SQL(sql, args...))
}

func (sb *SelectBuilder) OrderByExpr(expr Expression) *SelectBuilder {
	sb.orderBy = expr

	return sb
}

func (sb *SelectBuilder) Limit(limit uint64) *SelectBuilder {
	sb.limit = limit

	return sb
}

func (sb *SelectBuilder) Offset(offset uint64) *SelectBuilder {
	sb.offset = offset

	return sb
}

func Insert(into string) *InsertBuilder {
	return &InsertBuilder{
		into: into,
	}
}

type InsertBuilder struct {
	into    string
	columns []string
	data    [][]any
}

func (ib *InsertBuilder) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("INSERT INTO %s", ib.into)),
		If(len(ib.columns) > 0, SQL(fmt.Sprintf("(%s)", strings.Join(ib.columns, ", ")))),
		SQL("VALUES ?", ib.data),
	).ToSQL()
}

func (ib *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	ib.columns = columns

	return ib
}

func (ib *InsertBuilder) Values(values ...any) *InsertBuilder {
	ib.data = append(ib.data, values)

	return ib
}

func Update(table string) *UpdateBuilder {
	return &UpdateBuilder{
		table: table,
	}
}

type UpdateBuilder struct {
	table string
	sets  []Expression
	where Expression
}

func (ub *UpdateBuilder) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("UPDATE %s SET ?", ub.table), Join(", ", ub.sets...)),
		If(ub.where != nil, SQL("WHERE ?", ub.where)),
	).ToSQL()
}

func (ub *UpdateBuilder) Set(sql string, args ...any) *UpdateBuilder {
	return ub.SetExpr(SQL(sql, args...))
}

func (ub *UpdateBuilder) SetExpr(set Expression) *UpdateBuilder {
	ub.sets = append(ub.sets, set)

	return ub
}

func (ub *UpdateBuilder) Where(sql string, args ...any) *UpdateBuilder {
	return ub.WhereExpr(SQL(sql, args...))
}

func (ub *UpdateBuilder) WhereExpr(expr Expression) *UpdateBuilder {
	ub.where = expr

	return ub
}

func Delete(from string) *DeleteBuilder {
	return &DeleteBuilder{
		from: from,
	}
}

type DeleteBuilder struct {
	from  string
	where Expression
}

func (db *DeleteBuilder) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("DELETE FROM %s", db.from)),
		If(db.where != nil, SQL("WHERE ?", db.where)),
	).ToSQL()
}

func (db *DeleteBuilder) Where(sql string, args ...any) *DeleteBuilder {
	return db.WhereExpr(SQL(sql, args...))
}

func (db *DeleteBuilder) WhereExpr(expr Expression) *DeleteBuilder {
	db.where = expr

	return db
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
