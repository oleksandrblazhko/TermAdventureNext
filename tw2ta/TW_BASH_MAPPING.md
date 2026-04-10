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
| `precmd` | `mkdir -p $HOME/.tw2ta_game/chest_drawer && touch $HOME/.tw2ta_game/chest_drawer/.closed` |
| `test` | `test ! -f $HOME/.tw2ta_game/chest_drawer/.closed` |
| **Команда гравця** | `rm $HOME/.tw2ta_game/chest_drawer/.closed` |
| `postcmd` | `touch $HOME/.tw2ta_game/chest_drawer/.open` |

**Логіка:**
- Контейнер = директорія `$HOME/.tw2ta_game/<container_name>/`
- `.closed` — файл-прапорець що контейнер зачинений
- Гравець видаляє `.closed` → контейнер відчинено

---

### `close/c` — Зачинити контейнер

**TextWorld:** `close chest drawer`

| Поле | Значення |
|------|----------|
| `test` | `test -f $HOME/.tw2ta_game/chest_drawer/.closed` |
| **Команда гравця** | `touch $HOME/.tw2ta_game/chest_drawer/.closed` |
| `postcmd` | `rm -f $HOME/.tw2ta_game/chest_drawer/.open` |

---

### `take/c` — Взяти предмет з контейнера

**TextWorld:** `take old key from chest drawer`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p $HOME/.tw2ta_game/chest_drawer && cp /tmp/items/old_key $HOME/.tw2ta_game/chest_drawer/ 2>/dev/null || true` |
| `test` | `test -f ~/old_key` |
| **Команда гравця** | `cp $HOME/.tw2ta_game/chest_drawer/old_key ~/` |
| `postcmd` | `echo "old_key taken" >> $HOME/.tw2ta_game/inventory.log` |

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
| `test` | `test -f $HOME/.tw2ta_game/chest_drawer/old_key && test ! -f ~/old_key` |
| **Команда гравця** | `cp ~/old_key $HOME/.tw2ta_game/chest_drawer/` |
| `postcmd` | `rm ~/old_key` |

---

## 2. Робота з дверима (d)

### `open/d` — Відчинити двері

**TextWorld:** `open wooden door`

| Поле | Значення |
|------|----------|
| `precmd` | `echo "closed" > $HOME/.tw2ta_game/door_wooden_door.state` |
| `test` | `test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "open"` |
| **Команда гравця** | `echo "open" > $HOME/.tw2ta_game/door_wooden_door.state` |
| `postcmd` | `echo "door_wooden_door: open" >> $HOME/.tw2ta_game/doors.log` |

**Логіка:**
- Стан дверей = файл `.state` з текстом `closed`, `open`, `locked`
- Гравець перезаписує файл → двері відчинено

---

### `close/d` — Зачинити двері

**TextWorld:** `close wooden door`

| Поле | Значення |
|------|----------|
| `test` | `test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "closed"` |
| **Команда гравця** | `echo "closed" > $HOME/.tw2ta_game/door_wooden_door.state` |

---

### `unlock/d` — Відімкнути двері ключем

**TextWorld:** `unlock wooden door with old key`

| Поле | Значення |
|------|----------|
| `precmd` | `echo "locked" > $HOME/.tw2ta_game/door_wooden_door.state` |
| `test` | `test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "closed" && test -f ~/old_key` |
| **Команда гравця** | `echo "closed" > $HOME/.tw2ta_game/door_wooden_door.state` (якщо ключ є) |
| `postcmd` | `touch $HOME/.tw2ta_game/door_wooden_door.unlocked` |

**Логіка:**
- Двері спочатку `locked`
- Якщо гравець має ключ (`test -f ~/old_key`), може змінити на `closed`
- Потім `open/d` щоб відчинити

---

### `lock/d` — Замкнути двері ключем

**TextWorld:** `lock wooden door with old key`

| Поле | Значення |
|------|----------|
| `test` | `test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "locked" && test -f ~/old_key` |
| **Команда гравця** | `echo "locked" > $HOME/.tw2ta_game/door_wooden_door.state` |

---

## 3. Робота з поверхнями/опорами (s)

### `take/s` — Взяти предмет з поверхні

**TextWorld:** `take apple from stove`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p $HOME/.tw2ta_game/stove && touch $HOME/.tw2ta_game/stove/apple` |
| `test` | `test -f ~/apple` |
| **Команда гравця** | `cp $HOME/.tw2ta_game/stove/apple ~/` |
| `postcmd` | `rm $HOME/.tw2ta_game/stove/apple` |

---

### `put` — Покласти предмет на поверхню

