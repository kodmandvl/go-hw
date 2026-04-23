# Результатом выполнения следующих домашних заданий является сервис «Календарь»:
- [Домашнее задание №12 «Заготовка сервиса Календарь»](./docs/12_README.md)
- [Домашнее задание №13 «Внешние API от Календаря»](./docs/13_README.md)
- [Домашнее задание №14 «Кроликизация Календаря»](./docs/14_README.md)
- [Домашнее задание №15 «Докеризация и интеграционное тестирование Календаря»](./docs/15_README.md)

# Ветки при выполнении
- `hw12_calendar` (от `master`) -> Merge Request в `master`
- `hw13_calendar` (от `hw12_calendar`) -> Merge Request в `hw12_calendar` (если уже вмержена, то в `master`)
- `hw14_calendar` (от `hw13_calendar`) -> Merge Request в `hw13_calendar` (если уже вмержена, то в `master`)
- `hw15_calendar` (от `hw14_calendar`) -> Merge Request в `hw14_calendar` (если уже вмержена, то в `master`)
- `hw16_calendar` (от `hw15_calendar`) -> Merge Request в `hw15_calendar` (если уже вмержена, то в `master`)


**Домашнее задание не принимается, если не принято ДЗ, предшествующее ему.**

# Быстрый старт (состояние на момент ДЗ №14)

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

# Быстрый старт (состояние на момент ДЗ №15)

Для ДЗ №15 добавлены:

- `deployments/docker-compose.yaml` — поднимает PostgreSQL, RabbitMQ, `calendar`, `calendar_scheduler`, `calendar_sender`;
- `deployments/docker-compose.integration.yaml` — поднимает окружение + контейнер с интеграционными тестами;
- `make up`, `make down`, `make integration-tests`.

Важно: `make integration-tests` использует Docker Compose, поэтому должен быть запущен Docker daemon.

- 1) Поднять все сервисы в docker:

```bash
make up
```

Проверка API:

```bash
curl -s http://localhost:8888/
# ожидается: calendar ok
```

Статус и логи:

```bash
docker compose -f ./deployments/docker-compose.yaml ps
docker compose -f ./deployments/docker-compose.yaml logs postgres -f
docker compose -f ./deployments/docker-compose.yaml logs rabbitmq -f
docker compose -f ./deployments/docker-compose.yaml logs calendar -f
docker compose -f ./deployments/docker-compose.yaml logs scheduler -f
docker compose -f ./deployments/docker-compose.yaml logs sender -f
```

Проверка создания событий:

```bash
curl -s -X POST http://localhost:8888/api/events -H 'Content-Type: application/json' -d '{"title":"День рождения","date_time":"2026-04-18T19:30:00Z","duration":7200,"description":"Только раз в году","user_id":"1"}' | jq
curl -s -X POST http://localhost:8888/api/events -H 'Content-Type: application/json' -d '{"title":"Митап","date_time":"2026-04-18T11:30:00Z","duration":1800,"description":"","user_id":"2"}' | jq
curl -s -X POST http://localhost:8888/api/events -H 'Content-Type: application/json' -d '{"title":"Встреча","date_time":"2026-04-18T10:00:00Z","duration":3600,"description":"","user_id":"1"}' | jq
curl -s -X POST http://localhost:8888/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"Тест напоминания",
    "date_time":"2030-12-31T15:00:00Z",
    "duration":3600,
    "description":"",
    "user_id":"3",
    "time_notification":"2020-01-01T12:00:00Z"
  }' | jq
curl -s -X POST http://localhost:8888/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "title":"UPD 15",
    "date_time":"2031-01-01T10:00:00Z",
    "duration":3600,
    "description":"№15 с docker-compose",
    "user_id":"4"
  }' | jq
```

Смотрим события:

```bash
curl -s "http://localhost:8888/api/events/day?day=2026-04-18" | jq
curl -s "http://localhost:8888/api/events/day?day=2030-12-31" | jq
curl -s "http://localhost:8888/api/events/day?day=2031-01-01" | jq
```

Смотрим в БД:

