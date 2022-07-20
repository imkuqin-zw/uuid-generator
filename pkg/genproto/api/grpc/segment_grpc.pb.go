// Code generated by protoc-gen-yggdrasil-grpc. DO NOT EDIT.

package grpcimpl

import (
	context "context"
	api "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

type segmentClient struct {
	cc grpc.ClientConnInterface
}

func NewSegmentClient(cc grpc.ClientConnInterface) api.SegmentClient {
	return &segmentClient{cc}
}

func (c *segmentClient) FetchNext(ctx context.Context, in *api.FetchSegmentNextReq) (*api.UUID, error) {
	out := new(api.UUID)
	err := c.cc.Invoke(ctx, "/com.github.imkuqin_zw.uuid_generator.api.Segment/FetchNext", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Segment_FetchNext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(api.FetchSegmentNextReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(api.SegmentServer).FetchNext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/com.github.imkuqin_zw.uuid_generator.api.Segment/FetchNext",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(api.SegmentServer).FetchNext(ctx, req.(*api.FetchSegmentNextReq))
	}
	return interceptor(ctx, in, info, handler)
}

var SegmentServiceDesc = grpc.ServiceDesc{
	ServiceName: "com.github.imkuqin_zw.uuid_generator.api.Segment",
	HandlerType: (*api.SegmentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchNext",
			Handler:    _Segment_FetchNext_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/segment.proto",
}
