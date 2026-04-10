# Мапінг команд TextWorld → Bash-команди TermAdventure

## Огляд

Цей документ описує як кожна дія TextWorld конвертується у bash-команди для TermAdventure.

## Принципи мапінгу

Кожен крок у TextWorld стає **рівнем** у TermAdventure з:
- **`test`** — перевірка що дія виконана (bash-команда)
- **`precmd`** — підготовка середовища (створення файлів/директорій)
- **`postcmd`** — фіксація результату (оновлення стану)
- **текст** — інструкція для гравця з bash-командою

---

## 1. Робота з контейнерами (c)

### `open/c` — Відчинити контейнер

**TextWorld:** `open chest drawer`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p /tmp/game/chest_drawer && touch /tmp/game/chest_drawer/.closed` |
| `test` | `test ! -f /tmp/game/chest_drawer/.closed` |
| **Команда гравця** | `rm /tmp/game/chest_drawer/.closed` |
| `postcmd` | `touch /tmp/game/chest_drawer/.open` |

**Логіка:**
- Контейнер = директорія `/tmp/game/<container_name>/`
- `.closed` — файл-прапорець що контейнер зачинений
- Гравець видаляє `.closed` → контейнер відчинено

---

### `close/c` — Зачинити контейнер

**TextWorld:** `close chest drawer`

| Поле | Значення |
|------|----------|
| `test` | `test -f /tmp/game/chest_drawer/.closed` |
| **Команда гравця** | `touch /tmp/game/chest_drawer/.closed` |
| `postcmd` | `rm -f /tmp/game/chest_drawer/.open` |

---

### `take/c` — Взяти предмет з контейнера

**TextWorld:** `take old key from chest drawer`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p /tmp/game/chest_drawer && cp /tmp/items/old_key /tmp/game/chest_drawer/ 2>/dev/null || true` |
| `test` | `test -f ~/old_key` |
| **Команда гравця** | `cp /tmp/game/chest_drawer/old_key ~/` |
| `postcmd` | `echo "old_key taken" >> /tmp/game/inventory.log` |

**Логіка:**
- Предмети спочатку в `/tmp/items/<item_name>/`
- `precmd` копіює предмет у контейнер
- Гравець копіює додому → предмет "взято"
- Перевірка: `test -f ~/old_key`

---

### `insert` — Покласти предмет у контейнер

**TextWorld:** `insert old key into chest drawer`

| Поле | Значення |
|------|----------|
| `test` | `test -f /tmp/game/chest_drawer/old_key && test ! -f ~/old_key` |
| **Команда гравця** | `cp ~/old_key /tmp/game/chest_drawer/` |
| `postcmd` | `rm ~/old_key` |

---

## 2. Робота з дверима (d)

### `open/d` — Відчинити двері

**TextWorld:** `open wooden door`

| Поле | Значення |
|------|----------|
| `precmd` | `echo "closed" > /tmp/game/door_wooden_door.state` |
| `test` | `test "$(cat /tmp/game/door_wooden_door.state)" = "open"` |
| **Команда гравця** | `echo "open" > /tmp/game/door_wooden_door.state` |
| `postcmd` | `echo "door_wooden_door: open" >> /tmp/game/doors.log` |

**Логіка:**
- Стан дверей = файл `.state` з текстом `closed`, `open`, `locked`
- Гравець перезаписує файл → двері відчинено

---

### `close/d` — Зачинити двері

**TextWorld:** `close wooden door`

| Поле | Значення |
|------|----------|
| `test` | `test "$(cat /tmp/game/door_wooden_door.state)" = "closed"` |
| **Команда гравця** | `echo "closed" > /tmp/game/door_wooden_door.state` |

---

### `unlock/d` — Відімкнути двері ключем

**TextWorld:** `unlock wooden door with old key`

| Поле | Значення |
|------|----------|
| `precmd` | `echo "locked" > /tmp/game/door_wooden_door.state` |
| `test` | `test "$(cat /tmp/game/door_wooden_door.state)" = "closed" && test -f ~/old_key` |
| **Команда гравця** | `echo "closed" > /tmp/game/door_wooden_door.state` (якщо ключ є) |
| `postcmd` | `touch /tmp/game/door_wooden_door.unlocked` |

**Логіка:**
- Двері спочатку `locked`
- Якщо гравець має ключ (`test -f ~/old_key`), може змінити на `closed`
- Потім `open/d` щоб відчинити

---

### `lock/d` — Замкнути двері ключем

**TextWorld:** `lock wooden door with old key`

| Поле | Значення |
|------|----------|
| `test` | `test "$(cat /tmp/game/door_wooden_door.state)" = "locked" && test -f ~/old_key` |
| **Команда гравця** | `echo "locked" > /tmp/game/door_wooden_door.state` |

---

## 3. Робота з поверхнями/опорами (s)

### `take/s` — Взяти предмет з поверхні

