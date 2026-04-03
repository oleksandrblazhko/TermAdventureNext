# TermAdventure — Повна документація

## Огляд проекту

**TermAdventure** — це Go-бібліотека для створення ретро-текстових пригод у терміналі *nix. Вона надає фреймворк для інтерактивних shell-квестів, де користувач проходить рівні, виконуючи команди у bash (навігація, робота з файлами, пошук тощо).

Проект працює за такою схемою:
1. Рівні челенджу описуються у `.ta` файлах (YAML-метадані + Markdown текст, розділені горизонтальними лініями)
2. Скомпільований Go-бінарник читає `.ta` файл і керує прогресом
3. Інтеграція з bash через кастомний `ta_bashrc` — відслідковує прогрес і показує ідентифікатор рівня у промпті

---

## Структура директорій

```
TermAdventure/
├── main.go              # Точка входу; CLI прапори, завантаження челенджу, шифрування/дешифрування
├── levels/              # Основна бібліотека
│   ├── levels.go        # Типи Challenge/Level, парсинг YAML, хешування ID рівнів, управління конфігурацією
│   ├── utils.go         # Утиліти: друк у термінал, форматування markdown, AES шифрування, виконання shell-команд
│   └── template.go      # Генерація .ta файлів з Go шаблонів + YAML змінних
├── challenger.sh        # Shell-скрипт для запуску сесії челенджу
├── ta_bashrc            # Кастомний bashrc для сесій челенджу (промпт, історія, аліаси)
├── sample_challenge.ta  # Приклад челенджу з декількома рівнями
├── sample_challenge.yaml # Приклад YAML даних для генерації шаблонів
├── sample_challenge.tpl # Приклад шаблону
├── sample_challenge.gta.enc # Зашифрований приклад челенджу
├── print_test.go        # Тест крайнього випадку друку тексту
├── docs/                # Документація (Sphinx)
│   ├── Makefile
│   └── source/conf.py
└── README.md
```

---

## Формат файлів `.ta`

### Структура

Рівні розділяються горизонтальними лініями Markdown (`---` з 10+ дефісами):

```
name: level_name
test: test "$(pwd)" = "/tmp"
next: [level2]

Текст рівня у форматі **Markdown**.

--------------------

name: level2
test: true
...
```

### Розділення метаданих і тексту

Метадані (YAML) і текст розділяються **двома новими рядками** (`\n\n`):

```
<YAML метадані>
\n\n
<Markdown текст рівня>
```

### Поля метаданих рівня

| Поле | YAML ключ | Опис |
|------|-----------|------|
| `name` | — | Унікальний ідентифікатор рівня |
| `test` | `test` | Shell-команда; якщо exit code = 0, рівень пройдено |
| `next` | `next` | Список наступних рівнів (підтримує гілкування) |
| `precmd` | `precmd` | Команда, що виконується ПЕРЕД початком рівня |
| `postcmd` | `postcmd` | Команда, що виконується ПІСЛЯ проходження рівня |
| `postprintcmd` | `postprintcmd` | Команда, що виконується ПІСЛЯ друку тексту рівня |
| `bgjobs` | `bgjobs` | Boolean — чи використовує рівень фонові завдання |
| `timelimit` | `timelimit` | **int** — ліміт часу на рівень у секундах (0 = без обмежень) |

**Приклад використання `timelimit`:**

```yaml
name: quick_task
test: test -f /tmp/report.txt
timelimit: 300   # 5 хвилин

У вас є 5 хвилин щоб створити файл /tmp/report.txt!
```

---

## Детальні алгоритми роботи

### 1. Парсинг `.ta` файлу

**Файл:** `levels/levels.go`, функція `LoadFromString()`

**Алгоритм:**

