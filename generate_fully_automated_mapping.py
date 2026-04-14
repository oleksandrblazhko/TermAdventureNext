
import yaml
import re

# File paths
MARKDOWN_TABLE_PATH = "tw2ta/tw-simple_Table_Map_Lab1.md"
ORIGINAL_YAML_PATH = "tw2ta/tw-simple_mapping.yaml"
GENERATED_YAML_PATH = "tw2ta/tw-simple_mapping_generated.yaml"

def parse_markdown_table(file_path):
    """Parses the Markdown table into a list of dictionaries."""
    with open(file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    header = []
    data = []
    
    table_started = False
    for line in lines:
        stripped_line = line.strip()
        if not stripped_line:
            continue

        if stripped_line.startswith('|') and '---' in stripped_line:
            # This is the separator line between header and data
            table_started = True
            continue

        if stripped_line.startswith('|') and table_started:
            # This is a data row
            row_values = [v.strip() for v in stripped_line.strip('|').split('|')]
            if len(row_values) == len(header):
                data.append(dict(zip(header, row_values)))
        elif stripped_line.startswith('|') and not table_started:
            # This is the header row
            header = [h.strip() for h in stripped_line.strip('|').split('|')]
            
    return data

def get_static_yaml_sections(original_yaml_path):
    """Reads specific static sections from the original YAML file."""
    with open(original_yaml_path, 'r', encoding='utf-8') as f:
        full_yaml = yaml.safe_load(f)
    
    return {
        'game_dir': full_yaml.get('game_dir'),
        'inventory_dir': full_yaml.get('inventory_dir'),
        'world': full_yaml.get('world'),
        'template_vars': full_yaml.get('template_vars'),
    }

def generate_commands_section(parsed_data):
    """Generates the 'commands' section of the YAML."""
    commands = {}
    for row in parsed_data:
        # Assuming the 'назва Bash-команди' is the first bash command in the string
        bash_command_full = row.get('назва Bash-команди', '').replace('`', '').strip()
        bash_command_name = bash_command_full.split(' ')[0]

        if not bash_command_name or bash_command_name in ['/', '—', 'cat', 'less']: # Ignore placeholders or commands handled differently
            continue

        # Avoid duplicates, prioritize the first entry or combine descriptions
        if bash_command_name not in commands:
            commands[bash_command_name] = {
                'full': bash_command_full,
                'description': row.get('пояснення щодо відповідності', ''),
                'textworld_action': row.get('назва операції гравця', ''),
                'game_role': row.get('пояснення щодо відповідності', '') # Reusing description for now
            }
            # Handle specific cases for commands that appear multiple times with different examples
            if bash_command_name == 'mv':
                if 'Взяти предмет' in row.get('пояснення щодо відповідності', ''):
                    commands[bash_command_name]['game_role'] = "Взяти/покласти/перекласти предмет"
                elif 'Кинути предмет' in row.get('пояснення щодо відповідності', ''):
                     commands[bash_command_name]['game_role'] = "Взяти/покласти/перекласти предмет"
            elif bash_command_name == 'rm':
                if 'Відчинення контейнерів' in row.get('пояснення щодо відповідності', ''):
                    commands[bash_command_name]['game_role'] = "Відчинити контейнер, знищити предмет"
                elif 'видалення файлу-їжі' in row.get('пояснення щодо відповідності', ''):
                    commands[bash_command_name]['game_role'] = "Відчинити контейнер, знищити предмет"
            elif bash_command_name == 'touch':
                if 'створення "файлу-замка"' in row.get('пояснення щодо відповідності', ''):
                    commands[bash_command_name]['game_role'] = "Змінити стан, відзначити перемогу"
                elif 'Приготування' in row.get('пояснення щодо відповідності', ''):
                    commands[bash_command_name]['game_role'] = "Змінити стан, відзначити перемогу"


        # Special handling for cat/less as they are combined in the source YAML and have different examples
        if 'cat' in bash_command_name and 'cat' not in commands:
            commands['cat'] = {
                'full': "cat файл",
                'description': "Вивести вміст файлу на екран",
                'textworld_action': "examine <item>",
                'game_role': "Що написано?"
            }
        if 'less' in bash_command_name and 'less' not in commands:
            commands['less'] = {
                'full': "less файл",
                'description': "Посторінковий перегляд файлу (q — вихід)",
                'textworld_action': "read <item>",
                'game_role': "Детальне читання"
            }

    return commands

def generate_action_templates_section(parsed_data):
    """Generates the 'action_templates' section of the YAML."""
    action_templates = {}
    
    # Global variables for templates (matching tw-simple_mapping.yaml)
    game_dir = "{game_dir}"
    inventory_log = "{inventory_log}"
    doors_log = "{doors_log}"
    movement_log = "{movement_log}"
    current_room_path = "{current_room}"
    rooms_dir = "{rooms_dir}"
    win_condition_path = "{win_condition}"
    key = "{key}"
    element = "{element}" # Generic element for examine

    # Pre-define common direction templates to avoid duplication
    for direction in ['east', 'west', 'north', 'south']:
        template_key = f"go/{direction}"
        action_templates[template_key] = {
            'description': f"Йти на {direction}",
            'textworld_example': f"go {direction}",
            'player_command': f'echo "{{room}}" > {current_room_path}',
            'test': f'test "$(cat {current_room_path})" = "{{room}}"',
            'precmd': f"mkdir -p {rooms_dir}",
            'postcmd': f'echo "Moved to {{room_name}} at $(date)" >> {movement_log}'
        }

    for row in parsed_data:
        player_action_raw = row.get('назва операції гравця', '').strip()
        bash_command_raw = row.get('назва Bash-команди', '').strip()
        description = row.get('опис операції', '').strip()
        if not description: # Fallback to explanation if 'опис операції' is empty
             description = row.get('пояснення щодо відповідності', '').strip()
        example = row.get('приклад операції', '').replace('`', '').strip()

        # Normalize player action for easier matching
        player_action = player_action_raw.lower()
        child_entity = row.get('child-сутність', '').lower()
        parent_entity = row.get('parent-сутність', '').lower()

        # --- General Interactions ---
        if player_action == 'look' and 'look' not in action_templates:
             action_templates['look'] = {
                'description': 'Оглянутись у кімнаті',
                'textworld_example': 'look',
                'player_command': f'ls -la {game_dir}/{{room}}/',
                'test': 'true'
             }
        elif player_action == 'inventory' and 'inventory' not in action_templates:
            action_templates['inventory'] = {
                'description': 'Переглянути інвентар',
                'textworld_example': 'inventory',
                'player_command': 'ls ~/',
                'test': 'true'
            }
        elif player_action.startswith('examine') and 'examine' not in action_templates:
            action_templates['examine'] = {
                'description': description, # Use derived description
                'textworld_example': example,
                'player_command': f"cat {game_dir}/{{room}}/{element} || cat ~/{element}",
                'test': 'true'
            }
            
        # --- Door Interactions ---
        # "open/close/unlock/lock" with "двері"
        elif 'двері' in child_entity or 'door' in example:
            if player_action == 'open' and 'open/d' not in action_templates:
                action_templates['open/d'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f'echo "open" > {game_dir}/{{door}}.state',
                    'test': f'test "$(cat {game_dir}/{{door}}.state)" = "open"',
                    'precmd': f'echo "closed" > {game_dir}/{{door}}.state',
                    'postcmd': f'echo "{{door}}: open" >> {doors_log}'
                }
            elif player_action == 'close' and 'close/d' not in action_templates:
                action_templates['close/d'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f'echo "closed" > {game_dir}/{{door}}.state',
                    'test': f'test "$(cat {game_dir}/{{door}}.state)" = "closed"'
                }
            elif player_action.startswith('unlock') and 'unlock/d' not in action_templates:
                action_templates['unlock/d'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f'echo "closed" > {game_dir}/{{door}}.state', # This command changes state to closed, implying unlock was successful
                    'test': f'test "$(cat {game_dir}/{{door}}.state)" = "closed" && test -f ~/{key}', # Needs key in inventory
                    'precmd': f'echo "locked" > {game_dir}/{{door}}.state', # Initial state for test
                    'postcmd': f"touch {game_dir}/{{door}}.unlocked"
                }
            elif player_action.startswith('lock') and 'lock/d' not in action_templates:
                action_templates['lock/d'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f'echo "locked" > {game_dir}/{{door}}.state',
                    'test': f'test "$(cat {game_dir}/{{door}}.state)" = "locked" && test -f ~/{key}' # Needs key in inventory
                }

        # --- Container Interactions ---
        # "open/close/take from/insert into" with "контейнер"
        elif 'контейнер' in child_entity or 'container' in example:
            if player_action == 'open' and 'open/c' not in action_templates:
                action_templates['open/c'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"rm {game_dir}/{{room}}/{{container}}/.closed",
                    'test': f"test ! -f {game_dir}/{{room}}/{{container}}/.closed",
                    'precmd': f"mkdir -p {game_dir}/{{room}}/{{container}} && touch {game_dir}/{{room}}/{{container}}/.closed",
                    'postcmd': f"touch {game_dir}/{{room}}/{{container}}/.open"
                }
            elif player_action == 'close' and 'close/c' not in action_templates:
                action_templates['close/c'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"touch {game_dir}/{{room}}/{{container}}/.closed",
                    'test': f"test -f {game_dir}/{{room}}/{{container}}/.closed",
                    'postcmd': f"rm -f {game_dir}/{{room}}/{{container}}/.open"
                }
            elif player_action.startswith('take') and 'take/c' not in action_templates:
                action_templates['take/c'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv {game_dir}/{{room}}/{{container}}/{{item}} ~/",
                    'test': f"test -f ~/{{item}}",
                    'precmd': f"mkdir -p {game_dir}/{{room}}/{{container}}",
                    'postcmd': f"echo '{{item}} taken' >> {inventory_log}"
                }
            elif player_action.startswith('insert') and 'insert' not in action_templates:
                action_templates['insert'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv ~/{{item}} {game_dir}/{{room}}/{{container}}/",
                    'test': f"test -f {game_dir}/{{room}}/{{container}}/{{item}} && test ! -f ~/{{item}}",
                    'postcmd': f"rm ~/{{item}}"
                }
        
        # --- Surface Interactions ---
        # "put on/take from" with "поверхня"
        elif 'поверхня' in child_entity or 'surface' in example:
            if player_action.startswith('put') and 'put' not in action_templates:
                action_templates['put'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv ~/{{item}} {game_dir}/{{room}}/{{surface}}/",
                    'test': f"test -f {game_dir}/{{room}}/{{surface}}/{{item}}",
                    'precmd': f"mkdir -p {game_dir}/{{room}}/{{surface}}",
                    'postcmd': f"rm ~/{{item}}"
                }
            elif player_action.startswith('take') and 'take/s' not in action_templates:
                 action_templates['take/s'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv {game_dir}/{{room}}/{{surface}}/{{item}} ~/",
                    'test': f"test -f ~/{{item}}",
                    'precmd': f"mkdir -p {game_dir}/{{room}}/{{surface}} && touch {game_dir}/{{room}}/{{surface}}/{{item}}",
                    'postcmd': f"rm {game_dir}/{{room}}/{{surface}}/{{item}}"
                 }

        # --- Object Interactions (loose, generic) ---
        # "take/drop/read" with "об'єкт"
        elif "об'єкт" in child_entity or 'object' in example:
            if player_action == 'take' and 'take/o' not in action_templates:
                action_templates['take/o'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv {game_dir}/{{room}}/{{item}} ~/",
                    'test': f"test -f ~/{{item}}",
                    'precmd': f"mkdir -p {game_dir}/{{room}}",
                    'postcmd': f"echo '{{item}} taken' >> {inventory_log}"
                }
            elif player_action == 'drop' and 'drop/o' not in action_templates:
                action_templates['drop/o'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"mv ~/{{item}} {game_dir}/{{room}}/",
                    'test': f"test -f {game_dir}/{{room}}/{{item}} && test ! -f ~/{{item}}",
                    'postcmd': f"echo '{{item}} dropped' >> {movement_log}"
                }
            elif player_action == 'read' and 'read/o' not in action_templates: # read note/object
                action_templates['read/o'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"cat {game_dir}/{{room}}/{{item}}", # Assuming item is in room
                    'test': 'true'
                }
        # --- Food-Specific Actions ---
        # "cook/eat" with "їжа"
        elif 'їжа' in child_entity or 'food' in example:
            if player_action == 'cook' and 'cook' not in action_templates:
                action_templates['cook'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"touch {game_dir}/{{room}}/{{surface}}/cooked_{{item}}",
                    'test': f"test -f {game_dir}/{{room}}/{{surface}}/cooked_{{item}}"
                }
            elif player_action == 'eat' and 'eat' not in action_templates:
                action_templates['eat'] = {
                    'description': description,
                    'textworld_example': example,
                    'player_command': f"rm ~/{{item}}",
                    'test': f"test ! -f ~/{{item}}"
                }
    
    return action_templates

def generate_full_yaml():
    """Generates the complete YAML structure."""
    parsed_data = parse_markdown_table(MARKDOWN_TABLE_PATH)
    static_sections = get_static_yaml_sections(ORIGINAL_YAML_PATH)

    generated_commands = generate_commands_section(parsed_data)
    generated_action_templates = generate_action_templates_section(parsed_data)

    full_output = {}
    
    # Add static global settings
    full_output['game_dir'] = static_sections['game_dir']
    full_output['inventory_dir'] = static_sections['inventory_dir']

    # Add generated commands
    full_output['commands'] = generated_commands
    
    # Add static world data
    full_output['world'] = static_sections['world']

    # Add generated action templates
    full_output['action_templates'] = generated_action_templates

    # Add static template variables
    full_output['template_vars'] = static_sections['template_vars']

    return full_output

if __name__ == "__main__":
    generated_yaml_data = generate_full_yaml()
    with open(GENERATED_YAML_PATH, 'w', encoding='utf-8') as f:
        yaml.dump(generated_yaml_data, f, allow_unicode=True, indent=2, sort_keys=False)
    print(f"Generated YAML saved to {GENERATED_YAML_PATH}")