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
	cmd := action.Command
	matches := placeholderRegex.FindAllStringSubmatch(cmd, -1)
	
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		placeholder := match[1]
		entityType := string(placeholder[0])
		
		vars[entityType+"_id"] = placeholder
		if name := gs.GetEntityName(placeholder); name != "" {
			vars[entityType+"_name"] = name
			vars[entityType] = toBashName(name)
		}
	}
	
	if mapping != nil {
		vars["game_dir"] = mapping.GameDir
		vars["inventory_log"] = mapping.InventoryLog
	}
	return vars
}

// toBashName - конвертує читабельну назву у bash-безпечне ім'я
func toBashName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// generateFallbackCommands - виправлені вбудовані правила
func generateFallbackCommands(gs *GameState, action ActionStep) (test, precmd, postcmd, playerCmd string) {
	actionName := action.ActionName
	vars := extractTemplateVars(gs, action, nil)
	item, container, door := vars["item"], vars["container"], vars["door"]

	switch {
	case actionName == "open/c":
		test = fmt.Sprintf("test ! -f $HOME/.tw2ta_game/%s/.closed", container)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s && touch $HOME/.tw2ta_game/%s/.closed", container, container)
		playerCmd = fmt.Sprintf("rm $HOME/.tw2ta_game/%s/.closed", container)
	case actionName == "take/c":
		test = fmt.Sprintf("test -f ~/%s", item)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s && touch $HOME/.tw2ta_game/%s/%s", container, container, item)
		playerCmd = fmt.Sprintf("mv $HOME/.tw2ta_game/%s/%s ~/", container, item)
	case actionName == "unlock/d":
		test = fmt.Sprintf("test ! -f $HOME/.tw2ta_game/%s/.locked", door)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s && touch $HOME/.tw2ta_game/%s/.locked", door, door)
		playerCmd = fmt.Sprintf("rm $HOME/.tw2ta_game/%s/.locked", door)
	default:
		test, playerCmd = "true", "# Виконайте дію: "+action.Command
	}
	return
}

// generateFinalLevel - створює фінальний рівень
func generateFinalLevel(gs *GameState) string {
	totalReward := 0
	for _, quest := range gs.Quests {
		totalReward += quest.Reward
	}

	text := fmt.Sprintf(`# 🎉 Вітаємо! Квест пройдено!

Ви успішно виконали всі кроки челенджу!

**Статистика:**
- Квестів пройдено: %d
- Винагорода: %d балів

Тепер ви можете завершити сесію командою "exit".`, len(gs.Quests), totalReward)

	level := LevelData{
		Name:   "final",
		Test:   "true",
		PreCmd: fmt.Sprintf("echo 'Квест завершено! Загальна винагорода: %d балів'", totalReward),
		Text:   text,
	}

	return renderLevel(level)
}

// renderLevel - рендерить рівень у формат .ta
func renderLevel(level LevelData) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "name: %s\n", level.Name)
	formatYamlField(&buf, "test", level.Test)
	formatYamlField(&buf, "precmd", level.PreCmd)
	formatYamlField(&buf, "postcmd", level.PostCmd)
	formatYamlField(&buf, "postprintcmd", level.PostPrintCmd)

	if level.TimeLimit > 0 {
		fmt.Fprintf(&buf, "timelimit: %d\n", level.TimeLimit)
	}
	if len(level.NextLevels) > 0 {
		fmt.Fprintf(&buf, "next: [%s]\n", strings.Join(level.NextLevels, ", "))
	}
	if level.BackgroundJobs {
		fmt.Fprintf(&buf, "bgjobs: true\n")
	}

	buf.WriteString("\n") // Один перенос рядка перед текстом
	buf.WriteString(level.Text)

	return buf.String()
}

// formatYamlField - форматує поле YAML, обробляючи багаторядкові значення
func formatYamlField(buf *bytes.Buffer, key, value string) {
	if value == "" {
		return
	}
	if strings.Contains(value, "\n") {
		fmt.Fprintf(buf, "%s: |\n", key)
		for _, line := range strings.Split(strings.TrimRight(value, "\n"), "\n") {
			fmt.Fprintf(buf, "  %s\n", line)
		}
	} else {
		fmt.Fprintf(buf, "%s: %s\n", key, yamlQuote(value))
	}
}

// yamlQuote - безпечно квотує значення для YAML
func yamlQuote(s string) string {
	s = strings.TrimSpace(s)
	if !needsQuoting(s) {
		return s
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return strings.TrimSpace(string(data))
}

// needsQuoting — перевіряє чи значення потребує YAML-квотування
func needsQuoting(s string) bool {
	if s == "" {
		return true // Пусті рядки теж треба квотувати
	}
	// Спецсимволи YAML
	specialChars := []string{":", "{", "}", "[", "]", ",", "&", "*", "#", "?", "|", "-", "<", ">", "=", "!", "%", "@", "`"}
	if strings.ContainsAny(s, specialChars) {
		return true
	}
	// Починається або закінчується пробілом
	if s[0] == ' ' || s[len(s)-1] == ' ' {
		return true
	}
	// Схоже на число/булеве
	if s == "true" || s == "false" || s == "null" || s == "TRUE" || s == "FALSE" || s == "NULL" {
		return true
	}
	
    // Check if it can be parsed as a number
    if _, err := strconv.ParseFloat(s, 64); err == nil {
        return true
    }
    if _, err := strconv.ParseInt(s, 10, 64); err == nil {
        return true
    }

	return false
}


// generateLevelText, getActionTitle, getActionInstructions - без суттєвих змін
// ...
func generateLevelText(gs *GameState, action ActionStep, currentStep int, totalSteps int, playerCommand string) string {
	var text strings.Builder

	text.WriteString(fmt.Sprintf("## Крок %d/%d: %s\n\n", currentStep, totalSteps, getActionTitle(gs, action)))
	text.WriteString(getActionInstructions(gs, action))

	if playerCommand != "" && !strings.HasPrefix(playerCommand, "#") {
		text.WriteString(fmt.Sprintf("\n**Виконайте команду:**\n\n```bash\n%s\n```\n", playerCommand))
	}

	text.WriteString(fmt.Sprintf("\n*Оригінальна команда TextWorld:* `%s`*\n", cleanCommand(action.Command)))

	return text.String()
}

func getActionTitle(gs *GameState, action ActionStep) string {
	key := getActionKeyFromTemplate(action.ActionName, action.Command)
	switch key {
	case "open <container>":
		return "Відчиніть контейнер"
	case "take <object> from <container>":
		return "Візьміть предмет"
	case "unlock <door> with <key>":
		return "Відімкніть"
	default:
		return "Виконайте дію"
	}
}

func getActionInstructions(gs *GameState, action ActionStep) string {
	vars := extractTemplateVars(gs, action, nil)
	return fmt.Sprintf("Ви знаходитесь у кімнаті **%s**.\n\nВиконайте потрібну дію.\n", vars["room_name"])
}

func cleanCommand(command string) string {
	return strings.NewReplacer("{", "", "}", "").Replace(command)
}
