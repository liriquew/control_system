from sklearn.ensemble import GradientBoostingRegressor
import numpy as np
import grpc
import predictions_service.predictions_service_pb2 as pb


from database import Database, ExceptionDB
import predicator

class PredicatorException(Exception):
    def __init__(self, message, extra_info):
        super().__init__(message)
        self.extra_info = extra_info

class PredictionService():
    def __init__(self, db:Database, tag_classificator: predicator.TagsPredicator):
        self.db = db
        self.tag_predicator = tag_classificator
        
    def _validate_uid(self, uid: int):
        """
        Проверяет, что UID корректен
        """
        if uid <= 0:
            raise PredicatorException("UID must be greater than zero", grpc.StatusCode.INVALID_ARGUMENT)


    def create_model(self, UID: int) -> predicator.TaskPredicator:
        """
        Обучает и сохраняет модель
        """
        print("predictions_service.Predicator.fit_model()")
        self._validate_uid(UID)

        # выборка задач из бд
        tasks = self.db.get_user_tasks(UID)
        if len(tasks) == 0:
            self.db.delete_model(UID)
            raise PredicatorException(f"User with UID:{UID} does not have any completed tasks", grpc.StatusCode.FAILED_PRECONDITION)

        model = predicator.TaskPredicator.from_tasks(tasks=tasks)
        return model


    def get_model(self, UID: int):
        """
        Возвращает обученную модель пользователя из БД
        
        В случае если модель не актуальна или отсутствует в БД,
        то обучает и сохраняет ее для дальнейшего использования
        """
        print("predictions_service.Predicator.get_model()")
        try:
            model_blob = None
            model_blob = self.db.load_model(UID)
            model = predicator.TaskPredicator.from_model_blob(model_blob)
        except PredicatorException as pe:
            raise pe
        except ExceptionDB:
            pass
        if model_blob is None:
            model = self.create_model(UID)
            self.db.save_model(UID, model.dump())
        return model


    def make_predict(self, info: predicator.PredictInfo) -> float:
        """
        Загружает модель из бд и предсказывает итоговое время выполнения задачи
        """
        print("predictions_service.Predicator.make_predict()")

        self._validate_uid(info.uid)
        try:
            model = self.get_model(info.uid)
        except PredicatorException as pe:
            if pe.extra_info == grpc.StatusCode.FAILED_PRECONDITION:
                return 0.0
            raise pe

        return model.predict(info)
    

    def make_list_predict(self, predict_info: list[predicator.PredictInfo]) -> tuple[list[pb.PredictedInfo], list[int]]:
        """
        Загружает модель для каждого пользователя из бд
        и предсказывает итоговое время выполнения задач
        """
        print("predictions_service.Predicator.make_list_predict()")
        if len(predict_info) == 0:
            raise PredicatorException(f"Empty user with time list", grpc.StatusCode.INVALID_ARGUMENT)

        cache = dict()
        unpredicted_uids = set()

        for i, info in enumerate(predict_info):
            if info.uid not in cache:
                try:
                    model = self.get_model(info.uid)
                    cache[info.uid] = model
                except PredicatorException:
                    # user doesn't have any tasks, nothing to predict
                    predict_info[i].actual_time = predict_info[i].planned_time
                    unpredicted_uids.add(info.uid)
                    continue
            else:
                model = cache[info.uid]
            predict_info[i].actual_time = model.predict(info)

        return predict_info, list(unpredicted_uids)
    

    def predict_tags(self, title: str, description: str) -> list[predicator.Tag]:
        """
        Предсказывает теги по заголовку и описанию задачи
        """
        tags = self.tag_predicator.predict(title, description)

        return tags
    
    def get_tags_list(self) -> list[predicator.Tag]:
        try:
            return self.tag_predicator.get_tags_list()
        except Exception as e:
            print(e)

        return []
