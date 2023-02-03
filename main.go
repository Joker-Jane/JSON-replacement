package main

import (
	"github.com/Joker-Jane/JSON-replacement/json_select"
)

func main() {
	/*
		cfg := json_replace.NewConfigFromConsole()
		replace := json_replace.NewJSONReplace(cfg)
		replace.Exec()
	*/

	cfg := json_select.NewConfigFromConsole()
	s := json_select.NewJSONSelect(cfg)
	s.Exec()
}
