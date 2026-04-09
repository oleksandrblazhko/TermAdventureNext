#!/bin/bash
# game_state.sh - Helper для управління станом гри
# Використовується конвертованими квестами TextWorld → TermAdventure
#
# Використання:
#   game_state.sh init <challenge_name>   - Ініціалізація стану
#   game_state.sh check <type> <id>       - Перевірка умови
#   game_state.sh has <item_id>           - Перевірка наявності предмета
#   game_state.sh at <room_id>            - Перевірка поточної кімнати
#   game_state.sh update <type> <id>      - Оновлення стану
#   game_state.sh win                     - Перевірка умови перемоги

# Файл стану (зберігається у тимчасовій директорії)
STATE_DIR="/tmp/termadventure"
STATE_FILE=""

# ===========================
# Ініціалізація
# ===========================

init_state() {
    local challenge_name="${1:-default}"
    STATE_DIR="/tmp/termadventure_${challenge_name}"
    STATE_FILE="${STATE_DIR}/game_state.json"
    
    # Створюємо директорію
    mkdir -p "$STATE_DIR"
    
    # Ініціалізуємо порожній стан
    if [ ! -f "$STATE_FILE" ]; then
        cat > "$STATE_FILE" << 'EOF'
{
    "containers": {},
    "doors": {},
    "items": {},
    "inventory": [],
    "player_room": "",
    "won": false
}
EOF
        echo "Стан гри ініціалізовано: $STATE_FILE"
    fi
}

# Завантажуємо STATE_FILE при кожному виклику (якщо не init)
if [ "$1" != "init" ]; then
    # Шукаємо файл стану
    for dir in /tmp/termadventure_*; do
        if [ -d "$dir" ] && [ -f "$dir/game_state.json" ]; then
            STATE_DIR="$dir"
            STATE_FILE="$dir/game_state.json"
            break
        fi
    done
    
    if [ -z "$STATE_FILE" ] || [ ! -f "$STATE_FILE" ]; then
        echo "Помилка: стан гри не ініціалізовано. Виконайте: game_state.sh init <name>" >&2
        exit 1
    fi
fi

# ===========================
# Перевірка умов
# ===========================

check_condition() {
    local type="$1"
    local id="$2"
    
    case "$type" in
        open)
            # Перевірка чи контейнер/двері відчинені
            check_open "$id"
            ;;
        closed)
            # Перевірка чи контейнер/двері зачинені
            check_closed "$id"
            ;;
        unlocked)
            # Перевірка чи двері відімкнені
            check_unlocked "$id"
            ;;
        locked)
            # Перевірка чи двері замкнені
            check_locked "$id"
            ;;
        in)
            # Перевірка чи предмет у контейнері
            # Використання: game_state.sh check in <item_id>
            check_in_container "$id"
            ;;
        on)
            # Перевірка чи предмет на поверхні
            # Використання: game_state.sh check on <item_id> <supporter_id>
            check_on_supporter "$id" "$3"
            ;;
        win)
            # Перевірка умови перемоги
            check_win
            ;;
        *)
            echo "Невідомий тип перевірки: $type" >&2
            return 1
            ;;
    esac
}

check_open() {
    local id="$1"
    # Перевірка чи контейнер відчинений (не має прапорця "closed")
    if [ -f "$STATE_DIR/container_$id" ] && [ "$(cat "$STATE_DIR/container_$id")" = "open" ]; then
        return 0
    fi
    # Перевірка чи двері відчинені
    if [ -f "$STATE_DIR/door_$id" ] && [ "$(cat "$STATE_DIR/door_$id")" = "open" ]; then
        return 0
    fi
    return 1
}

check_closed() {
    local id="$1"
    # Перевірка чи контейнер зачинений
    if [ -f "$STATE_DIR/container_$id" ] && [ "$(cat "$STATE_DIR/container_$id")" = "closed" ]; then
        return 0
    fi
    # Перевірка чи двері зачинені
    if [ -f "$STATE_DIR/door_$id" ] && [ "$(cat "$STATE_DIR/door_$id")" = "closed" ]; then
        return 0
    fi
    return 1
}

