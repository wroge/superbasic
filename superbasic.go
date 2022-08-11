//nolint:exhaustivestruct,exhaustruct,ireturn
package superbasic

import (
	"fmt"
	"strings"
)

// NumberOfArgumentsError is returned if arguments doesn't match the number of placeholders.
type NumberOfArgumentsError struct {
	Got, Want int
}

func (e NumberOfArgumentsError) Error() string {
	return fmt.Sprintf("invalid number of arguments: got '%d' want '%d'", e.Got, e.Want)
}

// ExpressionError is returned if expressions are nil.
type ExpressionError struct {
	Err error
}

func (e ExpressionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid expression: %s", e.Err.Error())
	}

	return "invalid expression: expression is nil"
}

func (e ExpressionError) Unwrap() error {
	return e.Err
}

type ExpressionIndexError struct {
	Index int
}

func (e ExpressionIndexError) Error() string {
	return fmt.Sprintf("expression is nil at index '%d'", e.Index)
}

type Expression interface {
	ToSQL() (string, []any, error)
}

type Values []any

func (v Values) ToSQL() (string, []any, error) {
	return fmt.Sprintf("(%s)", strings.Repeat(", ?", len(v))[2:]), v, nil
}

// Compile takes a template with placeholders into which expressions can be compiled.
// Escape '?' by using '??'.
func Compile(template string, expressions ...Expression) Compiler {
	return Compiler{Template: template, Expressions: expressions}
}

type Compiler struct {
	Template    string
	Expressions []Expression
}

func (c Compiler) ToSQL() (string, []any, error) {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(c.Expressions))

	exprIndex := -1

	for {
		index := strings.IndexRune(c.Template, '?')
		if index < 0 {
			builder.WriteString(c.Template)

			break
		}

		if index < len(c.Template)-1 && c.Template[index+1] == '?' {
			builder.WriteString(c.Template[:index+2])
			c.Template = c.Template[index+2:]

			continue
		}

		exprIndex++

		if exprIndex >= len(c.Expressions) {
			return "", nil, NumberOfArgumentsError{Got: exprIndex, Want: len(c.Expressions)}
		}

		if c.Expressions[exprIndex] == nil {
			return "", nil, ExpressionError{
				Err: ExpressionIndexError{Index: exprIndex},
			}
		}

		builder.WriteString(c.Template[:index])
		c.Template = c.Template[index+1:]

		sql, args, err := c.Expressions[exprIndex].ToSQL()
		if err != nil {
			return "", nil, ExpressionError{Err: err}
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex >= len(c.Expressions) {
		return "", nil, NumberOfArgumentsError{Got: exprIndex, Want: len(c.Expressions)}
	}

	return builder.String(), arguments, nil
}

func Append(expressions ...Expression) Joiner {
	return Joiner{Expressions: expressions}
}

func Join(sep string, expressions ...Expression) Joiner {
	return Joiner{Sep: sep, Expressions: expressions}
}

type Joiner struct {
	Sep         string
	Expressions []Expression
}

func (j Joiner) ToSQL() (string, []any, error) {
	builder := &strings.Builder{}
	arguments := make([]any, 0, len(j.Expressions))

	isFirst := true

	for i, expr := range j.Expressions {
		if expr == nil {
			return "", nil, ExpressionError{
				Err: ExpressionIndexError{Index: i},
			}
		}

		sql, args, err := expr.ToSQL()
		if err != nil {
			return "", nil, ExpressionError{Err: err}
		}

		if sql == "" {
			continue
		}

		if isFirst {
			builder.WriteString(sql)

			isFirst = false
		} else {
			builder.WriteString(j.Sep + sql)
		}

		arguments = append(arguments, args...)
	}

	return builder.String(), arguments, nil
}

func If(condition bool, then Expression) Expression {
	if condition {
		return then
	}

	return Raw{}
}

func IfElse(condition bool, then, els Expression) Expression {
	if condition {
		return then
	}

	return els
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
		If(q.With != nil, Compile("WITH ?", q.With)),
		IfElse(q.Select != nil, Compile("SELECT ?", q.Select), SQL("SELECT *")),
		If(q.From != nil, Compile("FROM ?", q.From)),
		If(q.Where != nil, Compile("WHERE ?", q.Where)),
		If(q.GroupBy != nil, Compile("GROUP BY ?", q.GroupBy)),
		If(q.Having != nil, Compile("HAVING ?", q.Having)),
		If(q.Having != nil, Compile("WINDOW ?", q.Window)),
		If(q.OrderBy != nil, Compile("ORDER BY ?", q.OrderBy)),
		If(q.Limit > 0, SQL(fmt.Sprintf("LIMIT %d", q.Limit))),
		If(q.Offset > 0, SQL(fmt.Sprintf("OFFSET %d", q.Offset))),
	).ToSQL()
}

type Insert struct {
	Into    string
	Columns []string
	Data    []Values
}

func (i Insert) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("INSERT INTO %s", i.Into)),
		If(len(i.Columns) > 0, SQL(fmt.Sprintf("(%s)", strings.Join(i.Columns, ", ")))),
		Compile("VALUES ?", Join(", ", Slice(i.Data)...)),
	).ToSQL()
}

type Update struct {
	Table string
	Sets  []Expression
	Where Expression
}

func (u Update) ToSQL() (string, []any, error) {
	return Join(" ",
		Compile(fmt.Sprintf("UPDATE %s SET ?", u.Table), Join(", ", u.Sets...)),
		If(u.Where != nil, Compile("WHERE ?", u.Where)),
	).ToSQL()
}

type Delete struct {
	From  string
	Where Expression
}

func (d Delete) ToSQL() (string, []any, error) {
	return Join(" ",
		SQL(fmt.Sprintf("DELETE FROM %s", d.From)),
		If(d.Where != nil, Compile("WHERE ?", d.Where)),
	).ToSQL()
}

func SQL(sql string, args ...any) Raw {
	return Raw{SQL: sql, Args: args}
}

type Raw struct {
	SQL  string
	Args []any
	Err  error
}

func (r Raw) ToSQL() (string, []any, error) {
	return r.SQL, r.Args, r.Err
}

func Finalize(placeholder string, expr Expression) (string, []any, error) {
	if expr == nil {
		return "", nil, ExpressionError{}
	}

	sql, args, err := expr.ToSQL()
	if err != nil {
		return "", nil, ExpressionError{Err: err}
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
		return "", nil, NumberOfArgumentsError{Got: argIndex, Want: len(args)}
	}

	return build.String(), args, nil
}

// Slice creates a slice of expressions from any slice of expressions.
// This function will exists util there are better alternatives.
func Slice[T Expression](s []T) []Expression {
	out := make([]Expression, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}
