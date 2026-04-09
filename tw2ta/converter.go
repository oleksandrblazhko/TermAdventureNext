package main

import (
	"fmt"
	"strings"
	"text/template"
	"bytes"
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
func ConvertToTA(gs *GameState, challengeName string) (string, error) {
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
	for qIndex, quest := range gs.Quests {
		// Пропускаємо квести без дій
		if len(quest.Actions) == 0 {
			continue
		}

		for aIndex, action := range quest.Actions {
			actionCounter++
			
			level := generateActionLevel(gs, action, actionCounter, totalActions, qIndex)
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

// generateIntroLevel - створює вступний рівень
func generateIntroLevel(gs *GameState, challengeName string) string {
	startRoom := gs.Rooms[gs.PlayerRoom]
	startRoomDesc := gs.GetRoomDescription(gs.PlayerRoom)

	text := fmt.Sprintf(`# Ласкаво просимо до TextWorld!

**Челендж:** %s
**Тема:** %s

%s

**Ваше завдання:** Виконуйте кроки послідовно. Кожен крок — це окремий рівень.

Введіть \`help\` щоб побачити доступні команди.`, 
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
func generateActionLevel(gs *GameState, action ActionStep, currentStep int, totalSteps int, questIndex int) string {
	levelName := fmt.Sprintf("step_%02d", currentStep)
	
	// Визначаємо test команду
	testCmd := generateTestCommand(gs, action)
	
	// Визначаємо precmd
	preCmd := fmt.Sprintf("echo 'Крок %d/%d: %s'", currentStep, totalSteps, getActionDescription(gs, action))
	
	// Визначаємо наступний рівень
	var nextLevels []string
	if currentStep < totalSteps {
		nextLevels = []string{fmt.Sprintf("step_%02d", currentStep+1)}
	} else {
		nextLevels = []string{"final"}
	}

	// Генеруємо текст рівня
	text := generateLevelText(gs, action, currentStep, totalSteps)

	level := LevelData{
		Name:       levelName,
		Test:       testCmd,
		PreCmd:     preCmd,
		NextLevels: nextLevels,
		Text:       text,
	}

	return renderLevel(level)
}

// generateFinalLevel - створює фінальний рівень
func generateFinalLevel(gs *GameState) string {
	text := `# 🎉 Вітаємо! Квест пройдено!

Ви успішно виконали всі кроки челенджу!

**Статистика:**
- Квестів пройдено: %d
- Винагорода: %d балів

Тепер ви можете завершити сесію командою \`exit\`.`

	totalReward := 0
	for _, quest := range gs.Quests {
		totalReward += quest.Reward
	}

	level := LevelData{
		Name: "final",
		Test: "true",
		PreCmd: fmt.Sprintf("echo 'Квест завершено! Загальна винагорода: %d балів'", totalReward),
		Text: fmt.Sprintf(text, len(gs.Quests), totalReward),
	}

	return renderLevel(level)
}

// renderLevel - рендерить рівень у формат .ta
func renderLevel(level LevelData) string {
	var buf bytes.Buffer

	// YAML метадані
	fmt.Fprintf(&buf, "name: %s\n", level.Name)
	fmt.Fprintf(&buf, "test: %s\n", level.Test)
	
	if level.PreCmd != "" {
		fmt.Fprintf(&buf, "precmd: %s\n", level.PreCmd)
	}
	if level.PostCmd != "" {
		fmt.Fprintf(&buf, "postcmd: %s\n", level.PostCmd)
	}
	if level.PostPrintCmd != "" {
		fmt.Fprintf(&buf, "postprintcmd: %s\n", level.PostPrintCmd)
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

// generateTestCommand - генерує shell-команду для перевірки кроку
func generateTestCommand(gs *GameState, action ActionStep) string {
	actionName := action.ActionName
	
	// Мапінг типів дій у test-команди
	switch {
	case actionName == "open/c":
		// Відчинити контейнер
		containerID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check open %s", containerID)
		
	case actionName == "close/c":
		containerID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check closed %s", containerID)
		
	case actionName == "open/d":
		// Відчинити двері
		doorID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check open %s", doorID)
		
	case actionName == "close/d":
		doorID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check closed %s", doorID)
		
	case actionName == "unlock/d":
		// Відімкнути двері
		doorID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check unlocked %s", doorID)
		
	case actionName == "lock/d":
		doorID := extractEntityID(action.Command)
		return fmt.Sprintf("game_state.sh check locked %s", doorID)
		
	case actionName == "take/c":
		// Взяти з контейнера
		itemID := extractItemID(action.Command)
		return fmt.Sprintf("game_state.sh has %s", itemID)
		
	case actionName == "take/s":
		// Взяти з поверхні
		itemID := extractItemID(action.Command)
		return fmt.Sprintf("game_state.sh has %s", itemID)
		
	case actionName == "insert":
		// Покласти в контейнер
		itemID := extractItemID(action.Command)
		return fmt.Sprintf("game_state.sh check in %s", itemID)
		
	case actionName == "put":
		// Покласти на поверхню
		itemID := extractItemID(action.Command)
		supporterID := extractSupporterID(action.Command)
		return fmt.Sprintf("game_state.sh check on %s %s", itemID, supporterID)
		
	case strings.HasPrefix(actionName, "go/"):
		// Рух - перевірка поточної кімнати
		if action.SourceRoom != "" {
			return fmt.Sprintf("game_state.sh at %s", action.SourceRoom)
		}
		return "true"
		
	case actionName == "trigger":
		// Автоматична подія (перевірка умови перемоги)
		if len(action.Postconditions) > 0 {
			lastPost := action.Postconditions[len(action.Postconditions)-1]
			if lastPost.Name == "event" {
				return "game_state.sh check win"
			}
		}
		return "true"
		
	default:
		return "true"
	}
}

// generateLevelText - генерує Markdown текст для рівня
func generateLevelText(gs *GameState, action ActionStep, currentStep int, totalSteps int) string {
	var text strings.Builder

	// Заголовок кроку
	text.WriteString(fmt.Sprintf("## Крок %d/%d: %s\n\n", currentStep, totalSteps, getActionTitle(gs, action)))

	// Опис дії
	text.WriteString(getActionInstructions(gs, action))

	// Команда
	if action.Command != "" {
		text.WriteString(fmt.Sprintf("\n**Команда:** `%s`\n", cleanCommand(action.Command)))
	}

	// Підказка якщо є fail_conditions
	// (буде додано пізніше)

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
		if room, ok := gs.Rooms[action.SourceRoom]; ok {
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

// _ - заглушка для невикористаного імпорту (template)
var _ = template.FuncMap{}
