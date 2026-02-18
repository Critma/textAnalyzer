# TextAnalyzer

## Описание проекта

Проект TextAnalyzer состоит из двух основных компонентов:
- **receiver** - сервис, api-gateway, принимает текстовые данные, возвращает статус запроса
- **analyzer** - сервис, анализирующий тексты, отправленые от **receiver**

## Инструкция по запуску (docker-compose)
1. Создайте файл .env в корне, с переменными из exapmples/.example.env (или переименуйте)
1. Поменяйте значения в .env и yml, или оставьте по умолчанию
1. Для запуска проекта используйте команду

    ```bash
    docker-compose up --build
    ```
1. Запуск интеграционных тестов
    ```bash
    go test integration_tests/*.go
    ```

## Примеры использования
- Получить анализ текста
    1. Отправить текст для анализа:
        ```bash
        curl -X POST http://localhost:8080/api/v1/text -d '{"text": "Text to analyze"}'
        ```
    1. Получить статус анализа по id, {id} - замените на полученный из 1 шага.
        ```bash
        curl -X GET http://localhost:8080/api/v1/status/{id}
        ```

- Проверить подняты ли сервисы
    - Receiver
        ```bash
        curl -X GET http://localhost:8080/api/v1/health
        ```
    - Analyzer
        ```bash
        curl -X GET http://localhost:8081/api/v1/health
        ```

- Другое
    - Swagger receiver: http://localhost:8080/api/v1/swagger/index.html
    - Postan коллекция в examples, с запросами к сервисам

## Архитектура

- Межсервисное взаимодействие:
    1. POST с текстом от клиента поступает на receiver
    1. Receiver генерирует uuid, отправляет Post с текстом и id на analyzer
    1. Analyzer принимает запрос, посылает его в worker pool отправляет ответ receiver.
    1. Worker pool производит конкуретную обработку всех запросов, обработов задачу отправляет Post на receiver с результатом.
    1. Клиент в любое время может проверить статус по GET с uuid.

- Архитектура проект имеет модульную архитектуру с разделением на:
    - cmd - точка входа приложения
    - internal - внутренние пакеты приложения
        - analyze - анализ текста, воркеры. (analyzer)
        - config - конфигурация сервисов, загрузка env
        - models - структуры данных rest (analyzer)
        - routes - маршруты и обработчики HTTP-запросов
        - metrics - сбор и экспорт метрик
        - cache - кеширование данных (Redis)
        - store - хранилище данных, паттерн Repository (receiver)
            - local - In-memory хранение данных - реализация Store

## Используемые технологии

- Golang
- Docker Compose
- Swagger - swagger docs
- Redis - cache
- Prometheus - metrics
- Gin (WEB-фреймворк), CORS
- Zerolog - json logger
- google/uuid - uniq id