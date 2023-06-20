package hcj

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/alg/intgeom"
	"github.com/oakmound/oak/v4/render"
)

// RenderHTML as a sprite with the provided dimensions
func RenderHTML(htmlReader io.Reader, dims intgeom.Point2) (*render.Sprite, error) {
	_, sp, err := ParseAndRenderHTML(htmlReader, dims)
	return sp, err
}

func (p *ParsedNode) Draw(buff draw.Image, xOff, yOff float64) {
	// according to box model , first determine padding, border and last apply margin to the zone
	// https://www.w3.org/TR/CSS2/box.html#box-dimensions

	// empty html is blank white
	var bkgColor color.Color = color.RGBA{255, 255, 255, 255}

	// TODO: what to do if both of these are set?
	if col, ok := p.Style["background-color"]; ok {
		bkgColor, _, _ = parseHTMLColor(col)
	} else if col, ok := p.Style["background"]; ok {
		bkgColor, _, _ = parseHTMLColor(col)
	}

	bds := buff.Bounds()

	stack := render.NewDrawStack(
		render.NewDynamicHeap(),
		render.NewDynamicHeap(),
	)
	bkg := render.NewColorBoxR(bds.Dx(), bds.Dy(), bkgColor)
	stack.Draw(bkg, 0, -1)

	// remove space that is used for margin
	fullBodyMargin := parseMargin(p.Style)
	bodyMargin := fullBodyMargin.Min
	zone := floatgeom.Rect2{
		Min: bodyMargin,
		Max: floatgeom.Point2{float64(bds.Dx()), float64(bds.Dy())}.Sub(bodyMargin),
	}

	trackingStack := &trackingDrawStack{
		DrawStack: stack,
	}

	renderNode(p.FirstChild, trackingStack, zone, p.State)

	// slap it all onto the background
	stack.PreDraw()
	stack.DrawToScreen(buff, &intgeom.Point2{0, 0}, bds.Dx(), bds.Dy())
}

// ParseAndRenderHTML outputting the internal Node representation along with a sprite that has
// the given dimensions.
func ParseAndRenderHTML(htmlReader io.Reader, dims intgeom.Point2) (*ParsedNode, *render.Sprite, error) {
	parsed, err := ParseHTML(htmlReader)
	if err != nil || parsed == nil {
		return parsed, nil, err
	}

	// TODO: respect the width from css
	sp := render.NewEmptySprite(0, 0, dims.X(), dims.Y())
	parsed.Draw(sp.GetRGBA(), 0, 0)

	return parsed, sp, err
}

func ParseHTML(htmlReader io.Reader, opts ...ParseNodeOption) (*ParsedNode, error) {
	rootNode, err := html.Parse(htmlReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	// body style background-color
	var css CSS
	if headNode := findHTMLNode(rootNode, "head"); headNode != nil {
		if styleNode := findHTMLNode(headNode, "style"); styleNode != nil {
			if styleNode.FirstChild != nil {
				css = ParseCSS(string(styleNode.FirstChild.Data))
			}
		}
	}
	css = DefaultCSS().Merge(css)
	opts = append(opts, WithCSS(css))
	var bn *ParsedNode
	if bodyNode := findHTMLNode(rootNode, "body"); bodyNode != nil {
		bn = ParseNode(bodyNode, opts...)
	}

	return bn, nil
}

func findHTMLNode(root *html.Node, name string) *html.Node {
	scan := []*html.Node{root}
	for len(scan) > 0 {
		next := scan[0]
		scan = scan[1:]
		if next.Data == name {
			return next
		}

		if next.FirstChild != nil {
			scan = append(scan, next.FirstChild)
		}
		if next.NextSibling != nil {
			scan = append(scan, next.NextSibling)
		}
	}
	return nil
}

func getAttributes(node *html.Node, key string) []string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return strings.Split(attr.Val, " ")
		}
	}
	return []string{}
}

func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getMapAttribute(node *html.Node, key string) map[string]string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return parseSemiSeparatedMap(attr.Val)
		}
	}
	return nil
}

func parseSemiSeparatedMap(data string) map[string]string {
	_, content := parseSemiSeparatedMapWithOrder(data)
	return content
}

