#!/bin/bash

# Проверяем, передан ли ID контейнера
if [ -z "$1" ]; then
    echo "Usage: $0 <container_id>"
    exit 1
fi

CONTAINER_ID=$1

echo "Обработка контейнера: $CONTAINER_ID"

# 1. Отключаем автоперезапуск
sudo docker update --restart=no "$CONTAINER_ID" 2>/dev/null

# 2. Получаем PID контейнера
PID=$(sudo docker inspect "$CONTAINER_ID" --format '{{.State.Pid}}' 2>/dev/null)

if [ -z "$PID" ] || [ "$PID" -eq 0 ]; then
    echo "Контейнер не запущен или не имеет PID. Удаляем принудительно..."
    sudo docker rm -f "$CONTAINER_ID"
    exit 0
fi

echo "PID: $PID"

# 3. Находим путь до cgroup
CGROUP_PATH=$(cat /proc/$PID/cgroup 2>/dev/null | grep -o '/system.slice/docker-.*.scope' | head -1)

if [ -n "$CGROUP_PATH" ]; then
    echo "Cgroup путь: $CGROUP_PATH"
    echo "1" | sudo tee "/sys/fs/cgroup$CGROUP_PATH/cgroup.kill" > /dev/null
else
    echo "Не удалось найти cgroup путь. Убиваем процесс напрямую..."
    sudo kill -9 "$PID" 2>/dev/null
fi

# 4. Удаляем контейнер
sudo docker rm "$CONTAINER_ID" 2>/dev/null || sudo docker rm -f "$CONTAINER_ID" 2>/dev/null

echo "Контейнер $CONTAINER_ID удалён"
