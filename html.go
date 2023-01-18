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

func RenderHTML(htmlReader io.Reader, dims intgeom.Point2) (*render.Sprite, error) {
	rootNode, err := html.Parse(htmlReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}
	sp := render.NewEmptySprite(0, 0, dims.X(), dims.Y())
	// 1: empty html is blank white
	var bkgColor color.Color = color.RGBA{255, 255, 255, 255}
	// 2: body style background-color
	var css CSS
	if headNode := findHTMLNode(rootNode, "head"); headNode != nil {
		if styleNode := findHTMLNode(headNode, "style"); styleNode != nil {
			if styleNode.FirstChild != nil {
				css = ParseCSS(string(styleNode.FirstChild.Data))
			}
		}
	}
	css = DefaultCSS().Merge(css)
	if bodyNode := findHTMLNode(rootNode, "body"); bodyNode != nil {
		bn := ParseNode(bodyNode, WithCSS(css))
		// TODO: what to do if both of these are set?
		if col, ok := bn.Style["background-color"]; ok {
			bkgColor, _, _ = parseHTMLColor(col)
		} else if col, ok := bn.Style["background"]; ok {
			bkgColor, _, _ = parseHTMLColor(col)
		}
		for x := 0; x < dims.X(); x++ {
			for y := 0; y < dims.Y(); y++ {
				sp.Set(x, y, bkgColor)
			}
		}
		fullBodyMargin := parseMargin(bn.Style)
		bodyMargin := fullBodyMargin.Min
		zone := floatgeom.Rect2{
			Min: bodyMargin,
			Max: floatgeom.Point2{float64(dims.X()), float64(dims.Y())}.Sub(bodyMargin),
		}
		drawNode(bn.FirstChild, sp, zone)
	}
	return sp, nil
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
	// todo: how often do we need to strip spaces from these values
	vals := strings.Split(data, ";")
	out := make(map[string]string, len(vals))
	for _, v := range vals {
		splitVal := strings.Split(v, ":")
		if len(splitVal) != 2 {
			// ignore bad formatted thing
			continue
		}
		out[strings.TrimSpace(splitVal[0])] = strings.TrimSpace(splitVal[1])
	}
	return out
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
	return floatgeom.NewRect2WH(0, 0, w, h)
}

func drawBackground(node *ParsedNode, sp *render.Sprite, drawzone floatgeom.Rect2, noTallerThan, noWiderThan float64) floatgeom.Point2 {
	bkg, ok := node.Style["background"]
	if !ok {
		bkg, ok = node.Style["background-color"]
	}
	if ok {
		bkgDim := parseNodeDims(node, drawzone)
		if bkgDim.H() > noTallerThan {
			bkgDim.Max[1] = bkgDim.Min[1] + noTallerThan
		}
		if bkgDim.W() > noWiderThan {
			bkgDim.Max[0] = bkgDim.Min[0] + noWiderThan
		}
		parsedColor, _, inheritable := parseHTMLColor(bkg)
		if inheritable {
			parsedColor, _, _ = parseHTMLColor(node.Style["color"])
		}
		parsedColor = applyOpacity(parsedColor, node.Style["opacity"])
		bx := render.NewColorBox(int(bkgDim.W()), int(bkgDim.H()), parsedColor)
		bds := offsetBoundsByDrawzone(bx, drawzone)
		draw.Draw(sp.GetRGBA(), bds, bx.GetRGBA(), image.Point{}, draw.Over)
		return floatgeom.Point2{float64(bds.Dx()), float64(bds.Dy())}
	}
	return floatgeom.Point2{}
}

