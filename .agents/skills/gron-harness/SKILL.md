---
name: gron-harness
description: Inspect unfamiliar JSON tool output and design compact, path-aware responses for agent harnesses. Use when a tool returns large nested JSON or its output contract needs fewer tokens.
---

# Gron for harness tools

Use `gron` to discover paths in sample JSON. Keep the tool's production output structured and bounded.

1. Capture a representative response with secrets removed. Check input size before running `gron`. Default mode reads each JSON value into memory; `--no-sort` streams fields but still holds each scalar value.
2. Explore paths with `gron --json response.json`. Each line is `[path, value]`, where `path` is an array of keys and indexes. For consecutive JSON values or JSONL, use `gron --stream --json`.
3. Choose fields the agent actually needs. Filter by parsed path segments, not a text search that might match a value. Keep types and source indexes intact.
4. Return matching path/value records with an explicit limit and `truncated` flag. Redact secrets before tool output reaches the model. Test with nested arrays, missing fields, large values, and invalid JSON.

For an exploratory shell query:

```sh
gron --json response.json | jq -c 'select(.[0][-1] == "id")' | head -n 20
```

Do not send a full gron dump when a narrower query works. Discard partial stream output when `gron` exits nonzero. Do not evaluate gron's JavaScript-like text output. `gron --ungron` rebuilds filtered arrays with `null` in missing positions, so it is unsuitable for selecting a compact list. Omit `--no-sort` when paths must have sorted order across inputs.
