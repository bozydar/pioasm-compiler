package compiler

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"
)

type stateFn func(l *lexer) stateFn

type lexItem struct {
	typ   itemType
	val   string
	end   int
	start int
}

type itemType int

const (
	itemError = iota
	itemEOF
	itemEOL
	itemNumber
	itemSymbol
	itemDirDefine
	itemDirProgram
	itemDirOrigin
	itemDirSideSet
	itemDirWrapTarget
	itemDirWrap
	itemDirLangOpt
	itemDirWord
	itemInstrJMP
	itemInstrWAIT
	itemInstrIN
	itemInstrOUT
	itemInstrPUSH
	itemInstrPULL
	itemInstrMOV
	itemInstrIRQ
	itemInstrSET
	itemInstrNOP
	itemPublic
	itemOptional
	itemSide
	itemPin
	itemGPIO
	itemOSRE

	itemReverse
	itemComma
	itemLabel
	itemLBracket
	itemRBracket
	itemLParent
	itemRParent
	itemPlus
	itemMinus
	itemStar
	itemSlash
	itemBinAnd
	itemBinOr
	itemBinXor
	itemBang
	itemEqual
)

const (
	semicolon = ';'
	colon     = ':'
	dot       = '.'
	comma     = ','
	eol       = '\n'
	plus      = '+'
	minus     = '-'
	star      = '*'
	slash     = '/'
	binAnd    = '&'
	binOr     = '|'
	binXor    = '^'
	lBracket  = '['
	rBracket  = ']'
	lParent   = '('
	rParent   = ')'
	bang      = '!'
	equal     = '='

	eof = 0
)

var whites = []rune{'\t', '\n', '\r', ' '}

func (i lexItem) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}

type lexer struct {
	name     string
	input    string
	lines    []int
	start    int
	pos      int
	width    int
	items    chan lexItem
	lastItem *lexItem
}

func lex(name, input string) (*lexer, chan lexItem) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan lexItem),
	}

	go l.run()

	return l, l.items
}

func (l *lexer) run() {
	for state := lexContent; state != nil; {
		state = state(l)
	}

	close(l.items)
}

func (l *lexer) position(pos int) (line int, offset int) {
	linesLen := len(l.lines)
	line = sort.Search(linesLen, func(i int) bool {
		return l.lines[i] > pos
	}) - 1

	if line >= linesLen {
		line = linesLen - 1
	}

	lineOffset := 0
	if line >= 0 {
		lineOffset = l.lines[line]
	}

	offset = pos - lineOffset
	if lineOffset == 0 {
		offset += 1
	}
	line += 2

	return
}

func (l *lexer) emit(t itemType) {
	item := lexItem{typ: t, val: l.input[l.start:l.pos], start: l.start, end: l.pos}
	// Don't put item if the last item was eol and current is eol
	// or there was no last item (current item is the first one) and it is an EOL
	if !(item.typ == itemEOL && (l.lastItem == nil || l.lastItem != nil && l.lastItem.typ == itemEOL)) {
		l.items <- item
	}

	l.lastItem = &item
	l.start = l.pos
}

func (l *lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	if rune == '\n' {
		l.addLine(l.pos)
	}
	l.pos += l.width

	return rune
}

func (l *lexer) addLine(pos int) {
	if i := len(l.lines); i == 0 || l.lines[i-1] < pos {
		l.lines = append(l.lines, pos)
	}
}

func (l *lexer) back() {
	l.pos = l.start
	l.width = 0
}

func (l *lexer) acceptStringCI(s string) bool {
	for _, r := range s {
		nextUp := unicode.ToUpper(l.next())
		rUp := unicode.ToUpper(r)
		if nextUp != rUp {
			l.back()
			return false
		}
	}
	if isWordEnd(l.peek()) {
		return true
	}
	l.back()
	return false
}

