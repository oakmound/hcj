package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

const name = "selectortypes.htm"

func main() {
	readFile, err := os.Open(filepath.Join("..", name))

	if err != nil {
		fmt.Println(err)
	}

	rootNode, err := html.Parse(readFile)
	if err != nil {
		fmt.Printf("failed to parse html: %s", err.Error())
		return
	}

	out := "package sub\n\n"

	out += recurse(rootNode) + "\n}\n)"

	outByts := []byte(out)
	lastDot := strings.LastIndex(name, ".")
	outName := name[:lastDot] + ".go"

	err = os.WriteFile(filepath.Join("sub", outName), outByts, 0644)
	fmt.Println(err)

}

func recurse(node *html.Node) string {
	out := ""

	if strings.Contains(node.Data, "thead") {
		out += "const(\n me = []interface{"
	}

	// dont care anything after tbody
	if strings.Contains(node.Data, "tbody") {
		out += createTREntriesFromSiblings(node.FirstChild, []string{})
	}

	if node.FirstChild != nil {
		out += recurse(node.FirstChild)
	}
	if node.NextSibling != nil {
		out += recurse(node.NextSibling)
	}
	return out
}

// convert tr entries in the tbody to an entry for your slice
/*
    {
		"class": "entry_escaped",
		...
		...
		...
	},
*/
func createTREntriesFromSiblings(node *html.Node, fieldNames []string) string {
	row := ""

	entry := node.FirstChild
	for entry != nil {
		// weird input?
		if entry.FirstChild == nil {
			entry = entry.NextSibling
			continue
		}
		className := "!TODO!"
		for _, attr := range entry.Attr {
			if attr.Key != "class" {
				continue
			}
			className = attr.Val
			className = strings.ToTitle(className[0:1]) + className[1:]
			break
		}
		lineStr := fmt.Sprintf("%s: %q,\n", className, sortaRawDump(entry.FirstChild))
		row += lineStr

		entry = entry.NextSibling
	}

	row = fmt.Sprintf("{\n\t%s},\n", row)

	if node.NextSibling == nil {
		return row
	}
	return row + "\n" + createTREntriesFromSiblings(node.NextSibling, fieldNames)
}

func sortaRawDump(node *html.Node) string {

	entryStr := node.Data
	if node.FirstChild != nil {
		entryStr = sortaRawDump(node.FirstChild)
	}
	if node.NextSibling != nil {
		entryStr += sortaRawDump(node.NextSibling)
	}

	return strings.TrimSpace(entryStr)
}
