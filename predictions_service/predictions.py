from sklearn.ensemble import GradientBoostingRegressor
import numpy as np
import grpc 

from database import Database, ExceptionDB
from config import ConfigLoader

class PredicatorException(Exception):
    def __init__(self, message, extra_info):
        super().__init__(message)
        self.extra_info = extra_info

class Predicator():
    # параметры модели в зависимости от размера выборки (числа известных задач)
    model_params = [
        {
            'sample_size': 20,
            'n_estimators': 10,
            # 'learning_rate': 0.01,
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

    def __init__(self, config: ConfigLoader):
        self.db = Database(config.get_database_config())

    def get_params(self, n: int):
        '''
        Определяет параметры модели в зависимости от объема выборки
        '''
        for model_params in self.model_params:
            if n < model_params['sample_size']:
                params = model_params.copy()
                params.pop('sample_size')
                return params
            
        params = self.models_params[-1].copy()
        params.pop('sample_size')
        return params

    def recalulate_model(self, UID: int):
        '''
        Обучает и сохраняет модель
        '''
        if UID <= 0:
            raise PredicatorException(f"UID must be greater than zero", grpc.StatusCode.INVALID_ARGUMENT)
        # выборка задач из бд
        tasks = self.db.get_user_tasks(UID)
        if len(tasks[0]) == 0:
            self.db.delete_model(UID)
            raise PredicatorException("User does not have any completed tasks", grpc.StatusCode.FAILED_PRECONDITION)


        tasks_data = np.array(tasks)
        tasks_data = np.reshape(tasks_data, (-1, 3))

        _, planned_time, actual_time = np.array_split(tasks_data, 3, axis=1)
        actual_time = np.ravel(actual_time)

        # Определение и сохранение модели
        model_params = self.get_params(len(tasks_data))
        gb_model = GradientBoostingRegressor(**model_params)
        gb_model.fit(planned_time, actual_time)

        self.db.save_model(UID, gb_model)
        return gb_model

    def fit_model(self, **kwargs):
        '''
        Обучает и сохраняет полученную модель
        '''
        if "UID" not in kwargs:
            raise PredicatorException(f"UID required", grpc.StatusCode.INVALID_ARGUMENT)
        if "ID" not in kwargs:
            raise PredicatorException(f"ID (task_id) required", grpc.StatusCode.INVALID_ARGUMENT, )
        if "ActualTime" not in kwargs:
            raise PredicatorException(f"ActualTime required", grpc.StatusCode.INVALID_ARGUMENT)

        UID = int(kwargs.pop("UID"))
        task_id = int(kwargs.pop("ID"))
        new_actual_time = float(kwargs.pop("ActualTime"))

        print(f'UID: {UID}; task_id: {task_id}, actual_time: {new_actual_time}')
        
        # выборка задач из бд
        tasks = self.db.get_user_tasks(UID)
        tasks_data = np.array(tasks)
        tasks_data = np.reshape(tasks_data, (-1, 3))
        
        # Если изменено время выполнения уже существующей задачи,
        # то соответствующая запись в выборке будет изменена
        # в противном случае будет добавлена новая запись к выборке
        for task in tasks_data:
            if task[0] == task_id:
                # для измененной задачи ранее было известно итоговое время выполнения
                print("Task found in completed, change")
                task[2] = new_actual_time
                break
        else:
            # для измененной задачи ранее не было известно итоговое время выполнения
            print("Task not found in completed, add new one")
            changed_task = self.db.get_task(UID, task_id)
            if changed_task is None:
                raise PredicatorException(f"Task not found task_id: {task_id}", grpc.StatusCode.INVALID_ARGUMENT)
            changed_task = np.array(changed_task)

            changed_task[2] = new_actual_time
            tasks_data = np.append(tasks_data, [changed_task], axis=0)

        _, planned_time, actual_time = np.array_split(tasks_data, 3, axis=1)
        actual_time = np.ravel(actual_time)

        # Определение и сохранение модели
        model_params = self.get_params(len(tasks_data))
        gb_model = GradientBoostingRegressor(**model_params)
        gb_model.fit(planned_time, actual_time)

        self.db.save_model_and_task(UID, task_id, new_actual_time, gb_model, **kwargs)

    def make_predict(self, UID: int, PlannedTime: float) -> float:
        '''
        Загружает модель из бд и предсказывает итоговое время выполнения задачи
        '''
        if PlannedTime <= 0:
            raise PredicatorException(f"Invalid PlannedTime (must be greater than zero)", grpc.StatusCode.INVALID_ARGUMENT)
        try:
            model = self.db.load_model(UID)
        except ExceptionDB as e:
            if e.extra_info == ExceptionDB.NOT_FOUND:
                model = self.recalulate_model(UID)

        planned_time = np.reshape(np.array([PlannedTime]), shape=(-1, 1))
        predict = float(model.predict(planned_time)[0])

        print(predict)

        x_range = np.linspace(1, 72, 100).reshape(-1, 1)
        y_range_pred = model.predict(x_range)
        import matplotlib.pyplot as plt
        plt.plot(x_range, y_range_pred, color="orange", label="Модель Gradient Boosting", linewidth=2)
        plt.xlabel("Оценочное время (X)")
        plt.ylabel("Фактическое время (Y)")
        plt.legend()
        plt.grid(True)
        plt.savefig('foo.png')

        return predict