**TextWorld:** `put apple on stove`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p $HOME/.tw2ta_game/stove` |
| `test` | `test -f $HOME/.tw2ta_game/stove/apple` |
| **Команда гравця** | `cp ~/apple $HOME/.tw2ta_game/stove/` |
| `postcmd` | `rm ~/apple && touch $HOME/.tw2ta_game/win_condition` |

---

## 4. Переміщення між кімнатами

### `go/east`, `go/west`, `go/north`, `go/south` — Рух

**TextWorld:** `go east`

| Поле | Значення |
|------|----------|
| `precmd` | `mkdir -p $HOME/.tw2ta_game/rooms/bedroom $HOME/.tw2ta_game/rooms/kitchen` |
| `test` | `test "$(cat $HOME/.tw2ta_game/current_room)" = "kitchen"` |
| **Команда гравця** | `echo "kitchen" > $HOME/.tw2ta_game/current_room` |
| `postcmd` | `echo "Moved to kitchen at $(date)" >> $HOME/.tw2ta_game/movement.log` |

**Логіка:**
- Поточна кімната = файл `$HOME/.tw2ta_game/current_room`
- Гравець змінює вміст → перейшов у нову кімнату

---

## 5. Автоматичні події

### `trigger` — Перевірка умови перемоги

| Поле | Значення |
|------|----------|
| `test` | `test -f $HOME/.tw2ta_game/win_condition` |
| **Команда гравця** | (не потрібна — перевірка автоматична) |

---

## Повна таблиця відповідності

| TextWorld дія | Bash-команда гравця | Перевірка (test) |
|---------------|---------------------|------------------|
| `open container` | `rm $HOME/.tw2ta_game/<container>/.closed` | `test ! -f ...` |
| `close container` | `touch $HOME/.tw2ta_game/<container>/.closed` | `test -f ...` |
| `take item from container` | `cp $HOME/.tw2ta_game/<container>/<item> ~/` | `test -f ~/<item>` |
| `take item from surface` | `cp $HOME/.tw2ta_game/<surface>/<item> ~/` | `test -f ~/<item>` |
| `insert item into container` | `cp ~/<item> $HOME/.tw2ta_game/<container>/` | `test -f $HOME/.tw2ta_game/<container>/<item>` |
| `put item on surface` | `cp ~/<item> $HOME/.tw2ta_game/<surface>/` | `test -f $HOME/.tw2ta_game/<surface>/<item>` |
| `open door` | `echo "open" > $HOME/.tw2ta_game/door_<name>.state` | `test "$(cat ...)" = "open"` |
| `close door` | `echo "closed" > $HOME/.tw2ta_game/door_<name>.state` | `test "$(cat ...)" = "closed"` |
| `unlock door with key` | `echo "closed" > $HOME/.tw2ta_game/door_<name>.state` | `test "$(cat ...)" = "closed" && test -f ~/<key>` |
| `lock door with key` | `echo "locked" > $HOME/.tw2ta_game/door_<name>.state` | `test "$(cat ...)" = "locked" && test -f ~/<key>` |
| `go east/west/north/south` | `echo "<room>" > $HOME/.tw2ta_game/current_room` | `test "$(cat $HOME/.tw2ta_game/current_room)" = "<room>"` |

---

## Структура файлів гри

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

/tmp/items/                         # Початкові предмети (шаблони)
├── old_key
├── apple
├── milk
└── ...

~/old_key                           # Предмет в інвентарі гравця
```

---

## Приклад повного рівня

```yaml
name: step_03_unlock_wooden_door
test: test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "closed" && test -f ~/old_key
next: [step_04_open_wooden_door]
precmd: |
  echo "locked" > $HOME/.tw2ta_game/door_wooden_door.state
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
echo "closed" > $HOME/.tw2ta_game/door_wooden_door.state
```

Перевірте статус:
```bash
cat $HOME/.tw2ta_game/door_wooden_door.state
```
```

---

## Обробка помилок

Якщо гравець забув попередній крок:

```bash
# Не має ключа
test -f ~/old_key || echo "❌ Потрібен old_key! Виконайте попередній крок."

# Двері не відімкнено
test "$(cat $HOME/.tw2ta_game/door_wooden_door.state)" = "locked" && echo "⚠️  Двері замкнені!"

# Контейнер зачинено
test -f $HOME/.tw2ta_game/chest_drawer/.closed && echo "⚠️  Контейнер зачинений!"
```

---

## Інструкція для гравця

Кожен рівень містить:
1. **Що робити** — опис дії українською
2. **Яку команду виконати** — конкретна bash-команда
3. **Як перевірити** — команда для самоперевірки
4. **Що має статися** — очікуваний результат

Це замінює `game_state.sh` на **чисті bash-команди** без додаткових скриптів.
