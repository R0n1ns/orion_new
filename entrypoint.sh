#!/bin/sh
# Устанавливаем переменную SERVICE_ADDRES равной значению HOSTNAME контейнера
export SERVICE_ADDRES=$HOSTNAME

# Запускаем приложение с передачей аргументов, если они есть
exec ./app "$@"