func parseSemiSeparatedMapWithOrder(data string) ([]string, map[string]string) {
	// todo: how often do we need to strip spaces from these values
	vals := strings.Split(data, ";")
	out := make(map[string]string, len(vals))
	outOrder := make([]string, 0, len(vals))
	for _, v := range vals {
		splitVal := strings.Split(v, ":")
		if len(splitVal) != 2 {
			// ignore bad formatted thing
			continue
		}
		k := strings.TrimSpace(splitVal[0])
		out[k] = strings.TrimSpace(splitVal[1])
		outOrder = append(outOrder, k)
	}
	return outOrder, out
}

func parseLengthWithDefault(l string, def float64) float64 {
	v, ok := parseLength(l)
	if !ok {
		return def
	}
	return v
}

func parseLength(l string) (float64, bool) {
	if lpx := strings.TrimSuffix(l, "px"); lpx != l {
		length, err := strconv.Atoi(lpx)
		if err != nil {
			return 0, false
		}
		return float64(length), true
	}
	if lem := strings.TrimSuffix(l, "em"); lem != l {
		length, err := strconv.ParseFloat(lem, 64)
		if err != nil {
			return 0, false
		}
		return length * 16, true
	}
	// todo: others
	return 0, false
}

const maxSize = 3

// parseBorderAttributes and determine the drawing considerations for a given border line
// returns the computed size, color, style and optionally an error in the compute
func parseBorderAttributes(direction string, styles map[string]string) (int, color.Color, string, error) {

	// first see if we are taking from border or a sub value here
	width := styles[fmt.Sprintf("border-%s-width", direction)]

	colorString := styles[fmt.Sprintf("border-%s-color", direction)]

	bStyle := styles[fmt.Sprintf("border-%s-style", direction)]

	// Since the initial value of the border styles is 'none', no borders will be visible unless the border style is set.
	if bStyle == "" {
		bStyle = "none"
	}

	//  The interpretation of the first three values depends on the user agent

	size := 0
	switch width {
	case "thin":
		size = 1
	case "medium":
		size = 2
	case "thick":
		size = maxSize
	case "": // valid just not set so make size 0
		size = 0
	default:
		// attempt to parse as a non-negative int
		length, err := strconv.Atoi(width)
		if err != nil {
			return size, color.RGBA{0, 0, 0, 255}, bStyle, fmt.Errorf("invalid border length requested: %s", width)
		}
		size = length
	}

	// border-color
	parsedColor, _, inheritable := parseHTMLColor(colorString)
	if inheritable {
		parsedColor, _, _ = parseHTMLColor(styles["color"])
	}
	return size, parsedColor, bStyle, nil

}

func parseMargin(style map[string]string) floatgeom.Rect2 {
	// TODO: what do we do if both margin and margin-left/top/etc are set?
	// for now, the latter takes priority over the former
	margin := floatgeom.Rect2{}
	if m, ok := style["margin"]; ok {
		if m == "inherit" {
			// uh on
		} else if m == "initial" {
			// no!
		} else if m == "revert" {
			// ahh!
		} else if m == "revert-layer" {
			// what!?!?
		} else if m == "unset" {
			// oh god!
		} else {
			fds := strings.Fields(m)
			switch len(fds) {
			case 0:
				// ????????
			case 1:
				v, ok := parseLength(m)
				if ok {
					margin.Min = floatgeom.Point2{v, v}
					margin.Max = floatgeom.Point2{v, v}
				}
			case 2:
				vert, ok := parseLength(fds[0])
				if ok {
					margin.Min[1] = vert
					margin.Max[1] = vert
				}
				horz, ok := parseLength(fds[1])
				if ok {
					margin.Min[0] = horz
					margin.Max[0] = horz
				}
			case 3:
				top, ok := parseLength(fds[0])
				if ok {
					margin.Min[1] = top
				}
				horz, ok := parseLength(fds[1])
				if ok {
					margin.Min[0] = horz
					margin.Max[0] = horz
				}
				bot, ok := parseLength(fds[2])
				if ok {
					margin.Max[1] = bot
				}
			case 4:
				top, ok := parseLength(fds[0])
				if ok {
					margin.Min[1] = top
				}
				right, ok := parseLength(fds[1])
				if ok {
					margin.Max[0] = right
				}
				bot, ok := parseLength(fds[2])
				if ok {
					margin.Max[1] = bot
				}
				left, ok := parseLength(fds[1])
				if ok {
					margin.Min[0] = left
				}
			}
		}
	}
	if marginLeft, ok := parseLength(style["margin-left"]); ok {
		margin.Min[0] = marginLeft
	}
	if marginRight, ok := parseLength(style["margin-right"]); ok {
		margin.Max[0] = marginRight
	}
	if marginTop, ok := parseLength(style["margin-top"]); ok {
		margin.Min[1] = marginTop
	}
	if marginBottom, ok := parseLength(style["margin-bottom"]); ok {
		margin.Max[1] = marginBottom
	}
	return margin
}

