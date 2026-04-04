# Покрокові інструкції щодо створення опису квесту

## Крок 1: Визначення навчальних цілей

Спочатку склади список команд, які студент має опанувати:

```
Приклад набору команд:
1. cd - навігація по директоріях
2. ls, ls -l - перегляд вмісту
3. mkdir - створення директорій
4. cp - копіювання файлів
5. mv - переміщення файлів
6. rm - видалення файлів
7. cat - перегляд вмісту файлів
8. grep - пошук тексту
9. find - пошук файлів
10. chmod - зміна прав доступу
```

## Крок 2: Розробка сценарію квесту

Створи послідовну історію, де кожен рівень вимагає використання конкретної команди:

```
Сценарій: "Подорож системного адміністратора"

Рівень 1: cd + ls
  Завдання: Знайти прихований файл у /tmp
  
Рівень 2: mkdir + cd
  Завдання: Створити робочу директорію
  
Рівень 3: cp + mv
  Завдання: Скопіювати конфігураційний файл
  
Рівень 4: grep + cat
  Завдання: Знайти пароль у файлі
  
Рівень 5: chmod
  Завдання: Зробити скрипт виконуваним
  
Фінал: Виконати скрипт
```

## Крок 3: Створення файлу `.ta` вручну

Створи файл `my_quest.ta` за таким шаблоном:

```bash
touch my_quest.ta
```

Вміст файлу:

```yaml
name: start
test: test -f /tmp/secret.txt && grep "password" /tmp/secret.txt > /dev/null
timelimit: 300
score: 10
next: [level2]

# Рівень 1: Пошук секрету

Вітаю, майбутній адміністраторе! 🎯

Твоє перше завдання: знайти файл `/tmp/secret.txt` 
і перевірити, що він містить слово **password**.

Для цього тобі знадобляться команди:
- **cat** для перегляду вмісту
- **grep** для пошуку тексту

Підказка: спочатку перейди у директорію `/tmp`

--------------------
```

## Крок 4: Визначення тестів для кожного рівня

Кожен рівень має поле `test:` — це bash-команда, яка перевіряє виконання:

```yaml
# Приклади test команд:

# Перевірка що студент у правильній директорії
test: test "$(pwd)" = "/home/user/workspace"

# Перевірка що файл існує
test: test -f /tmp/report.txt

# Перевірка вмісту файлу
test: grep -q "success" /tmp/result.txt

# Перевірка прав доступу
test: test -x /tmp/my_script.sh

# Комбінована перевірка
test: test -d /tmp/project && test -f /tmp/project/main.py

# Завжди true (для інформаційних рівнів)
test: true
```

## Крок 5: Додавання pre/post команд

Використовуй `precmd` та `postcmd` для підготовки середовища:

```yaml
name: level2
precmd: mkdir -p /tmp/project && echo "print('Hello')" > /tmp/project/main.py
postcmd: rm -rf /tmp/project
test: test -f /tmp/project/main.py
timelimit: 180
score: 15
next: [level3]

# Рівень 2: Робота з файлами

Чудово! Тепер створи директорію `/tmp/project`
і перевір, що в ному є файл `main.py`.

Команди:
- **mkdir** для створення директорії
- **ls** для перегляду вмісту

--------------------
```

## Крок 6: Створення через шаблон (для великих квестів)

Якщо квест великий, використовуй шаблонізатор:

### Файл `quest.tpl`:

```yaml
{{range $index, $cmd := .commands}}
name: level_{{$cmd.Name}}
precmd: {{$cmd.PrepCmd}}
test: {{$cmd.TestCmd}}
timelimit: {{$cmd.TimeLimit}}
score: {{$cmd.Score}}
{{if $cmd.Next}}next: [{{$cmd.Next}}]{{end}}

# Рівень {{$cmd.Number}}: {{$cmd.Name}}

{{$cmd.Description}}

Команди для використання:
{{$cmd.CommandsText}}

--------------------
{{end}}
```

### Файл `quest.yaml`:

