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

//Uint8 returns an uint8 repr of a Castling struct
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

//Array returns a [6]bool repr of a Castling struct
func (cs Castling) Array() [6]bool {
	var b [6]bool
	for i := 0; i < 6; i++ {
		b[i] = cs[i>>1][i%2]
	}
	return b
}

//CastlingFromUint8 reproduces Castling from uint8 repr
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

//CastlingFromArray reproduces Castling from [6]bool repr
func CastlingFromArray(b [6]bool) Castling {
	var cs Castling
	for i := 0; i < 6; i++ {
		cs[i>>1][i%2] = b[i]
	}
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
func (pa *PlayersAlive) Die(who Color) {
	pa[who-1] = false
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
	*Board         `json:"board"`      //[color,figure_lowercase] //[0,0]
	MoatsState     `json:"moatsstate"` //moatsstate
	MovesNext      Color               `json:"movesnext"` //W G B
	Castling       `json:"castling"`   //0W 1G 2B  //0K 1Q
	EnPassant      `json:"enpassant"`  //[previousplayer,currentplayer]  [number,letter]
	HalfmoveClock  `json:"halfmoveclock"`
	FullmoveNumber `json:"fullmovenumber"`
	PlayersAlive   `json:"alivecolors"`
}

func (s *State) Equal(d *State) bool {
	return *s.Board == *d.Board && s.MoatsState == d.MoatsState && s.MovesNext == d.MovesNext &&
		s.Castling == d.Castling && s.EnPassant == d.EnPassant && s.HalfmoveClock == d.HalfmoveClock &&
		s.FullmoveNumber == d.FullmoveNumber && s.PlayersAlive == d.PlayersAlive
}

//StateData is a repr of State for db storage
type StateData struct {
	Board          [144]byte `json:"boardrepr"`
	Moats          [3]bool   `json:"moatsstate"`
	MovesNext      int8      `json:"movesnext"`
	Castling       [6]bool   `json:"castling"`
	EnPassant      [4]int8   `json:"enpassant"`
	HalfmoveClock  int8      `json:"halfmoveclock"`
	FullmoveNumber int16     `json:"fullmovenumber"`
	Alive          [3]bool   `json:"alivecolors"`
}

//FromData pulls data from StateData into State
func (s *State) FromData(d *StateData) {
	s.Board = BoardByte(d.Board[:])
	s.MoatsState = MoatsState(d.Moats)
	s.MovesNext = Color(d.MovesNext)
	s.Castling = CastlingFromArray(d.Castling)
	s.EnPassant = EnPassant{{d.EnPassant[0], d.EnPassant[1]}, {d.EnPassant[2], d.EnPassant[3]}}
	s.HalfmoveClock = HalfmoveClock(d.HalfmoveClock)
	s.FullmoveNumber = FullmoveNumber(d.FullmoveNumber)
	s.PlayersAlive = PlayersAlive(d.Alive)
}

//Data returns a StateData repr of a State
func (s *State) Data() *StateData {
	d := StateData{
		Board: s.Board.Byte(), MovesNext: int8(s.MovesNext),
		Moats:         [3]bool(s.MoatsState),
		Castling:      s.Castling.Array(),
		EnPassant:     [4]int8{s.EnPassant[0][0], s.EnPassant[0][1], s.EnPassant[1][0], s.EnPassant[1][1]},
		HalfmoveClock: int8(s.HalfmoveClock), FullmoveNumber: int16(s.FullmoveNumber),
		Alive: [3]bool(s.PlayersAlive),
	}
	return &d
}

func Byte144(s []byte) [144]byte {
	if len(s) != 144 {
		panic(s)
	}
	var d [144]byte
	for i := 0; i < 144; i++ {
		d[i] = s[i]
	}
	return d
}

//EvalDeath : evaluate the death of all players
func (s *State) EvalDeath() {
	if !(s.CanIMoveWOCheck(s.MovesNext)) { // next player to move cannot be checkmated
		s.PlayersAlive.Die(s.MovesNext)
	}
	for _, c := range COLORS { // all players must have theirs' kings
		if s.PlayersAlive.Give(c) && !s.Board.IsKingPresent(c) {
			s.PlayersAlive.Die(c)
		}
	}
}

func (s *State) String() string {
	return fmt.Sprintln(s.Board.String(), s.MoatsState, s.MovesNext, s.Castling, s.EnPassant, s.HalfmoveClock, s.FullmoveNumber, s.PlayersAlive)
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
	allposinit()
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
