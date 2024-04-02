package lir

import "github.com/mantton/calypso/internal/calypso/token"

// LLVM Langauge Reference : https://llvm.org/docs/LangRef.html
type ICompOp byte
type FCompOp byte

const (
	INVALID_ICOMP ICompOp = iota
	EQL
	NEQ

	// Unsigned
	ULSS
	UGTR
	UGEQ
	ULEQ

	// Signed
	SLSS
	SGTR
	SGEQ
	SLEQ
)

const (
	INVALID_FCOMP FCompOp = iota
	// Ordered
	OEQ
	OGT
	OGE
	OLT
	OLE
	ONE
	ORD

	// Unordered
	UEQ
	UGE
	ULT
	ULE
	UNE
	UNO
)

type Load struct {
	yielder
	Address Value
}

type Allocate struct {
	yielder
	OnHeap bool
}

type Store struct {
	Address Value
	Value   Value
}

type Call struct {
	yielder
	Target    *Function
	Arguments []Value
}

type Return struct {
	Result Value
}

type ReturnVoid struct {
}

type ConditionalBranch struct {
	Condition   Value
	Action      *Block // Then
	Alternative *Block // Else
}

type Branch struct {
	Block *Block
}

// Addition
type Add struct {
	yielder
	Left  Value
	Right Value
}

type FAdd struct {
	yielder
	Left  Value
	Right Value
}

// Subtraction
type Sub struct {
	yielder
	Left  Value
	Right Value
}

type FSub struct {
	yielder
	Left  Value
	Right Value
}

// Multiplication
type Mul struct {
	yielder
	Left  Value
	Right Value
}
type FMul struct {
	yielder
	Left  Value
	Right Value
}

// Division
type UDiv struct {
	yielder
	Left  Value
	Right Value
}

type SDiv struct {
	yielder
	Left  Value
	Right Value
}

type FDiv struct {
	yielder
	Left  Value
	Right Value
}

// Remainder
type URem struct {
	yielder
	Left  Value
	Right Value
}

type SRem struct {
	yielder
	Left  Value
	Right Value
}

type FRem struct {
	yielder
	Left  Value
	Right Value
}

// Negation

type INeg struct {
	yielder
	Right Value
}
type FNeg struct {
	yielder
	Right Value
}

// Comparisons
type ICmp struct {
	yielder
	Left       Value
	Right      Value
	Comparison ICompOp
}

type FCmp struct {
	yielder
	Left       Value
	Right      Value
	Comparison ICompOp
}

// XOR
type XOR struct {
	yielder
	Left  Value
	Right Value
}

var UOpMap = map[token.Token]ICompOp{
	token.L_CHEVRON: ULSS,
	token.R_CHEVRON: UGTR,
	token.EQL:       EQL,
	token.LEQ:       ULEQ,
	token.GEQ:       UGEQ,
}

var SOpMap = map[token.Token]ICompOp{
	token.L_CHEVRON: SLSS,
	token.R_CHEVRON: SGTR,
	token.EQL:       EQL,
	token.LEQ:       SLEQ,
	token.GEQ:       SGEQ,
}
