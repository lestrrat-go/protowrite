package protowrite_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/lestrrat-go/protowrite"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	var b protowrite.Builder

	file, err := b.File().
		Package(`foo.bar`).
		Import("google/protobuf/descriptor.proto", protowrite.ImportDefault).
		Enums(
			b.Enum("Unit").
				Element("VOID", 0).
				MustBuild(),
		).
		Messages(
			b.Message("Message").
				OneOfs(
					b.OneOf("id").
						StringField("name", 1).
						Uint64Field("num", 2).
						MustBuild(),
				).
				Messages(
					b.Message("NestedMessage").
						Extensions(
							b.Extension("google.protobuf.MessageOptions").
								StringField("fizz", 49999).
								MustBuild(),
						).
						Enums(
							b.Enum("Kind").
								Element("NULL", 0).
								Element("PRIMARY", 1).
								Element("SECONDARY", 2).
								MustBuild(),
						).
						Option("(NestedMessage.fizz)", strconv.Quote("buzz")).
						Field("Kind", "kind", 1).
						MustBuild(),
				).
				Field("NestedMessage", "extra", 3).
				MustBuild(),
		).
		Services(
			b.Service("FooService").
				Method("Bar", "Message", "Message").
				MustBuild(),
		).
		Build()
	require.NoError(t, err, `builder.Build should succeed`)

	buf, err := protowrite.Marshal(file)
	require.NoError(t, err, `protowrite.Marshal should succeed`)

	expected, err := os.ReadFile(`testdata/sanity.golden`)
	require.NoError(t, err, `io.ReadFile should succeed`)

	require.Equal(t, strings.TrimSpace(string(expected)), string(buf))
}
