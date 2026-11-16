# frame

A bordered panel on the panel colour: an optional head row, a body, an
optional foot row, with a rule between each.

- The caller gives the overall size; `BodySize` says how much of it the body gets.
- The caller places it. A frame knows nothing about position or margins.
- Head, body rows and foot arrive already drawn. The frame pads or cuts them to fit.

Used by the palette and by the file view of a search.
