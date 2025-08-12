# yandex-music-go

Неофициальный Go SDK для неофициального API сервиса Yandex Music.

> DISCLAIMER: Библиотека не аффилирована с Яндексом. Используя её, вы соглашаетесь с рисками блокировки аккаунта. Убедитесь, что ваши действия соответствуют условиям сервиса.
ээ....
> DISCLAIMER2: Данный продукт является следствием эксперемента над ChatGpt 5. Основной задачей стояло создание полной функциональной копией репозитория на .Net от K1llMan с использованием бестпрактисов и парадигм свойственных именно Go. Функциональное тестирование сдк не проводилось, его работоспособность неизвестна, это самое интересное. Я выложил потенциально лютый калл и горд собой. Отпишите в иссуесах че там да как вообще.


## Содержание
1. Установка
2. Быстрый старт
3. Примеры
4. Архитектура
5. Auth Flow (получение токена)
6. Сервисы
7. Ошибки
8. Линтеры и качество
9. Contributing
10. Публикация версий (Release Process)
11. Roadmap
12. Лицензия

## 1. Установка
```bash
go get github.com/Banjirome/yandex-music-go@latest
```

## 2. Быстрый старт
```go
ctx := context.Background()
cli := client.New(
  client.WithToken(os.Getenv("YANDEX_MUSIC_TOKEN")), // OAuth токен
)
resp, err := cli.Search.Tracks(ctx, "Muse Uprising", 0, 10)
if err != nil { log.Fatal(err) }
for _, tr := range resp.Result.Tracks.Results {
  artist := ""
  if len(tr.Artists) > 0 { artist = tr.Artists[0].Name }
  fmt.Printf("%s - %s\n", tr.Title, artist)
}
```

## 3. Примеры
Поиск по альбому:
```go
album, err := cli.Album.Get(ctx, "123456")
```

Получение артиста:
```go
artist, err := cli.Artist.Get(ctx, "312")
```

Плейлист пользователя:
```go
pl, err := cli.Playlist.Get(ctx, "user-id", "playlist-id")
```

Загрузка файла (прямая ссылка трека/обложки):
```go
err = cli.Downloader.ToFile(ctx, url, "cover.jpg")
```

## 4. Архитектура
- `client` — корневой клиент, HTTP, сервисы, downloader.
- `auth` — токен, device id, proxy.
- `models` — общие модели ответа / ошибок.
- `search` — реализация Search API (остальные сервисы пока упрощены в пакете `client`).

Шаблон вызова:
```go
resp, err := cli.Service.Method(ctx, params...)
```
Все методы принимают `context.Context` для отмены.

## 5. Auth Flow (получение токена)
API требует OAuth токен (используется как `Authorization: OAuth <token>`).
Поскольку это неофициальный API, общедоступной стабильной процедуры может не быть. Возможные пути:
1. Войти в web версию Yandex Music, открыть DevTools (F12) → вкладка Network → найти любой запрос к `api.music.yandex.net` с заголовком Authorization и скопировать токен.
2. Либо извлечь токен из мобильного приложения (перехват HTTPS через mitmproxy — на свой риск).
3. Некоторые токены могут быть краткоживущими; реализуйте обновление самостоятельно (SDK пока не включает refresh-flow).
Без токена доступны только ограниченные публичные данные.

Пример безопасной загрузки токена из окружения:
```go
cli := client.New(client.WithToken(os.Getenv("YANDEX_MUSIC_TOKEN")))
```

### Multi-step User Auth (QR / Letter / App Password) – пример
Если у вас нет готового OAuth токена, можно пройти многошаговый поток (требует корректных `client_id/secret` и cookie jar – по умолчанию создаётся):
```go
cli := client.New(
  client.WithClientCredentials(os.Getenv("YANDEX_CLIENT_ID"), os.Getenv("YANDEX_CLIENT_SECRET")),
  client.WithXClientCredentials(os.Getenv("YANDEX_X_CLIENT_ID"), os.Getenv("YANDEX_X_CLIENT_SECRET")),
)
ctx := context.Background()

// 1. Инициация сессии по логину
types, err := cli.User.CreateAuthSession(ctx, "your_login")
if err != nil { log.Fatal(err) }
log.Println("methods:", types.Result.AuthTypes)

// 2. Запрашиваем QR ссылку (например)
qrURL, err := cli.User.GetAuthQRLink(ctx)
if err != nil { log.Fatal(err) }
fmt.Println("Откройте на телефоне:", qrURL)

// 3. Пулинг статуса подтверждения
st, err := cli.User.AuthorizeByQR(ctx, 2*time.Second, 2*time.Minute)
if err != nil { log.Fatal(err) }
log.Println("QR confirmed:", st.MagicLinkConfirmed)

// 4. На этом шаге `loginByCookies` внутри потока установит AccessToken + OAuth Token
acc, err := cli.User.GetUserAuth(ctx)
if err != nil { log.Fatal(err) }
fmt.Println("Authorized UID:", acc.Result.Account.Uid)
```

