package protokit

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"strings"
)

const (
	// tag numbers in FileDescriptorProto
	packageCommentPath   = 2
	messageCommentPath   = 4
	enumCommentPath      = 5
	serviceCommentPath   = 6
	extensionCommentPath = 7
	syntaxCommentPath    = 12

	// tag numbers in DescriptorProto
	messageFieldCommentPath     = 2 // field
	messageMessageCommentPath   = 3 // nested_type
	messageEnumCommentPath      = 4 // enum_type
	messageExtensionCommentPath = 6 // extension

	// tag numbers in EnumDescriptorProto
	enumValueCommentPath = 2 // value

	// tag numbers in ServiceDescriptorProto
	serviceMethodCommentPath = 2
)

func getAllFileDescriptor(req *pluginpb.CodeGeneratorRequest) map[string]protoreflect.FileDescriptor {
	allFileDesc := make(map[string]protoreflect.FileDescriptor)
	fileDescSet := &descriptorpb.FileDescriptorSet{}
	for _, pf := range req.GetProtoFile() {
		fileDescSet.File = append(fileDescSet.File, pf)
	}
	files, err := protodesc.NewFiles(fileDescSet)
	if err != nil {
		log.Fatal(err)
	}
	for _, pf := range req.GetProtoFile() {
		f, err := protodesc.NewFile(pf, files)
		if err != nil {
			log.Fatal(err)
		}
		allFileDesc[pf.GetName()] = f
	}
	return allFileDesc
}

func registerAllExtensions(allFileDesc map[string]protoreflect.FileDescriptor) {
	for _, fileDesc := range allFileDesc {
		exts := fileDesc.Extensions()
		for i := 0; i < exts.Len(); i++ {
			ext := exts.Get(i)
			err := protoregistry.GlobalTypes.RegisterExtension(dynamicpb.NewExtensionType(ext))
			if err != nil {
				log.Fatal(err)
			}
		}

	}
}
func reUnmarshalReq(req *pluginpb.CodeGeneratorRequest) (err error) {
	reqData, err := proto.Marshal(req)
	if err != nil {
		return
	}
	err = proto.Unmarshal(reqData, req)
	if err != nil {
		return
	}
	return
}

func ParseCodeGenRequestAllFiles(req *pluginpb.CodeGeneratorRequest) ([]*PKFileDescriptor, error) {
	allFiles := make(map[string]*PKFileDescriptor)
	genFiles := make([]*PKFileDescriptor, 0, len(req.GetProtoFile()))

	allFileDesc := getAllFileDescriptor(req)
	registerAllExtensions(allFileDesc)
	err := reUnmarshalReq(req)
	if err != nil {
		return nil, err
	}

	for _, pf := range req.GetProtoFile() {
		allFiles[pf.GetName()] = parseFile(context.Background(), pf, allFileDesc[pf.GetName()])
	}

	for _, f := range allFiles {

		parseAllImports(f, allFiles)

		genFiles = append(genFiles, f)
	}
	for _, f := range req.FileToGenerate {
		allFiles[f].IsFileToGenerate = true
	}

	return genFiles, nil
}

func parseFile(ctx context.Context, fd *descriptorpb.FileDescriptorProto, f protoreflect.FileDescriptor) *PKFileDescriptor {
	comments := ParseComments(fd)

	file := &PKFileDescriptor{
		comments:            comments,
		FileDescriptorProto: fd,
		Comments:            comments.Get(fmt.Sprintf("%d", packageCommentPath)),
		PackageComments:     comments.Get(fmt.Sprintf("%d", packageCommentPath)),
		SyntaxComments:      comments.Get(fmt.Sprintf("%d", syntaxCommentPath)),
		FileDescriptor:      f,
	}

	if fd.Options != nil {
		file.setOptions(fd.Options)
	}

	fileCtx := ContextWithFileDescriptor(ctx, file)
	file.Enums = parseEnums(fileCtx, fd.GetEnumType())
	file.Extensions = parseExtensions(fileCtx, fd.GetExtension())
	file.Messages = parseMessages(fileCtx, fd.GetMessageType())
	file.Services = parseServices(fileCtx, fd.GetService())

	return file
}

