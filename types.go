package protokit

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type common struct {
	file     *PKFileDescriptor
	path     string
	LongName string
	FullName string

	OptionExtensions map[string]interface{}
}

func newCommon(f *PKFileDescriptor, path, longName string) common {
	fn := longName
	if !strings.HasPrefix(fn, ".") {
		fn = fmt.Sprintf("%s.%s", f.GetPackage(), longName)
		if !strings.HasPrefix(fn, ".") {
			fn = "." + fn
		}
	}

	return common{
		file:     f,
		path:     path,
		LongName: longName,
		FullName: fn,
	}
}

// GetFile returns the PKFileDescriptor that contains this object
func (c *common) GetFile() *PKFileDescriptor { return c.file }

// GetPackage returns the package this object is in
func (c *common) GetPackage() string { return c.file.GetPackage() }

// GetLongName returns the name prefixed with the dot-separated parent descriptor's name (if any)
func (c *common) GetLongName() string { return c.LongName }

// GetFullName returns the `LongName` prefixed with the package this object is in
func (c *common) GetFullName() string { return c.FullName }

// IsProto3 returns whether or not this is a proto3 object
func (c *common) IsProto3() bool { return c.file.GetSyntax() == "proto3" }

// GetOptionExtensions returns the options defined for this object
func (c *common) GetOptionExtensions() map[string]interface{} { return c.OptionExtensions }

func getOptions(options proto.Message) (m map[string]interface{}) {
	protoregistry.GlobalTypes.RangeExtensions(func(extensionType protoreflect.ExtensionType) bool {
		if extensionType.TypeDescriptor().ContainingMessage().FullName() ==
			options.ProtoReflect().Descriptor().FullName() &&
			options.ProtoReflect().Has(extensionType.TypeDescriptor()) {

			ext := proto.GetExtension(options, extensionType)
			if ext != nil {
				if m == nil {
					m = make(map[string]interface{})
				}
				m[string(extensionType.TypeDescriptor().FullName())] = ext
			}
		}
		return true
	})
	return m
}

func (c *common) setOptions(options proto.Message) {
	if opts := getOptions(options); len(opts) > 0 {
		if c.OptionExtensions == nil {
			c.OptionExtensions = opts
			return
		}
		for k, v := range opts {
			c.OptionExtensions[k] = v
		}
	}
}

// An PKImportedDescriptor describes a type that was imported by a PKFileDescriptor.
type PKImportedDescriptor struct {
	common
}

// A PKFileDescriptor describes a single proto file with all of its messages, enums, services, etc.
type PKFileDescriptor struct {
	comments Comments
	desc     *descriptorpb.FileDescriptorProto

	PackageComments *Comment
	SyntaxComments  *Comment

	Enums              []*PKEnumDescriptor
	Extensions         []*PKExtensionDescriptor
	Imports            []*PKImportedDescriptor
	Messages           []*PKDescriptor
	Services           []*PKServiceDescriptor
	Dependencies       []*PKFileDescriptor
	PublicDependencies []*PKFileDescriptor

	OptionExtensions map[string]interface{}

	FileDescriptor   protoreflect.FileDescriptor
	IsFileToGenerate bool
}

func (f *PKFileDescriptor) ProtoDesc() *descriptorpb.FileDescriptorProto { return f.desc }

func (f *PKFileDescriptor) GetName() string    { return f.ProtoDesc().GetName() }
func (f *PKFileDescriptor) GetPackage() string { return f.ProtoDesc().GetPackage() }
func (f *PKFileDescriptor) GetSyntax() string  { return f.ProtoDesc().GetSyntax() }

func (f *PKFileDescriptor) GetDependencies() []*PKFileDescriptor       { return f.Dependencies }
func (f *PKFileDescriptor) GetPublicDependencies() []*PKFileDescriptor { return f.PublicDependencies }

// IsProto3 returns whether or not this file is a proto3 file
func (f *PKFileDescriptor) IsProto3() bool { return f.GetSyntax() == "proto3" }

// GetPackageComments returns the file's package comments
func (f *PKFileDescriptor) GetPackageComments() *Comment { return f.PackageComments }