check_unlocked() {
    local id="$1"
    # Перевірка чи двері відімкнені (не мають прапорця "locked")
    if [ -f "$STATE_DIR/door_$id" ]; then
        local state=$(cat "$STATE_DIR/door_$id")
        if [ "$state" = "unlocked" ] || [ "$state" = "open" ]; then
            return 0
        fi
    fi
    return 1
}

check_locked() {
    local id="$1"
    # Перевірка чи двері замкнені
    if [ -f "$STATE_DIR/door_$id" ] && [ "$(cat "$STATE_DIR/door_$id")" = "locked" ]; then
        return 0
    fi
    return 1
}

check_in_container() {
    local item_id="$1"
    # Перевірка чи предмет НЕ у контейнері (гравець його взяв)
    # Якщо файл container_<container_id> містить цей предмет - він ще всередині
    for container_file in "$STATE_DIR"/container_contents_*; do
        if [ -f "$container_file" ] && grep -q "$item_id" "$container_file" 2>/dev/null; then
            return 1  # Предмет ще в контейнері
        fi
    done
    return 0  # Предмет не знайдено в жодному контейнері (гравець взяв)
}

check_on_supporter() {
    local item_id="$1"
    local supporter_id="$2"
    # Перевірка чи предмет на поверхні
    if [ -f "$STATE_DIR/supporter_${supporter_id}" ] && grep -q "$item_id" "$STATE_DIR/supporter_${supporter_id}" 2>/dev/null; then
        return 0
    fi
    return 1
}

check_win() {
    # Перевірка умови перемоги
    if [ -f "$STATE_DIR/won" ] && [ "$(cat "$STATE_DIR/won")" = "true" ]; then
        return 0
    fi
    return 1
}

# ===========================
# Перевірка інвентарю
# ===========================

has_item() {
    local item_id="$1"
    # Перевірка чи предмет в інвентарі гравця
    if [ -f "$STATE_DIR/inventory" ] && grep -q "$item_id" "$STATE_DIR/inventory" 2>/dev/null; then
        return 0
    fi
    return 1
}

# ===========================
# Перевірка кімнати
# ===========================

check_player_room() {
    local room_id="$1"
    # Перевірка поточної кімнати гравця
    if [ -f "$STATE_DIR/player_room" ] && [ "$(cat "$STATE_DIR/player_room")" = "$room_id" ]; then
        return 0
    fi
    return 1
}

# ===========================
# Оновлення стану
# ===========================

update_state() {
    local type="$1"
    local id="$2"
    
    case "$type" in
        open)
            update_open "$id"
            ;;
        closed)
            update_closed "$id"
            ;;
        unlocked)
            update_unlocked "$id"
            ;;
        locked)
            update_locked "$id"
            ;;
        *)
            echo "Невідомий тип оновлення: $type" >&2
            return 1
            ;;
    esac
}

update_open() {
    local id="$1"
    # Відчиняємо контейнер
    if [ -f "$STATE_DIR/container_$id" ] || [ -d "$STATE_DIR" ]; then
        echo "open" > "$STATE_DIR/container_$id"
        echo "container_$id: open" >&2
    fi
    # Або двері
    echo "open" > "$STATE_DIR/door_$id" 2>/dev/null
    echo "door_$id: open" >&2
}

update_closed() {
    local id="$1"
    # Зачиняємо контейнер
    echo "closed" > "$STATE_DIR/container_$id"
    echo "container_$id: closed" >&2
}

update_unlocked() {
    local id="$1"
    # Відмикаємо двері
    echo "unlocked" > "$STATE_DIR/door_$id"
    echo "door_$id: unlocked" >&2
}

update_locked() {
    local id="$1"
    # Замикаємо двері
    echo "locked" > "$STATE_DIR/door_$id"
    echo "door_$id: locked" >&2
}

