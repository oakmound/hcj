package hcj

// CSS3SelectorInfo as pulled from https://www.w3.org/TR/2018/REC-selectors-3-20181106/#selectors
type CSS3SelectorInfo struct {
	Pattern, Meaning, Described, Origin string
}

type CSS3Selector int

const (
	any_element CSS3Selector = iota
	an_element_of_type_E
	an_E_element_with_a_foo_attribute
	an_E_element_whose_foo_attribute_value_is_exactly_equal_to_bar
	an_E_element_whose_foo_attribute_value_is_a_list_of_whitespaceseparated_values_one_of_which_is_exactly_equal_to_bar
	an_E_element_whose_foo_attribute_value_begins_exactly_with_the_string_bar
	an_E_element_whose_foo_attribute_value_ends_exactly_with_the_string_bar
	an_E_element_whose_foo_attribute_value_contains_the_substring_bar
	an_E_element_whose_foo_attribute_has_a_hyphenseparated_list_of_values_beginning__with_en
	an_E_element_root_of_the_document
	an_E_element_the_nth_child_of_its_parent
	an_E_element_the_nth_child_of_its_parent_counting_from_the_last_one
	an_E_element_the_nth_sibling_of_its_type
	an_E_element_the_nth_sibling_of_its_type_counting_from_the_last_one
	an_E_element_first_child_of_its_parent
	an_E_element_last_child_of_its_parent
	an_E_element_first_sibling_of_its_type
	an_E_element_last_sibling_of_its_type
	an_E_element_only_child_of_its_parent
	an_E_element_only_sibling_of_its_type
	an_E_element_that_has_no_children_
	an_E_element_being_the_source_anchor_of_a_hyperlink_of_which_the_target_is_not_yet_visited_
	an_E_element_during_certain_user_actions
	an_E_element_being_the_target_of_the_referring_URI
	an_element_of_type_E_in_language
	a_user_interface_element_E_which_is_enabled_or_disabled
	a_user_interface_element_E_which_is_checkedor_in_an_indeterminate_state
	the_first_formatted_line_of_an_E_element
	the_first_formatted_letter_of_an_E_element
	generated_content_before_an_E_element
	generated_content_after_an_E_element
	an_E_element_whose_class_is_warning_
	an_E_element_with_ID_equal_to_myid
	an_E_element_that_does_not_match_simple_selector_s
	an_F_element_descendant_of_an_E_element
	an_F_element_child_of_an_E_element
	an_F_element_immediately_preceded_by_an_E_element
	an_F_element_preceded_by_an_E_element
)