// GetSyntaxComments returns the file's syntax comments
func (f *PKFileDescriptor) GetSyntaxComments() *Comment { return f.SyntaxComments }

// GetEnums returns the top-level enumerations defined in this file
func (f *PKFileDescriptor) GetEnums() []*PKEnumDescriptor { return f.Enums }

// GetExtensions returns the top-level (file) extensions defined in this file
func (f *PKFileDescriptor) GetExtensions() []*PKExtensionDescriptor { return f.Extensions }

// GetImports returns the proto files imported by this file
func (f *PKFileDescriptor) GetImports() []*PKImportedDescriptor { return f.Imports }

// GetMessages returns the top-level messages defined in this file
func (f *PKFileDescriptor) GetMessages() []*PKDescriptor { return f.Messages }

// GetServices returns the services defined in this file
func (f *PKFileDescriptor) GetServices() []*PKServiceDescriptor { return f.Services }

// GetOptionExtensions returns the file-level options defined in this file
func (f *PKFileDescriptor) GetOptionExtensions() map[string]interface{} { return f.OptionExtensions }

// GetFileDescriptor returns the underlying `protoreflect.FileDescriptor`
func (f *PKFileDescriptor) GetFileDescriptor() protoreflect.FileDescriptor { return f.FileDescriptor }

// GetIsFileToGenerate returns whether or not this file is to be generated
func (f *PKFileDescriptor) GetIsFileToGenerate() bool { return f.IsFileToGenerate }

// GetEnum returns the enumeration with the specified name (returns `nil` if not found)
func (f *PKFileDescriptor) GetEnum(name string) *PKEnumDescriptor {
	for _, e := range f.GetEnums() {
		if e.GetName() == name || e.GetLongName() == name {
			return e
		}
	}

	return nil
}

// GetMessage returns the message with the specified name (returns `nil` if not found)
func (f *PKFileDescriptor) GetMessage(name string) *PKDescriptor {
	for _, m := range f.GetMessages() {
		if m.GetName() == name || m.GetLongName() == name || m.GetFullName() == name {
			return m
		}
	}

	return nil
}

// GetService returns the service with the specified name (returns `nil` if not found)
func (f *PKFileDescriptor) GetService(name string) *PKServiceDescriptor {
	for _, s := range f.GetServices() {
		if s.GetName() == name || s.GetLongName() == name {
			return s
		}
	}

	return nil
}

func (f *PKFileDescriptor) setOptions(options proto.Message) {
	if opts := getOptions(options); len(opts) > 0 {
		if f.OptionExtensions == nil {
			f.OptionExtensions = opts
			return
		}
		for k, v := range opts {
			f.OptionExtensions[k] = v
		}
	}
}

// An PKEnumDescriptor describe an enum type
type PKEnumDescriptor struct {
	common
	desc     *descriptorpb.EnumDescriptorProto
	Parent   *PKDescriptor
	Values   []*PKEnumValueDescriptor
	Comments *Comment
}

// ProtoDesc returns the underlying `EnumDescriptorProto`
func (e *PKEnumDescriptor) ProtoDesc() *descriptorpb.EnumDescriptorProto { return e.desc }

// GetName returns the name of the enum
func (e *PKEnumDescriptor) GetName() string { return e.ProtoDesc().GetName() }

// GetComments returns a description of this enum
func (e *PKEnumDescriptor) GetComments() *Comment { return e.Comments }

// GetParent returns the parent message (if any) that contains this enum
func (e *PKEnumDescriptor) GetParent() *PKDescriptor { return e.Parent }

// GetValues returns the available values for this enum
func (e *PKEnumDescriptor) GetValues() []*PKEnumValueDescriptor { return e.Values }

// GetNamedValue returns the value with the specified name (returns `nil` if not found)
func (e *PKEnumDescriptor) GetNamedValue(name string) *PKEnumValueDescriptor {
	for _, v := range e.GetValues() {
		if v.GetName() == name {
			return v
		}
	}

	return nil
}

