package libaduk

import (
    "fmt"
    "log"
)

// Represents a Go board data structure
type AbstractBoard struct {
    BoardSize uint8
    data []BoardStatus
    undoStack []*Move
}

// Creates new Go Board
func NewBoard(boardSize uint8) (*AbstractBoard, error) {
    if boardSize < 1 {
        return nil, fmt.Errorf("Boardsize can not be less than 1!")
    }

    return &AbstractBoard {
        boardSize,
        make([]BoardStatus, boardSize * boardSize),
        make([]*Move, 0),
    }, nil
}

// Returns a string representation of the current board status
func (board *AbstractBoard) ToString() string {
    result := ""

    for y := uint8(0); y < board.BoardSize; y++ {
        for x := uint8(0); x < board.BoardSize; x++ {
            switch board.getStatus(x, y) {
            case EMPTY:
                result += ". "
            case BLACK:
                result += "X "
            case WHITE:
                result += "O "
            }
        }
        result += "\n"
    }

    return result
}

// Clears the board
func (board *AbstractBoard) Clear() {
    for i := 0; i < len(board.data); i++ {
        board.data[i] = EMPTY
    }
    board.undoStack = []*Move { }
}

// Returns the Top Move of the Undostack
func (board *AbstractBoard) UndostackTopMove() *Move {
    return board.undoStack[len(board.undoStack) - 1]
}

// Removes last Move from Undostack
func (board *AbstractBoard) UndostackPop() (move *Move) {
    if len(board.undoStack) > 0 {
        move = board.undoStack[len(board.undoStack) - 1]
        board.undoStack = board.undoStack[:len(board.undoStack) - 1]
    }

    return
}

// Adds the given Move to the Undostack
func (board *AbstractBoard) UndostackPush(move *Move) {
    log.Printf("Add Move to Undostack: %+v", move)

    board.undoStack = append(board.undoStack, move)
}

// Adds a Pass to the Undostack
func (board *AbstractBoard) UndostackPushPass() {
    board.UndostackPush(&Move { 255, 255, PASS, nil })
}

// Play move on board
func (board *AbstractBoard) PlayMove(move Move) (error) {
    return board.Play(move.X, move.Y, move.Color)
}

// Play stone at given position
func (board *AbstractBoard) Play(x uint8, y uint8, color BoardStatus) (error) {
    log.Printf("Play: X: %v, Y: %v, Color: %v", x, y, color)

    // Is move on the board?
    if x < 0 || x >= board.BoardSize || y < 0 || y >= board.BoardSize {
        return fmt.Errorf("Invalid move position!")
    }

    // Is already a stone on this position?
    if board.getStatus(x, y) != EMPTY {
        return fmt.Errorf("Position already occupied!")
    }

    // Check if move is legal and get captures
    captures, err := board.legal(x, y, color)
    if err != nil {
        return err
    }

    // Remove captures
    for i := 0; i < len(captures); i++ {
        board.setStatus(captures[i].X, captures[i].Y, EMPTY)
    }

    // Add them to undostack
    board.UndostackPush(&Move { x, y, color, captures })

    return nil
}

// Checks if move is legal and returns captured stones if necessary
func (board *AbstractBoard) legal(x uint8, y uint8, color BoardStatus) (captures []Position, err error) {
    captures = []Position { }
    neighbours := board.getNeighbours(x, y)

    log.SetPrefix("legal ")
    log.Printf("Neighbours for Playmove (X: %d, Y: %d) are %+v", x, y, neighbours)

    // Check if we capture neighbouring stones
    for i := 0; i < len(neighbours); i++ {
        // Is neighbour from another color?
        if board.getStatus(neighbours[i].X, neighbours[i].Y) == board.invertColor(color) {
            log.SetPrefix("legal ")
            log.Printf("Neighbour of Playmove (X: %d, Y: %d) at (X: %d, Y: %d) is %v. Get its No liberties...",
                x, y, neighbours[i].X, neighbours[i].Y, board.invertColor(color))

            // Get enemy stones with no liberties left
            noLibertyStones := board.getNoLibertyStones(neighbours[i].X, neighbours[i].Y, Position { x, y })
            for j := 0; j < len(noLibertyStones); j++ {
                captures = append(captures, noLibertyStones[j])
            }
        }
    }

    board.setStatus(x, y, color)

    // TODO: Delete Duplicates necessary????
    if len(captures) > 0 {
        return
    }

    // Check if the played move has no liberties and therefore is a suicide
    log.SetPrefix("legal ")
    log.Printf("Check if Playmove (%d, %d) is a suicide.", x, y)
    selfNoLiberties := board.getNoLibertyStones(x, y, Position { })
    if len(selfNoLiberties) > 0 {
        // Take move back
        board.setStatus(x, y, EMPTY)
        err = fmt.Errorf("Invalid move (Suicide not allowed)!")
    }

    log.SetPrefix("")
    return
}

