import grpc
from concurrent import futures

import predictions_service.predictions_service_pb2 as pb2
import predictions_service.predictions_service_pb2_grpc as pb2_grpc
from google.protobuf.json_format import MessageToDict, ParseDict

from predictions import Predicator, PredicatorException

class PredictService(pb2_grpc.PredictionsServicer):
    def __init__(self):
        self.service = Predicator()
        pass

    def RecalculateAndSaveTask(self, request: pb2.RecalculateAndSaveTaskRequest, context: grpc.aio.ServicerContext) -> pb2.RecalculateAndSaveTaskResponse:
        print('RecalculateAndSaveTask')
        
        updated_fields = MessageToDict(request)
        
        print(request)

        try:
            self.service.fit_model(**updated_fields)
        except PredicatorException as e:
            print(e, e.extra_info)
            context.set_code(e.extra_info)
            context.set_details(str(e))
            return pb2.RecalculateAndSaveTaskResponse(Status=str(e))
        except Exception as e:
            print("INTERNAL:", e)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb2.RecalculateAndSaveTaskResponse(Status=str(e))

        context.set_code(grpc.StatusCode.OK)
        return pb2.RecalculateAndSaveTaskResponse(Status="ok")

    def Predict(self, request: pb2.PredictRequest, context: grpc.aio.ServicerContext) -> pb2.PredictResponse:
        print('Predict')
        print(request)
        predict : float
        try:
            predict = self.service.make_predict(request.UID, request.PlannedTime)
        except PredicatorException as e:
            print(e, e.extra_info)
            context.set_code(e.extra_info)
            context.set_details(str(e))
            return pb2.PredictResponse(Status=str(e))
        except Exception as e:
            print("INTERNAL:", e)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb2.PredictResponse(Status=str(e))

        context.set_code(grpc.StatusCode.OK)
        return pb2.PredictResponse(ActualTime=predict, Status="ok")
    
    def Recalculete(self, request, context):
        print('Recalculate')

        print(request)

        try:
            self.service.recalulate_model(request.UID)
        except PredicatorException as e:
            print(e, e.extra_info)
            context.set_code(e.extra_info)
            context.set_details(str(e))
            return pb2.RecalculateResponse(Status=str(e))
        except Exception as e:
            print("INTERNAL:", e)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb2.RecalculateResponse(Status=str(e))

        context.set_code(grpc.StatusCode.OK)
        return pb2.RecalculateResponse(Status="ok")


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb2_grpc.add_PredictionsServicer_to_server(PredictService(), server)
    server.add_insecure_port('0.0.0.0:4041')
    print("RUN on 0.0.0.0:4041")
    print("python-server:4041")
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    serve()