func applyOpacity(parsedColor color.Color, opacity string) color.Color {
	if opacity != "" {
		opacityFloat, err := strconv.ParseFloat(opacity, 64)
		if err == nil {
			if opacityFloat < 0 {
				opacityFloat = 0
			} else if opacityFloat > 1 {
				opacityFloat = 1
			}
			// TODO: does the a here combine with global opacity or override it?
			r, g, b, _ := parsedColor.RGBA()
			parsedColor = color.NRGBA64{uint16(r), uint16(g), uint16(b), uint16(opacityFloat * 0xffff)}
		}
	}
	return parsedColor
}

// body default y buffer is 8 pix, x buffer ix 8 pix
// figure has default y buffer of 0, x buffer of 40

func parseNodeDims(node *ParsedNode, drawzone floatgeom.Rect2) floatgeom.Rect2 {
	w, _ := parseLength(node.Style["width"])
	h, _ := parseLength(node.Style["height"])
	// todo: percents
	if w == 0 {
		w = drawzone.W()
	}
	if h == 0 {
		h = drawzone.H()
	}

	pl, _ := parseLength(node.Style["padding-left"])
	pr, _ := parseLength(node.Style["padding-right"])

	pt, _ := parseLength(node.Style["padding-top"])
	pb, _ := parseLength(node.Style["padding-bottom"])

	w += pl + pr
	h += pt + pb

	return floatgeom.NewRect2WH(0, 0, w, h)
}

func drawboxModel(node *ParsedNode, stack *trackingDrawStack, drawzone floatgeom.Rect2, noTallerThan, noWiderThan float64) float64 {
	// TODO: Margin
	lateOffset := drawBorder(node, stack, drawzone, noTallerThan, noWiderThan)

	// TODO: padding
	drawBackground(node, stack, drawzone, noTallerThan, noWiderThan)
	return lateOffset[1]
}

// trackingDrawStack lets us consistently draw successive elements on top of each other,
// or in particular orders e.g. borders on top of backgrounds. This is not a correct implementation
// of html draw order, and is a stopgap.
type trackingDrawStack struct {
	nextMainLayer   int
	nextBorderLayer int
	*render.DrawStack
}

func (tds *trackingDrawStack) draw(r render.Renderable) {
	tds.DrawStack.Draw(r, 0, tds.nextMainLayer)
	tds.nextMainLayer++
}

func (tds *trackingDrawStack) drawBorder(r render.Renderable) {
	tds.DrawStack.Draw(r, 1, tds.nextBorderLayer)
	tds.nextBorderLayer++
}

