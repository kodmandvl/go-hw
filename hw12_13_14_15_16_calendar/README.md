#### Результатом выполнения следующих домашних заданий является сервис «Календарь»:
- [Домашнее задание №12 «Заготовка сервиса Календарь»](./docs/12_README.md)
- [Домашнее задание №13 «Внешние API от Календаря»](./docs/13_README.md)
- [Домашнее задание №14 «Кроликизация Календаря»](./docs/14_README.md)
- [Домашнее задание №15 «Докеризация и интеграционное тестирование Календаря»](./docs/15_README.md)

#### Ветки при выполнении
- `hw12_calendar` (от `master`) -> Merge Request в `master`
- `hw13_calendar` (от `hw12_calendar`) -> Merge Request в `hw12_calendar` (если уже вмержена, то в `master`)
- `hw14_calendar` (от `hw13_calendar`) -> Merge Request в `hw13_calendar` (если уже вмержена, то в `master`)
- `hw15_calendar` (от `hw14_calendar`) -> Merge Request в `hw14_calendar` (если уже вмержена, то в `master`)
- `hw16_calendar` (от `hw15_calendar`) -> Merge Request в `hw15_calendar` (если уже вмержена, то в `master`)


**Домашнее задание не принимается, если не принято ДЗ, предшествующее ему.**

#### Быстрый старт (состояние на момент ДЗ №14)

Как пользователю запустить приложение и познакомиться с функционалом.

- 1) Сборка (три бинарника: API, планировщик, рассыльщик):

```bash
make build
```

- 2) Конфигурирование (в `configs`)

- 3) Запуск PostgreSQL, например в Podman:

```bash
make run-postgres
# Проверяем:
podman logs postgres-calendar
podman exec -it postgres-calendar pg_isready
```

- 4) Запуск RabbitMQ (в Makefile есть цель `run-rabbit` под Docker; учётные данные по умолчанию совпадают с `configs/scheduler_config.yaml` и `configs/sender_config.yaml`):

```bash
make run-rabbit
# Проверяем:
podman logs rabbitmq
```

- 5) Запуск приложений (в трех окнах терминала):

(для запуска в таком виде с микросервисами и с RabbitMQ нужно выбрать `storage.type` => `sql`, иначе просто со `storage.type` => `memory` запустить только лишь `./bin/calendar`)

```bash
make run-calendar
# или: ./bin/calendar -c ./configs/calendar_config.yaml
# или: ./bin/calendar --config ./configs/calendar_config.yaml
```

```bash
make run-scheduler
# или: ./bin/calendar_scheduler --config ./configs/scheduler_config.yaml
```

```bash
make run-sender
# или: ./bin/calendar_sender --config ./configs/sender_config.yaml
```

- 6) Проверка gRPC:

```bash
curl -v --http2-prior-knowledge http://127.0.0.1:50051/
```

- 7) Добавление событий:

```bash
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"День рождения","date_time":"2026-04-18T19:30:00Z","duration":7200,"description":"Только раз в году","user_id":"1"}' | jq
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"Митап","date_time":"2026-04-18T11:30:00Z","duration":1800,"description":"","user_id":"2"}' | jq
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"Встреча","date_time":"2026-04-18T10:00:00Z","duration":3600,"description":"","user_id":"1"}' | jq
```

В логах всех трёх приложений будет соответствующая информация при этом.

- 8) Проверка напоминаний

```bash
# Добавление события:
curl -s -X POST http://localhost:8080/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"Тест напоминания",
    "date_time":"2030-12-31T15:00:00Z",
    "duration":3600,
    "description":"",
    "user_id":"3",
    "time_notification":"2020-01-01T12:00:00Z"
  }' | jq
```

Дальше в логах можно посмотреть, как было запланировано напоминание и отправлено.

При этом в столбце БД столбец time_notification станет `NULL`.

В нашем тестовом случае уведомления быстро отправляются. Если нужно именно в RabbitMQ посмотреть, как сообщение было в очереди, то нужно перед добавлением события отключить `./bin/sender`, тогда можно будет успеть увидеть, как сообщение о напоминании находится в очереди в http://localhost:15672 в браузере или вот так:

```bash
# Состояние очереди:
podman exec -it rabbitmq rabbitmqctl list_queues name messages messages_ready messages_unacknowledged
# Посмотреть сообщения из очереди:
curl -s -u rabbit:password -X POST "http://localhost:15672/api/queues/%2F/calendar_notifications/get" \
  -H "Content-Type: application/json" \
  -d '{"count":5,"ackmode":"ack_requeue_true","encoding":"auto","truncate":50000}' | jq
```

А в БД можем посмотреть события и что после добавления в очередь у события становится `notification_time` => `NULL`).

Затем снова запускаем `./bin/sender`, сообщение (уведомление) отправляется и очередь снова пуста.

- 9) Просмотр событий, примеры:

```bash
curl -s "http://localhost:8080/api/events/day?day=2026-04-18&user_id=1" | jq
curl -s "http://localhost:8080/api/events/day?day=2026-04-18&user_id=2" | jq
curl -s "http://localhost:8080/api/events/day?day=2026-04-18" | jq
curl -s "http://localhost:8080/api/events/day?day=2030-12-31" | jq
```

(или в браузере: http://localhost:8080/api/events/day?day=2026-04-18)

- 10) Просмотр записей в БД при `storage.type: sql`:

```bash
podman exec -it postgres-calendar psql -Upostgres -dbackend -c "select id,title,date_time,duration,description,user_id,notification_time from event order by user_id,date_time;"
```

Пример вывода:

```
$ podman exec -it postgres-calendar psql -Upostgres -dbackend -c "select id,title,date_time,duration,description,user_id,notification_time from event order by user_id,date_time;"
                  id                  |      title       |       date_time        | duration |    description    | user_id | notification_time 
--------------------------------------+------------------+------------------------+----------+-------------------+---------+-------------------
 3cb1de3f-44e6-4513-99d2-4159115b3e17 | Встреча          | 2026-04-18 10:00:00+00 |     3600 |                   |       1 | 
 3b7b0bc2-5228-471c-82f4-84c89793120f | День рождения    | 2026-04-18 19:30:00+00 |     7200 | Только раз в году |       1 | 
 e7a5f8b2-9114-435a-88e9-28a056628448 | Митап            | 2026-04-18 11:30:00+00 |     1800 |                   |       2 | 
 3af22993-cac2-45a7-b8ad-e5a8fe2ab0ab | Тест напоминания | 2030-12-31 15:00:00+00 |     3600 |                   |       3 | 
(4 rows)
```