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
				Classes: []string{"class"},
			},
		}, {
			input: ".class1.class2",
			expected: Selector{
				Classes: []string{"class1", "class2"},
			},
		}, {
			input: "#id",
			expected: Selector{
				IDs: []string{"id"},
			},
		}, {
			input: "#id1#id2",
			expected: Selector{
				IDs: []string{"id1", "id2"},
			},
		},
		{
			input: "tag.class#id",
			expected: Selector{
				IDs:     []string{"id"},
				Tag:     "tag",
				Classes: []string{"class"},
			},
		},
		{
			input: "tag.class1.class2#id",
			expected: Selector{
				IDs:     []string{"id"},
				Tag:     "tag",
				Classes: []string{"class1", "class2"},
			},
		}, {
			input: "tag#id",
			expected: Selector{
				IDs: []string{"id"},
				Tag: "tag",
			},
		}, {
			input: "tag#id.class",
			expected: Selector{
				IDs:     []string{"id"},
				Tag:     "tag",
				Classes: []string{"class"},
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
		// {
		// 	input: ".class1.class2",
		// },
		// {
		// 	input: ".class1",
		// },
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
	if len(s1.IDs) != len(s2.IDs) {
		return fmt.Errorf("mismatched ids length : %v vs %v", len(s1.IDs), len(s2.IDs))
	}
	for i, c1 := range s1.IDs {
		if s2.IDs[i] != c1 {
			return fmt.Errorf("mismatched id at index %d : %v vs %v", i, len(s1.IDs), len(s2.IDs))
		}
	}
	if s1.Global != s2.Global {
		return fmt.Errorf("mismatched Global: %v vs %v", s1.Global, s2.Global)
	}
	if len(s1.Classes) != len(s2.Classes) {
		return fmt.Errorf("mismatched class length : %v vs %v", len(s1.Classes), len(s2.Classes))
	}
	for i, c1 := range s1.Classes {
		if s2.Classes[i] != c1 {
			return fmt.Errorf("mismatched class at index %d : %v vs %v", i, len(s1.Classes), len(s2.Classes))
		}
	}
	return nil
}
