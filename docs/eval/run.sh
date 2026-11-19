#!/bin/sh
# Renders what lgtm writes about a fixed set of commits of go-chi/chi to
# files, so that the documents can be read side by side and judged.
#
#   docs/eval/run.sh out/round-1
#
# Each commit gets <hash>.md (the document), <hash>.log (what the model
# did, and what the checker refused) and <hash>.diff (the commit, to judge
# the document against). It asks a real model, so lgtm has to be installed
# and configured. COMMITS overrides the set.
set -eu
out=${1:?where to put the documents}
repo=${LGTM_EVAL_REPO:-/tmp/lgtm-demo/chi}
commits=${COMMITS:-"aba8618 9b6ddcd 87c34a2 1576723 3d1777a caf87e6"}

[ -d "$repo" ] || git clone -q --depth 300 https://github.com/go-chi/chi "$repo"
mkdir -p "$out/tmp"
out=$(cd "$out" && pwd)

one() {
	start=$(date +%s)
	(cd "$repo" && TMPDIR="$out/tmp" lgtm ask --plain "$1" >"$out/$1.md" 2>"$out/$1.log") || echo "FAILED" >>"$out/$1.log"
	git -C "$repo" show --stat --patch "$1" >"$out/$1.diff"
	echo "$1: $(( $(date +%s) - start ))s, $(wc -w <"$out/$1.md" | tr -d ' ') words, $(grep -c '^✗ ' "$out/$1.log" || true) refused"
}

# Three at a time: more than that and the endpoint slows every one of them.
n=0
for c in $commits; do
	one "$c" &
	n=$((n + 1))
	[ $((n % 3)) -ne 0 ] || wait
done
wait
