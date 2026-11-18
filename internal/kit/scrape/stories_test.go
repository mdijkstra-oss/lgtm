//go:build gallery

package scrape

import (
	"testing"

	"lgtm/internal/kit/story/storytest"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories()) }