```yaml
commands:
  - Name: navigation
    Number: 1
    PrepCmd: "mkdir -p /tmp/nav_test"
    TestCmd: "test \"$(pwd)\" = \"/tmp/nav_test\""
    TimeLimit: 300
    Score: 10
    Next: file_operations
    Description: |
      Твоє завдання: створи директорію та перейди в неї.
    CommandsText: "- **cd** - зміна директорії\n- **pwd** - поточна директорія"
    
  - Name: file_operations
    Number: 2
    PrepCmd: "echo 'test content' > /tmp/nav_test/file.txt"
    TestCmd: "test -f /tmp/nav_test/file.txt"
    TimeLimit: 240
    Score: 15
    Next: finish
    Description: |
      Тепер створи файл у поточній директорії.
    CommandsText: "- **touch** - створення файлів\n- **ls** - перегляд"
```

### Генерація:

```bash
./term-adventure -g quest.tpl quest.yaml > my_quest.ta
```

## Крок 7: Тестування квесту

```bash
# 1. Перегляд структури
./term-adventure --print my_quest.ta

# 2. Запуск сесії
export CHALLENGE_FILE=./my_quest.ta
./challenger.sh

# 3. Перевірка працездатності кожного рівня
./term-adventure --check-current-level my_quest.ta
echo $?  # 0 = пройдено, 1 = ні
```

## Крок 8: Фіналізація

Після тестування:

```bash
# Шифрування (якщо потрібно)
./term-adventure --enc my_quest.ta > my_quest.ta.enc

# Створення архіву для студентів
zip student-quest.zip term-adventure challenger.sh ta_bashrc my_quest.ta.enc

# Інструкція для студента
cat > README_STUDENT.txt << 'EOF'
Інструкція:
1. Розпакуй архів
2. Запусти: ./challenger.sh
3. Виконуй завдання рівнів
4. Для виходу: exit
EOF
```

---

## Повний приклад: "Файлова пригода"

Ось готовий шаблон для копіювання:

```yaml
name: intro
test: test "$(pwd)" = "/tmp"
timelimit: 300
score: 10
next: [create]

# Вступ: Навігація

Вітаю у квесті **"Файлова пригода"**! 🎮

Твої цілі:
1. Перейти у директорію `/tmp`
2. Створити робочу директорію
3. Створити та редагувати файли
4. Знайти приховані дані

Команди: **cd**, **pwd**, **ls**

Завдання: перейди у `/tmp`

--------------------

name: create
precmd: rm -rf /tmp/quest_work 2>/dev/null
test: test -d /tmp/quest_work && test -f /tmp/quest_work/notes.txt
timelimit: 240
score: 20
next: [search]

# Рівень 2: Створення

Чудово! Тепер створи:
1. Директорію `/tmp/quest_work`
2. Файл `notes.txt` всередині

Команди: **mkdir**, **touch**, **echo**

--------------------

name: search
test: grep -q "SECRET" /tmp/quest_work/notes.txt
timelimit: 300
score: 30
next: [finish]

# Рівень 3: Пошук

Додай у файл `notes.txt` рядок,
який містить слово **SECRET**.

Команди: **echo**, **>>**, **grep**

--------------------

name: finish
test: false
score: 40

# Фініш! 🏆

Вітаю! Ти пройшов усі рівні!

**Загальна кількість балів: 100**
```

---

## Корисні поради

### 1. Структура рівнів

Кожен рівень має містити:
- **Унікальну назву** (`name`)
- **Команду тесту** (`test`)
- **Наступні рівні** (`next`) — підтримує гілкування
- **Текст завдання** у Markdown форматі
- **Ліміт часу** (`timelimit`) у секундах
- **Кількість балів** (`score`)

### 2. Система балів

Бали нараховуються автоматично при проходженні рівня:
- ✅ **Вчасно** → повні бали
- ⏰ **Час вичерпано** → 50% штраф

Промпт відображає загальний рахунок:
```
[45 pts] [4m 23s] [quest l00][user:/tmp]$ 
```

### 3. Гілкування квесту

Використовуй `next: [level_a, level_b]` для створення альтернативних шляхів:

```yaml
name: decision_point
test: test -f /tmp/choice.txt
next: [path_easy, path_hard]
```

Програма випадково обирає один з наступних рівнів.

### 4. Підготовка середовища

Використовуй `precmd` для створення файлів/директорій перед рівнем:

```yaml
precmd: |
  mkdir -p /tmp/workspace
  echo "config_data" > /tmp/workspace/config.ini
  chmod 600 /tmp/workspace/config.ini
```

### 5. Очищення після рівня

Використовуй `postcmd` для видалення тимчасових файлів:

```yaml
postcmd: rm -rf /tmp/workspace
```

---

Ця інструкція покриває весь процес від планування до фінального продукту.
