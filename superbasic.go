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

// SQL takes b template with placeholders into which expressions can be compiled.
// []Expression is compiled to Join(", ", expr...).
// Escape '?' by using '??'.
func SQL(sql string, expressions ...any) Raw {
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
			return Raw{Err: NumberOfArgumentsError{Got: exprIndex, Want: len(expressions)}}
		}

		if expressions[exprIndex] == nil {
			return Raw{Err: ExpressionError{}}
		}

		builder.WriteString(sql[:index])
		sql = sql[index+1:]

		sql, args, err := compile(expressions[exprIndex])
		if err != nil {
			return Raw{Err: err}
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex >= len(expressions) {
		return Raw{Err: NumberOfArgumentsError{Got: exprIndex, Want: len(expressions)}}
	}

	return Raw{SQL: builder.String(), Args: arguments}
}

type Append []Expression

func (b Append) ToSQL() (string, []any, error) {
	return Join{Expressions: b}.ToSQL()
}

type Join struct {
	Sep         string
	Expressions []Expression
}

func (j Join) ToSQL() (string, []any, error) {
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
	return Join{
		Sep: " ",
		Expressions: []Expression{
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
		},
	}.ToSQL()
}

type Insert struct {
	Into    string
	Columns []string
	Data    []Values
}

func (i Insert) ToSQL() (string, []any, error) {
	return Join{
		Sep: " ",
		Expressions: []Expression{
			SQL(fmt.Sprintf("INSERT INTO %s", i.Into)),
			If(len(i.Columns) > 0, SQL(fmt.Sprintf("(%s)", strings.Join(i.Columns, ", ")))),
			SQL("VALUES ?", i.Data),
		},
	}.ToSQL()
}

type Update struct {
	Table string
	Sets  []Expression
	Where Expression
}

func (u Update) ToSQL() (string, []any, error) {
	return Join{
		Sep: " ",
		Expressions: []Expression{
			SQL(fmt.Sprintf("UPDATE %s SET ?", u.Table), u.Sets),
			If(u.Where != nil, SQL("WHERE ?", u.Where)),
		},
	}.ToSQL()
}

type Delete struct {
	From  string
	Where Expression
}

func (d Delete) ToSQL() (string, []any, error) {
	return Join{
		Sep: " ",
		Expressions: []Expression{
			SQL(fmt.Sprintf("DELETE FROM %s", d.From)),
			If(d.Where != nil, SQL("WHERE ?", d.Where)),
		},
	}.ToSQL()
}

type Raw struct {
	SQL  string
	Args []any
	Err  error
}

func (e Raw) ToSQL() (string, []any, error) {
	if e.Err != nil {
		return "", nil, e.Err
	}

	return e.SQL, e.Args, e.Err
}

func Build() *Builder {
	return &Builder{
		builder:   &strings.Builder{},
		arguments: make([]any, 0, 1),
	}
}

type Builder struct {
	builder   *strings.Builder
	arguments []any
	err       error
}

func (b *Builder) ToSQL() (string, []any, error) {
	if b.builder == nil {
		return "", nil, nil
	}

	return b.builder.String(), b.arguments, nil
}

func (b *Builder) Space() *Builder {
	if b.err != nil {
		return b
	}

	if b.builder == nil {
		b.builder = &strings.Builder{}
	}

	b.builder.WriteString(" ")

	return b
}

func (b *Builder) Append(expressions ...Expression) *Builder {
	if len(expressions) == 0 || b.err != nil {
		return b
	}

	if b.builder == nil {
		b.builder = &strings.Builder{}
	}

	for _, expr := range expressions {
		sql, args, err := expr.ToSQL()
		if err != nil {
			b.err = err

			return b
		}

		b.builder.WriteString(sql)

		b.arguments = append(b.arguments, args...)
	}

	return b
}

func (b *Builder) SQL(sql string, args ...any) *Builder {
	return b.Append(SQL(sql, args...))
}

func (b *Builder) Join(sep string, expressions ...Expression) *Builder {
	return b.Append(Join{Sep: sep, Expressions: expressions})
}

func (b *Builder) If(condition bool, then Expression) *Builder {
	if condition {
		return b.Append(then)
	}

	return b
}

func (b *Builder) IfElse(condition bool, then Expression, els Expression) *Builder {
	if condition {
		return b.Append(then)
	}

	return b.Append(els)
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

func compile(expression any) (string, []any, error) {
	switch expr := expression.(type) {
	case Expression:
		return expr.ToSQL()
	case []Expression:
		return Join{
			Sep:         ", ",
			Expressions: expr,
		}.ToSQL()
	}

	value := reflect.ValueOf(expression)

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		builder := &strings.Builder{}
		arguments := make([]any, 0, value.Len())

		for index := 0; index < value.Len(); index++ {
			sql, args, err := compile(value.Index(index).Interface())
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
		return "?", []any{expression}, nil
	}
}
