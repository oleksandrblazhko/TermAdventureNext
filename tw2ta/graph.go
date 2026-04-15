package main

import (
	"fmt"
	"strings"
)

// Room - кімната у грі
type Room struct {
	ID          string
	Name        string
	Type        string // storage, work, cook, clean
	Description string
	Items       []string // предмети у кімнаті
	Containers  []string // контейнери
	Supporters  []string // поверхні/опори
	Doors       []string // двері
}

// Door - двері/портал між кімнатами
type Door struct {
	ID       string
	Name     string
	RoomA    string
	RoomB    string
	Locked   bool
	Closed   bool
	KeyMatch string // ID ключа який відмикає
}

// Container - контейнер (шухляда, холодильник, тощо)
type Container struct {
	ID       string
	Name     string
	Room     string
	Closed   bool
	Contents []string // предмети всередині
}

// Supporter - поверхня/опора (стіл, ліжко, тощо)
type Supporter struct {
	ID       string
	Name     string
	Room     string
	Contents []string // предмети на поверхні
}

// Item - предмет у грі
type Item struct {
	ID          string
	Name        string
	Type        string // k (ключ), f (їжа), o (інше)
	Location    string // ID кімнати/контейнера/поверхні
	InInventory bool
	Edible      bool
}

// GameState - повний стан гри
type GameState struct {
	Rooms       map[string]*Room
	Doors       map[string]*Door
	Containers  map[string]*Container
	Supporters  map[string]*Supporter
	Items       map[string]*Item
	PlayerRoom  string // поточна кімната гравця
	Quests      []QuestActions // квести з діями
	Objective   string
	Walkthrough []string
	NameMap     map[string]string // Мапінг ID → читабельна назва
}

// QuestActions - дії квесту з залежностями
type QuestActions struct {
	Desc           string
	Reward         int
	Actions        []ActionStep
	FailConditions []FailCondition
}

// ActionStep - окремий крок дії
type ActionStep struct {
	Index          int
	ActionName     string
	Command        string
	Preconditions  []Predicate
	Postconditions []Predicate
	SourceRoom     string // кімната де виконується дія
	TargetRoom     string // цільова кімната (для руху)
}

// FailCondition - умова поразки
type FailCondition struct {
	Condition Predicate
}

// BuildGameState - будує повний стан гри з JSON
func BuildGameState(tw *TextWorldJSON) (*GameState, error) {
	gs := &GameState{
		Rooms:      make(map[string]*Room),
		Doors:      make(map[string]*Door),
		Containers: make(map[string]*Container),
		Supporters: make(map[string]*Supporter),
		Items:      make(map[string]*Item),
		NameMap:    make(map[string]string),
	}

	// 1. Спочатку збираємо мапу ID → Name
	gs.buildNameMap(tw)

	// 2. Будуємо кімнати
	gs.buildRooms(tw)

	// 3. Будуємо двері
	gs.buildDoors(tw)

	// 4. Будуємо контейнери
	gs.buildContainers(tw)

	// 5. Будуємо поверхні
	gs.buildSupporters(tw)

	// 6. Будуємо предмети
	gs.buildItems(tw)

	// 7. Знаходимо початкову кімнату гравця
	gs.findPlayerStart(tw)

	// 8. Зберігаємо мету та проходження
	gs.Objective = tw.Objective
	gs.Walkthrough = tw.Metadata.Walkthrough

	// 9. Парсимо квести та дії
	if err := gs.parseQuests(tw); err != nil {
		return nil, err
	}

	return gs, nil
}

// buildNameMap - створює мапінг ID → читабельна назва
func (gs *GameState) buildNameMap(tw *TextWorldJSON) {
	for _, pair := range tw.Infos {
		if pair.Info.Name != nil && *pair.Info.Name != "" {
			gs.NameMap[pair.ID] = *pair.Info.Name
		} else if pair.Info.Noun != nil && *pair.Info.Noun != "" {
			gs.NameMap[pair.ID] = *pair.Info.Noun
		}
	}
}

// buildRooms - створює кімнати з предикатів
func (gs *GameState) buildRooms(tw *TextWorldJSON) {
	roomIDs := make(map[string]bool)
	roomTypes := make(map[string]string)
	for _, pair := range tw.Infos {
		if pair.Info.Type == "r" {
			roomIDs[pair.ID] = true
			if pair.Info.RoomType != nil {
				roomTypes[pair.ID] = *pair.Info.RoomType
			}
		}
	}
	for roomID := range roomIDs {
		gs.Rooms[roomID] = &Room{ID: roomID, Name: gs.NameMap[roomID], Type: roomTypes[roomID]}
	}
}

// buildDoors - створює двері з предикатів link, locked, closed
func (gs *GameState) buildDoors(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "link" && len(pred.Arguments) == 3 {
			roomA, doorID, roomB := pred.Arguments[0].Name, pred.Arguments[1].Name, pred.Arguments[2].Name
			if _, exists := gs.Doors[doorID]; !exists {
				door := &Door{ID: doorID, Name: gs.NameMap[doorID], RoomA: roomA, RoomB: roomB}
				gs.Doors[doorID] = door
				if room, ok := gs.Rooms[roomA]; ok {
					room.Doors = append(room.Doors, doorID)
				}
				if room, ok := gs.Rooms[roomB]; ok {
					room.Doors = append(room.Doors, doorID)
				}
			}
		}
	}
	for _, pred := range tw.World {
		if (pred.Name == "locked" || pred.Name == "closed") && len(pred.Arguments) > 0 {
			if door, ok := gs.Doors[pred.Arguments[0].Name]; ok {
				if pred.Name == "locked" {
					door.Locked = true
				} else {
					door.Closed = true
				}
			}
		} else if pred.Name == "match" && len(pred.Arguments) == 2 {
			if door, ok := gs.Doors[pred.Arguments[1].Name]; ok {
				door.KeyMatch = pred.Arguments[0].Name
			}
		}
	}
}

