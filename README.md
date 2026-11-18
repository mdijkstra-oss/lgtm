# lgtm

> Looks good to me, said by someone who read it.

Everyone says "lgtm" on a change without understanding the code. `lgtm` answers questions about a codebase, so that saying it is true.

    lgtm "how does a request get from the router to the database?"

A model reads the repository you are in and answers with a document: the answer first, then the shape of the thing with the code that shows it, then a few questions to try yourself. The document is kept in the system's temporary folder, outside the repository, so nothing is written into the code you asked about, and the system clears it away in its own time. The model is asked every time, since the code may have changed. Plain `lgtm` opens the documents there are.

Which model is asked is named by the environment, or by `~/.lgtm/config.json`:

    {"LGTM_MODEL": "…", "OPENAI_BASE_URL": "…", "OPENAI_API_KEY": "…"}

`LGTM_MODEL` is the model's id. `LGTM_EFFORT` is how hard it thinks before it answers (`none`, `minimal`, `low`, `medium`, `high`); it is `low` when unset, and only an endpoint of your own is told. With `OPENAI_BASE_URL` it is whatever answers there; otherwise it is the first of `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENROUTER_API_KEY` and `GEMINI_API_KEY` that is set.
