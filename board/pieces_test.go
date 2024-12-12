package board

import "testing"

func TestGetPiece(t *testing.T) {
	for _, test := range getTestData() {
		t.Run(string(test.rune), func(t *testing.T) {
			result, err := GetPiece(test.rune)

			if err != nil {
				t.Errorf("encountered unexpected error: %v", err)
			}

			if result != test.piece {
				t.Errorf("expected %x, got %x", test.piece, result)
			}
		})
	}
}

func TestGetRune(t *testing.T) {
	for _, test := range getTestData() {
		t.Run(string(test.rune), func(t *testing.T) {
			rune := test.piece.GetRune()
			if rune != test.rune {
				t.Errorf("expected %x to be %c, got %c",
					test.piece, test.rune, rune)
			}
		})
	}
}

func getTestData() []struct {
	rune  rune
	piece Piece
} {
	return []struct {
		rune  rune
		piece Piece
	}{
		{rune: 'K', piece: King | White},
		{rune: 'Q', piece: Queen | White},
		{rune: 'R', piece: Rook | White},
		{rune: 'B', piece: Bishop | White},
		{rune: 'N', piece: Knight | White},
		{rune: 'P', piece: Pawn | White},
		{rune: 'k', piece: King | Black},
		{rune: 'q', piece: Queen | Black},
		{rune: 'r', piece: Rook | Black},
		{rune: 'b', piece: Bishop | Black},
		{rune: 'n', piece: Knight | Black},
		{rune: 'p', piece: Pawn | Black},
	}
}