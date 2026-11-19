# The recording at the top of README.md: lgtm in a clone of a router many
# people know (github.com/go-chi/chi), picking a commit from its history
# and reading what comes back. make demo makes the clone and runs this. It
# asks a real model, so it needs lgtm on the PATH and a model configured.
Output docs/demo.gif

Require lgtm

Set Shell zsh
Set Width 1300
Set Height 1000
Set FontSize 15
Set Framerate 20
Set TypingSpeed 90ms
Set WaitTimeout 120s

# The clone has a temporary folder of its own, emptied by make demo, so
# that no commit was explained before and the model is seen at work.
Hide
Type "cd /tmp/lgtm-demo && export TMPDIR=/tmp/lgtm-demo/tmp && clear"
Enter
Show

Sleep 600ms
Type "cd chi"
Sleep 300ms
Enter
Sleep 500ms
Type "lgtm"
Sleep 500ms
Enter

# The page opens under the slop, and the blade takes about three seconds
# to cross. Then the history: down a few rows, and narrowed to one commit
# by a word of it.
Sleep 5s
Down
Sleep 350ms
Down
Sleep 350ms
Down
Sleep 900ms
Type "duplicate"
Sleep 1500ms
Enter

# The model at work, for a few seconds. The rest of the wait is cut out of
# the recording: it takes the model about twenty seconds.
Sleep 4s
Hide

# The index beside a document is headed "contents" and a rule: the
# document is open. The word alone can be on the page before that.
Wait+Screen /contents ─/
Show
Sleep 4500ms

# What it did before, and what it does now.
Type "2"
Sleep 1200ms
Down@250ms 2
Sleep 3s
Right
Sleep 3500ms

# How it works, a step at a time.
Type "3"
Sleep 1200ms
Down@250ms 2
Sleep 3s
Right
Sleep 3500ms

# The last section is the question to answer before accepting, whichever
# number it has: the question, a hint, and the answer.
Hide
Type "4567"
Show
Sleep 1200ms
Down@250ms 2
Sleep 3500ms
Right
Sleep 2500ms
Right
Sleep 2s
Right
Sleep 4500ms

Type "q"
Sleep 400ms
