package main

import (
	"github.com/Joker-Jane/JSON-replacement/json_replace"
)

func main() {
	cfg := json_replace.NewConfigFromConsole()
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}
