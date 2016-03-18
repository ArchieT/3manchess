package constsitval

//© Copyright 2015-2016 Michał Krzysztof Feiler & Paweł Zacharek

import "github.com/ArchieT/3manchess/game"
import "github.com/ArchieT/3manchess/simple"
import "github.com/ArchieT/3manchess/player"

//import "sync"
//import "sync/atomic"
import "fmt"
import "github.com/ArchieT/3manchess/ai"
import "encoding/json"

const MAXDEPTHCONSIDERED int8 = 8 // should be renamed to MINDEPTHNOTCONSIDERED

const DEFFIXDEPTH int8 = 2

const DEFOWN2THRTHD = 4.0

const WhoAmI string = "3manchess-ai_constsitval"

type AIPlayer struct {
	Name      string
	errchan   chan error
	ErrorChan chan<- error
	hurry     chan bool
	HurryChan chan<- bool
	Conf      AIConfig
	gp        *player.Gameplay
	waiting   bool
}

func (a *AIPlayer) Config() ai.Config {
	return a.Conf
}

type AIConfig struct {
	Depth             int8
	OwnedToThreatened float64
}

func (c AIConfig) Byte() []byte {
	o, e := json.Marshal(c)
	if e != nil {
		panic(e)
	}
	return o
}

func (c AIConfig) String() string {
	return string(c.Byte())
}

func (a *AIPlayer) Initialize(gp *player.Gameplay) {
	if a.Conf.Depth == 0 {
		a.Conf.Depth = DEFFIXDEPTH
	}
	if a.Conf.OwnedToThreatened == 0.0 {
		a.Conf.OwnedToThreatened = DEFOWN2THRTHD
	}
	errchan := make(chan error)
	a.errchan = errchan
	a.ErrorChan = errchan
	hurry := make(chan bool)
	a.hurry = hurry
	a.HurryChan = hurry
	a.gp = gp

	go func() {
		for b := range a.errchan {
			panic(b)
		}
	}()
}

func (a *AIPlayer) HurryChannel() chan<- bool {
	return a.HurryChan
}

func (a *AIPlayer) ErrorChannel() chan<- error {
	return a.ErrorChan
}

//Worker is the routine behind the Think; exported just in case
func (a *AIPlayer) Worker(s *game.State, whoarewe game.Color, depth int8) []float64 {
	minmax_slice := make([]float64, depth+1)
	if depth < 0 { // negative depth may be considered error in the future
		minmax_slice[0] = a.SitValue(s)
		return minmax_slice
	}
	var mythoughts map[int]*[MAXDEPTHCONSIDERED]float64
	var index int
	index = 0
	newthought := make([]float64, MAXDEPTHCONSIDERED)
	var bestthought [MAXDEPTHCONSIDERED]float64
	var move_to_apply game.Move
	var bestsitval float64
	for state := range game.ASAOMGen(s, whoarewe) {
		mythoughts[index][0] = a.SitValue(state)
		if int(depth) > 0 {
			bestsitval = -1000000
			for mymove := range game.VFTPGen(state) {
				move_to_apply = game.Move{mymove.FromTo[0], mymove.FromTo[1], state, mymove.PawnPromotion}
				newstate, _ := move_to_apply.After()
				newthought := append([]float64{mythoughts[index][0]}, a.Worker(newstate, s.MovesNext, depth)...)
				if newthought[a.Conf.Depth-1] > bestsitval {
					bestsitval = minmax_slice[depth-1]
					bestthought[0] = a.SitValue(newstate)
					for i := 1; i <= int(depth); i++ {
						mythoughts[index][i] = newthought[i]
					}
				}
			}
			for i := 1; i <= int(depth); i++ {
				mythoughts[index][i] = bestthought[i]
			}
		}
		index++
	}
	bestsitval = 1000000
	minmax_slice[0] = mythoughts[0][0] // we assume game hadn't finished so far
	for i := 0; i < index; i++ {
		if newthought[depth-1] < bestsitval { // we need to find the best opponents' moves to test our strategy
			for j := i; j <= int(depth); j++ {
				minmax_slice[j] = mythoughts[i][j]
			}
		}
	}
	return minmax_slice
}

//Think is the function generating the Move
func (a *AIPlayer) Think(s *game.State, hurry <-chan bool) game.Move {
	hurryup := simple.MergeBool(hurry, a.hurry)
	for i := len(hurryup); i > 0; i-- {
		<-hurryup
	}
	var thoughts map[game.FromToProm]*[MAXDEPTHCONSIDERED]float64 // so "bloated" for future use of hurry channel
	var bestmove game.FromToProm
	var bestsitval float64
	bestsitval = -1000000
	for move := range game.VFTPGen(s) {
		move_to_apply := game.Move{move.FromTo[0], move.FromTo[1], s, move.PawnPromotion}
		newstate, _ := move_to_apply.After()
		minmax_slice := a.Worker(newstate, s.MovesNext, a.Conf.Depth)
		for i, sitval := range minmax_slice {
			thoughts[move][i] = sitval
		}
		if thoughts[move][a.Conf.Depth] > bestsitval {
			bestmove = move
			bestsitval = thoughts[move][a.Conf.Depth]
		}
	}
	return game.Move{bestmove.FromTo[0], bestmove.FromTo[1], s, bestmove.PawnPromotion}
}

func (a *AIPlayer) HeyItsYourMove(s *game.State, hurryup <-chan bool) game.Move {
	return a.Think(s, hurryup)
}

func (a *AIPlayer) HeySituationChanges(_ *game.Move, _ *game.State) {}
func (a *AIPlayer) HeyYouLost(_ *game.State)                        {}
func (a *AIPlayer) HeyYouWon(_ *game.State)                         {}
func (a *AIPlayer) HeyYouDrew(_ *game.State)                        {}

func (a *AIPlayer) String() string {
	return fmt.Sprintf("%v%v", "SVBotDepth", a.Conf.Depth) //TODO: print whoami and conf
}

func (a *AIPlayer) AreWeWaitingForYou() bool {
	return a.waiting
}

func (a *AIPlayer) HeyWeWaitingForYou(b bool) {
	a.waiting = b
}