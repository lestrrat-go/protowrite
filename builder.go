package protowrite

func StringField(name string, id int) *Field {
	return &Field{
		Type: "string",
		Name: name,
		ID:   id,
	}
}

func Uint64Field(name string, id int) *Field {
	return &Field{
		Type: "uint64",
		Name: name,
		ID:   id,
	}
}

type Builder struct{}

func (b *Builder) Enum(name string) *EnumBuilder {
	return &EnumBuilder{object: &Enum{Name: name}}
}

func (b *Builder) EnumElement(name string, value int) *EnumElementBuilder {
	return &EnumElementBuilder{object: &EnumElement{Name: name, Value: value}}
}

func (b *Builder) Extension(name string) *ExtensionBuilder {
	return &ExtensionBuilder{object: &Extension{Name: name}}
}

func (b *Builder) File() *FileBuilder {
	return &FileBuilder{object: &File{}}
}

func (b *Builder) Message(name string) *MessageBuilder {
	return &MessageBuilder{object: &Message{Name: name}}
}

func (b *Builder) MessageLiteral() *MessageLiteralBuilder {
	return &MessageLiteralBuilder{object: &MessageLiteral{}}
}

func (b *Builder) OneOf(name string) *OneOfBuilder {
	return &OneOfBuilder{object: &OneOf{Name: name}}
}

func (b *Builder) Service(name string) *ServiceBuilder {
	return &ServiceBuilder{object: &Service{Name: name}}
}

type ExtensionBuilder struct {
	object *Extension
}

func (b *ExtensionBuilder) StringField(name string, id int) *ExtensionBuilder {
	b.object.Fields = append(b.object.Fields, StringField(name, id))
	return b
}

func (b *ExtensionBuilder) Uint64Field(name string, id int) *ExtensionBuilder {
	b.object.Fields = append(b.object.Fields, Uint64Field(name, id))
	return b
}

func (b *ExtensionBuilder) Field(typ, name string, id int) *ExtensionBuilder {
	b.object.Fields = append(b.object.Fields, &Field{
		Type: typ,
		Name: name,
		ID:   id,
	})
	return b
}

func (b *ExtensionBuilder) MustBuild() *Extension {
	return b.object
}

type FileBuilder struct {
	object *File
}

func (b *FileBuilder) Package(s string) *FileBuilder {
	b.object.Package = s
	return b
}

// Import adds a single import type.
func (b *FileBuilder) Import(path string, typ ImportType) *FileBuilder {
	b.Imports(&Import{
		Path: path,
		Type: typ,
	})
	return b
}

func (b *FileBuilder) Imports(v ...*Import) *FileBuilder {
	b.object.Imports = append(b.object.Imports, v...)
	return b
}

func (b *FileBuilder) Enums(v ...*Enum) *FileBuilder {
	b.object.Enums = append(b.object.Enums, v...)
	return b
}

func (b *FileBuilder) Extensions(v ...*Extension) *FileBuilder {
	b.object.Extensions = append(b.object.Extensions, v...)
	return b
}

func (b *FileBuilder) Messages(v ...*Message) *FileBuilder {
	b.object.Messages = append(b.object.Messages, v...)
	return b
}

func (b *FileBuilder) Option(name string, value interface{}) *FileBuilder {
	b.object.Options = append(b.object.Options, &Option{
		Name:  name,
		Value: value,
	})
	return b
}

func (b *FileBuilder) Services(v ...*Service) *FileBuilder {
	b.object.Services = append(b.object.Services, v...)
	return b
}

func (b *FileBuilder) Build() (*File, error) {
	return b.object, nil
}

type EnumBuilder struct {
	object *Enum
}

func (b *EnumBuilder) MustBuild() *Enum {
	return b.object
}

func (b *EnumBuilder) Comment(s string) *EnumBuilder {
	b.object.Comment = s
	return b
}

func (b *EnumBuilder) Element(name string, value int) *EnumBuilder {
	return b.EnumElements(&EnumElement{
		Name:  name,
		Value: value,
	})
}

func (b *EnumBuilder) EnumElements(el ...*EnumElement) *EnumBuilder {
	b.object.Elements = append(b.object.Elements, el...)
	return b
}

type EnumElementBuilder struct {
	object *EnumElement
}

func (b *EnumElementBuilder) Comment(s string) *EnumElementBuilder {
	b.object.Comment = s
	return b
}

func (b *EnumElementBuilder) MustBuild() *EnumElement {
	return b.object
}

type MessageBuilder struct {
	object *Message
}

func (b *MessageBuilder) Comment(s string) *MessageBuilder {
	b.object.Comment = s
	return b
}

func (b *MessageBuilder) StringField(name string, id int) *MessageBuilder {
	b.object.Fields = append(b.object.Fields, StringField(name, id))
	return b
}

func (b *MessageBuilder) Uint64Field(name string, id int) *MessageBuilder {
	b.object.Fields = append(b.object.Fields, Uint64Field(name, id))
	return b
}

func (b *MessageBuilder) Option(name string, value interface{}) *MessageBuilder {
	b.object.Options = append(b.object.Options, &Option{
		Name:  name,
		Value: value,
	})
	return b
}

func (b *MessageBuilder) Extensions(v ...*Extension) *MessageBuilder {
	b.object.Extensions = append(b.object.Extensions, v...)
	return b
}

func (b *MessageBuilder) Messages(v ...*Message) *MessageBuilder {
	b.object.Messages = append(b.object.Messages, v...)
	return b
}

func (b *MessageBuilder) OneOfs(v ...*OneOf) *MessageBuilder {
	b.object.OneOfs = append(b.object.OneOfs, v...)
	return b
}

func (b *MessageBuilder) Enums(v ...*Enum) *MessageBuilder {
	b.object.Enums = append(b.object.Enums, v...)
	return b
}

func (b *MessageBuilder) Field(typ, name string, id int) *MessageBuilder {
	return b.Fields(&Field{
		Type: typ,
		Name: name,
		ID:   id,
	})
}

func (b *MessageBuilder) Fields(v ...*Field) *MessageBuilder {
	b.object.Fields = append(b.object.Fields, v...)
	return b
}

func (b *MessageBuilder) Build() (*Message, error) {
	return b.object, nil
}

func (b *MessageBuilder) MustBuild() *Message {
	obj, err := b.Build()
	if err != nil {
		panic(err)
	}
	return obj
}

type OneOfBuilder struct {
	object *OneOf
}

func (b *OneOfBuilder) StringField(name string, id int) *OneOfBuilder {
	b.object.Fields = append(b.object.Fields, StringField(name, id))
	return b
}

func (b *OneOfBuilder) Uint64Field(name string, id int) *OneOfBuilder {
	b.object.Fields = append(b.object.Fields, Uint64Field(name, id))
	return b
}

func (b *OneOfBuilder) MustBuild() *OneOf {
	return b.object
}

type ServiceBuilder struct {
	object *Service
}

func (b *ServiceBuilder) Method(name, input, output string) *ServiceBuilder {
	b.object.Methods = append(b.object.Methods, &Method{
		Name:   name,
		Input:  input,
		Output: output,
	})
	return b
}

func (b *ServiceBuilder) MustBuild() *Service {
	return b.object
}

type MessageLiteralBuilder struct {
	object *MessageLiteral
}

func (b *MessageLiteralBuilder) Field(name string, value interface{}) *MessageLiteralBuilder {
	b.object.Fields = append(b.object.Fields, &MessageLiteralField{
		Name:  name,
		Value: value,
	})
	return b
}

func (b *MessageLiteralBuilder) MustBuild() *MessageLiteral {
	return b.object
}
