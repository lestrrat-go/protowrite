protowrite
==========

Simple Go Objects to construct and serialize a protobuf file programmatically

# SYNOPSIS

To generate the following protobuf:

```protobuf
syntax = "proto3";

package foo.bar;

import "google/protobuf/descriptor.proto";

message Message {
    oneof id {
        string name = 1;
        uint64 num = 2;
    }
    message NestedMessage {
        extend google.protobuf.MessageOptions {
            string fizz = 49999;
        }
        option (NestedMessage.fizz) = "buzz";
        enum Kind {
            NULL = 0;
            PRIMARY = 1;
            SECONDARY = 2;
        }
        Kind kind = 1;
    }
    NestedMessage extra = 3;
}

enum Unit {
    VOID = 0;
}

service FooService {
    rpc Bar(Message) returns (Message);
}
```

You can execute the followning code

```go
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
```
