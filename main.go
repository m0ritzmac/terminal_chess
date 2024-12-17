package main

import (
	"fmt"
	"strings"
)

type Player byte

type Piece string

type Board [][]Piece

type Position struct {
	row    int
	column int
}

const (
	white Player = iota
	black
)

const (
	whitePawn   Piece = "♙"
	whiteRook   Piece = "♖"
	whiteKnight Piece = "♘"
	whiteBishop Piece = "♗"
	whiteQueen  Piece = "♕"
	whiteKing   Piece = "♔"
	blackPawn   Piece = "♟"
	blackRook   Piece = "♜"
	blackKnight Piece = "♞"
	blackBishop Piece = "♝"
	blackQueen  Piece = "♛"
	blackKing   Piece = "♚"
	empty       Piece = "·"
)

func newBoard() *Board {
	return &Board{
		{blackRook, blackKnight, blackBishop, blackQueen, blackKing, blackBishop, blackKnight, blackRook},
		{blackPawn, blackPawn, blackPawn, blackPawn, blackPawn, blackPawn, blackPawn, blackPawn},
		{empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty},
		{empty, empty, empty, empty, empty, empty, empty, empty},
		{whitePawn, whitePawn, whitePawn, whitePawn, whitePawn, whitePawn, whitePawn, whitePawn},
		{whiteRook, whiteKnight, whiteBishop, whiteQueen, whiteKing, whiteBishop, whiteKnight, whiteRook},
	}
}

func (b *Board) draw() {
	for _, row := range *b {
		for _, cell := range row {
			if cell == "" {
				fmt.Print(empty, " ")
			} else {
				fmt.Print(cell, " ")
			}
		}
		fmt.Println()
	}
}

func (b *Board) move(old Position, new Position, piece Piece) {
	(*b)[old.row][old.column] = empty
	(*b)[new.row][new.column] = piece
}

func (b *Board) getPiece(p Position) Piece {
	return (*b)[p.row][p.column]
}

func (p *Player) switchPlayer() {
	if *p == white {
		*p = black
	} else {
		*p = white
	}
}

func (p Player) String() string {
	if p == white {
		return "White"
	}
	return "Black"
}

func getMoveFromNotation(n string) (Position, Position) {
	positions := strings.Split(n, "-")
	oldNot := positions[0]
	newNot := positions[1]

	oldColumn := int(oldNot[0] - 'a')
	oldRow := int(oldNot[1] - '1')
	newColumn := int(newNot[0] - 'a')
	newRow := int(newNot[1] - '1')
	oldRow = 7 - oldRow
	newRow = 7 - newRow

	oldPosition := Position{row: oldRow, column: oldColumn}
	newPosition := Position{row: newRow, column: newColumn}

	return oldPosition, newPosition
}

func main() {
	board := newBoard()

	player := white

	for {
		board.draw()
		fmt.Printf("\nPlayer %s, please make a move!\n", player)

		var move string

		for {
			fmt.Print("Enter your move")
			fmt.Scan(&move)

			oldPos, newPos := getMoveFromNotation(move)

			piece := board.getPiece(oldPos)
			if piece == empty {
				fmt.Println("Invalid move: No piece at the source position")
				continue
			}

			board.move(oldPos, newPos, piece)
			break
		}

		player.switchPlayer()
	}
}
