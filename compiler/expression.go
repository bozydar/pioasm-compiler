package compiler

import (
	"fmt"
	"strconv"
)

type exprParser struct {
	compiler *compiler
	line     line
}

type operator string

const (
	opPlus   = operator("+")
	opMinus  = operator("-")
	opMul    = operator("*")
	opDiv    = operator("/")
	opBinAnd = operator("&")
	opBinXor = operator("^")
	opBinOr  = operator("|")
)

type pioInt int32

type AstExpr interface {
	Ast
	eval(c *compiler) pioInt
	inParenthesis() bool
}

type AstValue struct {
	inParenthesisVal bool
	value            pioInt
}

func (a *AstValue) inParenthesis() bool {
	return a.inParenthesisVal
}

func (a *AstValue) eval(*compiler) pioInt {
	return a.value
}

func (a *AstValue) ToSource() string {
	return fmt.Sprintf("%d", a.value)
}

type AstBinOp struct {
	inParenthesisVal bool
	name             operator
	left             AstExpr
	right            AstExpr
}

func (a *AstBinOp) inParenthesis() bool {
	return a.inParenthesisVal
}

func (a *AstBinOp) ToSource() string {
	left := a.left.ToSource()
	right := a.right.ToSource()
	result := fmt.Sprintf("%s %s %s", left, a.name, right)
	if a.inParenthesis() {
		result = fmt.Sprintf("(%s)", result)
	}
	return result
}

func (a *AstBinOp) eval(c *compiler) pioInt {
	left := a.left.eval(c)
	right := a.right.eval(c)
	switch a.name {
	case opPlus:
		return left + right
	case opMinus:
		return left - right
	case opMul:
		return left * right
	case opDiv:
		return left / right
	case opBinOr:
		return left | right
	case opBinXor:
		return left ^ right
	case opBinAnd:
		return left & right
	default:
		panic(fmt.Sprintf("Unknown operation `%s`", a.name))
	}
}

type AstIdentifier struct {
	inParenthesisVal bool
	name             string
}

func (a *AstIdentifier) inParenthesis() bool {
	return a.inParenthesisVal
}

func (a *AstIdentifier) ToSource() string {
	return a.name
}

func (a *AstIdentifier) eval(c *compiler) pioInt {
	return c.getValueByIdentifier(a.name)
}

func (c *compiler) parseExpr(l line) AstExpr {
	ep := exprParser{line: l, compiler: c}
	return ep.parseExprBinOr(false)
}

func (ep *exprParser) parseExprBinOr(inParent bool) AstExpr {
	left := ep.parseExprBinXor(inParent)

	for len(ep.line) > 0 && ep.line[0].typ == itemBinOr {
		ep.next()
		right := ep.parseExprBinXor(inParent)
		left = &AstBinOp{name: opBinOr, left: left, right: right, inParenthesisVal: inParent}
	}

	return left
}

func (ep *exprParser) parseExprBinXor(inParent bool) AstExpr {
	left := ep.parseExprBinAnd(inParent)

	for len(ep.line) > 0 && ep.line[0].typ == itemBinXor {
		ep.next()
		right := ep.parseExprBinAnd(inParent)
		left = &AstBinOp{name: opBinXor, left: left, right: right, inParenthesisVal: inParent}
	}

	return left
}

func (ep *exprParser) parseExprBinAnd(inParent bool) AstExpr {
	left := ep.parseExprPlusMinus(inParent)

	for len(ep.line) > 0 && ep.line[0].typ == itemBinAnd {
		ep.next()
		right := ep.parseExprPlusMinus(inParent)
		left = &AstBinOp{name: opBinAnd, left: left, right: right, inParenthesisVal: inParent}
	}

	return left
}

func (ep *exprParser) parseExprPlusMinus(inParen bool) AstExpr {
	left := ep.parseExprMulDiv()

	for len(ep.line) > 0 && (ep.line[0].typ == itemPlus || ep.line[0].typ == itemMinus) {
		lexItem := ep.next()
		var name operator
		if lexItem.typ == itemPlus {
			name = opPlus
		} else {
			name = opMinus
		}
		right := ep.parseExprMulDiv()
		left = &AstBinOp{name: name, left: left, right: right, inParenthesisVal: inParen}
	}

	return left
}

func (ep *exprParser) parseExprMulDiv() AstExpr {
	left := ep.parseExprSymbolsConsParens()

	for len(ep.line) > 0 && (ep.line[0].typ == itemStar || ep.line[0].typ == itemSlash) {
		lexItem := ep.next()
		var name operator
		if lexItem.typ == itemStar {
			name = opMul
		} else {
			name = opDiv
		}
		right := ep.parseExprSymbolsConsParens()
		left = &AstBinOp{name: name, left: left, right: right}
	}

	return left
}

// TODO OR, AND, XOR
// TODO MINUS _
// TODO REVERSE _

func (ep *exprParser) parseExprSymbolsConsParens() AstExpr {
	if ep.line[0].typ == itemNumber {
		lexItem := ep.next()
		value, _ := strconv.ParseInt(lexItem.val, 0, 0)
		return &AstValue{value: pioInt(value)}
	} else if ep.line[0].typ == itemSymbol {
		lexItem := ep.next()
		if ep.compiler.getDefineDeclared(lexItem.val) == nil {
			ep.compiler.raiseError("Unknown identifier in expression", lexItem)
		}
		return &AstIdentifier{name: lexItem.val}
	}
	ep.expect(itemLParent)
	result := ep.parseExprBinOr(true)
	ep.expect(itemRParent)

	return result
}

func (ep *exprParser) next() *lexItem {
	result := ep.line[0]
	ep.line = ep.line[1:]

	return result
}

func (ep *exprParser) expect(itemType itemType) {
	lexItem := ep.next()
	if lexItem.typ != itemType {
		ep.compiler.raiseError("Syntax error in expression", lexItem)
	}
}
