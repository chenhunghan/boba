# Contributing to boba

Thanks for your interest — issues and pull requests are welcome.

## Development

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # must print nothing
```

CI runs these (plus `staticcheck` and `go test -race`) across the
supported Go versions on every pull request.

## Pull request titles: Conventional Commits

`boba` releases automatically with
[release-please](https://github.com/googleapis/release-please): the next
version and `CHANGELOG.md` are derived from
[Conventional Commits](https://www.conventionalcommits.org/).

Pull requests are **squash-merged**, so the **PR title** becomes the
commit on `main` and is what release-please reads. A required check (via
`amannn/action-semantic-pull-request`) validates the title.

Format — `<type>[optional scope][!]: <description>`:

```
feat: add tooltip component
fix(tab): keep selection in bounds when the active tab closes
docs: document the focus state machine
feat!: rename Stack.Active to Stack.Focused
```

Effect on the next release (pre-1.0: the API is unstable, so breaking
changes bump the minor, not the major):

| Title prefix | Release |
| --- | --- |
| `fix:` | patch — 0.1.0 → 0.1.1 |
| `feat:` | minor — 0.1.0 → 0.2.0 |
| `feat!:` or a `BREAKING CHANGE:` footer | minor — stays in 0.x |
| `docs:` `chore:` `refactor:` `test:` `ci:` `build:` `style:` | no release |

The `!` is shorthand for a `BREAKING CHANGE:` footer in the PR body.

## Style

- Library packages are fully tested — add or extend tests with your change.
- Keep components **headless**: packages own behavior, callers own
  styling. No color or theme defaults in package code.
- Keep comments to non-obvious rationale or constraints; don't restate
  what the code already says.
