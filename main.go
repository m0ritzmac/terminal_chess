package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Player int

type Piece struct {
	Player   Player
	Type     PieceType
	Icon     string
	HasMoved bool // Track if piece has moved (for castling and pawn first move)
}

func (p *Piece) String() string {
	return p.Icon
}

type Board struct {
	squares   [8][8]*Piece
	lastMove  Move // Track last move for en passant
	moveCount int
	whiteKing Position
	blackKing Position
}

type Move struct {
	From        Position
	To          Position
	Piece       *Piece
	Captured    *Piece
	IsEnPassant bool
	IsCastling  bool
}

type Position struct {
	Row, Col int
}

const (
	White Player = iota
	Black
)

func (p Player) String() string {
	if p == White {
		return "White"
	}
	return "Black"
}

type PieceType byte

const (
	Pawn PieceType = iota
	Rook
	Knight
	Bishop
	Queen
	King
)

var pieceIcons = map[PieceType]string{
	Pawn:   "♙♟",
	Rook:   "♖♜",
	Knight: "♘♞",
	Bishop: "♗♝",
	Queen:  "♕♛",
	King:   "♔♚",
}

func NewPiece(pt PieceType, player Player) *Piece {
	icon := string([]rune(pieceIcons[pt])[player])
	return &Piece{Player: player, Type: pt, Icon: icon, HasMoved: false}
}

func NewBoard() *Board {
	b := &Board{}
	// Initialize pieces
	b.squares = [8][8]*Piece{
		{NewPiece(Rook, Black), NewPiece(Knight, Black), NewPiece(Bishop, Black), NewPiece(Queen, Black), NewPiece(King, Black), NewPiece(Bishop, Black), NewPiece(Knight, Black), NewPiece(Rook, Black)},
		{NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black), NewPiece(Pawn, Black)},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White), NewPiece(Pawn, White)},
		{NewPiece(Rook, White), NewPiece(Knight, White), NewPiece(Bishop, White), NewPiece(Queen, White), NewPiece(King, White), NewPiece(Bishop, White), NewPiece(Knight, White), NewPiece(Rook, White)},
	}
	// Store initial king positions
	b.whiteKing = Position{7, 4}
	b.blackKing = Position{0, 4}
	return b
}

func (b *Board) Draw() {
	fmt.Println("   a b c d e f g h")
	fmt.Println("  ─────────────────")
	for row := 0; row < 8; row++ {
		fmt.Printf("%d│ ", 8-row)
		for col := 0; col < 8; col++ {
			if b.squares[row][col] == nil {
				fmt.Print(". ")
			} else {
				fmt.Print(b.squares[row][col], " ")
			}
		}
		fmt.Printf("│%d\n", 8-row)
	}

	fmt.Println("  ─────────────────")
	fmt.Println("   a b c d e f g h")
}

func (b *Board) Move(oldPos, newPos Position, currentPlayer Player) error {
	piece := b.squares[oldPos.Row][oldPos.Col]
	if piece == nil {
		return fmt.Errorf("no piece at source position")
	}
	if piece.Player != currentPlayer {
		return fmt.Errorf("it's not your turn")
	}

	// Check if the move is valid
	move, err := b.ValidateMove(oldPos, newPos, currentPlayer)
	if err != nil {
		return err
	}

	// Make the move
	b.makeMove(move)

	// Check if the move puts the current player in check
	if b.IsInCheck(currentPlayer) {
		b.undoMove(move)
		return fmt.Errorf("move would leave king in check")
	}

	return nil
}

func (b *Board) ValidateMove(oldPos, newPos Position, currentPlayer Player) (Move, error) {
	piece := b.squares[oldPos.Row][oldPos.Col]
	move := Move{
		From:     oldPos,
		To:       newPos,
		Piece:    piece,
		Captured: b.squares[newPos.Row][newPos.Col],
	}

	// Basic validation
	if !isValidPosition(newPos) {
		return move, fmt.Errorf("destination position is outside the board")
	}

	if move.Captured != nil && move.Captured.Player == currentPlayer {
		return move, fmt.Errorf("cannot capture your own piece")
	}

	// Validate piece-specific movement
	if !b.IsValidPieceMove(piece, oldPos, newPos, &move) {
		return move, fmt.Errorf("invalid move for %s", piece)
	}

	return move, nil
}

