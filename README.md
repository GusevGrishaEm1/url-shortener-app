«Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

## Использование

Ваш сервис сокращения URL должен иметь следующий функционал:

- Сокращение URL с помощью POST-запроса на `/shorten` с JSON-телом, содержащим поле `url` с URL для сокращения.
- Расширение сокращённого URL с помощью GET-запроса на `/<token>`, где `<token>` — токен сокращённого URL.

Примеры использования можно найти в файле `internal/app/server/example_test.go`.

## Лицензия

Этот проект лицензируется под лицензией [MIT](LICENSE).
