# Upstream issue and PR review

Reviewed 2026-10-14 against `tomnomnom/gron` open items. This fork cannot close upstream issues. `Fixed` means implemented locally; `Release` depends on published binaries. `Partial`, `Mitigated`, `External`, and `Declined` state remaining limits or scope decisions.

## Issues

| Upstream | Decision | Reason |
| --- | --- | --- |
| [#131](https://github.com/tomnomnom/gron/issues/131) | Fixed | Version comes from module build info or release ldflags. |
| [#129](https://github.com/tomnomnom/gron/issues/129) | Fixed | Ungron accepts tabs around assignments; adapted PR #132. |
| [#127](https://github.com/tomnomnom/gron/issues/127) | Fixed | Replaced dead jsontest examples with GitHub API examples. |
| [#121](https://github.com/tomnomnom/gron/issues/121) | Fixed | Accepts a UTF-8 BOM at the start of JSON input. |
| [#119](https://github.com/tomnomnom/gron/issues/119) | Declined | GitHub CLI download tip is external documentation, not a gron defect. |
| [#118](https://github.com/tomnomnom/gron/issues/118) | Fixed | `--stream --no-sort` now emits keys in input order. |
| [#116](https://github.com/tomnomnom/gron/issues/116) | Declined | Bash and fish completions already exist; a new CLI framework would add needless dependency. |
| [#110](https://github.com/tomnomnom/gron/issues/110) | Partial | Ungron accepts SGR-colored grep output; highlight styling is not retained in JSON. |
| [#108](https://github.com/tomnomnom/gron/issues/108) | Declined | An asdf plugin is a separate distribution project. |
| [#107](https://github.com/tomnomnom/gron/issues/107) | Declined | No public Go library API in this release; CLI and `--json` serve harness integrations. |
| [#105](https://github.com/tomnomnom/gron/issues/105) | Declined | Trailing commas are invalid JSON; accepting them would change the input language. |
| [#104](https://github.com/tomnomnom/gron/issues/104) | Declined | Use `--json` for machine-readable path/value records instead of configurable punctuation. |
| [#103](https://github.com/tomnomnom/gron/issues/103) | Declined | JSON object key order is not semantic; `--no-sort` applies to gron output. |
| [#102](https://github.com/tomnomnom/gron/issues/102) | Fixed | `--no-sort` now preserves input field order via token streaming. |
| [#96](https://github.com/tomnomnom/gron/issues/96) | Declined | JavaScript constructors are outside gron statement grammar. |
| [#94](https://github.com/tomnomnom/gron/issues/94) | Release | Release workflow builds static Linux binaries with `CGO_ENABLED=0`. |
| [#93](https://github.com/tomnomnom/gron/issues/93) | Fixed | Ungron merges incrementally and reuses array storage. |
| [#90](https://github.com/tomnomnom/gron/issues/90) | Release | Fork release will be marked latest. |
| [#87](https://github.com/tomnomnom/gron/issues/87) | Fixed | Later scalar values replace earlier scalar or container values consistently. |
| [#86](https://github.com/tomnomnom/gron/issues/86) | Fixed | `--stream` reads consecutive JSON values. |
| [#81](https://github.com/tomnomnom/gron/issues/81) | Release | Fork publishes tagged releases and binaries. |
| [#79](https://github.com/tomnomnom/gron/issues/79) | Declined | Use `curl` piped to gron for headers; putting API keys in CLI arguments exposes them to process listings. |
| [#73](https://github.com/tomnomnom/gron/issues/73) | Release | GoReleaser matrix includes Linux ARM and ARM64 plus macOS ARM64. |
| [#71](https://github.com/tomnomnom/gron/issues/71) | Fixed | `--ungron --stream` reconstructs JSONL. |
| [#70](https://github.com/tomnomnom/gron/issues/70) | Fixed | Stream mode uses `json.Decoder`, without a Scanner line-size ceiling. |
| [#68](https://github.com/tomnomnom/gron/issues/68) | Declined | The fixed `json` root keeps output interoperable with ungron; text rewrites can use shell tools. |
| [#65](https://github.com/tomnomnom/gron/issues/65) | Declined | Use explicit `--stream` for JSONL; implicit format switching hides malformed input. |
| [#64](https://github.com/tomnomnom/gron/issues/64) | Declined | No public Go library API in this release; CLI and `--json` serve harness integrations. |
| [#63](https://github.com/tomnomnom/gron/issues/63) | Declined | Removing null padding would shift array indexes and change the selected data. |
| [#62](https://github.com/tomnomnom/gron/issues/62) | Mitigated | Use `find` with `gron -H` to retain each file path. |
| [#61](https://github.com/tomnomnom/gron/issues/61) | Mitigated | `--no-sort` streams fields with much lower memory; a huge single scalar still needs memory. |
| [#57](https://github.com/tomnomnom/gron/issues/57) | Declined | Use a YAML-to-JSON tool before gron; YAML adds another grammar and dependency. |
| [#53](https://github.com/tomnomnom/gron/issues/53) | Declined | RPM spec packaging remains downstream of portable release archives. |
| [#51](https://github.com/tomnomnom/gron/issues/51) | External | Snap confinement causes file permission denial; use a native release binary or stdin. |
| [#50](https://github.com/tomnomnom/gron/issues/50) | Mitigated | `--no-sort` streams fields with much lower memory; a huge single scalar still needs memory. |
| [#48](https://github.com/tomnomnom/gron/issues/48) | Fixed | Added `--color` and `--no-color` aliases. |
| [#47](https://github.com/tomnomnom/gron/issues/47) | Fixed | Ungron strips ANSI SGR codes while parsing statements. |
| [#46](https://github.com/tomnomnom/gron/issues/46) | Fixed | Ungron reads arbitrarily long statement lines without Scanner limit. |
| [#39](https://github.com/tomnomnom/gron/issues/39) | Partial | Ungron errors report input line; JSON decoding still reports byte offset. |
| [#38](https://github.com/tomnomnom/gron/issues/38) | Fixed | Release workflow and GoReleaser config use Go modules and tested builds. |
| [#28](https://github.com/tomnomnom/gron/issues/28) | Mitigated | Use `find ... -exec gron -H {} \;` for labeled multi-file output. |
| [#23](https://github.com/tomnomnom/gron/issues/23) | Fixed | `--stream` accepts consecutive JSON values, even without line breaks. |

## Pull requests

| Upstream | Decision | Reason |
| --- | --- | --- |
| [#133](https://github.com/tomnomnom/gron/pull/133) | Applied | Buffered stdout and checked write errors. |
| [#132](https://github.com/tomnomnom/gron/pull/132) | Applied | Tabs accepted by ungron. |
| [#130](https://github.com/tomnomnom/gron/pull/130) | Applied | Added mise instructions, labeled as upstream distribution. |
| [#92](https://github.com/tomnomnom/gron/pull/92) | Replaced | `--stream` decodes consecutive values without loading the whole input. |
| [#91](https://github.com/tomnomnom/gron/pull/91) | Declined | Slurping full input makes large-file memory use worse. |
| [#74](https://github.com/tomnomnom/gron/pull/74) | Declined | YAML belongs in a separate conversion step. |
| [#72](https://github.com/tomnomnom/gron/pull/72) | Adapted | `-H` prefixes lines even when a write contains multiple lines. |
| [#67](https://github.com/tomnomnom/gron/pull/67) | Adapted | Added a native Go round-trip fuzz target without package migration. |
| [#66](https://github.com/tomnomnom/gron/pull/66) | Declined | Broad package move is unnecessary for CLI and fuzz target. |

## Performance spot check

Single local run on macOS arm64, output sent to `/dev/null`; synthetic fixtures, not a general benchmark. For a 3.7 MB JSON array (200,000 objects), `gron --no-sort -m` used 403 MB before and 19 MB after; elapsed time was 0.71 s before and 0.31 s after. For 11.5 MB of gron JSON records (262,144 leaves), `gron -u -j -m` used 1.31 GB before and 200 MB after; elapsed time was 2.38 s before and 2.70 s after. Memory is the primary improvement for ungron.
