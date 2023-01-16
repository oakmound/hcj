package hcj

import (
	"sort"

	"golang.org/x/net/html"
)

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
	//Parent *ParsedNode
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
		"color":      {},
		"opacity":    {},
		"background": {},
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
			default:
				continue INHERIT_LOOP
			}
			pn.Style[k] = v
		}
	}
	if node.FirstChild != nil {
		pn.FirstChild = ParseNode(node.FirstChild, WithCSS(cfg.CSS), WithParentStyle(pn.Style))
		//pn.FirstChild.setParent(pn)
	}
	if node.NextSibling != nil {
		pn.NextSibling = ParseNode(node.NextSibling, WithCSS(cfg.CSS))
	}
	return pn
}

// func (pn *ParsedNode) setParent(parent *ParsedNode) {
// 	pn.Parent = parent
// 	if pn.NextSibling != nil {
// 		pn.NextSibling.setParent(parent)
// 	}
// }

type styleWithPriority struct {
	priority uint16 // max ~1000
	style    map[string]string
}

func (pn *ParsedNode) CalculateStyle(css CSS) {
	styles := []styleWithPriority{}
	for selStr, style := range css.Selectors {
		sel, err := ParseSelector(selStr)
		if err != nil {
			continue
		}
		priority := uint16(0)
		if sel.Global {
			styles = append(styles, styleWithPriority{
				priority: priority,
				style:    style,
			})
			continue
		}
		if sel.ID != "" && sel.ID == pn.ID {
			priority += 100
		}
		if sel.Tag == pn.Tag {
			priority += 1
		}
		for _, c := range pn.Classes {
			for _, c2 := range sel.Classes {
				if c == c2 {
					priority += 10
				}
			}
		}
		if sel.Attribute != "" {
			matcher, err := ParseAttributeSelector(sel.Attribute)
			if err != nil {
				continue
			}
			if matcher.Match(pn) {
				priority += 10
			}
		}
		if priority != 0 {
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
