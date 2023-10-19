# go run examples/rpc/rpc_server.go
# python3 examples/rpc/python/rpc_client.py

import grpc
from google.protobuf.struct_pb2 import Struct

import scheduler_pb2


class SchedulerStub(object):
    def __init__(self, channel):
        self.AddJob = channel.unary_unary(
            "/scheduler.Scheduler/AddJob",
            request_serializer=scheduler_pb2.JobId.SerializeToString,
            response_deserializer=scheduler_pb2.Job.FromString,
        )


def run():
    with grpc.insecure_channel("127.0.0.1:36363") as channel:
        args = Struct()
        args.update({"arg1": "1", "arg2": "2", "arg3": "3"}),
        stub = SchedulerStub(channel)
        pb_job = stub.AddJob(
            scheduler_pb2.Job(
                name="Job1",
                type="interval",
                interval="2s",
                timezone="UTC",
                func_name="main.printMsg",
                args=args,
            )
        )
        print("Scheduler add job: ", pb_job)


if __name__ == "__main__":
    run()
