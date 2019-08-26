package main

import (
	"strings"
	"testing"
)

var getJoinTestingTable = []struct {
	in       string
	expected string
}{
	{"likipiki", "@likipiki"},
	{"", "Здравствуй"},
	{"testing", "@testing"},
}

func TestGetJoin(t *testing.T) {
	for _, el := range getJoinTestingTable {
		result := GetJoin(el.in)
		if !strings.HasPrefix(result, el.expected) {
			t.Errorf("got %q, expected %q", result, el.expected)
		}
	}
}
