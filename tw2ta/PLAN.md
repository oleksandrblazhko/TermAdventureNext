# План: Конвертер TextWorld JSON → TermAdventure .ta

## Мета
Створити Go-утиліту `tw2ta` для конвертації JSON-файлів TextWorld у формат `.ta` TermAdventure. 
Підхід **"1 дія = 1 рівень"** для навчання студентів — максимальний контроль та деталізація кожного кроку.

## Підхід
- **Варіант 1:** Кожна команда з TextWorld (`open`, `take`, `go`, `put`) стає окремим рівнем у `.ta`
- **Цільова аудиторія:** Студенти/початківці
- **Переваги:** Максимальна деталізація, легкий контроль прогресу, не можна "перестрибнути" крок

## Структура файлів
```
TermAdventure/
├── tw2ta/
│   ├── main.go              # Точка входу утиліти конвертації
│   ├── parser.go            # Парсинг TextWorld JSON (структури +读取)
│   ├── graph.go             # Побудова графу станів (кімнати, зв'язки, предмети)
│   ├── converter.go         # Генерація .ta файлів з графу
│   └── templates.go         # Markdown шаблони для текстів рівнів
├── game_state.sh            # Helper-скрипт для управління станом гри
└── converted_challenges/    # Директорія для вихідних .ta файлів
```

## Кроки реалізації

### 1. Структури даних для TextWorld JSON
```go
type TextWorldJSON struct {
    Version int
    World   []Predicate
    Quests  []Quest
    Infos   []EntityInfo
    Grammar Grammar
}

type Predicate struct {
    Name      string
    Arguments []Argument
}

type Quest struct {
    Desc       string
    Reward     int
    Commands   []string
    WinEvents  []WinEvent
    FailEvents []FailEvent
}

type WinEvent struct {
    Commands []string
    Actions  []Action
}

type Action struct {
    Name              string
    Preconditions     []Predicate
    Postconditions    []Predicate
    CommandTemplate   string
    ReverseName       string
    ReverseCommandTpl string
}

type EntityInfo struct {
    ID       string
    Type     string
    Name     string
    Noun     string
    Adj      string
    Desc     string
    RoomType string
}
```

### 2. Парсинг JSON (tw2ta/parser.go)
- Зчитати JSON-файл
- Розібрати у структури Go
- Валідація обов'язкових полів
- Помилки при некоректному форматі

### 3. Побудова графу станів (tw2ta/graph.go)
- Витягти кімнати з `world[]` та `infos[]`
- Витягти предмети, контейнери, поверхні, двері
- Побудувати зв'язки між кімнатами (east_of, west_of, тощо)
- Визначити початковий стан гравця
- Витягти відповідності ключів до дверей (match)

### 4. Генерація .ta рівнів (tw2ta/converter.go)
Для кожної дії з квесту створити рівень:
```yaml
name: step_01_open_chest_drawer
test: game_state.sh check open c_0
precmd: |
  game_state.sh init
  echo "Крок 1/8: Відчиніть шухляду"
next: [step_02_take_old_key]
postcmd: game_state.sh update open c_0

## Крок 1: Відчиніть шухляду

Ви у **спальні**. Тут стоїть **шухляда (chest drawer)**.
Вона зачинена. Відчиніть її.

**Команда:** `open chest_drawer`
```

### 5. Шаблони рівнів (tw2ta/templates.go)
- Вступний рівень (intro)
- Рівень дії (action step)
- Рівень переходу між кімнатами (movement)
- Фінальний рівень (completion)
- Рівень-пастка (fail event)

### 6. Helper game_state.sh
```bash
#!/bin/bash
# Управління станом гри для конвертованих квестів

STATE_FILE="/tmp/game_state_${CHALLENGE_NAME}.json"

case "$1" in
  init)
    # Ініціалізація початкового стану з JSON
    ;;
  check)
    # Перевірка умови (напр., "check open c_0")
    ;;
  update)
    # Оновлення стану після дії
    ;;
  has)
    # Перевірка наявності предмета
    ;;
esac
```

