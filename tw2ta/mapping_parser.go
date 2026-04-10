package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// ActionMapping - правила конвертації однієї дії TextWorld → Bash
type ActionMapping struct {
	Name           string // Назва дії (open/c, take/c, тощо)
	TextWorldCmd   string // Приклад команди TextWorld
	Description    string // Опис українською
	PlayerCommand  string // Команда для гравця
	Test           string // test (перевірка)
	PreCmd         string // precmd (підготовка)
	PostCmd        string // postcmd (фіксація)
}

// ActionTemplate - сирий шаблон з YAML
type ActionTemplate struct {
	Description   string `yaml:"description"`
	TextWorldExample string `yaml:"textworld_example"`
	PlayerCommand string `yaml:"player_command"`
	Test          string `yaml:"test"`
	PreCmd        string `yaml:"precmd"`
	PostCmd       string `yaml:"postcmd"`
}

// TemplateVars - глобальні змінні для підстановки
type TemplateVars struct {
	GameDir       string `yaml:"game_dir"`
	InventoryLog  string `yaml:"inventory_log"`
	DoorsLog      string `yaml:"doors_log"`
	MovementLog   string `yaml:"movement_log"`
	CurrentRoom   string `yaml:"current_room"`
	WinCondition  string `yaml:"win_condition"`
	RoomsDir      string `yaml:"rooms_dir"`
}

// Lab1Mapping - коренева структура YAML-файлу
type Lab1Mapping struct {
	Actions         map[string]ActionTemplate `yaml:"action_templates"`
	TemplateVars    TemplateVars              `yaml:"template_vars"`
}

// BashMapping - всі правила мапінгу
type BashMapping struct {
	Actions map[string]*ActionMapping

	// Глобальні налаштування
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
	var raw Lab1Mapping
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("помилка парсингу YAML: %w", err)
	}

	mapping := &BashMapping{
		Actions: make(map[string]*ActionMapping),
		GameDir: raw.TemplateVars.GameDir,
		InventoryLog: raw.TemplateVars.InventoryLog,
		DoorsLog: raw.TemplateVars.DoorsLog,
		MovementLog: raw.TemplateVars.MovementLog,
		CurrentRoom: raw.TemplateVars.CurrentRoom,
		WinCondition: raw.TemplateVars.WinCondition,
		RoomsDir: raw.TemplateVars.RoomsDir,
	}

	// Конвертуємо сирые шаблони у ActionMapping
	for name, tpl := range raw.Actions {
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
				fmt.Fprintf(os.Stderr, "⚠️  Помилка читання %s: %v\n", path, err)
				continue
			}
			fmt.Fprintf(os.Stderr, "✅ Завантажено мапінг: %s (%d дій)\n", path, len(mapping.Actions))
			return mapping
		}
	}
	fmt.Fprintf(os.Stderr, "⚠️  lab1_mapping.yaml не знайдено, використовуємо вбудовані правила\n")
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
