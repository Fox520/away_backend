// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PropertyServiceClient is the client API for PropertyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PropertyServiceClient interface {
	GetMinimalInfoProperties(ctx context.Context, in *GetMinimalPropertiesRequest, opts ...grpc.CallOption) (PropertyService_GetMinimalInfoPropertiesClient, error)
	GetSingleProperty(ctx context.Context, in *GetSinglePropertyRequest, opts ...grpc.CallOption) (*SinglePropertyResponse, error)
	GetUserProperties(ctx context.Context, in *GetUserPropertiesRequest, opts ...grpc.CallOption) (*GetUserPropertiesResponse, error)
	GetMultipleProperties(ctx context.Context, in *GetMultiplePropertyRequest, opts ...grpc.CallOption) (PropertyService_GetMultiplePropertiesClient, error)
	CreateProperty(ctx context.Context, in *CreatePropertyRequest, opts ...grpc.CallOption) (*Property, error)
	UpdateProperty(ctx context.Context, in *Property, opts ...grpc.CallOption) (*Property, error)
	DeleteProperty(ctx context.Context, in *DeletePropertyRequest, opts ...grpc.CallOption) (*DeletePropertyResponse, error)
	GetFeaturedAreas(ctx context.Context, in *FeaturedAreasRequest, opts ...grpc.CallOption) (*FeaturedAreasResponse, error)
	GetPromotedProperties(ctx context.Context, in *PromotedRequest, opts ...grpc.CallOption) (PropertyService_GetPromotedPropertiesClient, error)
	LocationSearch(ctx context.Context, opts ...grpc.CallOption) (PropertyService_LocationSearchClient, error)
}

type propertyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPropertyServiceClient(cc grpc.ClientConnInterface) PropertyServiceClient {
	return &propertyServiceClient{cc}
}

