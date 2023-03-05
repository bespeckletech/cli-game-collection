package main

import (
	"bored/games"
	"bored/games/pacman"
	"bored/games/snake"
	"bored/games/tetris"
	"math/rand"
	"time"
)

func main() {
	min := 1
	max := 5
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
		pacman.Start()
		break
	case 3:
		tetris.Start()
		break
	default:
		println("Bad luck no games available")
	}

}
