import signal
import grpc
from concurrent import futures
from typing import Dict, Any

import predictions_service.predictions_service_pb2 as pb2
import predictions_service.predictions_service_pb2_grpc as pb2_grpc
from google.protobuf.json_format import MessageToDict
from google.protobuf.internal import containers as _containers

from service import PredictionService, PredicatorException
from config import ConfigLoader
from database import Database
from consumer import KafkaMLConsumer

class PredictionsServer(pb2_grpc.PredictionsServicer):
    def __init__(self, config: Dict[str, Any]):
        self.service = PredictionService(config)


    def _handle_exception(self, context: grpc.ServicerContext, exception: Exception):
        print(f"Error: {exception}, Details: {exception.extra_info if hasattr(exception, "extra_info") else ""}")
        if isinstance(exception, PredicatorException):
            context.abort(exception.extra_info, str(exception))
        else:
            context.abort(grpc.StatusCode.INTERNAL, str(exception))


    def Predict(self, request: pb2.PredictRequest, context: grpc.ServicerContext) -> pb2.PredictResponse:
        print("Predict")
        try:
            predict = self.service.make_predict(request.UID, request.PlannedTime)
        except Exception as e:
            self._handle_exception(context, e)

        context.set_code(grpc.StatusCode.OK)
        return pb2.PredictResponse(ActualTime=predict, Status="ok")

    
    def PredictList(self, request: pb2.PredictListRequest, context: grpc.ServicerContext) -> pb2.PredictListResponse:
        print("PredictList")
        try:
            print(MessageToDict(request))
            predicts, unpredicted_uids = self.service.make_list_predict(request.PlannedUserTime)
        except Exception as e:
            self._handle_exception(context, e)

        context.set_code(grpc.StatusCode.OK)
        return pb2.PredictListResponse(PredictedUserTime=predicts, UnpredictedUIDs=unpredicted_uids)

def serve():
    app_config = ConfigLoader()
    service_config = app_config.get_service_config()

    db = Database(app_config.get_database_config())

    consumer = KafkaMLConsumer(app_config.get_kafka_config(), db)

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb2_grpc.add_PredictionsServicer_to_server(PredictionsServer(db), server)
    
    conn_str = f"{service_config["host"]}:{service_config["port"]}"
    server.add_insecure_port(conn_str)

    def handle_signal(sig, frame):
        print("\nReceived shutdown signal, stopping server...")
        server.stop(0)
        exit(0)
    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    print(f"Server running on {conn_str}")
    consumer.start()
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()

    