package hcj

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type PseudoClassType uint8

// PseudoClasses
const (
	PseudoClassTypeUnknown       PseudoClassType = iota
	PseudoClassTypeActive        PseudoClassType = iota
	PseudoClassTypeChecked       PseudoClassType = iota
	PseudoClassTypeDisabled      PseudoClassType = iota
	PseudoClassTypeEmpty         PseudoClassType = iota
	PseudoClassTypeEnabled       PseudoClassType = iota
	PseudoClassTypeFirstChild    PseudoClassType = iota
	PseudoClassTypeFirstOfType   PseudoClassType = iota
	PseudoClassTypeFocus         PseudoClassType = iota
	PseudoClassTypeHover         PseudoClassType = iota
	PseudoClassTypeInRange       PseudoClassType = iota
	PseudoClassTypeInvalid       PseudoClassType = iota
	PseudoClassTypeLang          PseudoClassType = iota
	PseudoClassTypeLastChild     PseudoClassType = iota
	PseudoClassTypeLastOfType    PseudoClassType = iota
	PseudoClassTypeLink          PseudoClassType = iota
	PseudoClassTypeNot           PseudoClassType = iota
	PseudoClassTypeNthChild      PseudoClassType = iota
	PseudoClassTypeNthLastChild  PseudoClassType = iota
	PseudoClassTypeNthLastOfType PseudoClassType = iota
	PseudoClassTypeNthOfType     PseudoClassType = iota
	PseudoClassTypeOnlyOfType    PseudoClassType = iota
	PseudoClassTypeOnlyChild     PseudoClassType = iota
	PseudoClassTypeOptional      PseudoClassType = iota
	PseudoClassTypeOutOfRange    PseudoClassType = iota
	PseudoClassTypeReadOnly      PseudoClassType = iota
	PseudoClassTypeReadWrite     PseudoClassType = iota
	PseudoClassTypeRequired      PseudoClassType = iota
	PseudoClassTypeRoot          PseudoClassType = iota
	PseudoClassTypeTarget        PseudoClassType = iota
	PseudoClassTypeValid         PseudoClassType = iota
	PseudoClassTypeVisited       PseudoClassType = iota
	// TODO: pseudo elements
)

type PseudoClass struct {
	Type        PseudoClassType
	SubSelector Selector
}

func stringToPseudoClassType(s string) PseudoClassType {
	switch s {
	case "active":
		return PseudoClassTypeActive
	case "checked":
		return PseudoClassTypeChecked
	case "disabled":
		return PseudoClassTypeDisabled
	case "empty":
		return PseudoClassTypeEmpty
	case "enabled":
		return PseudoClassTypeEnabled
	case "first-child":
		return PseudoClassTypeFirstChild
	case "first-of-type":
		return PseudoClassTypeFirstOfType
	case "focus":
		return PseudoClassTypeFocus
	case "hover":
		return PseudoClassTypeHover
	case "in-rage":
		return PseudoClassTypeInRange
	case "invalid":
		return PseudoClassTypeInvalid
	case "lang":
		return PseudoClassTypeLang
	case "last-child":
		return PseudoClassTypeLastChild
	case "last-of-type":
		return PseudoClassTypeLastOfType
	case "link":
		return PseudoClassTypeLink
	case "not":
		return PseudoClassTypeNot
	case "nth-child":
		return PseudoClassTypeNthChild
	case "nth-last-child":
		return PseudoClassTypeNthLastChild
	case "nth-last-of-type":
		return PseudoClassTypeNthLastOfType
	case "nth-of-type":
		return PseudoClassTypeNthOfType
	case "only-of-type":
		return PseudoClassTypeOnlyOfType
	case "only-child":
		return PseudoClassTypeOnlyChild
	case "optional":
		return PseudoClassTypeOptional
	case "out-of-range":
		return PseudoClassTypeOutOfRange
	case "read-only":
		return PseudoClassTypeReadOnly
	case "read-write":
		return PseudoClassTypeReadWrite
	case "required":
		return PseudoClassTypeRequired
	case "root":
		return PseudoClassTypeRoot
	case "target":
		return PseudoClassTypeTarget
	case "valid":
		return PseudoClassTypeValid
	case "visited":
		return PseudoClassTypeVisited
	default:
		return PseudoClassTypeUnknown
	}
}

// we don't support !important

type CSS struct {
	Selectors map[string]map[string]string
}

