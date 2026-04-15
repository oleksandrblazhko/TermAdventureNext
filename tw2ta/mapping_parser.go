package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// ActionMapping - правила конвертації однієї дії TextWorld → Bash
// Ця структура використовується внутрішньою логікою конвертера.
type ActionMapping struct {
	Name           string // Назва дії (open/c, take/c, тощо) - в новому форматі це player_action
	TextWorldCmd   string // Приклад команди TextWorld - в новому форматі не прямо мапиться
	Description    string // Опис українською - в новому форматі не прямо мапиться
	PlayerCommand  string // Команда для гравця - в новому форматі це bash_command
	Test           string // test (перевірка)
	PreCmd         string // precmd (підготовка)
	PostCmd        string // postcmd (фіксація)
}

// OldActionTemplate - сирий шаблон зі старого YAML
type OldActionTemplate struct {
	Description   string `yaml:"description"`
	TextWorldExample string `yaml:"textworld_example"`
	PlayerCommand string `yaml:"player_command"`
	Test          string `yaml:"test"`
	PreCmd        string `yaml:"precmd"`
	PostCmd       string `yaml:"postcmd"`
}

// NewYAMLAction - структура для нового формату YAML-мапінгу
type NewYAMLAction struct {
	ID           string `yaml:"id"`
	PlayerAction string `yaml:"player_action"`
	BashCommand  string `yaml:"bash_command"`
	PreCmd       string `yaml:"precmd"`
	Test         string `yaml:"test"`
	PostCmd      string `yaml:"postcmd"`
}

// NewYAMLRoot - коренева структура для нового формату YAML-мапінгу
type NewYAMLRoot struct {
	Actions      []NewYAMLAction `yaml:"actions"`
	TemplateVars TemplateVars    `yaml:"template_vars"` // Додано для глобальних змінних
}

// TemplateVars - глобальні змінні для підстановки (зберігаємо, якщо вони все ще потрібні в майбутньому)
// Для нового формату ці змінні не читаються безпосередньо з YAML.
type TemplateVars struct {
	GameDir       string `yaml:"game_dir"`
	InventoryLog  string `yaml:"inventory_log"`
	DoorsLog      string `yaml:"doors_log"`
	MovementLog   string `yaml:"movement_log"`
	CurrentRoom   string `yaml:"current_room"`
	WinCondition  string `yaml:"win_condition"`
	RoomsDir      string `yaml:"rooms_dir"`
}

// OldLab1Mapping - коренева структура СТАРОГО YAML-файлу (зберігаємо для можливої сумісності)
type OldLab1Mapping struct {
	Actions         map[string]OldActionTemplate `yaml:"action_templates"`
	TemplateVars    TemplateVars              `yaml:"template_vars"`
}

// BashMapping - всі правила мапінгу
type BashMapping struct {
	Actions map[string]*ActionMapping

	// Глобальні налаштування (можливо, будуть заповнюватись з інших джерел або матимуть значення за замовчуванням)
	GameDir       string // $HOME/.tw2ta_game
	InventoryLog  string
	DoorsLog      string
	MovementLog   string
	CurrentRoom   string
	WinCondition  string
	RoomsDir      string
}

// LoadMappingFromFile - читає YAML-файл маппінгу
func LoadMappingFromFile(path string) (*BashMapping, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("помилка читання %s: %w", path, err)
	}

	return ParseYAMLMapping(string(data))
}

