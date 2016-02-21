package game

//© Copyright 2015-2016 Michał Krzysztof Feiler & Paweł Zacharek

import "fmt"

//MoatsState :  Black-White, White-Gray, Gray-Black  //true: bridged. Originally, true meant still active, i.e. non-bridged!!!
type MoatsState [3]bool

//var DEFMOATSSTATE = MoatsState{true, true, true}

//DEFMOATSSTATE , ie. nothig is bridged     const
var DEFMOATSSTATE = MoatsState{false, false, false} //are they bridged?

//Castling : White,Gray,Black King-side,Queen-side
type Castling [3][2]bool

func (cs Castling) Uint8() uint8 {
	var u uint8
	if cs[0][0] {
		u++
	}
	if cs[0][1] {
		u += 2
	}
	if cs[1][0] {
		u += 4
	}
	if cs[1][1] {
		u += 8
	}
	if cs[2][0] {
		u += 16
	}
	if cs[2][1] {
		u += 32
	}
	return u
}

func CastlingFromUint8(u uint8) Castling {
	var cs Castling
	cs[0][0] = u%2 == 1
	cs[0][1] = u>>1%2 == 1
	cs[1][0] = u>>2%2 == 1
	cs[1][1] = u>>3%2 == 1
	cs[2][0] = u>>4%2 == 1
	cs[2][1] = u>>5 == 1
	return cs
}

func forcastlingconv(c Color, b byte) (uint8, uint8) {
	var ct uint8
	switch b {
	case 'k', 'K':
		ct = 0
	case 'q', 'Q':
		ct = 1
	}
	return uint8(c) - 1, ct
}

//Give (color, K/Q)
func (cs Castling) Give(c Color, b byte) bool {
	col, ct := forcastlingconv(c, b)
	return cs[col][ct]
}

//Change (color, K/Q, bool)
func (cs Castling) Change(c Color, b byte, w bool) Castling {
	cso := cs
	col, ct := forcastlingconv(c, b)
	cso[col][ct] = w
	return cso
}

//OffRook : Rook can no longr castle
func (cs Castling) OffRook(c Color, b byte) Castling {
	return cs.Change(c, b, false)
}

//OffKing : No castling anymore for this player
func (cs Castling) OffKing(c Color) Castling {
	return cs.OffRook(c, 'K').OffRook(c, 'Q')
}

//EnPassant type : two positions of enpassant, moving one left on each move
type EnPassant [2]Pos

//Appeared : new EnPassant possibility
func (e EnPassant) Appeared(p Pos) EnPassant {
	ep := e
	ep[0] = ep[1]
	ep[1] = p
	return ep
}

//Nothing : just a move, no new enpassant possibility
func (e EnPassant) Nothing() EnPassant {
	return EnPassant{e[1], Pos{127, 127}}
}

//HalfmoveClock : not used atm, TODO
type HalfmoveClock uint8

//FullmoveNumber : not used atm, TODO
type FullmoveNumber uint16

//PlayersAlive : which players are still active
type PlayersAlive [3]bool

//Give : tell if a player is active by color
func (pa PlayersAlive) Give(who Color) bool {
	return pa[who-1]
}

//Die : disactivate a player
func (pa PlayersAlive) Die(who Color) PlayersAlive {
	pan := pa
	pan[who-1] = false
	return pan
}

//ListEm is simplified Subc2's Winner(*State) from e396e2b & 17685ad
func (pa PlayersAlive) ListEm() []Color {
	to := make([]Color, 0, 3)
	for _, j := range COLORS {
		if pa.Give(j) {
			to = append(to, j)
		}
	}
	return to
}

//DEFPLAYERSALIVE : true,true,true const
var DEFPLAYERSALIVE = [3]bool{true, true, true}

//State : single gamestate
type State struct {
	*Board //[color,figure_lowercase] //[0,0]
	MoatsState
	MovesNext Color //W G B
	Castling        //0W 1G 2B  //0K 1Q
	EnPassant       //[previousplayer,currentplayer]  [number,letter]
	HalfmoveClock
	FullmoveNumber
	PlayersAlive
}

