package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// TextWorldJSON - коренева структура для JSON-файлу TextWorld
type TextWorldJSON struct {
	Version int            `json:"version"`
	World   []Predicate    `json:"world"`
	Grammar Grammar        `json:"grammar"`
	Quests  []Quest        `json:"quests"`
	Infos   []EntityInfoPair `json:"infos"`
}

// Predicate - предикат стану світу (at, closed, locked, тощо)
type Predicate struct {
	Name      string     `json:"name"`
	Arguments []Argument `json:"arguments"`
}

// Argument - аргумент предикату
type Argument struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Grammar - налаштування граматики гри
type Grammar struct {
	Theme                    string   `json:"theme"`
	NamesToExclude           []string `json:"names_to_exclude"`
	IncludeAdj               bool     `json:"include_adj"`
	BlendDescriptions        bool     `json:"blend_descriptions"`
	AmbiguousInstructions    bool     `json:"ambiguous_instructions"`
	OnlyLastAction           bool     `json:"only_last_action"`
	BlendInstructions        bool     `json:"blend_instructions"`
	AllowedVariablesNumbering bool    `json:"allowed_variables_numbering"`
	UniqueExpansion          bool     `json:"unique_expansion"`
}

// Quest - визначення квесту
type Quest struct {
	Desc       string     `json:"desc"`
	Reward     int        `json:"reward"`
	Commands   []string   `json:"commands"`
	WinEvents  []WinEvent `json:"win_events"`
	FailEvents []FailEvent `json:"fail_events"`
	Optional   bool       `json:"optional"`
	Repeatable bool       `json:"repeatable"`
}

// WinEvent - подія перемоги
type WinEvent struct {
	Commands []string `json:"commands"`
	Actions  []Action `json:"actions"`
}

// FailEvent - подія поразки
type FailEvent struct {
	Commands  []string    `json:"commands"`
	Condition *Predicate  `json:"condition"`
}

// Action - дія у квесті
type Action struct {
	Name              string      `json:"name"`
	Preconditions     []Predicate `json:"preconditions"`
	Postconditions    []Predicate `json:"postconditions"`
	CommandTemplate   string      `json:"command_template"`
	ReverseName       string      `json:"reverse_name"`
	ReverseCommandTpl string      `json:"reverse_command_template"`
}

// EntityInfoPair - пара ID + інформація про сутність
type EntityInfoPair struct {
	ID   string     `json:"id"`
	Info EntityInfo `json:"info"`
}

// EntityInfo - інформація про сутність
type EntityInfo struct {
	Type     string   `json:"type"`
	Name     *string  `json:"name"`
	Noun     *string  `json:"noun"`
	Adj      *string  `json:"adj"`
	Desc     *string  `json:"desc"`
	RoomType *string  `json:"room_type"`
	Definite *string  `json:"definite"`
	Indefinite *string `json:"indefinite"`
	Synonyms []string `json:"synonyms"`
}

// ParseJSONFile - читає та парсить JSON-файл TextWorld
func ParseJSONFile(filepath string) (*TextWorldJSON, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("помилка читання файлу %s: %w", filepath, err)
	}

	var twJSON TextWorldJSON
	if err := json.Unmarshal(data, &twJSON); err != nil {
		return nil, fmt.Errorf("помилка парсингу JSON: %w", err)
	}

	// Валідація обов'язкових полів
	if err := twJSON.Validate(); err != nil {
		return nil, err
	}

	return &twJSON, nil
}

// Validate - перевірка коректності структури
func (tw *TextWorldJSON) Validate() error {
	if tw.Version != 1 {
		return fmt.Errorf("непідтримувана версія JSON: %d (очікується 1)", tw.Version)
	}

	if len(tw.World) == 0 {
		return fmt.Errorf("світ порожній - відсутні предикати")
	}

	if len(tw.Quests) == 0 {
		return fmt.Errorf("квести відсутні")
	}

	if len(tw.Infos) == 0 {
		return fmt.Errorf("інформація про сутності відсутня")
	}

	// Перевірка що хоча б один квест має win_events
	hasWinEvent := false
	for _, quest := range tw.Quests {
		if len(quest.WinEvents) > 0 && len(quest.WinEvents[0].Actions) > 0 {
			hasWinEvent = true
			break
		}
	}

	if !hasWinEvent {
		return fmt.Errorf("жоден квест не має win_events")
	}

	return nil
}
