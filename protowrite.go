// Package protowrite provides simple objects to help you programmatically
// generate a protobuf specification. The protobuf specification emmited through
// the use of this tool IS NOT guaranteed to be syntactically valid, nor are they
// pretty-formatted: It is the caller's responsibility to use tools to
// perform post processing and validation on the generated code.
//
// On the other hand, protowrite provides a saner way to generate protobuf
// specifications compared to, for example, generating them through the use
// of templates. This is achieved by providing users a pseudo-AST style
// API that allows them to treat pieces of information as building blocks to
// compose a protobuf specification
//
// The implementation is based on the specification at https://protobuf.com/docs/language-spec

package protowrite

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

type encoder interface {
	encode(context.Context, io.Writer) error
}

type encodeIndentKey struct{}
type encodeIndentOnceKey struct{}

var Indent = "    "

func getIndentOnce(ctx context.Context) string {
	return ctx.Value(encodeIndentOnceKey{}).(string)
}

func getIndent(ctx context.Context) string {
	v := ctx.Value(encodeIndentKey{})
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func moreIndent(ctx context.Context) context.Context {
	indentOnce := getIndentOnce(ctx)

	cur := getIndent(ctx)
	return context.WithValue(ctx, encodeIndentKey{}, cur+indentOnce)
}

func lessIndent(ctx context.Context) context.Context {
	indentOnce := getIndentOnce(ctx)

	cur := getIndent(ctx)
	return context.WithValue(ctx, encodeIndentKey{}, strings.TrimSuffix(cur, indentOnce))
}

func multilineComment(ctx context.Context, dst io.Writer, s string) {
	indent := getIndent(ctx)
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		txt := scanner.Text()
		fmt.Fprintf(dst, "\n%s// %s", indent, txt)
	}
}

// File represents a protobuf file, which is the top-most level resource
// that this package can generate
type File struct {
	// Package describes the package name
	Package    string
	Imports    []*Import
	Messages   []*Message
	Enums      []*Enum
	Options    []*Option
	Extensions []*Extension
	Services   []*Service
}

func (f *File) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)

	fmt.Fprintf(dst, "%ssyntax = \"proto3\";", indent)
	fmt.Fprintf(dst, "\n\n%spackage %s;", indent, f.Package)

	if list := f.Imports; len(list) > 0 {
		fmt.Fprint(dst, "\n")
		for i, v := range list {
			if err := v.encode(ctx, dst); err != nil {
				return fmt.Errorf(`failed to encode import statement %d: %w`, i, err)
			}
		}
	}

	if list := f.Options; len(list) > 0 {
		fmt.Fprint(dst, "\n")
		for i, v := range list {
			if err := v.encode(ctx, dst); err != nil {
				return fmt.Errorf(`failed to encode option declaration %d: %w`, i, err)
			}
		}
	}

	for i, v := range f.Extensions {
		fmt.Fprintf(dst, "\n")
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode extension declaration %d: %w`, i, err)
		}
	}
	for i, v := range f.Messages {
		fmt.Fprintf(dst, "\n")
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode message declaration %d: %w`, i, err)
		}
	}
	for i, v := range f.Enums {
		fmt.Fprintf(dst, "\n")
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode enum declaration %d: %w`, i, err)
		}
	}
	for i, v := range f.Services {
		fmt.Fprintf(dst, "\n")
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode service declaration %d: %w`, i, err)
		}
	}

	return nil
}

type ImportType int

const (
	ImportDefault ImportType = iota
	ImportPublic
	ImportWeak
)

type Import struct {
	Path string
	Type ImportType
}

func (f *Import) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%simport", indent)
	switch f.Type {
	case ImportPublic:
		fmt.Fprintf(dst, " public")
	case ImportWeak:
		fmt.Fprintf(dst, " weak")
	}
	fmt.Fprintf(dst, " %q;", f.Path)
	return nil
}

// Option represents a protobuf option. No check whatsoever is performed on the syntax of the
// option. The caller is responsible to quote strings, use correct braces, etc.
type Option struct {
	Name    string
	Value   interface{} // Scalar or MessageLiteral
	Compact bool
}

type MessageLiteral struct {
	SingleLine bool
	Fields     []*MessageLiteralField
}

func (ml *MessageLiteral) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprint(dst, "{")
	ctx = moreIndent(ctx)
	for i, field := range ml.Fields {
		if !ml.SingleLine {
			fmt.Fprintf(dst, "\n%s", getIndent(ctx))
		} else if i > 0 {
			fmt.Fprintf(dst, " ")
		}

		if err := field.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode field %d for message literal: %w`, i, err)
		}

	}
	ctx = lessIndent(ctx)
	if !ml.SingleLine {
		fmt.Fprintf(dst, "\n%s", indent)
	}
	fmt.Fprint(dst, "}")
	return nil
}

type MessageLiteralField struct {
	Name  string
	Value interface{}
}

func (mlf *MessageLiteralField) encode(ctx context.Context, dst io.Writer) error {
	var val string
	if e, ok := mlf.Value.(encoder); ok {
		var buf strings.Builder
		if err := e.encode(ctx, &buf); err != nil {
			return fmt.Errorf(`failed to encode option value for message literal %q: %w`, mlf.Name, err)
		}
		val = buf.String()
	} else {
		val = fmt.Sprintf("%#v", mlf.Value)
	}
	fmt.Fprintf(dst, "%s: %s", mlf.Name, val)
	return nil
}

func (o *Option) encode(ctx context.Context, dst io.Writer) error {
	if o.Compact {
		// CompactOption is much like Option, but is used as part of other declarations.
		// These have no newlines, are enclosed within '[' and ']', and are concatenated using commas
		fmt.Fprintf(dst, "%s = %s", o.Name, o.Value)
		return nil
	}
	indent := getIndent(ctx)

	var val string
	if e, ok := o.Value.(encoder); ok {
		var buf strings.Builder
		if err := e.encode(ctx, &buf); err != nil {
			return fmt.Errorf(`failed to encode option value for option %q: %w`, o.Name, err)
		}
		val = buf.String()
	} else {
		val = fmt.Sprintf("%#v", o.Value)
	}

	fmt.Fprintf(dst, "\n%soption %s = %s;", indent, o.Name, val)
	return nil
}

type OneOf struct {
	Name   string
	Fields []*Field
}

func (oo *OneOf) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%soneof %s {", indent, oo.Name)
	for i, v := range oo.Fields {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode field declaration %d for oneof %q: %w`, i, oo.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	fmt.Fprintf(dst, "\n%s}", indent)
	return nil
}

