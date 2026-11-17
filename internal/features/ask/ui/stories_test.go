//go:build gallery

package ui

import (
	"testing"

	"strigil/internal/kit/story/storytest"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories()) }
