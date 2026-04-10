package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// LevelData - дані для шаблону рівня
type LevelData struct {
	Name        string
	Test        string
	PreCmd      string
	PostCmd     string
	PostPrintCmd string
	NextLevels  []string
	BackgroundJobs bool
	TimeLimit   int
	Text        string
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
	totalActions := 0
	for _, quest := range gs.Quests {
		totalActions += len(quest.Actions)
	}

	actionCounter := 0
	for _, quest := range gs.Quests {
		// Пропускаємо квести без дій
		if len(quest.Actions) == 0 {
			continue
		}

		for _, action := range quest.Actions {
			actionCounter++
			
			level := generateActionLevel(gs, action, actionCounter, totalActions, mapping)
			ta.WriteString(level)
			
			// Додаємо роздільник між рівнями (крім останнього)
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
	// Якщо явно вказано файл
	if mappingPath != "" {
		mapping, err := LoadMappingFromFile(mappingPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Помилка читання %s: %v\n", mappingPath, err)
			return nil
		}
		fmt.Fprintf(os.Stderr, "✅ Завантажено мапінг: %s (%d дій)\n", mappingPath, len(mapping.Actions))
		return mapping
	}

	// Шукаємо YAML-файл у різних місцях (дефолт)
	searchPaths := []string{
		"tw-simple_mapping.yaml",
		"../tw-simple_mapping.yaml",
		filepath.Join(os.Getenv("HOME"), "TermAdventureNext/tw2ta/tw-simple_mapping.yaml"),
	}

	return LoadMappingFromPaths(searchPaths)
}

// generateIntroLevel - створює вступний рівень
func generateIntroLevel(gs *GameState, challengeName string) string {
	startRoomDesc := gs.GetRoomDescription(gs.PlayerRoom)

	text := fmt.Sprintf(`# Ласкаво просимо до TextWorld!

**Челендж:** %s
**Тема:** %s

%s

**Ваше завдання:** Виконуйте кроки послідовно. Кожен крок — це окремий рівень.

Введіть "help" щоб побачити доступні команди.`, 
		challengeName, 
		gs.Quests[0].Desc,
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

// generateCommandsFromMapping - генерує команди з мапінгу
func generateCommandsFromMapping(gs *GameState, action ActionStep, mapping *BashMapping) (test, precmd, postcmd, playerCmd string) {
	// Отримуємо правила з мапінгу
	var actionMapping *ActionMapping
	if mapping != nil {
		actionMapping, _ = mapping.GetAction(action.ActionName)
	}

	// Якщо мапінг знайдено, використовуємо його
	if actionMapping != nil {
		// Збираємо змінні для шаблону
		vars := extractTemplateVars(gs, action, mapping)

		// Застосовуємо шаблон
		precmd, playerCmd, test, postcmd = actionMapping.ApplyTemplate(vars)
		return
	}

	// Fallback: вбудовані правила
	return generateFallbackCommands(gs, action)
}

// toBashName - конвертує читабельну назву у bash-безпечне ім'я
// "chest drawer" → "chest_drawer", "wooden door" → "wooden_door"
func toBashName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	// Прибираємо зайві підкреслення
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")
	return name
}

// extractTemplateVars - витягує змінні для шаблону
func extractTemplateVars(gs *GameState, action ActionStep, mapping *BashMapping) map[string]string {
	vars := make(map[string]string)

	// Витягуємо ID з command_template
	cmd := action.Command

	// Контейнери
	if idx := strings.Index(cmd, "{c_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			containerID := cmd[idx+1 : idx+end]
			if c, ok := gs.Containers[containerID]; ok {
				vars["container"] = toBashName(c.Name)
				vars["container_name"] = c.Name
			} else {
				vars["container"] = containerID
				vars["container_name"] = gs.GetEntityName(containerID)
			}
		}
	}

	// Предмети
	if idx := strings.Index(cmd, "{f_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			itemID := cmd[idx+1 : idx+end]
			if item, ok := gs.Items[itemID]; ok {
				vars["item"] = toBashName(item.Name)
				vars["item_name"] = item.Name
			} else {
				vars["item"] = itemID
				vars["item_name"] = gs.GetEntityName(itemID)
			}
		}
	} else if idx := strings.Index(cmd, "{k_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			itemID := cmd[idx+1 : idx+end]
			if item, ok := gs.Items[itemID]; ok {
				vars["item"] = toBashName(item.Name)
				vars["item_name"] = item.Name
			} else {
				vars["item"] = itemID
				vars["item_name"] = gs.GetEntityName(itemID)
			}
		}
	} else if idx := strings.Index(cmd, "{o_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			itemID := cmd[idx+1 : idx+end]
			if item, ok := gs.Items[itemID]; ok {
				vars["item"] = toBashName(item.Name)
				vars["item_name"] = item.Name
			} else {
				vars["item"] = itemID
				vars["item_name"] = gs.GetEntityName(itemID)
			}
		}
	}

	// Двері
	if idx := strings.Index(cmd, "{d_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			doorID := cmd[idx+1 : idx+end]
			if door, ok := gs.Doors[doorID]; ok {
				vars["door"] = "door_" + toBashName(door.Name)
				vars["door_name"] = door.Name
			} else {
				vars["door"] = doorID
				vars["door_name"] = gs.GetEntityName(doorID)
			}
		}
	}

	// Поверхні
	if idx := strings.Index(cmd, "{s_"); idx != -1 {
		end := strings.Index(cmd[idx:], "}")
		if end != -1 {
			supporterID := cmd[idx+1 : idx+end]
			if s, ok := gs.Supporters[supporterID]; ok {
				vars["surface"] = toBashName(s.Name)
				vars["surface_name"] = s.Name
			} else {
				vars["surface"] = supporterID
				vars["surface_name"] = gs.GetEntityName(supporterID)
			}
		}
	}

	// Ключі (для unlock)
	if strings.Contains(cmd, "with {") {
		if idx := strings.Index(cmd, "with {"); idx != -1 {
			start := idx + 6
			end := strings.Index(cmd[start:], "}")
			if end != -1 {
				keyID := cmd[start : start+end]
				if item, ok := gs.Items[keyID]; ok {
					vars["key"] = toBashName(item.Name)
					vars["key_name"] = item.Name
				} else {
					vars["key"] = keyID
					vars["key_name"] = gs.GetEntityName(keyID)
				}
			}
		}
	}
	
	// Кімнати (для руху)
	if strings.HasPrefix(action.ActionName, "go/") {
		direction := strings.TrimPrefix(action.ActionName, "go/")
		vars["direction"] = direction
		if action.TargetRoom != "" {
			vars["room"] = gs.GetEntityName(action.TargetRoom)
			vars["room_name"] = gs.GetEntityName(action.TargetRoom)
		}
	} else if action.SourceRoom != "" {
		// Для всіх інших дій — кімната де виконується дія
		vars["room"] = toBashName(gs.GetEntityName(action.SourceRoom))
		vars["room_name"] = gs.GetEntityName(action.SourceRoom)
	}

	// Глобальні змінні з YAML
	if mapping != nil {
		vars["inventory_log"] = mapping.InventoryLog
		vars["doors_log"] = mapping.DoorsLog
		vars["movement_log"] = mapping.MovementLog
		vars["current_room"] = mapping.CurrentRoom
		vars["win_condition"] = mapping.WinCondition
		vars["rooms_dir"] = mapping.RoomsDir
		vars["game_dir"] = mapping.GameDir
	}

	return vars
}