func parseEnums(ctx context.Context, protos []*descriptorpb.EnumDescriptorProto) []*PKEnumDescriptor {
	enums := make([]*PKEnumDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)
	parent, hasParent := DescriptorFromContext(ctx)

	for i, ed := range protos {
		longName := ed.GetName()
		commentPath := fmt.Sprintf("%d.%d", enumCommentPath, i)

		if hasParent {
			longName = fmt.Sprintf("%s.%s", parent.GetLongName(), longName)
			commentPath = fmt.Sprintf("%s.%d.%d", parent.path, messageEnumCommentPath, i)
		}

		enums[i] = &PKEnumDescriptor{
			common:              newCommon(file, commentPath, longName),
			EnumDescriptorProto: ed,
			Comments:            file.comments.Get(commentPath),
			Parent:              parent,
		}
		if ed.Options != nil {
			enums[i].setOptions(ed.Options)
		}

		subCtx := ContextWithEnumDescriptor(ctx, enums[i])
		enums[i].Values = parseEnumValues(subCtx, ed.GetValue())
	}

	return enums
}

func parseEnumValues(ctx context.Context, protos []*descriptorpb.EnumValueDescriptorProto) []*PKEnumValueDescriptor {
	values := make([]*PKEnumValueDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)
	enum, _ := EnumDescriptorFromContext(ctx)

	for i, vd := range protos {
		longName := fmt.Sprintf("%s.%s", enum.GetLongName(), vd.GetName())

		values[i] = &PKEnumValueDescriptor{
			common:                   newCommon(file, "", longName),
			EnumValueDescriptorProto: vd,
			Enum:                     enum,
			Comments:                 file.comments.Get(fmt.Sprintf("%s.%d.%d", enum.path, enumValueCommentPath, i)),
		}
		if vd.Options != nil {
			values[i].setOptions(vd.Options)
		}
	}

	return values
}

func parseExtensions(ctx context.Context, protos []*descriptorpb.FieldDescriptorProto) []*PKExtensionDescriptor {
	exts := make([]*PKExtensionDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)
	parent, hasParent := DescriptorFromContext(ctx)

	for i, ext := range protos {
		commentPath := fmt.Sprintf("%d.%d", extensionCommentPath, i)
		longName := fmt.Sprintf("%s.%s", ext.GetExtendee(), ext.GetName())

		if strings.Contains(longName, file.GetPackage()) {
			parts := strings.Split(ext.GetExtendee(), ".")
			longName = fmt.Sprintf("%s.%s", parts[len(parts)-1], ext.GetName())
		}

		if hasParent {
			commentPath = fmt.Sprintf("%s.%d.%d", parent.path, messageExtensionCommentPath, i)
		}

		exts[i] = &PKExtensionDescriptor{
			common:               newCommon(file, commentPath, longName),
			FieldDescriptorProto: ext,
			Comments:             file.comments.Get(commentPath),
			Parent:               parent,
			ExtensionDescriptor:  file.FileDescriptor.Extensions().ByName(protoreflect.Name(ext.GetName())),
		}
		if ext.Options != nil {
			exts[i].setOptions(ext.Options)
		}
	}

	return exts
}

func parseAllImports(fd *PKFileDescriptor, allFiles map[string]*PKFileDescriptor) {
	fd.Imports = make([]*PKImportedDescriptor, 0)

	for _, fileName := range fd.GetDependency() {
		file := allFiles[fileName]

		for _, d := range file.GetMessages() {
			// skip map entry objects
			if !d.GetOptions().GetMapEntry() {
				fd.Imports = append(fd.Imports, &PKImportedDescriptor{d.common})
			}
		}

		for _, e := range file.GetEnums() {
			fd.Imports = append(fd.Imports, &PKImportedDescriptor{e.common})
		}

		for _, ext := range file.GetExtensions() {
			fd.Imports = append(fd.Imports, &PKImportedDescriptor{ext.common})
		}
	}
}