func drawBackground(node *ParsedNode, stack *trackingDrawStack, drawzone floatgeom.Rect2, noTallerThan, noWiderThan float64) floatgeom.Point2 {
	bkg, ok := node.Style["background"]
	if !ok {
		bkg, ok = node.Style["background-color"]
		if !ok {
			return floatgeom.Point2{}
		}
	}

	// dont consume the child drawzone as we actually want to start at the default start spot
	_, offsetTop, offsetBot := applyPaddingForFinalOutput(node, drawzone)

	bkgDim := parseNodeDims(node, drawzone)
	// TODO: How do we actually handle no taller?
	if bkgDim.H() > noTallerThan {
		bkgDim.Max[1] = bkgDim.Min[1] + noTallerThan
	}
	if bkgDim.W() > noWiderThan {
		bkgDim.Max[0] = bkgDim.Min[0] + noWiderThan
	}
	// apply padding total overhead to the max to make sure the background covers all that it needs to
	bkgDim.Max = bkgDim.Max.Add(offsetTop, offsetBot)

	parsedColor, _, inheritable := parseHTMLColor(bkg)
	if inheritable {
		parsedColor, _, _ = parseHTMLColor(node.Style["color"])
	}
	parsedColor = applyOpacity(parsedColor, node.Style["opacity"])
	bx := render.NewColorBox(int(bkgDim.W()), int(bkgDim.H()), parsedColor)
	bds := offsetBoundsByDrawzone(bx, drawzone)
	bx.SetPos(drawzone.Min.X(), drawzone.Min.Y())
	stack.draw(bx)
	return floatgeom.Point2{float64(bds.Dx()), float64(bds.Dy())}

}

func drawBorder(node *ParsedNode, stack *trackingDrawStack, drawzone floatgeom.Rect2, noTallerThan, noWiderThan float64) floatgeom.Point2 {

	bkgDim := parseNodeDims(node, drawzone)
	if bkgDim.H() > noTallerThan {
		bkgDim.Max[1] = bkgDim.Min[1] + noTallerThan
	}
	if bkgDim.W() > noWiderThan {
		bkgDim.Max[0] = bkgDim.Min[0] + noWiderThan
	}

	_, offsetTop, offsetBot := applyPaddingForFinalOutput(node, drawzone)
	bkgDim.Max = bkgDim.Max.Add(offsetTop, offsetBot)
	minOffset := floatgeom.Point2{}
	offset := floatgeom.Point2{}
	box := render.NewColorBox(int(bkgDim.W()+maxSize*2), int(bkgDim.H()+maxSize*2), color.RGBA{0, 0, 0, 0})

	backingW := int(bkgDim.W())
	backingH := int(bkgDim.H())

	// remove space for border and the populate the border element
	// TODO: actually do this per direction
	width, brdColor, style, err := parseBorderAttributes("top", node.Style)
	if err != nil {
		width = 0
		// TODO: err handling?
		fmt.Println("encountered a bad top border", err)
	}
	if width > 0 {
		switch style {
		case "hidden", "none":
		case "solid":
			for x := 0; x < int(bkgDim.W()); x++ {
				for y := 0; y < width; y++ {
					box.Set(x, y, brdColor)
				}
			}
			minOffset[1] = float64(width)
		}

	}
	width, brdColor, style, err = parseBorderAttributes("bottom", node.Style)
	if err != nil {
		width = 0
		fmt.Println("encountered a bad bottom border", err)
	}
	if width > 0 {
		switch style {
		case "hidden", "none":
		case "solid":
			for x := 0; x < backingW; x++ {
				for y := backingH; y < int(bkgDim.Max.Y())+width; y++ {
					box.Set(x, y, brdColor)
				}
			}
			offset[1] = float64(width)
		}

	}
	width, brdColor, style, err = parseBorderAttributes("left", node.Style)
	if err != nil {
		width = 0
		fmt.Println("encountered a bad left border", err)
	}
	if width > 0 {
		switch style {
		case "hidden", "none":
		case "solid":
			for x := 0; x < width; x++ {
				for y := 0; y < backingH; y++ {
					box.Set(x, y, brdColor)
				}
			}
			minOffset[0] = float64(width)
		}
	}
	width, brdColor, style, err = parseBorderAttributes("right", node.Style)
	if err != nil {
		width = 0
		fmt.Println("encountered a bad right border", err)
	}
	if width > 0 {
		switch style {
		case "hidden", "none":
		case "solid":
			for x := backingW - width; x <= backingW; x++ {
				for y := 0; y < backingH; y++ {
					box.Set(x, y, brdColor)
				}
			}
			offset[0] = float64(width)
		}

	}
	box.SetPos(drawzone.Min.X(), drawzone.Min.Y())
	stack.drawBorder(box)

	// offset drawzone by the portions that matter for now
	drawzone.Min.Add(minOffset)
	return offset
}