// An PKEnumValueDescriptor describes an enum value
type PKEnumValueDescriptor struct {
	common
	desc     *descriptorpb.EnumValueDescriptorProto
	Enum     *PKEnumDescriptor
	Comments *Comment
}

// ProtoDesc returns the underlying `EnumValueDescriptorProto`
func (v *PKEnumValueDescriptor) ProtoDesc() *descriptorpb.EnumValueDescriptorProto { return v.desc }

// GetName returns the name of the value
func (v *PKEnumValueDescriptor) GetName() string { return v.ProtoDesc().GetName() }

// GetComments returns a description of the value
func (v *PKEnumValueDescriptor) GetComments() *Comment { return v.Comments }

// GetEnum returns the parent enumeration that contains this value
func (v *PKEnumValueDescriptor) GetEnum() *PKEnumDescriptor { return v.Enum }

// An PKExtensionDescriptor describes a protobuf extension. If it's a top-level extension it's parent will be `nil`
type PKExtensionDescriptor struct {
	common
	desc                *descriptorpb.FieldDescriptorProto
	Parent              *PKDescriptor
	Comments            *Comment
	ExtensionDescriptor protoreflect.ExtensionDescriptor
}

// ProtoDesc returns the underlying `desc`
func (e *PKExtensionDescriptor) ProtoDesc() *descriptorpb.FieldDescriptorProto { return e.desc }

// GetExtensionDescriptor returns the underlying `protoreflect.ExtensionDescriptor`
func (e *PKExtensionDescriptor) GetExtensionDescriptor() protoreflect.ExtensionDescriptor {
	return e.ExtensionDescriptor
}

// ExtensionType returns a new `protoreflect.ExtensionType` for this extension
func (e *PKExtensionDescriptor) ExtensionType() protoreflect.ExtensionType {
	return dynamicpb.NewExtensionType(e.GetExtensionDescriptor())
}

// GetName returns the name of the extension
func (e *PKExtensionDescriptor) GetName() string { return e.ProtoDesc().GetName() }

// GetComments returns a description of the extension
func (e *PKExtensionDescriptor) GetComments() *Comment { return e.Comments }

// GetParent returns the descriptor that defined this extension (if any)
func (e *PKExtensionDescriptor) GetParent() *PKDescriptor { return e.Parent }

// A PKDescriptor describes a message
type PKDescriptor struct {
	common
	desc       *descriptorpb.DescriptorProto
	Parent     *PKDescriptor
	Comments   *Comment
	Enums      []*PKEnumDescriptor
	Extensions []*PKExtensionDescriptor
	Fields     []*PKFieldDescriptor
	Messages   []*PKDescriptor
}

func (m *PKDescriptor) ProtoDesc() *descriptorpb.DescriptorProto { return m.desc }

func (m *PKDescriptor) GetName() string { return m.ProtoDesc().GetName() }

// GetComments returns a description of the message
func (m *PKDescriptor) GetComments() *Comment { return m.Comments }

// GetParent returns the parent descriptor (if any) that defines this descriptor
func (m *PKDescriptor) GetParent() *PKDescriptor { return m.Parent }

// GetEnums returns the nested enumerations within the message
func (m *PKDescriptor) GetEnums() []*PKEnumDescriptor { return m.Enums }

// GetExtensions returns the message-level extensions defined by this message
func (m *PKDescriptor) GetExtensions() []*PKExtensionDescriptor { return m.Extensions }

// GetMessages returns the nested messages within the message
func (m *PKDescriptor) GetMessages() []*PKDescriptor { return m.Messages }

// GetMessageFields returns the message fields
func (m *PKDescriptor) GetMessageFields() []*PKFieldDescriptor { return m.Fields }

// GetEnum returns the enum with the specified name. The name can be either simple, or fully qualified (returns `nil` if
// not found)
func (m *PKDescriptor) GetEnum(name string) *PKEnumDescriptor {
	for _, e := range m.GetEnums() {
		// can lookup by name or message prefixed name (qualified)
		if e.GetName() == name || e.GetLongName() == name {
			return e
		}
	}

	return nil
}

