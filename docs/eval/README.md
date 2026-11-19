# eval

How to tell whether what lgtm writes about a commit is any good.

There is no automatic score. `run.sh` renders the documents for a fixed set
of commits of go-chi/chi to files, and a reader (a person, or a model given
this page) judges each one against its commit with the rubric below. What
goes wrong in several documents at once is a fault of `prompt.md` or of the
checker, and is fixed there; one document's slip is not.

    make install            # or: go install ./cmd/lgtm
    docs/eval/run.sh out/round-1

Each commit gets its document (`.md`), what the model did and what the
checker refused (`.log`), and the commit itself (`.diff`).

## What the document is for

A person has to understand a commit quickly, before they read the diff
themselves. After thirty seconds they know what behaves differently, what
kind of change it is and where it is. After three minutes they could sketch
the diff and say how the new code works.

The document explains what is there. It is not a review: it does not look
for bugs, missing tests or risks, and it reports none. Other tools do that.
What the judge checks hard is the other kind of error: whether the document
itself is right about the code.

## Rubric

Score each from 1 to 5.

| | 5 | 1 |
|---|---|---|
| Opening | Says what behaves differently, what kind of change it is and where it is, in three sentences | Restates the commit message, or describes edits to code |
| Order | What changed comes first; context comes after, if at all | The reader goes through background to reach the change |
| Before and after | A real input and what came back, before and now; the old code is shown when code was removed | Describes the edit, not the behaviour |
| Accuracy | Every line range, test name and claim checks out against the commit and the files | A wrong range, a test that does not assert the case, an invented reason |
| Explains, does not review | Describes what the code did, does, and how; says what the tests check without judging what they leave out | Reports faults, gaps or risks; marks things right or wrong; tells the reader what to worry about |
| How it works | A reader who never opened the file could follow one input through the new code, step by step | Names functions without saying what happens in them |
| Proportion | A 20-line commit gets a document a fraction the size of a 300-line one | The same template whatever the size |
| Load | One idea on screen at a time; the names of the steps and cases read as the findings; nothing is said twice | Stacked blocks, long code, the same fact in three places |
| Voice | A senior engineer talking to a colleague: plain words, literal | Slogans, balanced contrasts, code described as a person |

## The commits

| commit | what it is | size |
|---|---|---|
| `9b6ddcd` | Don't duplicate methods in the Allow header of a 405 | 16 lines |
| `87c34a2` | ContentCharset skips requests without a body | 44 lines |
| `aba8618` | Panic when a router is mounted onto itself | 45 lines |
| `caf87e6` | Support the HTTP QUERY method | 69 lines, 5 files |
| `1576723` | Add a Clone method to Context | 80 lines |
| `3d1777a` | Honor Discard when flushing | 100 lines, rewrites code |
