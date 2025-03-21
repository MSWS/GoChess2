package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Single byte to store 2D coordinates,
// as the max value we need to store in a given
// dimension is 8.
// We store the X coordinate in the upper 4 bits,
// and the Y coordinate in the lower 4 bits
type Coordinate byte

func (coord Coordinate) String() string {
	return coord.GetAlgebra()
}

func (coord Coordinate) GetCoords() (byte, byte) {
	return byte(coord) >> 4, byte(coord) & 0b1111
}

func CreateCoordInt(row int, col int) Coordinate {
	return CreateCoordByte(byte(row), byte(col))
}

func CreateCoordByte(row byte, col byte) Coordinate {
	if row > 7 || col > 7 {
		panic(fmt.Errorf("invalid row, col (%v, %v)", row, col))
	}
	return Coordinate(row<<4 + col)
}

func CreateCoordAlgebra(alg string) Coordinate {
	col := alg[0] - 'a'
	row := alg[1] - '1'

	return CreateCoordByte(row, col)
}

func (coord Coordinate) GetAlgebra() string {
	row, col := coord.GetCoords()
	return fmt.Sprintf("%c%c", col+'a', row+'1')
}

type Castlability struct {
	CanQueenSide bool
	CanKingSide  bool
}

func (board Board) Get(coord Coordinate) Piece {
	row, col := coord.GetCoords()
	return board.Board[row][col]
}

func (board Board) GetStr(coord string) Piece {
	return board.Get(CreateCoordAlgebra(coord))
}

func (board Board) Set(coord Coordinate, piece Piece) {
	row, col := coord.GetCoords()
	board.Board[row][col] = piece
}

func (board Board) Move(from Coordinate, to Coordinate) {
	fromRow, fromCol := from.GetCoords()
	toRow, toCol := to.GetCoords()
	board.Board[toRow][toCol] = board.Board[fromRow][fromCol]
	board.Board[fromRow][fromCol] = 0
}

func (board *Board) MakeMove(move Move) Move {
	board.WhiteCastleHistory = append(board.WhiteCastleHistory, board.WhiteCastling)
	board.BlackCastleHistory = append(board.BlackCastleHistory, board.BlackCastling)
	captured := board.Get(move.to)
	move.capture = captured

	board.Move(move.from, move.to)

	row, _ := move.to.GetCoords()

	if move.piece.GetType() == Pawn && (row == 0 || row == 7) {
		if move.promotionTo == 0 {
			move.promotionTo = Queen | move.piece.GetColor()
		}
		board.Set(move.to, move.promotionTo|move.piece.GetColor())
	}

	castlability := &board.WhiteCastling
	if move.piece.GetColor() == Black {
		castlability = &board.BlackCastling
	}

	if move.piece.GetType() == Rook {
		_, col := move.from.GetCoords()
		if col == 0 {
			castlability.CanQueenSide = false
		} else if col == 7 {
			castlability.CanKingSide = false
		}
	}

	if move.piece.GetType() == King {
		castlability.CanKingSide = false
		castlability.CanQueenSide = false

		if move.IsCastle() {
			board.applyCastle(move)
		}
	}

	board.Active = (^board.Active).GetColor()
	board.Moves = append(board.Moves, move)
	return move
}

func (board *Board) UndoMove() {
	move := board.Moves[len(board.Moves)-1]
	board.Set(move.to, move.capture)
	board.Set(move.from, move.piece)

	castling := &board.WhiteCastling
	if move.piece.GetColor() == Black {
		castling = &board.BlackCastling
	}

	if move.IsCastle() {
		toRow, toCol := move.to.GetCoords()

		if toCol == 0 {
			castling.CanQueenSide = true
			board.Set(CreateCoordInt(int(toRow), 1), 0)
			board.Set(CreateCoordInt(int(toRow), 2), 0)
			board.Set(CreateCoordInt(int(toRow), 3), 0)
		} else {
			castling.CanKingSide = true
			board.Set(CreateCoordInt(int(toRow), 5), 0)
			board.Set(CreateCoordInt(int(toRow), 6), 0)
		}
	}

	board.Active = (^board.Active).GetColor()
	board.Moves = board.Moves[0 : len(board.Moves)-1]
	board.WhiteCastling = board.WhiteCastleHistory[len(board.WhiteCastleHistory)-1]
	board.BlackCastling = board.BlackCastleHistory[len(board.BlackCastleHistory)-1]

	board.WhiteCastleHistory = board.WhiteCastleHistory[0 : len(board.WhiteCastleHistory)-1]
	board.BlackCastleHistory = board.BlackCastleHistory[0 : len(board.BlackCastleHistory)-1]
}

