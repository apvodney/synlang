%{

package main

import (
	"fmt"
	"unicode/utf8"
	"strings"
)

%}


%%

// itemType identifies the type of lex items.
type itemType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	itemError        itemType = iota // error occurred; value is text of error
	itemAssign                       // equals ('=') introducing an assignment
	itemEOF
	itemIdentifier   // alpha identifier
	itemLeftBlock    // '{'
	itemLeftList     // '['
	itemLeftParen    // '('
	itemNumber       // simple number, including imaginary
	itemRightBlock   // '}'
	itemRightList    // ']'
	itemRightParen   // ')'
	itemSpace        // run of spaces separating arguments
	// Keywords appear after all the rest.
	itemKeyword  // used only to delimit the keywords
	itemDot      // the cursor, spelled '.'
)

var key = map[string]itemType{
	".":        itemDot,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input      string    // the string being scanned
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
	line       int       // 1+number of newlines seen
	startLine  int       // start line of this item
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
	l.start = l.pos
	l.startLine = l.line
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		input:      input,
		items:      make(chan item),
		line:       1,
		startLine:  1,
	}
	go func(){
		for state := lexBlock; state != nil; {
			state = state(l)
		}
		close(l.items)
	}()
	return l
}

const (
	leftBlock = "{"
	rightBlock = "}"
	leftList = "["
	rightList = "]"
	leftParen = "("
	rightParen = ")"
)

func lexBlock(l *lexer) stateFn {
	// Either number, quoted string, or identifier.
	// Spaces separate arguments; runs of spaces turn into itemSpace.
	// Pipe symbols separate and are emitted.
	switch r := l.next(); {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed action")
	case isSpace(r):
		return lexSpace
	case r == '=':
		l.emit(itemAssign)
	case unicode.IsLetter(r):
		l.backup()
		return lexIdentifier
	case r == '.':
		// special look-ahead for ".field" so we don't break l.backup().
		if l.pos < Pos(len(l.input)) {
			r := l.input[l.pos]
			if !('0' <= r && r <= '9') {
				l.backup()
				return lexIdentifier
			}
		}
		fallthrough // '.' can start a number.
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case r == leftParen:
		l.emit(itemLeftParen)
	case r == rightParen:
		l.emit(itemRightParen)
	case r == leftList:
		l.emit(itemLeftList)
	case r == rightList:
		l.emit(itemRightList)
	case r == leftBlock:
		l.emit(itemLeftBlock)
	case r == rightBlock:
		l.emit(itemRightBlock)
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexBlock
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return lexBlock
}

func lexIdentifier(l *lexer) stateFn {
	if r := l.next(); r == '.' {
		l.emit(itemIdentifier)
	}
Loop:
	for {
		switch r := l.next(); {
		case unicode.IsLetter(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			l.emit(itemIdentifier)
			if l.peek() != '.' {
				break Loop
			}
			l.next()
			l.ignore()
		}
	}
	return lexBlock
}

