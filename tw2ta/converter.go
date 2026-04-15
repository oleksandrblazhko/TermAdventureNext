package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	// Шукаємо та парсимо YAML-мапінг
	mapping := loadMappingFile(mappingPath)

	var ta strings.Builder

	// 1. Вступний рівень
	introLevel := generateIntroLevel(gs, challengeName)
	ta.WriteString(introLevel)
	ta.WriteString("\n\n" + strings.Repeat("-", 20) + "\n\n")

	// 2. Генеруємо рівні для кожного кроку квесту
	// Використовуємо walkthrough для правильного порядку дій
	walkthroughActions := gs.GetWalkthroughActions()
	totalActions := len(walkthroughActions)

	for i, action := range walkthroughActions {
		actionCounter := i + 1
		level := generateActionLevel(gs, action, actionCounter, totalActions, mapping)
		ta.WriteString(level)

		// Додаємо роздільник між рівнями (крім останнього)
		if actionCounter < totalActions {
			ta.WriteString("\n\n" + strings.Repeat("-", 20) + "\n\n")
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
	// Якщо явно вказано файл
	if mappingPath != "" {
		mapping, err := LoadMappingFromFile(mappingPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Помилка читання %s: %v\n", mappingPath, err)
			// Пробуємо шляхи за замовчуванням, якщо вказаний не знайдено
		} else {
			fmt.Fprintf(os.Stderr, "✅ Завантажено мапінг: %s (%d дій)\n", mappingPath, len(mapping.Actions))
			return mapping
		}
	}

	// Шукаємо YAML-файл у різних місцях (дефолт)
	searchPaths := []string{
		"prompts/Lab1_Bash_Scripts.yaml",
		"../prompts/Lab1_Bash_Scripts.yaml",
		"Lab1_Bash_Scripts.yaml",
	}

	return LoadMappingFromPaths(searchPaths)
}

// generateIntroLevel - створює вступний рівень
func generateIntroLevel(gs *GameState, challengeName string) string {
	startRoomDesc := gs.GetRoomDescription(gs.PlayerRoom)

	text := fmt.Sprintf(`# Ласкаво просимо до TermAdventure!

**Челендж:** %s
**Тема:** %s

%s

**Ваше завдання:** Виконуйте кроки послідовно. Кожен крок — це окремий рівень.

Введіть "help" щоб побачити доступні команди.`,
		challengeName,
		gs.Objective,
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
func generateActionLevel(gs *GameState, action ActionStep, currentStep int, totalSteps int, mapping *BashMapping) string {
	levelName := fmt.Sprintf("step_%02d", currentStep)

	// Визначаємо test, precmd, postcmd команди з мапінгу
	testCmd, preCmd, postCmd, playerCommand := generateCommandsFromMapping(gs, action, mapping)

	// Визначаємо наступний рівень
	var nextLevels []string
	if currentStep < totalSteps {
		nextLevels = []string{fmt.Sprintf("step_%02d", currentStep+1)}
	} else {
		nextLevels = []string{"final"}
	}

	// Генеруємо текст рівня
	text := generateLevelText(gs, action, currentStep, totalSteps, playerCommand)

	level := LevelData{
		Name:       levelName,
		Test:       testCmd,
		PreCmd:     preCmd,
		PostCmd:    postCmd,
		NextLevels: nextLevels,
		Text:       text,
	}

	return renderLevel(level)
}

// getActionKeyFromTemplate - визначає ключ для мапінгу з шаблону команди
func getActionKeyFromTemplate(actionName, commandTemplate string) string {
	// Спрощені правила для визначення ключа
	switch {
	case strings.HasPrefix(actionName, "open/c"):
		return "open <container>"
	case strings.HasPrefix(actionName, "close/c"):
		return "close <container>"
	case strings.HasPrefix(actionName, "unlock/c"):
		return "unlock <container> with <key>"
	case strings.HasPrefix(actionName, "lock/c"):
		return "lock <container> with <key>"
	case strings.HasPrefix(actionName, "take/c"):
		return "take <object> from <container>"
	case strings.HasPrefix(actionName, "insert"):
		return "insert <object> into <container>"
	case strings.HasPrefix(actionName, "open/d"):
		return "open <door>"
	case strings.HasPrefix(actionName, "close/d"):
		return "close <door>"
	case strings.HasPrefix(actionName, "unlock/d"):
		return "unlock <door> with <key>"
	case strings.HasPrefix(actionName, "lock/d"):
		return "lock <door> with <key>"
	case strings.HasPrefix(actionName, "take/s"):
		return "take <object> from <supporter>"
	case strings.HasPrefix(actionName, "put"):
		return "put <object> on <supporter>"
	case actionName == "take":
		return "take <object>"
	case actionName == "drop":
		return "drop <object>"
	case actionName == "eat":
		return "eat <food>"
	case actionName == "inventory":
		return "inventory"
	case strings.HasPrefix(actionName, "go/"):
		return "go <direction>"
	case strings.HasPrefix(actionName, "examine"):
		return "examine <object>" // Узагальнюємо
	}
	return actionName // Повертаємо як є, якщо не знайшли відповідності
}

// generateCommandsFromMapping - генерує команди з мапінгу
func generateCommandsFromMapping(gs *GameState, action ActionStep, mapping *BashMapping) (test, precmd, postcmd, playerCmd string) {
	var actionMapping *ActionMapping
	if mapping != nil {
		// Використовуємо нову функцію для отримання правильного ключа
		actionKey := getActionKeyFromTemplate(action.ActionName, action.Command)
		actionMapping, _ = mapping.GetAction(actionKey)
	}

	if actionMapping != nil {
		vars := extractTemplateVars(gs, action, mapping)
		precmd, playerCmd, test, postcmd = actionMapping.ApplyTemplate(vars)
		return
	}

	// Fallback: вбудовані правила, якщо мапінг не знайдено
	fmt.Fprintf(os.Stderr, "⚠️  Мапінг для '%s' не знайдено, використовуються вбудовані правила.\n", action.ActionName)
	return generateFallbackCommands(gs, action)
}

var placeholderRegex = regexp.MustCompile(`\{([^}]+)\}`)

// extractTemplateVars - витягує змінні для шаблону, використовуючи Regex
func extractTemplateVars(gs *GameState, action ActionStep, mapping *BashMapping) map[string]string {
	vars := make(map[string]string)
	cmd := action.Command

	matches := placeholderRegex.FindAllStringSubmatch(cmd, -1)
	var key, item, container, door, supporter string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		placeholder := match[1] // вміст дужок
		entityType := string(placeholder[0])
		entityID := placeholder

		switch entityType {
		case "k":
			key = entityID
			item = entityID
		case "f", "o":
			item = entityID
		case "c":
			container = entityID
		case "d":
			door = entityID
		case "s":
			supporter = entityID
		}
	}

	if key != "" {
		vars["key"] = toBashName(gs.GetEntityName(key))
		vars["key_name"] = gs.GetEntityName(key)
	}
	if item != "" {
		vars["item"] = toBashName(gs.GetEntityName(item))
		vars["item_name"] = gs.GetEntityName(item)
	}
	if container != "" {
		vars["container"] = toBashName(gs.GetEntityName(container))
		vars["container_name"] = gs.GetEntityName(container)
	}
	if door != "" {
		vars["door"] = "door_" + toBashName(gs.GetEntityName(door))
		vars["door_name"] = gs.GetEntityName(door)
	}
	if supporter != "" {
		vars["surface"] = toBashName(gs.GetEntityName(supporter))
		vars["surface_name"] = gs.GetEntityName(supporter)
	}
	
	if strings.HasPrefix(action.ActionName, "go/") {
		vars["direction"] = strings.TrimPrefix(action.ActionName, "go/")
		if action.TargetRoom != "" {
			vars["room"] = gs.GetEntityName(action.TargetRoom)
			vars["room_name"] = gs.GetEntityName(action.TargetRoom)
		}
	} else if action.SourceRoom != "" {
		vars["room"] = toBashName(gs.GetEntityName(action.SourceRoom))
		vars["room_name"] = gs.GetEntityName(action.SourceRoom)
	}

	// Глобальні змінні
	if mapping != nil {
		vars["game_dir"] = mapping.GameDir
		vars["inventory_log"] = mapping.InventoryLog
		vars["doors_log"] = mapping.DoorsLog
		vars["movement_log"] = mapping.MovementLog
		vars["current_room"] = mapping.CurrentRoom
		vars["win_condition"] = mapping.WinCondition
		vars["rooms_dir"] = mapping.RoomsDir
	}

	return vars
}

// toBashName - конвертує читабельну назву у bash-безпечне ім'я
func toBashName(name string) string {
	if name == "" {
		return ""
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	return strings.Trim(name, "_")
}

// generateFallbackCommands - виправлені вбудовані правила
func generateFallbackCommands(gs *GameState, action ActionStep) (test, precmd, postcmd, playerCmd string) {
	actionName := action.ActionName
	vars := extractTemplateVars(gs, action, nil) // Використовуємо новий екстрактор

	item := vars["item"]
	container := vars["container"]
	door := vars["door"]
	supporter := vars["surface"]

	switch {
	case actionName == "open/c":
		test = fmt.Sprintf("test ! -f %s/%s/.closed", vars["game_dir"], container)
		precmd = fmt.Sprintf("mkdir -p %s/%s && touch %s/%s/.closed", vars["game_dir"], container, vars["game_dir"], container)
		playerCmd = fmt.Sprintf("rm %s/%s/.closed", vars["game_dir"], container)
		postcmd = fmt.Sprintf("touch %s/%s/.open", vars["game_dir"], container)

	case actionName == "take/c":
		test = fmt.Sprintf("test -f ~/%s", item)
		precmd = fmt.Sprintf("mkdir -p %s/%s && touch %s/%s/%s", vars["game_dir"], container, vars["game_dir"], container, item)
		playerCmd = fmt.Sprintf("mv %s/%s/%s ~/", vars["game_dir"], container, item) // Виправлено cp на mv
		postcmd = fmt.Sprintf("echo '%s taken' >> %s", item, vars["inventory_log"])

	case actionName == "put":
		test = fmt.Sprintf("test -f %s/%s/%s", vars["game_dir"], supporter, item)
		precmd = fmt.Sprintf("mkdir -p %s/%s && touch ~/%s", vars["game_dir"], supporter, item)
		playerCmd = fmt.Sprintf("mv ~/%s %s/%s/", item, vars["game_dir"], supporter) // Виправлено cp на mv
		postcmd = fmt.Sprintf("echo '%s put on %s' >> %s", item, supporter, vars["inventory_log"])

	case actionName == "unlock/d":
		test = fmt.Sprintf("test ! -f %s/%s/.locked", vars["game_dir"], door)
		precmd = fmt.Sprintf("mkdir -p %s/%s && touch %s/%s/.locked", vars["game_dir"], door, vars["game_dir"], door)
		playerCmd = fmt.Sprintf("rm %s/%s/.locked", vars["game_dir"], door) // Правильна дія для розблокування
		postcmd = fmt.Sprintf("echo 'door %s unlocked' >> %s", door, vars["doors_log"])

	case actionName == "open/d":
		test = fmt.Sprintf("test ! -f %s/%s/.closed", vars["game_dir"], door)
		precmd = fmt.Sprintf("mkdir -p %s/%s && touch %s/%s/.closed", vars["game_dir"], door, vars["game_dir"], door)
		playerCmd = fmt.Sprintf("rm %s/%s/.closed", vars["game_dir"], door)
		postcmd = fmt.Sprintf("echo 'door %s open' >> %s", door, vars["doors_log"])

	case strings.HasPrefix(actionName, "go/"):
		if action.TargetRoom != "" {
			test = fmt.Sprintf("test \"$(cat %s)\" = \"%s\"", vars["current_room"], action.TargetRoom)
			precmd = fmt.Sprintf("mkdir -p %s && echo '%s' > %s", vars["rooms_dir"], action.SourceRoom, vars["current_room"])
			playerCmd = fmt.Sprintf("echo '%s' > %s", action.TargetRoom, vars["current_room"])
			postcmd = fmt.Sprintf("echo 'Moved to %s at $(date)' >> %s", action.TargetRoom, vars["movement_log"])
		} else {
			test, playerCmd = "true", "echo 'Перейдіть у наступну кімнату'"
		}

	default:
		test = "true"
		playerCmd = "# Виконайте дію: " + action.Command
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

	fmt.Fprintf(&buf, "name: %s\n", yamlQuote(level.Name))
	fmt.Fprintf(&buf, "test: %s\n", yamlQuote(level.Test))
	if level.PreCmd != "" {
		fmt.Fprintf(&buf, "precmd: %s\n", yamlQuote(level.PreCmd))
	}
	if level.PostCmd != "" {
		fmt.Fprintf(&buf, "postcmd: %s\n", yamlQuote(level.PostCmd))
	}
	if level.PostPrintCmd != "" {
		fmt.Fprintf(&buf, "postprintcmd: %s\n", yamlQuote(level.PostPrintCmd))
	}
	if len(level.NextLevels) > 0 {
		fmt.Fprintf(&buf, "next: [%s]\n", strings.Join(level.NextLevels, ", "))
	}
	if level.BackgroundJobs {
		fmt.Fprintf(&buf, "bgjobs: true\n")
	}
	if level.TimeLimit > 0 {
		fmt.Fprintf(&buf, "timelimit: %d\n", level.TimeLimit)
	}

	buf.WriteString("\n\n")
	buf.WriteString(level.Text)

	return buf.String()
}

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

func needsQuoting(s string) bool {
	if s == "" {
		return false
	}
	specialChars := []string{":", "\"", "'", "#", "{", "}", "[", "]", ",", "&", "*", "?", "|", "-", "<", ">", "=", "!", "%", "@", "`"}
	for _, ch := range specialChars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	if s[0] == ' ' {
		return true
	}
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	return false
}

func generateLevelText(gs *GameState, action ActionStep, currentStep int, totalSteps int, playerCommand string) string {
	var text strings.Builder

	text.WriteString(fmt.Sprintf("## Крок %d/%d: %s\n\n", currentStep, totalSteps, getActionTitle(gs, action)))
	text.WriteString(getActionInstructions(gs, action))

	if playerCommand != "" && !strings.HasPrefix(playerCommand, "#") {
		text.WriteString("\n**Виконайте команду:**\n")
		text.WriteString(fmt.Sprintf("\n```bash\n%s\n```\n", playerCommand))
	} else if strings.HasPrefix(playerCommand, "#") {
		text.WriteString(fmt.Sprintf("\n%s\n", strings.TrimPrefix(playerCommand, "# ")))
	}

	if action.Command != "" {
		text.WriteString(fmt.Sprintf("\n*Оригінальна команда TextWorld:* `%s`*\n", cleanCommand(action.Command)))
	}

	return text.String()
}

func getActionTitle(gs *GameState, action ActionStep) string {
	// Використовуємо getActionKeyFromTemplate для узагальнення
	key := getActionKeyFromTemplate(action.ActionName, action.Command)
	switch key {
	case "open <container>":
		return "Відчиніть контейнер"
	case "close <container>":
		return "Зачиніть контейнер"
	case "open <door>":
		return "Відчиніть двері"
	case "close <door>":
		return "Зачиніть двері"
	case "unlock <door> with <key>", "unlock <container> with <key>":
		return "Відімкніть"
	case "lock <door> with <key>", "lock <container> with <key>":
		return "Замкніть"
	case "take <object> from <container>", "take <object> from <supporter>", "take <object>":
		return "Візьміть предмет"
	case "insert <object> into <container>":
		return "Покладіть в контейнер"
	case "put <object> on <supporter>":
		return "Покладіть на поверхню"
	case "go <direction>":
		return fmt.Sprintf("Йдіть %s", vars["direction"]) // Vars may not be available here, simple title
	case "eat <food>":
		return "З'їжте їжу"
	default:
		return "Виконайте дію"
	}
}

func getActionInstructions(gs *GameState, action ActionStep) string {
	var text strings.Builder
	vars := extractTemplateVars(gs, action, nil)

	roomName := vars["room_name"]
	itemName := vars["item_name"]
	containerName := vars["container_name"]
	doorName := vars["door_name"]
	supporterName := vars["surface_name"]
	keyName := vars["key_name"]

	if roomName != "" {
		text.WriteString(fmt.Sprintf("Ви знаходитесь у кімнаті **%s**.\n\n", roomName))
	}

	key := getActionKeyFromTemplate(action.ActionName, action.Command)
	switch key {
	case "open <container>":
		text.WriteString(fmt.Sprintf("Перед вами **%s** — він зачинений.\n", containerName))
	case "take <object> from <container>":
		text.WriteString(fmt.Sprintf("Всередині **%s** ви бачите **%s**.\n", containerName, itemName))
	case "unlock <door> with <key>", "unlock <container> with <key>":
		entity := doorName
		if entity == "" {
			entity = containerName
		}
		text.WriteString(fmt.Sprintf("**%s** замкнені. Вам потрібен **%s**.\n", entity, keyName))
	case "open <door>":
		text.WriteString(fmt.Sprintf("**%s** зачинені, але відімкнені. Відчиніть їх.\n", doorName))
	case "go <direction>":
		if vars["room_name"] != "" {
			text.WriteString(fmt.Sprintf("Перейдіть до кімнати **%s**.\n", vars["room_name"]))
		}
	case "put <object> on <supporter>":
		text.WriteString(fmt.Sprintf("У вас є **%s**. Покладіть його на **%s**.\n", itemName, supporterName))
	case "take <object>":
		text.WriteString(fmt.Sprintf("Перед вами **%s**. Візьміть його.\n", itemName))
	}

	text.WriteString("\nВиконайте потрібну дію.\n")

	return text.String()
}

func cleanCommand(command string) string {
	command = strings.ReplaceAll(command, "{", "")
	command = strings.ReplaceAll(command, "}", "")
	return command
}
