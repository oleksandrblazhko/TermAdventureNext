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
	ID       string
	Name     string
	Type     string // k (ключ), f (їжа), o (інше)
	Location string // ID кімнати/контейнера/поверхні
	InInventory bool
	Edible   bool
}

// GameState - повний стан гри
type GameState struct {
	Rooms      map[string]*Room
	Doors      map[string]*Door
	Containers map[string]*Container
	Supporters map[string]*Supporter
	Items      map[string]*Item
	
	PlayerRoom string // поточна кімната гравця
	Quests     []QuestActions // квести з діями
	
	// Мапінг ID → читабельна назва
	NameMap map[string]string
}

// QuestActions - дії квесту з залежностями
type QuestActions struct {
	Desc    string
	Reward  int
	Actions []ActionStep
	FailConditions []FailCondition
}

// ActionStep - окремий крок дії
type ActionStep struct {
	Index         int
	ActionName    string
	Command       string
	Preconditions []Predicate
	Postconditions []Predicate
	SourceRoom    string // кімната де виконується дія
	TargetRoom    string // цільова кімната (для руху)
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

	// 8. Парсимо квести та дії
	if err := gs.parseQuestes(tw); err != nil {
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
	// Знаходимо всі унікальні кімнати
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
		room := &Room{
			ID:   roomID,
			Name: gs.NameMap[roomID],
			Type: roomTypes[roomID],
		}
		gs.Rooms[roomID] = room
	}
}

// buildDoors - створює двері з предикатів link, locked, closed
func (gs *GameState) buildDoors(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "link" && len(pred.Arguments) == 3 {
			roomA := pred.Arguments[0].Name
			doorID := pred.Arguments[1].Name
			roomB := pred.Arguments[2].Name

			// Якщо двері вже існують, пропускаємо (зворотній link)
			if _, exists := gs.Doors[doorID]; exists {
				continue
			}

			door := &Door{
				ID:    doorID,
				Name:  gs.NameMap[doorID],
				RoomA: roomA,
				RoomB: roomB,
			}
			gs.Doors[doorID] = door

			// Додаємо двері до кімнат
			if roomA, ok := gs.Rooms[roomA]; ok {
				roomA.Doors = append(roomA.Doors, doorID)
			}
			if roomB, ok := gs.Rooms[roomB]; ok {
				roomB.Doors = append(roomB.Doors, doorID)
			}
		}
	}

	// Стан замків
	for _, pred := range tw.World {
		if pred.Name == "locked" && len(pred.Arguments) > 0 {
			doorID := pred.Arguments[0].Name
			if door, ok := gs.Doors[doorID]; ok {
				door.Locked = true
			}
		}
	}

	// Стан дверей (відчинені/зачинені)
	for _, pred := range tw.World {
		if pred.Name == "closed" && len(pred.Arguments) > 0 {
			argName := pred.Arguments[0].Name
			// Перевіряємо чи це двері (тип d)
			if _, isDoor := gs.Doors[argName]; isDoor {
				gs.Doors[argName].Closed = true
			}
		}
	}

	// Відповідності ключів
	for _, pred := range tw.World {
		if pred.Name == "match" && len(pred.Arguments) == 2 {
			keyID := pred.Arguments[0].Name
			doorID := pred.Arguments[1].Name
			if door, ok := gs.Doors[doorID]; ok {
				door.KeyMatch = keyID
			}
		}
	}
}

// buildContainers - створює контейнери
func (gs *GameState) buildContainers(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) == 2 {
			objID := pred.Arguments[0].Name
			roomID := pred.Arguments[1].Name

			// Перевіряємо чи це контейнер (тип c)
			for _, pair := range tw.Infos {
				if pair.ID == objID && pair.Info.Type == "c" {
					container := &Container{
						ID:   objID,
						Name: gs.NameMap[objID],
						Room: roomID,
					}
					gs.Containers[objID] = container

					if room, ok := gs.Rooms[roomID]; ok {
						room.Containers = append(room.Containers, objID)
					}
					break
				}
			}
		}
	}

	// Стан контейнерів
	for _, pred := range tw.World {
		if pred.Name == "closed" && len(pred.Arguments) > 0 {
			objID := pred.Arguments[0].Name
			if container, ok := gs.Containers[objID]; ok {
				container.Closed = true
			}
		}
	}

	// Вміст контейнерів
	for _, pred := range tw.World {
		if pred.Name == "in" && len(pred.Arguments) == 2 {
			itemID := pred.Arguments[0].Name
			containerID := pred.Arguments[1].Name
			if container, ok := gs.Containers[containerID]; ok {
				container.Contents = append(container.Contents, itemID)
			}
		}
	}
}

// buildSupporters - створює поверхні
func (gs *GameState) buildSupporters(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) == 2 {
			objID := pred.Arguments[0].Name
			roomID := pred.Arguments[1].Name

			for _, pair := range tw.Infos {
				if pair.ID == objID && pair.Info.Type == "s" {
					supporter := &Supporter{
						ID:   objID,
						Name: gs.NameMap[objID],
						Room: roomID,
					}
					gs.Supporters[objID] = supporter

					if room, ok := gs.Rooms[roomID]; ok {
						room.Supporters = append(room.Supporters, objID)
					}
					break
				}
			}
		}
	}

	// Предмети на поверхнях
	for _, pred := range tw.World {
		if pred.Name == "on" && len(pred.Arguments) == 2 {
			itemID := pred.Arguments[0].Name
			supporterID := pred.Arguments[1].Name
			if supporter, ok := gs.Supporters[supporterID]; ok {
				supporter.Contents = append(supporter.Contents, itemID)
			}
		}
	}
}

