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
			// some pseudo classes have sub selectors
			pc := PseudoClass{
				Type: stringToPseudoClassType(string(id.Raw)),
			}
			if pc.Type == PseudoClassTypeUnknown {
				return sel, fmt.Errorf("unknown pseudo-class %q", string(id.Raw))
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
		parsedDefOrder, parsedDef := parseSemiSeparatedMapWithOrder(def)
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
			for _, k := range parsedDefOrder {
				v := parsedDef[k]

				// validate directive
				// per https://www.w3.org/TR/css-color-3/#rgb-def which references https://www.w3.org/TR/2011/REC-CSS2-20110607/
				// which actually no longer contains the info as its been moved to it's errata https://www.w3.org/Style/css2-updates/REC-CSS2-20110607-errata.html
				// we have to validate whether the rule is valid prior to applying it. See t040204-hsl-parsing-f.htm

				if !isValidKnownCss(k, v) {
					continue
				}

				c.Selectors[selectorStr][k] = v

				// TODO: order dependant actions need to occur here
				if proc, ok := compositeAttributeMapping[k]; ok {
					proc(c.Selectors[selectorStr], k, v)
				}
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

type procesor func(map[string]string, string, string)

// map of composite attributes and what they should equate to
var compositeAttributeMapping = map[string]procesor{
	"border-top": decomposeBorder, "border-right": decomposeBorder, "border-bottom": decomposeBorder, "border-left": decomposeBorder, "border": decomposeBorder,
	"border-color": allDirLikeBorder, "border-style": allDirLikeBorder, "border-size": allDirLikeBorder,
	"margin": allDirLikeBorder, "padding": allDirLikeBorder,
}

func decomposeBorder(styles map[string]string, key, val string) {

	// if key == "border" --> break into alldir like border
	if key == "border" {
		decomposeBorder(styles, fmt.Sprintf("%s-top", key), val)
		decomposeBorder(styles, fmt.Sprintf("%s-bottom", key), val)
		decomposeBorder(styles, fmt.Sprintf("%s-left", key), val)
		decomposeBorder(styles, fmt.Sprintf("%s-right", key), val)
		return
	}

	// here we inspect the val to determine what sub elements we are setting
	valueParts := strings.Fields(val)
	for _, valPart := range valueParts {
		if isABorderWidth(valPart) {
			setStyleIfIsValid(styles, fmt.Sprintf("%s-width", key), valPart)
			continue
		}
		if _, ok := bStyles[valPart]; ok {
			setStyleIfIsValid(styles, fmt.Sprintf("%s-style", key), valPart)
			continue
		}

		setStyleIfIsValid(styles, fmt.Sprintf("%s-color", key), valPart)
	}

}

// allDirLikeBorder sets all directions like how margina and border do it. Namely following the below lines directive
// If there is only one component value, it applies to all sides. If there are two values, the top and bottom margins are set to the first value and the right and left margins are set to the second. If there are three values, the top is set to the first value, the left and right are set to the second, and the bottom is set to the third. If there are four values, they apply to the top, right, bottom, and left, respectively.
func allDirLikeBorder(styles map[string]string, key, val string) {
	if !isValidKnownCss(key, val) {
		return
	}
	// if multi part such as border-style then it becomes border-<direction>-style
	keyParts := strings.Split(key, "-")
	postDirectionPiece := ""
	if len(keyParts) != 1 {
		postDirectionPiece = fmt.Sprintf("-%s", strings.Join(keyParts[1:], "-"))
	}
	top := fmt.Sprintf("%s-top%s", keyParts[0], postDirectionPiece)
	bottom := fmt.Sprintf("%s-bottom%s", keyParts[0], postDirectionPiece)
	left := fmt.Sprintf("%s-left%s", keyParts[0], postDirectionPiece)
	right := fmt.Sprintf("%s-right%s", keyParts[0], postDirectionPiece)

	valueParts := strings.Fields(val)

	switch len(valueParts) {
	case 1:
		styles[top] = valueParts[0]
		styles[bottom] = valueParts[0]
		styles[left] = valueParts[0]
		styles[right] = valueParts[0]
	case 2:
		styles[top] = valueParts[0]
		styles[bottom] = valueParts[0]
		styles[left] = valueParts[1]
		styles[right] = valueParts[1]
	case 3:
		styles[top] = valueParts[0]
		styles[bottom] = valueParts[2]
		styles[left] = valueParts[1]
		styles[right] = valueParts[1]
	case 4:
		styles[top] = valueParts[0]
		styles[right] = valueParts[1]
		styles[bottom] = valueParts[2]
		styles[left] = valueParts[3]

	default:
		// TODO: Don't just print. decide on a nicer way to bubble. Also should this count as a parse error?
		fmt.Println("detected invalid len for ", key, "with value", val, "and len", len(valueParts))
	}
}

var bStyles = map[string]struct{}{"none": struct{}{},
	"hidden": struct{}{}, "dotted": struct{}{},
	"dashed": struct{}{}, "solid": struct{}{},
	"double": struct{}{}, "groove": struct{}{},
	"ridge": struct{}{}, "inset": struct{}{},
	"outset": struct{}{},
}
var bWidths = map[string]struct{}{"thin": struct{}{}, "medium": struct{}{}, "thick": struct{}{}}

func isABorderWidth(part string) bool {
	_, err := strconv.Atoi(part)
	if err == nil {
		return true
	}
	_, ok := bWidths[part]
	return ok
}

func setStyleIfIsValid(styles map[string]string, key, val string) {
	if !isValidKnownCss(key, val) {
		return
	}
	styles[key] = val
}

// isValidKnownCss probably shouldnt exist as we are now duplicating work over in other places where we do final parsing
// adding a stopgap until we have a better understanding of the scope of things that need validation
func isValidKnownCss(key, val string) bool {
	ok := true // pretend that things are ok unless we know otherwise. Technically against the spirit of css
	switch key {
	case "color", "background-color", "border-color":
		_, ok, _ = parseHTMLColor(val)
	}

	return ok
}