### 7. main.go для tw2ta
```bash
# Використання
./tw2ta simple_game.json > converted_challenges/simple_game.ta
./tw2ta --output output.ta simple_game.json
./tw2ta --help
```

### 8. Приклад вихідного .ta файлу
```
name: intro
test: true
precmd: game_state.sh init simple_game

# Ласкаво просимо у TextWorld!

Ви прокидаєтесь у незнайомому будинку. 
Ваше завдання — знайти яблуко та покласти його на плиту.

Огляньтесь: `look`

--------------------

name: step_01_open_chest_drawer
test: game_state.sh check open c_0
next: [step_02_take_old_key]
precmd: echo "Крок 1/8: Відчиніть шухляду у спальні"

## Крок 1: Відчиніть шухляду

Ви у **спальні**. Тут стоїть **шухляда (chest drawer)**.
Вона зачинена. Відчиніть її.

**Команда:** `open chest_drawer`

--------------------

name: step_02_take_old_key
test: game_state.sh has k_0
next: [step_03_unlock_wooden_door]
precmd: echo "Крок 2/8: Візьміть старий ключ"

## Крок 2: Візьміть ключ

Шухляда відчинена! Всередині ви бачите **старий ключ (old key)**.
Візьміть його.

**Команда:** `take old_key`

... (продовження для всіх 8 кроків)

--------------------

name: final
test: game_state.sh check win
postcmd: echo "Вітаємо! Квест пройдено!"

## Квест пройдено! 🎉

Ви успішно поклали яблуко на плиту!
```

## Технічні деталі

| Параметр | Значення |
|----------|----------|
| Мова | Go (як весь TermAdventure) |
| Залежності | `encoding/json`, `os`, `strings`, `text/template`, `flag` |
| Вихід | `.ta` файл + `game_state.sh` |
| Команда | `./tw2ta <input.json> [output.ta]` |
| Платформа | Linux/macOS (як TermAdventure) |

## Інтеграція з TermAdventure

Конвертований `.ta` файл:
- ✅ Працює з існуючим бінарником `termadventure`
- ✅ Використовує `challenger.sh` та `ta_bashrc`
- ✅ Підтримує шифрування (`--enc`)
- ✅ Зберігає прогрес у `$HOME/.config/<challenge_name>/`
- ✅ Показує ідентифікатор рівня у промпті

## Наступні кроки після реалізації

1. Додати підтримку інших шаблонів (`tw-cooking`, `tw-treasure_hunter`)
2. Генерація українською мовою (поточна — англійська)
3. Веб-інтерфейс для перегляду графу квесту
4. Автоматичне шифрування вихідного файлу
5. Генерація README для кожного квесту

## Візуалізація процесу конвертації

```
┌─────────────────────┐
│  simple_game.json   │  (TextWorld JSON)
│  (2467 рядків)      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   tw2a/parser.go    │  Розбір JSON у структури Go
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   tw2a/graph.go     │  Побудова графу станів
│                     │  - 6 кімнат
│                     │  - 13 предметів
│                     │  - 8 кроків квесту
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  tw2a/converter.go  │  Генерація .ta рівнів
│                     │  Для кожної дії:
│                     │  - name
│                     │  - test
│                     │  - precmd
│                     │  - next
│                     │  - Markdown текст
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  simple_game.ta     │  Готовий файл для TermAdventure
│  (~150 рядків)      │  + game_state.sh
└─────────────────────┘
```

## Статус виконання

- [x] 1. Зберегти план у файл PLAN.md ✅
- [x] 2. Створити структури даних для TextWorld JSON ✅
- [x] 3. Реалізувати парсинг JSON ✅
- [x] 4. Побудова графу станів ✅
- [x] 5. Генерація .ta рівнів ✅
- [x] 6. Створити шаблони рівнів ✅
- [x] 7. Створити game_state.sh helper ✅
- [x] 8. Реалізувати main.go для tw2ta ✅
- [x] 9. Тестування на simple_game.json ✅
- [x] 10. Створити інструкцію TW2TA_GUIDE.md ✅
