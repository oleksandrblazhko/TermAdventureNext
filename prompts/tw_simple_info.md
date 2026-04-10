# Огляд гри (Game Overview) типу tw_simple

**тема (Theme):** Будинок (House)
**Квести (Quests):** 2 (1 основний + 1 умова поразки (fail condition))

### Приклад команди генерації опису гри
```bash
tw-make tw-simple --goal brief --seed 5 --output simple_game.z8 --rewards sparse
```
**Параметри:**
- `tw-simple` — тип шаблону гри;
- `--goal brief` — стиль опису мети квесту;
- `--seed 5` — випадкове зерно для процедурної генерації;
- `--output simple_game.z8` — вихідний файл гри (формат Z-machine);
- `--rewards sparse` — тип системи винагород.

---

## Карта світу — розташування кімнат (World Map - Room Layout)

```
                    ┌─────────────┐
                    │   r_2       │
                    │ Вітальня    │
                    │   (clean)   │
                    └──────┬──────┘
                           │ north_of
                           │
┌─────────────┐    ┌──────┴──────┐    ┌─────────────┐
│   r_0       │◄──►│   r_1       │◄──►│   r_4       │
│ Спальня     │ d_0│ Кухня       │ d_1│ Задній двір │
│ (storage)   │    │ (work)      │    │ (cook)      │
└─────────────┘    └──────┬──────┘    └──────┬──────┘
                          │                   │
                   south_of│            north_of│
                          │                   │
                    ┌─────┴──────┐    ┌───────┴─────┐
                    │   r_3      │    │   r_5       │
                    │ Ванна      │    │ Сад         │
                    │ (work)     │    │ (cook)      │
                    └────────────┘    └─────────────┘
```

### З'єднання кімнат (Room Connections)

