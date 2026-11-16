package search

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func collect(t *testing.T, q Query) []Match {
	t.Helper()
	var got []Match
	if err := Run(context.Background(), q, func(m Match) { got = append(got, m) }); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestRunFindsEveryOccurrence(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "a.txt", "foo bar foo\nnothing here\nFoo\n")

	got := collect(t, Query{Text: "foo", Dir: dir})
	want := []Match{
		{Path: "a.txt", Line: 1, Col: 1, Length: 3, Text: "foo bar foo"},
		{Path: "a.txt", Line: 1, Col: 9, Length: 3, Text: "foo bar foo"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d matches, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("match %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	if n := len(collect(t, Query{Text: "foo", Dir: dir, IgnoreCase: true})); n != 3 {
		t.Errorf("ignore case found %d, want 3", n)
	}
}

func TestRunTreatsQueryAsLiteral(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "a.txt", "a.b\naxb\n")
	if n := len(collect(t, Query{Text: "a.b", Dir: dir})); n != 1 {
		t.Errorf("found %d, want 1", n)
	}
}

func TestRunSkipsHiddenFoldersAndBinaryFiles(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "sub/keep.txt", "needle\n")
	put(t, dir, ".hidden/skip.txt", "needle\n")
	put(t, dir, "blob.bin", "needle\x00needle")

	got := collect(t, Query{Text: "needle", Dir: dir})
	if len(got) != 1 || got[0].Path != filepath.Join("sub", "keep.txt") {
		t.Errorf("got %+v, want only sub/keep.txt", got)
	}
}

func TestRunStopsWhenCancelled(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "a.txt", "needle\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := Run(ctx, Query{Text: "needle", Dir: dir}, func(Match) { calls++ })
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Errorf("err = %v, calls = %d; want context.Canceled and 0", err, calls)
	}
}

func TestRunRejectsMissingFolder(t *testing.T) {
	err := Run(context.Background(), Query{Text: "x", Dir: filepath.Join(t.TempDir(), "nope")}, func(Match) {})
	if err == nil {
		t.Error("want an error for a missing folder")
	}
}
