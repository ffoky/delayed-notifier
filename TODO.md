# TODO

Архитектурные и технические задачи, выявленные при ревью.

## Архитектура

- [ ] **Перенести интерфейсы сервисов из domain в application**  
  `domain/notification_service.go` и `domain/user_service.go` — use-case интерфейсы, не доменные.
  Должны лежать в `application/` или определяться локально в пакете-потребителе.

- [ ] **Перенести `Publisher` из domain в application**  
  `domain/publisher.go` потребляется application-сервисом, а не domain-слоем.

- [ ] **Рассмотреть Transactional Outbox вместо прямой публикации в Create()**  
  Сейчас при недоступности RabbitMQ уведомление создаётся в БД, но не публикуется.
  Первая отправка задержится до следующего тика планировщика (до 30с).
  Outbox-паттерн устраняет этот gap и делает Create() проще.

## Отказоустойчивость

- [ ] **Circuit Breaker на Telegram и Email sender**  
  Если провайдер недоступен, каждое уведомление выжигает все ретраи.
  Библиотека: `github.com/sony/gobreaker`.

- [ ] **Timeout на HTTP-клиент в TelegramSender**  
  `&http.Client{}` без таймаута — бесконечное ожидание.
  Добавить `Timeout: 10 * time.Second`.

- [ ] **Пересмотреть Nack.Requeue: false в consumer**  
  При транзиентной ошибке БД (timeout, disconnect) сообщение теряется.
  Нужно различать доменные ошибки (ACK) и инфраструктурные (Nack + requeue или retry).

- [ ] **Rate limiting в ProcessReady**  
  Scheduler публикует до 100 сообщений подряд без задержки.
  При большом накоплении это создаёт пик нагрузки на consumer.

## Код

- [ ] **`IncrementRetries` возвращать новое значение**  
  Сейчас в памяти `n.Retries++` после DB-инкремента — два источника правды.
  Безопаснее: `newRetries, err := repo.IncrementRetries(ctx, id)`.

- [ ] **Интерфейс `UserCreator` в telegram/bot.go**  
  Переименовать в `userRegistrar`, метод — `GetOrCreate` вместо `CreateUser`,
  чтобы бот не создавал дубли при повторном `/start`.

- [ ] **Заменить `log.Fatalf` на wbf-логгер + `os.Exit(1)` в main**  
  Стандартный `log.Fatalf` не пишет в структурированный лог.

- [ ] **Добавить индекс на `notifications(user_id)`**  
  При фильтрации уведомлений по пользователю (будущая фича) без индекса — seq scan.

## Тесты

- [ ] **Unit-тесты для `application.Service`** с моками репозитория и publisher
- [ ] **Integration-тест для `NotificationRepo`** через testcontainers (PostgreSQL)
- [ ] **Тест сценария retry**: уведомление → ошибка sender → статус planned → ретрай
