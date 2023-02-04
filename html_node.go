package hcj

import (
	"fmt"
	"sort"

	"golang.org/x/net/html"
)

// ParsedNode is the internal hcj representation of an html node.
type ParsedNode struct {
	Raw              *html.Node
	Tag              string
	ID               string
	Classes          []string
	Style            map[string]string
	PseudoClassStyle map[PseudoClass]map[string]string
	// PseudoClassSupers are PseudoClasses with parameters like lang(in)
	// they are not supported yet
	PseudoClassSuperStyle map[string]map[string]string
	FirstChild            *ParsedNode
	// LastChild
	NextSibling *ParsedNode
	// LastSibling
	// Parent *ParsedNode
}

func (pn *ParsedNode) String() string {

	return pn.NestedString(0)
}

// NestedString is a convience method for attempting to pretty print a ParsedNode
func (pn *ParsedNode) NestedString(indent int) string {
	out := ""
	nest := ""
	startLine := ""
	if indent > 0 {
		nest = "\n|"
		startLine = "-"
		for i := 0; i < indent; i++ {
			nest += " "
			startLine += "-"
		}
	}
	if pn.Tag != "" {
		out += nest + fmt.Sprintf("Tag:'%s' ", pn.Tag)
	}
	if pn.Raw != nil {
		out += nest + fmt.Sprintf("SubNodeType:'%v' ", nodeTypeString(int(pn.Raw.Type)))
	}
	if pn.ID != "" {
		out += nest + fmt.Sprintf("ID:'%s' ", pn.ID)
	}
	if len(pn.Classes) != 0 {
		out += nest + fmt.Sprintf("Classes: %v ", pn.Classes)
	}
	if len(pn.Style) != 0 {
		out += nest + fmt.Sprintf("Style: %v ", pn.Style)
	}
	if pn.FirstChild != nil {
		out += nest + fmt.Sprintf("First: {%v} ", pn.FirstChild.NestedString(indent+1))
	}
	if pn.NextSibling != nil {
		out += nest + fmt.Sprintf("Next: {%v} ", pn.NextSibling.NestedString(indent))
	}
	if len(out) == 0 {
		return out
	}

	return "[\n" + startLine + out + "]"
}

func nodeTypeString(enumT int) string {
	switch enumT {
	case 0:
		return "error"
	case 1:
		return "text"
	case 2:
		return "document"
	case 3:
		return "element"
	case 4:
		return "comment"
	case 5:
		return "doc"
	case 6:
		return "raw"
	}
	return "unknown"
}

type ParseNodeOptions struct {
	CSS         CSS
	ParentStyle map[string]string
}

func ParseNode(node *html.Node, opts ...ParseNodeOption) *ParsedNode {
	cfg := ParseNodeOptions{}
	for _, o := range opts {
		cfg = o(cfg)
	}
	pn := &ParsedNode{
		Raw:     node,
		ID:      getAttribute(node, "id"),
		Tag:     node.Data,
		Classes: getAttributes(node, "class"),
		Style:   make(map[string]string),
	}
	// todo: what properties are inherited?
	// find a table
	inheritedProps := map[string]struct{}{
		"color":            {},
		"opacity":          {},
		"background":       {},
		"background-color": {},
		"font-style":       {},
	}
	pn.CalculateStyle(cfg.CSS)
INHERIT_LOOP:
	for k, v := range cfg.ParentStyle {
		if _, ok := inheritedProps[k]; ok {
			switch pn.Style[k] {
			case "inherit":
			case "currentColor":
				if k != "color" {
					continue INHERIT_LOOP
				}
			case "":
			default:
				continue INHERIT_LOOP
			}
			pn.Style[k] = v
		}
	}
	if node.FirstChild != nil {
		pn.FirstChild = ParseNode(node.FirstChild, WithCSS(cfg.CSS), WithParentStyle(pn.Style))
		// pn.FirstChild.setParent(pn)
	}
	if node.NextSibling != nil {
		pn.NextSibling = ParseNode(node.NextSibling, WithCSS(cfg.CSS), WithParentStyle(cfg.ParentStyle))
	}
	return pn
}

// setParent reference on the node.
// This may be removable in the future if the full set of inheritance is understood.
// May be needed to determine parent node type and what implicit changes that poses.
// func (pn *ParsedNode) setParent(parent *ParsedNode) {
// 	pn.Parent = parent
// 	if pn.NextSibling != nil {
// 		pn.NextSibling.setParent(parent)
// 	}
// }

// styleWithPriority contains a decided on style
type styleWithPriority struct {
	priority int16 // max ~1000
	style    map[string]string
}

func (pn *ParsedNode) CalculateStyle(css CSS) {
	styles := []styleWithPriority{}
	for selStr, style := range css.Selectors {
		sel, err := ParseSelector(selStr)
		if err != nil {
			continue
		}
		priority := int16(0)
		if sel.Global {
			styles = append(styles, styleWithPriority{
				priority: priority,
				style:    style,
			})
			continue
		}
		for _, id := range sel.IDs {
			if pn.ID == id {
				priority += 100
			} else {
				priority = -1000
			}
		}
		if sel.Tag == pn.Tag {
			priority += 1
		}
		for _, c2 := range sel.Classes {
			match := false
			for _, c := range pn.Classes {
				if c == c2 {
					match = true
					priority += 10
					break
				}
			}
			if !match {
				priority = -1000
			}
		}
		if sel.Attribute != "" {
			matcher, err := ParseAttributeSelector(sel.Attribute)
			if err != nil {
				continue
			}
			if matcher.Match(pn) {
				priority += 10
			} else {
				priority = -1000
			}
		}
		if priority > 0 {
			styles = append(styles, styleWithPriority{
				priority: priority,
				style:    style,
			})
		}
	}
	sort.Slice(styles, func(i, j int) bool {
		return styles[i].priority < styles[j].priority
	})
	for _, style := range styles {
		pn.Style = mergeMaps(pn.Style, style.style)
	}

	// inline always has highest priority
	if style := getMapAttribute(pn.Raw, "style"); style != nil {
		pn.Style = mergeMaps(pn.Style, style)
	}
	// !important, not supported
}
