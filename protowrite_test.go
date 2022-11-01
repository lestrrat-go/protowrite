package protowrite_test

import (
	"os"
	"strings"
	"testing"

	"github.com/lestrrat-go/protowrite"
	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	var b protowrite.Builder

	cmpProtobuf := func(file *protowrite.File, filename string) {
		buf, err := protowrite.Marshal(file)
		require.NoError(t, err, `protowrite.Marshal should succeed`)

		expected, err := os.ReadFile(filename)
		require.NoError(t, err, `io.ReadFile should succeed`)
		require.Equal(t, strings.TrimSpace(string(expected)), string(buf))
	}

	t.Run("Sanity", func(t *testing.T) {
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
							Option("(NestedMessage.fizz)", "buzz").
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

		cmpProtobuf(file, `testdata/sanity.golden`)
	})
	t.Run("Any", func(t *testing.T) {
		file, err := b.File().
			Package(`foo.bar`).
			Import("google/protobuf/descriptor.proto", protowrite.ImportDefault).
			Extensions(
				b.Extension("google.protobuf.MessageOptions").
					Field("google.protobuf.Any", "extra", 33333).
					MustBuild(),
			).
			Messages(
				b.Message("MyOptionData").
					StringField("name", 1).
					Uint64Field("id", 3).
					MustBuild(),
				b.Message("MyMessage").
					Option("(extra)", b.MessageLiteral().
						Field("[googleapis.com/foo.bar.MyOptionData]", b.MessageLiteral().
							Field("name", "foobar").
							Field("id", 42).
							MustBuild(),
						).
						MustBuild(),
					).
					MustBuild(),
			).
			Build()
		require.NoError(t, err, `builder.Build should succeed`)

		cmpProtobuf(file, `testdata/any.golden`)
	})
}
