//go:build gallery

package viewer

import (
	"testing"

	"lgtm/internal/kit/story/storytest"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories()) }
