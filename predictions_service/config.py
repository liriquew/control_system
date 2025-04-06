import yaml
import os

class ConfigLoader:
    """
    Класс для загрузки конфигурации подключения к базе данных из YAML файла.
    """
    
    def __init__(self, config_path: str = 'config.yaml'):
        """
        Инициализация класса ConfigLoader.
        """
        self.config_path = "./config/config.yaml"
        if os.getenv("DOCKER_CONFIG") is not None:
            self.config_path = os.getenv("DOCKER_CONFIG")
        
        self.config = self.load_config()
    
    def load_config(self) -> dict:
        """
        Загрузка конфигурации из YAML файла.
        """
        if not os.path.exists(self.config_path):
            raise FileNotFoundError(f"Конфигурационный файл не найден: {self.config_path}")
        
        with open(self.config_path, 'r', encoding='utf-8') as file:
            try:
                config = yaml.safe_load(file)
                if not config:
                    raise ValueError("Конфигурационный файл пустой или имеет неправильный формат.")
                return config
            except yaml.YAMLError as e:
                raise ValueError(f"Ошибка при разборе YAML файла: {e}")
    
    def get_database_config(self) -> dict:
        """
        Возвращает параметры подключения к базе данных.
        """
        db_config = self.config.get('postgres')
        if not db_config:
            raise KeyError("В конфигурации отсутствует секция 'database'.")

        return db_config

    def get_service_config(self) -> dict:
        """
        Возвращает параметры подключения к базе данных.
        """
        app_config = self.config.get('service_config')
        if not app_config:
            raise KeyError("В конфигурации отсутствует секция 'service_config'.")

        return app_config
    
    def get_kafka_config(self) -> dict:
        """
        Возвращает параметры для брокера
        """
        app_config = self.config.get('kafka')
        if not app_config:
            raise KeyError("В конфигурации отсутствует секция 'kafka'.")

        return app_config