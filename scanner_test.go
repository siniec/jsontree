package jsontree

import (
	"errors"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		in   string
		want []*Node
		err  error
	}{
		{
			in:  `"a":{"b":{"c":"v1","d":{"e":"v2","f":"v3"}},"g":{"h":"v4"}}`,
			err: errors.New(`Invalid input: did not start with '{'. Found '"'`),
		},
		{
			in:  `{a":{"b":{"c":"v1","d":{"e":"v2","f":"v3"}},"g":{"h":"v4"}}`,
			err: errors.New(`Invalid input: JSON object is incomplete (EOF reached)`),
		},
		{
			in:  `{"a":{"b":{"c":"v1","d":{"e":"v2","f":"v3"}`,
			err: errors.New(`Invalid input: JSON object is incomplete (EOF reached)`),
		},
		{
			in:  `{"a":{"b":"v1"}`,
			err: errors.New(`Invalid input: JSON object is incomplete (EOF reached)`),
		},
		{
			in:  `{"a":{"b":"v1"}-}`,
			err: errors.New(`Invalid input: unknown token encountered: '-'. Expected ',' or '}'`),
		},
		{
			in:   `{}`,
			want: []*Node{},
		},
		{
			in: `{"a":{"b":"v1","c":{"d":"v2","e":"v3"}},"f":{"g":"v4"},"h":{"i":{"j":"v5"}}}`,
			want: []*Node{
				&Node{Key: "a", Nodes: []*Node{&Node{Key: "b", Value: "v1"}, &Node{Key: "c", Nodes: []*Node{&Node{Key: "d", Value: "v2"}, &Node{Key: "e", Value: "v3"}}}}},
				&Node{Key: "f", Nodes: []*Node{&Node{Key: "g", Value: "v4"}}},
				&Node{Key: "h", Nodes: []*Node{&Node{Key: "i", Nodes: []*Node{&Node{Key: "j", Value: "v5"}}}}},
			},
		},
		{
			in: `{"1":{"c1":{"r1":{"c":"A"}},"c2":{"r2":{"d":"B"}},"c3":{"r1":{"d":"C"}}},"2":{"c1":{"r1":{"d":"D"}}}}`,
			want: []*Node{
				&Node{Key: "1", Nodes: []*Node{
					&Node{Key: "c1", Nodes: []*Node{
						&Node{Key: "r1", Nodes: []*Node{
							&Node{Key: "c", Value: "A"},
						}},
					}},
					&Node{Key: "c2", Nodes: []*Node{
						&Node{Key: "r2", Nodes: []*Node{
							&Node{Key: "d", Value: "B"},
						}},
					}},
					&Node{Key: "c3", Nodes: []*Node{
						&Node{Key: "r1", Nodes: []*Node{
							&Node{Key: "d", Value: "C"},
						}},
					}},
				}},
				&Node{Key: "2", Nodes: []*Node{
					&Node{Key: "c1", Nodes: []*Node{
						&Node{Key: "r1", Nodes: []*Node{
							&Node{Key: "d", Value: "D"},
						}},
					}},
				}},
			},
		},
	}
	for _, test := range tests {
		r := strings.NewReader(test.in)
		scanner := NewScanner(r)
		var got []*Node
		for scanner.Scan() {
			got = append(got, scanner.Node())
		}
		err := scanner.Err()
		if test.err != nil {
			if !errEqual(test.err, err) {
				t.Errorf("%s\nWrong error.\nWant %v\nGot  %v", test.in, test.err, err)
			}
			if got != nil {
				t.Errorf("%s\nWrong nodes returned. Want none, got %v", test.in, nodesString(got))
			}
			continue
		} else {
			if err != nil {
				t.Errorf("%s\nUnexpected error %v", test.in, err)
				continue
			}
		}
		want := test.want
		if nGot, nWant := len(got), len(want); nGot != nWant {
			t.Errorf(test.in+"\n"+"Wrong number of nodes (%d vs %d)\nWant%v\nGot%v", nWant, nGot, nodesString(want), nodesString(got))
		} else {
			for i := range want {
				if !nodeEqual(want[i], got[i]) {
					t.Errorf(test.in+"\n"+"Wrong node at index %d\nWant%v\nGot%v", i, nodesString(want), nodesString(got))
				}
			}
		}
	}
}