// generateFallbackCommands - вбудовані правила якщо мапінг не знайдено
func generateFallbackCommands(gs *GameState, action ActionStep) (test, precmd, postcmd, playerCmd string) {
	actionName := action.ActionName

	switch {
	case actionName == "open/c":
		containerID := extractEntityID(action.Command)
		test = fmt.Sprintf("test ! -f $HOME/.tw2ta_game/%s/.closed", containerID)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s && touch $HOME/.tw2ta_game/%s/.closed", containerID, containerID)
		playerCmd = fmt.Sprintf("rm $HOME/.tw2ta_game/%s/.closed", containerID)
		postcmd = fmt.Sprintf("touch $HOME/.tw2ta_game/%s/.open", containerID)

	case actionName == "take/c":
		itemID := extractItemID(action.Command)
		containerID := extractContainerFromCommand(action.Command)
		test = fmt.Sprintf("test -f ~/%s", itemID)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s && touch $HOME/.tw2ta_game/%s/%s", containerID, containerID, itemID)
		playerCmd = fmt.Sprintf("cp $HOME/.tw2ta_game/%s/%s ~/", containerID, itemID)
		postcmd = fmt.Sprintf("echo '%s taken' >> $HOME/.tw2ta_game/inventory.log", itemID)

	case actionName == "put":
		itemID := extractItemID(action.Command)
		supporterID := extractSupporterID(action.Command)
		test = fmt.Sprintf("test -f $HOME/.tw2ta_game/%s/%s", supporterID, itemID)
		precmd = fmt.Sprintf("mkdir -p $HOME/.tw2ta_game/%s", supporterID)
		playerCmd = fmt.Sprintf("cp ~/%s $HOME/.tw2ta_game/%s/", itemID, supporterID)
		postcmd = fmt.Sprintf("rm ~/%s", itemID)

	case actionName == "unlock/d":
		doorID := extractEntityID(action.Command)
		test = fmt.Sprintf("test \"$(cat $HOME/.tw2ta_game/door_%s.state)\" = \"closed\"", doorID)
		precmd = fmt.Sprintf("echo 'locked' > $HOME/.tw2ta_game/door_%s.state", doorID)
		playerCmd = fmt.Sprintf("echo 'closed' > $HOME/.tw2ta_game/door_%s.state", doorID)
		postcmd = fmt.Sprintf("touch $HOME/.tw2ta_game/door_%s.unlocked", doorID)

	case actionName == "open/d":
		doorID := extractEntityID(action.Command)
		test = fmt.Sprintf("test \"$(cat $HOME/.tw2ta_game/door_%s.state)\" = \"open\"", doorID)
		precmd = fmt.Sprintf("echo 'closed' > $HOME/.tw2ta_game/door_%s.state", doorID)
		playerCmd = fmt.Sprintf("echo 'open' > $HOME/.tw2ta_game/door_%s.state", doorID)
		postcmd = fmt.Sprintf("echo 'door %s open' >> $HOME/.tw2ta_game/doors.log", doorID)

	case strings.HasPrefix(actionName, "go/"):
		if action.TargetRoom != "" {
			test = fmt.Sprintf("test \"$(cat $HOME/.tw2ta_game/current_room)\" = \"%s\"", action.TargetRoom)
			precmd = "mkdir -p $HOME/.tw2ta_game/rooms"
			playerCmd = fmt.Sprintf("echo '%s' > $HOME/.tw2ta_game/current_room", action.TargetRoom)
			postcmd = fmt.Sprintf("echo 'Moved to %s at $(date)' >> $HOME/.tw2ta_game/movement.log", action.TargetRoom)
		} else {
			test = "true"
			playerCmd = "echo 'Перейдіть у наступну кімнату'"
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
		Name: "final",
		Test: "true",
		PreCmd: fmt.Sprintf("echo 'Квест завершено! Загальна винагорода: %d балів'", totalReward),
		Text: text,
	}

	return renderLevel(level)
}

// renderLevel - рендерить рівень у формат .ta
func renderLevel(level LevelData) string {
	var buf bytes.Buffer

	// YAML метадані — використовуємо yaml.Marshal для безпечного екранування
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

	// Порожній рядок між метаданими та текстом
	buf.WriteString("\n\n")

	// Markdown текст
	buf.WriteString(level.Text)

	return buf.String()
}

// yamlQuote - безпечно квотує значення для YAML
// Якщо значення містить спеціальні символи — обгортає в подвійні лапки
// з екрануванням через yaml.Marshal
func yamlQuote(s string) string {
	s = strings.TrimSpace(s)
	// Прості значення без спецсимволів — залишаємо як є
	if !needsQuoting(s) {
		return s
	}
	// Використовуємо yaml.Marshal для коректного екранування
	data, err := yaml.Marshal(s)
	if err != nil {
		// Fallback: обгортаємо в подвійні лапки з екрануванням
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return strings.TrimSpace(string(data))
}

// needsQuoting — перевіряє чи значення потребує YAML-квайтування
func needsQuoting(s string) bool {
	if s == "" {
		return false
	}
	// Спецсимволи YAML
	specialChars := []string{":", "\"", "'", "#", "{", "}", "[", "]", ",", "&", "*", "?", "|", "-", "<", ">", "=", "!", "%", "@", "`"}
	for _, ch := range specialChars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	// Починається з пробілу
	if s[0] == ' ' {
		return true
	}
	// Схоже на число/булеве
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	return false
}


// generateLevelText - генерує Markdown текст для рівня
func generateLevelText(gs *GameState, action ActionStep, currentStep int, totalSteps int, playerCommand string) string {
	var text strings.Builder

	// Заголовок кроку
	text.WriteString(fmt.Sprintf("## Крок %d/%d: %s\n\n", currentStep, totalSteps, getActionTitle(gs, action)))

	// Опис дії
	text.WriteString(getActionInstructions(gs, action))

	// Команда для гравця
	if playerCommand != "" && !strings.HasPrefix(playerCommand, "#") {
		text.WriteString("\n**Виконайте команду:**\n")
		text.WriteString(fmt.Sprintf("\n```bash\n%s\n```\n", playerCommand))
	} else if strings.HasPrefix(playerCommand, "#") {
		// Коментар як команда
		text.WriteString(fmt.Sprintf("\n%s\n", strings.TrimPrefix(playerCommand, "# ")))
	}

	// Підказка щодо перевірки
	if action.Command != "" {
		text.WriteString(fmt.Sprintf("\n*Оригінальна команда TextWorld:* `%s`*\n", cleanCommand(action.Command)))
	}

	return text.String()
}

// getActionDescription - коротка назва дії для precmd
func getActionDescription(gs *GameState, action ActionStep) string {
	return getActionTitle(gs, action)
}

// getActionTitle - заголовок дії українською
func getActionTitle(gs *GameState, action ActionStep) string {
	switch action.ActionName {
	case "open/c":
		return "Відчиніть контейнер"
	case "close/c":
		return "Зачиніть контейнер"
	case "open/d":
		return "Відчиніть двері"
	case "close/d":
		return "Зачиніть двері"
	case "unlock/d":
		return "Відімкніть двері"
	case "lock/d":
		return "Замкніть двері"
	case "take/c":
		return "Візьміть предмет"
	case "take/s":
		return "Візьміть предмет"
	case "insert":
		return "Покладіть в контейнер"
	case "put":
		return "Покладіть на поверхню"
	case "go/east":
		return "Йдіть на схід"
	case "go/west":
		return "Йдіть на захід"
	case "go/north":
		return "Йдіть на північ"
	case "go/south":
		return "Йдіть на південь"
	case "trigger":
		return "Перевірка умови"
	default:
		return action.ActionName
	}
}

// getActionInstructions - детальна інструкція для рівня
func getActionInstructions(gs *GameState, action ActionStep) string {
	var text strings.Builder

	// Опис кімнати
	if action.SourceRoom != "" {
		if _, ok := gs.Rooms[action.SourceRoom]; ok {
			text.WriteString(fmt.Sprintf("Ви знаходитесь у кімнаті **%s**.\n\n", gs.GetEntityName(action.SourceRoom)))
			
			// Додаємо контекст про предмети
			switch action.ActionName {
			case "open/c":
				containerID := extractEntityID(action.Command)
				if container, ok := gs.Containers[containerID]; ok {
					text.WriteString(fmt.Sprintf("Перед вами **%s** — він зачинений.\n", gs.GetEntityName(containerID)))
					if len(container.Contents) > 0 {
						text.WriteString(fmt.Sprintf("Всередині щось є...\n"))
					}
				}
				
			case "take/c":
				itemID := extractItemID(action.Command)
				containerID := extractContainerFromCommand(action.Command)
				text.WriteString(fmt.Sprintf("Всередині **%s** ви бачите **%s**.\n", 
					gs.GetEntityName(containerID), gs.GetEntityName(itemID)))
				
			case "unlock/d":
				doorID := extractEntityID(action.Command)
				if door, ok := gs.Doors[doorID]; ok {
					text.WriteString(fmt.Sprintf("**%s** замкнені. ", gs.GetEntityName(doorID)))
					if door.KeyMatch != "" {
						text.WriteString(fmt.Sprintf("Вам потрібен **%s**.\n", gs.GetEntityName(door.KeyMatch)))
					}
				}
				
			case "open/d":
				doorID := extractEntityID(action.Command)
				text.WriteString(fmt.Sprintf("**%s** зачинені, але відімкнені. Відчиніть їх.\n", 
					gs.GetEntityName(doorID)))
				
			case "go/east", "go/west", "go/north", "go/south":
				if action.TargetRoom != "" {
					text.WriteString(fmt.Sprintf("Перейдіть до кімнати **%s**.\n", 
						gs.GetEntityName(action.TargetRoom)))
				}
				
			case "put":
				itemID := extractItemID(action.Command)
				supporterID := extractSupporterID(action.Command)
				text.WriteString(fmt.Sprintf("У вас є **%s**. Покладіть його на **%s**.\n", 
					gs.GetEntityName(itemID), gs.GetEntityName(supporterID)))
			}
		}
	}

	text.WriteString("\nВиконайте потрібну дію.\n")
	
	return text.String()
}

// Допоміжні функції для парсингу команд

func extractEntityID(command string) string {
	// Приклад: "open {c_0}" → "c_0"
	command = strings.TrimSpace(command)
	command = strings.Trim(command, "{}")
	parts := strings.SplitN(command, " ", 2)
	if len(parts) > 1 {
		return strings.Trim(parts[1], "{}")
	}
	return ""
}

func extractItemID(command string) string {
	// Приклад: "take {f_1} from {c_2}" → "f_1"
	command = strings.TrimSpace(command)
	parts := strings.SplitN(command, " ", 2)
	if len(parts) > 1 {
		return strings.Trim(parts[1], "{}")
	}
	return ""
}

func extractContainerFromCommand(command string) string {
	// Приклад: "take {k_0} from {c_0}" → "c_0"
	if idx := strings.Index(command, "from {"); idx != -1 {
		start := idx + 6
		end := strings.Index(command[start:], "}")
		if end != -1 {
			return command[start:start+end]
		}
	}
	return ""
}

func extractSupporterID(command string) string {
	// Приклад: "put {f_1} on {s_2}" → "s_2"
	if idx := strings.Index(command, "on {"); idx != -1 {
		start := idx + 4
		end := strings.Index(command[start:], "}")
		if end != -1 {
			return command[start:start+end]
		}
	}
	return ""
}

func cleanCommand(command string) string {
	// Прибираємо фігурні дужки для показу гравцю
	command = strings.ReplaceAll(command, "{", "")
	command = strings.ReplaceAll(command, "}", "")
	return command
}