| Звідки (From) | Напрямок (Direction) | двері (Door) | Куди (To) | Примітки (Notes) |
|------|-----------|------|----|----------|
| r_0 (Спальня (Bedroom)) | схід (east) | d_0 (дерев'яні двері (wooden door)) | r_1 (Кухня (Kitchen)) | **замкнені (locked)** спочатку |
| r_1 (Кухня (Kitchen)) | захід (west) | d_0 (дерев'яні двері (wooden door)) | r_0 (Спальня (Bedroom)) | **замкнені (locked)** спочатку |
| r_1 (Кухня (Kitchen)) | схід (east) | d_1 (двері з сіткою (screen door)) | r_4 (Задній двір (Backyard)) | **зачинені (closed)** спочатку |
| r_4 (Задній двір (Backyard)) | захід (west) | d_1 (двері з сіткою (screen door)) | r_1 (Кухня (Kitchen)) | **зачинені (closed)** спочатку |
| r_1 (Кухня (Kitchen)) | північ (north) | — | r_2 (Вітальня (Living Room)) | без перешкод (unguarded) |
| r_2 (Вітальня (Living Room)) | південь (south) | — | r_1 (Кухня (Kitchen)) | без перешкод (unguarded) |
| r_1 (Кухня (Kitchen)) | південь (south) | — | r_3 (Ванна (Bathroom)) | без перешкод (unguarded) |
| r_3 (Ванна (Bathroom)) | північ (north) | — | r_1 (Кухня (Kitchen)) | без перешкод (unguarded) |
| r_4 (Задній двір (Backyard)) | південь (south) | — | r_5 (Сад (Garden)) | вільний прохід (unblocked) |
| r_5 (Сад (Garden)) | північ (north) | — | r_4 (Задній двір (Backyard)) | вільний прохід (unblocked) |

---

## Сутності та їх початкове розташування (Entities and Initial Locations)

### Контейнери (Containers (c))

| ID | Назва (Name) | Розташування (Location) | початковий стан (Initial State) |
|----|-------|--------------|-----------------|
| c_0 | шухляда (chest drawer) | r_0 (Спальня (Bedroom)) | **зачинена (closed)** |
| c_1 | старовинний сундук (antique trunk) | r_0 (Спальня (Bedroom)) | **зачинений (closed)** |
| c_2 | холодильник (refrigerator) | r_1 (Кухня (Kitchen)) | **зачинений (closed)** |
| c_3 | унітаз (toilet) | r_3 (Ванна (Bathroom)) | **зачинений (closed)** |
| c_4 | ванна (bath) | r_3 (Ванна (Bathroom)) | **зачинена (closed)** |

### поверхні/опори (Supporters/Surfaces (s))

| ID | Назва (Name) | Розташування (Location) |
|----|-------|--------------|
| s_0 | ліжко king-size (king-size bed) | r_0 (Спальня (Bedroom)) |
| s_1 | стільниця (counter) | r_1 (Кухня (Kitchen)) |
| s_2 | плита (stove) | r_1 (Кухня (Kitchen)) |
| s_3 | кухонний острів (kitchen island) | r_1 (Кухня (Kitchen)) |
| s_4 | раковина (sink) | r_3 (Ванна (Bathroom)) |
| s_5 | диван (couch) | r_2 (Вітальня (Living Room)) |
| s_6 | журнальний столик (low table) | r_2 (Вітальня (Living Room)) |
| s_7 | телевізор (tv) | r_2 (Вітальня (Living Room)) |
| s_8 | барбекю (bbq) | r_4 (Задній двір (Backyard)) |
| s_9 | садовий столик (patio table) | r_4 (Задній двір (Backyard)) |
| s_10 | набір стільців (set of chairs) | r_4 (Задній двір (Backyard)) |

### предмети (Items)

| ID | тип (Type) | Назва (Name) | початкове розташування (Initial Location) |
|----|-----|-------|------------------------|
| k_0 | ключ (key) | старий ключ (old key) | **у c_0** (шухляда (chest drawer)) |
| f_0 | їжа (food) | молоко (milk) | у c_2 (холодильник (refrigerator)) |
| f_1 | їжа (food) | яблуко (apple) | **у c_2** (холодильник (refrigerator)) |
| f_2 | їжа (food) | — | на s_5 (диван (couch)) |
| f_3 | їжа (food) | кущ помідорів (tomato plant) | r_5 (Сад (Garden)) |
| f_4 | їжа (food) | перець (bell pepper) | r_5 (Сад (Garden)) |
| f_5 | їжа (food) | салат-латук (lettuce) | r_5 (Сад (Garden)) |
| o_0 | об'єкт (object) | записка (note) | на s_4 (раковина (sink)) |
| o_1 | об'єкт (object) | зубна щітка (toothbrush) | у c_4 (ванна (bath)) |
| o_2 | об'єкт (object) | шматок мила (soap bar) | на s_6 (журнальний столик (low table)) |
| o_3 | об'єкт (object) | лопата (shovel) | r_5 (Сад (Garden)) |
| o_4 | об'єкт (object) | пульт (remote) | на s_3 (кухонний острів (kitchen island)) |

### двері (Doors)

| ID | Назва (Name) | З'єднує (Connects) | початковий стан (Initial State) |
|----|-------|---------|-----------------|
| d_0 | дерев'яні двері (wooden door) | r_0 ↔ r_1 | **замкнені (locked)** |
| d_1 | двері з сіткою (screen door) | r_1 ↔ r_4 | **зачинені (closed)** |

### Відповідність ключів і дверей (Key-Door Matching)

| Ключ (Key) | Відмикає (Unlocks) |
|------|----------|
| k_0 | d_0 (дерев'яні двері (wooden door)) |

---

## Квест 1 — Основна мета (Quest 1 - Main Objective) (Винагорода (Reward): 1)

**Мета (Goal):** покласти яблуко на плиту (Put apple on stove)

**потрібні команди (Required Commands):**
1. open chest drawer (відчинити шухляду)
2. take old key from chest drawer (взяти старий ключ із шухляди)
3. unlock wooden door with old key (відімкнути дерев'яні двері старим ключем)
4. open wooden door (відчинити дерев'яні двері)
5. go east (йти на схід)
6. open refrigerator (відчинити холодильник)
7. take apple from refrigerator (взяти яблуко з холодильника)
8. put apple on stove (покласти яблуко на плиту)

### Граф переходів дій (Action Transition Graph)

```
поЧАТОК (START)
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 1: open/c (відчинити шухляду (open chest drawer))      │
│                                                             │
│ Команда (Command): "open {c_0}"                             │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_0)                    • at(P, r_0)                │
│ • at(c_0, r_0)                  • at(c_0, r_0)              │
│ • closed(c_0)                   • open(c_0)                 │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 2: take/c (взяти ключ із шухляди                       │
│          (take old key from chest drawer))                  │
│                                                             │
│ Команда (Command): "take {k_0} from {c_0}"                  │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_0)                    • at(P, r_0)                │
│ • at(c_0, r_0)                  • at(c_0, r_0)              │
│ • open(c_0)                     • open(c_0)                 │
│ • in(k_0, c_0)                  • in(k_0, I) ← ключ         │
│                                    отримано (key acquired)  │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 3: unlock/d (відімкнути дерев'яні двері ключем         │
│          (unlock wooden door with old key))                 │
│                                                             │
│ Команда (Command): "unlock {d_0} with {k_0}"                │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_0)                    • at(P, r_0)                │
│ • link(r_0, d_0, r_1)           • link(r_0, d_0, r_1)       │
│ • link(r_1, d_0, r_0)           • link(r_1, d_0, r_0)       │
│ • in(k_0, I)                    • in(k_0, I)                │
│ • match(k_0, d_0)               • match(k_0, d_0)           │
│ • locked(d_0)                   • closed(d_0) ← відімкнено! │
│                                    (unlocked!)              │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 4: open/d (відчинити дерев'яні двері                   │
│          (open wooden door))                                │
│                                                             │
│ Команда (Command): "open {d_0}"                             │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_0)                    • at(P, r_0)                │
│ • link(r_0, d_0, r_1)           • link(r_0, d_0, r_1)       │
│ • link(r_1, d_0, r_0)           • link(r_1, d_0, r_0)       │
│ • closed(d_0)                   • open(d_0)                 │
│                                 • free(r_0, r_1) ← прохід   │
│                                    (passable)               │
│                                 • free(r_1, r_0)            │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 5: go/east (перейти на кухню (move to kitchen))        │
│                                                             │
│ Команда (Command): "go east"                                │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_0)                    • west_of(r_0, r_1)         │
│ • west_of(r_0, r_1)             • free(r_0, r_1)            │
│ • free(r_0, r_1)                • free(r_1, r_0)            │
│ • free(r_1, r_0)                • at(P, r_1) ← нова кімната!│
│                                    (new room!)              │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 6: open/c (відчинити холодильник                       │
│          (open refrigerator))                               │
│                                                             │
│ Команда (Command): "open {c_2}"                             │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_1)                    • at(P, r_1)                │
│ • at(c_2, r_1)                  • at(c_2, r_1)              │
│ • closed(c_2)                   • open(c_2)                 │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 7: take/c (взяти яблуко з холодильника                 │
│          (take apple from refrigerator))                    │
│                                                             │
│ Команда (Command): "take {f_1} from {c_2}"                  │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_1)                    • at(P, r_1)                │
│ • at(c_2, r_1)                  • at(c_2, r_1)              │
│ • open(c_2)                     • open(c_2)                 │
│ • in(f_1, c_2)                  • in(f_1, I) ← яблуко       │
│                                    взято! (apple taken!)    │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ Крок 8: put (покласти яблуко на плиту)                      │
│         ← УМОВА пеРЕМОГИ (WIN CONDITION)                    │
│                                                             │
│ Команда (Command): "put {f_1} on {s_2}"                     │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • at(P, r_1)                    • at(P, r_1)                │
│ • at(s_2, r_1)                  • at(s_2, r_1)              │
│ • in(f_1, I)                    • on(f_1, s_2) ← пеРЕМОГА!  │
│                                    (QUEST WIN!)             │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
┌─────────────────────────────────────────────────────────────┐
│ подія TRIGGER (перевірка умови перемоги                     │
│          (win condition check))                             │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • on(f_1, s_2)                  • on(f_1, s_2)              │
│                                 • event(f_1, s_2) ←         │
│                                    пеРЕМОГА! (WIN!)         │
└─────────────────────────────────────────────────────────────┘
  │
  ▼
КВЕСТ ЗАВЕРШЕНО (QUEST COMPLETE) (Винагорода (Reward): 1)
```

---

## Квест 2 — Умова поразки (Quest 2 - Fail Condition) (Винагорода (Reward): 0)

**тип (Type):** тригер поразки (Failure trigger)
**Умова (Condition):** З'їсти яблуко (f_1) (Eating the apple (f_1))

```
┌─────────────────────────────────────────────────────────────┐
│ подія ПОРАЗКИ (FAIL EVENT): eaten(f_1)                      │
│                                                             │
│ Якщо гравець з'їсть яблуко, квест провалено!                │
│ (If the player eats the apple, quest fails!)                │
│                                                             │
│ пеРЕДУМОВИ (PRECONDITIONS):     НАСЛІдКИ (POSTCONDITIONS):  │
│ • —                             • eaten(f_1)                │
│                                 • event(f_1) ← ПОРАЗКА!     │
│                                    (FAIL!)                  │
└─────────────────────────────────────────────────────────────┘
```

**призначення (Purpose):** Запобігає тому, щоб гравець з'їв яблуко замість того, щоб покласти його на плиту (Prevents the player from eating the apple instead of putting it on the stove).

---

## підсумок скінченного автомата (State Machine Summary)

### Критичний шлях — стани (Critical Path States)

```
[початковий стан (Initial State)]
  ├─ P у r_0 (Спальня (Bedroom))
  ├─ k_0 у c_0 (шухляда (chest drawer))
  ├─ c_0 зачинена (closed)
  ├─ d_0 замкнені (locked)
  ├─ f_1 у c_2 (холодильник (refrigerator))
  └─ c_2 зачинений (closed)
       │
       ▼ [Крок 1: відчинити c_0 (open c_0)]
       │ c_0 тепер відчинена (c_0 now open)
       │
       ▼ [Крок 2: взяти k_0 (take k_0)]
       │ k_0 тепер в інвентарі (k_0 now in Inventory)
       │
       ▼ [Крок 3: відімкнути d_0 (unlock d_0)]
       │ d_0 тепер зачинені, не замкнені (d_0 now closed (not locked))
       │
       ▼ [Крок 4: відчинити d_0 (open d_0)]
       │ d_0 тепер відчинені (d_0 now open)
       │ free(r_0, r_1) = true
       │
       ▼ [Крок 5: йти на схід (go east)]
       │ P тепер у r_1 (Кухня (Kitchen))
       │
       ▼ [Крок 6: відчинити c_2 (open c_2)]
       │ c_2 тепер відчинений (c_2 now open)
       │
       ▼ [Крок 7: взяти f_1 (take f_1)]
       │ f_1 тепер в інвентарі (f_1 now in Inventory)
       │
       ▼ [Крок 8: покласти f_1 на s_2 (put f_1 on s_2)]
       │ f_1 на s_2 (плита (stove))
       │
       ▼ [TRIGGER]
       │ event(f_1, s_2) = пеРЕМОГА! (WIN!)
       │
       ▼
[КВЕСТ ЗАВЕРШЕНО (QUEST COMPLETE)]
```

### Залежності дій (Action Dependencies)

```
open(c_0)
    │
    └──► take(k_0, c_0)
              │
              └──► unlock(d_0, k_0)
                        │
                        └──► open(d_0)
                                  │
                                  └──► go/east (r_0 → r_1)
                                              │
                                              ├──► open(c_2)
                                              │       │
                                              │       └──► take(f_1, c_2)
                                              │                   │
                                              └───────────────────┴──► put(f_1, s_2)
                                                                              │
                                                                              └──► пеРЕМОГА (WIN)
```

### Зворотні дії (доступні) (Reverse Actions (Available))

Кожна дія має зворотну (Every action has a reverse):
- `open/c` ↔ `close/c`
- `take/c` ↔ `insert`
- `unlock/d` ↔ `lock/d`
- `open/d` ↔ `close/d`
- `go/east` ↔ `go/west`
- `put` ↔ `take/s`

Це означає, що гравець може **скасувати** будь-яку дію, але це може завадити завершенню квесту (This means the player can **undo** any action, but doing so may prevent quest completion).

---

## Статистика графу (Graph Statistics)

| Метрика (Metric) | Значення (Value) |
|---------|----------|
| Всього кімнат (Total rooms) | 6 |
| Всього контейнерів (Total containers) | 5 |
| Всього поверхонь (Total supporters) | 11 |
| Всього предметів (Total items) | 13 |
| Всього дверей (Total doors) | 2 |
| Кроків основного квесту (Main quest steps) | 8 |
| Умов поразки (Fail conditions) | 1 |
| типів дій (Action types) | 13 |
| Винагорода за перемогу (Win reward) | 1 |

---

## Як взаємодіяти з грою (How to Interact with the Game)

### Запуск гри (Launching the Game)

Файл `.z8` — це скомпільована гра для віртуальної машини Z-machine (Z-machine virtual machine). Потрібен інтерпретатор (interpreter required), наприклад **Frotz**:

```bash
frotz simple_game.z8
```

### Як гравець дізнається команди (How the Player Discovers Commands)

Гравець **не бачить JSON-файл** (The player does **not** see the JSON file) — він отримує лише текстові описи кімнат та об'єктів, як у класичних текстових квестах (Interactive Fiction).

#### 1. Стандартні команди Interactive Fiction (Standard Interactive Fiction Commands)

Існує усталений набір команд, які працюють у більшості Z-machine ігор (There is a standard set of commands that work in most Z-machine games):

| Команда (Command) | Скорочення (Shortcut) | Дія (Action) |
|---------|------------|-----|
| `look` | `l` | Описати кімнату (Describe the room) |
| `inventory` | `i` | Що в кишенях (What's in pockets) |
| `go <direction>` | `n`, `s`, `e`, `w` | Йти (Go) |
| `take <obj>` | — | Взяти (Take) |
| `open <obj>` | — | Відчинити (Open) |
| `close <obj>` | — | Зачинити (Close) |
| `unlock <door> with <key>` | — | Відімкнути (Unlock) |
| `put <obj> on <surface>` | — | Покласти на (Put on) |
| `insert <obj> into <container>` | — | Покласти в (Insert into) |
| `examine <obj>` | `x <obj>` | Оглянути (Examine) |
| `eat <obj>` | — | З'їсти (Eat) |

#### 2. Текстові підказки у грі (In-Game Text Hints)

Коли гравець вводить `look`, він бачить опис кімнати з переліком об'єктів (When the player types `look`, they see a room description listing objects):

```
Bedroom
You are in a bedroom. A chest drawer and an antique trunk are here.
You also see a king-size bed.
```

Це **неявно** підказує (This **implicitly** hints):
- `chest drawer` — можна `open`, `take` з нього (can `open`, `take` from it)
- `antique trunk` — те саме (same thing)
- `king-size bed` — поверхня, можна `put` / `on` (a surface, can `put` / `on`)

#### 3. Метод спроб і помилок (Trial and Error)

Гравець пробує логічні дії, і гра дає зворотний зв'язок (The player tries logical actions, and the game gives feedback):

```
> open chest drawer     → "You open the chest drawer."
> look                  → "You see an old key inside."
> take old key          → "Taken."
> go east               → "The door is locked."  ← підказка: треба unlock (hint: need to unlock)
> unlock wooden door with old key  → "Unlocked."
```

Кожна відповідь гри дає **зворотний зв'язок**, який направляє далі (Every game response gives **feedback** that guides the next step).

### Правила, закладені в TextWorld (Rules Embedded in TextWorld)

З JSON видно, які дії **дозволені** через `preconditions` (The JSON shows which actions are **allowed** via `preconditions`):

```
дія take/c можлива ТІЛЬКИ якщо (take/c action is possible ONLY if):
  • at(P, r_0) — гравець у кімнаті з контейнером (player in room with container)
  • open(c_0)  — контейнер відчинений (container is open)
  • in(k_0, c_0) — ключ всередині (key is inside)
```

Тобто гравець має спочатку `open`, потім `take` — інший порядок не спрацює (So the player must `open` first, then `take` — a different order won't work).

### Що гравець НЕ знає (What the Player Does NOT Know)

- Немає списку квестів (No quest list)
- Немає підказки "зроби 8 кроків" (No "do 8 steps" hint)
- `fail_events` (з'їсти яблуко) — гравець дізнається тільки після програшу (fail events — player only learns after failing)

Гравець орієнтується виключно на (The player relies solely on):
1. **Описи кімнат** (`look`) — **Room descriptions**
2. **Описи об'єктів** (`examine`) — **Object descriptions**
3. **Зворотний зв'язок** від невдалих дій ("The door is locked") — **Feedback** from failed actions
4. **Логіку** та здоровий глузд — **Logic** and common sense

