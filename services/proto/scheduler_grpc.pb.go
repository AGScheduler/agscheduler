// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: scheduler.proto

package services

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Scheduler_AddJob_FullMethodName        = "/services.Scheduler/AddJob"
	Scheduler_GetJob_FullMethodName        = "/services.Scheduler/GetJob"
	Scheduler_GetAllJobs_FullMethodName    = "/services.Scheduler/GetAllJobs"
	Scheduler_UpdateJob_FullMethodName     = "/services.Scheduler/UpdateJob"
	Scheduler_DeleteJob_FullMethodName     = "/services.Scheduler/DeleteJob"
	Scheduler_DeleteAllJobs_FullMethodName = "/services.Scheduler/DeleteAllJobs"
	Scheduler_PauseJob_FullMethodName      = "/services.Scheduler/PauseJob"
	Scheduler_ResumeJob_FullMethodName     = "/services.Scheduler/ResumeJob"
	Scheduler_RunJob_FullMethodName        = "/services.Scheduler/RunJob"
	Scheduler_ScheduleJob_FullMethodName   = "/services.Scheduler/ScheduleJob"
	Scheduler_Start_FullMethodName         = "/services.Scheduler/Start"
	Scheduler_Stop_FullMethodName          = "/services.Scheduler/Stop"
)

