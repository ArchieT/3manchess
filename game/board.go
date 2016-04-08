package game

//© Copyright 2015-2016 Michał Krzysztof Feiler & Paweł Zacharek

import "fmt"

//FigType : type of a figure, piece type
type FigType uint8

//const FigTypes
const (
	ZeroFigType FigType = iota
	Rook
	Knight
	Bishop
	Queen
	King
	Pawn
)

func (f *Fig) String() string {
	r := f.Rune()
	otoczka := "."
	if f.PawnCenter {
		otoczka = "!"
	}
	return otoczka + string(r)
}

//PawnCenter : whether the pawn had already passed through the center
type PawnCenter bool

//Byte returns "Y" or "N"
func (pc PawnCenter) Byte() byte {
	return ynbool(bool(pc))
}

//Fig : a struct describing a single piece: it's type, it's color, and, in case of a pawn, whether is had already passed through the center
type Fig struct {
	FigType
	Color
	PawnCenter
}

//Square : a struct describing a single square: whether it is empty, and what is on it
type Square struct {
	Fig
	NotEmpty bool
}

//Empty : return !s.NotEmpty
func (s Square) Empty() bool {
	return !s.NotEmpty
}

//Color : legacy : return s.Fig.Color
func (s Square) Color() Color {
	return s.Fig.Color
}

//What : legacy : return s.Fig.FigType
func (s Square) What() FigType {
	return s.Fig.FigType
}

//EMPTYOURSTR is the string that is the value of Square.String() if Square.Empty()
var EMPTYOURSTR = "__"

func (s Square) String() string {
	if s.NotEmpty {
		//ourbyte = []byte{byte(s.Fig.Color), byte(s.Fig.FigType), s.Fig.PawnCenter.Byte()}
		return s.Fig.String()
	}
	return EMPTYOURSTR
}

func (b *Board) String() string {
	var s string = "\n"
	for i := 0; i < 6; i++ {
		s += fmt.Sprintln((*b)[i])
	}
	return s
}

//Pos : coordinates
type Pos [2]int8

//Pos.String : give a nice [0,0] string
func (p Pos) String() string {
	return fmt.Sprintf("[%v,%v]", p[0], p[1])
}

//Board : board array
type Board [6][24]Square

//GPos gets the position's square pointer
func (b *Board) GPos(p Pos) *Square {
	if err := p.Correct(); err != nil {
		panic(err)
	}
	return &((*b)[p[0]][p[1]])
}

//Color : color type
type Color uint8

//Next returns the next color: White, Gray, Black,  White, etc.
func (c Color) Next() Color {
	return Color(c%3 + 1)
}

//String returns string "White"/"Gray"/"Black"
func (c Color) String() string {
	switch c {
	case White:
		return "White"
	case Gray:
		return "Gray"
	case Black:
		return "Black"
	default:
		panic(uint8(c))
	}
}

//Color constants
const (
	ZeroColor Color = iota
	White
	Gray
	Black
)

//COLORS : array of colors, ordered	const
var COLORS = [3]Color{White, Gray, Black}

//FIRSTRANKNEWGAME : first rank, from 0mod8 to 7mod8
var FIRSTRANKNEWGAME = [8]FigType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}

//ALLPOS : all valid positions
var ALLPOS [6 * 24]Pos

func allposinit() { //initialize ALLPOS
	for y := 0; y < 6; y++ {
		for x := 0; x < 24; x++ {
			ALLPOS[y*24 + x] = Pos{int8(y), int8(x)}
		}
	}
}

//BOARDFORNEWGAME — a newgame board
var BOARDFORNEWGAME Board //newgame board

func boardinit() { //initialize BOARDFORNEWGAME module pseudoconstant
	for ci, c := range COLORS {
		for fi, f := range FIRSTRANKNEWGAME {
			a := ci*8 + fi
			BOARDFORNEWGAME[0][a].FigType = f
			BOARDFORNEWGAME[0][a].Fig.Color = c
			BOARDFORNEWGAME[0][a].NotEmpty = true
			BOARDFORNEWGAME[1][a].Fig.Color = c
			BOARDFORNEWGAME[1][a].Fig.FigType = Pawn
			BOARDFORNEWGAME[1][a].NotEmpty = true
			for l := 2; l < 6; l++ {
				BOARDFORNEWGAME[l][a].Fig.Color = 0
				BOARDFORNEWGAME[l][a].Fig.FigType = 0
				BOARDFORNEWGAME[l][a].NotEmpty = false
			}
		}
	}
}

//NewBoard is a replacement for NEWBOARD
func NewBoard() Board {
	return BOARDFORNEWGAME
}