// Merge combines two CSS maps, preferring settings in c2 over c. c is modified as a result.
func (c CSS) Merge(c2 CSS) CSS {
	for sel, decl := range c2.Selectors {
		if c.Selectors[sel] == nil {
			c.Selectors[sel] = make(map[string]string)
		}
		for prop, val := range decl {
			c.Selectors[sel][prop] = val
		}
	}
	return c
}

type Selector struct {
	Tag           string
	IDs           []string
	Attribute     string
	Classes       []string
	Global        bool
	PseudoClasses []PseudoClass
}

var ErrInvalidSelector = fmt.Errorf("invalid selector")

func ParseSelector(s string) (Selector, error) {
	sel := Selector{}
	tkz := TokenizeSelector(s)
	i := 0
	for {
		i++
		tok, err := tkz.Next()
		//fmt.Println(tok, err, tok.Type.String(), string(tok.Raw))
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return sel, err
		}
		if i == 1 && tok.Type == SelectorTokenTypeIdentifier {
			sel.Tag = string(tok.Raw)
			continue
		}
		switch tok.Type {
		case SelectorTokenTypeIdentifier:
			return sel, fmt.Errorf("impossible: identifier with no qualifier %q", string(tok.Raw))
		case SelectorTokenTypeGlobal:
			if i == 1 {
				sel.Global = true
			} else {
				return sel, fmt.Errorf("invalid: global * after selector start")
			}
		case SelectorTokenTypeIDStart:
			id, err := tkz.ExpectNext(SelectorTokenTypeIdentifier)
			if err != nil {
				return sel, err
			}
			sel.IDs = append(sel.IDs, string(id.Raw))
		case SelectorTokenTypeClassStart:
			id, err := tkz.ExpectNext(SelectorTokenTypeIdentifier)
			if err != nil {
				return sel, err
			}
			sel.Classes = append(sel.Classes, string(id.Raw))
		case SelectorTokenTypeAttributeStart:
			if sel.Attribute != "" {
				return sel, fmt.Errorf("duplicate attribute definition")
			}
			attr := []byte{}
			for {
				// attributes may contain characters normally associated with
				// non-identifier selector types e.g. *; treat as a comment
				// and grab the entire contents to be parsed later
				t2, err := tkz.Next()
				if err != nil {
					return sel, fmt.Errorf("error parsing attribute: %w", err)
				}
				if t2.Type == SelectorTokenTypeAttributeStop {
					break
				}
				attr = append(attr, t2.Raw...)
			}
			sel.Attribute = string(attr)
		case SelectorTokenTypeAttributeStop:
			return sel, fmt.Errorf("attribute stop without matching start")
		case SelectorTokenTypePseudoClassStart:
			id, err := tkz.ExpectNext(SelectorTokenTypeIdentifier)
			if err != nil {
				if id.Type == SelectorTokenTypePseudoClassStart {
					// pseudo element, OK, but not supported yet
					return sel, fmt.Errorf("pseudo elements are unimplemented")
				} else {
					return sel, err
				}
			}
			// some psuedo classes have sub selectors
			pc := PseudoClass{
				Type: stringToPseudoClassType(string(id.Raw)),
			}
			switch string(id.Raw) {
			case "lang", "not", "nth-child", "nth-last-child", "nth-last-of-type", "nth-of-type":
				_, err := tkz.ExpectNext(SelectorTokenTypeSubSelectorStart)
				if err != nil {
					return sel, err
				}
				subSel := []byte{}
				for {
					// TODO: this does not support nested sub selectors
					t2, err := tkz.Next()
					if err != nil {
						return sel, fmt.Errorf("error parsing sub selector: %w", err)
					}
					if t2.Type == SelectorTokenTypeSubSelectorStop {
						break
					}
					subSel = append(subSel, t2.Raw...)
				}
				parsed, err := ParseSelector(string(subSel))
				if err != nil {
					return sel, fmt.Errorf("error parsing sub selector: %w", err)
				}
				pc.SubSelector = parsed
			default:
				// no sub selector
			}
			sel.PseudoClasses = append(sel.PseudoClasses, pc)
		case SelectorTokenTypeSubSelectorStart:
			return sel, fmt.Errorf("sub-selector start outside of pseudo-class")
		case SelectorTokenTypeSubSelectorStop:
			return sel, fmt.Errorf("sub-selector stop outside of pseudo-class")
		case SelectorTokenTypeDescendantStart:
			return sel, fmt.Errorf("descendants are unimplemented")
		case SelectorTokenTypeChildStart:
			return sel, fmt.Errorf("children are unimplemented")
		case SelectorTokenTypeNextSiblingStart:
			return sel, fmt.Errorf("next siblings are unimplemented")
		case SelectorTokenTypeSubsequentSiblingStart:
			return sel, fmt.Errorf("subsequent siblings are unimplemented")
		}
	}
	return sel, nil
}

