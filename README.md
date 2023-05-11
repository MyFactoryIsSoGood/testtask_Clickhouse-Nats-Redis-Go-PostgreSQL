# Тестовое задание для компании Hezzl

## Задача
Развернуть сервис на Golang, Postgres, Clickhouse, Nats , Redis

- Развернуть БД PostgreSQL, Clickhouse и Redis
- Реализовать CRUD методы на GET-POST-PATCH-DELETE данных в таблице ITEMS в Postgres
- При редактировании данных в Postgres ставить блокировку на чтение записи и оборачивать все в транзакцию. Валидировать поля при редактировании. 
- При редактировании данных в ITEMS инвалидировать данные в REDIS
- Если записи нет (проверяем на PATCH-DELETE), выдаем ошибку (статус 404)
- При GET запросе данных из Postgres кешировать данные в Redis на минуту. Проверять данные сперва в Redis, если их нет, брать из БД и кэшировать
- При добавлении, редактировании или удалении записи в Postgres писать лог события в Clickhouse через очередь Nats.

## Описание проекта
### Схема
service<br />
├───cache - **кэширование через Redis**<br />
├───controllers - **обработчики для CRUD**<br />
├───db - **PostgreSQL**<br />
├───logs - **Clickhouse**<br />
├───migrations - **миграции**<br />
│   ├───clickhouse<br />
│   └───postgres<br />
├───models - **модели**<br />
└───nats  - **Подключение к серверу Nats**<br />

### Эндпоинты
- `/items/list` **GET** - список предметов<br />
- `/logs` **GET** - список логов событий<br />
- `/item/create?campaignId=X` **POST** - создание предмета<br />
{<br />
    "name":"XXX",<br />
    "description":"XXX"<br />
}<br />
- `/item/update?id=X&campaignId=X` **PATCH** - изменение предмета<br />
{<br />
    "name":"XXX",<br />
    "description":"XXX"<br />
}<br />
- `/item/remove?id=X&campaignId=X` **DELETE** - удаление<br />



## Результат
Все требования выполнены, сервис развернут в Docker. Поднимается с помощью 
`docker-compose up`
Доступен по `0.0.0.0:8080`. При миграции создается компания с id=1

- Сервис конфигурируется через переменные окружения в файле docker-compose.yml
- Коммуникация с PostgreSQL и Clickhouse ведется с помощью `database/sql` без ORM и маппинга для меньшей нагрузки и большей прозрачности происходящего, но я умею работать и с ORM.
- `Subscriber` и `Publisher` в **Nats** находятся в одном сервисе для наглядности и упрощения. Подписчик живет в горутине. Разделение на два сервиса можно провести без проблем.
- Миграции производятся с помощью https://github.com/golang-migrate/migrate в docker-compose
- Реализован дополнительный эндпоинт позволяющий получить записи о событиях из Clickhouse
