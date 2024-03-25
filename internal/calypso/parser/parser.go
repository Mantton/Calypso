package parser

import (
	"errors"
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/commands/utils"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

type NodeType byte

const (
	DECL NodeType = iota
	STMT
	EXPR
)

type Parser struct {
	file   *lexer.File
	errors lexer.ErrorList

	inSwitch bool
	cursor   int

	modifiers []token.Token
}

func New(file *lexer.File) *Parser {
	return &Parser{file: file}
}

func ParseString(input string) (*ast.File, lexer.ErrorList) {
	lF := lexer.NewFileFromString(input)
	p := New(lF)
	return p.Parse(), p.errors
}

func (p *Parser) Parse() *ast.File {
	moduleName := ""
	var declarations []ast.Declaration

	file := &ast.File{
		ModuleName:   moduleName,
		Declarations: declarations,
		LexerFile:    p.file,
	}
	defer func() {
		file.Errors = p.errors
	}()
	// - Parse Module Declaration

	// Module Header
	_, err := p.expect(token.MODULE) // Consume module token
	if err != nil {
		p.handleError(err, DECL)
		return file
	}

	tok, err := p.expect(token.IDENTIFIER) // Identifier
	if err != nil {
		p.handleError(err, DECL)
		return file
	}
	file.ModuleName = tok.Lit
	_, err = p.expect(token.SEMICOLON) // Consume semicolon
	if err != nil {
		p.handleError(err, DECL)
		return file
	}

	// Imports

	// Declarations

	for p.current() != token.EOF {
		decl, err := p.parseDeclaration()
		if err != nil {
			p.handleError(err, DECL)
			return file
		}
		declarations = append(declarations, decl)
	}

	file.Declarations = declarations
	return file
}

// 0 -> Decl

func (p *Parser) handleError(err error, node NodeType) {
	var hasMoved bool

	switch node {
	case DECL:
		hasMoved = p.advance(token.IsDeclaration)
	case STMT:
		hasMoved = p.advance(token.IsStatement)
	}

	// avoid infinite loop
	if !hasMoved {
		p.next()
	}

	p.errors.Add(err)
}

func ParseFileSet(set *utils.FileSet) (*ast.FileSet, error) {

	// Lex all Files

	files := &ast.FileSet{}
	for _, path := range set.FilesPaths {

		// Lex
		f, err := lexer.NewFile(path)
		if err != nil {
			return nil, err
		}

		l := lexer.New(f)
		l.ScanAll()

		// Parse
		p := New(f)
		file := p.Parse()

		if len(file.Errors) != 0 {
			return nil, errors.New(file.Errors.String())
		}

		// Add to Set
		if files.ModuleName == "" {
			files.ModuleName = file.ModuleName
		} else if files.ModuleName != file.ModuleName {
			return nil, fmt.Errorf("multiple modules in the same directory: %s, %s", files.ModuleName, file.ModuleName)
		}

		files.Files = append(files.Files, file)
	}

	return files, nil
}