```
1. Отримати вміст файлу як рядок
2. Застосувати регулярний вираз: (?s)(.*?)\n\n------------+\n
   - (?s) — режим dotall (. захоплює \n)
   - (.*?) — ліниве захоплення всього до роздільника
   - \n\n------------+\n — два нових рядки + 10+ дефісів + новий рядок
3. Для кожного знайденого блоку:
   a. Розділити по \n\n → частини [0] = YAML, частини [1:] = текст
   b. YAML.Unmarshal() → структура Level
   c. Додати рівень до Challenge.Levels
4. Встановити поточний рівень = перший у списку
```

**Функція `buildLevel()`:**

```go
func buildLevel(text string) Level {
    parts := strings.Split(text, "\n\n")
    metadata := parts[0]
    clean_text := strings.Join(parts[1:len(parts)], "\n\n")

    level := Level{}
    yaml.Unmarshal([]byte(metadata), &level)
    level.Text = clean_text
    return level
}
```

### 2. Генерація ID рівня (хешування)

**Файл:** `levels/levels.go`, функції `LevelToID()`, `GetMD5Hash()`

**Формат вхідного рядка:**

```
i<назва_челенджу>j<назва_рівня>k<домашня_директорія>l
```

Наприклад: `isample_challengejl00k/home/blazkol`

**Алгоритм:**

```
1. Отримати домашню директорію поточного користувача (user.Current().HomeDir)
2. Сформувати рядок: fmt.Sprintf("i%sj%dk%sl", challenge_name, level, homeDir)
3. Обчислити MD5 хеш:
   a. hasher := md5.New()
   b. hasher.Write([]byte(text))
   c. hex.EncodeToString(hasher.Sum(nil))
4. Повернути 32-символьний hex рядок
```

**Ключові наслідки:**
- Хеш **унікальний для кожного користувача** (бо включає home directory)
- Один і той самий рівень дасть різні ID для різних студентів
- Це ускладнює списування — не можна передати хеш сусідові

### 3. Перевірка поточного рівня

**Файл:** `levels/levels.go`, функція `CheckCurrentLevel()`

**Алгоритм:**

```
1. Знайти поточний рівень за ID (c.IDToLevel(*c.CurrentLevel))
2. Якщо testcmd порожній → повернути true
3. Інакше виконати: exec.Command("/bin/bash", "-c", testcmd).Output()
4. Якщо exit code = 0 → рівень пройдено
5. Якщо в testcmd є вивід — спробувати інтерпретувати його як назву наступного рівня
   (але зазвичай testcmd — це просто перевірка, вивід ігнорується)
```

**Функція `CmdOK()` з `utils.go`:**

```go
func CmdOK(cmd string) (bool, string) {
    if cmd == "" {
        return true, ""
    }
    output, err := exec.Command(DefaultShell, "-c", cmd).Output()
    return err == nil, string(output)
}
```

### 4. Перехід на наступний рівень

**Файл:** `levels/levels.go`, функція `GoToNextLevel()`

**Алгоритм:**

```
1. Знайти поточний рівень
2. Виконати PostLevelCmd поточного рівня
3. Випадково обрати наступний рівень зі списку NextLevels:
   - rand.Seed(time.Now().UTC().UnixNano())
   - i := rand.Intn(len(c.Levels[index].NextLevels))
   - next_level = c.Levels[index].NextLevels[i]
4. Виконати PreLevelCmd нового рівня
5. Згенерувати новий ID: LevelToID(next_level, challenge_name)
6. Зберегти у конфігурацію: c.SetConfigVal("level", new_id)
7. Встановити last_level_printed = "no"
```

**Гілкування:** Якщо `next: [path_a, path_b, path_c]` — один обирається **випадково**.

### 5. Шифрування / Дешифрування

**Файл:** `levels/utils.go`, функції `Encrypt()`, `Decrypt()`

**Алгоритм:** AES-128/256 у режимі CFB (Cipher Feedback)

**Шифрування `Encrypt()`:**

