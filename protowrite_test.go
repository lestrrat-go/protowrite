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
	var file protowrite.File

	file.Package = `foo.bar`
	file.Imports = []*protowrite.Import{
		{
			Path: "google/protobuf/descriptor.proto",
		},
	}
	file.Enums = []*protowrite.Enum{
		{
			Name: "Unit",
			Elements: []*protowrite.EnumElement{
				{
					Name:  "VOID",
					Value: 0,
				},
			},
		},
	}
	file.Messages = []*protowrite.Message{
		{
			Name: "Message",
			OneOfs: []*protowrite.OneOf{
				{
					Name: "id",
					Fields: []*protowrite.Field{
						{
							Type: "string",
							Name: "name",
							ID:   1,
						},
						{
							Type: "uint64",
							Name: "num",
							ID:   2,
						},
					},
				},
			},
			Messages: []*protowrite.Message{
				{
					Name: "NestedMessage",
					Options: []*protowrite.Option{
						{
							Name:  "(NestedMessage.fizz)",
							Value: strconv.Quote("buzz"),
						},
					},
					Extensions: []*protowrite.Extension{
						{
							Name: "google.protobuf.MessageOptions",
							Fields: []*protowrite.Field{
								{
									Name: "fizz",
									Type: "string",
									ID:   49999,
								},
							},
						},
					},
					Enums: []*protowrite.Enum{
						{
							Name: "Kind",
							Elements: []*protowrite.EnumElement{
								{
									Name:  "NULL",
									Value: 0,
								},
								{
									Name:  "PRIMARY",
									Value: 1,
								},
								{
									Name:  "SECONDARY",
									Value: 2,
								},
							},
						},
					},
					Fields: []*protowrite.Field{
						{
							Type: "Kind",
							Name: "kind",
							ID:   1,
						},
					},
				},
			},
			Fields: []*protowrite.Field{
				{
					Type: "NestedMessage",
					Name: "extra",
					ID:   3,
				},
			},
		},
	}
	file.Services = []*protowrite.Service{
		{
			Name: "FooService",
			Methods: []*protowrite.Method{
				{
					Name:   "Bar",
					Input:  "Message",
					Output: "Message",
				},
			},
		},
	}
	buf, err := protowrite.Marshal(&file)
	require.NoError(t, err, `protowrite.Marshal should succeed`)

	expected, err := os.ReadFile(`testdata/sanity.golden`)
	require.NoError(t, err, `io.ReadFile should succeed`)

	require.Equal(t, strings.TrimSpace(string(expected)), string(buf))
}
