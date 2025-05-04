import json
import logging
import threading
import sys

from confluent_kafka import Consumer, KafkaException

from database import Database
from predicator import PredictInfo

class KafkaMLConsumer:
    def __init__(
        self,
        cfg: dict,
        db: Database,
    ):
        self._consumer = Consumer({
            "bootstrap.servers": cfg["bootstrap_servers"],
            "group.id": cfg["group_id"],
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True,
            "fetch.min.bytes": 1,         # Минимум данных для возврата
            "fetch.wait.max.ms": 100,     # Макс время ожидания новых данных
            "queued.min.messages": 10000,  # Минимум сообщений в локальной очереди
        })
        self._topic = cfg["topic"]
        
        self._consumer_delete = Consumer({
            "bootstrap.servers": cfg["bootstrap_servers"],
            "group.id": cfg["group_id"],
            "auto.offset.reset": "earliest",
            "enable.auto.commit": True
        })
        self._topic_delete = cfg["delete_topic"]
        
        self._db = db

        self._logger = logging.getLogger(__name__)
        self._logger.setLevel(logging.DEBUG)
        self._logger.addHandler(logging.StreamHandler(sys.stdout))

    def start(self):
        self._logger.info("Consumer started")
        self._thread_pred_data = threading.Thread(target=self.consume_prediction_data, daemon=True)
        self._thread_pred_data.start()

        self._thread_pred_data_delete = threading.Thread(target=self.consume_prediction_data_delete, daemon=True)
        self._thread_pred_data_delete.start()

    def consume_prediction_data(self):
        self._consumer.subscribe([self._topic])
        
        try:
            while True:
                msg = self._consumer.poll(0.001)
                
                if msg is None:
                    continue
                if msg.error():
                    raise KafkaException(msg.error())
                
                try:
                    data = json.loads(msg.value().decode("utf-8"))
                    self._logger.info(f"recieved message: {data}")
                    self._db.save_task_prediction_data(PredictInfo.from_dict(data))
                except json.JSONDecodeError as e:
                    self._logger.error(f"Invalid JSON: {e}")
                except Exception as e:
                    self._logger.error(f"Handler failed: {e}")

        except KeyboardInterrupt:
            self._logger.info("Consumer stopped by user")
        finally:
            self._consumer.close()

    def consume_prediction_data_delete(self):
        self._consumer_delete.subscribe([self._topic_delete])
        
        try:
            while True:
                msg = self._consumer_delete.poll(0.001)
                
                if msg is None:
                    continue
                if msg.error():
                    raise KafkaException(msg.error())
                
                try:
                    data = json.loads(msg.value().decode("utf-8"))
                    self._logger.info(f"recieved message: {data}")
                    self._db.delete_task_prediction_data(task_id=data["ID"])
                except json.JSONDecodeError as e:
                    self._logger.error(f"Invalid JSON: {e}")
                except Exception as e:
                    self._logger.error(f"Handler failed: {e}")

        except KeyboardInterrupt:
            self._logger.info("Consumer stopped by user")
        finally:
            self._consumer_delete.close()

