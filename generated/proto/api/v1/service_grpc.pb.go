// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api_v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// VersionServiceClient is the client API for VersionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VersionServiceClient interface {
	GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Version, error)
}

type versionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVersionServiceClient(cc grpc.ClientConnInterface) VersionServiceClient {
	return &versionServiceClient{cc}
}

func (c *versionServiceClient) GetVersion(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*Version, error) {
	out := new(Version)
	err := c.cc.Invoke(ctx, "/api.v1.VersionService/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VersionServiceServer is the server API for VersionService service.
// All implementations should embed UnimplementedVersionServiceServer
// for forward compatibility
type VersionServiceServer interface {
	GetVersion(context.Context, *emptypb.Empty) (*Version, error)
}

// UnimplementedVersionServiceServer should be embedded to have forward compatible implementations.
type UnimplementedVersionServiceServer struct {
}

func (UnimplementedVersionServiceServer) GetVersion(context.Context, *emptypb.Empty) (*Version, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVersion not implemented")
}

// UnsafeVersionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VersionServiceServer will
// result in compilation errors.
type UnsafeVersionServiceServer interface {
	mustEmbedUnimplementedVersionServiceServer()
}

func RegisterVersionServiceServer(s grpc.ServiceRegistrar, srv VersionServiceServer) {
	s.RegisterService(&VersionService_ServiceDesc, srv)
}

func _VersionService_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VersionServiceServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.VersionService/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VersionServiceServer).GetVersion(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// VersionService_ServiceDesc is the grpc.ServiceDesc for VersionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VersionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.VersionService",
	HandlerType: (*VersionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVersion",
			Handler:    _VersionService_GetVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/api/v1/service.proto",
}

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	// Whoami provides information about the currently authenticated principal.
	Whoami(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*WhoamiResponse, error)
	// CreateToken generates an arbitrary service account token
	CreateToken(ctx context.Context, in *ServiceAccount, opts ...grpc.CallOption) (*TokenResponse, error)
	// Token generates a service account token for the current user.
	Token(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TokenResponse, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) Whoami(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*WhoamiResponse, error) {
	out := new(WhoamiResponse)
	err := c.cc.Invoke(ctx, "/api.v1.UserService/Whoami", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CreateToken(ctx context.Context, in *ServiceAccount, opts ...grpc.CallOption) (*TokenResponse, error) {
	out := new(TokenResponse)
	err := c.cc.Invoke(ctx, "/api.v1.UserService/CreateToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Token(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TokenResponse, error) {
	out := new(TokenResponse)
	err := c.cc.Invoke(ctx, "/api.v1.UserService/Token", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations should embed UnimplementedUserServiceServer
// for forward compatibility
type UserServiceServer interface {
	// Whoami provides information about the currently authenticated principal.
	Whoami(context.Context, *emptypb.Empty) (*WhoamiResponse, error)
	// CreateToken generates an arbitrary service account token
	CreateToken(context.Context, *ServiceAccount) (*TokenResponse, error)
	// Token generates a service account token for the current user.
	Token(context.Context, *emptypb.Empty) (*TokenResponse, error)
}

// UnimplementedUserServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUserServiceServer struct {
}

func (UnimplementedUserServiceServer) Whoami(context.Context, *emptypb.Empty) (*WhoamiResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Whoami not implemented")
}
func (UnimplementedUserServiceServer) CreateToken(context.Context, *ServiceAccount) (*TokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateToken not implemented")
}
func (UnimplementedUserServiceServer) Token(context.Context, *emptypb.Empty) (*TokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Token not implemented")
}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_Whoami_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Whoami(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.UserService/Whoami",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Whoami(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CreateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CreateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.UserService/CreateToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CreateToken(ctx, req.(*ServiceAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Token_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Token(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.UserService/Token",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Token(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Whoami",
			Handler:    _UserService_Whoami_Handler,
		},
		{
			MethodName: "CreateToken",
			Handler:    _UserService_CreateToken_Handler,
		},
		{
			MethodName: "Token",
			Handler:    _UserService_Token_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/api/v1/service.proto",
}

// FlavorServiceClient is the client API for FlavorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FlavorServiceClient interface {
	// List provides information about the available flavors.
	List(ctx context.Context, in *FlavorListRequest, opts ...grpc.CallOption) (*FlavorListResponse, error)
	// Info provides information about a specific flavor.
	Info(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*Flavor, error)
}

type flavorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFlavorServiceClient(cc grpc.ClientConnInterface) FlavorServiceClient {
	return &flavorServiceClient{cc}
}

func (c *flavorServiceClient) List(ctx context.Context, in *FlavorListRequest, opts ...grpc.CallOption) (*FlavorListResponse, error) {
	out := new(FlavorListResponse)
	err := c.cc.Invoke(ctx, "/api.v1.FlavorService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *flavorServiceClient) Info(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*Flavor, error) {
	out := new(Flavor)
	err := c.cc.Invoke(ctx, "/api.v1.FlavorService/Info", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FlavorServiceServer is the server API for FlavorService service.
// All implementations should embed UnimplementedFlavorServiceServer
// for forward compatibility
type FlavorServiceServer interface {
	// List provides information about the available flavors.
	List(context.Context, *FlavorListRequest) (*FlavorListResponse, error)
	// Info provides information about a specific flavor.
	Info(context.Context, *ResourceByID) (*Flavor, error)
}

// UnimplementedFlavorServiceServer should be embedded to have forward compatible implementations.
type UnimplementedFlavorServiceServer struct {
}

func (UnimplementedFlavorServiceServer) List(context.Context, *FlavorListRequest) (*FlavorListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedFlavorServiceServer) Info(context.Context, *ResourceByID) (*Flavor, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}

// UnsafeFlavorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FlavorServiceServer will
// result in compilation errors.
type UnsafeFlavorServiceServer interface {
	mustEmbedUnimplementedFlavorServiceServer()
}

func RegisterFlavorServiceServer(s grpc.ServiceRegistrar, srv FlavorServiceServer) {
	s.RegisterService(&FlavorService_ServiceDesc, srv)
}

func _FlavorService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FlavorListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FlavorServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.FlavorService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FlavorServiceServer).List(ctx, req.(*FlavorListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FlavorService_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceByID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FlavorServiceServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.FlavorService/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FlavorServiceServer).Info(ctx, req.(*ResourceByID))
	}
	return interceptor(ctx, in, info, handler)
}

// FlavorService_ServiceDesc is the grpc.ServiceDesc for FlavorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FlavorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.FlavorService",
	HandlerType: (*FlavorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _FlavorService_List_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _FlavorService_Info_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/api/v1/service.proto",
}

// ClusterServiceClient is the client API for ClusterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClusterServiceClient interface {
	// Info provides information about a specific cluster.
	Info(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*Cluster, error)
	// List provides information about the available clusters.
	List(ctx context.Context, in *ClusterListRequest, opts ...grpc.CallOption) (*ClusterListResponse, error)
	// Lifespan updates the lifespan for a specific cluster.
	Lifespan(ctx context.Context, in *LifespanRequest, opts ...grpc.CallOption) (*durationpb.Duration, error)
	// Create launches a new cluster.
	Create(ctx context.Context, in *CreateClusterRequest, opts ...grpc.CallOption) (*ResourceByID, error)
	// Artifacts returns the artifacts for a specific cluster.
	Artifacts(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*ClusterArtifacts, error)
	// Delete deletes an existing cluster.
	Delete(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Logs returns the logs for a specific cluster.
	Logs(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*LogsResponse, error)
}

type clusterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClusterServiceClient(cc grpc.ClientConnInterface) ClusterServiceClient {
	return &clusterServiceClient{cc}
}

func (c *clusterServiceClient) Info(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*Cluster, error) {
	out := new(Cluster)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Info", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) List(ctx context.Context, in *ClusterListRequest, opts ...grpc.CallOption) (*ClusterListResponse, error) {
	out := new(ClusterListResponse)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) Lifespan(ctx context.Context, in *LifespanRequest, opts ...grpc.CallOption) (*durationpb.Duration, error) {
	out := new(durationpb.Duration)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Lifespan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) Create(ctx context.Context, in *CreateClusterRequest, opts ...grpc.CallOption) (*ResourceByID, error) {
	out := new(ResourceByID)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) Artifacts(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*ClusterArtifacts, error) {
	out := new(ClusterArtifacts)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Artifacts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) Delete(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) Logs(ctx context.Context, in *ResourceByID, opts ...grpc.CallOption) (*LogsResponse, error) {
	out := new(LogsResponse)
	err := c.cc.Invoke(ctx, "/api.v1.ClusterService/Logs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClusterServiceServer is the server API for ClusterService service.
// All implementations should embed UnimplementedClusterServiceServer
// for forward compatibility
type ClusterServiceServer interface {
	// Info provides information about a specific cluster.
	Info(context.Context, *ResourceByID) (*Cluster, error)
	// List provides information about the available clusters.
	List(context.Context, *ClusterListRequest) (*ClusterListResponse, error)
	// Lifespan updates the lifespan for a specific cluster.
	Lifespan(context.Context, *LifespanRequest) (*durationpb.Duration, error)
	// Create launches a new cluster.
	Create(context.Context, *CreateClusterRequest) (*ResourceByID, error)
	// Artifacts returns the artifacts for a specific cluster.
	Artifacts(context.Context, *ResourceByID) (*ClusterArtifacts, error)
	// Delete deletes an existing cluster.
	Delete(context.Context, *ResourceByID) (*emptypb.Empty, error)
	// Logs returns the logs for a specific cluster.
	Logs(context.Context, *ResourceByID) (*LogsResponse, error)
}

// UnimplementedClusterServiceServer should be embedded to have forward compatible implementations.
type UnimplementedClusterServiceServer struct {
}

func (UnimplementedClusterServiceServer) Info(context.Context, *ResourceByID) (*Cluster, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}
func (UnimplementedClusterServiceServer) List(context.Context, *ClusterListRequest) (*ClusterListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedClusterServiceServer) Lifespan(context.Context, *LifespanRequest) (*durationpb.Duration, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lifespan not implemented")
}
func (UnimplementedClusterServiceServer) Create(context.Context, *CreateClusterRequest) (*ResourceByID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedClusterServiceServer) Artifacts(context.Context, *ResourceByID) (*ClusterArtifacts, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Artifacts not implemented")
}
func (UnimplementedClusterServiceServer) Delete(context.Context, *ResourceByID) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedClusterServiceServer) Logs(context.Context, *ResourceByID) (*LogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logs not implemented")
}

// UnsafeClusterServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClusterServiceServer will
// result in compilation errors.
type UnsafeClusterServiceServer interface {
	mustEmbedUnimplementedClusterServiceServer()
}

func RegisterClusterServiceServer(s grpc.ServiceRegistrar, srv ClusterServiceServer) {
	s.RegisterService(&ClusterService_ServiceDesc, srv)
}

func _ClusterService_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceByID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Info",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Info(ctx, req.(*ResourceByID))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClusterListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).List(ctx, req.(*ClusterListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_Lifespan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LifespanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Lifespan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Lifespan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Lifespan(ctx, req.(*LifespanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateClusterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Create(ctx, req.(*CreateClusterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_Artifacts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceByID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Artifacts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Artifacts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Artifacts(ctx, req.(*ResourceByID))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceByID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Delete(ctx, req.(*ResourceByID))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClusterService_Logs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceByID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClusterServiceServer).Logs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.ClusterService/Logs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClusterServiceServer).Logs(ctx, req.(*ResourceByID))
	}
	return interceptor(ctx, in, info, handler)
}

// ClusterService_ServiceDesc is the grpc.ServiceDesc for ClusterService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClusterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.ClusterService",
	HandlerType: (*ClusterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Info",
			Handler:    _ClusterService_Info_Handler,
		},
		{
			MethodName: "List",
			Handler:    _ClusterService_List_Handler,
		},
		{
			MethodName: "Lifespan",
			Handler:    _ClusterService_Lifespan_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _ClusterService_Create_Handler,
		},
		{
			MethodName: "Artifacts",
			Handler:    _ClusterService_Artifacts_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _ClusterService_Delete_Handler,
		},
		{
			MethodName: "Logs",
			Handler:    _ClusterService_Logs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/api/v1/service.proto",
}

// CliServiceClient is the client API for CliService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CliServiceClient interface {
	// Upgrade - gets an updated binary if it exists.
	Upgrade(ctx context.Context, in *CliUpgradeRequest, opts ...grpc.CallOption) (CliService_UpgradeClient, error)
}

type cliServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCliServiceClient(cc grpc.ClientConnInterface) CliServiceClient {
	return &cliServiceClient{cc}
}

func (c *cliServiceClient) Upgrade(ctx context.Context, in *CliUpgradeRequest, opts ...grpc.CallOption) (CliService_UpgradeClient, error) {
	stream, err := c.cc.NewStream(ctx, &CliService_ServiceDesc.Streams[0], "/api.v1.CliService/Upgrade", opts...)
	if err != nil {
		return nil, err
	}
	x := &cliServiceUpgradeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CliService_UpgradeClient interface {
	Recv() (*CliUpgradeResponse, error)
	grpc.ClientStream
}

type cliServiceUpgradeClient struct {
	grpc.ClientStream
}

func (x *cliServiceUpgradeClient) Recv() (*CliUpgradeResponse, error) {
	m := new(CliUpgradeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CliServiceServer is the server API for CliService service.
// All implementations should embed UnimplementedCliServiceServer
// for forward compatibility
type CliServiceServer interface {
	// Upgrade - gets an updated binary if it exists.
	Upgrade(*CliUpgradeRequest, CliService_UpgradeServer) error
}

// UnimplementedCliServiceServer should be embedded to have forward compatible implementations.
type UnimplementedCliServiceServer struct {
}

func (UnimplementedCliServiceServer) Upgrade(*CliUpgradeRequest, CliService_UpgradeServer) error {
	return status.Errorf(codes.Unimplemented, "method Upgrade not implemented")
}

// UnsafeCliServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CliServiceServer will
// result in compilation errors.
type UnsafeCliServiceServer interface {
	mustEmbedUnimplementedCliServiceServer()
}

func RegisterCliServiceServer(s grpc.ServiceRegistrar, srv CliServiceServer) {
	s.RegisterService(&CliService_ServiceDesc, srv)
}

func _CliService_Upgrade_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CliUpgradeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CliServiceServer).Upgrade(m, &cliServiceUpgradeServer{stream})
}

type CliService_UpgradeServer interface {
	Send(*CliUpgradeResponse) error
	grpc.ServerStream
}

type cliServiceUpgradeServer struct {
	grpc.ServerStream
}

func (x *cliServiceUpgradeServer) Send(m *CliUpgradeResponse) error {
	return x.ServerStream.SendMsg(m)
}

// CliService_ServiceDesc is the grpc.ServiceDesc for CliService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CliService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.CliService",
	HandlerType: (*CliServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Upgrade",
			Handler:       _CliService_Upgrade_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/api/v1/service.proto",
}

// InfraStatusServiceClient is the client API for InfraStatusService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InfraStatusServiceClient interface {
	// GetStatus gets the maintenance
	GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*InfraStatus, error)
	// ResetStatus resets the maintenance
	ResetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*InfraStatus, error)
	// SetStatus sets the maintenance
	SetStatus(ctx context.Context, in *InfraStatus, opts ...grpc.CallOption) (*InfraStatus, error)
}

type infraStatusServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInfraStatusServiceClient(cc grpc.ClientConnInterface) InfraStatusServiceClient {
	return &infraStatusServiceClient{cc}
}

func (c *infraStatusServiceClient) GetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*InfraStatus, error) {
	out := new(InfraStatus)
	err := c.cc.Invoke(ctx, "/api.v1.InfraStatusService/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *infraStatusServiceClient) ResetStatus(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*InfraStatus, error) {
	out := new(InfraStatus)
	err := c.cc.Invoke(ctx, "/api.v1.InfraStatusService/ResetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *infraStatusServiceClient) SetStatus(ctx context.Context, in *InfraStatus, opts ...grpc.CallOption) (*InfraStatus, error) {
	out := new(InfraStatus)
	err := c.cc.Invoke(ctx, "/api.v1.InfraStatusService/SetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InfraStatusServiceServer is the server API for InfraStatusService service.
// All implementations should embed UnimplementedInfraStatusServiceServer
// for forward compatibility
type InfraStatusServiceServer interface {
	// GetStatus gets the maintenance
	GetStatus(context.Context, *emptypb.Empty) (*InfraStatus, error)
	// ResetStatus resets the maintenance
	ResetStatus(context.Context, *emptypb.Empty) (*InfraStatus, error)
	// SetStatus sets the maintenance
	SetStatus(context.Context, *InfraStatus) (*InfraStatus, error)
}

// UnimplementedInfraStatusServiceServer should be embedded to have forward compatible implementations.
type UnimplementedInfraStatusServiceServer struct {
}

func (UnimplementedInfraStatusServiceServer) GetStatus(context.Context, *emptypb.Empty) (*InfraStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedInfraStatusServiceServer) ResetStatus(context.Context, *emptypb.Empty) (*InfraStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetStatus not implemented")
}
func (UnimplementedInfraStatusServiceServer) SetStatus(context.Context, *InfraStatus) (*InfraStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetStatus not implemented")
}

// UnsafeInfraStatusServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InfraStatusServiceServer will
// result in compilation errors.
type UnsafeInfraStatusServiceServer interface {
	mustEmbedUnimplementedInfraStatusServiceServer()
}

func RegisterInfraStatusServiceServer(s grpc.ServiceRegistrar, srv InfraStatusServiceServer) {
	s.RegisterService(&InfraStatusService_ServiceDesc, srv)
}

func _InfraStatusService_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InfraStatusServiceServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.InfraStatusService/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InfraStatusServiceServer).GetStatus(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _InfraStatusService_ResetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InfraStatusServiceServer).ResetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.InfraStatusService/ResetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InfraStatusServiceServer).ResetStatus(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _InfraStatusService_SetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InfraStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InfraStatusServiceServer).SetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.v1.InfraStatusService/SetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InfraStatusServiceServer).SetStatus(ctx, req.(*InfraStatus))
	}
	return interceptor(ctx, in, info, handler)
}

// InfraStatusService_ServiceDesc is the grpc.ServiceDesc for InfraStatusService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var InfraStatusService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.v1.InfraStatusService",
	HandlerType: (*InfraStatusServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStatus",
			Handler:    _InfraStatusService_GetStatus_Handler,
		},
		{
			MethodName: "ResetStatus",
			Handler:    _InfraStatusService_ResetStatus_Handler,
		},
		{
			MethodName: "SetStatus",
			Handler:    _InfraStatusService_SetStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/api/v1/service.proto",
}