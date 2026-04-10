# Інструкція: Конвертер TextWorld JSON → TermAdventure .ta

## Швидкий старт на Linux-сервері

### Крок 1: Отримати код через git

```bash
# Увійди на Linux-сервер
cd ~/TermAdventureNext

# Отримай останні зміни
git pull origin main

# Або якщо репозиторій ще не клоновано:
git clone <your-repo-url> ~/TermAdventureNext
cd ~/TermAdventureNext
```

### Крок 2: Перевірити файли

Переконайся що з'явилися нові файли:

```bash
ls -la tw2ta/
# Має показати:
# ├── main.go
# ├── parser.go
# ├── graph.go
# ├── converter.go
# ├── mapping_parser.go
# └── go.mod

ls -la TW_BASH_MAPPING.md
```

### Крок 3: Зібрати утиліту tw2ta

```bash
cd ~/TermAdventureNext

# Збери утиліту
go build -o tw2ta ./tw2ta/

# Перевір що працює
./tw2ta --version
./tw2ta --help
```

### Крок 4: Згенерувати TextWorld JSON

Якщо ще не маєш JSON-файлу:

```bash
# Згенеруй просту гру
tw-make tw-simple --seed 42 --output test_game.z8 --json test_game.json --goal brief

# Або використовуй існуючий
ls -la prompts/simple_game.json
```

### Крок 5: Конвертувати JSON → .ta

```bash
# Базова конвертація
./tw2ta test_game.json

# Це створить test_game.ta у поточній директорії

# З явною назвою челенджу
./tw2ta --challenge "My First Quest" test_game.json my_quest.ta

# Переглянути результат
head -n 50 test_game.ta
```

**Важливо:** `tw2ta` автоматично читає `TW_BASH_MAPPING.md` для генерації bash-команд. Якщо змінити цей файл — наступна конвертація використає нові правила!

### Крок 6: Підготувати до запуску

```bash
# game_state.sh більше не потрібен!
# Конвертований .ta використовує прямі bash-команди

# Просто переконайся що /tmp/game існує
mkdir -p /tmp/game
```

### Крок 7: Збери TermAdventure (якщо ще не зібрано)

```bash
cd ~/TermAdventureNext

# Збери основний бінарник
go build -o termadventure

# Або з ключем шифрування
go build -ldflags "-X main.encryption_key=my_secret_key" -o termadventure
```

### Крок 8: Тестовий запуск

```bash
# Перегляд згенерованого квесту
./termadventure --print test_game.ta

# Запуск через challenger.sh
export CHALLENGE_FILE=./test_game.ta
./challenger.sh
```

## Повний приклад від генерації до гри

```bash
#!/bin/bash
set -e

# 1. Генерація TextWorld гри
tw-make tw-simple --seed 42 --output my_game.z8 --json my_game.json

# 2. Конвертація у TermAdventure формат
./tw2ta --copy-game-state my_game.json

# 3. Підготовка
chmod +x game_state.sh

# 4. Перегляд
./termadventure --print my_game.ta

# 5. Запуск
export CHALLENGE_FILE=./my_game.ta
./challenger.sh
```

## Структура згенерованого .ta файлу

Після конвертації `simple_game.json` отримаєш файл з такою структурою:

```
name: intro
test: true
precmd: echo 'Початок челенджу: simple_game'
next: [step_01]

# Ласкаво просимо до TextWorld!
...

--------------------

name: step_01
test: game_state.sh check open c_0
precmd: echo 'Крок 1/8: Відчиніть контейнер'
next: [step_02]

## Крок 1/8: Відчиніть контейнер
...

--------------------

name: step_02
test: game_state.sh has k_0
precmd: echo 'Крок 2/8: Візьміть предмет'
next: [step_03]

...

--------------------

name: final
test: true
precmd: echo 'Квест завершено!'

# 🎉 Вітаємо! Квест пройдено!
```

## Як працюють bash-команди у конвертованих квестах

Кожен рівень у конвертованому `.ta` файлі використовує **прямі bash-команди** замість `game_state.sh`:

### Структура файлів гри

```
/tmp/game/                          # Робоча директорія гри
├── chest_drawer/                   # Контейнер
│   ├── .closed                     # Прапорець стану
│   ├── .open                       # Прапорець стану
│   └── old_key                     # Предмет всередині
├── stove/                          # Поверхня
│   └── apple                       # Предмет на поверхні
├── door_wooden_door.state          # Стан дверей: locked/closed/open
├── door_wooden_door.unlocked       # Прапорець що відімкнено
├── doors.log                       # Лог дій з дверима
├── movement.log                    # Лог переміщень
├── inventory.log                   # Лог інвентарю
└── win_condition                   # Файл-прапорець перемоги

~/.current_room                     # Поточна кімната гравця
~/old_key                           # Предмет в інвентарі гравця
```