func (l *lexer) peek() (rune rune) {
	if l.pos >= len(l.input) {
		return eof
	}
	rune, _ = utf8.DecodeRuneInString(l.input[l.pos:])

	return rune
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func isWhite(r rune) bool {
	if r == eof || r == eol {
		return true
	}
	for _, v := range whites {
		if v == r {
			return true
		}
	}
	return false
}

func isWordEnd(r rune) bool {
	return isWhite(r)
}

func lexContent(l *lexer) stateFn {
	for {
		if l.acceptStringCI("jmp") {
			l.emit(itemInstrJMP)
			return lexContent
		} else if l.acceptStringCI("WAIT") {
			l.emit(itemInstrWAIT)
			return lexContent
		} else if l.acceptStringCI("IN") {
			l.emit(itemInstrIN)
			return lexContent
		} else if l.acceptStringCI("OUT") {
			l.emit(itemInstrOUT)
			return lexContent
		} else if l.acceptStringCI("PUSH") {
			l.emit(itemInstrPUSH)
			return lexContent
		} else if l.acceptStringCI("PULL") {
			l.emit(itemInstrPULL)
			return lexContent
		} else if l.acceptStringCI("MOV") {
			l.emit(itemInstrMOV)
			return lexContent
		} else if l.acceptStringCI("IRQ") {
			l.emit(itemInstrIRQ)
			return lexContent
		} else if l.acceptStringCI("SET") {
			l.emit(itemInstrSET)
			return lexContent
		} else if l.acceptStringCI("NOP") {
			l.emit(itemInstrNOP)
			return lexContent
		} else if l.acceptStringCI("PUBLIC") {
			l.emit(itemPublic)
			return lexContent
		} else if l.acceptStringCI("OPTIONAL") || l.acceptStringCI("OPT") {
			l.emit(itemOptional)
			return lexContent
		} else if l.acceptStringCI("SIDE") || l.acceptStringCI("SIDESET") || l.acceptStringCI("SIDE_SET") {
			l.emit(itemSide)
			return lexContent
		} else if l.acceptStringCI("PIN") {
			l.emit(itemPin)
			return lexContent
		} else if l.acceptStringCI("GPIO") {
			l.emit(itemGPIO)
			return lexContent
		} else if l.acceptStringCI("OSRE") {
			l.emit(itemOSRE)
			return lexContent
		}

		next := l.next()

		if isEOF(next) {
			l.emit(itemEOF)
			return nil
		} else if isEOL(next) {
			l.emit(itemEOL)
			return lexContent
		} else if isComment(next, l) {
			l.ignore()
			return lexComment
		} else if next == eol {
			l.emit(itemEOL)
		} else if isWhite(next) {
			l.ignore()
		} else if isSymbol(next) {
			return lexSymbolOrLabel
		} else if isDirective(next) {
			return lexDirective
		} else if isValue(next) {
			return lexValue
		} else if next == lBracket {
			l.emit(itemLBracket)
		} else if next == rBracket {
			l.emit(itemRBracket)
		} else if next == lParent {
			l.emit(itemLParent)
		} else if next == rParent {
			l.emit(itemRParent)
		} else if next == plus {
			l.emit(itemPlus)
		} else if next == minus {
			l.emit(itemMinus)
		} else if next == star {
			l.emit(itemStar)
		} else if next == slash {
			l.emit(itemSlash)
		} else if next == binAnd {
			l.emit(itemBinAnd)
		} else if next == binOr {
			l.emit(itemBinOr)
		} else if next == binXor {
			l.emit(itemBinXor)
		} else if next == bang {
			l.emit(itemBang)
		} else if next == equal {
			l.emit(itemEqual)
		} else if next == comma {
			l.emit(itemComma)
		}
	}
}

func lexSymbolOrLabel(l *lexer) stateFn {
	for {
		next := l.next()
		if next != dot && !unicode.IsNumber(next) && !isSymbol(next) {
			if next == colon {
				// TODO "public label:"
				l.emit(itemLabel)
			} else {
				l.backup()
				l.emit(itemSymbol)
			}
			return lexContent
		}
	}
}

func lexDirective(l *lexer) stateFn {
	for {
		next := l.next()
		if !isSymbol(next) {
			l.backup()
			directive := l.input[l.start:l.pos]
			switch directive {
			case ".define":
				l.emit(itemDirDefine)
			case ".program":
				l.emit(itemDirProgram)
			case ".origin":
				l.emit(itemDirOrigin)
			case ".side_set":
				l.emit(itemDirSideSet)
			case ".wrap_target":
				l.emit(itemDirWrapTarget)
			case ".wrap":
				l.emit(itemDirWrap)
			case ".lang_opt":
				l.emit(itemDirLangOpt)
			case ".word":
				l.emit(itemDirWord)
			}
			return lexContent
		}
	}
}

func lexValue(l *lexer) stateFn {
	// TODO hex (0xf), binary (0b1001)
	for {
		next := l.next()
		if !isValue(next) {
			if isSymbol(next) {
				l.emit(itemError)
				return nil
			}
			l.backup()
			l.emit(itemNumber)
			return lexContent
		}
	}
}

func lexComment(l *lexer) stateFn {
	for {
		peek := l.peek()

		if isEOL(peek) || isEOF(peek) {
			return lexContent
		}

		l.next()
		l.ignore()
	}
}

func isSymbol(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isValue(r rune) bool {
	return unicode.IsNumber(r)
}

func isDirective(r rune) bool {
	return r == dot
}

func isComment(r rune, l *lexer) bool {
	// Comment is: ";" or "//"
	return semicolon == r || (slash == r && slash == l.peek())
}

func isEOF(r rune) bool {
	return r == eof
}

func isEOL(r rune) bool {
	return r == eol
}
