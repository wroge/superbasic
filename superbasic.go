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
		return Join(", ", anySlice(expr)...).ToSQL()
	case []any:
		return fmt.Sprintf("(%s)", strings.Repeat(", ?", len(expr))[2:]), expr, nil
	default:
		return "?", []any{expr}, nil
	}
}

// SQL takes a template with placeholders into which expressions can be compiled.
// []Expression is compiled to Join(sep, expr...). (default sep is ", ")
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
	return SQL("? BETWEEN ? AND ?", expr, lower, higher)
}

// BetweenIdent returns a BETWEEN expression.
func BetweenIdent(ident string, lower, higher any) Expression {
	return Between(SQL(ident), lower, higher)
}

// NotBetween returns a NOT BETWEEN expression.
func NotBetween(expr, lower, higher any) Expression {
	return SQL("? NOT BETWEEN ? AND ?", expr, lower, higher)
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

func SelectSQL(sql string, args ...any) *SelectBuilder {
	return Select(SQL(sql, args...))
}

func Select(expr any) *SelectBuilder {
	return &SelectBuilder{
		sel: expr,
	}
}

type SelectBuilder struct {
	sel     any
	from    any
	where   any
	groupBy any
	having  any
	orderBy any
	limit   uint64
	offset  uint64
}

func (sb *SelectBuilder) ToSQL() (string, []any, error) {
	return Append(
		IfElse(sb.sel != nil, SQL("SELECT ?", sb.sel), SQL("SELECT *")),
		If(sb.from != nil, SQL(" FROM ?", sb.from)),
		If(sb.where != nil, SQL(" WHERE ?", sb.where)),
		If(sb.groupBy != nil, SQL(" GROUP BY ?", sb.groupBy)),
		If(sb.having != nil, SQL(" HAVING ?", sb.having)),
		If(sb.orderBy != nil, SQL(" ORDER BY ?", sb.orderBy)),
		If(sb.limit > 0, SQL(fmt.Sprintf(" LIMIT %d", sb.limit))),
		If(sb.offset > 0, SQL(fmt.Sprintf(" OFFSET %d", sb.offset))),
	).ToSQL()
}

func (sb *SelectBuilder) FromSQL(sql string, args ...any) *SelectBuilder {
	return sb.From(SQL(sql, args...))
}

func (sb *SelectBuilder) From(expr any) *SelectBuilder {
	sb.from = expr

	return sb
}

func (sb *SelectBuilder) WhereSQL(sql string, args ...any) *SelectBuilder {
	return sb.Where(SQL(sql, args...))
}

func (sb *SelectBuilder) Where(expr any) *SelectBuilder {
	sb.where = expr

	return sb
}

func (sb *SelectBuilder) GroupBySQL(sql string, args ...any) *SelectBuilder {
	return sb.GroupBy(SQL(sql, args...))
}

func (sb *SelectBuilder) GroupBy(expr any) *SelectBuilder {
	sb.groupBy = expr

	return sb
}

func (sb *SelectBuilder) HavingSQL(sql string, args ...any) *SelectBuilder {
	return sb.Having(SQL(sql, args...))
}

func (sb *SelectBuilder) Having(expr any) *SelectBuilder {
	sb.having = expr

	return sb
}

func (sb *SelectBuilder) OrderBySQL(sql string, args ...any) *SelectBuilder {
	return sb.OrderBy(SQL(sql, args...))
}

func (sb *SelectBuilder) OrderBy(expr any) *SelectBuilder {
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
	return Append(
		SQL(fmt.Sprintf("INSERT INTO %s", ib.into)),
		If(len(ib.columns) > 0, SQL(fmt.Sprintf(" (%s)", strings.Join(ib.columns, ", ")))),
		SQL(" VALUES ?", ib.data),
	).ToSQL()
}

func (ib *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	ib.columns = columns

	return ib
}

func (ib *InsertBuilder) AddRow(values ...any) *InsertBuilder {
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
	sets  []any
	where any
}

func (ub *UpdateBuilder) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("UPDATE %s SET ?", ub.table), Join(", ", ub.sets...)),
		If(ub.where != nil, SQL(" WHERE ?", ub.where)),
	).ToSQL()
}

func (ub *UpdateBuilder) AddSetSQL(sql string, args ...any) *UpdateBuilder {
	return ub.AddSet(SQL(sql, args...))
}

func (ub *UpdateBuilder) AddSet(set any) *UpdateBuilder {
	ub.sets = append(ub.sets, set)

	return ub
}

func (ub *UpdateBuilder) WhereSQL(sql string, args ...any) *UpdateBuilder {
	return ub.Where(SQL(sql, args...))
}

func (ub *UpdateBuilder) Where(expr any) *UpdateBuilder {
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
	where any
}

func (db *DeleteBuilder) ToSQL() (string, []any, error) {
	return Append(
		SQL(fmt.Sprintf("DELETE FROM %s", db.from)),
		If(db.where != nil, SQL(" WHERE ?", db.where)),
	).ToSQL()
}

func (db *DeleteBuilder) WhereSQL(sql string, args ...any) *DeleteBuilder {
	return db.Where(SQL(sql, args...))
}

func (db *DeleteBuilder) Where(expr any) *DeleteBuilder {
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

func anySlice[T any](s []T) []any {
	out := make([]any, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}