func (b *Board) IsValidPieceMove(piece *Piece, oldPos, newPos Position, move *Move) bool {
	dr, dc := newPos.Row-oldPos.Row, newPos.Col-oldPos.Col

	switch piece.Type {
	case Pawn:
		return b.validatePawnMove(piece, oldPos, newPos, dr, dc, move)
	case Rook:
		return (dr == 0 || dc == 0) && b.isPathClear(oldPos, newPos)
	case Knight:
		return (abs(dr) == 2 && abs(dc) == 1) || (abs(dr) == 1 && abs(dc) == 2)
	case Bishop:
		return abs(dr) == abs(dc) && b.isPathClear(oldPos, newPos)
	case Queen:
		return (dr == 0 || dc == 0 || abs(dr) == abs(dc)) && b.isPathClear(oldPos, newPos)
	case King:
		if abs(dr) <= 1 && abs(dc) <= 1 {
			return true
		}
		// Check for castling
		return b.validateCastling(piece, oldPos, newPos, move)
	}
	return false
}

// Add this method to the Board struct implementation
func (b *Board) isPathClear(oldPos, newPos Position) bool {
	dr := sign(newPos.Row - oldPos.Row)
	dc := sign(newPos.Col - oldPos.Col)

	currentRow := oldPos.Row + dr
	currentCol := oldPos.Col + dc

	for currentRow != newPos.Row || currentCol != newPos.Col {
		if b.squares[currentRow][currentCol] != nil {
			return false
		}
		currentRow += dr
		currentCol += dc
	}

	return true
}

// Also modify the validatePawnMove method to remove the unused startRow variable
func (b *Board) validatePawnMove(piece *Piece, oldPos, newPos Position, dr, dc int, move *Move) bool {
	forward := -1
	if piece.Player == Black {
		forward = 1
	}

	// Normal forward move
	if dc == 0 && dr == forward && b.squares[newPos.Row][newPos.Col] == nil {
		return true
	}

	// First move - two squares
	if !piece.HasMoved && dc == 0 && dr == 2*forward &&
		b.squares[newPos.Row][newPos.Col] == nil &&
		b.squares[oldPos.Row+forward][oldPos.Col] == nil {
		return true
	}

	// Capture
	if dr == forward && abs(dc) == 1 {
		// Normal capture
		if b.squares[newPos.Row][newPos.Col] != nil {
			return true
		}
		// En passant
		if b.canEnPassant(oldPos, newPos, piece.Player) {
			move.IsEnPassant = true
			return true
		}
	}

	return false
}

func (b *Board) validateCastling(piece *Piece, oldPos, newPos Position, move *Move) bool {
	if piece.HasMoved {
		return false
	}

	// Check if it's a castling move
	if oldPos.Row != newPos.Row || abs(newPos.Col-oldPos.Col) != 2 {
		return false
	}

	row := oldPos.Row
	isKingSide := newPos.Col > oldPos.Col
	rookCol := 7
	if !isKingSide {
		rookCol = 0
	}

	// Check if rook is in place and hasn't moved
	rook := b.squares[row][rookCol]
	if rook == nil || rook.Type != Rook || rook.HasMoved {
		return false
	}

	// Check if path is clear
	startCol := min(oldPos.Col, rookCol) + 1
	endCol := max(oldPos.Col, rookCol)
	for col := startCol; col < endCol; col++ {
		if b.squares[row][col] != nil {
			return false
		}
	}

	// Check if king is not in check and doesn't pass through check
	if b.IsInCheck(piece.Player) {
		return false
	}

	// Check intermediate square
	intermediateCol := oldPos.Col + sign(newPos.Col-oldPos.Col)
	b.squares[row][intermediateCol] = piece
	b.squares[oldPos.Row][oldPos.Col] = nil
	inCheck := b.IsInCheck(piece.Player)
	b.squares[oldPos.Row][oldPos.Col] = piece
	b.squares[row][intermediateCol] = nil

	if inCheck {
		return false
	}

	move.IsCastling = true
	return true
}

