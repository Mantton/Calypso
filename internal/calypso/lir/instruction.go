package lir

import (
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

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
	Address Value
}

type Allocate struct {
	TypeOf types.Type
	OnHeap bool
}

type Store struct {
	Address Value
	Value   Value
}

type Call struct {
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
	Left  Value
	Right Value
}

type FAdd struct {
	Left  Value
	Right Value
}

// Subtraction
type Sub struct {
	Left  Value
	Right Value
}

type FSub struct {
	Left  Value
	Right Value
}

// Multiplication
type Mul struct {
	Left  Value
	Right Value
}
type FMul struct {
	Left  Value
	Right Value
}

// Division
type UDiv struct {
	Left  Value
	Right Value
}

type SDiv struct {
	Left  Value
	Right Value
}

type FDiv struct {
	Left  Value
	Right Value
}

// Remainder
type URem struct {
	Left  Value
	Right Value
}

type SRem struct {
	Left  Value
	Right Value
}

type FRem struct {
	Left  Value
	Right Value
}

// Negation

type INeg struct {
	Right Value
}
type FNeg struct {
	Right Value
}

// Comparisons
type ICmp struct {
	Left       Value
	Right      Value
	Comparison ICompOp
}

type FCmp struct {
	Left       Value
	Right      Value
	Comparison ICompOp
}

// XOR
type XOR struct {
	Left, Right Value
}

type ShiftLeft struct {
	Left, Right Value
}

type ArithmeticShiftRight struct {
	Left, Right Value
}

type LogicalShiftRight struct {
	Left, Right Value
}

type AND struct {
	Left, Right Value
}

type OR struct {
	Left, Right Value
}

type PHI struct {
	Nodes []*PhiNode
}

type PhiNode struct {
	Value Value
	Block *Block
}

// Get Element Pointer
type GEP struct {
	Index     int
	Address   Value
	Composite *Composite
}

// Extract Value
type ExtractValue struct {
	Index     int
	Address   Value
	Composite *Composite
}

var UOpMap = map[token.Token]ICompOp{
	token.L_CHEVRON: ULSS,
	token.R_CHEVRON: UGTR,
	token.LEQ:       ULEQ,
	token.GEQ:       UGEQ,
	token.EQL:       EQL,
	token.NEQ:       NEQ,
}

var SOpMap = map[token.Token]ICompOp{
	token.NEQ: NEQ,
	token.EQL: EQL, token.L_CHEVRON: SLSS,
	token.R_CHEVRON: SGTR,
	token.LEQ:       SLEQ,
	token.GEQ:       SGEQ,
}

// Conformance
func (c *Call) Yields() types.Type     { return c.Target.Signature().Result.Type() }
func (c *Load) Yields() types.Type     { return types.Dereference(c.Address.Yields()) }
func (c *Allocate) Yields() types.Type { return types.NewPointer(c.TypeOf) }
func (c *GEP) Yields() types.Type      { return types.NewPointer(c.Composite.Members[c.Index]) }

func (c *Add) Yields() types.Type  { return c.Left.Yields() }
func (c *FAdd) Yields() types.Type { return c.Left.Yields() }
func (c *Sub) Yields() types.Type  { return c.Left.Yields() }
func (c *FSub) Yields() types.Type { return c.Left.Yields() }
func (c *Mul) Yields() types.Type  { return c.Left.Yields() }
func (c *FMul) Yields() types.Type { return c.Left.Yields() }
func (c *UDiv) Yields() types.Type { return c.Left.Yields() }
func (c *SDiv) Yields() types.Type { return c.Left.Yields() }
func (c *FDiv) Yields() types.Type { return c.Left.Yields() }
func (c *URem) Yields() types.Type { return c.Left.Yields() }
func (c *SRem) Yields() types.Type { return c.Left.Yields() }
func (c *FRem) Yields() types.Type { return c.Left.Yields() }
func (c *ICmp) Yields() types.Type { return c.Left.Yields() }
func (c *FCmp) Yields() types.Type { return c.Left.Yields() }

func (c *INeg) Yields() types.Type { return c.Right.Yields() }
func (c *FNeg) Yields() types.Type { return c.Right.Yields() }

func (c *XOR) Yields() types.Type                  { return c.Left.Yields() }
func (c *ShiftLeft) Yields() types.Type            { return c.Left.Yields() }
func (c *ArithmeticShiftRight) Yields() types.Type { return c.Left.Yields() }
func (c *LogicalShiftRight) Yields() types.Type    { return c.Left.Yields() }
func (c *AND) Yields() types.Type                  { return c.Left.Yields() }
func (c *OR) Yields() types.Type                   { return c.Left.Yields() }
func (c *PHI) Yields() types.Type                  { return c.Nodes[0].Value.Yields() }
func (c *ExtractValue) Yields() types.Type         { return c.Composite.Members[c.Index] }
