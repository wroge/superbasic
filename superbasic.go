//nolint:exhaustivestruct,exhaustruct,ireturn
package superbasic

import (
	"fmt"
	"strings"
)

// Error wraps any error in this package and can be used to create an Expression.
type Error struct {
	Err error
}

func (e Error) Error() string {
	if e.Err != nil {
		//nolint:errorlint
		if err, ok := e.Err.(Error); ok {
			return err.Error()
		}

		return fmt.Sprintf("superbasic.Error: %s", e.Err.Error())
	}

	return "superbasic.Error"
}

func (e Error) Unwrap() error {
	return e.Err
}

func (e Error) ToSQL() (string, []any, error) {
	return "", nil, e
}

// ExpressionError is returned by the Compile Expression, if an expression is nil.
type ExpressionError struct {
	Position int
}

func (e ExpressionError) Error() string {
	return fmt.Sprintf("expression at position '%d' is nil", e.Position)
}

// NumberOfArgumentsError is returned if arguments doesn't match the number of placeholders.
type NumberOfArgumentsError struct {
	SQL                     string
	Placeholders, Arguments int
}

func (e NumberOfArgumentsError) Error() string {
	argument := "argument"

	if e.Arguments > 1 {
		argument += "s"
	}

	placeholder := "placeholder"

	if e.Placeholders > 1 {
		placeholder += "s"
	}

	return fmt.Sprintf("%d %s and %d %s in '%s'",
		e.Placeholders, placeholder, e.Arguments, argument, e.SQL)
}

type Expression interface {
	ToSQL() (string, []any, error)
}

func Value(a any) Raw {
	return Raw{SQL: "?", Args: []any{a}}
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
			return "", nil, Error{
				Err: NumberOfArgumentsError{
					SQL:          builder.String(),
					Placeholders: exprIndex,
					Arguments:    len(c.Expressions),
				},
			}
		}

		if c.Expressions[exprIndex] == nil {
			return "", nil, Error{Err: ExpressionError{Position: exprIndex}}
		}

		builder.WriteString(c.Template[:index])
		c.Template = c.Template[index+1:]

		sql, args, err := c.Expressions[exprIndex].ToSQL()
		if err != nil {
			return "", nil, Error{Err: err}
		}

		builder.WriteString(sql)

		arguments = append(arguments, args...)
	}

	if exprIndex != len(c.Expressions)-1 {
		return "", nil, Error{
			Err: NumberOfArgumentsError{
				SQL:          builder.String(),
				Placeholders: exprIndex,
				Arguments:    len(c.Expressions),
			},
		}
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
			return "", nil, Error{Err: ExpressionError{}}
		}

		sql, args, err := expr.ToSQL()
		if err != nil {
			return "", nil, Error{Err: err}
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

// Finalize takes a static placeholder like '?' or a positional placeholder containing '%d'.
// Escaped placeholders ('??') are replaced to '?' when placeholder argument is not '?'.
func Finalize(placeholder string, expression Expression) (string, []any, error) {
	if expression == nil {
		return "", nil, Error{Err: ExpressionError{}}
	}

	sql, args, err := expression.ToSQL()
	if err != nil {
		return "", nil, Error{Err: err}
	}

	var count int

	sql, count = Replace(placeholder, sql)

	if count != len(args) {
		return "", nil, Error{Err: NumberOfArgumentsError{SQL: sql, Placeholders: count, Arguments: len(args)}}
	}

	return sql, args, nil
}

// Replace takes a static placeholder like '?' or a positional placeholder containing '%d'.
// Escaped placeholders ('??') are replaced to '?' when placeholder argument is not '?'.
func Replace(placeholder string, sql string) (string, int) {
	build := &strings.Builder{}
	count := 0

	question := "?"
	positional := false

	if placeholder == "?" {
		question = "??"
	}

	if strings.Contains(placeholder, "%d") {
		positional = true
	}

	for {
		index := strings.IndexRune(sql, '?')
		if index < 0 {
			build.WriteString(sql)

			break
		}

		if index < len(sql)-1 && sql[index+1] == '?' {
			build.WriteString(sql[:index] + question)
			sql = sql[index+2:]

			continue
		}

		count++

		build.WriteString(sql[:index])

		if positional {
			build.WriteString(fmt.Sprintf(placeholder, count))
		} else {
			build.WriteString(placeholder)
		}

		sql = sql[index+1:]
	}

	return build.String(), count
}