// SchedulerClient is the client API for Scheduler service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchedulerClient interface {
	AddJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error)
	GetJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error)
	GetAllJobs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*JobsResp, error)
	UpdateJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error)
	DeleteJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteAllJobs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	PauseJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error)
	ResumeJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error)
	RunJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ScheduleJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Start(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Stop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type schedulerClient struct {
	cc grpc.ClientConnInterface
}

func NewSchedulerClient(cc grpc.ClientConnInterface) SchedulerClient {
	return &schedulerClient{cc}
}

func (c *schedulerClient) AddJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Job)
	err := c.cc.Invoke(ctx, Scheduler_AddJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) GetJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Job)
	err := c.cc.Invoke(ctx, Scheduler_GetJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) GetAllJobs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*JobsResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(JobsResp)
	err := c.cc.Invoke(ctx, Scheduler_GetAllJobs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) UpdateJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*Job, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Job)
	err := c.cc.Invoke(ctx, Scheduler_UpdateJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) DeleteJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_DeleteJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) DeleteAllJobs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_DeleteAllJobs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) PauseJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Job)
	err := c.cc.Invoke(ctx, Scheduler_PauseJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) ResumeJob(ctx context.Context, in *JobReq, opts ...grpc.CallOption) (*Job, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Job)
	err := c.cc.Invoke(ctx, Scheduler_ResumeJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) RunJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_RunJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) ScheduleJob(ctx context.Context, in *Job, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_ScheduleJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) Start(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_Start_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) Stop(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Scheduler_Stop_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SchedulerServer is the server API for Scheduler service.
// All implementations must embed UnimplementedSchedulerServer
// for forward compatibility.
type SchedulerServer interface {
	AddJob(context.Context, *Job) (*Job, error)
	GetJob(context.Context, *JobReq) (*Job, error)
	GetAllJobs(context.Context, *emptypb.Empty) (*JobsResp, error)
	UpdateJob(context.Context, *Job) (*Job, error)
	DeleteJob(context.Context, *JobReq) (*emptypb.Empty, error)
	DeleteAllJobs(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	PauseJob(context.Context, *JobReq) (*Job, error)
	ResumeJob(context.Context, *JobReq) (*Job, error)
	RunJob(context.Context, *Job) (*emptypb.Empty, error)
	ScheduleJob(context.Context, *Job) (*emptypb.Empty, error)
	Start(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	Stop(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedSchedulerServer()
}

// UnimplementedSchedulerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSchedulerServer struct{}

func (UnimplementedSchedulerServer) AddJob(context.Context, *Job) (*Job, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddJob not implemented")
}
func (UnimplementedSchedulerServer) GetJob(context.Context, *JobReq) (*Job, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetJob not implemented")
}
func (UnimplementedSchedulerServer) GetAllJobs(context.Context, *emptypb.Empty) (*JobsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllJobs not implemented")
}
func (UnimplementedSchedulerServer) UpdateJob(context.Context, *Job) (*Job, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateJob not implemented")
}
func (UnimplementedSchedulerServer) DeleteJob(context.Context, *JobReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteJob not implemented")
}
func (UnimplementedSchedulerServer) DeleteAllJobs(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAllJobs not implemented")
}
func (UnimplementedSchedulerServer) PauseJob(context.Context, *JobReq) (*Job, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseJob not implemented")
}
func (UnimplementedSchedulerServer) ResumeJob(context.Context, *JobReq) (*Job, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResumeJob not implemented")
}
func (UnimplementedSchedulerServer) RunJob(context.Context, *Job) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunJob not implemented")
}
func (UnimplementedSchedulerServer) ScheduleJob(context.Context, *Job) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ScheduleJob not implemented")
}
func (UnimplementedSchedulerServer) Start(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedSchedulerServer) Stop(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedSchedulerServer) mustEmbedUnimplementedSchedulerServer() {}
func (UnimplementedSchedulerServer) testEmbeddedByValue()                   {}

// UnsafeSchedulerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SchedulerServer will
// result in compilation errors.
type UnsafeSchedulerServer interface {
	mustEmbedUnimplementedSchedulerServer()
}

func RegisterSchedulerServer(s grpc.ServiceRegistrar, srv SchedulerServer) {
	// If the following call pancis, it indicates UnimplementedSchedulerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Scheduler_ServiceDesc, srv)
}

func _Scheduler_AddJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).AddJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_AddJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).AddJob(ctx, req.(*Job))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_GetJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).GetJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_GetJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).GetJob(ctx, req.(*JobReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_GetAllJobs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).GetAllJobs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_GetAllJobs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).GetAllJobs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_UpdateJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).UpdateJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_UpdateJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).UpdateJob(ctx, req.(*Job))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_DeleteJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).DeleteJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_DeleteJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).DeleteJob(ctx, req.(*JobReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_DeleteAllJobs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).DeleteAllJobs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_DeleteAllJobs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).DeleteAllJobs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_PauseJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).PauseJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_PauseJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).PauseJob(ctx, req.(*JobReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_ResumeJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JobReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).ResumeJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_ResumeJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).ResumeJob(ctx, req.(*JobReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_RunJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).RunJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_RunJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).RunJob(ctx, req.(*Job))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_ScheduleJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Job)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).ScheduleJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_ScheduleJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).ScheduleJob(ctx, req.(*Job))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_Start_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).Start(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Scheduler_Stop_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).Stop(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Scheduler_ServiceDesc is the grpc.ServiceDesc for Scheduler service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Scheduler_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "services.Scheduler",
	HandlerType: (*SchedulerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddJob",
			Handler:    _Scheduler_AddJob_Handler,
		},
		{
			MethodName: "GetJob",
			Handler:    _Scheduler_GetJob_Handler,
		},
		{
			MethodName: "GetAllJobs",
			Handler:    _Scheduler_GetAllJobs_Handler,
		},
		{
			MethodName: "UpdateJob",
			Handler:    _Scheduler_UpdateJob_Handler,
		},
		{
			MethodName: "DeleteJob",
			Handler:    _Scheduler_DeleteJob_Handler,
		},
		{
			MethodName: "DeleteAllJobs",
			Handler:    _Scheduler_DeleteAllJobs_Handler,
		},
		{
			MethodName: "PauseJob",
			Handler:    _Scheduler_PauseJob_Handler,
		},
		{
			MethodName: "ResumeJob",
			Handler:    _Scheduler_ResumeJob_Handler,
		},
		{
			MethodName: "RunJob",
			Handler:    _Scheduler_RunJob_Handler,
		},
		{
			MethodName: "ScheduleJob",
			Handler:    _Scheduler_ScheduleJob_Handler,
		},
		{
			MethodName: "Start",
			Handler:    _Scheduler_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Scheduler_Stop_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "scheduler.proto",
}
