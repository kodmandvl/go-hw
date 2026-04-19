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

#### Быстрый старт (состояние на момент ДЗ №13)

Как пользователю запустить приложение и познакомиться с функционалом.

- 1) Build:

```bash
make build
```

- 2) В конфигурационном файле задать `storage.type`

Если БД не требуется, то можно взять конфигурационый файл по умолчанию.

Если БД нужна (рассмотрим этот вариант), то `storage.type` => `sql`.

- 3) Запуск БД, например, вот так в Podman:

```bash
make run-postgres-16-in-podman
# Проверяем:
podman logs postgres-calendar
podman exec -it postgres-calendar pg_isready
```

- 4) Запустить (если берем конфиг по умолчанию, указывать конфигурационный файл нам необязательно):

```bash
# Через make run:
make run
# Или:
./bin/calendar -c ./configs/config.yaml
# Или:
./bin/calendar --config ./configs/config.yaml
```

- 5) Проверка gRPC:

```bash
curl -v --http2-prior-knowledge http://127.0.0.1:50051/
```

- 6) Добавление событий:

```bash
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"День рождения","date_time":"2026-04-18T19:30:00Z","duration":7200,"description":"Только раз в году","user_id":"1"}' | jq
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"Митап","date_time":"2026-04-18T11:30:00Z","duration":1800,"description":"","user_id":"2"}' | jq
curl -s -X POST http://localhost:8080/api/events -H 'Content-Type: application/json' -d '{"title":"Встреча","date_time":"2026-04-18T10:00:00Z","duration":3600,"description":"","user_id":"1"}' | jq
```

- 7) Просмотр событий, примеры:

```bash
curl -s "http://localhost:8080/api/events/day?day=2026-04-18&user_id=1" | jq
curl -s "http://localhost:8080/api/events/day?day=2026-04-18&user_id=2" | jq
curl -s "http://localhost:8080/api/events/day?day=2026-04-18" | jq
```

( или в браузере открыть http://localhost:8080/api/events/day?day=2026-04-18 )

- 8) Просмотр записей событий в БД (если запустили со `storage.type` => `sql`):

```bash
podman exec -it postgres-calendar psql -Upostgres -dbackend -c "select title,date_time,duration,description,user_id from event order by 2;"
```

Вывод:

```
$ podman exec -it postgres-calendar psql -Upostgres -dbackend -c "select title,date_time,duration,description,user_id from event order by 2;"
     title     |       date_time        | duration |    description    | user_id 
---------------+------------------------+----------+-------------------+---------
 Встреча       | 2026-04-18 10:00:00+00 |     3600 |                   |       1
 Митап         | 2026-04-18 11:30:00+00 |     1800 |                   |       2
 День рождения | 2026-04-18 19:30:00+00 |     7200 | Только раз в году |       1
(3 rows)
```