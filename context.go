package protokit

import (
	"context"
)

type contextKey string

const (
	allFilesContextKey   = contextKey("all_files")
	fileContextKey       = contextKey("file")
	descriptorContextKey = contextKey("descriptor")
	enumContextKey       = contextKey("enum")
	serviceContextKey    = contextKey("service")
)

// ContextWithAllFiles returns a new context with the attached `AllFiles`
func ContextWithAllFiles(ctx context.Context, allFiles map[string]*PKFileDescriptor) context.Context {
	return context.WithValue(ctx, allFilesContextKey, allFiles)
}

// AllFilesFromContext returns the `AllFiles` from the context and whether or not the key was found.
func AllFilesFromContext(ctx context.Context) (map[string]*PKFileDescriptor, bool) {
	val, ok := ctx.Value(allFilesContextKey).(map[string]*PKFileDescriptor)
	return val, ok
}

// ContextWithFileDescriptor returns a new context with the attached `PKFileDescriptor`
func ContextWithFileDescriptor(ctx context.Context, fd *PKFileDescriptor) context.Context {
	return context.WithValue(ctx, fileContextKey, fd)
}

// FileDescriptorFromContext returns the `PKFileDescriptor` from the context and whether or not the key was found.
func FileDescriptorFromContext(ctx context.Context) (*PKFileDescriptor, bool) {
	val, ok := ctx.Value(fileContextKey).(*PKFileDescriptor)
	return val, ok
}

// ContextWithDescriptor returns a new context with the specified `PKDescriptor`
func ContextWithDescriptor(ctx context.Context, d *PKDescriptor) context.Context {
	return context.WithValue(ctx, descriptorContextKey, d)
}

// DescriptorFromContext returns the associated `PKDescriptor` for the context and whether or not it was found
func DescriptorFromContext(ctx context.Context) (*PKDescriptor, bool) {
	val, ok := ctx.Value(descriptorContextKey).(*PKDescriptor)
	return val, ok
}

// ContextWithEnumDescriptor returns a new context with the specified `PKEnumDescriptor`
func ContextWithEnumDescriptor(ctx context.Context, d *PKEnumDescriptor) context.Context {
	return context.WithValue(ctx, enumContextKey, d)
}

// EnumDescriptorFromContext returns the associated `PKEnumDescriptor` for the context and whether or not it was found
func EnumDescriptorFromContext(ctx context.Context) (*PKEnumDescriptor, bool) {
	val, ok := ctx.Value(enumContextKey).(*PKEnumDescriptor)
	return val, ok
}

// ContextWithServiceDescriptor returns a new context with `service`
func ContextWithServiceDescriptor(ctx context.Context, service *PKServiceDescriptor) context.Context {
	return context.WithValue(ctx, serviceContextKey, service)
}

// ServiceDescriptorFromContext returns the `Service` from the context and whether or not the key was found.
func ServiceDescriptorFromContext(ctx context.Context) (*PKServiceDescriptor, bool) {
	val, ok := ctx.Value(serviceContextKey).(*PKServiceDescriptor)
	return val, ok
}
