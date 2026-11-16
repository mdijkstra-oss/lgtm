# shell

The interactive interface: open tabs, the palette over them, the keys and the
footer. It imports only `kit`, so it shows any feature without knowing it.

- `shell.go`, `view.go`: the model and how it draws.
- `cli/`: `strigil tui`, which hands the features to the shell.
