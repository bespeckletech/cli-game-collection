package main

import (
	"bored/games"
	"bored/games/snake"
	"bored/games/tetris"
	"math/rand"
	"time"
)

func main() {
	min := 0
	max := 4
	var gameselection int
	rand.Seed(time.Now().UnixNano())
	gameselection = rand.Intn(max-min) + min
	print(gameselection)
	// return
	switch gameselection {
	case 0:
		games.Tictactoe()
		break
	case 1:
		snake.NewGame().Start()
		break
	case 2:
		// pacman.Start()
		println("Under process")
		break
	case 3:
		tetris.Start()
		break
	default:
		println("Bad luck no games available")
	}

}
