package hcj

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type PseudoClass uint8

// PseudoClasses
const (
	PseudoClassActive      PseudoClass = iota
	PseudoClassChecked     PseudoClass = iota
	PseudoClassDisabled    PseudoClass = iota
	PseudoClassEmpty       PseudoClass = iota
	PseudoClassEnabled     PseudoClass = iota
	PseudoClassFirstChild  PseudoClass = iota
	PseudoClassFirstOfType PseudoClass = iota
	PseudoClassFocus       PseudoClass = iota
	PseudoClassHover       PseudoClass = iota
	PseudoClassInRange     PseudoClass = iota
	PseudoClassInvalid     PseudoClass = iota
	PseudoClassLastChild   PseudoClass = iota
	PseudoClassLastOfType  PseudoClass = iota
	PseudoClassLink        PseudoClass = iota
	PseudoClassOnlyOfType  PseudoClass = iota
	PseudoClassOnlyChild   PseudoClass = iota
	PseudoClassOptional    PseudoClass = iota
	PseudoClassOutOfRange  PseudoClass = iota
	PseudoClassReadOnly    PseudoClass = iota
	PseudoClassReadWrite   PseudoClass = iota
	PseudoClassRequired    PseudoClass = iota
	PseudoClassRoot        PseudoClass = iota
	PseudoClassTarget      PseudoClass = iota
	PseudoClassValid       PseudoClass = iota
	PseudoClassVisited     PseudoClass = iota
)

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
	Tag       string
	ID        string // Multiple IDs are supported, but we lack an example where they would be used. Styles with multiple IDs should be discarded for now.
	Attribute string
	Class     string // Multiple classes are not supported see css3-modsel-14b
	Global    bool
}

var ErrInvalidSelector = fmt.Errorf("invalid selector")

func ParseSelector(s string) (Selector, error) {
	sel := Selector{}
	if s == "*" {
		return Selector{
			Global: true,
		}, nil
	}
	// utf8?
	var next []rune
	var nextIsID, nextIsClass, nextIsAttribute bool
	for _, c := range s {
		switch c {
		case '[':
			// TODO: are some of these conditions workable?
			if nextIsID || nextIsAttribute || nextIsClass || len(next) == 0 {
				return sel, ErrInvalidSelector
			}
			sel.Tag = string(next)
			next = []rune{}
			nextIsAttribute = true
		case ']':
			if !nextIsAttribute || len(next) == 0 {
				return sel, ErrInvalidSelector
			}
			sel.Attribute = string(next)
			next = []rune{}
			nextIsAttribute = false
		case '#':
			if sel.ID != "" || nextIsID {
				return sel, ErrInvalidSelector
			}
			if len(next) == 0 && (nextIsClass || nextIsID) {
				// invalid .# or ##
				return sel, ErrInvalidSelector
			} else if len(next) != 0 {
				if nextIsClass {
					sel.Class = string(next)
					next = []rune{}
				} else {
					sel.Tag = string(next)
					next = []rune{}
				}
			}
			nextIsClass = false
			nextIsID = true
		case '.':
			if nextIsClass || sel.Class != "" {
				return sel, ErrInvalidSelector
			}
			if len(next) == 0 && (nextIsClass || nextIsID) {
				// invalid #. or ..
				return sel, ErrInvalidSelector
			} else if len(next) != 0 {
				if nextIsClass {
					sel.Class = string(next)
					next = []rune{}
				} else if nextIsID {
					sel.ID = string(next)
					next = []rune{}
				} else {
					sel.Tag = string(next)
					next = []rune{}
				}
			}
			nextIsID = false
			nextIsClass = true
		case ' ', '\n', '\t', '\r':
			return sel, ErrInvalidSelector
		default:
			// utf8?
			if c > unicode.MaxASCII {
				return sel, ErrInvalidSelector
			}
			next = append(next, c)
		}
	}
	if len(next) != 0 {
		if nextIsAttribute {
			return sel, ErrInvalidSelector
		}
		if nextIsClass {
			sel.Class = string(next)
		} else if nextIsID {
			sel.ID = string(next)
		} else {
			sel.Tag = string(next)
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