**TextWorld:** `take apple from stove`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p /tmp/game/stove && touch /tmp/game/stove/apple` |
| `test` | `test -f ~/apple` |
| **Команда гравця** | `cp /tmp/game/stove/apple ~/` |
| `postcmd` | `rm /tmp/game/stove/apple` |

---

### `put` — Покласти предмет на поверхню

**TextWorld:** `put apple on stove`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p /tmp/game/stove` |
| `test` | `test -f /tmp/game/stove/apple` |
| **Команда гравця** | `cp ~/apple /tmp/game/stove/` |
| `postcmd` | `rm ~/apple && touch /tmp/game/win_condition` |

---

## 4. Переміщення між кімнатами

### `go/east`, `go/west`, `go/north`, `go/south` — Рух

**TextWorld:** `go east`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p /tmp/rooms/bedroom /tmp/rooms/kitchen` |
| `test` | `test "$(cat ~/.current_room)" = "kitchen"` |
| **Команда гравця** | `echo "kitchen" > ~/.current_room` |
| `postcmd` | `echo "Moved to kitchen at $(date)" >> /tmp/game/movement.log` |

**Логіка:**
- Поточна кімната = файл `~/.current_room`
- Гравець змінює вміст → перейшов у нову кімнату

---

## 5. Автоматичні події

### `trigger` — Перевірка умови перемоги

| Поле | Значення |
|------|----------|
| `test` | `test -f /tmp/game/win_condition` |
| **Команда гравця** | (не потрібна — перевірка автоматична) |

---

## Повна таблиця відповідності

| TextWorld дія | Bash-команда гравця | Перевірка (test) |
|---------------|---------------------|------------------|
| `open container` | `rm /tmp/game/<container>/.closed` | `test ! -f ...` |
| `close container` | `touch /tmp/game/<container>/.closed` | `test -f ...` |
| `take item from container` | `cp /tmp/game/<container>/<item> ~/` | `test -f ~/<item>` |
| `take item from surface` | `cp /tmp/game/<surface>/<item> ~/` | `test -f ~/<item>` |
| `insert item into container` | `cp ~/<item> /tmp/game/<container>/` | `test -f /tmp/game/<container>/<item>` |
| `put item on surface` | `cp ~/<item> /tmp/game/<surface>/` | `test -f /tmp/game/<surface>/<item>` |
| `open door` | `echo "open" > /tmp/game/door_<name>.state` | `test "$(cat ...)" = "open"` |
| `close door` | `echo "closed" > /tmp/game/door_<name>.state` | `test "$(cat ...)" = "closed"` |
| `unlock door with key` | `echo "closed" > /tmp/game/door_<name>.state` | `test "$(cat ...)" = "closed" && test -f ~/<key>` |
| `lock door with key` | `echo "locked" > /tmp/game/door_<name>.state` | `test "$(cat ...)" = "locked" && test -f ~/<key>` |
| `go east/west/north/south` | `echo "<room>" > ~/.current_room` | `test "$(cat ~/.current_room)" = "<room>"` |

---

## Структура файлів гри

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

/tmp/items/                         # Початкові предмети (шаблони)
├── old_key
├── apple
├── milk
└── ...

~/.current_room                     # Поточна кімната гравця
~/old_key                           # Предмет в інвентарі гравця
```

---

## Приклад повного рівня

```yaml
name: step_03_unlock_wooden_door
test: test "$(cat /tmp/game/door_wooden_door.state)" = "closed" && test -f ~/old_key
next: [step_04_open_wooden_door]
precmd: |
  echo "locked" > /tmp/game/door_wooden_door.state
  echo "У вас має бути: old_key"
  ls ~/ | grep old_key || echo "⚠️  old_key не знайдено!"

## Крок 3/8: Відімкніть дерев'яні двері

Ви перед **дерев'яними дверима (wooden door)**.
Вони **замкнені**. У вас має бути **старий ключ (old_key)**.

Перевірте що ключ у вас:
```bash
ls ~/ | grep old_key
```

Відімкніть двері (ключ залишиться у вас):
```bash
echo "closed" > /tmp/game/door_wooden_door.state
```

Перевірте статус:
```bash
cat /tmp/game/door_wooden_door.state
```
```

---

## Обробка помилок

Якщо гравець забув попередній крок:

```bash
# Не має ключа
test -f ~/old_key || echo "❌ Потрібен old_key! Виконайте попередній крок."

# Двері не відімкнено
test "$(cat /tmp/game/door_wooden_door.state)" = "locked" && echo "⚠️  Двері замкнені!"

# Контейнер зачинено
test -f /tmp/game/chest_drawer/.closed && echo "⚠️  Контейнер зачинений!"
```

---

## Інструкція для гравця

Кожен рівень містить:
1. **Що робити** — опис дії українською
2. **Яку команду виконати** — конкретна bash-команда
3. **Як перевірити** — команда для самоперевірки
4. **Що має статися** — очікуваний результат

Це замінює `game_state.sh` на **чисті bash-команди** без додаткових скриптів.
