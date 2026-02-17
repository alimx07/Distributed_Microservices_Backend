package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alimx07/Distributed_Microservices_Backend/services/api_gateway/models"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCInvoker handles dynamic gRPC invocation and HTTP route mapping
type GRPCInvoker struct {
	serviceDescriptors map[string]*ServiceDescriptor
	files              map[string]protoreflect.FileDescriptor
	httpRoutes         map[string]map[string]*models.RouteConfig // method -> path pattern -> RouteConfig
	routeOptions       map[string]*models.RouteOption
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

func NewGRPCInvoker(routeOptions map[string]*models.RouteOption) *GRPCInvoker {
	return &GRPCInvoker{
		serviceDescriptors: make(map[string]*ServiceDescriptor),
		files:              make(map[string]protoreflect.FileDescriptor),
		httpRoutes:         make(map[string]map[string]*models.RouteConfig),
		routeOptions:       routeOptions,
	}
}

// LoadProtoset loads service definitions and HTTP annotations from protoset files
func (g *GRPCInvoker) LoadProtoset(protosetPath, serviceName string) error {
	data, err := os.ReadFile(protosetPath)
	if err != nil {
		return fmt.Errorf("failed to read protoset file %s: %w", protosetPath, err)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return fmt.Errorf("failed to unmarshal protoset: %w", err)
	}

	resolver := &fileResolver{filesByPath: make(map[string]protoreflect.FileDescriptor)}

	for _, fdProto := range fds.File {
		fd, err := protodesc.NewFile(fdProto, resolver)
		if err != nil {
			log.Printf("Skipping file %s: %v", fdProto.GetName(), err)
			continue
		}

		resolver.filesByPath[fd.Path()] = fd
		g.files[fd.Path()] = fd

		// Process each service in the file
		for i := 0; i < fd.Services().Len(); i++ {
			svc := fd.Services().Get(i)
			g.registerService(svc)
			g.registerHttpRoutes(fdProto, svc, serviceName)
		}
	}

	log.Printf("Loaded protoset: %s", protosetPath)
	return nil
}

// registerService registers gRPC service methods for invocation
func (g *GRPCInvoker) registerService(svc protoreflect.ServiceDescriptor) {
	grpcServiceName := string(svc.FullName())

	sd := &ServiceDescriptor{
		serviceName: grpcServiceName,
		methods:     make(map[string]*MethodDescriptor),
	}

	for i := 0; i < svc.Methods().Len(); i++ {
		method := svc.Methods().Get(i)
		methodName := string(method.Name())

		sd.methods[methodName] = &MethodDescriptor{
			methodName:       methodName,
			inputDescriptor:  method.Input(),
			outputDescriptor: method.Output(),
			fullMethodName:   fmt.Sprintf("/%s/%s", grpcServiceName, methodName),
		}
		log.Printf("Registered gRPC method: /%s/%s", grpcServiceName, methodName)
	}

	g.serviceDescriptors[grpcServiceName] = sd
}

// registerHttpRoutes parses google.api.http annotations and registers HTTP routes
func (g *GRPCInvoker) registerHttpRoutes(fdProto *descriptorpb.FileDescriptorProto, svc protoreflect.ServiceDescriptor, serviceName string) {
	grpcServiceName := string(svc.FullName())
	simpleServiceName := string(svc.Name())

	for _, svcProto := range fdProto.GetService() {
		if svcProto.GetName() != simpleServiceName {
			continue
		}

		for _, methodProto := range svcProto.GetMethod() {
			opts := methodProto.GetOptions()
			if opts == nil {
				continue
			}

			// Use typed API to extract google.api.http annotation
			httpRule := g.getHttpRule(opts)
			if httpRule == nil {
				continue
			}

			httpMethod, httpPath := extractMethodAndPath(httpRule)
			if httpMethod == "" || httpPath == "" {
				continue
			}

			route := &models.RouteConfig{
				Path:   httpPath,
				Method: httpMethod,
				Body:   httpRule.Body,
				// Service:     serviceName,
				GRPCService: grpcServiceName,
				GRPCMethod:  methodProto.GetName(),
			}

			if g.httpRoutes[httpMethod] == nil {
				g.httpRoutes[httpMethod] = make(map[string]*models.RouteConfig)
			}
			if val, ok := g.routeOptions[route.Path]; ok {
				route.RateLimitEnabled = val.RateLimitEnabled
				route.RequireAuth = val.RequireAuth
			}
			g.httpRoutes[httpMethod][httpPath] = route

			log.Printf("Registered HTTP route: %s %s -> %s/%s", httpMethod, httpPath, grpcServiceName, methodProto.GetName())
		}
	}
}

// getHttpRule extracts the google.api.http annotation using the typed extension API
func (g *GRPCInvoker) getHttpRule(opts *descriptorpb.MethodOptions) *annotations.HttpRule {
	if !proto.HasExtension(opts, annotations.E_Http) {
		return nil
	}

	ext := proto.GetExtension(opts, annotations.E_Http)
	if httpRule, ok := ext.(*annotations.HttpRule); ok {
		return httpRule
	}
	return nil
}

// extractMethodAndPath extracts HTTP method and path from HttpRule
func extractMethodAndPath(rule *annotations.HttpRule) (method, path string) {
	switch pattern := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		return "GET", pattern.Get
	case *annotations.HttpRule_Post:
		return "POST", pattern.Post
	case *annotations.HttpRule_Delete:
		return "DELETE", pattern.Delete
	case *annotations.HttpRule_Put:
		return "PUT", pattern.Put
	case *annotations.HttpRule_Patch:
		return "PATCH", pattern.Patch
	}
	return "", ""
}

