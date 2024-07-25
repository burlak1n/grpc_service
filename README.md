# SSO
- Auth - авторизация и аутентификация (https://www.youtube.com/watch?v=EURjTg5fw-E&t=3017s)
- Permisions
- User Info

## Protobuf
Для генерации файлов необходима утилита `protoc`: https://grpc.io/docs/languages/go/quickstart/
> protoc -I proto proto/sso/sso.proto --go_out=./gen/go --go_opt=paths=source_relative --go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative

# TODO
- Поменять в gen-файлах AppId uint32 на uint8
- Улучшить логгер 1:02 - yt
  - ~~Обработка ошибок~~
  - ~~Цвета уровней~~
  - Шифрование паролей и секретной информации - Звёздочки
- Тесты для jwt
- Реализовать обработку ошибок в сервисном слое
- is_admin вынести в таблицу админов
- обработка ошибок в слое работы с данными (pg) 2:28
- улучшить migrate (избавиться от drop) +-

Нужно ли передавать в isAdmin appID? 2:02
Refresh token?
