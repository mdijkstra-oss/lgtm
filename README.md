# lgtm

> Looks good to me, said by someone who read it.

Ask a question about the repository you are in. A model reads the code and answers with a short document you can step through in the terminal.

![lgtm asked "What happens, step by step, when a request comes in?" in a clone of the chi router](docs/demo.gif)

    cd some/repository
    lgtm "what happens, step by step, when a request comes in?"

Everyone types "lgtm" on a change without having read it, because there is more code than anyone can read. `lgtm` is for understanding the code first, so that saying it is true.

## What you get

- **The answer first.** Two to four sentences, before anything else.
- **The shape of it.** A sequence of steps or a map of the parts. Pick a step and you see what it does and the lines of code that do it, read from your files.
- **A question to try.** One or two at the end, each with two hints and an answer, to check that you now understand it.
- **A short read.** The document is kept to a few hundred words.

While the model works you see what it is doing: the files it reads, what it searches for, and what it is thinking.

The model can only list, search and read files, and read git history. It cannot change anything, and it cannot read files that git ignores. The document is saved in your system's temporary folder, not in the repository.

## Install

You need Go and git.

    git clone https://github.com/mdijkstra-oss/lgtm
    cd lgtm
    go install ./cmd/lgtm

## Configure

`lgtm` needs to know which model to ask. Set these in your environment, or put them in `~/.lgtm/config.json`:

    {"LGTM_MODEL": "…", "OPENAI_BASE_URL": "…", "OPENAI_API_KEY": "…"}

| Name | What it is |
|---|---|
| `LGTM_MODEL` | The model's id, as its provider writes it. Required. |
| `OPENAI_BASE_URL` | An endpoint that speaks OpenAI's API. When set, this is what gets asked, with `OPENAI_API_KEY` as its key. |
| `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENROUTER_API_KEY`, `GEMINI_API_KEY` | Without a base URL, the first of these that is set picks the provider. |
| `LGTM_EFFORT` | How hard the model thinks before it answers: `none`, `minimal`, `low`, `medium` or `high`. `low` when unset. Only sent to an endpoint of your own. |

What is in the environment wins over the file.

## Use

| Command | What it does |
|---|---|
| `lgtm "a question"` | Asks the model and opens its answer. The model is asked every time, since the code may have changed. |
| `lgtm` | Opens the answers that are still in the temporary folder for this repository. |
| `lgtm ask --plain "a question"` | Prints the document instead of opening it. This is also what happens when the output is not a terminal. |
| `lgtm intro` | The opening animation on its own. |

In a document: the arrow keys move and step, a number goes to a section, enter shows more of the code, `q` quits.

## Working on it

See `AGENTS.md` for the make targets and the rules of the repository. `make demo` records the picture at the top again; it asks a real model, so it needs `lgtm` installed and configured.