type StateData struct {
	Board          []byte
	MoatZero       bool
	MoatOne        bool
	MoatTwo        bool
	MovesNext      int8
	Castling       uint8
	EnPasPrevRank  int8
	EnPasPrevFile  int8
	EnPasCurRank   int8
	EnPasCurFile   int8
	HalfmoveClock  int8
	FullmoveNumber int16
	AliveWhite     bool
	AliveGray      bool
	AliveBlack     bool
}

func (s *State) FromData(d *StateData) {
	s.Board = BoardByte(d.Board)
	s.MoatsState = MoatsState{d.MoatZero, d.MoatOne, d.MoatTwo}
	s.MovesNext = Color(d.MovesNext)
	s.Castling = CastlingFromUint8(d.Castling)
	s.EnPassant = EnPassant{{d.EnPasPrevRank, d.EnPasPrevFile}, {d.EnPasCurRank, d.EnPasCurFile}}
	s.HalfmoveClock = HalfmoveClock(d.HalfmoveClock)
	s.FullmoveNumber = FullmoveNumber(d.FullmoveNumber)
	s.PlayersAlive = PlayersAlive{d.AliveWhite, d.AliveGray, d.AliveBlack}
}

func (s *State) Data() *StateData {
	d := StateData{
		Board: s.Board.Byte(), MovesNext: int8(s.MovesNext),
		MoatZero: s.MoatsState[0], MoatOne: s.MoatsState[1], MoatTwo: s.MoatsState[2],
		Castling:      s.Castling.Uint8(),
		EnPasPrevRank: s.EnPassant[0][0], EnPasPrevFile: s.EnPassant[0][1],
		EnPasCurRank: s.EnPassant[1][0], EnPasCurFile: s.EnPassant[1][1],
		HalfmoveClock: int8(s.HalfmoveClock), FullmoveNumber: int16(s.FullmoveNumber),
		AliveWhite: s.PlayersAlive[0], AliveGray: s.PlayersAlive[1], AliveBlack: s.PlayersAlive[2],
	}
	return &d
}

//EvalDeath evaluates the death of whom is about to move next and returns the same pointer it got
func (s *State) EvalDeath() *State {
	if !(s.CanIMoveWOCheck(s.MovesNext)) {
		s.PlayersAlive.Die(s.MovesNext)
	}
	return s
}

func (s *State) String() string {
	return fmt.Sprintln("Board: ", (*s.Board), s.MoatsState, s.MovesNext, s.Castling, s.EnPassant, s.HalfmoveClock, s.FullmoveNumber, s.PlayersAlive)
}

//AnyPiece : if a piece could move (any piece, whatever stays there)
func (s *State) AnyPiece(from, to Pos) bool {
	return s.Board.AnyPiece(from, to, s.MoatsState, s.Castling, s.EnPassant)
}

//DEFENPASSANT : empty enpassant   const
var DEFENPASSANT = EnPassant{Pos{127, 127}, Pos{127, 127}}

//DEFCASTLING : everybody capable of castling everywhere  const
var DEFCASTLING = [3][2]bool{
	{true, true},
	{true, true},
	{true, true},
}

//FALSECASTLING : nobody can castle anymore   const
var FALSECASTLING = [3][2]bool{
	{false, false},
	{false, false},
	{false, false},
}

//NEWGAME : !!!LEGACY — use NewState() instead!!!  gamestate of a new game   const
var NEWGAME State

func init() { //initialize module pseudoconstants
	boardinit()
	NEWGAME = State{&BOARDFORNEWGAME, DEFMOATSSTATE, White, DEFCASTLING, DEFENPASSANT, HalfmoveClock(0), FullmoveNumber(1), DEFPLAYERSALIVE}
}

//NewState returns a newgame State
func NewState() State {
	nb := NewBoard()
	return State{&nb, DEFMOATSSTATE, White, DEFCASTLING, DEFENPASSANT, HalfmoveClock(0), FullmoveNumber(1), DEFPLAYERSALIVE}
}

//func (s *State) String() string {   // returns some kind of string that is also parsable
//}

//func ParsBoard3FEN([]byte) *[8][24][2]byte {
//}

//func Pars3FEN([]byte) *State {
//}
