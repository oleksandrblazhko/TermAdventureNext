package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ActionMapping - правила конвертації однієї дії TextWorld → Bash
type ActionMapping struct {
	Name        string // Назва дії (open/c, take/c, тощо)
	TextWorldCmd string // Приклад команди TextWorld
	Description string // Опис українською

	PreCmd    string // precmd (підготовка)
	Command   string // Команда для гравця
	Test      string // test (перевірка)
	PostCmd   string // postcmd (фіксація)
}

// BashMapping - всі правила мапінгу
type BashMapping struct {
	Actions map[string]*ActionMapping
	
	// Глобальні налаштування
	GameDir       string // $HOME/.tw2ta_game
	ItemsDir      string // /tmp/items
	InventoryDir  string // ~/
}

// LoadMappingFromFile - читає та парсить TW_BASH_MAPPING.md
func LoadMappingFromFile(filepath string) (*BashMapping, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("помилка читання %s: %w", filepath, err)
	}

	return ParseMarkdownMapping(string(data))
}

// ParseMarkdownMapping - парсить Markdown контент у структури
func ParseMarkdownMapping(content string) (*BashMapping, error) {
	mapping := &BashMapping{
		Actions: make(map[string]*ActionMapping),
		GameDir: "$HOME/.tw2ta_game",
		ItemsDir: "/tmp/items",
		InventoryDir: "~",
	}

	// 1. Знаходимо секції дій (### `action_name`)
	actionSectionRegex := regexp.MustCompile(`^###\s+` + "`([^`]+)`" + `\s*—\s*(.+)$`)
	
	// 2. Знаходимо коди в блоках
	codeBlockRegex := regexp.MustCompile("```bash\\n([\\s\\S]*?)```")

	// Розділяємо контент на секції
	sections := strings.Split(content, "\n### ")
	
	for _, section := range sections {
		lines := strings.Split(strings.TrimSpace(section), "\n")
		if len(lines) == 0 {
			continue
		}
		
		// Парсимо заголовок секції
		headerLine := lines[0]
		headerMatch := actionSectionRegex.FindStringSubmatch("### " + headerLine)
		if headerMatch == nil {
			// Можливо це не секція дії (огляд, тощо)
			continue
		}

		actionName := headerMatch[1]
		description := headerMatch[2]
		
		mapping.Actions[actionName] = &ActionMapping{
			Name: actionName,
			Description: description,
		}
		
		action := mapping.Actions[actionName]
		
		// Шукаємо TextWorld команду
		twCmdRegex := regexp.MustCompile(`\*\*TextWorld:\*\*\s+` + "`([^`]+)`")
		if match := twCmdRegex.FindStringSubmatch(section); match != nil {
			action.TextWorldCmd = match[1]
		}
		
		// Шукаємо таблицю з полями
		sectionText := strings.Join(lines[1:], "\n")
		tableLines := strings.Split(sectionText, "\n")
		inTable := false
		
		for _, line := range tableLines {
			line = strings.TrimSpace(line)
			
			// Початок таблиці
			if strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "|---") {
				inTable = true
				parts := strings.Split(line, "|")
				
				// Очищаємо частини
				cleanParts := make([]string, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						cleanParts = append(cleanParts, p)
					}
				}
				
				// Маємо 3 колонки: Поле, Значення, (опціонально)
				if len(cleanParts) >= 2 {
					fieldName := cleanParts[0]
					value := cleanParts[1]
					
					switch fieldName {
					case "precmd", "`precmd`":
						action.PreCmd = value
					case "test", "`test`":
						action.Test = value
					case "Команда гравця", "**Команда гравця**":
						action.Command = value
					case "postcmd", "`postcmd`":
						action.PostCmd = value
					}
				}
			} else if strings.HasPrefix(line, "|---") {
				inTable = true
				continue
			} else if line == "" || strings.HasPrefix(line, "**Логіка:**") {
				if inTable && line == "" {
					// Кінець таблиці
					inTable = false
				}
			}
		}
		
		// Якщо команда не знайдена в таблиці, шукаємо в блоці коду
		if action.Command == "" {
			if codeMatch := codeBlockRegex.FindStringSubmatch(sectionText); codeMatch != nil {
				// Беремо перший рядок блоку як команду
				lines := strings.Split(strings.TrimSpace(codeMatch[1]), "\n")
				if len(lines) > 0 {
					action.Command = strings.TrimSpace(lines[0])
				}
			}
		}
	}

	return mapping, nil
}

// GetAction - повертає правила для конкретної дії
func (bm *BashMapping) GetAction(actionName string) (*ActionMapping, error) {
	action, exists := bm.Actions[actionName]
	if !exists {
		return nil, fmt.Errorf("дія '%s' не знайдена у TW_BASH_MAPPING.md", actionName)
	}
	return action, nil
}

// ApplyTemplate - замінює плейсхолдери {container}, {item}, тощо
func (am *ActionMapping) ApplyTemplate(vars map[string]string) (precmd, command, test, postcmd string) {
	precmd = am.replaceVars(am.PreCmd, vars)
	command = am.replaceVars(am.Command, vars)
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
