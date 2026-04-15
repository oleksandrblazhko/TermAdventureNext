# Person
Раніше ти виконав завдання - prompts\prompt_3_TextWord_Bash_Mapping.md
Було створено таблицю мапінгу між діями гравця та діями Bash-команд.
Мапінг є проміжною структурою, яка буде використовуватися для створення опису реальних Bash-команд,
які повинен виконувати гравець, намагаючись виконати дії у грі.

# Task
1) прочитай таблицю мапінгу
2) для кожного рядка мапінгу опиши реальні Bash-команди, які розбито на три частини:
- bash-команди для підготовки bash-середовища перед виконанням завдання гри
- перевірка правильності виконання bash-команд, які визначають правильне виконання завдання гри
- bash-команди для очистки bash-середовища після виконанням завдання гри

# Context
1) таблиця мапінгу - prompts/TW-simple_Mapping.md
2) для ідентифікації рядка перед описом Bash-команд вказуй елементи:
- id
- player_action
- bash_command
3) кожна частина починається з ключового слова:
- precmd: bash-команди для підготовки bash-середовища перед виконанням завдання
- test: перевірка правильності виконання bash-команд
- postcmd: bash-команди для очистки bash-середовища після виконанням завдання
4) для перевірка правильності виконання bash-команд необхідно використовувати утиліту test
5) приклади опису частин:
- приклад 1:
precmd: echo "Hi there!" > /tmp/ta_test
test: test $(pwd) = "$HOME"
postcmd: rm /tmp/ta_test
- приклад 2:
precmd: mkdir -p $HOME/.room
test: test "$(cat $HOME/.rooms/info.txt)" = "bedroom"
postcmd: rm -fR $HOME/.rooms
- приклад 3:
precmd: mkdir -p $HOME/refrigerator && touch $HOME/refrigerator/.closed
test: test ! -f $HOME/refrigerator/.closed
postcmd: touch $HOME/refrigerator/.open

# Format
1) формат представлення - yaml
2) опис збережи у файлі prompts/Lab1_Bash_Scripts.yaml