```
1. Створити AES cipher: aes.NewCipher(key)
2. Доповнити повідомлення до блоку (padding PKCS7):
   - padding = aes.BlockSize - len(msg) % aes.BlockSize
   - padtext = bytes.Repeat([padding], padding)
3. Згенерувати випадковий IV (Initialization Vector): io.ReadFull(rand.Reader, iv)
4. Зашифрувати: cfb.XORKeyStream(ciphertext[aes.BlockSize:], msg)
5. Закодувати у base64 URL-safe: base64.URLEncoding.EncodeToString(ciphertext)
```

**Дешифрування `Decrypt()`:**

```
1. Створити AES cipher: aes.NewCipher(key)
2. Декодувати base64: base64.URLEncoding.DecodeString(text)
3. Виділити IV: decodedMsg[:aes.BlockSize]
4. Розшифрувати: cfb.XORKeyStream(msg, msg)
5. Прибрати padding:
   - unpadding = int(msg[length-1])
   - повернути msg[:(length - unpadding)]
```

**Використання у main.go:**

```
1. Якщо файл має суфікс .enc → автоматично дешифрується
2. Якщо прапор --enc → зашифрувати і вивести
3. Якщо прапор --dec → дешифрувати і вивести
```

### 6. Друк тексту у термінал

**Файл:** `levels/utils.go`, функції `PrintText()`, `PrettyPrintText()`, `print_line()`

**Алгоритм:**

```
1. Конвертувати Markdown у ANSI: MarkdownToTerminal()
   - **bold** → \033[1m...\033[0m
   - *italic* → \033[7m...\033[0m (inverse)
   - # Header → \033[1;4m...\033[0m (underline bold)
2. Якщо pretty_print = false:
   a. Встановити термінал у cbreak режим: stty cbreak min 1
   b. Вимкнути відображення вводу: stty -echo
   c. Читати по одному байту з stdin
3. Для кожного символу:
   a. Надрукувати символ
   b. Перевірити чи натиснуто Space (32) або Enter (10)
   c. Якщо так → вимкнути затримку між символами (пропуск тексту)
   d. Інакше → Sleep(print_sleep_time * ms)
   e. Для '.', '!', '?' → затримка x10 (ефект паузи)
4. Відновити налаштування терміналу при виході
```

**Обробка сигналів:**

```go
sigs := make(chan os.Signal, 1)
signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sigs
    if echo_state == false {
        exec.Command("stty", "-F", "/dev/tty", "echo").Run()
    }
    os.Exit(0)
}()
```

При Ctrl+C відновлюється відображення вводу перед виходом.

### 7. Markdown → Terminal форматування

**Файл:** `levels/utils.go`, функція `MarkdownToTerminal()`

**Підтримувані елементи:**

| Markdown | ANSI код | Результат |
|----------|----------|-----------|
| `**text**` | `\033[1m` | Жирний |
| `*text*` | `\033[7m` | Інверсія |
| `# Header` | `\033[1;4m` | Підкреслений жирний |

**Регулярні вирази:**