// GetMessage returns the nested message with the specified name. The name can be simple or fully qualified (returns
// `nil` if not found)
func (m *PKDescriptor) GetMessage(name string) *PKDescriptor {
	for _, msg := range m.GetMessages() {
		// can lookup by name or message prefixed name (qualified)
		if msg.GetName() == name || msg.GetLongName() == name {
			return msg
		}
	}

	return nil
}

// GetMessageField returns the field with the specified name (returns `nil` if not found)
func (m *PKDescriptor) GetMessageField(name string) *PKFieldDescriptor {
	for _, f := range m.GetMessageFields() {
		if f.GetName() == name || f.GetLongName() == name {
			return f
		}
	}

	return nil
}

// A PKFieldDescriptor describes a message field
type PKFieldDescriptor struct {
	common
	desc     *descriptorpb.FieldDescriptorProto
	Comments *Comment
	Message  *PKDescriptor
}

// ProtoDesc returns the underlying `desc`
func (mf *PKFieldDescriptor) ProtoDesc() *descriptorpb.FieldDescriptorProto { return mf.desc }

// GetName returns the name of the field
func (mf *PKFieldDescriptor) GetName() string { return mf.ProtoDesc().GetName() }

// GetComments returns a description of the field
func (mf *PKFieldDescriptor) GetComments() *Comment { return mf.Comments }

// GetMessage returns the descriptor that defines this field
func (mf *PKFieldDescriptor) GetMessage() *PKDescriptor { return mf.Message }

// A PKServiceDescriptor describes a service
type PKServiceDescriptor struct {
	common
	desc              *descriptorpb.ServiceDescriptorProto
	Comments          *Comment
	Methods           []*PKMethodDescriptor
	ServiceDescriptor protoreflect.ServiceDescriptor
}

// ProtoDesc returns the underlying `desc`
func (s *PKServiceDescriptor) ProtoDesc() *descriptorpb.ServiceDescriptorProto { return s.desc }

// GetName returns the name of the service
func (s *PKServiceDescriptor) GetName() string { return s.ProtoDesc().GetName() }

// GetComments returns a description of the service
func (s *PKServiceDescriptor) GetComments() *Comment { return s.Comments }

// GetMethods returns the methods for the service
func (s *PKServiceDescriptor) GetMethods() []*PKMethodDescriptor { return s.Methods }

// GetNamedMethod returns the method with the specified name (if found)
func (s *PKServiceDescriptor) GetNamedMethod(name string) *PKMethodDescriptor {
	for _, m := range s.GetMethods() {
		if m.GetName() == name || m.GetLongName() == name {
			return m
		}
	}

	return nil
}

// A PKMethodDescriptor describes a method in a service
type PKMethodDescriptor struct {
	common
	desc             *descriptorpb.MethodDescriptorProto
	Comments         *Comment
	Service          *PKServiceDescriptor
	MethodDescriptor protoreflect.MethodDescriptor
	InputType        *PKDescriptor
	OutputType       *PKDescriptor
}

// ProtoDesc returns the underlying `desc`
func (m *PKMethodDescriptor) ProtoDesc() *descriptorpb.MethodDescriptorProto { return m.desc }

// GetName returns the name of the method
func (m *PKMethodDescriptor) GetName() string { return m.ProtoDesc().GetName() }

// GetInputType returns the input message type
func (m *PKMethodDescriptor) GetInputType() *PKDescriptor { return m.InputType }

// GetOutputType returns the output message type
func (m *PKMethodDescriptor) GetOutputType() *PKDescriptor { return m.OutputType }

// GetComments returns a description of the method
func (m *PKMethodDescriptor) GetComments() *Comment { return m.Comments }

// GetService returns the service descriptor that defines this method
func (m *PKMethodDescriptor) GetService() *PKServiceDescriptor { return m.Service }

// GetMethodDescriptor returns the underlying `protoreflect.MethodDescriptor`
func (m *PKMethodDescriptor) GetMethodDescriptor() protoreflect.MethodDescriptor {
	return m.MethodDescriptor
}
