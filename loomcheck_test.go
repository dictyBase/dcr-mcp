package main

import (
	"testing"

	"github.com/dictyBase/fp-go-loom/loomcheck"
)

func TestNoHandRolledPredicates(t *testing.T) {
	loomcheck.Require(t, loomcheck.Config{Roots: []string{"."}})
}
