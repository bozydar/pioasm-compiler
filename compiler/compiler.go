package compiler

import (
	"bytes"
	"fmt"
	"strconv"
)

type Ast interface {
	ToSource() string
}

type Options struct {
	evalDefine bool
}

type CompileError struct {
	line    int
	offset  int
	message string
}

func (ce *CompileError) ToString() string {
	return fmt.Sprintf("%s: %d:%d", ce.message, ce.line, ce.offset)
}

type compiler struct {
	options        *Options
	lex            *lexer
	file           *AstFile
	pos            int
	error          *CompileError
	programs       map[string]*AstProgram
	currentProgram *AstProgram
	globalSymbols  map[string]*AstDefine
	programSymbols map[string]map[string]*AstDefine
}

type AstFile struct {
	defines  []*AstDefine
	programs []*AstProgram
}

func (a *AstFile) ToSource() string {
	var b bytes.Buffer

	for _, define := range a.defines {
		b.WriteString(define.ToSource() + "\n")
	}

	for _, program := range a.programs {
		b.WriteString(program.ToSource())
	}

	return b.String()
}

type AstProgram struct {
	name      string
	sideSet   uint8
	defines   []*AstDefine
	assembler []uint16
}

func (a *AstProgram) ToSource() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf(".program %s\n", a.name))
	for _, define := range a.defines {
		b.WriteString(define.ToSource() + "\n")
	}

	return b.String()
}

type AstDefine struct {
	name      string
	expr      AstExpr
	evaluated bool
	value     pioInt
}

func (a *AstDefine) ToSource() string {
	result := fmt.Sprintf(".define %s %s", a.name, a.expr.ToSource())
	_, ok := a.expr.(*AstBinOp)
	if a.evaluated && ok {
		result += fmt.Sprintf(" ; = %d", a.value)
	}

	return result
}

func Compile(source string, options *Options) (astFile *AstFile, error *CompileError) {
	lexer, _ := lex("lex", source)
	c := compiler{
		options:        options,
		lex:            lexer,
		pos:            -1,
		programs:       make(map[string]*AstProgram),
		currentProgram: nil,
		globalSymbols:  make(map[string]*AstDefine),
		programSymbols: make(map[string]map[string]*AstDefine),
	}

	defer func() {
		if err, ok := recover().(*CompileError); ok && err != nil {
			error = err
			astFile = nil
		}
	}()

	astFile = c.parseFile()

	return
}

func (c *compiler) next() *lexItem {
	item, ok := <-c.lex.items
	if !ok {
		return nil
	}

	return &item
}

func (c *compiler) backup() {
	c.pos = 0
}

type line []*lexItem

func (c *compiler) nextLine() line {
	result := make([]*lexItem, 0)
	for item := c.next(); item != nil && item.typ != itemEOL; item = c.next() {
		result = append(result, item)
	}

	return result
}

func (c *compiler) parseProgram(l line) *AstProgram {
	var id string
	if l[0].typ == itemDirProgram && l[1].typ == itemSymbol {
		id = l[1].val
	} else {
		c.raiseError("Syntax error near .program", l[0])
	}
	ast := &AstProgram{name: id}
	c.registerProgram(ast, l[0])
	return ast
}

func (c *compiler) registerProgram(program *AstProgram, item *lexItem) {
	for k := range c.programs {
		if k == program.name {
			c.raiseError("Program already exists", item)
		}
	}

	c.currentProgram = program
	c.programs[program.name] = program
}

func (c *compiler) parseSymbol(item *lexItem) string {
	return item.val
}

func (c *compiler) parseNumber(item *lexItem) int {
	result, ok := strconv.Atoi(item.val)
	if ok != nil {
		c.raiseError("Can't convert string to int", item)
	}

	return result
}

func (c *compiler) parseDefine(l line) *AstDefine {
	var name string
	var value AstExpr

	// TODO ensure if expressions always needs parents around. If so the matching should be simpler.
	// e.g. l[3].typ == itemLParen && l[len(l) - 1].typ == itemRParen
	if l[0].typ == itemDirDefine && l[1].typ == itemPublic && l[2].typ == itemSymbol {
		name = c.parseSymbol(l[2])
		value = c.parseExpr(l[3:])
	} else if l[0].typ == itemDirDefine && l[1].typ == itemSymbol {
		name = c.parseSymbol(l[1])
		value = c.parseExpr(l[2:])
	} else {
		c.raiseError("Syntax error near `.define`", l[0])
	}
	ast := &AstDefine{name: name, expr: value}

	// NOTE don't swap these to lines. It may allow for self-reference
	c.evaluateDefine(ast)
	c.registerDefine(ast, l[0])

	return ast
}

func (c *compiler) registerDefine(define *AstDefine, item *lexItem) {
	if d := c.getDefineDeclared(define.name); d != nil {
		c.raiseError("Symbol already defined", item)
	}

	if c.currentProgram != nil {
		if c.programSymbols[c.currentProgram.name] == nil {
			c.programSymbols[c.currentProgram.name] = make(map[string]*AstDefine)
		}
		c.programSymbols[c.currentProgram.name][define.name] = define
	} else {
		c.globalSymbols[define.name] = define
	}
}

func (c *compiler) evaluateDefine(define *AstDefine) {
	define.value = define.expr.eval(c)
	define.evaluated = true
}

func (c *compiler) getDefineDeclared(name string) *AstDefine {
	for k := range c.globalSymbols {
		if name == k {
			return c.globalSymbols[name]
		}
	}

	if c.currentProgram != nil {
		for k := range c.programSymbols[c.currentProgram.name] {
			if name == k {
				return c.programSymbols[c.currentProgram.name][k]
			}
		}
	}

	return nil
}

func (c *compiler) getValueByIdentifier(name string) pioInt {
	define := c.getDefineDeclared(name)
	if !define.evaluated {
		c.evaluateDefine(define)
	}

	return define.value
}

func (c *compiler) parseLine() (interface{}, line) {
	l := c.nextLine()
	if len(l) == 0 {
		return nil, l
	}
	item := l[0]
	switch item.typ {
	case itemDirProgram:
		return c.parseProgram(l), l
	case itemDirDefine:
		return c.parseDefine(l), l
	case itemEOF:
		return nil, l
	default:
		c.raiseError("Unexpected item", item)
		return nil, l
	}
}

func (c *compiler) parseFile() *AstFile {
	programs := make([]*AstProgram, 0)
	fileDefines := make([]*AstDefine, 0)

	for ast, _ := c.parseLine(); ast != nil; ast, _ = c.parseLine() {
		switch v := ast.(type) {
		case *AstDefine:
			// It might be c.currentProgram as well
			if len(programs) > 0 {
				programs[len(programs)-1].defines = append(programs[len(programs)-1].defines, v)
			} else {
				fileDefines = append(fileDefines, v)
			}
		case *AstProgram:
			programs = append(programs, v)
		}
	}

	result := AstFile{
		defines:  fileDefines,
		programs: programs,
	}

	return &result
}

func (c *compiler) position(pos int) (line int, offset int) {
	return c.lex.position(pos)
}

func (c *compiler) raiseError(message string, item *lexItem) {
	line, offset := c.position(item.start)
	c.error = &CompileError{message: message, line: line, offset: offset}
	panic(c.error)
}