func drawNode(node *ParsedNode, sp *render.Sprite, drawzone floatgeom.Rect2) (heightConsumed float64) {
	if node == nil {
		return 0
	}
	startHeight := drawzone.Min.Y()
	margin := parseMargin(node.Style)
	drawzone.Min[1] += margin.Min[1]
	if drawzone.Min.X() < margin.Min.X() {
		drawzone.Min[0] = margin.Min[0]
	}
	childDrawzoneModifier := floatgeom.Point2{}
	// TODO: inline vs block / content categories
	switch node.Tag {
	case "div":
		// TODO: div and p are really similar and yet subtly different, why?
		drawBackground(node, sp, drawzone, math.MaxFloat64, math.MaxFloat64)
		// TODO: this is not right; children are drawn below already;
		// how do we know whether a node is text content?
		if node.FirstChild != nil && node.FirstChild.FirstChild == nil {
			text := node.FirstChild.Raw.Data
			rText, textSize, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			textVBuffer := textSize / 5 // todo: where is this from?
			draw.Draw(sp.GetRGBA(), bds, rText.GetRGBA(), image.Point{}, draw.Over)
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, textVBuffer + float64(bds.Dy())})
		}
	case "span":
		fallthrough
	case "address":
		if node.FirstChild != nil {
			text := node.FirstChild.Raw.Data
			// This is not correct?
			if !unicode.IsSpace(rune(text[len(text)-1])) {
				text += " "
			}
			rText, textSize, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			textVBuffer := textSize / 5 // todo: where is this from?
			// todo: is this the background of p or the background of the text content child?
			drawBackground(node, sp, drawzone, textSize+textVBuffer, math.MaxFloat64)
			draw.Draw(sp.GetRGBA(), bds, rText.GetRGBA(), image.Point{}, draw.Over)
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{float64(bds.Dx()), 0})
		}
	case "p":
		if node.FirstChild != nil {
			text := node.FirstChild.Raw.Data
			rText, textSize, bds := formatTextAsSprite(node, drawzone, 16.0, text)
			textVBuffer := textSize / 5 // todo: where is this from?
			// todo: is this the background of p or the background of the text content child?
			drawBackground(node, sp, drawzone, textSize+textVBuffer, math.MaxFloat64)
			draw.Draw(sp.GetRGBA(), bds, rText.GetRGBA(), image.Point{}, draw.Over)
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, textVBuffer + float64(bds.Dy())})
		}
	case "img":
		for _, atr := range node.Raw.Attr {
			if atr.Key == "src" {
				r, err := loadSrc(atr.Val)
				if err != nil {
					fmt.Println(err)
					r.Close()
					continue
				}
				img, _, err := image.Decode(r)
				if err != nil {
					r.Close()
					continue
				}
				r.Close()
				bds := offsetBoundsByDrawzone(img, drawzone)
				draw.Draw(sp.GetRGBA(), bds, img, image.Point{}, draw.Over)
				drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, float64(bds.Dy())})
			}
		}
	case "ul":
		// move right, defer to li to look back upward at ul to determine its prefix
		// todo: this number appears to be too big compared to firefox; I think padding doesn't take into account the bullet width
		padding, _ := parseLength(node.Style["padding-left"])
		childDrawzoneModifier[0] = padding
		drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, 16})
	case "li":
		switch node.Raw.Parent.Data {
		case "ul":
			if node.FirstChild != nil {
				text := node.FirstChild.Raw.Data
				textSize := 16.0
				if size, ok := parseLength(node.Style["font-size"]); ok {
					textSize = size
				}

				textVBuffer := textSize / 5 // todo: where is this from?

				// draw bullet
				bulletRadius := textSize / 2
				bulletOffset := textSize / 3
				render.DrawCircle(sp.GetRGBA(), getTextColor(node.FirstChild), bulletRadius/2, 1, drawzone.Min.X(), drawzone.Min.Y()+bulletOffset)

				// TODO: this number
				bulletGap := bulletRadius * 2
				drawzone.Min = drawzone.Min.Add(floatgeom.Point2{bulletGap, 0})

				// todo: is this the background of ul or the background of the text content child?
				// TODO: this background does not extend down to the bottom of letters like 'g' and 'y'
				drawBackground(node, sp, drawzone, textSize+textVBuffer, math.MaxFloat64)

				// draw text
				rText, _, bds := formatTextAsSprite(node, drawzone, 16.0, text)
				draw.Draw(sp.GetRGBA(), bds, rText.GetRGBA(), image.Point{}, draw.Over)
				drawzone.Min = drawzone.Min.Add(floatgeom.Point2{-bulletGap, textVBuffer + float64(bds.Dy())})
			}
		}
	case "table":
		// TODO: thead
		// TODO: tfoot
		// TODO: colgroup
		// TODO: col
		// TODO: caption
		// TODO: tbody
		// TODO: th
		// Assert we are getting a series of 'tr's, each with the same count of 'td's or 'th's.
		unknown := node.FirstChild
		if unknown == nil {
			fmt.Println("found table without unknown buffer cell")
			return
		}
		tBody := unknown.NextSibling
		if tBody.Tag != "tbody" {
			fmt.Println("found table without body")
			return
		}
		nextRow := tBody.FirstChild
		for nextRow != nil {
			if nextRow.Tag != "tr" {
				if nextRow.NextSibling != nil && nextRow.NextSibling.Tag == "tr" {
					nextRow = nextRow.NextSibling
				} else {
					break
				}
			}
			col := nextRow.FirstChild
			tallestColumnHeight := 0.0
			rowWidth := 0.0
			for col != nil {
				if strings.TrimSpace(col.Tag) == "" && col.NextSibling != nil {
					col = col.NextSibling
					continue
				}
				if col.Tag != "th" && col.Tag != "td" {
					break
				}
				// draw, then move right
				rText, textSize, bds := formatTextAsSprite(node, drawzone, 16.0, node.FirstChild.Raw.Data)
				textVBuffer := textSize / 5 // todo: where is this from?

				// TODO: how wide are spaces? 3 spaces is 4 pixels here
				w := float64(bds.Dx())
				if w < 10 {
					w = 10
				}
				if wd, ok := col.Style["width"]; ok {
					parsed, ok := parseLength(wd)
					if ok {
						w = parsed
					}
				}
				bkgDiff := drawBackground(col, sp, drawzone, textSize+textVBuffer, w)

				bds.Min.X = int(drawzone.Min.X())
				bds.Min.Y = int(drawzone.Min.Y())
				bds.Max.X += int(drawzone.Min.X())
				bds.Max.Y += int(drawzone.Min.Y())
				draw.Draw(sp.GetRGBA(), bds, rText.GetRGBA(), image.Point{}, draw.Over)
				rowWidth += bkgDiff.X()
				drawzone.Min = drawzone.Min.Add(floatgeom.Point2{bkgDiff.X(), 0})
				col = col.NextSibling
				if bkgDiff.Y() > tallestColumnHeight {
					tallestColumnHeight = bkgDiff.Y()
				}
			}
			// move down
			drawzone.Min = drawzone.Min.Add(floatgeom.Point2{-rowWidth, float64(tallestColumnHeight)})
			nextRow = nextRow.NextSibling
			if nextRow != nil && nextRow.Tag == "" { // ?????
				nextRow = nextRow.NextSibling
			}
		}
	}
	drawzone.Min = drawzone.Min.Add(childDrawzoneModifier)
	childrenHeight := drawNode(node.FirstChild, sp, drawzone)
	drawzone.Min = drawzone.Min.Sub(childDrawzoneModifier)
	drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, float64(childrenHeight)})
	siblingsHeight := drawNode(node.NextSibling, sp, drawzone)
	drawzone.Min = drawzone.Min.Add(floatgeom.Point2{0, float64(siblingsHeight)})
	return drawzone.Min.Y() - startHeight
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

// formatTextAsSprite from a node where styling decisions have already been taken care of given wrapping elem
// introduced to dedupe a portion of div versus p
// this is purely formatting as oak and should not contain any novel formatting decisions
// consumer will often call the draw on the returned sprite and may add the bds sizing to the overall drawzone
func formatTextAsSprite(node *ParsedNode, drawzone floatgeom.Rect2, textSizeDefault float64, inText string) (*render.Sprite, float64, image.Rectangle) {
	textSize := textSizeDefault
	if size, ok := parseLength(node.Style["font-size"]); ok {
		textSize = float64(size)
	}
	newTxt := strings.Builder{}
	foundNewline := false
	var lastRn rune
	for _, rn := range inText {
		// trim all whitespace following newlines
		if foundNewline && (rn == ' ' || rn == '\t' || rn == '\r') {
			continue
		}
		foundNewline = false
		if rn == '\n' {
			// newlines become spaces iff a space is not already present before them
			if lastRn != ' ' {
				newTxt.WriteRune(' ')
			}
			foundNewline = true
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
		} else if node.Style["font-stlye"] == "bold" {
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
