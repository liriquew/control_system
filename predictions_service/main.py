#!/home/ql/project/GW/repo/predictions_service/venv/bin/python
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'
os.environ['TF_XLA_FLAGS'] = '--tf_xla_enable_xla_devices=false'

import absl.logging
absl.logging.set_verbosity(absl.logging.ERROR)

import tensorflow as tf
tf.get_logger().setLevel('ERROR')
tf.config.set_visible_devices([], 'GPU')

import warnings
from sklearn.exceptions import InconsistentVersionWarning
warnings.filterwarnings("ignore", category=InconsistentVersionWarning)

import signal
import grpc
from concurrent import futures
from typing import Dict, Any

import predictions_service.predictions_service_pb2 as pb
import predictions_service.predictions_service_pb2_grpc as pb2_grpc
from google.protobuf.json_format import MessageToDict
from google.protobuf.internal import containers as _containers

from service import PredictionService, PredicatorException
from config import ConfigLoader
from database import Database
from consumer import KafkaMLConsumer
import predicator

class PredictionsServer(pb2_grpc.PredictionsServicer):
    def __init__(self, db: Database, tag_classificator: predicator.TagsPredicator):
        self.service = PredictionService(db, tag_classificator)


    def _handle_exception(self, context: grpc.ServicerContext, exception: Exception):
        print(f"Error: {exception}, Details: {exception.extra_info if hasattr(exception, "extra_info") else ""}")
        if isinstance(exception, PredicatorException):
            context.abort(exception.extra_info, str(exception))
        else:
            context.abort(grpc.StatusCode.INTERNAL, str(exception))


    @staticmethod
    def PredictInfo(info: pb.PredictInfo) -> predicator.PredictInfo:
        return predicator.PredictInfo(
                id=info.ID,
                uid=info.UID,
                planned_time=info.PlannedTime,
                tags=info.TagsIDs,
            )


    @staticmethod
    def PredictedInfo(info: predicator.PredictInfo) -> pb.PredictedInfo:
        return pb.PredictedInfo(
            ID=info.id,
            UID=info.uid,
            PredictedTime=info.actual_time,
        )


    def Predict(self, request: pb.PredictRequest, context: grpc.ServicerContext) -> pb.PredictResponse:
        print("Predict")
        try:
            task = self.PredictInfo(request.Info)
            predict = self.service.make_predict(task)
        except Exception as e:
            self._handle_exception(context, e)

        context.set_code(grpc.StatusCode.OK)
        return pb.PredictResponse(ActualTime=predict)

    
    def PredictList(self, request: pb.PredictListRequest, context: grpc.ServicerContext) -> pb.PredictListResponse:
        print("PredictList")
        try:
            print(MessageToDict(request))
            predict_info_list = [self.PredictInfo(info) for info in request.Infos]
            predicts, unpredicted_uids = self.service.make_list_predict(predict_info_list)
            predicted_infos = [self.PredictedInfo(predict) for predict in predicts]
        except Exception as e:
            self._handle_exception(context, e)

        context.set_code(grpc.StatusCode.OK)
        return pb.PredictListResponse(PredictedUserTime=predicted_infos, UnpredictedUIDs=unpredicted_uids)


    def PredictTags(self, request: pb.PredictTagRequest, context: grpc.ServicerContext) -> pb.PredictTagResponse:
        print("PredictTags")
        try:
            print(MessageToDict(request))
            tags = self.service.predict_tags(request.Description)
            tags = [tag.toPb() for tag in tags]
        except Exception as e:
            self._handle_exception(context, e)

        context.set_code(grpc.StatusCode.OK)
        return pb.PredictTagResponse(Tags=tags)


    def GetTags(self, reuest, context: grpc.ServicerContext) -> pb.TagList:
        print("GetTags")
        return pb.TagList(Tags=[tag.toPb() for tag in self.service.get_tags_list()])


def serve():
    app_config = ConfigLoader()
    service_config = app_config.get_service_config()

    db = Database(app_config.get_database_config())

    consumer = KafkaMLConsumer(app_config.get_kafka_config(), db)
    tags_predicator = predicator.TagsPredicator(app_config.get_classificator_config())

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb2_grpc.add_PredictionsServicer_to_server(PredictionsServer(db, tags_predicator), server)
    
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

    