```
$ docker exec -it calendar-postgres psql -U postgres -d backend -c "select * from event order by date_time, user_id;"
                  id                  |      title       |       date_time        | duration |     description      | user_id | notification_time |         created_at         
--------------------------------------+------------------+------------------------+----------+----------------------+---------+-------------------+----------------------------
 db514eb7-d32d-4643-a0a5-c1b61dd30b8c | Встреча          | 2026-04-18 10:00:00+00 |     3600 |                      |       1 |                   | 2026-04-22 22:34:42.357188
 a0f001d9-0889-4843-ab7e-10d8ed9608c0 | Митап            | 2026-04-18 11:30:00+00 |     1800 |                      |       2 |                   | 2026-04-22 22:34:42.326307
 bae7829b-1f67-42ef-817c-ba9b98ee2532 | День рождения    | 2026-04-18 19:30:00+00 |     7200 | Только раз в году    |       1 |                   | 2026-04-22 22:33:53.806586
 ad7822e6-b7be-4f7e-b5fe-ff547c5e494c | Тест напоминания | 2030-12-31 15:00:00+00 |     3600 |                      |       3 |                   | 2026-04-22 22:35:18.087817
 aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa | UPD 15           | 2031-01-01 10:00:00+00 |     3600 | №15 с docker-compose |       4 |                   | 2026-04-22 22:35:25.796707
(5 rows)
```

Смотрим статус очередей :

```
$ docker exec -it calendar-rabbitmq rabbitmqctl list_queues name messages messages_ready messages_unacknowledged
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
name    messages        messages_ready  messages_unacknowledged
calendar_notifications_status   5       5       0
calendar_notifications  0       0       0
```

Посмотреть сообщения из новой очереди calendar_notifications_status:

```bash
curl -s -u rabbit:password -X POST "http://localhost:15672/api/queues/%2F/calendar_notifications_status/get" \
  -H "Content-Type: application/json" \
  -d '{"count":5,"ackmode":"ack_requeue_true","encoding":"auto","truncate":50000}' | jq
```

- 2) Остановить и удалить окружение:

```bash
make down
```

- 3) Запустить интеграционные тесты:

```bash
make integration-tests
docker compose -f ./deployments/docker-compose.integration.yaml ps | egrep -v '(^NAME)' | wc -l
```

Команда:

- поднимает окружение из `deployments/docker-compose.integration.yaml`,
- запускает тесты из `tests/integration`,
- удаляет окружение за собой (через `down -v --remove-orphans`).

Сценарии интеграционных тестов:

- создание события + проверка бизнес-ошибки (конфликт),
- листинг событий на день/неделю/месяц,
- доставка уведомления: scheduler -> rabbitmq -> sender -> очередь статусов `calendar_notifications_status`.

- 4) Сборка образов и пуш в реджистри (на будущее для ДЗ №16):

```bash
# Сборка образов:
make build-img
docker images | grep kodmandvl[/]calendar
# Пуш:
docker push docker.io/kodmandvl/calendar-api:develop
docker push docker.io/kodmandvl/calendar-scheduler:develop
docker push docker.io/kodmandvl/calendar-sender:develop
```

Указанные образы можно увидеть здесь: https://hub.docker.com/search?q=kodmandvl&sort=updated_at&order=desc

# Быстрый старт (состояние на момент ДЗ №16)

Добавлен **Helm chart** в каталоге `calendar-chart/`

## Что входит в chart

- **Deployment** и **Service** для трёх процессов: HTTP/gRPC API (`calendar-api`), планировщик (`calendar-scheduler`), рассыльщик (`calendar-sender`).
- **Ingress** (по умолчанию выключён) — если включен, наружу выводится `calendar-api` (планировщик и рассыльщик работают внутри кластера).
- **ConfigMap** с YAML-конфигами, как в docker-сценарии: приложения читают `/app/configs/*.yaml` из смонтированного тома.

## Что в chart не входит

- PostgreSQL и RabbitMQ **не** создаются этим chart-ом.

