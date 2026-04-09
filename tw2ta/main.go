package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	version = "1.0.0"
	usage   = `tw2ta - Конвертер TextWorld JSON → TermAdventure .ta

Використання:
  tw2ta [OPTIONS] <input.json> [output.ta]

Опис:
  Конвертує JSON-файл TextWorld у формат .ta для TermAdventure.
  Кожна дія з квесту TextWorld стає окремим рівнем.

Приклади:
  tw2ta simple_game.json
  tw2ta --output output.ta simple_game.json
  tw2ta --challenge "My Quest" game.json
  tw2ta --copy-game-state game.json

Опції:`
)

func main() {
	// Прапори
	outputFlag := flag.String("output", "", "Вихідний файл .ta (за замовчуванням: <input>.ta)")
	challengeFlag := flag.String("challenge", "", "Назва челенджу (за замовчуванням: з імені файлу)")
	copyGameState := flag.Bool("copy-game-state", false, "Копіювати game_state.sh до поточної директорії")
	versionFlag := flag.Bool("version", false, "Показати версію")
	helpFlag := flag.Bool("help", false, "Показати допомогу")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Показати версію
	if *versionFlag {
		fmt.Printf("tw2ta v%s\n", version)
		os.Exit(0)
	}

	// Показати допомогу
	if *helpFlag {
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Перевірка аргументів
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Помилка: відсутній вхідний файл JSON\n\n")
		flag.Usage()
		os.Exit(1)
	}

	inputFile := args[0]
	
	// Визначаємо вихідний файл
	var outputFile string
	if *outputFlag != "" {
		outputFile = *outputFlag
	} else {
		// <input.json> → <input>.ta
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".ta"
	}

	// Визначаємо назву челенджу
	var challengeName string
	if *challengeFlag != "" {
		challengeName = *challengeFlag
	} else {
		// З імені файлу: simple_game.json → simple_game
		challengeName = strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	}

	// Парсинг JSON
	fmt.Fprintf(os.Stderr, "📖 Читання JSON: %s\n", inputFile)
	twJSON, err := ParseJSONFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Помилка парсингу: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "✅ JSON успішно прочитано (версія: %d, квести: %d)\n", 
		twJSON.Version, len(twJSON.Quests))

	// Побудова графу станів
	fmt.Fprintf(os.Stderr, "🔨 Побудова графу станів...\n")
	gameState, err := BuildGameState(twJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Помилка побудови графу: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "✅ Граф побудовано:\n")
	fmt.Fprintf(os.Stderr, "   - Кімнат: %d\n", len(gameState.Rooms))
	fmt.Fprintf(os.Stderr, "   - Дверей: %d\n", len(gameState.Doors))
	fmt.Fprintf(os.Stderr, "   - Контейнерів: %d\n", len(gameState.Containers))
	fmt.Fprintf(os.Stderr, "   - Поверхонь: %d\n", len(gameState.Supporters))
	fmt.Fprintf(os.Stderr, "   - Предметів: %d\n", len(gameState.Items))
	fmt.Fprintf(os.Stderr, "   - Дій у квестах: %d\n", countTotalActions(gameState.Quests))

	// Конвертація у .ta формат
	fmt.Fprintf(os.Stderr, "🔄 Конвертація у формат .ta...\n")
	taContent, err := ConvertToTA(gameState, challengeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Помилка конвертації: %v\n", err)
		os.Exit(1)
	}

	// Запис .ta файлу
	fmt.Fprintf(os.Stderr, "💾 Запис у файл: %s\n", outputFile)
	if err := os.WriteFile(outputFile, []byte(taContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Помилка запису файлу: %v\n", err)
		os.Exit(1)
	}

	// Копіювання game_state.sh
	if *copyGameState {
		if err := copyGameStateFile(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Попередження: не вдалося скопіювати game_state.sh: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "✅ game_state.sh скопійовано до поточної директорії\n")
		}
	}

	// Фінальне повідомлення
	fmt.Fprintf(os.Stderr, "\n✅ Конвертацію завершено!\n")
	fmt.Fprintf(os.Stderr, "\n📋 Результат:\n")
	fmt.Fprintf(os.Stderr, "   Файл квесту: %s\n", outputFile)
	fmt.Fprintf(os.Stderr, "   Назва челенджу: %s\n", challengeName)
	
	if *copyGameState {
		fmt.Fprintf(os.Stderr, "\n🔧 Для запуску:\n")
		fmt.Fprintf(os.Stderr, "   1. Зробіть game_state.sh виконуваним: chmod +x game_state.sh\n")
		fmt.Fprintf(os.Stderr, "   2. Додайте game_state.sh до PATH або поточної директорії\n")
		fmt.Fprintf(os.Stderr, "   3. Запустіть: ./termadventure --print %s\n", outputFile)
	}
	
	fmt.Fprintf(os.Stderr, "\n🎮 Гарної гри!\n")
}

// countTotalActions - підраховує загальну кількість дій у квестах
func countTotalActions(quests []QuestActions) int {
	count := 0
	for _, quest := range quests {
		count += len(quest.Actions)
	}
	return count
}

// copyGameStateFile - копіює game_state.sh до поточної директорії
func copyGameStateFile() error {
	// Шукаємо game_state.sh у директорії виконуваного файлу
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("не вдалося визначити шлях: %w", err)
	}
	
	execDir := filepath.Dir(execPath)
	sourcePath := filepath.Join(execDir, "game_state.sh")
	
	// Якщо не знайдено, шукаємо у поточній директорії
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		sourcePath = "game_state.sh"
	}
	
	// Перевіряємо що файл існує
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("game_state.sh не знайдено")
	}
	
	// Читаємо та записуємо
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("помилка читання: %w", err)
	}
	
	destPath := filepath.Join(".", "game_state.sh")
	if err := os.WriteFile(destPath, content, 0755); err != nil {
		return fmt.Errorf("помилка запису: %w", err)
	}
	
	return nil
}