// buildItems - створює предмети
func (gs *GameState) buildItems(tw *TextWorldJSON) {
	for _, pair := range tw.Infos {
		// Пропускаємо кімнати (вже оброблені)
		if pair.Info.Type == "r" {
			continue
		}

		// Обробляємо тільки предмети (k, f, o)
		if pair.Info.Type == "k" || pair.Info.Type == "f" || pair.Info.Type == "o" {
			item := &Item{
				ID:   pair.ID,
				Name: gs.NameMap[pair.ID],
				Type: pair.Info.Type,
			}

			// Перевіряємо чи їстівний
			for _, pred := range tw.World {
				if pred.Name == "edible" && len(pred.Arguments) > 0 {
					if pred.Arguments[0].Name == pair.ID {
						item.Edible = true
					}
				}
			}

			gs.Items[pair.ID] = item
		}
	}

	// Визначаємо локацію кожного предмета
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) == 2 {
			itemID := pred.Arguments[0].Name
			roomID := pred.Arguments[1].Name
			if item, ok := gs.Items[itemID]; ok {
				item.Location = roomID
			}
		} else if pred.Name == "in" && len(pred.Arguments) == 2 {
			itemID := pred.Arguments[0].Name
			containerID := pred.Arguments[1].Name
			if item, ok := gs.Items[itemID]; ok {
				item.Location = "in:" + containerID
			}
		} else if pred.Name == "on" && len(pred.Arguments) == 2 {
			itemID := pred.Arguments[0].Name
			supporterID := pred.Arguments[1].Name
			if item, ok := gs.Items[itemID]; ok {
				item.Location = "on:" + supporterID
			}
		}
	}
}

// findPlayerStart - знаходить початкову кімнату гравця
func (gs *GameState) findPlayerStart(tw *TextWorldJSON) {
	for _, pred := range tw.World {
		if pred.Name == "at" && len(pred.Arguments) == 2 {
			if pred.Arguments[0].Name == "P" && pred.Arguments[0].Type == "P" {
				gs.PlayerRoom = pred.Arguments[1].Name
				return
			}
		}
	}
}

// parseQuestes - парсить квести та їх дії
func (gs *GameState) parseQuestes(tw *TextWorldJSON) error {
	for _, quest := range tw.Quests {
		qa := QuestActions{
			Desc:   quest.Desc,
			Reward: quest.Reward,
		}

		// Парсимо fail_conditions
		for _, fail := range quest.FailEvents {
			if fail.Condition != nil {
				qa.FailConditions = append(qa.FailConditions, FailCondition{
					Condition: *fail.Condition,
				})
			}
		}

		// Парсимо win_events
		for _, winEvent := range quest.WinEvents {
			for i, action := range winEvent.Actions {
				step := ActionStep{
					Index:          i,
					ActionName:     action.Name,
					Command:        action.CommandTemplate,
					Preconditions:  action.Preconditions,
					Postconditions: action.Postconditions,
				}

				// Визначаємо кімнату дії з precondition at(P, room)
				for _, pre := range action.Preconditions {
					if pre.Name == "at" && len(pre.Arguments) == 2 {
						if pre.Arguments[0].Name == "P" {
							step.SourceRoom = pre.Arguments[1].Name
						}
					}
					// Для руху - визначаємо цільову кімнату
					if strings.HasPrefix(pre.Name, "east_of") || strings.HasPrefix(pre.Name, "west_of") ||
						strings.HasPrefix(pre.Name, "north_of") || strings.HasPrefix(pre.Name, "south_of") {
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
	desc.WriteString(fmt.Sprintf("Ви у кімнаті **%s**", gs.GetEntityName(roomID)))
	
	if room.Name != "" {
		desc.WriteString(fmt.Sprintf(" (%s)", room.Name))
	}
	desc.WriteString(".\n\n")

	// Контейнери
	if len(room.Containers) > 0 {
		desc.WriteString("Тут є:\n")
		for _, cID := range room.Containers {
			container := gs.Containers[cID]
			state := "відчинений"
			if container.Closed {
				state = "зачинений"
			}
			desc.WriteString(fmt.Sprintf("- **%s** (%s)\n", gs.GetEntityName(cID), state))
		}
		desc.WriteString("\n")
	}

	// Поверхні
	if len(room.Supporters) > 0 {
		desc.WriteString("Поверхні:\n")
		for _, sID := range room.Supporters {
			supporter := gs.Supporters[sID]
			if len(supporter.Contents) > 0 {
				desc.WriteString(fmt.Sprintf("- **%s** (на ньому: %s)\n", 
					gs.GetEntityName(sID), 
					gs.GetEntityName(supporter.Contents[0])))
			} else {
				desc.WriteString(fmt.Sprintf("- **%s**\n", gs.GetEntityName(sID)))
			}
		}
		desc.WriteString("\n")
	}

	// Двері
	if len(room.Doors) > 0 {
		desc.WriteString("Двері:\n")
		for _, dID := range room.Doors {
			door := gs.Doors[dID]
			state := ""
			if door.Locked {
				state = " (замкнені)"
			} else if door.Closed {
				state = " (зачинені)"
			}
			desc.WriteString(fmt.Sprintf("- до %s через **%s**%s\n", 
				gs.GetEntityName(door.RoomA), 
				gs.GetEntityName(dID), 
				state))
		}
	}

	return desc.String()
}
