//nolint:exhaustivestruct,exhaustruct,wrapcheck,ireturn,exhaustive
package superbasic

import (
	"fmt"
	"reflect"
	"strings"
)

// NumberOfArgumentsError is returned if arguments doesn't match the number of placeholders.
type NumberOfArgumentsError struct {
	Got, Want int
}

func (e NumberOfArgumentsError) Error() string {
	if e.Got > 0 || e.Want > 0 {
		return fmt.Sprintf("invalid number of arguments: got '%d' want '%d'", e.Got, e.Want)
	}

	return "invalid number of arguments"
}

// ExpressionError is returned if expressions are nil.
type ExpressionError struct {
	Err error
}

func (e ExpressionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid expression: %s", e.Err.Error())
	}

	return "invalid expression"
}

func (e ExpressionError) Unwrap() error {
	return e.Err
}

type Expression interface {
	ToSQL() (string, []any, error)
}

type Values []any

func (v Values) ToSQL() (string, []any, error) {
	return fmt.Sprintf("(%s)", strings.Repeat(", ?", len(v))[2:]), v, nil
}

// Compile takes a template with placeholders into which expressions can be compiled.
// []Expression is compiled to Join(", ", expr...).
// Escape '?' by using '??'.
func Compile(template string, expressions ...any) Compiler {
	return Compiler{Template: template, Expressions: expressions}
}

type Compiler struct {
	Template    string
	Expressions []any
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
			return "", nil, ExpressionError{}
		}

		builder.WriteString(c.Template[:index])
		c.Template = c.Template[index+1:]

		sql, args, err := compile(c.Expressions[exprIndex], nil)
		if err != nil {
			return "", nil, err
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

	for _, expr := range j.Expressions {
		if expr == nil {
			return "", nil, ExpressionError{}
		}

		sql, args, err := expr.ToSQL()
		if err != nil {
			return "", nil, err
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
		SQL("SELECT"),
		IfElse(q.Select != nil, q.Select, SQL("*")),
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
		Compile("VALUES ?", i.Data),
	).ToSQL()
}

type Update struct {
	Table string
	Sets  []Expression
	Where Expression
}

func (u Update) ToSQL() (string, []any, error) {
	return Join(" ",
		Compile(fmt.Sprintf("UPDATE %s SET ?", u.Table), u.Sets),
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
}

func (r Raw) ToSQL() (string, []any, error) {
	return r.SQL, r.Args, nil
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
		return "", nil, NumberOfArgumentsError{Got: argIndex, Want: len(args)}
	}

	return build.String(), args, nil
}

func toExpressionSlice[T Expression](s []T) []Expression {
	out := make([]Expression, len(s))

	for i := range out {
		out[i] = s[i]
	}

	return out
}

func compile(start any, expression any) (string, []any, error) {
	if expression == nil {
		expression = start
	}

	switch expr := expression.(type) {
	case Expression:
		return expr.ToSQL()
	case []Expression:
		return Join(", ", expr...).ToSQL()
	}

	value := reflect.ValueOf(expression)

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		builder := &strings.Builder{}
		arguments := make([]any, 0, value.Len())

		for index := 0; index < value.Len(); index++ {
			sql, args, err := compile(start, value.Index(index).Interface())
			if err != nil {
				return "", nil, err
			}

			if index != 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(sql)
			arguments = append(arguments, args...)
		}

		return builder.String(), arguments, nil
	default:
		return "?", []any{start}, nil
	}
}
