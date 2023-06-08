package main

import (
	"github.com/Joker-Jane/JSON-replacement/json_flat"
)

func main() {
	cfg := json_flat.NewConfigFromConsole()
	s := json_flat.NewJSONFlat(cfg)
	s.Exec()
}