var (
	CSS3SelectorInfoLookup = []CSS3SelectorInfo{
		{
			Pattern:   "*",
			Meaning:   "any element",
			Described: "Universal selector",
			Origin:    "2",
		},

		{
			Pattern:   "E",
			Meaning:   "an element of type E",
			Described: "Type selector",
			Origin:    "1",
		},

		{
			Pattern:   "E[foo]",
			Meaning:   "an E element with a \"foo\" attribute",
			Described: "Attribute\n       selectors",
			Origin:    "2",
		},

		{
			Pattern:   "E[foo=\"bar\"]",
			Meaning:   "an E element whose \"foo\" attribute value is exactly\n       equal to \"bar\"",
			Described: "Attribute\n       selectors",
			Origin:    "2",
		},

		{
			Pattern:   "E[foo~=\"bar\"]",
			Meaning:   "an E element whose \"foo\" attribute value is a list of\n       whitespace-separated values, one of which is exactly equal to \"bar\"",
			Described: "Attribute\n       selectors",
			Origin:    "2",
		},

		{
			Pattern:   "E[foo^=\"bar\"]",
			Meaning:   "an E element whose \"foo\" attribute value begins\n       exactly with the string \"bar\"",
			Described: "Attribute\n       selectors",
			Origin:    "3",
		},

		{
			Pattern:   "E[foo$=\"bar\"]",
			Meaning:   "an E element whose \"foo\" attribute value ends exactly\n       with the string \"bar\"",
			Described: "Attribute\n       selectors",
			Origin:    "3",
		},

		{
			Pattern:   "E[foo*=\"bar\"]",
			Meaning:   "an E element whose \"foo\" attribute value contains the\n       substring \"bar\"",
			Described: "Attribute\n       selectors",
			Origin:    "3",
		},

		{
			Pattern:   "E[foo|=\"en\"]",
			Meaning:   "an E element whose \"foo\" attribute has a\n       hyphen-separated list of values beginning (from the left) with \"en\"",
			Described: "Attribute\n       selectors",
			Origin:    "2",
		},

		{
			Pattern:   "E:root",
			Meaning:   "an E element, root of the document",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:nth-child(n)",
			Meaning:   "an E element, the n-th child of its parent",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:nth-last-child(n)",
			Meaning:   "an E element, the n-th child of its parent, counting\n       from the last one",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:nth-of-type(n)",
			Meaning:   "an E element, the n-th sibling of its type",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:nth-last-of-type(n)",
			Meaning:   "an E element, the n-th sibling of its type, counting\n       from the last one",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:first-child",
			Meaning:   "an E element, first child of its parent",
			Described: "Structural\n       pseudo-classes",
			Origin:    "2",
		},

		{
			Pattern:   "E:last-child",
			Meaning:   "an E element, last child of its parent",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:first-of-type",
			Meaning:   "an E element, first sibling of its type",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:last-of-type",
			Meaning:   "an E element, last sibling of its type",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:only-child",
			Meaning:   "an E element, only child of its parent",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:only-of-type",
			Meaning:   "an E element, only sibling of its type",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:empty",
			Meaning:   "an E element that has no children (including text\n       nodes)",
			Described: "Structural\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:linkbrE:visited",
			Meaning:   "an E element being the source anchor of a hyperlink of\n       which the target is not yet visited (:link) or already visited\n       (:visited)",
			Described: "The link pseudo-classes",
			Origin:    "1",
		},

		{
			Pattern:   "E:activebrE:hoverbrE:focus",
			Meaning:   "an E element during certain user actions",
			Described: "The user action\n       pseudo-classes",
			Origin:    "1 and 2",
		},

		{
			Pattern:   "E:target",
			Meaning:   "an E element being the target of the referring URI",
			Described: "The target pseudo-class",
			Origin:    "3",
		},

		{
			Pattern:   "E:lang(fr)",
			Meaning:   "an element of type E in language \"fr\" (the document\n       language specifies how language is determined)",
			Described: "The :lang() pseudo-class",
			Origin:    "2",
		},

		{
			Pattern:   "E:enabledbrE:disabled",
			Meaning:   "a user interface element E which is enabled or\n       disabled",
			Described: "The UI element states\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E:checked<br>E:indeterminate",
			Meaning:   "a user interface element E which is\n       checkedor in an\n       indeterminate state(for instance a\n       radio-button or checkbox)",
			Described: "The UI element states\n       pseudo-classes",
			Origin:    "3",
		},

		{
			Pattern:   "E::first-line",
			Meaning:   "the first formatted line of an E element",
			Described: "The ::first-line\n       pseudo-element",
			Origin:    "1",
		},

		{
			Pattern:   "E::first-letter",
			Meaning:   "the first formatted letter of an E element",
			Described: "The ::first-letter\n       pseudo-element",
			Origin:    "1",
		},

		{
			Pattern:   "E::before",
			Meaning:   "generated content before an E element",
			Described: "The ::before\n       pseudo-element",
			Origin:    "2",
		},

		{
			Pattern:   "E::after",
			Meaning:   "generated content after an E element",
			Described: "The ::after\n       pseudo-element",
			Origin:    "2",
		},

		{
			Pattern:   "E.warning",
			Meaning:   "an E element whose class is \"warning\" (the document\n       language specifies how class is determined).",
			Described: "Class selectors",
			Origin:    "1",
		},

		{
			Pattern:   "E#myid",
			Meaning:   "an E element with ID equal to \"myid\".",
			Described: "ID selectors",
			Origin:    "1",
		},

		{
			Pattern:   "E:not(s)",
			Meaning:   "an E element that does not match simple selector s",
			Described: "Negation pseudo-class",
			Origin:    "3",
		},

		{
			Pattern:   "E F",
			Meaning:   "an F element descendant of an E element",
			Described: "Descendant\n       combinator",
			Origin:    "1",
		},

		{
			Pattern:   "E > F",
			Meaning:   "an F element child of an E element",
			Described: "Child combinator",
			Origin:    "2",
		},

		{
			Pattern:   "E + F",
			Meaning:   "an F element immediately preceded by an E element",
			Described: "Next-sibling\n       combinator",
			Origin:    "2",
		},

		{
			Pattern:   "E ~ F",
			Meaning:   "an F element preceded by an E element",
			Described: "Subsequent-sibling combinator",
			Origin:    "3",
		},
	}
)
