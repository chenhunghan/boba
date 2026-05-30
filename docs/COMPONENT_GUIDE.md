# Component guide

How to add a component to boba. Every component follows the same shape, so
the library stays predictable and a new package never surprises a caller.

## Principles

1. **Headless.** The package owns *behavior* (layout, state, hit-testing,
   key handling) and emits plain strings. The caller owns *styling* via a
   per-instance `Style` field. No package ships a color or theme default.
2. **Value types.** A component is a struct the caller stores on their model
   and replaces with the updated copy returned by its methods. Nothing lives
   in package-level state.
3. **Pure core + ergonomic drop-in.** Expose the pure methods *and* the
   bubbles-style `Update`/`View` pair. The pure core is what the tests pin;
   the drop-in is what most callers use.

## The contract

A component package `foo` exposes:

| Member | When | Shape |
| --- | --- | --- |
| `Foo` struct | always | persistent state + a `Style` field |
| `Style` struct | always | per-state `lipgloss.Style` fields; glyph strings with `""`-fallbacks |
| `Render() string` / `Render(width …) string` | always | pure; exactly the cells it occupies |
| `View() string` | when size is intrinsic | alias for `Render`; matches Bubble Tea / bubbles |
| `Update(tea.Msg) (Foo, tea.Cmd)` | if interactive | routes keys; returns new value + a cmd carrying any event |
| `ApplyKey(key string) …` | if keyboard-driven | the pure key handler `Update` wraps |
| `HitTest(x, y int) …` | if pointer-interactive | panel-local coords → what was hit |
| `ClickAt`/`HoverAt(x, y int)` | if pointer-interactive | panel-local coords from `panel.HitTest`; act / hover |
| `RenderState[…]` | if hover/focus is per-render | passed into `Render`, not stored |
| `<Event>Msg` | if it emits events | e.g. `ToggledMsg`, `ChangedMsg`; delivered via the cmd from `Update`/`ClickAt` |

Rules:

- `Update` is a no-op unless the component is active/focused, and ignores
  messages it doesn't handle.
- Mouse: self-positioned components (those that know their own screen coords,
  like `menu`) may handle `tea.MouseMsg` inside `Update`; panel-positioned
  ones take `ClickAt`/`HoverAt(localX, localY)` fed from `panel.HitTest`.
- Default key bindings live in `nil`-able `[]string` fields, resolved
  internally (`nil → ["enter", " "]`).

## References

- Static, minimal: `separator`
- Interactive (key + mouse + event): `checkbox`, `button`, `menu`
- Per-render hover/focus (`RenderState`): `tab`
- Self-positioned popup (handles mouse in `Update`): `menu`

## Checklist for a new component

- [ ] `foo/foo.go` — struct, `Style`, pure core, `Update`/`View`, event msgs
- [ ] No color/theme defaults; glyphs have `""`-fallbacks
- [ ] `foo/foo_test.go` — pin the pure core (render shape, key handling, hits)
- [ ] `foo/example_test.go` — a runnable `Example` (deterministic `// Output:` where feasible)
- [ ] `go build ./... && go vet ./... && gofmt -l . && go test ./...` clean, and `staticcheck ./...` clean
- [ ] Lean comments: exported-symbol docs + genuine *why* only; no inline narration
- [ ] Commit as `feat(foo): …`
