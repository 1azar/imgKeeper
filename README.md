Сервис imgKeeper. 
protobuf контракты реализованы в отдельном репозитории https://github.com/1azar/imgKeeper-api-contracts сгенерированный код используется как зависимость в этом проекте (сервере) и в клиентском приложении (https://github.com/1azar/imgKeeper-client)

файлы хранятся в дириктории `storage/imgs`, при этом информация о них (имя файла, дата создания, дата обновления) хранится в sqlite файле `/storage/imgKeeper.db` 