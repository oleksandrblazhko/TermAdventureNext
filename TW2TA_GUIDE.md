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
# └── go.mod (якщо є)

ls -la game_state.sh
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

# З копіюванням game_state.sh
./tw2ta --copy-game-state test_game.json

# Переглянути результат
head -n 50 test_game.ta
```

### Крок 6: Підготувати до запуску

```bash
# Зроби game_state.sh виконуваним
chmod +x game_state.sh

# Перемісти його до директорії з бінарником (якщо потрібно)
cp game_state.sh /usr/local/bin/
# АБО залиш у поточній директорії (шлях має бути в PATH)

# Ініціалізуй стан (опціонально, зазвичай робиться автоматично)
./game_state.sh init test_game
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

## Як працює game_state.sh

`game_state.sh` — це менеджер стану гри. Він зберігає інформацію про:

- Стан контейнерів (відчинені/зачинені)
- Стан дверей (замкнені/відімкнені/відчинені)
- Інвентар гравця
- Поточну кімнату гравця
- Умову перемоги

### Команди game_state.sh

```bash
# Ініціалізація
./game_state.sh init <challenge_name>

# Перевірка умов (використовується у test:)
./game_state.sh check open c_0        # Контейнер відчинений?
./game_state.sh check unlocked d_0    # Двері відімкнені?
./game_state.sh check win             # Квест пройдено?
./game_state.sh has k_0               # Ключ в інвентарі?
./game_state.sh at r_0                # Гравець у кімнаті r_0?

# Оновлення стану
./game_state.sh update open c_0       # Відчинити контейнер
./game_state.sh update unlocked d_0   # Відімкнути двері

# Перегляд стану (для дебагу)
./game_state.sh state

# Скинути стан
./game_state.sh reset
```

### Де зберігається стан

```bash
/tmp/termadventure_<challenge_name>/
├── game_state.json          # Основний файл стану
├── container_c_0            # Стан контейнера (open/closed)
├── container_contents_c_0   # Вміст контейнера
├── door_d_0                 # Стан дверей (locked/unlocked/open/closed)
├── inventory                # Предмети в інвентарі
├── player_room              # Поточна кімната гравця
└── won                      # Флаг перемоги (true/false)
```

## Вирішення проблем

### Помилка: "game_state.sh: command not found"

```bash
# Варіант 1: Додати до PATH
export PATH="$HOME/TermAdventureNext:$PATH"

# Варіант 2: Використовувати повний шлях
test: ./game_state.sh check open c_0

# Варіант 3: Скопіювати до /usr/local/bin
sudo cp game_state.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/game_state.sh
```

### Помилка: "стан гри не ініціалізовано"

```bash
# Ініціалізуй вручну
./game_state.sh init <challenge_name>

# Або скинь і почни знову
./game_state.sh reset
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