Подразумевается, что это либо развёрнутые сервисы в кластере (вручную или с помощью операторов), либо это внешние сервисы за пределами кластера. Соответствено, подразумевается, что это уже существующие сервисы и нужно отредактировать values, связанные с подключением к этим сервисам.

В нашем примере к ним будем обращаться так (это же и значения по умолчанию):

  - БД: `postgresql-server.postgres.svc.cluster.local`
  - RabbitMQ: `rabbitmq-server.rabbitmq.svc.cluster.local`

Учётные данные PostgreSQL и RabbitMQ (`database.*`, `rabbitmq.*` в `values.yaml`) можно переопределить при установке (`--set` или со своим файлом values).

Очереди в RabbitMQ будут созданы после запуска приложения, а вот целевую прикладную БД и пользователя в PostgreSQL нам нужно предварительно создать, например:

```bash
# Удалить БД и пользователя (если пересоздаём для чистого деплоя):
kubectl exec -it -n postgres svc/postgresql-server -- psql -Upostgres -c "drop database backend;"
kubectl exec -it -n postgres svc/postgresql-server -- psql -Upostgres -c "drop user backend;"
# Создать пользователя и БД:
kubectl exec -it -n postgres svc/postgresql-server -- psql -Upostgres -c "create user backend with encrypted password 'password';"
kubectl exec -it -n postgres svc/postgresql-server -- psql -Upostgres -c "create database backend owner backend;"
# Проверить подключение (как бы удалённое):
kubectl exec -it -n postgres svc/postgresql-server -- psql postgres://backend:password@postgresql-server.postgres.svc.cluster.local:5432/backend
```

## K8s кластер

Я использовал свой домашний тестовый одноузловой кластер K8s `v1.30.1` на ВМ `Rocky Linux 8.8`, который в свое время инициализировал с помощью `kubeadm`.

В кластере уже есть `ingress-nginx-controller` (с как бы внешним IP-адресом `172.17.255.1` через `MetalLB`), `cert-manager` (и `clusterissuer` `myca`).

Также в кластере в качестве как бы внешних сервисов запущены простенький StefulSet+Service для PostgreSQL и простенький Deployment+Service для RabbitMQ (+ Ingress для UI).

## Проверка без установки

Проверку чарта и генерацию им манифестов можно выполнить так:

```bash
make helm-lint
make helm-template
```

Также можно посмотреть, какие манифесты сгенерируются при дефолтных values и при моих кастомных значениях.

```bash
# Вот так можно посмотреть, какие ресурсы будут созданы при деплое с дефолтными values:
helm template calendar -n calendar ./calendar-chart
# А вот так с нашими кастомными values:
helm template calendar -n calendar ./calendar-chart -f ./calendar-my-custom-values.yaml
# В частности, у меня изменён schedulerConfig.intervalSeconds и изменены параметры ingress.*
# Остальные параметры, которые не меняем, можно было бы удалить из кастомного файла, но оставил для наглядности.
```

## Упаковка чарта

Готовый чарт также можно упаковать (например, для дальнейшей отправки в реджистри чартов):

```bash
make helm-package
```

## Деплой

Теперь деплоим наш Calendar:

```bash
helm upgrade --install calendar -n calendar --create-namespace ./calendar-chart -f ./calendar-my-custom-values.yaml
```

## Проверка состояния

Смотрим статус и логи приложений:

```bash
helm ls -n calendar
kubectl get all,secrets,cm,ing -n calendar
kubectl -n calendar logs deployments/calendar-calendar-api
kubectl -n calendar logs deployments/calendar-calendar-scheduler
kubectl -n calendar logs deployments/calendar-calendar-sender
```

Проверка API (через curl или в браузере):

```bash
# Через ингресс:
curl -k https://calendar.172.17.255.1.nip.io
# Без ингресса:
nohup kubectl port-forward -n calendar svc/calendar-calendar-api 8080:8080 &
curl http://127.0.0.1:8080
```

(ожидается: `calendar ok`)

Далее будем через ингресс обращаться к API.

## Проверка создания событий

Создадим те же события, что и в прошлом ДЗ + еще одно:

