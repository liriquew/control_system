import tensorflow as tf
tf.keras.config.disable_interactive_logging()

from keras.api.models import load_model
from scipy.sparse import hstack
import numpy as np
import joblib

import heapq

from sklearn.preprocessing import MultiLabelBinarizer
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.ensemble import GradientBoostingRegressor

import pickle
import functools

import predictions_service.predictions_service_pb2 as pb

class Tag:
    tag_name: str
    probability: float
    id: int


    def __init__(self, tag_name: str, id: int, probability: float=None):
        self.tag_name = tag_name
        self.probability = probability
        self.id = id


    def toPb(self) -> pb.Tag:
        res  = pb.Tag(
            Id=self.id,
            Name=self.tag_name,
            Probability=self.probability,
        )
        return res


    def __str__(self):
        return f"id: {self.id}, probability: {self.probability}, tag_name: {self.tag_name}"


class PredictInfo:
    id: int
    uid: int
    planned_time: float
    actual_time: float
    tags: list[int]
    tags_vector: list[float]

    def __init__(self, id: int, uid: int, tags: list[int], planned_time: float, actual_time: float=0.0):
        self.planned_time = planned_time
        self.actual_time = actual_time
        self.tags = tags
        self.id = id
        self.uid = uid
        self.tags_vector = None

    @classmethod
    def from_dict(cls, data: dict[str, int|float|list[int]], UID: int=None):
        if "user_id" in data:
            UID = data["user_id"]
        return cls(
            id=data["id"],
            uid=UID,
            tags=data["tags"],
            planned_time=data["planned_time"],
            actual_time=data["actual_time"],
        )

    def tags_ids(self) -> list[int]:
        return self.tags
    
    def set_tags_vector(self, vector: list[float]):
        self.tags_vector = vector

    def get_predict_vector(self) -> list[float]:
        return [self.planned_time] + self.tags_vector
    
    def get_predicted_value(self) -> float:
        return self.actual_time

    def __str__(self) -> str:
        s = f"id: {self.id}, uid: {self.uid}, plannned_time: {self.planned_time}, actual_time: {self.actual_time}, tags: {self.tags}"
        if self.tags_vector is not None:
            s += f"\ntags_vector: {self.tags_vector}"
        return s


class TagsPredicator:
    def __init__(self, config: dict[str, str]):
        self.classifacator = load_model(config["classificator_path"])
        self.vectorizer_body: TfidfVectorizer = joblib.load(config["body_vectorizer"])
        self.binarizer_tags: MultiLabelBinarizer = joblib.load(config["tags_vectorizer"])


    @staticmethod
    def get_top_ten(arr: list[float]):
        h = []
        for i, val in enumerate(arr):
            h.append((-val, [i, val]))
        heapq.heapify(h)

        res = []
        for i in range(10):
            res.append(heapq.heappop(h)[1])

        return res


    def predict(self, body: str) -> list[Tag]:
        """
        make tags predict, returns (tags, tags ids)
        """
        model_input = self.vectorizer_body.transform([body])

        model_input = hstack([model_input]).toarray()
        tags_predict = self.classifacator.predict(model_input)[0]
        recent_tags = self.get_top_ten(tags_predict)

        tags = self.binarizer_tags.classes_[[t for t, _ in recent_tags]]

        predictions = []
        print("LOG:")
        for tag_name, (id, probability) in zip(tags, recent_tags):
            predictions.append(Tag(
                id=id,
                tag_name=tag_name,
                probability=probability
            ))
    
        return predictions
    

    @functools.cache
    def get_tags_list(self) -> list[Tag]:
        return [Tag(tag_name, id) for id, tag_name in enumerate(self.binarizer_tags.classes_)]


class TaskPredicator:
    model: GradientBoostingRegressor
    tags_map: dict[int, int]

    # параметры модели в зависимости от размера выборки (числа известных задач)
    MODEL_PARAMS = [
        {
            'sample_size': 20,
            'n_estimators': 10,
            'max_depth': 2,
            'min_samples_split': 2,
        },
        {
            'sample_size': 100,
            'n_estimators': 50,
            'learning_rate': 0.05,
            'max_depth': 4,
            'min_samples_split': 3,
        },
        {
            'sample_size': 500,
            'n_estimators': 100,
            'learning_rate': 0.1,
            'max_depth': 5,
            'min_samples_split': 5,
        }
    ]

    @classmethod
    def get_model_params(self, sample_size: int):
        """
        Определяет параметры модели в зависимости от объема выборки
        """
        for params in self.MODEL_PARAMS:
            if sample_size < params["sample_size"]:
                return {k: v for k, v in params.items() if k != "sample_size"}
        return {k: v for k, v in self.MODEL_PARAMS[-1].items() if k != "sample_size"}


    def __init__(self, model: GradientBoostingRegressor, tags_map: dict[int, int]):
        self.model = model
        self.tags_map = tags_map


    @classmethod
    def from_model_blob(cls, blob: bytes):
        model, tags_map = pickle.loads(blob)
        return cls(model, tags_map)


    @classmethod
    def from_tasks(cls, tasks: list[PredictInfo]):
        tags_map = cls.vectorize_task_tags(tasks)
        x, y = cls.prepare_tasks_data(tasks)

        model_params = cls.get_model_params(len(tasks))
        model = GradientBoostingRegressor(**model_params)
        model.fit(x, y)

        return cls(model, tags_map)
        

    @staticmethod
    def vectorize_task_tags(tasks: list[PredictInfo]) -> dict[int, int]:
        """
        Устанавливает для каждой задачи сжатый вектор тегов

        всего тегов 50 (пока что), в простом случае предополагается, что в выборке не будет столько задач,
        что множество тегов будет включать все теги, поэтому, все теги сжимаются.

        каждому тегу присваивается уникальный идентификатор из [0, |set(tags)|]
        потом для каждой задачи формируется разреженный вектор длины |set(tags)|,
        в котором компорнента i с значением 1 указывает, что у задачи есть тег i
        """
        tags_map = dict() # tag_id -> index in compressed vector

        # determine index in vector for each tag_id
        vector_tag_id = 0
        for task in tasks:
            for tag_id in task.tags_ids():
                if not tag_id in tags_map:
                    tags_map[tag_id] = vector_tag_id
                    vector_tag_id += 1

        # create for each task tags_vector
        for task in tasks:
            task_tags_vector = [float(0.0)] * len(tags_map)
            for tag in task.tags_ids():
                task_tags_vector[tags_map[tag]] = 1.0

            task.set_tags_vector(task_tags_vector)

        return tags_map


    @staticmethod
    def prepare_tasks_data(tasks: list[PredictInfo]) -> tuple[np.ndarray, np.ndarray]:
        """
        Подготавливает данные задач для обучения модели
        """
        x, y = [], []
        for task in tasks:
            x.append(task.get_predict_vector())
            y.append(task.get_predicted_value())
        
        return np.array(x), np.array(y)


    def predict(self, task: PredictInfo) -> float:
        x = [task.planned_time] + [0.0] * len(self.tags_map)

        for tag_id in task.tags_ids():
            if tag_id in self.tags_map:
                x[self.tags_map[tag_id] + 1] = 1.0
        
        return float(self.model.predict(np.array([x]))[0])
    

    def dump(self) -> bytes:
        return pickle.dumps((self.model, self.tags_map))