package main
import (
	"testing"
	"fmt"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input string
		expected []string
	}{
		{
			input: "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},{
			input: "  hello  world  ",
			expected: []string{"hello", "world"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			fmt.Printf("not the same number of words")
			fmt.Printf("actual: %v", actual)
			fmt.Printf("expected: %v", c.expected)
			continue
		}
		for i := range actual {
			if actual[i] != c.expected[i] {
				t.Errorf("mismatch: %s != %s", actual[i], c.expected[i])
				return
			}
		}
	}
	return
}
