# The recording at the top of README.md: a question a newcomer would ask,
# about a small router many people know (github.com/go-chi/chi). make demo
# clones it and runs this. It asks a real model, so it needs lgtm on the
# PATH and a model configured.
Output docs/demo.gif

Require lgtm

Set Shell zsh
Set Width 1300
Set Height 800
Set FontSize 15
Set Framerate 20
Set TypingSpeed 45ms
Set WaitTimeout 120s

Hide
Type "cd /tmp/lgtm-demo && clear"
Enter
Show

Sleep 600ms
Type "cd chi"
Sleep 300ms
Enter
Sleep 500ms
Type 'lgtm "What happens, step by step, when a request comes in?"'
Sleep 700ms
Enter

# The index beside a document says "contents": the document is open.
Wait+Screen /contents/
Sleep 3s

# The first section, and through the parts of its drawing.
Type "2"
Sleep 1500ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 800ms
Right
Sleep 2200ms
Right
Sleep 2200ms
Right
Sleep 2500ms

# The last section is a practice question, whichever number it has.
Hide
Type "34567"
Show
Sleep 1500ms
Down
Sleep 250ms
Down
Sleep 1500ms
Right
Sleep 2s
Right
Sleep 2s
Right
Sleep 4s

Type "q"
Sleep 400ms
