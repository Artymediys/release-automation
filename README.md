# ARel – Auto Release
Программа для автоматического релиза из исходной ветки в целевую, с формированием новой версии и тега.
Также поддерживается релиз в нескольких репозиториях.

## Сборка и запуск
В директории `./bin` лежат сборки приложения под разные ОС.

### Зависимости
Для работы c собранным приложением, требуется установленный **git** на устройстве.

Для самостоятельной сборки и последующего запуска, требуется установленный **Go** версии `>=1.22.5`.

### Запуск собранного приложения
Приложение не имеет каких-либо аргументов и флагов.
Для запуска требуется обратиться к файлу, который находится в директории с требуемой ОС и архитектурой: 
```shell
./bin/<ОС>/arel_<архитектура>
```

### Самостоятельная сборка
В корне репозитория находится bash-скрипт `build.sh` для сборки приложения под необходимую ОС и архитектуру.

Скрипт можно запустить вместе с аргументами:
1) Сборка приложения под предустановленные ОС и архитектуры, а именно –> **windows amd64**; **linux amd64**;
**macos amd64** и **arm64**
```shell
./build.sh all
```

2) Сборка приложения под конкретную ОС. Из доступных вариантов -> **windows**; **linux**, **macos**
```shell
./build.sh <ОС>
```

Также скрипт можно запустить без аргументов. Тогда он соберёт приложение под текущую ОС и архитектуру, 
на которой был запущен скрипт:
```shell
./build.sh
```

Если требуется собрать приложение под ОС или архитектуру, которой нет в стандартном наборе скрипта `build.sh`,
тогда можно посмотреть список доступных для сборки ОС и архитектуры командой:
```shell
go tool dist list
```
после чего выполнить следующую команду для сборки:

**Linux/MacOS**
```shell
env GOOS=<ОС> GOARCH=<архитектура> go build -o <название итогового файла>
```

**Windows/Powershell**
```shell
$env:GOOS="<ОС>"; $env:GOARCH="<архитектура>"; go build -o <название итогового файла>
```

**Windows/cmd**
```shell
set GOOS=<ОС> && set GOARCH=<архитектура> && go build -o <название итогового файла>
```

## Конфигурация
При первом запуске приложение попросит **GitLab Personal Access Token** и **GitLab URL**.

- **GitLab Personal Access Token** – генерируются в настройках аккаунта в разделе **Access Tokens**
- **GitLab URL** – адрес GitLab, указывающийся в формате https://gitlab.example.com

Файл конфигурации автоматически создаётся в домашней директории пользователя с названием `.arel.yaml`