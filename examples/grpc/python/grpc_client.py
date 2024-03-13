# 1. go run examples/grpc/grpc_server.go
# 2. python3 examples/grpc/python/grpc_client.py

import grpc
from google.protobuf.empty_pb2 import Empty
from google.protobuf.struct_pb2 import Struct

from proto import scheduler_pb2
from proto import scheduler_pb2_grpc


def run():
    with grpc.insecure_channel("127.0.0.1:36360") as channel:
        args = Struct()
        args.update({"arg1": "1", "arg2": "2", "arg3": "3"}),
        stub = scheduler_pb2_grpc.SchedulerStub(channel)
        pb_job = stub.AddJob(
            scheduler_pb2.Job(
                name="Job1",
                type="interval",
                interval="2s",
                timezone="UTC",
                func_name="github.com/kwkwc/agscheduler/examples.PrintMsg",
                args=args,
            )
        )
        print(pb_job)

        stub.Start(Empty())


if __name__ == "__main__":
    run()
