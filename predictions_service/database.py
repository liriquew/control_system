import psycopg2
import numpy as np

from predicator import PredictInfo
import psycopg2.extras

from predicator import Tag, PredictInfo

class ExceptionDB(Exception):
    NOT_FOUND = 0
    
    def __init__(self, message, extra_info):
        super().__init__(message)
        self.extra_info = extra_info

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
 

    def load_model(self, UID: int) -> bytes:
        """
        Загружает созданную ранее модель пользователя
        """
        print("database.Database.load_model()")
        try:
            cursor = self.conn.cursor()
            cursor.execute("SELECT model, is_active FROM models WHERE user_id=%s", (UID,))
            result = cursor.fetchone()
            
            if result is None:
                raise ExceptionDB(f"model for UID={UID} not found", ExceptionDB.NOT_FOUND)

            if not result[1]:
                return None

            return result[0]
        except Exception as e:
            print(f"database.Database.load_model(): Error while loading model: {e}")
            raise e
        finally:
            cursor.close()


    def save_model(self, UID: int, model_blob: bytes):
        """
        Сохраняет модель и задачу пользователя в базу данных
        """
        print("database.Database.save_model()")
        try:
            cursor = self.conn.cursor()
            
            cursor.execute("""
                INSERT INTO models (user_id, model, is_active) 
                VALUES (%s, %s, %s) 
                ON CONFLICT (user_id) DO UPDATE 
                SET model = EXCLUDED.model, is_active=true""",
                (UID, psycopg2.Binary(model_blob), True)
            )

            self.conn.commit()
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.save_model(): Error while saving model: {e}")
            raise e
        finally:
            cursor.close()


    def delete_model(self, UID: int):
        """
        Удаляет запись с моделью пользователя 
        (необходимо в случае, если задач нет, т.е. все удалены, и дальнейшие прогнозы не нужны)
        """
        print("database.Database.delete_model()")
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
        
        
    def get_user_tasks(self, UID:int) -> list[PredictInfo]:
        """
        Выбирает все задачи пользователя, 
        которые имеют известное действительное время выполнения
        """
        print("database.Database.get_user_tasks()")
        try:
            cursor = self.conn.cursor(cursor_factory=psycopg2.extras.DictCursor)
            cursor.execute("""
                SELECT id, planned_time, actual_time, tags 
                FROM tasks t 
                WHERE user_id=%s AND actual_time IS NOT NULL
            """,
                (UID,)
            )
            result = cursor.fetchall()
            cursor.close()

            if not result:
                print(f"database.Database.get_user_tasks(): User UID: {UID} doesn't have completed tasks")
                return []

            infos = []
            for info in result:
                infos.append(PredictInfo.from_dict(dict(info.items()), UID))

            return infos
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.get_user_tasks(): Error while retrieving user tasks: {e}")
            raise e
        finally:
            cursor.close()

    
    def save_task_prediction_data(self, prediction_data: PredictInfo):
        """
        kafka consumer method
        """
        try:
            cursor = self.conn.cursor()
            
            cursor.execute("""
                INSERT INTO tasks (id, user_id, planned_time, actual_time, tags) 
                VALUES (%s, %s, %s, %s, %s) 
                ON CONFLICT (id) DO UPDATE 
                SET planned_time=EXCLUDED.planned_time, actual_time=EXCLUDED.actual_time, tags=EXCLUDED.tags""",
                (prediction_data.id, prediction_data.uid, 
                 prediction_data.planned_time, prediction_data.actual_time, prediction_data.tags)
            )

            self.conn.commit()
            print("task data saved")
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.save_task_prediction_data(): Error while saving task data: {e}")
            raise e
        finally:
            cursor.close()


    def delete_task_prediction_data(self, task_id: int):
        """
        kafka consumer method
        """
        try:
            cursor = self.conn.cursor()
            
            cursor.execute("""
                DELETE FROM tasks WHERE id=%s""",
                (task_id,)
            )

            self.conn.commit()
            print("task data deleted")
        except Exception as e:
            self.conn.rollback()
            print(f"database.Database.save_task_prediction_data(): Error while deleting task data: {e}")
            raise e
        finally:
            cursor.close()
