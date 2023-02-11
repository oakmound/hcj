package hcj

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
)

type SelectorTokenStream interface {
	// Next will return io.EOF if no tokens remain, or ErrInvalidToken if invalid input is seen
	Next() (SelectorToken, error)
	ExpectNext(typ SelectorTokenType) (SelectorToken, error)
}

type SelectorToken struct {
	Type SelectorTokenType
	Raw  []byte
}

type SelectorTokenType int

const (
	SelectorTokenTypeIdentifier SelectorTokenType = iota
	SelectorTokenTypeGlobal
	SelectorTokenTypeIDStart
	SelectorTokenTypeClassStart
	SelectorTokenTypeAttributeStart
	SelectorTokenTypeAttributeStop
	SelectorTokenTypePseudoClassStart
	SelectorTokenTypeSubSelectorStart
	SelectorTokenTypeSubSelectorStop
	SelectorTokenTypeDescendantStart
	SelectorTokenTypeChildStart
	SelectorTokenTypeNextSiblingStart
	SelectorTokenTypeSubsequentSiblingStart
	SelectorTokenTypeOr
)

func (stt SelectorTokenType) String() string {
	switch stt {
	case SelectorTokenTypeIdentifier:
		return "Identifier"
	case SelectorTokenTypeGlobal:
		return "Global"
	case SelectorTokenTypeIDStart:
		return "IDStart"
	case SelectorTokenTypeClassStart:
		return "ClassStart"
	case SelectorTokenTypeAttributeStart:
		return "AttributeStart"
	case SelectorTokenTypeAttributeStop:
		return "AttributeStop"
	case SelectorTokenTypePseudoClassStart:
		return "PseudoClassStart"
	case SelectorTokenTypeSubSelectorStart:
		return "SubSelectorStart"
	case SelectorTokenTypeSubSelectorStop:
		return "SubSelectorStop"
	case SelectorTokenTypeDescendantStart:
		return "DescendantStart"
	case SelectorTokenTypeChildStart:
		return "ChildStart"
	case SelectorTokenTypeNextSiblingStart:
		return "NextSiblingStart"
	case SelectorTokenTypeSubsequentSiblingStart:
		return "SubsequentSiblingStart"
	case SelectorTokenTypeOr:
		return "Or"
	default:
		return "Unknown"
	}
}

func TokenizeSelector(s string) SelectorTokenStream {
	return selectorTokenStream{
		r: bufio.NewReader(strings.NewReader(s)),
	}
}

type selectorTokenStream struct {
	r         *bufio.Reader
	lastToken SelectorToken
}

func (sts selectorTokenStream) ExpectNext(typ SelectorTokenType) (SelectorToken, error) {
	tok, err := sts.Next()
	if errors.Is(err, io.EOF) {
		return tok, io.ErrUnexpectedEOF
	} else if err != nil {
		return tok, fmt.Errorf("unexpected error after %s: %w", sts.lastToken.Type, err)
	}
	if tok.Type != typ {
		return tok, fmt.Errorf("expected %s after %s, got: %q", typ, sts.lastToken.Type, string(tok.Raw))
	}
	return tok, nil
}

func (sts selectorTokenStream) Next() (tok SelectorToken, err error) {
	defer func() {
		sts.lastToken = tok
	}()
	b1, err := sts.r.ReadByte()
	if err != nil {
		return SelectorToken{}, fmt.Errorf("failed to read: %w", err)
	}
	tok = SelectorToken{
		Raw:  []byte{b1},
		Type: SelectorTokenTypeIdentifier,
	}
	switch b1 {
	case '*':
		tok.Type = SelectorTokenTypeGlobal
	case '.':
		tok.Type = SelectorTokenTypeClassStart
	case ':':
		tok.Type = SelectorTokenTypePseudoClassStart
	case '[':
		tok.Type = SelectorTokenTypeAttributeStart
	case ']':
		tok.Type = SelectorTokenTypeAttributeStop
	case '#':
		tok.Type = SelectorTokenTypeIDStart
	case ' ':
		tok.Type = SelectorTokenTypeDescendantStart
	case '(':
		tok.Type = SelectorTokenTypeSubSelectorStart
	case ')':
		tok.Type = SelectorTokenTypeSubSelectorStop
	case '>':
		tok.Type = SelectorTokenTypeChildStart
	case '+':
		tok.Type = SelectorTokenTypeNextSiblingStart
	case '~':
		tok.Type = SelectorTokenTypeSubsequentSiblingStart
	case '|':
		tok.Type = SelectorTokenTypeOr
	case '\r', '\n', '\t':
		return tok, ErrInvalidSelector
	}
	if tok.Type != SelectorTokenTypeIdentifier {
		return tok, nil
	}
	// We've now determined this is an identifier
	// Identifiers cannot begin with numbers (css3-modsel-155)
	if unicode.IsDigit(rune(b1)) {
		return tok, ErrInvalidSelector
	}
	for {
		b, err := sts.r.ReadByte()
		if err == io.EOF {
			return tok, nil
		}
		if err != nil && err != io.EOF {
			return SelectorToken{}, fmt.Errorf("failed to read: %w", err)
		}
		switch b {
		case '\r', '\n', '\t':
			return tok, ErrInvalidSelector
		case '*', '.', ':', '[', ']', '#', ' ', '(', ')', '>', '+', '~', '|':
			sts.r.UnreadByte()
			return tok, nil
		}
		tok.Raw = append(tok.Raw, b)
	}
}
