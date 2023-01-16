package hcj

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"golang.org/x/image/colornames"
)

func parseHTMLColor(col string) (c color.Color, ok bool, inheritFromParent bool) {
	if namedColor, ok := colornames.Map[col]; ok {
		return namedColor, true, false
	}
	if col == "transparent" {
		return color.RGBA{0, 0, 0, 0}, true, false
	}
	if col == "inherit" || col == "currentColor" || col == "currentcolor" {
		return color.RGBA{0, 0, 0, 0}, true, true
	}
	if len(col) < 3 {
		return color.RGBA{255, 255, 255, 255}, false, false
	}
	// parse as hex
	if col[0] == '#' {
		switch len(col) {
		case 4:
			col = string([]byte{'#',
				col[1],
				col[1],
				col[2],
				col[2],
				col[3],
				col[3],
			})
			fallthrough
		case 7:
			var r, g, b uint8
			_, err := fmt.Sscanf(col, "#%2x%2x%2x", &r, &g, &b)
			if err != nil {
				return color.RGBA{255, 255, 255, 255}, false, false
			}
			return color.RGBA{r, g, b, 255}, true, false
		default:
			return color.RGBA{255, 255, 255, 255}, false, false
		}
	}
	rp := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "")
	col = rp.Replace(col)

	if strings.HasPrefix(col, "rgb(") {
		percCt := strings.Count(col, "%")
		if percCt == 0 {
			var r, g, b uint8
			_, err := fmt.Sscanf(col, "rgb(%d,%d,%d)", &r, &g, &b)
			if err != nil {
				return color.RGBA{255, 255, 255, 255}, false, false
			}
			return color.RGBA{r, g, b, 255}, true, false
		} else if percCt == 3 {
			var r, g, b float32
			_, err := fmt.Sscanf(col, "rgb(%f%%,%f%%,%f%%)", &r, &g, &b)
			if err != nil {
				return color.RGBA{255, 255, 255, 255}, false, false
			}
			return color.RGBA{uint8((r / 100) * 255), uint8((g / 100) * 255), uint8((b / 100) * 255), 255}, true, false
		} else {
			return color.RGBA{255, 255, 255, 255}, false, false
		}
	}
	if strings.HasPrefix(col, "rgba(") {
		percCt := strings.Count(col, "%")
		if percCt == 0 {
			var r, g, b uint8
			var a float64
			_, err := fmt.Sscanf(col, "rgba(%d,%d,%d,%f)", &r, &g, &b, &a)
			if err != nil {
				return color.RGBA{255, 255, 255, 255}, false, false
			}
			if a < 0 {
				a = 0
			}
			if a > 1 {
				a = 1
			}
			return color.RGBA{r, g, b, uint8(a * 255)}, true, false
		} else if percCt == 3 {
			var r, g, b float32
			var a float64
			_, err := fmt.Sscanf(col, "rgba(%f%%,%f%%,%f%%,%f)", &r, &g, &b, &a)
			if err != nil {
				return color.RGBA{255, 255, 255, 255}, false, false
			}
			if a < 0 {
				a = 0
			}
			if a > 1 {
				a = 1
			}
			return color.RGBA{uint8((r / 100) * 255), uint8((g / 100) * 255), uint8((b / 100) * 255), uint8(a * 255)}, true, false
		} else {
			return color.RGBA{255, 255, 255, 255}, false, false
		}
	}
	// https://github.com/gerow/go-color/blob/master/color.go
	if strings.HasPrefix(col, "hsl(") {
		var h uint16
		var s, l float64
		col = strings.ReplaceAll(col, " ", "")
		_, err := fmt.Sscanf(col, "hsl(%d,%f%%,%f%%)", &h, &s, &l)
		if err != nil {
			return color.RGBA{255, 255, 255, 255}, false, false
		}
		rgb := Hsl2Rgb(float64(h), s, l)
		return color.RGBA{
			R: uint8(rgb[0]),
			G: uint8(rgb[1]),
			B: uint8(rgb[2]),
			A: 255,
		}, true, false
	}
	if strings.HasPrefix(col, "hsla(") {
		var h uint16
		var s, l, a float64
		col = strings.ReplaceAll(col, " ", "")
		_, err := fmt.Sscanf(col, "hsla(%d,%f%%,%f%%,%f)", &h, &s, &l, &a)
		if err != nil {
			return color.RGBA{255, 255, 255, 255}, false, false
		}
		if a < 0 {
			a = 0
		}
		if a > 1 {
			a = 1
		}
		rgb := Hsl2Rgb(float64(h), s, l)
		return color.NRGBA{
			R: uint8(rgb[0]),
			G: uint8(rgb[1]),
			B: uint8(rgb[2]),
			A: uint8(a * 255),
		}, true, false
	}
	// todo: hwb
	// todo: cmyk
	// todo: ncol
	return color.RGBA{255, 255, 255, 255}, true, false
}

// from https://github.com/hisamafahri/coco
func Hsl2Rgb(h float64, s float64, l float64) [3]float64 {
	h = h / 360
	s = s / 100
	l = l / 100

	var t2 float64
	var t3 float64
	var val float64
	var result [3]float64

	if s == 0 {
		val = l * 255
		result[0] = math.Round(val)
		result[1] = math.Round(val)
		result[2] = math.Round(val)

		return result
	}

	if l < 0.5 {
		t2 = l * (1 + s)
	} else {
		t2 = l + s - l*s
	}

	t1 := 2*l - t2

	rgb := [3]float64{0, 0, 0}

	for i := 0; i < 3; i++ {
		t3 = h + 1.0/3.0*(-(float64(i) - 1))

		if t3 < 0 {
			t3++
		}

		if t3 > 1 {
			t3--
		}

		if (6 * t3) < 1 {
			val = t1 + (t2-t1)*6*t3
		} else if (2 * t3) < 1 {
			val = t2
		} else if (3 * t3) < 2 {
			val = t1 + (t2-t1)*(2.0/3.0-t3)*6
		} else {
			val = t1
		}
		rgb[i] = val * 255
	}

	result[0] = math.Round(rgb[0])
	result[1] = math.Round(rgb[1])
	result[2] = math.Round(rgb[2])

	return result
}
