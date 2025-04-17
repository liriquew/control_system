#!/bin/bash

# Создаем папку для логов
mkdir -p logs

# Объявляем массив для хранения PID процессов
declare -A PIDS

# Функция для остановки сервисов
stop_services() {
    echo "Stop..."
    for pid in "${PIDS[@]}"; do
        kill $pid
    done
    exit 0
}

# Перехватываем Ctrl+C
trap stop_services SIGINT

# Запускаем сервисы
start_service() {
    local name=$1
    local cmd=$2
    local log="logs/${name}.log"
    
    echo "Start $name..."
    eval "$cmd &"
    sleep 1
    PIDS[$name]=$!
}

# Python-сервис
start_service "predictions" "cd predictions_service && source venv/bin/activate && python main.py"

# Go-сервисы
start_service "auth" "cd auth_service && go build -o auth_service ./cmd/main.go && ./auth_service"
start_service "graphs" "cd graphs_service && go build -o graphs_service ./cmd/main.go && ./graphs_service"
start_service "groups" "cd groups_service && go build -o groups_service ./cmd/main.go && ./groups_service"
start_service "tasks" "cd tasks_service && go build -o tasks_service ./cmd/main.go && ./tasks_service"
start_service "api_gateway" "cd api && go build -o api ./cmd/main.go && ./api"

echo "Press Ctrl+C to stop all"

while true; do
    sleep 1
done
