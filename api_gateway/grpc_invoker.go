package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GRPCInvoker struct {
	serviceDescriptors map[string]*ServiceDescriptor
	files              map[string]protoreflect.FileDescriptor
}

type ServiceDescriptor struct {
	serviceName string
	methods     map[string]*MethodDescriptor
}

type MethodDescriptor struct {
	methodName       string
	inputDescriptor  protoreflect.MessageDescriptor
	outputDescriptor protoreflect.MessageDescriptor
	fullMethodName   string
}

func NewGRPCInvoker() *GRPCInvoker {
	return &GRPCInvoker{
		serviceDescriptors: make(map[string]*ServiceDescriptor),
		files:              make(map[string]protoreflect.FileDescriptor),
	}
}

// LoadProtoset loads service definitions from protoset files
func (g *GRPCInvoker) LoadProtoset(protosetPath string) error {

	data, err := os.ReadFile(protosetPath)
	if err != nil {
		return fmt.Errorf("failed to read protoset file %s: %w", protosetPath, err)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return fmt.Errorf("failed to unmarshal protoset: %w", err)
	}

	files := &fakeFiles{
		filesByPath: make(map[string]protoreflect.FileDescriptor),
	}

	for _, fdProto := range fds.File {

		fd, err := protodesc.NewFile(fdProto, files)
		if err != nil {
			log.Printf("failed to create file descriptor for %s: %w", fdProto.GetName(), err)
			continue
		}

		files.filesByPath[fd.Path()] = fd
		g.files[fd.Path()] = fd

		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			svc := services.Get(i)
			if err := g.registerService(svc); err != nil {
				log.Printf("Warning: failed to register service %s: %v", svc.FullName(), err)
			}
		}
	}

	log.Printf("Successfully loaded protoset from %s", protosetPath)
	return nil
}

func (g *GRPCInvoker) registerService(svc protoreflect.ServiceDescriptor) error {
	serviceName := string(svc.FullName())

	sd := &ServiceDescriptor{
		serviceName: serviceName,
		methods:     make(map[string]*MethodDescriptor),
	}

	methods := svc.Methods()
	for j := 0; j < methods.Len(); j++ {
		method := methods.Get(j)
		methodName := string(method.Name())

		inputMsg := method.Input()
		outputMsg := method.Output()

		md := &MethodDescriptor{
			methodName:       methodName,
			inputDescriptor:  inputMsg,
			outputDescriptor: outputMsg,
			fullMethodName:   fmt.Sprintf("/%s/%s", serviceName, methodName),
		}

		sd.methods[methodName] = md
		log.Printf("Registered method: %s", md.fullMethodName)
	}

	g.serviceDescriptors[serviceName] = sd
	return nil
}

func (g *GRPCInvoker) Invoke(ctx context.Context, conn *grpc.ClientConn, serviceName string, methodName string, requestJSON []byte) ([]byte, error) {

	sd, exists := g.serviceDescriptors[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	md, exists := sd.methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
	}

	reqMsg := dynamicpb.NewMessage(md.inputDescriptor)

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	log.Println("Message: ", reqMsg.String())
	if err := unmarshaler.Unmarshal(requestJSON, reqMsg); err != nil {
		log.Println("MARSHAL PROBLEM IN GRPC INVOKER: ", reqMsg.String())
		log.Println(err.Error())
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	respMsg := dynamicpb.NewMessage(md.outputDescriptor)

	// Invoke the method
	err := conn.Invoke(ctx, md.fullMethodName, reqMsg, respMsg)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("failed to invoke gRPC method: %w", err)
	}

	// Marshal response to JSON
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}
	responseJSON, err := marshaler.Marshal(respMsg)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return responseJSON, nil
}

type fakeFiles struct {
	filesByPath map[string]protoreflect.FileDescriptor
}

func (f *fakeFiles) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if fd, ok := f.filesByPath[path]; ok {
		return fd, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (f *fakeFiles) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {

	for _, fd := range f.filesByPath {
		// Try to find in messages
		if desc := fd.Messages().ByName(protoreflect.Name(name)); desc != nil {
			return desc, nil
		}
		// Try to find in services
		if desc := fd.Services().ByName(protoreflect.Name(name)); desc != nil {
			return desc, nil
		}
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}

func (f *fakeFiles) RangeFiles(fn func(protoreflect.FileDescriptor) bool) {
	for _, fd := range f.filesByPath {
		if !fn(fd) {
			return
		}
	}
}

func (f *fakeFiles) NumFiles() int {
	return len(f.filesByPath)
}

func (f *fakeFiles) RegisterFile(fd protoreflect.FileDescriptor) error {
	f.filesByPath[fd.Path()] = fd
	return nil
}

// func (g *GRPCInvoker) close() {

// }
