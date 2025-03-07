import psycopg2
import numpy as np
import pickle

class ExceptionDB(Exception):
    def __init__(self, message, extra_info):
        super().__init__(message)
        self.extra_info = extra_info

    NOT_FOUND = 0

class Database():
    """
    Класс для работы с базой данных PostgreSQL
    """
    def __init__(self, config: dict):
        """
        Подключение к базе данных
        """
        self.conn = psycopg2.connect(
            dbname=config["db_name"],
            user=config["username"],
            password=config["password"],
            host=config["host"]
        )

    def load_model(self, UID: int):
        '''
        Загружает созданную ранее модель пользователя
        '''
        try:
            cursor = self.conn.cursor()
            cursor.execute("SELECT model FROM models WHERE user_id = %s", (UID,))
            result = cursor.fetchone()
            
            if result is None:
                raise ExceptionDB(f"model for UID={UID} not found", ExceptionDB.NOT_FOUND)

            # десериализация бинарных данных в объект модели
            model = pickle.loads(result[0])
            return model
        except Exception as e:
            print(f"database.Database.load_model(): Error while loading model: {e}")
            raise e
        finally:
            cursor.close()

    def save_model(self, UID: int, model):
        """
        Сохраняет модель и задачу пользователя в базу данных
        """
        try:
            model_binary = pickle.dumps(model)

            cursor = self.conn.cursor()
            
            # save model
            cursor.execute("""
                INSERT INTO models (user_id, model) 
                VALUES (%s, %s) 
                ON CONFLICT (user_id) DO UPDATE 
                SET model = EXCLUDED.model RETURNING id""",
                (UID, psycopg2.Binary(model_binary))
            )
            model_id = cursor.fetchone()[0]

            self.conn.commit()
            print(f"model saved id: {model_id}")
            return model_id
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.save_model(): Error when saving model: {e}")
            raise e
        finally:
            cursor.close()

    def delete_model(self, UID: int):
        """
        Удаляет запись с моделью пользователя 
        (необходимо в случае, если задач нет, т.е. все удалены, и дальнейшие прогнозы не нужны)
        """
        try:
            cursor = self.conn.cursor()
            cursor.execute("""
                DELETE FROM models WHERE user_id=%s
                """,
                (UID, ),
            )

            self.conn.commit()
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.delete_model(): Error when deleting model: {e}")
            raise e
        finally:
            cursor.close()
        

    def save_model_and_task(self, UID: int, task_id: int, new_actual_time: float, model, **kwargs):
        """
        Сохраняет модель и задачу пользователя в базу данных
        """
        try:
            model_binary = pickle.dumps(model)

            cursor = self.conn.cursor()
            
            # save model
            cursor.execute("""
                INSERT INTO models (user_id, model) 
                VALUES (%s, %s) 
                ON CONFLICT (user_id) DO UPDATE 
                SET model = EXCLUDED.model RETURNING id""",
                (UID, psycopg2.Binary(model_binary))
            )
            model_id = cursor.fetchone()[0]

            # save task
            map_names = {
                'Title': 'title',
                'Description': 'description',
                'PlannedTime': 'planned_time',
            }
            cursor = self.conn.cursor()

            set_parts = [(f"{map_names[k]} = %s", v) for k, v in kwargs.items()]
            set_clause = ", ".join([sp[0] for sp in set_parts] + ['actual_time = %s'])

            query = f"UPDATE tasks SET {set_clause} WHERE id = %s AND user_id = %s"
            values = [sp[1] for sp in set_parts] + [new_actual_time, task_id, UID]
            cursor.execute(query, values)

            self.conn.commit()
            print(f"model saved id: {model_id}")
            return model_id
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.save_model_and_task(): Error when saving model and users`s task: {e}")
            raise e
        finally:
            cursor.close()
    
    def get_task(self, UID: int, task_id: int) -> tuple[int, float, float]:
        '''
        Возвращает задачу пользователя по id из базы данных
        '''
        try:

            cursor = self.conn.cursor()
            cursor.execute(
                "SELECT id, planned_time, actual_time FROM tasks WHERE id=%s AND user_id=%s",
                (task_id, UID)
            )
            result = cursor.fetchone()
            cursor.close()
            
            if not result:
                print(f"Task with task_id: {task_id} not found")
                return None

            return result
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.get_task(): Error collecting task from db: {e}")
        finally:
            cursor.close()

    def get_user_tasks(self, UID:int) -> list[tuple[int, float, float]]:
        '''
        Выбирает все задачи пользователя, 
        которые имеют известное действительное время выполнения
        '''
        try:
            cursor = self.conn.cursor()
            cursor.execute(
                "SELECT id, planned_time, actual_time FROM tasks WHERE user_id=%s AND actual_time IS NOT NULL",
                (UID,)
            )
            result = cursor.fetchall()
            cursor.close()

            if not result:
                print(f"database.Database.get_user_tasks(): User UID: {UID} doesn't have completed tasks")
                return np.array([]), np.array([])
            
            return result
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.get_user_tasks(): Error while retrieving user tasks: {e}")
            raise e
        finally:
            cursor.close()