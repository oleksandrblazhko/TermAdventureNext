# Інструкція: Конвертер TextWorld JSON → TermAdventure .ta

## Швидкий старт на Linux-сервері

### Крок 1: Отримати код через git

```bash
# Увійди на Linux-сервер
cd ~/TermAdventureNext

# Отримати останні зміни
git pull origin main
```
### Крок 2: Зібрати утиліту tw2ta

```bash
# Зібрати утиліту
go build -o tw2ta ./tw2ta/

# Перевірити роботу
./tw2ta/tw2ta --version
./tw2ta/tw2ta --help
```

### Крок 3: Згенерувати TextWorld JSON

```bash
# Згенерувати квест-гру типу tw-simple
tw-make tw-simple --seed 1 --output test_game.z8 --goal brief --rewards dense 

```

### Крок 4: Конвертувати JSON → .ta

```bash
# З явною назвою мапінгу (для інших типів ігор)
./tw2ta/tw2ta --mapping ./prompts/Lab1_Bash_Scripts.yaml ./prompts/test_game_1.json --output test_game_1.ta 

```

**Важливо:** За замовчуванням `tw2ta` використовує `tw-simple_mapping.yaml`. Прапор `--mapping` дозволяє вказати інший файл мапінгу для інших типів ігор (tw-cooking, tw-treasure_hunter тощо).

### Крок 5: Підготувати до запуску

```bash
# Конвертований .ta використовує прямі bash-команди

# Просто переконайся що $HOME/.tw2ta_game існує
mkdir -p $HOME/.tw2ta_game
```

### Крок 6: Збірка TermAdventure (якщо ще не зібрано)

```bash
cd ~/TermAdventureNext

# Збери основний бінарник
go build -o termadventure

# Або з ключем шифрування
go build -ldflags "-X main.encryption_key=my_secret_key" -o termadventure
```

### Крок 7: Тестовий запуск

```bash
# Перегляд згенерованого квесту
./termadventure --print test_game.ta

# Запуск через challenger.sh
export CHALLENGE_FILE=./test_game_1.ta
./challenger.sh
```

## Повний приклад від генерації до гри

```bash
#!/bin/bash
set -e

# 1. Генерація TextWorld гри
tw-make tw-simple --seed 1 --output my_game.z8 --json my_game.json

# 2. Конвертація у TermAdventure формат (з явним вказанням мапінгу)
./tw2ta --mapping tw-simple_mapping.yaml my_game.json

# 3. Перегляд
./termadventure --print my_game.ta

# 4. Запуск
export CHALLENGE_FILE=./my_game.ta
./challenger.sh
```

## Структура згенерованого .ta файлу

Після конвертації `simple_game.json` отримаєш файл з такою структурою:

```yaml
name: intro
test: "true"
precmd: echo 'Початок челенджу: simple_game'
next: [step_01]

# Ласкаво просимо до TextWorld!
...

--------------------

name: step_01
test: test ! -f $HOME/.tw2ta_game/bedroom/chest_drawer/.closed
precmd: mkdir -p $HOME/.tw2ta_game/bedroom/chest_drawer && touch $HOME/.tw2ta_game/bedroom/chest_drawer/.closed
postcmd: touch $HOME/.tw2ta_game/bedroom/chest_drawer/.open
next: [step_02]

## Крок 1/8: Відчиніть контейнер
...

--------------------

name: step_02
test: test -f ~/old_key
precmd: mkdir -p $HOME/.tw2ta_game/bedroom/chest_drawer
next: [step_03]

...

--------------------

name: final
test: "true"
precmd: echo 'Квест завершено!'

# 🎉 Вітаємо! Квест пройдено!
```

## Як працюють bash-команди у конвертованих квестах

Кожен рівень у конвертованому `.ta` файлі використовує **прямі bash-команди** замість вбудованих команд для TextWorld

### Структура файлів гри

```
$HOME/.tw2ta_game/                    # Робоча директорія гри (унікальна для кожного користувача)
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
├── current_room                    # Поточна кімната гравця
└── win_condition                   # Файл-прапорець перемоги

~/.current_room                     # Поточна кімната гравця (legacy)
~/old_key                           # Предмет в інвентарі гравця
```

### Приклади команд для гравця

| Дія TextWorld | Що робить гравець | Перевірка (test) |
|---------------|-------------------|------------------|
| `open chest_drawer` | `mv $HOME/.tw2ta_game/bedroom/chest_drawer/.closed` | `test ! -f .../.closed` |
| `take old_key` | `mv $HOME/.tw2ta_game/bedroom/chest_drawer/old_key ~/` | `test -f ~/old_key` |
| `unlock wooden_door` | `echo "closed" > $HOME/.tw2ta_game/door_wooden_door.state` | `test "$(cat ...)" = "closed"` |
| `open wooden_door` | `echo "open" > $HOME/.tw2ta_game/door_wooden_door.state` | `test "$(cat ...)" = "open"` |
| `go east` | `echo "kitchen" > $HOME/.tw2ta_game/current_room` | `test "$(cat $HOME/.tw2ta_game/current_room)" = "kitchen"` |
| `put apple on stove` | `mv ~/apple $HOME/.tw2ta_game/kitchen/stove/` | `test -f $HOME/.tw2ta_game/kitchen/stove/apple` |

