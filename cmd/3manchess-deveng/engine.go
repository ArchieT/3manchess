package main

import (
	"fmt"
	"github.com/ArchieT/3manchess/game"
	"github.com/ArchieT/3manchess/interface/deveng"
	"github.com/ArchieT/3manchess/player"
	//	"os"
)

func main() {
	fmt.Println("3manchess experimental engine")
	var players map[game.Color]*deveng.Developer
	players[game.White].Name = "Whitey"
	players[game.Gray].Name = "Greyey"
	players[game.Black].Name = "Blackey"
	proceed := make(chan bool)
	var interplayers map[game.Color]player.Player
	for _, ci := range game.COLORS {
		interplayers[ci] = player.Player(players[ci])
	}
	player.NewGame(interplayers, proceed)
}
