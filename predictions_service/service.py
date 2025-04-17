from sklearn.ensemble import GradientBoostingRegressor
import numpy as np
import grpc
import predictions_service.predictions_service_pb2 as pb2


from database import Database, ExceptionDB
from config import ConfigLoader

class PredicatorException(Exception):
    def __init__(self, message, extra_info):
        super().__init__(message)
        self.extra_info = extra_info

class PredictionService():
    # параметры модели в зависимости от размера выборки (числа известных задач)
    MODEL_PARAMS = [
        {
            'sample_size': 20,
            'n_estimators': 10,
            'max_depth': 3,
            'min_samples_split': 2
        },
        {
            'sample_size': 100,
            'n_estimators': 50,
            'learning_rate': 0.05,
            'max_depth': 4,
            'min_samples_split': 3
        },
        {
            'sample_size': 500,
            'n_estimators': 100,
            'learning_rate': 0.1,
            'max_depth': 5,
            'min_samples_split': 5
        }
    ]


    def __init__(self, db:Database):
        self.db = db


    def get_model_params(self, sample_size: int):
        """
        Определяет параметры модели в зависимости от объема выборки
        """
        for params in self.MODEL_PARAMS:
            if sample_size < params["sample_size"]:
                return {k: v for k, v in params.items() if k != "sample_size"}
        return {k: v for k, v in self.MODEL_PARAMS[-1].items() if k != "sample_size"}


    def _validate_uid(self, uid: int):
        """
        Проверяет, что UID корректен
        """
        if uid <= 0:
            raise PredicatorException("UID must be greater than zero", grpc.StatusCode.INVALID_ARGUMENT)


    def _prepare_tasks_data(self, tasks: list) -> tuple[np.ndarray, np.ndarray]:
        """
        Подготавливает данные задач для обучения модели
        """
        tasks_data = np.array(tasks).reshape(-1, 3)
        _, planned_time, actual_time = np.array_split(tasks_data, 3, axis=1)
        return planned_time, np.ravel(actual_time)


    def fit_model(self, UID: int):
        """
        Обучает и сохраняет модель
        """
        print("predictions_service.Predicator.fit_model()")
        self._validate_uid(UID)

        # выборка задач из бд
        tasks = self.db.get_user_tasks(UID)
        if len(tasks[0]) == 0:
            self.db.delete_model(UID)
            raise PredicatorException(f"User with UID:{UID} does not have any completed tasks", grpc.StatusCode.FAILED_PRECONDITION)

        planned_time, actual_time = self._prepare_tasks_data(tasks)

        # Определение и сохранение модели
        model_params = self.get_model_params(len(planned_time))
        model = GradientBoostingRegressor(**model_params)
        model.fit(planned_time, actual_time)
        self.db.save_model(UID, model)

        return model


    def get_model(self, UID: int):
        """
        Возвращает обученную модель пользователя из БД
        
        В случае если модель не актуальна или отсутствует в БД,
        то обучает и сохраняет ее для дальнейшего использования
        """
        print("predictions_service.Predicator.get_model()")
        try:
            model = None
            model = self.db.load_model(UID)
        except ExceptionDB as e:
            pass
        if model is None:
            model = self.fit_model(UID)
        return model


    def predict_single_value(self, model, value: float) -> float:
        return float(model.predict([[value]])[0])


    def make_predict(self, UID: int, PlannedTime: float) -> float:
        """
        Загружает модель из бд и предсказывает итоговое время выполнения задачи
        """
        print("predictions_service.Predicator.make_predict()")
        self._validate_uid(UID)
        if PlannedTime <= 0:
            raise PredicatorException(f"Invalid PlannedTime (must be greater than zero)", grpc.StatusCode.INVALID_ARGUMENT)
        try:
            model = self.get_model(UID)
        except PredicatorException as pe:
            if pe.extra_info == grpc.StatusCode.FAILED_PRECONDITION:
                return 0.0
            raise pe
        return self.predict_single_value(model, PlannedTime)
    

    def make_list_predict(self, users_with_times: list[pb2.UserWithTime]) -> tuple[list[pb2.UserWithTime], list[int]]:
        """
        Загружает модель для каждого пользователя из бд
        и предсказывает итоговое время выполнения задач
        """
        print("predictions_service.Predicator.make_list_predict()")
        if len(users_with_times) == 0:
            raise PredicatorException(f"Empty user with time list", grpc.StatusCode.INVALID_ARGUMENT)

        cache = dict()
        unpredicted_uids = set()
        
        for i, user_with_time in enumerate(users_with_times):
            if user_with_time.UID not in cache:
                try:
                    model = self.get_model(user_with_time.UID)
                    cache[user_with_time.UID] = model
                except PredicatorException:
                    # user doesn't have any tasks, nothing to predict
                    unpredicted_uids.add(user_with_time.UID)
                    continue
            else:
                model = cache[user_with_time.UID]
            users_with_times[i].Time = self.predict_single_value(model, user_with_time.Time)

        return users_with_times, list(unpredicted_uids)