func (board Board) applyCastle(move Move) {
	kingSquare := 2
	rookSquare := 3
	castleRow, castleCol := move.to.GetCoords()

	if castleCol == 7 {
		kingSquare = 6
		rookSquare = 5
	}

	board.Set(move.to, 0)
	board.Set(CreateCoordByte(castleRow, byte(kingSquare)), move.piece)
	board.Set(CreateCoordByte(castleRow, byte(rookSquare)), move.capture)
}

func (board *Board) MakeMoveStr(str string) {
	switch str {
	case "e4":
		board.MakeMove(board.CreateMoveStr("e2", "e4"))
	case "e5":
		board.MakeMove(board.CreateMoveStr("e7", "e5"))
	default:
		panic(fmt.Errorf("unsupported make move call with %v", str))
	}
}

// A board that represents a given game.
// Represents all data that is stored in a FEN string.
// i.e. given a Board, you can find the FEN, and vice-versa
type Board struct {
	// The game's pieces, in [row][col]
	// with the first row, first column being the bottom
	// left of the board from white's perspective (i.e. a1)
	Board *[8][8]Piece

	// Active player's turn
	Active Piece

	WhiteCastling Castlability
	BlackCastling Castlability
	EnPassant     *Coordinate
	HalfMoves     int
	FullMoves     int

	Moves              []Move
	WhiteCastleHistory []Castlability
	BlackCastleHistory []Castlability
}

type Bitboard uint64

func (board Board) Equal(other Board) bool {
	return board.ToFEN() == other.ToFEN()
}

func (board Board) String() string {
	return fmt.Sprintf("Board{%s}", board.ToFEN())
}

func (board Board) PrettySPrint() string {
	var builder strings.Builder

	for row := 0; row < len(board.Board); row++ {
		for col := 0; col < len(board.Board[row]); col++ {
			piece := board.Board[row][col]
			if piece == 0 {
				builder.WriteRune('0')
			} else {
				builder.WriteRune(rune(board.Board[row][col].GetRune()))
			}
		}
		builder.WriteRune('\n')
	}

	return builder.String()
}

func FromFEN(str string) (*Board, error) {
	result := Board{}

	records := strings.Split(strings.TrimSpace(str), " ")

	if len(records) != 6 {
		return nil, errors.New("malformed FEN string, expected 6 records")
	}

	board, err := GenerateBoard(records[0])

	if err != nil {
		return &result, err
	}

	result.Board = board

	switch records[1][0] {
	case 'w':
		result.Active = White
	case 'b':
		result.Active = Black
	default:
		return &result, fmt.Errorf("invalid active color %c", records[1][0])
	}

	result.WhiteCastling = getCastlability(records[2], White)
	result.BlackCastling = getCastlability(records[2], Black)

	if records[3][0] != '-' {
		coords := CreateCoordAlgebra(records[3])
		result.EnPassant = &coords
	}

	halfMoves, err := strconv.Atoi(records[4])

	if err != nil {
		return &result, err
	}

	result.HalfMoves = halfMoves

	fullMoves, err := strconv.Atoi(records[5])

	if err != nil {
		return &result, err
	}

	result.FullMoves = fullMoves

	return &result, nil
}

func (board Board) ToFEN() string {
	var result strings.Builder
	result.WriteString(GeneratePieceString(*board.Board))
	result.WriteRune(' ')
	if board.Active == White {
		result.WriteRune('w')
	} else {
		result.WriteRune('b')
	}

	result.WriteRune(' ')
	oldLen := result.Len()

	if board.WhiteCastling.CanKingSide {
		result.WriteRune('K')
	}

	if board.WhiteCastling.CanQueenSide {
		result.WriteRune('Q')
	}

	if board.BlackCastling.CanKingSide {
		result.WriteRune('k')
	}

	if board.BlackCastling.CanQueenSide {
		result.WriteRune('q')
	}

	if oldLen == result.Len() {
		result.WriteRune('-')
	}

	result.WriteRune(' ')

	if board.EnPassant == nil {
		result.WriteRune('-')
	} else {
		result.WriteString(board.EnPassant.GetAlgebra())
	}

	result.WriteRune(' ')
	result.WriteString(fmt.Sprint(board.HalfMoves))

	result.WriteRune(' ')
	result.WriteString(fmt.Sprint(board.FullMoves))

	return result.String()
}

func getCastlability(str string, color Piece) Castlability {
	if color == White {
		return Castlability{
			CanQueenSide: strings.Contains(str, "Q"),
			CanKingSide:  strings.Contains(str, "K"),
		}
	} else if color == Black {
		return Castlability{
			CanQueenSide: strings.Contains(str, "q"),
			CanKingSide:  strings.Contains(str, "k"),
		}
	}

	return Castlability{}
}