// GetHttpRoutes returns all parsed HTTP routes
func (g *GRPCInvoker) GetHttpRoutes() map[string]map[string]*models.RouteConfig {
	return g.httpRoutes
}

// Invoke calls a gRPC method dynamically
func (g *GRPCInvoker) Invoke(ctx context.Context, conn *grpc.ClientConn, serviceName, methodName string, requestJSON []byte) ([]byte, error) {
	sd, exists := g.serviceDescriptors[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	md, exists := sd.methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
	}

	// Create request message and unmarshal JSON
	reqMsg := dynamicpb.NewMessage(md.inputDescriptor)
	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaler.Unmarshal(requestJSON, reqMsg); err != nil {
		log.Printf("Failed to unmarshal request for %s: %v", md.fullMethodName, err)
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Create response message and invoke
	respMsg := dynamicpb.NewMessage(md.outputDescriptor)
	if err := conn.Invoke(ctx, md.fullMethodName, reqMsg, respMsg); err != nil {
		log.Printf("gRPC call failed for %s: %v", md.fullMethodName, err)
		return nil, fmt.Errorf("failed to invoke gRPC method: %w", err)
	}

	// Marshal response to JSON
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	responseJSON, err := marshaler.Marshal(respMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return responseJSON, nil
}

// MatchPath checks if a request path matches a route pattern with path parameters
func MatchPath(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, pp := range patternParts {
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			continue // Path parameter matches anything
		}
		if pp != pathParts[i] {
			return false
		}
	}
	return true
}

// fileResolver implements protodesc.Resolver for building file descriptors
type fileResolver struct {
	filesByPath map[string]protoreflect.FileDescriptor
}

func (r *fileResolver) FindFileByPath(path string) (protoreflect.FileDescriptor, error) {
	if fd, ok := r.filesByPath[path]; ok {
		return fd, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func (r *fileResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	for _, fd := range r.filesByPath {
		if desc := fd.Messages().ByName(protoreflect.Name(name)); desc != nil {
			return desc, nil
		}
		if desc := fd.Services().ByName(protoreflect.Name(name)); desc != nil {
			return desc, nil
		}
	}
	return nil, fmt.Errorf("descriptor not found: %s", name)
}