func (b *Board) canEnPassant(oldPos, newPos Position, player Player) bool {
	if b.lastMove.Piece == nil || b.lastMove.Piece.Type != Pawn {
		return false
	}

	// Check if the last move was a two-square pawn advance
	if abs(b.lastMove.From.Row-b.lastMove.To.Row) != 2 {
		return false
	}

	// Check if the capturing pawn is on the correct rank
	correctRank := 3
	if player == Black {
		correctRank = 4
	}
	if oldPos.Row != correctRank {
		return false
	}

	// Check if the captured pawn is adjacent
	return b.lastMove.To.Col == newPos.Col && b.lastMove.To.Row == oldPos.Row
}

func (b *Board) makeMove(move Move) {
	// Update piece's HasMoved status
	move.Piece.HasMoved = true

	// Handle castling
	if move.IsCastling {
		rookFromCol := 0
		rookToCol := 3
		if move.To.Col > move.From.Col { // King-side castling
			rookFromCol = 7
			rookToCol = 5
		}
		// Move rook
		rook := b.squares[move.From.Row][rookFromCol]
		b.squares[move.From.Row][rookToCol] = rook
		b.squares[move.From.Row][rookFromCol] = nil
		rook.HasMoved = true
	}

	// Handle en passant
	if move.IsEnPassant {
		b.squares[move.From.Row][move.To.Col] = nil // Remove captured pawn
	}

	// Move piece
	b.squares[move.To.Row][move.To.Col] = move.Piece
	b.squares[move.From.Row][move.From.Col] = nil

	// Update king position if king was moved
	if move.Piece.Type == King {
		if move.Piece.Player == White {
			b.whiteKing = move.To
		} else {
			b.blackKing = move.To
		}
	}

	// Store last move for en passant
	b.lastMove = move
	b.moveCount++
}

func (b *Board) undoMove(move Move) {
	// Restore piece to original position
	b.squares[move.From.Row][move.From.Col] = move.Piece
	b.squares[move.To.Row][move.To.Col] = move.Captured

	// Restore HasMoved status
	move.Piece.HasMoved = false

	// Handle castling undo
	if move.IsCastling {
		rookFromCol := 3
		rookToCol := 0
		if move.To.Col > move.From.Col { // King-side castling
			rookFromCol = 5
			rookToCol = 7
		}
		rook := b.squares[move.From.Row][rookFromCol]
		b.squares[move.From.Row][rookToCol] = rook
		b.squares[move.From.Row][rookFromCol] = nil
		rook.HasMoved = false
	}

	// Handle en passant undo
	if move.IsEnPassant {
		capturedPawnRow := move.From.Row
		b.squares[capturedPawnRow][move.To.Col] = move.Captured
	}

	// Restore king position if necessary
	if move.Piece.Type == King {
		if move.Piece.Player == White {
			b.whiteKing = move.From
		} else {
			b.blackKing = move.From
		}
	}

	b.moveCount--
}

func (b *Board) IsInCheck(player Player) bool {
	kingPos := b.whiteKing
	if player == Black {
		kingPos = b.blackKing
	}

	// Check if any opponent's piece can capture the king
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.squares[row][col]
			if piece != nil && piece.Player != player {
				move, err := b.ValidateMove(Position{row, col}, kingPos, piece.Player)
				if err == nil && b.IsValidPieceMove(piece, Position{row, col}, kingPos, &move) {
					return true
				}
			}
		}
	}
	return false
}