func renderNode(node *ParsedNode, stack *trackingDrawStack, drawzone floatgeom.Rect2, state InteractiveState) (consumed floatgeom.Point2) {
	if node == nil {
		return consumed
	}
	var siblingDrawn bool
	start := drawzone.Min
	margin := parseMargin(node.Style)
	drawzone.Min[1] += margin.Min[1]
	if drawzone.Min.X() < margin.Min.X() {
		drawzone.Min[0] = margin.Min[0]
	}
	if node.Raw.Type == html.TextNode {
		textSize := parseLengthWithDefault(node.Style["font-size"], 16)
		textVBuffer := textSize / 5 // todo: where is this from?

		text := node.Raw.Data

		// This is not correct?
		if !unicode.IsSpace(rune(text[len(text)-1])) {
			text += " "
		}
		// TODO: move more blocks from tag switch here
		switch node.Raw.Parent.Data {
		case "th", "td":
			// Why don't the other text blocks need this width handling?
			rText, _, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			w := float64(bds.Dx())
			if w < 10 {
				w = 10
			}
			if wd, ok := node.Style["width"]; ok {
				parsed, ok := parseLength(wd)
				if ok {
					w = parsed
				}
			}
			// TODO: other text blocks should also draw their own backgrounds
			bkgDiff := drawBackground(node, stack, drawzone, textSize+textVBuffer, w)
			setIntPos(rText, bds)
			stack.draw(rText)
			consumed = bkgDiff
		case "div", "li":
			rText, _, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			setIntPos(rText, bds)
			stack.draw(rText)

			// Not sure if this is needed but definitely isn't if there is no text. see hcj02
			consumed[1] = textVBuffer + float64(bds.Dy())
		case "span", "address", "h1", "h2", "h3", "h4", "h5", "h6", "a":
			rText, _, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			setIntPos(rText, bds)
			stack.draw(rText)
			consumed[0] = float64(bds.Dx())
		}
	} else {
		// TODO: inline vs block / content categories
		switch node.Tag {
		case "figure", "tbody", "table", "tr":
			consumed = renderNode(node.FirstChild, stack, drawzone, state)
		case "div", "span", "address", "h1", "h2", "h3", "h4", "h5", "h6", "a":
			textSize := parseLengthWithDefault(node.Style["font-size"], 16)
			textVBuffer := textSize / 5 // todo: where is this from?
			// TODO: spacing around p and div is incorrect
			// TODO: div and p are really similar and yet subtly different, why?
			drawBackground(node, stack, drawzone, textSize+textVBuffer, math.MaxFloat64)
			consumed = renderNode(node.FirstChild, stack, drawzone, state)
		case "p":
			nextChild := node.FirstChild
			texts := []string{""}
			textIndex := 0
			for nextChild != nil {
				switch nextChild.Raw.Type {
				case html.TextNode:
					text := nextChild.Raw.Data
					texts[textIndex] += text
				default:
					if nextChild.Tag == "br" {
						textIndex++
						texts = append(texts, "")
					}
				}
				nextChild = nextChild.NextSibling
			}
			childDraw, _, offsetsBot := applyPaddingForFinalOutput(node, drawzone)

			var textVBuffer float64
			borderYOff := 0.0
			for i, text := range texts {
				rText, textSize, bds := formatTextAsSprite(node, childDraw, 16.0, text)
				textVBuffer = textSize / 5 // todo: where is this from?
				// todo: is this the background of p or the background of the text content child?
				if i == 0 {
					borderYOff = drawboxModel(node, stack, drawzone, (textSize+textVBuffer)*float64(len(texts)), math.MaxFloat64)

				}
				setIntPos(rText, bds)
				stack.draw(rText)
				childDraw.Min = childDraw.Min.Add(floatgeom.Point2{0, float64(bds.Dy())})
			}
			// We only care about y increment from padding and the like
			drawzone.Min[1] = childDraw.Min.Y()
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, textVBuffer})
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, borderYOff})
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, offsetsBot.Y()})

			consumed = renderNode(node.FirstChild, stack, drawzone, state)
		case "img":
			srcAttr := getAttribute(node.Raw, "src")
			if srcAttr != "" {
				r, err := loadSrc(srcAttr)
				if err != nil {
					fmt.Println(err)
					r.Close()
					break
				}
				img, _, err := image.Decode(r)
				if err != nil {
					r.Close()
					break
				}
				r.Close()
				bds := offsetBoundsByDrawzone(img, drawzone)
				var rgba *image.RGBA
				switch v := img.(type) {
				case *image.RGBA:
					rgba = v
				default:
					rgba = image.NewRGBA(img.Bounds())
					for x := 0; x < rgba.Rect.Dx(); x++ {
						for y := 0; y < rgba.Rect.Dy(); y++ {
							rgba.Set(x, y, img.At(x, y))
						}
					}
				}
				imgSprite := render.NewSprite(drawzone.Min.X(), drawzone.Min.Y(), rgba)
				stack.draw(imgSprite)
				drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, float64(bds.Dy())})
			}
			consumed = renderNode(node.FirstChild, stack, drawzone, state)
		case "ul":
			// move right, defer to li to look back upward at ul to determine its prefix
			// todo: this number appears to be too big compared to firefox; I think padding doesn't take into account the bullet width
			padding, _ := parseLength(node.Style["padding-left"])
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, 16})
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{padding, 0})
			consumed = renderNode(node.FirstChild, stack, drawzone, state)
			drawzone.Min = drawzone.Min.Sub(floatgeom.Point2{padding, 0})
		case "li":
			switch node.Style["list-style-type"] {
			case "disc", "":
				textSize := 16.0
				if size, ok := parseLength(node.Style["font-size"]); ok {
					textSize = size
				}

				textVBuffer := textSize / 5 // todo: where is this from?

				// draw bullet
				bulletRadius := textSize / 2
				bulletOffset := textSize / 3
				cir := render.NewCircle(getTextColor(node), bulletRadius/2, 1)
				cir.SetPos(drawzone.Min.X(), drawzone.Min.Y()+bulletOffset)
				stack.draw(cir)

				paddingL, _ := parseLength(node.Style["padding-left"])
				paddingT, _ := parseLength(node.Style["padding-top"])

				childDrawZone := drawzone
				childDrawZone.Min = childDrawZone.Min.Add(floatgeom.Point2{paddingL, paddingT})
				childDrawZone.Min = childDrawZone.Min.Add(floatgeom.Point2{bulletRadius * 2, 0})

				// q: is this the background of ul or the background of the text content child? a: child
				// TODO: this background does not extend down to the bottom of letters like 'g' and 'y'

				drawBackground(node, stack, childDrawZone, textSize+textVBuffer, math.MaxFloat64)
				if node.FirstChild != nil {
					consumed = renderNode(node.FirstChild, stack, childDrawZone, state)
				}
			default:
				panic("unhandled")
			}
			// TODO: thead
			// TODO: tfoot
			// TODO: colgroup
			// TODO: col
			// TODO: caption
			// TODO: tbody
			// TODO: th
		case "th", "td":
			consumed = renderNode(node.FirstChild, stack, drawzone, state)

			// handle drawing of siblings according to the table draw style
			// - siblings should not get our yDelta
			// - but parents should get larger of this yDelta and sibling yDelta
			drawzone.Min[0] += consumed[0]
			siblingsConsumed := renderNode(node.NextSibling, stack, drawzone, state)
			if consumed[1] < siblingsConsumed[1] {
				consumed[1] = siblingsConsumed[1]
			}
			drawzone.Min[1] += consumed[1]
			siblingDrawn = true
		}
	}
	if !siblingDrawn {
		drawzone.Min = drawzone.Min.Add(consumed)
		siblingsConsumed := renderNode(node.NextSibling, stack, drawzone, state)
		drawzone.Min = drawzone.Min.Add(siblingsConsumed)
	}

	// TODO: x distance should only be increased in parent if parent is an inline component
	// i.e. this 'p' / 'tr' is an incomplete list of parent tags
	consumed = drawzone.Min.Sub(start)
	switch node.Raw.Parent.Data {
	case "p", "tr":
		consumed[0] = 0
	}
	return consumed
}

