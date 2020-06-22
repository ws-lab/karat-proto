# Karat Proto
Karat Proto это сервис для преобразования пакетов данных, полученных с приборов учета производства [НПО “Карат”](https://www.karat-npo.com/) по сети LoRaWAN.

  - Работа по REST API (порт по умолчанию 8085)
  - Работа по GRPC (порт по умолчанию 8081)  
  - SWAGGER для визуализации и тестирования (http://127.0.0.1:8085/swagger-ui)

Описание по работе с сервисом можно посмотреть [здесь](https://www.karat-npo.com/).

### Docker
Необходимо установить актуальную версию [Docker](https://docs.docker.com/get-docker/).

```sh
git clone https://github.com/ws-lab/karat-proto.git
cd karat-proto
sudo docker build -t karat-proto --network=host --no-cache .
```
Это создаст образ Karat Proto и вытянет необходимые зависимости.

После этого запустите образ Docker и сопоставьте порт с тем, что вы хотите на своем хосте. В этом примере мы просто отображаем порт 8085 хоста на порт 8085 Docker  и порт 8081 на порт 8081:

```sh
sudo docker run -d -p 8085:8085 -p 8081:8081 karat-proto:latest
```
Проверьте развертывание, перейдя по адресу вашего сервера в выбранном вами браузере.

```sh
http://127.0.0.1:8085/swagger-ui
```

#### Building for source
Для сборки необходим golang версии не ниже 1.11
```sh
git clone https://github.com/ws-lab/karat-proto.git
```
Клонируем репозиторий
```sh
cd karat-proto
GO111MODULE=on make build
```
Эта операция запустит сборку приложения.
```sh
./build/karat-proto
```
Запускаем приложение и проверяем:
```sh
http://127.0.0.1:8085/swagger-ui
```






License
----

MIT