```bash
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events -H 'Content-Type: application/json' -d '{"title":"День рождения","date_time":"2026-04-18T19:30:00Z","duration":7200,"description":"Только раз в году","user_id":"1"}' | jq
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events -H 'Content-Type: application/json' -d '{"title":"Митап","date_time":"2026-04-18T11:30:00Z","duration":1800,"description":"","user_id":"2"}' | jq
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events -H 'Content-Type: application/json' -d '{"title":"Встреча","date_time":"2026-04-18T10:00:00Z","duration":3600,"description":"","user_id":"1"}' | jq
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"Тест напоминания",
    "date_time":"2030-12-31T15:00:00Z",
    "duration":3600,
    "description":"",
    "user_id":"3",
    "time_notification":"2020-01-01T12:00:00Z"
  }' | jq
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "title":"UPD 15",
    "date_time":"2031-01-01T10:00:00Z",
    "duration":3600,
    "description":"№15 с docker-compose",
    "user_id":"4"
  }' | jq
curl -k -s -X POST https://calendar.172.17.255.1.nip.io/api/events \
  -H 'Content-Type: application/json' \
  -d '{
    "title":"UPD 16",
    "date_time":"2026-04-23T22:30:00Z",
    "duration":72000,
    "description":"№16 с K8s и Helm",
    "user_id":"4"
  }' | jq
```

Смотрим события (через curl или в браузере):

```bash
curl -k -s "https://calendar.172.17.255.1.nip.io/api/events/day?day=2026-04-18" | jq
curl -k -s "https://calendar.172.17.255.1.nip.io/api/events/day?day=2030-12-31" | jq
curl -k -s "https://calendar.172.17.255.1.nip.io/api/events/day?day=2031-01-01" | jq
curl -k -s "https://calendar.172.17.255.1.nip.io/api/events/day?day=2026-04-23" | jq
```

Смотрим в БД:

```bash
kubectl exec -it -n postgres svc/postgresql-server -- psql postgres://backend:password@postgresql-server.postgres.svc.cluster.local:5432/backend -c "select * from event order by date_time, user_id;"
```

Т.к. schedulerConfig.intervalSeconds увеличил, можно успеть увидеть непустой столбец notification_time до обработки:

```
$ kubectl exec -it -n postgres svc/postgresql-server -- psql postgres://backend:password@postgresql-server.postgres.svc.cluster.local:5432/backend -c "select * from event order by date_time, user_id;"
Defaulted container "postgres" out of: postgres, change-pgdata-owner (init)
                  id                  |      title       |       date_time        | duration |     description      | user_id |   notification_time    |         created_at         
--------------------------------------+------------------+------------------------+----------+----------------------+---------+------------------------+----------------------------
 90e46ece-7329-4009-9568-df60a4dd93aa | Встреча          | 2026-04-18 10:00:00+00 |     3600 |                      |       1 | 0001-01-01 00:00:00+00 | 2026-04-23 14:04:28.920859
 79a3cb0a-cd87-4de9-be67-05174848d307 | Митап            | 2026-04-18 11:30:00+00 |     1800 |                      |       2 | 0001-01-01 00:00:00+00 | 2026-04-23 14:04:28.901348
 a70c99d3-01d2-4a28-945a-49175a287198 | День рождения    | 2026-04-18 19:30:00+00 |     7200 | Только раз в году    |       1 | 0001-01-01 00:00:00+00 | 2026-04-23 14:04:28.882242
 c8f0fd41-9902-4807-a930-ca2ee053142a | UPD 16           | 2026-04-23 22:30:00+00 |    72000 | №16 с K8s и Helm     |       4 | 0001-01-01 00:00:00+00 | 2026-04-23 14:04:28.982419
 4f2911da-0fc0-4d9e-a61c-f10b1946dcf0 | Тест напоминания | 2030-12-31 15:00:00+00 |     3600 |                      |       3 | 2020-01-01 12:00:00+00 | 2026-04-23 14:04:28.942182
 aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa | UPD 15           | 2031-01-01 10:00:00+00 |     3600 | №15 с docker-compose |       4 | 0001-01-01 00:00:00+00 | 2026-04-23 14:04:28.962583
(6 rows)
```