```go
bold_regex, _   := regexp.Compile(`\*\*([^\*]+)\*\*`)   // **text**
italic_regex, _ := regexp.Compile(`\*([^\*]+)\*`)        // *text*
header_regex, _ := regexp.Compile(`^\s*\#+\s*(.+)`)     // # Header
```

### 8. Система конфігурації

**Файл:** `levels/levels.go`, функції `NewChallenge()`, `SetConfigVal()`, `LoadCfg()`

**Бібліотека:** `github.com/rakyll/globalconf`

**Де зберігається:** `$HOME/.config/<challenge_name>/`

**Які дані:**

| Ключ | Значення | Коли встановлюється |
|------|----------|---------------------|
| `level` | MD5 хеш поточного рівня | При старті, при переході |
| `last_level_printed` | `"yes"` або `"no"` | Після друку рівня |

**Алгоритм збереження:**

```go
func (c *Challenge) SetConfigVal(name string, value string) {
    fint := &flagValue{str: value}
    f := &flag.Flag{Name: name, Value: fint}
    c.conf.Set("", f)
}
```

Використовується `flagValue` — кастомний тип, щоб зберегти рядок через систему flag + globalconf.

### 9. Робота `challenger.sh`

**Файл:** `challenger.sh`

**Алгоритм:**

```
1. Визначити CURRENT_DIR = директорія де знаходиться скрипт
   - $( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
2. Встановити TA_BIN (дефолт: $CURRENT_DIR/term-adventure)
   - ${TA_BIN:-$CURRENT_DIR/term-adventure}
3. Перетворити CHALLENGE_FILE на абсолютний шлях:
   - Якщо починається з / → залишити як є
   - Інакше → $(cd "$(dirname file)" && pwd)/$(basename file)
4. Перевірити що файл завдання існує
5. Витягти CHALLENGE_NAME = basename без розширення
   - $(basename "$CHALLENGE_FILE" | sed 's/\.[^.]*$//')
6. Експортувати TA_BIN, CHALLENGE_FILE
7. Запустити: bash --rcfile $CURRENT_DIR/ta_bashrc
8. Після виходу з bash:
   - Видалити $HOME/.tahistory
   - Видалити $HOME/.config/$CHALLENGE_NAME
   - Скасувати експортовані змінні
```

### 10. Робота `ta_bashrc`

**Файл:** `ta_bashrc`

**Алгоритм:**

```
1. Увімкнути cmdhist (багаторядкові команди = один запис в історії)
   - shopt -s cmdhist
2. Визначити PROMPT_COMMAND = prompt_command
3. Функція prompt_command виконується ПЕРЕД кожним промптом:
   a. history -a → зберегти історію у файл
   b. $TA_BIN $CHALLENGE_FILE → перевірити рівень, надрукувати якщо змінено
   c. export PS1='[$($TA_BIN --print-identifier $CHALLENGE_FILE)][\u@\H:\w]$ '
4. Історія:
   - HISTFILESIZE=500, HISTSIZE=500
   - HISTFILE=$HOME/.tahistory
   - HISTCONTROL=ignoredups
5. Колірний промпт: force_color_prompt=yes
6. Аліаси: ls, ll, la, grep тощо
```

### 11. Головний цикл програми (main.go)

**Файл:** `main.go`, функція `main()`

**Алгоритм:**

```
1. Парсинг CLI прапорів (flag.Parse())
2. Обробка спеціальних режимів (кожен викликає os.Exit(0) після виконання):
   - --generate-from-template → генерація .ta з шаблону
   - --enc → шифрування файлу
   - --dec → дешифрування файлу
   - --print → друк всіх рівнів
   - --detect-level → декодування хешу у назву рівня
   - --print-identifier → друк [challenge level]
   - --print-level → друк назви рівня
   - --print-challenge → друк назви челенджу
   - --print-current-level → друк тексту поточного рівня
   - --check-current-level → exit code 0/1
   - --background-jobs → exit code 0/1
3. Якщо жоден спеціальний режим:
   a. challenge.SanityCheck() — перевірка що всі рівні мають імена
   b. challenge.LoadCfg() — завантажити збережений прогрес
   c. Перевірити рівень: challenge.CheckCurrentLevel()
   d. Якщо пройдено: challenge.GoToNextLevel()
   e. Надрукувати рівень якщо:
      - last_level_printed != "yes"
      - АБО існує $HOME/.ta_print_again
   f. Встановити last_level_printed = "yes"
   g. Видалити $HOME/.ta_print_again
   h. Якщо існує $HOME/.ta_level_restart:
      - Виконати PreLevelCmd
      - Видалити $HOME/.ta_level_restart
```

### 12. Система шаблонів

**Файл:** `levels/template.go`, функція `Template()`

**Алгоритм:**

```
1. Прочитати YAML файл → templateData (map[string]interface{})
2. Прочитати Go template файл
3. Створити template.FuncMap з кастомними функціями:
   - generate_levels(name, levels, format) → "level1, level2, ..."
   - add(num, add) → num + add
4. template.New("template").Funcs(funcMap).Parse(templ)
5. t.Execute(os.Stdout, &templateData)
```

**Приклад використання:**

```bash
./term-adventure --generate-from-template challenge.tpl challenge.yaml
# або коротко:
./term-adventure -g challenge.tpl
# (автоматично шукає challenge.yaml)
```

### 13. Спеціальні файли-сентинелі

| Файл | Призначення |
|------|-------------|
| `$HOME/.ta_print_again` | Якщо існує — текст рівня буде надруковано повторно |
| `$HOME/.ta_level_restart` | Якщо існує — буде перезапущено PreLevelCmd поточного рівня |

Обидва видаляються після обробки.

### 14. Система таймерів (ліміт часу на рівень)

**Новий функціонал** — можна встановити обмеження часу на виконання кожного рівня окремо.

#### Як працює

```
1. При переході на новий рівень (GoToNextLevel):
   - challenge.SaveLevelStartTime() → записує поточний час (unix timestamp)
     у файл $HOME/.ta_level_start_time

2. При друці рівня (якщо last_level_printed != "yes"):
   - challenge.SaveLevelStartTime() → оновлює час старту

3. При кожному prompt_command у ta_bashrc:
   a. $TA_BIN --get-level-timelimit → отримує ліміт поточного рівня
   б. Якщо ліміт > 0:
      - $TA_BIN --check-level-time → перевіряє чи час вичерпано
      - Якщо так → виводить "⏰ Час на цей рівень вичерпано!"
      - $TA_BIN --print-level-timer → отримує залишок часу
      - Форматує у "Xm YYs" і додає у PS1
```

#### Структура Level (оновлена)

**Файл:** `levels/levels.go`

```go
type Level struct {
    Name              string
    PreLevelCmd       string `yaml:"precmd"`
    PostLevelCmd      string `yaml:"postcmd"`
    PostLevelPrintCmd string `yaml:"postprintcmd"`
    Text              string
    TestCmd           string   `yaml:"test"`
    NextLevels        []string `yaml:"next"`
    BackgroundJobs    bool     `yaml:"bgjobs"`
    TimeLimit         int      `yaml:"timelimit"`  // ← нове поле
}
```

#### Нові методи Challenge

| Метод | Опис | Повертає |
|-------|------|----------|
| `GetLevelTimeLimit()` | Ліміт часу поточного рівня | `int` (секунди), 0 = немає |
| `SaveLevelStartTime()` | Записує час старту у файл | — |
| `CheckLevelTimeExpired()` | Чи вичерпано час | `bool` |
| `GetLevelTimeRemaining()` | Залишок часу | `int` (секунди), -1 = немає ліміту |

#### Нові CLI прапори (main.go)

| Прапор | Опис | Вивід |
|--------|------|-------|
| `--get-level-timelimit` | Ліміт часу поточного рівня | `300` |
| `--check-level-time` | Чи час вичерпано | exit 0 = так, 1 = ні |
| `--print-level-timer` | Залишок часу | `245` (секунди) |

#### Формат промпту з таймером

Без таймеру:
```
14:32:07 [sample_challenge l00][user@host:/tmp]$ 
```

З таймером:
```
14:32:07 [4m 23s] [sample_challenge l00][user@host:/tmp]$ 
             ^^^^^^
          залишок часу
```

Якщо час вичерпано:
```
⏰ Час на цей рівень вичерпано!

14:32:07 [0m 00s] [sample_challenge l00][user@host:/tmp]$ 
```

#### Файл збереження часу

**Шлях:** `$HOME/.ta_level_start_time`

**Вміст:** unix timestamp (кількість секунд з 1970-01-01)

```
1712150400
```

Видаляється при завершенні сесії (`challenger.sh`).

---

## Збірка та запуск

### Залежності

- Go (без go.mod, legacy GOPATH-іморти `./levels`)
- `github.com/rakyll/globalconf`
- `gopkg.in/yaml.v2`

### Збірка

```bash
go build -ldflags "-X main.encryption_key=your_secret_key"
```

**Що робить `-ldflags "-X"`:**

```
-X main.encryption_key=value
```

Встановлює значення змінної `encryption_key` у пакеті `main` **на етапі компіляції**. Змінна оголошена у `main.go` рядок 13:

```go
var encryption_key string  // порожня за замовчуванням
```

### Запуск челенджу

```bash
export CHALLENGE_FILE=./my_challenge.ta
./challenger.sh
```

Або через `sample_challenge.sh`:

```bash
./sample_challenge.sh
```

### Корисні команди

| Команда | Опис |
|---------|------|
| `./term-adventure --print file.ta` | Надрукувати всі рівні у структурованому вигляді |
| `./term-adventure --enc file.ta` | Зашифрувати файл |
| `./term-adventure --dec file.enc` | Дешифрувати файл |
| `./term-adventure --print-level file.ta` | Назва поточного рівня |
| `./term-adventure --print-current-level file.ta` | Повний текст поточного рівня |
| `./term-adventure --print-identifier file.ta` | Ідентифікатор `[challenge level]` |
| `./term-adventure --print-challenge file.ta` | Назва челенджу |
| `./term-adventure --check-current-level file.ta` | Exit 0 = пройдено, 1 = ні |
| `./term-adventure --detect-level file.ta <hash> <homedir>` | Розшифрувати хеш → назву рівня |
| `./term-adventure -g template.tpl vars.yaml` | Генерація .ta з шаблону |

---

## Контроль прогресу студентів

### Механізми для викладача

| Механізм | Як використати |
|----------|---------------|
| **Промпт** `[challenge level]` | Видно поточний рівень прямо у терміналі |
| **`--print-level`** | Отримати назву рівня |
| **`--print-current-level`** | Побачити повний текст рівня |
| **`--check-current-level`** | Автоматична перевірка pass/fail (exit code) |
| **`$HOME/.tahistory`** | Перегляд усіх команд студента |
| **`--detect-level <hash> <homedir>`** | Розшифрувати хеш → назву рівня |
| **`touch $HOME/.ta_print_again`** | Примусово показати інструкцію знову |

### Історія команд

Файл `$HOME/.tahistory` містить до 500 останніх команд студента. Перевіряти треба **під час активної сесії** — після виходу файл видаляється.

### Конфігурація прогресу

Зберігається у `$HOME/.config/<challenge_name>/`. Містить MD5 хеш поточного рівня.

---

## Що передавати студенту

### Обов'язкові файли

| Файл | Призначення |
|------|-------------|
| `term-adventure` | Зібраний бінарник (НЕ вихідний код!) |
| `challenger.sh` | Скрипт запуску сесії |
| `ta_bashrc` | Кастомне середовище bash |
| `challenge.ta` або `.ta.enc` | Файл завдання |

### З шифруванням (рекомендовано)

```bash
# Викладач шифрує:
./term-adventure --enc my_challenge.ta > my_challenge.ta.enc

# Студенту передається архів:
student-challenge.zip
├── term-adventure      # зібраний з тим самим ключем!
├── challenger.sh
├── ta_bashrc
└── challenge.ta.enc
```

> **Важливо:** бінарник має бути зібраний з **тим самим ключем**, яким зашифровано `.ta.enc`. Інакше дешифрування не спрацює.

### Що НЕ передавати

- Вихідний код (`main.go`, `levels/`)
- Ключ шифрування
- Файли з відповідями

---

## Платформні обмеження

- **Linux/macOS only** — використовує `/dev/tty`, `stty -F`, `bash --rcfile`
- **WSL на Windows** — працює
- **Нативний Windows** — не працює

---

## Розробка

### Тестування

Мінімальне — один тест `print_test.go` перевіряє крайній випадок переповнення лічильника рядків при пропуску тексту:

```bash
go test
```

### Відома проблема (виправлено)

- `${TA_BIN=:...}` → `${TA_BIN:-...}` — невірний синтаксис parameter expansion давав `:path` замість `path`
- Відносні шляхи `CHALLENGE_FILE` → автоматичне перетворення на абсолютні
