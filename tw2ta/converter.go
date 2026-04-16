package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// LevelData - дані для шаблону рівня
type LevelData struct {
	Name           string
	Test           string
	PreCmd         string
	PostCmd        string
	PostPrintCmd   string
	NextLevels     []string
	BackgroundJobs bool
	TimeLimit      int
	Text           string
}

// ConvertToTA - конвертує GameState у формат .ta
func ConvertToTA(gs *GameState, challengeName string, mappingPath string) (string, error) {
	mapping := loadMappingFile(mappingPath)
	var ta strings.Builder

	// 1. Вступний рівень
	introLevel := generateIntroLevel(gs, challengeName)
	ta.WriteString(introLevel)
	ta.WriteString("\n\n" + strings.Repeat("-", 20) + "\n\n")

	// 2. Генеруємо рівні для кожного кроку квесту (використовуємо оригінальний цикл)
	totalActions := 0
	for _, quest := range gs.Quests {
		totalActions += len(quest.Actions)
	}

	actionCounter := 0
	for _, quest := range gs.Quests {
		if len(quest.Actions) == 0 {
			continue
		}
		for _, action := range quest.Actions {
			actionCounter++
			level := generateActionLevel(gs, action, actionCounter, totalActions, mapping)
			ta.WriteString(level)
			if actionCounter < totalActions {
				ta.WriteString("\n\n" + strings.Repeat("-", 20) + "\n\n")
			}
		}
	}

	// 3. Фінальний рівень
	ta.WriteString("\n\n" + strings.Repeat("-", 20) + "\n\n")
	finalLevel := generateFinalLevel(gs)
	ta.WriteString(finalLevel)

	return ta.String(), nil
}

// loadMappingFile - завантажує YAML-мапінг
func loadMappingFile(mappingPath string) *BashMapping {
	if mappingPath != "" {
		mapping, err := LoadMappingFromFile(mappingPath)
		if err == nil {
			fmt.Fprintf(os.Stderr, "✅ Завантажено мапінг: %s (%d дій)\n", mappingPath, len(mapping.Actions))
			return mapping
		}
		fmt.Fprintf(os.Stderr, "❌ Помилка читання %s: %v. Пошук за замовчуванням...\n", mappingPath, err)
	}
	searchPaths := []string{"prompts/Lab1_Bash_Scripts.yaml", "../prompts/Lab1_Bash_Scripts.yaml", "Lab1_Bash_Scripts.yaml"}
	return LoadMappingFromPaths(searchPaths)
}

// generateIntroLevel - створює вступний рівень
func generateIntroLevel(gs *GameState, challengeName string) string {
	startRoomDesc := gs.GetRoomDescription(gs.PlayerRoom)
	objective := ""
	if len(gs.Quests) > 0 {
		objective = gs.Quests[0].Desc // Беремо опис з першого квесту
	}

	text := fmt.Sprintf(`# Ласкаво просимо до TermAdventure!

**Челендж:** %s
**Тема:** %s

%s

**Ваше завдання:** Виконуйте кроки послідовно. Кожен крок — це окремий рівень.

Введіть "help" щоб побачити доступні команди.`,
		challengeName,
		objective,
		startRoomDesc)

	level := LevelData{
		Name:       "intro",
		Test:       "true",
		PreCmd:     fmt.Sprintf("echo 'Початок челенджу: %s'", challengeName),
		NextLevels: []string{"step_01"},
		Text:       text,
	}
	return renderLevel(level)
}

// generateActionLevel - створює рівень для однієї дії
func generateActionLevel(gs *GameState, action ActionStep, currentStep, totalSteps int, mapping *BashMapping) string {
	levelName := fmt.Sprintf("step_%02d", currentStep)
	testCmd, preCmd, postCmd, playerCommand := generateCommandsFromMapping(gs, action, mapping)
	nextLevels := []string{fmt.Sprintf("step_%02d", currentStep+1)}
	if currentStep >= totalSteps {
		nextLevels = []string{"final"}
	}
	text := generateLevelText(gs, action, currentStep, totalSteps, playerCommand)

	level := LevelData{
		Name: levelName, Test: testCmd, PreCmd: preCmd, PostCmd: postCmd,
		NextLevels: nextLevels, Text: text,
	}
	return renderLevel(level)
}

// getActionKeyFromTemplate - визначає ключ для мапінгу з шаблону команди
func getActionKeyFromTemplate(actionName, commandTemplate string) string {
	switch {
	case strings.HasPrefix(actionName, "open/c"):
		return "open <container>"
	case strings.HasPrefix(actionName, "take/c"):
		return "take <object> from <container>"
	case strings.HasPrefix(actionName, "unlock/d"):
		return "unlock <door> with <key>"
	case strings.HasPrefix(actionName, "open/d"):
		return "open <door>"
	case strings.HasPrefix(actionName, "go/"):
		return "go <direction>"
	case strings.HasPrefix(actionName, "put"):
		return "put <object> on <supporter>"
	// Додайте інші правила тут
	}
	return actionName // Fallback
}

