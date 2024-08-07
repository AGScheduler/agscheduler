# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from proto import recorder_pb2 as recorder__pb2
from proto import scheduler_pb2 as scheduler__pb2

GRPC_GENERATED_VERSION = '1.65.4'
GRPC_VERSION = grpc.__version__
EXPECTED_ERROR_RELEASE = '1.66.0'
SCHEDULED_RELEASE_DATE = 'August 6, 2024'
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    warnings.warn(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in recorder_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
        + f' This warning will become an error in {EXPECTED_ERROR_RELEASE},'
        + f' scheduled for release on {SCHEDULED_RELEASE_DATE}.',
        RuntimeWarning
    )


class RecorderStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetRecords = channel.unary_unary(
                '/services.Recorder/GetRecords',
                request_serializer=recorder__pb2.RecordsReq.SerializeToString,
                response_deserializer=recorder__pb2.RecordsResp.FromString,
                _registered_method=True)
        self.GetAllRecords = channel.unary_unary(
                '/services.Recorder/GetAllRecords',
                request_serializer=recorder__pb2.RecordsAllReq.SerializeToString,
                response_deserializer=recorder__pb2.RecordsResp.FromString,
                _registered_method=True)
        self.DeleteRecords = channel.unary_unary(
                '/services.Recorder/DeleteRecords',
                request_serializer=scheduler__pb2.JobReq.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                _registered_method=True)
        self.DeleteAllRecords = channel.unary_unary(
                '/services.Recorder/DeleteAllRecords',
                request_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                _registered_method=True)


class RecorderServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GetRecords(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetAllRecords(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def DeleteRecords(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def DeleteAllRecords(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_RecorderServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetRecords': grpc.unary_unary_rpc_method_handler(
                    servicer.GetRecords,
                    request_deserializer=recorder__pb2.RecordsReq.FromString,
                    response_serializer=recorder__pb2.RecordsResp.SerializeToString,
            ),
            'GetAllRecords': grpc.unary_unary_rpc_method_handler(
                    servicer.GetAllRecords,
                    request_deserializer=recorder__pb2.RecordsAllReq.FromString,
                    response_serializer=recorder__pb2.RecordsResp.SerializeToString,
            ),
            'DeleteRecords': grpc.unary_unary_rpc_method_handler(
                    servicer.DeleteRecords,
                    request_deserializer=scheduler__pb2.JobReq.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'DeleteAllRecords': grpc.unary_unary_rpc_method_handler(
                    servicer.DeleteAllRecords,
                    request_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'services.Recorder', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('services.Recorder', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class Recorder(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GetRecords(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/services.Recorder/GetRecords',
            recorder__pb2.RecordsReq.SerializeToString,
            recorder__pb2.RecordsResp.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def GetAllRecords(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/services.Recorder/GetAllRecords',
            recorder__pb2.RecordsAllReq.SerializeToString,
            recorder__pb2.RecordsResp.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def DeleteRecords(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/services.Recorder/DeleteRecords',
            scheduler__pb2.JobReq.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def DeleteAllRecords(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/services.Recorder/DeleteAllRecords',
            google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