# ===========================
# Дії гравця
# ===========================

take_item() {
    local item_id="$1"
    local source="$2"  # container або supporter ID
    
    # Додаємо до інвентарю
    echo "$item_id" >> "$STATE_DIR/inventory"
    
    # Видаляємо з джерела
    if [[ "$source" == container_* ]]; then
        local container_id="${source#container_}"
        if [ -f "$STATE_DIR/container_contents_${container_id}" ]; then
            sed -i "/$item_id/d" "$STATE_DIR/container_contents_${container_id}"
        fi
    elif [[ "$source" == supporter_* ]]; then
        local supporter_id="${source#supporter_}"
        if [ -f "$STATE_DIR/supporter_${supporter_id}" ]; then
            sed -i "/$item_id/d" "$STATE_DIR/supporter_${supporter_id}"
        fi
    fi
    
    echo "Предмет $item_id взято" >&2
}

put_item() {
    local item_id="$1"
    local target="$2"  # supporter ID
    
    # Видаляємо з інвентарю
    if [ -f "$STATE_DIR/inventory" ]; then
        sed -i "/$item_id/d" "$STATE_DIR/inventory"
    fi
    
    # Додаємо на поверхню
    echo "$item_id" >> "$STATE_DIR/supporter_${target}"
    
    echo "Предмет $item_id покладено на $target" >&2
}

insert_item() {
    local item_id="$1"
    local target="$2"  # container ID
    
    # Видаляємо з інвентарю
    if [ -f "$STATE_DIR/inventory" ]; then
        sed -i "/$item_id/d" "$STATE_DIR/inventory"
    fi
    
    # Додаємо в контейнер
    echo "$item_id" >> "$STATE_DIR/container_contents_${target}"
    
    echo "Предмет $item_id покладено в $target" >&2
}

move_to() {
    local room_id="$1"
    
    echo "$room_id" > "$STATE_DIR/player_room"
    echo "Гравець перейшов до $room_id" >&2
}

set_win() {
    echo "true" > "$STATE_DIR/won"
    echo "Квест пройдено!" >&2
}

# ===========================
# Головний перемикач
# ===========================

case "$1" in
    init)
        init_state "$2"
        ;;
    check)
        check_condition "$2" "$3" "$4"
        ;;
    has)
        has_item "$2"
        ;;
    at)
        check_player_room "$2"
        ;;
    update)
        update_state "$2" "$3"
        ;;
    take)
        take_item "$2" "$3"
        ;;
    put)
        put_item "$2" "$3"
        ;;
    insert)
        insert_item "$2" "$3"
        ;;
    move)
        move_to "$2"
        ;;
    win)
        set_win
        ;;
    state)
        # Показати поточний стан (для дебагу)
        if [ -f "$STATE_FILE" ]; then
            cat "$STATE_FILE"
        else
            echo "Стан не знайдено. Директорія: $STATE_DIR" >&2
            ls -la "$STATE_DIR"/ 2>/dev/null
        fi
        ;;
    reset)
        # Скинути стан
        rm -rf "$STATE_DIR"
        echo "Стан гри скинуто"
        ;;
    *)
        echo "Використання:" >&2
        echo "  game_state.sh init <challenge_name>" >&2
        echo "  game_state.sh check <type> <id>" >&2
        echo "  game_state.sh has <item_id>" >&2
        echo "  game_state.sh at <room_id>" >&2
        echo "  game_state.sh update <type> <id>" >&2
        echo "  game_state.sh take <item_id> <source>" >&2
        echo "  game_state.sh put <item_id> <target>" >&2
        echo "  game_state.sh insert <item_id> <target>" >&2
        echo "  game_state.sh move <room_id>" >&2
        echo "  game_state.sh win" >&2
        echo "  game_state.sh state" >&2
        echo "  game_state.sh reset" >&2
        exit 1
        ;;
esac