func (c *propertyServiceClient) GetMinimalInfoProperties(ctx context.Context, in *GetMinimalPropertiesRequest, opts ...grpc.CallOption) (PropertyService_GetMinimalInfoPropertiesClient, error) {
	stream, err := c.cc.NewStream(ctx, &PropertyService_ServiceDesc.Streams[0], "/property.service.PropertyService/GetMinimalInfoProperties", opts...)
	if err != nil {
		return nil, err
	}
	x := &propertyServiceGetMinimalInfoPropertiesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PropertyService_GetMinimalInfoPropertiesClient interface {
	Recv() (*GetMinimalPropertiesResponse, error)
	grpc.ClientStream
}

type propertyServiceGetMinimalInfoPropertiesClient struct {
	grpc.ClientStream
}

func (x *propertyServiceGetMinimalInfoPropertiesClient) Recv() (*GetMinimalPropertiesResponse, error) {
	m := new(GetMinimalPropertiesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *propertyServiceClient) GetSingleProperty(ctx context.Context, in *GetSinglePropertyRequest, opts ...grpc.CallOption) (*SinglePropertyResponse, error) {
	out := new(SinglePropertyResponse)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/GetSingleProperty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) GetUserProperties(ctx context.Context, in *GetUserPropertiesRequest, opts ...grpc.CallOption) (*GetUserPropertiesResponse, error) {
	out := new(GetUserPropertiesResponse)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/GetUserProperties", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) GetMultipleProperties(ctx context.Context, in *GetMultiplePropertyRequest, opts ...grpc.CallOption) (PropertyService_GetMultiplePropertiesClient, error) {
	stream, err := c.cc.NewStream(ctx, &PropertyService_ServiceDesc.Streams[1], "/property.service.PropertyService/GetMultipleProperties", opts...)
	if err != nil {
		return nil, err
	}
	x := &propertyServiceGetMultiplePropertiesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PropertyService_GetMultiplePropertiesClient interface {
	Recv() (*GetMultiplePropertyResponse, error)
	grpc.ClientStream
}

type propertyServiceGetMultiplePropertiesClient struct {
	grpc.ClientStream
}

func (x *propertyServiceGetMultiplePropertiesClient) Recv() (*GetMultiplePropertyResponse, error) {
	m := new(GetMultiplePropertyResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *propertyServiceClient) CreateProperty(ctx context.Context, in *CreatePropertyRequest, opts ...grpc.CallOption) (*Property, error) {
	out := new(Property)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/CreateProperty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) UpdateProperty(ctx context.Context, in *Property, opts ...grpc.CallOption) (*Property, error) {
	out := new(Property)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/UpdateProperty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) DeleteProperty(ctx context.Context, in *DeletePropertyRequest, opts ...grpc.CallOption) (*DeletePropertyResponse, error) {
	out := new(DeletePropertyResponse)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/DeleteProperty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) GetFeaturedAreas(ctx context.Context, in *FeaturedAreasRequest, opts ...grpc.CallOption) (*FeaturedAreasResponse, error) {
	out := new(FeaturedAreasResponse)
	err := c.cc.Invoke(ctx, "/property.service.PropertyService/GetFeaturedAreas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *propertyServiceClient) GetPromotedProperties(ctx context.Context, in *PromotedRequest, opts ...grpc.CallOption) (PropertyService_GetPromotedPropertiesClient, error) {
	stream, err := c.cc.NewStream(ctx, &PropertyService_ServiceDesc.Streams[2], "/property.service.PropertyService/GetPromotedProperties", opts...)
	if err != nil {
		return nil, err
	}
	x := &propertyServiceGetPromotedPropertiesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PropertyService_GetPromotedPropertiesClient interface {
	Recv() (*PromotedResponse, error)
	grpc.ClientStream
}

type propertyServiceGetPromotedPropertiesClient struct {
	grpc.ClientStream
}

func (x *propertyServiceGetPromotedPropertiesClient) Recv() (*PromotedResponse, error) {
	m := new(PromotedResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *propertyServiceClient) LocationSearch(ctx context.Context, opts ...grpc.CallOption) (PropertyService_LocationSearchClient, error) {
	stream, err := c.cc.NewStream(ctx, &PropertyService_ServiceDesc.Streams[3], "/property.service.PropertyService/LocationSearch", opts...)
	if err != nil {
		return nil, err
	}
	x := &propertyServiceLocationSearchClient{stream}
	return x, nil
}

type PropertyService_LocationSearchClient interface {
	Send(*LocationSearchRequest) error
	Recv() (*LocationSearchResponse, error)
	grpc.ClientStream
}

type propertyServiceLocationSearchClient struct {
	grpc.ClientStream
}

func (x *propertyServiceLocationSearchClient) Send(m *LocationSearchRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *propertyServiceLocationSearchClient) Recv() (*LocationSearchResponse, error) {
	m := new(LocationSearchResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PropertyServiceServer is the server API for PropertyService service.
// All implementations must embed UnimplementedPropertyServiceServer
// for forward compatibility
type PropertyServiceServer interface {
	GetMinimalInfoProperties(*GetMinimalPropertiesRequest, PropertyService_GetMinimalInfoPropertiesServer) error
	GetSingleProperty(context.Context, *GetSinglePropertyRequest) (*SinglePropertyResponse, error)
	GetUserProperties(context.Context, *GetUserPropertiesRequest) (*GetUserPropertiesResponse, error)
	GetMultipleProperties(*GetMultiplePropertyRequest, PropertyService_GetMultiplePropertiesServer) error
	CreateProperty(context.Context, *CreatePropertyRequest) (*Property, error)
	UpdateProperty(context.Context, *Property) (*Property, error)
	DeleteProperty(context.Context, *DeletePropertyRequest) (*DeletePropertyResponse, error)
	GetFeaturedAreas(context.Context, *FeaturedAreasRequest) (*FeaturedAreasResponse, error)
	GetPromotedProperties(*PromotedRequest, PropertyService_GetPromotedPropertiesServer) error
	LocationSearch(PropertyService_LocationSearchServer) error
	mustEmbedUnimplementedPropertyServiceServer()
}

// UnimplementedPropertyServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPropertyServiceServer struct {
}

func (UnimplementedPropertyServiceServer) GetMinimalInfoProperties(*GetMinimalPropertiesRequest, PropertyService_GetMinimalInfoPropertiesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetMinimalInfoProperties not implemented")
}
func (UnimplementedPropertyServiceServer) GetSingleProperty(context.Context, *GetSinglePropertyRequest) (*SinglePropertyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSingleProperty not implemented")
}
func (UnimplementedPropertyServiceServer) GetUserProperties(context.Context, *GetUserPropertiesRequest) (*GetUserPropertiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserProperties not implemented")
}
func (UnimplementedPropertyServiceServer) GetMultipleProperties(*GetMultiplePropertyRequest, PropertyService_GetMultiplePropertiesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetMultipleProperties not implemented")
}
func (UnimplementedPropertyServiceServer) CreateProperty(context.Context, *CreatePropertyRequest) (*Property, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProperty not implemented")
}
func (UnimplementedPropertyServiceServer) UpdateProperty(context.Context, *Property) (*Property, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProperty not implemented")
}
func (UnimplementedPropertyServiceServer) DeleteProperty(context.Context, *DeletePropertyRequest) (*DeletePropertyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteProperty not implemented")
}
func (UnimplementedPropertyServiceServer) GetFeaturedAreas(context.Context, *FeaturedAreasRequest) (*FeaturedAreasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFeaturedAreas not implemented")
}
func (UnimplementedPropertyServiceServer) GetPromotedProperties(*PromotedRequest, PropertyService_GetPromotedPropertiesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetPromotedProperties not implemented")
}
func (UnimplementedPropertyServiceServer) LocationSearch(PropertyService_LocationSearchServer) error {
	return status.Errorf(codes.Unimplemented, "method LocationSearch not implemented")
}
func (UnimplementedPropertyServiceServer) mustEmbedUnimplementedPropertyServiceServer() {}

// UnsafePropertyServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PropertyServiceServer will
// result in compilation errors.
type UnsafePropertyServiceServer interface {
	mustEmbedUnimplementedPropertyServiceServer()
}

func RegisterPropertyServiceServer(s grpc.ServiceRegistrar, srv PropertyServiceServer) {
	s.RegisterService(&PropertyService_ServiceDesc, srv)
}

func _PropertyService_GetMinimalInfoProperties_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetMinimalPropertiesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PropertyServiceServer).GetMinimalInfoProperties(m, &propertyServiceGetMinimalInfoPropertiesServer{stream})
}

type PropertyService_GetMinimalInfoPropertiesServer interface {
	Send(*GetMinimalPropertiesResponse) error
	grpc.ServerStream
}

type propertyServiceGetMinimalInfoPropertiesServer struct {
	grpc.ServerStream
}

func (x *propertyServiceGetMinimalInfoPropertiesServer) Send(m *GetMinimalPropertiesResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PropertyService_GetSingleProperty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSinglePropertyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).GetSingleProperty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/GetSingleProperty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).GetSingleProperty(ctx, req.(*GetSinglePropertyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_GetUserProperties_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserPropertiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).GetUserProperties(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/GetUserProperties",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).GetUserProperties(ctx, req.(*GetUserPropertiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_GetMultipleProperties_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetMultiplePropertyRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PropertyServiceServer).GetMultipleProperties(m, &propertyServiceGetMultiplePropertiesServer{stream})
}

type PropertyService_GetMultiplePropertiesServer interface {
	Send(*GetMultiplePropertyResponse) error
	grpc.ServerStream
}

type propertyServiceGetMultiplePropertiesServer struct {
	grpc.ServerStream
}

func (x *propertyServiceGetMultiplePropertiesServer) Send(m *GetMultiplePropertyResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PropertyService_CreateProperty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePropertyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).CreateProperty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/CreateProperty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).CreateProperty(ctx, req.(*CreatePropertyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_UpdateProperty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Property)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).UpdateProperty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/UpdateProperty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).UpdateProperty(ctx, req.(*Property))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_DeleteProperty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePropertyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).DeleteProperty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/DeleteProperty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).DeleteProperty(ctx, req.(*DeletePropertyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_GetFeaturedAreas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FeaturedAreasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PropertyServiceServer).GetFeaturedAreas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/property.service.PropertyService/GetFeaturedAreas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PropertyServiceServer).GetFeaturedAreas(ctx, req.(*FeaturedAreasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PropertyService_GetPromotedProperties_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PromotedRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PropertyServiceServer).GetPromotedProperties(m, &propertyServiceGetPromotedPropertiesServer{stream})
}

type PropertyService_GetPromotedPropertiesServer interface {
	Send(*PromotedResponse) error
	grpc.ServerStream
}

type propertyServiceGetPromotedPropertiesServer struct {
	grpc.ServerStream
}

func (x *propertyServiceGetPromotedPropertiesServer) Send(m *PromotedResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PropertyService_LocationSearch_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PropertyServiceServer).LocationSearch(&propertyServiceLocationSearchServer{stream})
}

type PropertyService_LocationSearchServer interface {
	Send(*LocationSearchResponse) error
	Recv() (*LocationSearchRequest, error)
	grpc.ServerStream
}

type propertyServiceLocationSearchServer struct {
	grpc.ServerStream
}

func (x *propertyServiceLocationSearchServer) Send(m *LocationSearchResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *propertyServiceLocationSearchServer) Recv() (*LocationSearchRequest, error) {
	m := new(LocationSearchRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PropertyService_ServiceDesc is the grpc.ServiceDesc for PropertyService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PropertyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "property.service.PropertyService",
	HandlerType: (*PropertyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSingleProperty",
			Handler:    _PropertyService_GetSingleProperty_Handler,
		},
		{
			MethodName: "GetUserProperties",
			Handler:    _PropertyService_GetUserProperties_Handler,
		},
		{
			MethodName: "CreateProperty",
			Handler:    _PropertyService_CreateProperty_Handler,
		},
		{
			MethodName: "UpdateProperty",
			Handler:    _PropertyService_UpdateProperty_Handler,
		},
		{
			MethodName: "DeleteProperty",
			Handler:    _PropertyService_DeleteProperty_Handler,
		},
		{
			MethodName: "GetFeaturedAreas",
			Handler:    _PropertyService_GetFeaturedAreas_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetMinimalInfoProperties",
			Handler:       _PropertyService_GetMinimalInfoProperties_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetMultipleProperties",
			Handler:       _PropertyService_GetMultipleProperties_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetPromotedProperties",
			Handler:       _PropertyService_GetPromotedProperties_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "LocationSearch",
			Handler:       _PropertyService_LocationSearch_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "property_service.proto",
}