// Get all stones with no liberties left on given position
func (board *AbstractBoard) getNoLibertyStones(x uint8, y uint8, orgPosition Position) (noLibertyStones []Position) {
    log.SetPrefix("getNoLibertyStones ")
    log.Printf("Get no liberty stones for (%d, %d)", x, y)

    noLibertyStones = []Position { }
    newlyFoundStones := []Position { Position { x, y } }
    foundNew := true
    var groupStones []Position = nil

    // Search until no new stones are found
    for foundNew == true {
        foundNew = false
        groupStones = []Position { }

        for i := 0; i < len(newlyFoundStones); i++ {
            x1 := newlyFoundStones[i].X
            y1 := newlyFoundStones[i].Y
            neighbours := board.getNeighbours(x1, y1)

            // Check liberties of stone x1, y1 by checking the neighbours
            for j := 0; j < len(neighbours); j++ {
                nbX := neighbours[j].X
                nbY := neighbours[j].Y

                // Has x1, y1 a free liberty?
                if board.getStatus(nbX, nbY) == EMPTY && !neighbours[j].isSamePosition(orgPosition) {
                    log.SetPrefix("getNoLibertyStones ")
                    log.Printf("Neighbour (%d, %d) is empty and not (%d, %d) so (%d, %d) has at least liberty",
                        nbX, nbY, orgPosition.X, orgPosition.Y, x, y)
                    return noLibertyStones[:0]
                } else {
                    // Is the neighbour of x1, y1 the same color? Then we have a group here
                    if board.getStatus(x1, y1) == board.getStatus(nbX, nbY) {
                        foundNewHere := true
                        groupStone := Position { nbX, nbY }

                        log.SetPrefix("getNoLibertyStones ")
                        log.Printf("Found group stone for (%d, %d) at %+v", x1, y1, groupStone)

                        // Check if found stone is already in our group list
                        for k := 0; k < len(groupStones); k++ {
                            if groupStones[k].isSamePosition(groupStone) {
                                foundNewHere = false
                                break
                            }
                        }

                        // Check if found stone is already in result set list
                        if foundNewHere {
                            for k := 0; k < len(noLibertyStones); k++ {
                                if noLibertyStones[k].isSamePosition(groupStone) {
                                    foundNewHere = false
                                    break
                                }
                            }
                        }

                        // If groupStone is not known yet, add it
                        if foundNewHere {
                            groupStones = append(groupStones, groupStone)
                            foundNew = true
                        }
                    }
                }
            }
        }

        // Add newly found stones to the resultset
        noLibertyStones = append(noLibertyStones, newlyFoundStones...)

        // Now check the found group stones
        newlyFoundStones = groupStones
    }

    log.SetPrefix("getNoLibertyStones ")
    log.Printf("Found these stones with no liberties: %+v", noLibertyStones)
    log.SetPrefix("")

    return
}

// Returns the neighbour array positions for a given point
func (board *AbstractBoard) getNeighbours(x uint8, y uint8) (neighbourIndexes []Position) {
    neighbourIndexes = []Position { }

    // Check for board borders
    if x > 0 {
        neighbourIndexes = append(neighbourIndexes, Position { (x - 1), y })
    }
    if x < board.BoardSize - 1 {
        neighbourIndexes = append(neighbourIndexes, Position { (x + 1), y })
    }
    if y > 0 {
        neighbourIndexes = append(neighbourIndexes, Position { x, y - 1 })
    }
    if y < board.BoardSize - 1 {
        neighbourIndexes = append(neighbourIndexes, Position { x, y + 1 })
    }

    return
}

func (board *AbstractBoard) getStatus(x uint8, y uint8) BoardStatus {
    return board.data[board.BoardSize * x + y]
}

func (board *AbstractBoard) setStatus(x uint8, y uint8, status BoardStatus) {
    board.data[board.BoardSize * x + y] = status
}

// Inverts Black to White or White to Black
func (board *AbstractBoard) invertColor(color BoardStatus) BoardStatus {
    if color == WHITE {
        return BLACK
    }

    if color == BLACK {
        return WHITE
    }

    return EMPTY
}