func loadSrc(src string) (io.ReadCloser, error) {
	// src can be an internet thing or a base64 string, but we're going to ignore that for right now
	// just do local paths
	return os.Open(src)
}

func mergeMaps(m1, m2 map[string]string) map[string]string {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

func getTextColor(node *ParsedNode) color.Color {
	var parsedColor color.Color = color.RGBA{0, 0, 0, 255}
	if col := node.Style["color"]; col != "" {
		parsedColor, _, _ = parseHTMLColor(col)
	}
	parsedColor = applyOpacity(parsedColor, node.Style["opacity"])
	return parsedColor
}

var nonBreakingWhitespace = rune(0xA0)

// formatTextAsSprite from a node where styling decisions have already been taken care of given wrapping elem
// introduced to dedupe a portion of div versus p
// this is purely formatting as oak and should not contain any novel formatting decisions
// consumer will often call the draw on the returned sprite and may add the bds sizing to the overall drawzone
func formatTextAsSprite(node *ParsedNode, drawzone floatgeom.Rect2, textSizeDefault float64, inText string) (*render.Sprite, float64, image.Rectangle) {
	textSize := parseLengthWithDefault(node.Style["font-size"], textSizeDefault)
	newTxt := strings.Builder{}
	skipWhitespace := false
	var lastRn rune
	for i, rn := range inText {
		// trim all whitespace following newlines or 0xA0
		if skipWhitespace && (rn == ' ' || rn == '\t' || rn == '\r' || rn == nonBreakingWhitespace) {
			continue
		}
		skipWhitespace = false
		if rn == '\n' || rn == nonBreakingWhitespace {
			// newlines become spaces iff a space is not already present before them
			if rn == '\n' && lastRn != ' ' && i != 0 {
				newTxt.WriteRune(' ')
			}
			skipWhitespace = true
			continue
		}
		newTxt.WriteRune(rn)
		lastRn = rn
	}
	text := newTxt.String()
	fnt, _ := render.DefaultFont().RegenerateWith(func(fg render.FontGenerator) render.FontGenerator {
		fg.Color = image.NewUniform(getTextColor(node))
		fg.Size = textSize
		if node.Style["font-style"] == "italic" {
			fg.RawFile = luxisriTTF
		} else if node.Style["font-style"] == "bold" {
			fg.RawFile = luxisbTTF
		}
		return fg
	})
	textRenderable := fnt.NewText(text, 0, 0).ToSprite()
	// Drawzone location should be respected so add to bounds
	return textRenderable, textSize, offsetBoundsByDrawzone(textRenderable, drawzone)
}

// ofsetBoundsByDrawzone retireves the bounds of an object, offsets it by drawzone and returns that bounds
func offsetBoundsByDrawzone(boundable hasBounds, drawzone floatgeom.Rect2) image.Rectangle {
	bds := boundable.Bounds()
	bds.Min.X = int(drawzone.Min.X())
	bds.Min.Y = int(drawzone.Min.Y())
	bds.Max.X += int(drawzone.Min.X())
	bds.Max.Y += int(drawzone.Min.Y())
	return bds
}

type hasBounds interface {
	// Bounds returns the domain for which At can return non-zero color.
	// The bounds do not necessarily contain the point (0, 0).
	Bounds() image.Rectangle
}

// TODO: Make the image rectangle be a float64 bounds type thing so we dont do this
func setIntPos(s *render.Sprite, bds image.Rectangle) {
	s.SetPos(float64(bds.Min.X), float64(bds.Min.Y))
}

func applyPaddingForFinalOutput(node *ParsedNode, drawzone floatgeom.Rect2) (floatgeom.Rect2, floatgeom.Point2, floatgeom.Point2) {
	paddingL, _ := parseLength(node.Style["padding-left"])
	paddingT, _ := parseLength(node.Style["padding-top"])

	childDrawTop := floatgeom.Point2{paddingL, paddingT}

	paddingR, _ := parseLength(node.Style["padding-right"])
	paddingB, _ := parseLength(node.Style["padding-bottom"])
	childDrawBot := floatgeom.Point2{paddingR, paddingB}

	childDrawZone := floatgeom.Rect2{}
	childDrawZone.Min = drawzone.Min.Add(childDrawTop)
	childDrawZone.Max = drawzone.Max.Add(childDrawBot)
	return childDrawZone, childDrawTop, childDrawBot
}