Альтернативно можно использовать `AuthorizeByAppPassword(password)` или письмо (`GetAuthLetter` -> `AuthorizeByLetter`). После успешной авторизации `cli.User.Authorize(...)` не нужен – токен уже будет в хранилище.

## 6. Сервисы
| Сервис | Поле клиента | Пример | Статус |
|--------|--------------|--------|--------|
| Search | `cli.Search` | `cli.Search.Tracks(...)` | Полностью примерный слой (модели частично) |
| Album  | `cli.Album`  | `cli.Album.Get(ctx, id)` | Основные (with-tracks, batch) |
| Artist | `cli.Artist` | `cli.Artist.Get(ctx, id)` | Основные (brief-info, batch, tracks/all-tracks) |
| Track  | `cli.Track`  | `cli.Track.Get(ctx, id)` | Основные (get, metadata/link, supplement, similar, play-audio) |
| Playlist | `cli.Playlist` | `cli.Playlist.Get(ctx, userID, playlistID)` | Основные операции (get, batch, create, rename, delete, change, favorites) |
| User   | `cli.User`   | `cli.User.Authorize(ctx, token)` | Расширено (token auth + многошаговые методы QR/Captcha/Letter/AppPassword, access token) |
| Queue  | `cli.Queue`  | `cli.Queue.List(ctx, device)` | Основные (list, get, create, update-position) |
| Radio  | `cli.Radio`  | `cli.Radio.Dashboard(ctx)` | Основные (dashboard, list, station, tracks, settings2, feedback) |
| Library| `cli.Library`| `cli.Library.LikedTracks(ctx)` | Лайки/дизлайки + RecentlyListened |
| Landing| `cli.Landing`| `cli.Landing.Get(ctx, landing.BlockPersonalPlaylists)` | Основные (landing3, feed, children) |
| Feed   | `cli.Feed`   | `cli.Feed.Get(ctx)` | Лента (минимальные модели) |
| Label  | `cli.Label`  | `cli.Label.Get(ctx, id)` | Минимально |
| Ugc    | `cli.Ugc`    | `cli.Ugc.GetUploadLink(ctx, pl, name)` | Upload (get upload link, upload bytes) |
| Ynison | `cli.Ynison` | `p,_:=cli.Ynison.Connect(ctx); st := p.State()` | Realtime state (websocket, keep-alive), без Play/Pause (паритет) |
| Account | `cli.Account` | `cli.Account.Status(ctx)` | Статус uid |

Минимально = возвращает обобщённую структуру (map/json.RawMessage). Полная типизация может быть расширена в будущих версиях.