func (b *Board) IsCheckmate(player Player) bool {
	if !b.IsInCheck(player) {
		return false
	}

	// Try all possible moves for all pieces
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.squares[row][col]
			if piece != nil && piece.Player == player {
				for newRow := 0; newRow < 8; newRow++ {
					for newCol := 0; newCol < 8; newCol++ {
						oldPos := Position{row, col}
						newPos := Position{newRow, newCol}

						move, err := b.ValidateMove(oldPos, newPos, player)
						if err != nil {
							continue
						}

						b.makeMove(move)
						stillInCheck := b.IsInCheck(player)
						b.undoMove(move)

						if !stillInCheck {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func (b *Board) IsStalemate(player Player) bool {
	if b.IsInCheck(player) {
		return false
	}

	// Check if the player has any legal moves
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.squares[row][col]
			if piece != nil && piece.Player == player {
				for newRow := 0; newRow < 8; newRow++ {
					for newCol := 0; newCol < 8; newCol++ {
						oldPos := Position{row, col}
						newPos := Position{newRow, newCol}

						move, err := b.ValidateMove(oldPos, newPos, player)
						if err != nil {
							continue
						}

						// Try the move
						b.makeMove(move)
						inCheck := b.IsInCheck(player)
						b.undoMove(move)

						if !inCheck {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func isValidPosition(pos Position) bool {
	return pos.Row >= 0 && pos.Row < 8 && pos.Col >= 0 && pos.Col < 8
}

func ParseMove(notation string) (Position, Position, error) {
	notation = strings.ToLower(strings.TrimSpace(notation))
	if len(notation) != 5 || notation[2] != '-' {
		return Position{}, Position{}, fmt.Errorf("invalid move format (example: e2-e4)")
	}

	fromCol := int(notation[0] - 'a')
	fromRow := 8 - int(notation[1]-'0')
	toCol := int(notation[3] - 'a')
	toRow := 8 - int(notation[4]-'0')

	if !isValidPosition(Position{fromRow, fromCol}) || !isValidPosition(Position{toRow, toCol}) {
		return Position{}, Position{}, fmt.Errorf("invalid position")
	}

	return Position{fromRow, fromCol}, Position{toRow, toCol}, nil
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}

func main() {
	board := NewBoard()
	currentPlayer := White
	scanner := bufio.NewScanner(os.Stdin)
	moveHistory := make([]string, 0)

	for {
		ClearScreen()

		// Display move history
		fmt.Println("\nMove History:")
		for i, move := range moveHistory {
			if i%2 == 0 {
				fmt.Printf("%d. %s", (i/2)+1, move)
			} else {
				fmt.Printf(" %s\n", move)
			}
		}
		fmt.Println("\n")

		// Display the board
		board.Draw()

		// Check for checkmate or stalemate
		if board.IsCheckmate(currentPlayer) {
			winner := Black
			if currentPlayer == Black {
				winner = White
			}
			fmt.Printf("\nCheckmate! %s wins!\n", winner)
			break
		}

		if board.IsStalemate(currentPlayer) {
			fmt.Println("\nStalemate! The game is a draw.")
			break
		}

		// Show if the current player is in check
		if board.IsInCheck(currentPlayer) {
			fmt.Printf("\n%s is in check!\n", currentPlayer)
		}

		// Prompt for move
		fmt.Printf("\n%s to move (example: e2-e4): ", currentPlayer)
		if !scanner.Scan() {
			break
		}
		moveStr := scanner.Text()

		// Handle special commands
		switch moveStr {
		case "quit":
			fmt.Println("Game ended.")
			return
		case "help":
			fmt.Println("\nCommands:")
			fmt.Println("- Enter moves in the format: e2-e4")
			fmt.Println("- 'quit' to end the game")
			fmt.Println("- 'help' to show this help message")
			fmt.Println("\nPress Enter to continue...")
			scanner.Scan()
			continue
		}

		// Parse and make the move
		oldPos, newPos, err := ParseMove(moveStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Press Enter to continue...")
			scanner.Scan()
			continue
		}

		err = board.Move(oldPos, newPos, currentPlayer)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Press Enter to continue...")
			scanner.Scan()
			continue
		}

		// Record the move
		moveHistory = append(moveHistory, moveStr)

		// Switch player
		currentPlayer = 1 - currentPlayer
	}

	fmt.Println("\nPress Enter to exit...")
	scanner.Scan()
}
