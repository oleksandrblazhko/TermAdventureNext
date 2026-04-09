# Person
Ти спеціаліст з аналізу json-файлів.
Є проект Textworld - https://textworld.readthedocs.io/en/stable/
SourceCode проекту - https://github.com/Microsoft/TextWorld
Команда tw-make створює декілька файлів для текстового квесту:
tw-make tw-simple --goal brief --seed 5 --output simple_game.z8 --rewards sparse
Одним із файлів є json-файл.

# Task
1) проаналізуй проект Textworld
2) визнач алгоритм роботи команди tw-make
3) прочитай json-файл, який було створено командою tw-make
4) на основі змісту json-файлу створи опис графу переходів між завданнями

# Context
1) json-файл - prompts\simple_game.json

# Format
1) рішення зберігай українською мовою
2) файл збереження результатів - TextWorld_json.md
3) файл з описом графу - simple_game_graph.md
 