// ParseYAMLMapping - парсить YAML контент у структури
func ParseYAMLMapping(content string) (*BashMapping, error) {
	// Спроба розпарсити новий формат
	var newRaw NewYAMLRoot
	if err := yaml.Unmarshal([]byte(content), &newRaw); err == nil && len(newRaw.Actions) > 0 {
		mapping := &BashMapping{
			Actions: make(map[string]*ActionMapping),
			// Заповнюємо глобальні налаштування з нового YAML
			GameDir: newRaw.TemplateVars.GameDir,
			InventoryLog: newRaw.TemplateVars.InventoryLog,
			DoorsLog: newRaw.TemplateVars.DoorsLog,
			MovementLog: newRaw.TemplateVars.MovementLog,
			CurrentRoom: newRaw.TemplateVars.CurrentRoom,
			WinCondition: newRaw.TemplateVars.WinCondition,
			RoomsDir: newRaw.TemplateVars.RoomsDir,
		}

		for _, action := range newRaw.Actions {
			// Використовуємо PlayerAction як ключ для внутрішньої мапи
			mapping.Actions[action.PlayerAction] = &ActionMapping{
				Name:          action.PlayerAction, // Name тепер PlayerAction
				PlayerCommand: action.BashCommand,
				PreCmd:        action.PreCmd,
				Test:          action.Test,
				PostCmd:       action.PostCmd,
				// TextWorldCmd та Description не мапляться безпосередньо з нового формату
			}
		}
		return mapping, nil
	}

	// Якщо новий формат не спрацював, спробуємо старий формат (для сумісності)
	var oldRaw OldLab1Mapping
	if err := yaml.Unmarshal([]byte(content), &oldRaw); err != nil {
		return nil, fmt.Errorf("помилка парсингу YAML: не вдалося розпарсити ні новий, ні старий формат: %w", err)
	}

	mapping := &BashMapping{
		Actions: make(map[string]*ActionMapping),
		GameDir: oldRaw.TemplateVars.GameDir,
		InventoryLog: oldRaw.TemplateVars.InventoryLog,
		DoorsLog: oldRaw.TemplateVars.DoorsLog,
		MovementLog: oldRaw.TemplateVars.MovementLog,
		CurrentRoom: oldRaw.TemplateVars.CurrentRoom,
		WinCondition: oldRaw.TemplateVars.WinCondition,
		RoomsDir: oldRaw.TemplateVars.RoomsDir,
	}

	// Конвертуємо старі шаблони у ActionMapping
	for name, tpl := range oldRaw.Actions {
		mapping.Actions[name] = &ActionMapping{
			Name:          name,
			Description:   tpl.Description,
			TextWorldCmd:  tpl.TextWorldExample,
			PlayerCommand: tpl.PlayerCommand,
			Test:          tpl.Test,
			PreCmd:        tpl.PreCmd,
			PostCmd:       tpl.PostCmd,
		}
	}

	return mapping, nil
}

// LoadMappingFromPaths - шукає YAML-файл у списку шляхів
func LoadMappingFromPaths(searchPaths []string) *BashMapping {
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			mapping, err := LoadMappingFromFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Помилка читання %s: %v
", path, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "✅ Завантажено мапінг: %s (%d дій)
", path, len(mapping.Actions))
			return mapping
		}
	}
	fmt.Fprintf(os.Stderr, "⚠️  lab1_mapping.yaml не знайдено, використовуємо вбудовані правила
")
	return nil
}

// GetAction - повертає правила для конкретної дії
func (bm *BashMapping) GetAction(actionName string) (*ActionMapping, error) {
	action, exists := bm.Actions[actionName]
	if !exists {
		return nil, fmt.Errorf("дія '%s' не знайдена у мапінгу", actionName)
	}
	return action, nil
}

// ApplyTemplate - замінює плейсхолдери {container}, {item}, тощо
func (am *ActionMapping) ApplyTemplate(vars map[string]string) (precmd, command, test, postcmd string) {
	precmd = am.replaceVars(am.PreCmd, vars)
	command = am.replaceVars(am.PlayerCommand, vars)
	test = am.replaceVars(am.Test, vars)
	postcmd = am.replaceVars(am.PostCmd, vars)
	return
}

// replaceVars - замінює {var} на значення з мапи
func (am *ActionMapping) replaceVars(template string, vars map[string]string) string {
	result := template
	for key, value := range vars {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// FindAllActions - повертає список всіх знайдених дій
func (bm *BashMapping) FindAllActions() []string {
	actions := make([]string, 0, len(bm.Actions))
	for name := range bm.Actions {
		actions = append(actions, name)
	}
	return actions
}