func parseMessages(ctx context.Context, protos []*descriptorpb.DescriptorProto) []*PKDescriptor {
	msgs := make([]*PKDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)
	parent, hasParent := DescriptorFromContext(ctx)

	for i, md := range protos {
		longName := md.GetName()
		commentPath := fmt.Sprintf("%d.%d", messageCommentPath, i)

		if hasParent {
			longName = fmt.Sprintf("%s.%s", parent.GetLongName(), longName)
			commentPath = fmt.Sprintf("%s.%d.%d", parent.path, messageMessageCommentPath, i)
		}

		msgs[i] = &PKDescriptor{
			common:          newCommon(file, commentPath, longName),
			DescriptorProto: md,
			Comments:        file.comments.Get(commentPath),
			Parent:          parent,
		}
		if md.Options != nil {
			msgs[i].setOptions(md.Options)
		}

		msgCtx := ContextWithDescriptor(ctx, msgs[i])
		msgs[i].Enums = parseEnums(msgCtx, md.GetEnumType())
		msgs[i].Extensions = parseExtensions(msgCtx, md.GetExtension())
		msgs[i].Fields = parseMessageFields(msgCtx, md.GetField())
		msgs[i].Messages = parseMessages(msgCtx, md.GetNestedType())
	}

	return msgs
}

func parseMessageFields(ctx context.Context, protos []*descriptorpb.FieldDescriptorProto) []*PKFieldDescriptor {
	fields := make([]*PKFieldDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)
	message, _ := DescriptorFromContext(ctx)

	for i, fd := range protos {
		longName := fmt.Sprintf("%s.%s", message.GetLongName(), fd.GetName())

		fields[i] = &PKFieldDescriptor{
			common:               newCommon(file, "", longName),
			FieldDescriptorProto: fd,
			Comments:             file.comments.Get(fmt.Sprintf("%s.%d.%d", message.path, messageFieldCommentPath, i)),
			Message:              message,
		}
		if fd.Options != nil {
			fields[i].setOptions(fd.Options)
		}
	}

	return fields
}

func parseServices(ctx context.Context, protos []*descriptorpb.ServiceDescriptorProto) []*PKServiceDescriptor {
	svcs := make([]*PKServiceDescriptor, len(protos))
	file, _ := FileDescriptorFromContext(ctx)

	for i, sd := range protos {
		longName := sd.GetName()
		commentPath := fmt.Sprintf("%d.%d", serviceCommentPath, i)

		svcs[i] = &PKServiceDescriptor{
			common:                 newCommon(file, commentPath, longName),
			ServiceDescriptorProto: sd,
			Comments:               file.comments.Get(commentPath),
			ServiceDescriptor:      file.FileDescriptor.Services().ByName(protoreflect.Name(sd.GetName())),
		}
		if sd.Options != nil {
			svcs[i].setOptions(sd.Options)
		}

		svcCtx := ContextWithServiceDescriptor(ctx, svcs[i])
		svcs[i].Methods = parseServiceMethods(svcCtx, sd.GetMethod())
	}

	return svcs
}

func parseServiceMethods(ctx context.Context, protos []*descriptorpb.MethodDescriptorProto) []*PKMethodDescriptor {
	methods := make([]*PKMethodDescriptor, len(protos))

	file, _ := FileDescriptorFromContext(ctx)
	svc, _ := ServiceDescriptorFromContext(ctx)

	for i, md := range protos {
		longName := fmt.Sprintf("%s.%s", svc.GetLongName(), md.GetName())

		methods[i] = &PKMethodDescriptor{
			common:                newCommon(file, "", longName),
			MethodDescriptorProto: md,
			Service:               svc,
			Comments:              file.comments.Get(fmt.Sprintf("%s.%d.%d", svc.path, serviceMethodCommentPath, i)),
			MethodDescriptor:      svc.ServiceDescriptor.Methods().ByName(protoreflect.Name(md.GetName())),
		}
		if md.Options != nil {
			methods[i].setOptions(md.Options)
		}
	}

	return methods
}