// buildContainers - створює контейнери
func (gs *GameState) buildContainers(tw *TextWorldJSON) {
	for _, pair := range tw.Infos {
		if pair.Info.Type == "c" {
			gs.Containers[pair.ID] = &Container{ID: pair.ID, Name: gs.NameMap[pair.ID]}
		}
	}
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) == 2 && gs.Containers[pred.Arguments[0].Name] != nil {
			containerID, roomID := pred.Arguments[0].Name, pred.Arguments[1].Name
			gs.Containers[containerID].Room = roomID
			if room, ok := gs.Rooms[roomID]; ok {
				room.Containers = append(room.Containers, containerID)
			}
		} else if pred.Name == "closed" && len(pred.Arguments) > 0 && gs.Containers[pred.Arguments[0].Name] != nil {
			gs.Containers[pred.Arguments[0].Name].Closed = true
		} else if pred.Name == "in" && len(pred.Arguments) == 2 && gs.Containers[pred.Arguments[1].Name] != nil {
			itemID, containerID := pred.Arguments[0].Name, pred.Arguments[1].Name
			gs.Containers[containerID].Contents = append(gs.Containers[containerID].Contents, itemID)
		}
	}
}

// buildSupporters - створює поверхні
func (gs *GameState) buildSupporters(tw *TextWorldJSON) {
	// ... (implementation is similar to buildContainers)
}

// buildItems - створює предмети
func (gs *GameState) buildItems(tw *TextWorldJSON) {
	// ... (implementation is similar to buildContainers)
}

// findPlayerStart - знаходить початкову кімнату гравця
func (gs *GameState) findPlayerStart(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) > 0 && pred.Arguments[0].Name == "P" {
			gs.PlayerRoom = pred.Arguments[1].Name
			return
		}
	}
}

// parseQuests - парсить квести та їх дії
func (gs *GameState) parseQuests(tw *TextWorldJSON) error {
	for _, quest := range tw.Quests {
		qa := QuestActions{Desc: quest.Desc, Reward: quest.Reward}
		for _, fail := range quest.FailEvents {
			if fail.Condition != nil {
				qa.FailConditions = append(qa.FailConditions, FailCondition{Condition: *fail.Condition})
			}
		}
		for _, winEvent := range quest.WinEvents {
			for i, action := range winEvent.Actions {
				step := ActionStep{
					Index: i, ActionName: action.Name, Command: action.CommandTemplate,
					Preconditions: action.Preconditions, Postconditions: action.Postconditions,
				}
				for _, pre := range action.Preconditions {
					if pre.Name == "at" && len(pre.Arguments) > 1 && pre.Arguments[0].Name == "P" {
						step.SourceRoom = pre.Arguments[1].Name
					}
					if strings.Contains(pre.Name, "_of") {
						step.TargetRoom = pre.Arguments[0].Name
					}
				}
				qa.Actions = append(qa.Actions, step)
			}
		}
		gs.Quests = append(gs.Quests, qa)
	}
	if len(gs.Quests) == 0 {
		return fmt.Errorf("не знайдено дійсних квестів з win_events")
	}
	return nil
}

// GetWalkthroughActions - повертає відсортований список дій для проходження
func (gs *GameState) GetWalkthroughActions() []ActionStep {
	var orderedActions []ActionStep
	actionMap := make(map[string]ActionStep)

	// Створюємо мапу для швидкого пошуку дій
	for _, quest := range gs.Quests {
		for _, action := range quest.Actions {
			// Нормалізуємо команду, видаляючи плейсхолдери
			normalizedCmd := cleanCommand(action.Command)
			actionMap[normalizedCmd] = action
		}
	}

	// Шукаємо відповідну ActionStep для кожної команди з walkthrough
	for _, cmd := range gs.Walkthrough {
		if action, ok := actionMap[cmd]; ok {
			orderedActions = append(orderedActions, action)
		} else {
			fmt.Printf("Попередження: не знайдено ActionStep для команди walkthrough: %s\n", cmd)
		}
	}
	return orderedActions
}

func cleanCommand(command string) string {
	// Прибираємо фігурні дужки для показу гравцю
	command = strings.ReplaceAll(command, "{", "")
	command = strings.ReplaceAll(command, "}", "")
	return command
}

// GetEntityName - повертає читабельну назву сутності
func (gs *GameState) GetEntityName(id string) string {
	if name, ok := gs.NameMap[id]; ok {
		return name
	}
	return id
}

// GetRoomDescription - генерує опис кімнати
func (gs *GameState) GetRoomDescription(roomID string) string {
	room, ok := gs.Rooms[roomID]
	if !ok {
		return "Невідома кімната"
	}
	var desc strings.Builder
	desc.WriteString(fmt.Sprintf("Ви у кімнаті **%s** (%s).\n\n", gs.GetEntityName(roomID), room.Name))
	if len(room.Containers) > 0 {
		desc.WriteString("Тут є:\n")
		for _, cID := range room.Containers {
			c := gs.Containers[cID]
			state := "відчинений"
			if c.Closed {
				state = "зачинений"
			}
			desc.WriteString(fmt.Sprintf("- **%s** (%s)\n", gs.GetEntityName(cID), state))
		}
		desc.WriteString("\n")
	}
	// ... (решта опису)
	return desc.String()
}