Смотрим статус очередей:

```
kubectl exec -it -n rabbitmq svc/rabbitmq-server -- rabbitmqadmin --host rabbitmq-server.rabbitmq.svc.cluster.local --port 15672 --username rabbit --password password list queues
kubectl exec -it -n rabbitmq svc/rabbitmq-server -- rabbitmqctl list_queues name messages messages_ready messages_unacknowledged
```

```
$ kubectl exec -it -n rabbitmq svc/rabbitmq-server -- rabbitmqadmin --host rabbitmq-server.rabbitmq.svc.cluster.local --port 15672 --username rabbit --password password list queues
╭───────────────────────────────┬───────┬────────────┬─────────┬─────────────┬───────────┬─────────────────────────┬─────────────────────────────────────────┬─────────┬────────┬─────────┬────────┬────────┬────────────────┬──────────────────────┬────────┬───────────────┬───────────────┬──────────────────────────────╮
│ name                          │ vhost │ queue_type │ durable │ auto_delete │ exclusive │ arguments               │ node                                    │ state   │ leader │ members │ online │ memory │ consumer_count │ consumer_utilisation │ policy │ message_bytes │ message_count │ unacknowledged_message_count │
├───────────────────────────────┼───────┼────────────┼─────────┼─────────────┼───────────┼─────────────────────────┼─────────────────────────────────────────┼─────────┼────────┼─────────┼────────┼────────┼────────────────┼──────────────────────┼────────┼───────────────┼───────────────┼──────────────────────────────┤
│ calendar_notifications        │ /     │ classic    │ true    │ false       │ false     │ x-queue-type: "classic" │ rabbit@rabbitmq-server-7b8b4c48c6-jkt7p │ running │        │         │        │ 42888  │ 1              │ 1                    │        │ 0             │ 0             │ 0                            │
│ calendar_notifications_status │ /     │ classic    │ true    │ false       │ false     │ x-queue-type: "classic" │ rabbit@rabbitmq-server-7b8b4c48c6-jkt7p │ running │        │         │        │ 42736  │ 0              │ 0                    │        │ 1502          │ 6             │ 0                            │
╰───────────────────────────────┴───────┴────────────┴─────────┴─────────────┴───────────┴─────────────────────────┴─────────────────────────────────────────┴─────────┴────────┴─────────┴────────┴────────┴────────────────┴──────────────────────┴────────┴───────────────┴───────────────┴──────────────────────────────╯
$ kubectl exec -it -n rabbitmq svc/rabbitmq-server -- rabbitmqctl list_queues name messages messages_ready messages_unacknowledged
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
name    messages        messages_ready  messages_unacknowledged
calendar_notifications_status   6       6       0
calendar_notifications  0       0       0
```

Посмотреть сообщения из новой очереди calendar_notifications_status через ингресс для RabbitMQ:

```bash
curl -k -s -u rabbit:password -X POST "https://rabbitmq.172.17.255.1.nip.io/api/queues/%2F/calendar_notifications_status/get" \
  -H "Content-Type: application/json" \
  -d '{"count":6,"ackmode":"ack_requeue_true","encoding":"auto","truncate":50000}' | jq
```

## Скриншоты

Скриншоты по состоянию приложений ниже.

* Кластер и ресурсы в неймспейсе calendar:

![pictures/k8s-and-app.png](pictures/k8s-and-app.png)

* Очередь calendar_notifications_status:

![pictures/rabbitmq.png](pictures/rabbitmq.png)

* События в браузере, пример:

![pictures/api.png](pictures/api.png)

* События в БД (но уже после обработки):

![pictures/postgres.png](pictures/postgres.png)

## Удаление

Если нужно удалить релиз из кластера:

```bash
helm uninstall -n calendar calendar
kubectl delete ns calendar
```

Также нужно будет удалить БД и пользователя в PostgreSQL и удалить созданные очереди в RabbitMQ.