### Приклади команд для гравця

| Дія TextWorld | Що робить гравець | Перевірка (test) |
|---------------|-------------------|------------------|
| `open chest_drawer` | `rm /tmp/game/chest_drawer/.closed` | `test ! -f ...` |
| `take old_key` | `cp /tmp/game/chest_drawer/old_key ~/` | `test -f ~/old_key` |
| `unlock wooden_door` | `echo "closed" > /tmp/game/door_wooden_door.state` | `test "$(cat ...)" = "closed"` |
| `open wooden_door` | `echo "open" > /tmp/game/door_wooden_door.state` | `test "$(cat ...)" = "open"` |
| `go east` | `echo "kitchen" > ~/.current_room` | `test "$(cat ~/.current_room)" = "kitchen"` |
| `put apple on stove` | `cp ~/apple /tmp/game/stove/` | `test -f /tmp/game/stove/apple` |

### Повний приклад рівня

```yaml
name: step_01_open_chest_drawer
test: test ! -f /tmp/game/chest_drawer/.closed
precmd: |
  mkdir -p /tmp/game/chest_drawer
  touch /tmp/game/chest_drawer/.closed
  echo "Крок 1/8: Відчиніть шухляду"
next: [step_02_take_old_key]
postcmd: touch /tmp/game/chest_drawer/.open

## Крок 1/8: Відчиніть контейнер

Ви у кімнаті **спальня**. Перед вами **шухляда (chest_drawer)** — вона зачинена.

**Виконайте команду:**

```bash
rm /tmp/game/chest_drawer/.closed
```

*Оригінальна команда TextWorld:* `open {c_0}`*
```

### Правила мапінгу (TW_BASH_MAPPING.md)

Файл `TW_BASH_MAPPING.md` визначає всі правила конвертації. Якщо його змінити:

```bash
# 1. Відредагуй правила
vim TW_BASH_MAPPING.md

# 2. Перезапусти конвертацію
./tw2ta test_game.json

# 3. Результат автоматично оновиться!
cat test_game.ta
```

**Не треба чіпати код `converter.go`** — просто зміни markdown!

## Вирішення проблем

### Помилка: "стан гри не ініціалізовано"

```bash
# Створи директорію гри
mkdir -p /tmp/game

# Скинь стан
rm -rf /tmp/game/*
```

### Помилка компіляції Go

```bash
# Перевір версію Go
go version
# Має бути >= 1.18

# Очисти кеш
go clean -cache -modcache

# Збери заново
go build -o tw2ta ./tw2ta/
```

### JSON не парситься

```bash
# Перевір валідність JSON
python3 -m json.tool simple_game.json > /dev/null

# Переглянь перші рядки
head -n 20 simple_game.json
```

## Генерація складніших ігор

```bash
# Більший будинок (різний seed = різна гра)
tw-make tw-simple --seed 100 --output big_house.z8 --json big_house.json

# Кулінарний квест (якщо доступний)
tw-make tw-cooking --level 3 --output cooking.z8 --json cooking.json

# Пошук скарбів
tw-make tw-treasure_hunter --level 5 --output treasure.z8 --json treasure.json

# Конвертуй будь-який
./tw2ta big_house.json
./tw2ta cooking.json
./tw2ta treasure.json
```

## Порівняння: TextWorld vs TermAdventure

| TextWorld | TermAdventure |
|-----------|---------------|
| `open chest drawer` | Рівень з `test: game_state.sh check open c_0` |
| `take old key` | Рівень з `test: game_state.sh has k_0` |
| `unlock wooden door` | Рівень з `test: game_state.sh check unlocked d_0` |
| `go east` | Рівень з `next: [step_02]` |
| `put apple on stove` | Рівень з `test: game_state.sh check on f_1 s_2` |
| Інтерпретатор Z-machine | Bash + termadventure бінарник |
| Стан у пам'яті | Стан у файлах `/tmp/` |

## Наступні кроки

1. ✅ **Базова конвертація** — працює
2. 🔲 **Підтримка інших шаблонів** (tw-cooking, tw-treasure_hunter)
3. 🔲 **Генерація українською мовою**
4. 🔲 **Веб-перегляд графу квесту**
5. 🔲 **Автоматичне шифрування**

## Довідка по tw2ta

```bash
./tw2ta --help

# Опції:
#   --output        Вихідний файл .ta
#   --challenge     Назва челенджу
#   --copy-game-state  Копіювати game_state.sh
#   --version       Показати версію
#   --help          Показати допомогу
```