func ParseCSS(s string) CSS {
	// ignore invalid selectors
	c := CSS{
		Selectors: make(map[string]map[string]string),
	}
	replaceWhiteSpace := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "", "{", "")
	for {
		selectorSplit := strings.SplitAfterN(s, "{", 2)
		if len(selectorSplit) <= 1 {
			break
		}
		endDef := strings.Index(selectorSplit[1], "}")
		if endDef == -1 {
			break
		}
		s = selectorSplit[1][endDef+1:]
		selectorStr := replaceWhiteSpace.Replace(selectorSplit[0])
		_, err := ParseSelector(selectorStr)
		if err != nil {
			continue
		}
		def := selectorSplit[1][:endDef]
		def = strings.Trim(def, "{} \n\t\r")
		parsedDef := parseSemiSeparatedMap(def)
		// validation
		for k, v := range parsedDef {
			if k == "color" || k == "background-color" || k == "background" {
				_, ok, inheritable := parseHTMLColor(v)
				if !ok && !inheritable {
					delete(parsedDef, k)
				}
			}
		}
		selectorKeys := strings.Split(selectorStr, ",")
		for _, selectorStr := range selectorKeys {
			for {
				i := strings.Index(selectorStr, "/*")
				if i == -1 {
					break
				}
				j := strings.Index(selectorStr, "*/")
				selectorStr = selectorStr[:i] + selectorStr[j+2:]
			}
			selectorStr = strings.TrimSpace(selectorStr)
			if c.Selectors[selectorStr] == nil {
				c.Selectors[selectorStr] = make(map[string]string)
			}
			for k, v := range parsedDef {
				c.Selectors[selectorStr][k] = v
			}
		}
	}
	return c
}

type AttributeSelector interface {
	Match(pn *ParsedNode) bool
}

type attributeSelectorFn struct {
	matchFn func(*ParsedNode) bool
}

func (asf *attributeSelectorFn) Match(pn *ParsedNode) bool {
	return asf.matchFn(pn)
}

func ParseAttributeSelector(selector string) (AttributeSelector, error) {
	var match func(*ParsedNode) bool
	if i := strings.Index(selector, "="); i != -1 {
		attr, val := selector[:i], selector[i+1:]
		if len(attr) == 0 || len(val) == 0 {
			return nil, ErrInvalidSelector
		}
		if v, err := strconv.Unquote(val); err == nil {
			val = v
		}

		switch attr[len(attr)-1] {
		case '*':
			match = func(pn *ParsedNode) bool {
				a := getAttribute(pn.Raw, attr[:len(attr)-1])
				if len(a) == 0 {
					return false
				}
				if strings.Contains(a, val) {
					return true
				}
				return false
			}
		case '^':
			match = func(pn *ParsedNode) bool {
				a := getAttribute(pn.Raw, attr[:len(attr)-1])
				if len(a) == 0 {
					return false
				}
				if strings.HasPrefix(a, val) {
					return true
				}
				return false
			}
		case '$':
			match = func(pn *ParsedNode) bool {
				a := getAttribute(pn.Raw, attr[:len(attr)-1])
				if len(a) == 0 {
					return false
				}
				if strings.HasSuffix(a, val) {
					return true
				}
				return false
			}
		case '|':
			match = func(pn *ParsedNode) bool {
				a := getAttribute(pn.Raw, attr[:len(attr)-1])
				if len(a) == 0 {
					return false
				}
				if a == val {
					return true
				}
				if strings.HasPrefix(a, val+"-") {
					return true
				}
				return false
			}
		case '~':
			match = func(pn *ParsedNode) bool {
				a := getAttribute(pn.Raw, attr[:len(attr)-1])
				if len(a) == 0 {
					return false
				}
				splitA := strings.Split(a, " ")
				for _, a := range splitA {
					if a == val {
						return true
					}
				}
				return false
			}
		default:
			match = func(pn *ParsedNode) bool {
				return getAttribute(pn.Raw, attr) == val
			}
		}

	} else {
		match = func(pn *ParsedNode) bool {
			return getAttribute(pn.Raw, selector) != ""
		}
	}
	return &attributeSelectorFn{
		matchFn: match,
	}, nil
}