// generateCommandsFromMapping - генерує команди з мапінгу
func generateCommandsFromMapping(gs *GameState, action ActionStep, mapping *BashMapping) (test, precmd, postcmd, playerCmd string) {
	if mapping != nil {
		actionKey := getActionKeyFromTemplate(action.ActionName, action.Command)
		if actionMapping, err := mapping.GetAction(actionKey); err == nil {
			vars := extractTemplateVars(gs, action, mapping)
			precmd, playerCmd, test, postcmd = actionMapping.ApplyTemplate(vars)
			return
		}
	}
	// Fallback, якщо мапінг не знайдено
	fmt.Fprintf(os.Stderr, "⚠️  Мапінг для '%s' ('%s') не знайдено, використовуються вбудовані правила.\n", action.ActionName, getActionKeyFromTemplate(action.ActionName, action.Command))
	return generateFallbackCommands(gs, action)
}

var placeholderRegex = regexp.MustCompile(`\{([^}]+)\}`)

// extractTemplateVars - витягує змінні для шаблону, використовуючи Regex
func extractTemplateVars(gs *GameState, action ActionStep, mapping *BashMapping) map[string]string {
	vars := make(map[string]string)
	cmd := action.Command // e.g., "take {k_0} from {c_1}"
	
	// Add global variables if available
	if mapping != nil {
		vars["game_dir"] = mapping.GameDir
		vars["inventory_log"] = mapping.InventoryLog
		vars["doors_log"] = mapping.DoorsLog
		vars["movement_log"] = mapping.MovementLog
		vars["current_room"] = mapping.CurrentRoom
		vars["win_condition"] = mapping.WinCondition
		vars["rooms_dir"] = mapping.RoomsDir
	}

	// Extract specific variables based on command type and placeholders
	key := getActionKeyFromTemplate(action.ActionName, action.Command)

	// Generic function to extract ID from placeholder (e.g., "{c_1}" -> "c_1")
	extractID := func(placeholder string) string {
		matches := placeholderRegex.FindStringSubmatch(placeholder)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}

	switch key {
	case "open <container>", "close <container>":
		// Command example: "open {c_1}"
		containerID := extractID(strings.TrimPrefix(cmd, "open "))
		if containerID == "" { containerID = extractID(strings.TrimPrefix(cmd, "close ")) }
		vars["container"] = toBashName(gs.GetEntityName(containerID))
		vars["container_name"] = gs.GetEntityName(containerID)
	
	case "take <object> from <container>":
		// Command example: "take {k_0} from {c_1}"
		parts := strings.Split(cmd, " from ")
		if len(parts) == 2 {
			itemID := extractID(strings.TrimPrefix(parts[0], "take "))
			containerID := extractID(parts[1])
			vars["item"] = toBashName(gs.GetEntityName(itemID))
			vars["item_name"] = gs.GetEntityName(itemID)
			vars["container"] = toBashName(gs.GetEntityName(containerID))
			vars["container_name"] = gs.GetEntityName(containerID)
		}

	case "unlock <door> with <key>":
		// Command example: "unlock {d_0} with {k_0}"
		parts := strings.Split(cmd, " with ")
		if len(parts) == 2 {
			doorID := extractID(strings.TrimPrefix(parts[0], "unlock "))
			keyID := extractID(parts[1])
			vars["door"] = toBashName(gs.GetEntityName(doorID))
			vars["door_name"] = gs.GetEntityName(doorID)
			vars["key"] = toBashName(gs.GetEntityName(keyID))
			vars["key_name"] = gs.GetEntityName(keyID)
		}

	case "open <door>", "close <door>":
		// Command example: "open {d_0}"
		doorID := extractID(strings.TrimPrefix(cmd, "open "))
		if doorID == "" { doorID = extractID(strings.TrimPrefix(cmd, "close ")) }
		vars["door"] = toBashName(gs.GetEntityName(doorID))
		vars["door_name"] = gs.GetEntityName(doorID)
	
	case "go <direction>":
		// Command example: "go east"
		direction := strings.TrimPrefix(cmd, "go ")
		vars["direction"] = direction
		if action.TargetRoom != "" {
			vars["room"] = toBashName(gs.GetEntityName(action.TargetRoom))
			vars["room_name"] = gs.GetEntityName(action.TargetRoom)
		} else if action.SourceRoom != "" {
			// Fallback to source room if target not explicitly set
			vars["room"] = toBashName(gs.GetEntityName(action.SourceRoom))
			vars["room_name"] = gs.GetEntityName(action.SourceRoom)
		}

	case "put <object> on <supporter>":
		// Command example: "put {f_2} on {s_2}"
		parts := strings.Split(cmd, " on ")
		if len(parts) == 2 {
			itemID := extractID(strings.TrimPrefix(parts[0], "put "))
			supporterID := extractID(parts[1])
			vars["item"] = toBashName(gs.GetEntityName(itemID))
			vars["item_name"] = gs.GetEntityName(itemID)
			vars["supporter"] = toBashName(gs.GetEntityName(supporterID))
			vars["supporter_name"] = gs.GetEntityName(supporterID)
		}
	
	case "take <object>":
		// Command example: "take {k_0}"
		itemID := extractID(strings.TrimPrefix(cmd, "take "))
		vars["item"] = toBashName(gs.GetEntityName(itemID))
		vars["item_name"] = gs.GetEntityName(itemID)
	}

	// Always add room from action.SourceRoom if available, especially for descriptions.
	if action.SourceRoom != "" && vars["room"] == "" { // Only if not already set by "go" action
		vars["room"] = toBashName(gs.GetEntityName(action.SourceRoom))
		vars["room_name"] = gs.GetEntityName(action.SourceRoom)
	}
	
	return vars
}