### Повний приклад рівня

```yaml
name: step_01
test: test ! -f $HOME/.tw2ta_game/bedroom/chest_drawer/.closed
precmd: mkdir -p $HOME/.tw2ta_game/bedroom/chest_drawer && touch $HOME/.tw2ta_game/bedroom/chest_drawer/.closed
postcmd: touch $HOME/.tw2ta_game/bedroom/chest_drawer/.open
next: [step_02]

## Крок 1/8: Відчиніть контейнер

Ви у кімнаті **bedroom**. Перед вами **chest drawer** — він зачинений.

**Виконайте команду:**

```bash
rm $HOME/.tw2ta_game/bedroom/chest_drawer/.closed
```

*Оригінальна команда TextWorld:* `open c_0`*
```

### Правила мапінгу (tw-simple_mapping.yaml)

Файл `tw2ta/tw-simple_mapping.yaml` визначає всі правила конвертації. Якщо його змінити:

```bash
# 1. Відредагуй шаблони дій
vim tw2ta/tw-simple_mapping.yaml

# 2. Перезапусти конвертацію
./tw2ta test_game.json

# 3. Результат автоматично оновиться!
cat test_game.ta
```

**Не треба чіпати код `converter.go`** — просто зміни YAML!

## Вирішення проблем

### Помилка: "стан гри не ініціалізовано"

```bash
# Створи директорію гри
mkdir -p $HOME/.tw2ta_game

# Скинь стан
rm -rf $HOME/.tw2ta_game/*
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

# Кулінарний квест (якщо доступний) — потрібен свій мапінг!
tw-make tw-cooking --level 3 --output cooking.z8 --json cooking.json

# Пошук скарбів — потрібен свій мапінг!
tw-make tw-treasure_hunter --level 5 --output treasure.z8 --json treasure.json

# Конвертація з відповідним мапінгом
./tw2ta --mapping tw2ta/tw-simple_mapping.yaml big_house.json
./tw2ta --mapping tw2ta/tw-cooking_mapping.yaml cooking.json       # коли буде створено
./tw2ta --mapping tw2ta/tw-treasure_hunter_mapping.yaml treasure.json  # коли буде створено
```

## Порівняння: TextWorld vs TermAdventure

| TextWorld | TermAdventure |
|-----------|---------------|
| `open chest drawer` | Рівень з `test: test ! -f $HOME/.tw2ta_game/bedroom/chest_drawer/.closed` |
| `take old key` | Рівень з `test: test -f ~/old_key` |
| `unlock wooden door` | Рівень з `test: test "$(cat .../door_wooden_door.state)" = "closed"` |
| `go east` | Рівень з `test: test "$(cat .../current_room)" = "kitchen"` |
| `put apple on stove` | Рівень з `test: test -f .../kitchen/stove/apple` |
| Інтерпретатор Z-machine | Bash + termadventure бінарник |
| Стан у пам'яті | Стан у файлах `$HOME/.tw2ta_game/` |

## Наступні кроки

1. ✅ **Базова конвертація** — працює з YAML мапінгом
2. ✅ **Підтримка різних типів ігор** — через `--mapping` прапор
3. 🔲 **Мапінг для tw-cooking**
4. 🔲 **Мапінг для tw-treasure_hunter**
5. 🔲 **Генерація українською мовою**
6. 🔲 **Веб-перегляд графу квесту**
7. 🔲 **Автоматичне шифрування**

## Довідка по tw2ta

```bash
./tw2ta --help

# Опції:
#   --output         Вихідний файл .ta (за замовчуванням: <input>.ta)
#   --mapping        YAML-файл мапінгу (за замовчуванням: tw-simple_mapping.yaml)
#   --challenge      Назва челенджу (за замовчуванням: з імені файлу)
#   --copy-game-state  Копіювати game_state.sh
#   --version        Показати версію
#   --help           Показати допомогу
```

## Підтримувані типи ігор

| Тип гри | Файл мапінгу | Статус |
|---------|-------------|--------|
| `tw-simple` | `tw-simple_mapping.yaml` | ✅ Створено |
| `tw-cooking` | `tw-cooking_mapping.yaml` | 🔲 Планується |
| `tw-treasure_hunter` | `tw-treasure_hunter_mapping.yaml` | 🔲 Планується |
| `tw-commonsense` | `tw-commonsense_mapping.yaml` | 🔲 Планується |
