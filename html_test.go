package hcj_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/oakmound/hcj"
	"github.com/oakmound/oak/v4/render"
)

var testDirs = []string{"htmlin"}

func Test_RenderHTML_Golden(t *testing.T) {
	t.Parallel()

	type input struct {
		f   io.Reader
		cfg testConfig
	}

	inputs := make(map[string]input)
	for _, inDirName := range testDirs {
		inDir, err := os.ReadDir(filepath.Join("testdata", inDirName))
		if err != nil {
			t.Fatal(err)
		}

		for _, fi := range inDir {
			ext := filepath.Ext(fi.Name())
			if ext == ".json" {
				continue
			}

			path := filepath.Join("testdata", inDirName, fi.Name())
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			htmlData, err := io.ReadAll(f)
			if err != nil {
				f.Close()
				t.Fatal(err)
			}
			in := input{
				f: bytes.NewReader(htmlData),
			}

			testName := strings.TrimSuffix(path, ext)
			cfgFileName := testName + ".json"
			if cfgFile, err := os.ReadFile(cfgFileName); err == nil {
				var cfg testConfig
				err = json.Unmarshal(cfgFile, &cfg)
				if err != nil {
					t.Fatal(err)
				}
				in.cfg = cfg
			}

			inputs[fi.Name()] = in
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

			ext := filepath.Ext(i)
			testName := strings.TrimSuffix(i, ext)

			if len(in.cfg.Steps) == 0 {
				in.cfg.Steps = append(in.cfg.Steps, testStep{})
			}
			for i, step := range in.cfg.Steps {
				in.f.(io.Seeker).Seek(0, io.SeekStart)
				stepTestName := testName
				if len(in.cfg.Steps) != 1 {
					stepTestName += "-" + strconv.Itoa(i)
				}

				node, err := hcj.ParseHTML(in.f, hcj.WithInteractiveState(translateConfigState(step)))
				if err != nil {
					t.Error(err)
					return
				}

				sp := render.NewEmptySprite(0, 0, 500, 300)
				node.Draw(sp.GetRGBA(), 0, 0)

				rgba1 := *sp.GetRGBA()
				pngFileName := stepTestName + ".png"
				rgba2, ok := outputs[pngFileName]
				if !ok || os.Getenv("OVERRIDE_TESTDATA") != "" {
					// create the baseline
					fmt.Println("creating baseline for", pngFileName)
					outfile, err := os.Create(filepath.Join("testdata", "htmlout", pngFileName))
					if err != nil {
						t.Fatal(err)
					}
					png.Encode(outfile, &rgba1)
					outfile.Close()
					continue
				}
				if !cmp.Equal(rgba1, rgba2) {
					t.Error("mismatch for key " + stepTestName)
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
			}
		})
	}
}

type testConfig struct {
	Steps []testStep `json:"steps"`
}

type testStep struct {
	VisitedAddresses []string `json:"visitedAddresses"`
}

func translateConfigState(s testStep) hcj.InteractiveState {
	addresses := make(map[string]struct{})
	for _, address := range s.VisitedAddresses {
		addresses[address] = struct{}{}
	}
	return hcj.InteractiveState{
		VisitedAddresses: addresses,
	}
}