Примечания:
- Ynison: методы управления воспроизведением (Play / Pause / Next / Previous) удалены для строгого паритета (в оригинале закомментированы). Предоставляется только websocket state + Current().
- User: поддержаны `Authorize`, `CreateAuthSession`, `GetAuthQRLink`, `AuthorizeByQR`, `GetCaptcha`, `AuthorizeByCaptcha`, `GetAuthLetter`, `AuthorizeByLetter`, `AuthorizeByAppPassword`, `GetAccessToken`, `GetLoginInfo` (требуют корректных client credentials и cookie jar). Используйте опции `WithClientCredentials`, `WithXClientCredentials`, `WithMobileProxyBaseURL`.
- Playlist: `InsertTracks` выполняет изменения и затем перечитывает плейлист (как в C#). `DeleteTracks` теперь возвращает результат diff без дополнительного чтения (паритет). Удалены неоригинальные методы `RemoveTracks` и заглушка `Personal()`.
- Track: реализованы перегрузки для работы как по ключу, так и по объекту (FileLinkTrack, DownloadMetadataTrack, SupplementTrack, SimilarTrack, Extract*Track). `SendPlayInfo` требует успешной `Authorize`, иначе вернёт ошибку.
- Auth: хранилище теперь содержит `IsAuthorized`, `User.Uid`, `User.Login` после успешной авторизации.
- Отсутствующие части перечислены ниже в разделе Parity Status / Missing.

### Дополнительные примеры
Получить персональные плейлисты ("Of The Day", "Premiere" и т.п.):
```go
pls, err := cli.Playlist.GetPersonalPlaylists(ctx)
```

Вставка трека в плейлист с обработкой конфликтов ревизии:
```go
plResp, _ := cli.Playlist.Get(ctx, uid, kind)
pl := plResp.Result
add := []playlist.TrackKey{{Id: "<trackId>"}}
res, err := cli.Playlist.InsertTracks(ctx, &pl, add)
if err != nil && strings.Contains(err.Error(), "revision conflict") {
  // refetch & retry
  plResp, _ = cli.Playlist.Get(ctx, uid, kind)
  pl = plResp.Result
  res, err = cli.Playlist.InsertTracks(ctx, &pl, add)
}
```

Удаление треков из плейлиста:
```go
del := []playlist.Track{{ID: "t1"}, {ID: "t2"}}
res, err := cli.Playlist.DeleteTracks(ctx, &pl, del)
```

Получение прямой ссылки и скачивание трека:
```go
link, err := cli.Track.FileLink(ctx, "<trackId>:<albumId>")
data, err := cli.Track.ExtractData(ctx, "<trackId>:<albumId>")
err = cli.Track.ExtractToFile(ctx, "<trackId>:<albumId>", "song.mp3")
```

Отправка play-audio статистики:
```go
tr := client.Track{ID: "123", DurationMs: 180000}
_ = cli.Track.SendPlayInfo(ctx, tr, "feed", "", "", false, 15.2, 15.2)
```

## 6.1 Опции конфигурации клиента
```go
client.New(
  client.WithBaseURL("https://api.music.yandex.net/"),
  client.WithToken("<oauth>"),
  client.WithClientCredentials(id, secret),          // oauth client credentials
  client.WithXClientCredentials(xid, xsecret),       // mobile proxy credentials
  client.WithMobileProxyBaseURL("https://mobileproxy.passport.yandex.net/"),
  client.WithUserAgent("my-app/1.0"),
  client.WithAuthStorage(customStorage),
)
```

## 7. Ошибки
- Сетевые: прямой `error` из `http.Client`.
- API >400: `*models.APIError` (попытка декодировать тело). Поля: `StatusCode`, `InvocationInfo`, `Error{Name, Message}`.
- Контекстная отмена: `context.Canceled`, дедлайн: `context.DeadlineExceeded`.
- Конфликт ревизии плейлиста: ошибка с сообщением `playlist revision conflict` (обновите плейлист и повторите diff).

## 8. Линтеры и качество
Используется `golangci-lint` (конфиг `.golangci.yml`). Запуск:
```bash
golangci-lint run ./...
go test ./...
```

## 9. Contributing
1. Форк / ветка из `main` (`feat/<feature>` или `fix/<issue>`)
2. Реализовать изменения + тесты + обновить README/CHANGELOG
3. Прогнать линтеры и тесты
4. Открыть PR (описать что изменено и почему)
5. После merge — мейнтейнер тегирует релиз

Code style: идиоматичный Go, без преждевременной абстракции. Пакеты маленькие и сфокусированные. Добавляйте generics только если это уменьшает дублирование.

## 10. Публикация версий
1. Обновить README / CHANGELOG
2. SemVer: несовместимое изменение → мажор (v2 → поправить module path)
3. Тег: `git tag v0.x.y && git push origin v0.x.y`
4. Проверить `pkg.go.dev`

## 11. Roadmap
- Расширить модели для всех сервисов
- Retry + backoff + rate limit handling
- Кэширование и ETag
- Полноценный auth refresh / device registration
- Генератор кода моделей из JSON фикстур
- Streaming API для плейлистов и радио

### Parity Status / Missing
Нереализованные части оригинальной C# библиотеки:
- Управляющие действия Ynison (Play/Pause/Next/Previous)
- Дополнительные перегрузки некоторых Playlist методов по объекту плейлиста (не критичны)
- Полная типизация всех моделей (используются частичные структуры / подмножества)
- Механизмы повторов, rate limiting, кэширование ответов

Дополнительно желательные улучшения (не влияющие на паритет): централизованный retry/backoff, error kinds, кэш сущностей (альбом, артист), расширенный TrackSupplement.

## 11.1 Тесты и покрытие
В репозитории присутствуют unit-тесты для ключевых путей (album batch, playlist change/revision, track link hashing, send play info, file link выбор). 
Запуск:
```bash
go test ./... -cover
```
Покрытие будет расти по мере расширения моделей. PR с тестами приветствуются.

Если цель — абсолютный паритет, эти пункты остаются к реализации; PR приветствуются.

## 12. Лицензия
MIT