type Enum struct {
	Name     string
	Elements []*EnumElement
	Comment  string
}

func (e *Enum) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)

	if s := e.Comment; s != "" {
		multilineComment(ctx, dst, s)
	}
	fmt.Fprintf(dst, "\n%senum %s {", indent, e.Name)
	for i, v := range e.Elements {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode enum declaration %d for enum %q: %w`, i, e.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	fmt.Fprintf(dst, "\n%s}", indent)
	return nil
}

type EnumElement struct {
	Name    string
	Value   int
	Comment string
}

func (ee *EnumElement) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%s%s = %d;", indent, ee.Name, ee.Value)
	if s := ee.Comment; s != "" {
		// no new lines here
		s = strings.Replace(s, "\n", " ", -1)
		fmt.Fprintf(dst, " // %s", s)
	}
	return nil
}

type Extension struct {
	Name   string
	Fields []*Field
}

func (e *Extension) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%sextend %s {", indent, e.Name)
	for i, v := range e.Fields {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode field declaration %d for extension %q: %w`, i, e.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	fmt.Fprintf(dst, "\n%s}", indent)
	return nil
}

type Message struct {
	Name       string
	Fields     []*Field
	OneOfs     []*OneOf
	Messages   []*Message
	Enums      []*Enum
	Extensions []*Extension
	Options    []*Option
}

func (m *Message) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%smessage %s {", indent, m.Name)

	for i, v := range m.OneOfs {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode nested oneof declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	for i, v := range m.Extensions {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode nested extension declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	for i, v := range m.Options {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode nested option declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	for i, v := range m.Enums {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode nested enum declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	for i, v := range m.Messages {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode nested message declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	for i, v := range m.Fields {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode field declaration %d for message %q: %w`, i, m.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	fmt.Fprintf(dst, "\n%s}", indent)
	return nil
}

type FieldCardinality int

const (
	CardinalityDefault FieldCardinality = iota
	CardinalityRequired
	CardinalityOptional
	CardinalityRepeated
)

type Field struct {
	Type        string
	Name        string
	ID          int
	Cardinality FieldCardinality
	Options     []*Option
}

func (f *Field) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%s", indent)
	switch f.Cardinality {
	case CardinalityRequired:
		fmt.Fprintf(dst, "required")
	case CardinalityOptional:
		fmt.Fprintf(dst, "optional")
	case CardinalityRepeated:
		fmt.Fprintf(dst, "repeated")
	}
	fmt.Fprintf(dst, "%s %s = %d", f.Type, f.Name, f.ID)

	if options := f.Options; len(options) > 0 {
		fmt.Fprintf(dst, " [")
		for i, option := range options {
			if err := option.encode(ctx, dst); err != nil {
				return fmt.Errorf(`failed to encode option %d for field %q: %w`, i, f.Name, err)
			}
		}
		fmt.Fprintf(dst, "]")
	}
	fmt.Fprintf(dst, ";")

	return nil
}

type Service struct {
	Name    string
	Methods []*Method
}

func (s *Service) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%sservice %s {", indent, s.Name)
	for i, v := range s.Methods {
		ctx = moreIndent(ctx)
		if err := v.encode(ctx, dst); err != nil {
			return fmt.Errorf(`failed to encode method %d for service %q: %w`, i, s.Name, err)
		}
		ctx = lessIndent(ctx)
	}
	fmt.Fprint(dst, "\n}", indent)
	return nil
}

type Method struct {
	Name    string
	Input   string
	Output  string
	Options []*Option
}

func (m *Method) encode(ctx context.Context, dst io.Writer) error {
	indent := getIndent(ctx)
	fmt.Fprintf(dst, "\n%srpc %s(%s) returns (%s)", indent, m.Name, m.Input, m.Output)
	if options := m.Options; len(options) > 0 {
		fmt.Fprintf(dst, " {")
		ctx = moreIndent(ctx)
		for i, option := range options {
			if err := option.encode(ctx, dst); err != nil {
				return fmt.Errorf(`failed to encode option %d for method %q: %w`, i, m.Name, err)
			}
		}
		ctx = lessIndent(ctx)
		fmt.Fprintf(dst, "\n%s}", indent)
	}
	fmt.Fprintf(dst, ";")
	return nil
}

func Marshal(f *File) ([]byte, error) {
	ctx := context.WithValue(context.Background(), encodeIndentOnceKey{}, Indent)

	var dst bytes.Buffer
	if err := f.encode(ctx, &dst); err != nil {
		return nil, fmt.Errorf(`failed to write protobuf: %w`, err)
	}
	return dst.Bytes(), nil
}
