# MINIPIC
[![build](https://github.com/bardex/minipic/actions/workflows/build.yml/badge.svg?branch=dev)](https://github.com/bardex/minipic/actions/workflows/builder.yml)
[![Go Report Card](https://goreportcard.com/badge/bardex/minipic)](https://goreportcard.com/report/bardex/minipic)

HTTP сервис для генерации миниатюр изображений (учебный проект).

## Возможности сервиса

- Загрузка изображения с удаленного хоста с проксированием http-заголовков от клиента к хосту и обратно
- Ресайз скачанного изображения
- Кеширование обработанных изображений вместе с http заголовками с использованием стратегии Least Recently Used
- Поддерживаемые форматы изображений: 
  - JPEG
  - PNG
- Поддерживаемые режимы ресайза: 
  - `fit` - вписать изображение целиком в заданные размеры (ресайз по большей стороне)
  - `fill` - заполнить заданные размеры изображением (ресайз по меньшей стороне + центрирование и подрезка лишнего) 

## Выбор библиотеки для работы с изображениями
В отборе участвовали три библиотеки
- [github.com/h2non/bimg](https://github.com/h2non/bimg) - адаптер к C библиотеке [libvips](https://github.com/jcupitt/libvips) требует её наличия в запускаемой среде,
  имеет проблемы с компиляцией в режиме CGO_ENABLED=1
- [github.com/anthonynsimon/bild](https://github.com/anthonynsimon/bild) - pure Go
- [github.com/disintegration/imaging](https://github.com/disintegration/imaging) - pure Go

Тесты производительности проводились для ресайза jpeg изображения 1920x1080px (707kB) в изображение 800x600px.
```
$ go test -bench=. -benchmem  bench_test.go 
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-6700T CPU @ 2.80GHz
BenchmarkBimg-8            88          12814039 ns/op          114736 B/op         12 allocs/op
BenchmarkBild-8             4         301796907 ns/op        15247708 B/op         40 allocs/op
BenchmarkImaging-8       1506            781762 ns/op         1937763 B/op         14 allocs/op
PASS
ok      command-line-arguments  7.917s
```
[Код бенчмарков](https://gist.github.com/bardex/b3668ea77c9ad5d47dd72e01101438d6)

По результатам бенчмарков была выбрана библиотека [github.com/disintegration/imaging](https://github.com/disintegration/imaging).

## API
Сервис имеет единственный http endpoint:

```
GET http://SERVICE_ADDR/MODE/WIDTH/HEIGHT/SRC
```

- SERVICE_ADDR - хост и порт сервиса (указывается в конфигурационном файле)
- MODE - режим ресайза изображения (fit, fill)
- WIDTH - целевая ширина изображения в px
- HEIGHT - целевая высота изображения в px
- SRC - полный URL исходного изображения

Например: [http://127.0.0.1:9011/fit/600/400/https://cdn.pixabay.com/photo/2020/11/01/10/35/street-5703332_960_720.jpg](http://127.0.0.1:9011/fit/600/400/https://cdn.pixabay.com/photo/2020/11/01/10/35/street-5703332_960_720.jpg)


## Makefile
Для автоматизации рутинных операций в проекте используется команда `make`:

- `make build` - скомпилировать исходный код в исполняемый файл в локальной операционной системе.
- `make run` - скомпилировать исходный код и запустить исполняемый файл в локальной операционной системе. Для конфигурирования сервиса будет использован файл `confgis/config.toml`
- `make build-image` - собрать сервис в докер-образ с именем `minipic`
- `make run-image` - собрать и запустить докер-образ сервиса, контейнер будет автоматически удален после остановки
- `make install-lint` - установить линтер golangci-lint
- `make lint` - запустить линтер по всему проекту

## Конфигурирование
Для управления настройками сервиса используется конфигурационный *.toml файл. 
По-умолчанию используется файл `configs/config.toml`:
```
[server]
listen = ":9011"

[cache]
limit=10
directory="/tmp"
```

## TODO
1. Тесты на всё
2. Shutdown сервера
3. прикрутить GitHub Actions для CI/CD

