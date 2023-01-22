package hcj_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/oakmound/hcj"
	"github.com/oakmound/oak/v4/alg/intgeom"
)

var testDirs = []string{"htmlin"}

func Test_RenderHTML_Golden(t *testing.T) {
	t.Parallel()
	inputs := make(map[string]io.Reader)
	for _, inDirName := range testDirs {
		inDir, err := os.ReadDir(filepath.Join("testdata", inDirName))
		if err != nil {
			t.Fatal(err)
		}

		for _, fi := range inDir {
			f, err := os.Open(filepath.Join("testdata", inDirName, fi.Name()))
			if err != nil {
				t.Fatal(err)
			}
			htmlData, err := io.ReadAll(f)
			if err != nil {
				f.Close()
				t.Fatal(err)
			}
			inputs[fi.Name()] = bytes.NewReader(htmlData)
			f.Close()
		}
	}

	outDir, err := os.ReadDir(filepath.Join("testdata", "htmlout"))
	if err != nil {
		t.Fatal(err)
	}
	outputs := make(map[string]image.RGBA, len(outDir))
	for _, fi := range outDir {
		f, err := os.Open(filepath.Join("testdata", "htmlout", fi.Name()))
		if err != nil {
			t.Fatal(err)
		}
		pngData, err := png.Decode(f)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		bds := pngData.Bounds()
		rgba := image.NewRGBA(bds)
		for x := 0; x < bds.Max.X; x++ {
			for y := 0; y < bds.Max.Y; y++ {
				rgba.Set(x, y, pngData.At(x, y))
			}
		}
		outputs[fi.Name()] = *rgba
	}
	for i, in := range inputs {
		i := i
		in := in
		t.Run(i, func(t *testing.T) {
			t.Parallel()
			node, sp, err := hcj.ParseAndRenderHTML(in, intgeom.Point2{500, 300})
			if err != nil {
				t.Error(err)
				return
			}
			rgba1 := *sp.GetRGBA()
			pngFileName := strings.ReplaceAll(i, ".html", ".png")
			pngFileName = strings.ReplaceAll(pngFileName, ".htm", ".png")
			rgba2, ok := outputs[pngFileName]
			if !ok {
				// create the baseline
				fmt.Println("creating baseline for", pngFileName)
				outfile, err := os.Create(filepath.Join("testdata", "htmlout", pngFileName))
				if err != nil {
					t.Fatal(err)
				}
				png.Encode(outfile, &rgba1)
				outfile.Close()
				return
			}
			if !cmp.Equal(rgba1, rgba2) {
				t.Error("mismatch for key " + i)
				os.MkdirAll(filepath.Join("testdata", "htmlfail"), 0700)
				outfile, err := os.Create(filepath.Join("testdata", "htmlfail", pngFileName))
				if err != nil {
					t.Fatal(err)
				}
				png.Encode(outfile, &rgba1)

				// print out internal rep
				fmt.Println(node)
				outfile.Close()
			}
		})
	}
}
