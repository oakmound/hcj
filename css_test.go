package hcj

import (
	"fmt"
	"strconv"
	"testing"
)

func TestParseSelector(t *testing.T) {
	t.Parallel()
	type testCase struct {
		input    string
		expected Selector
	}
	tcs := []testCase{
		{
			input: "*",
			expected: Selector{
				Global: true,
			},
		}, {
			input: "tag",
			expected: Selector{
				Tag: "tag",
			},
		}, {
			input: ".class",
			expected: Selector{
				Class: "class",
			},
		}, {
			input: "#id",
			expected: Selector{
				ID: "id",
			},
		}, {
			input: "tag.class#id",
			expected: Selector{
				ID:    "id",
				Tag:   "tag",
				Class: "class",
			},
		}, {
			input: "tag#id",
			expected: Selector{
				ID:  "id",
				Tag: "tag",
			},
		}, {
			input: "tag#id.class",
			expected: Selector{
				ID:    "id",
				Tag:   "tag",
				Class: "class",
			},
		}, {
			input: "a[target]",
			expected: Selector{
				Tag:       "a",
				Attribute: "target",
			},
		},
	}
	for i, tc := range tcs {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			sel, err := ParseSelector(tc.input)
			if err != nil {
				t.Fatal("unexpected error", err)
			}
			if err := SelectorsEqual(sel, tc.expected); err != nil {
				fmt.Println(tc.input, sel, tc.expected)
				t.Fatal(err)
			}
		})
	}
}

func TestParseSelectorInvalid(t *testing.T) {
	t.Parallel()
	type testCase struct {
		input string
	}
	tcs := []testCase{
		{
			input: "\u0500",
		},
		{
			input: "\n",
		},
		{
			input: "\r",
		},
		{
			input: " ",
		},
		{
			input: "\t",
		},
		{
			input: "#id1#id2",
		},
		{
			input: "##",
		},
		{
			input: "..",
		},
		{
			input: ".#",
		},
		{
			input: "#.",
		},
		{
			input: ".class1.class2",
		},
		{
			input: "tag.class1.class2#id",
		},
	}
	for i, tc := range tcs {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			_, err := ParseSelector(tc.input)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func SelectorsEqual(s1, s2 Selector) error {
	if s1.Tag != s2.Tag {
		return fmt.Errorf("mismatched tag: %v vs %v", s1.Tag, s2.Tag)
	}
	if s1.ID != s2.ID {
		return fmt.Errorf("mismatched ID: %v vs %v", s1.ID, s2.ID)
	}
	if s1.Global != s2.Global {
		return fmt.Errorf("mismatched Global: %v vs %v", s1.Global, s2.Global)
	}
	if s2.Class != s1.Class {
		return fmt.Errorf("mismatched class: %v vs %v", s1.Class, s2.Class)
	}
	return nil
}
