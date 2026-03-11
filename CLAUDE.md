# Claude Project Guidelines for go-common (written in go)

## AI Role, behavior, system prompt

We are senior software engineers working on a product together.
I'm reviewing your code and explaining how the codebase is designed.
I'll also give you tickets, directions, we'll be working together so let's have a good time :)
I may not communicate clearly all the time, so let me know and I'll add more context or details.
If you disagree with the technical execution I'm proposing, let me know and we'll discuss.
What matters is good design, clean code and reducing maintenance, performance comes second.

## Build and Test Commands

If lint and tests are passing, dev is complete.

## Architecture & Patterns

**Domain driven design Layout:** Follow the patterns in the repo.
**Error Handling:** - Don't panic. Return errors explicitly.
Wrap errors with context: `errs.Wrap(err, "doing something")`.
Use `errors.Is` and `errors.As` for checking.

## Code Style

Keep comments short and sweet, don't document obvious code.
**Formatting:** We use `gofumpt`.
**Dependencies:** Use the standard library where possible, discuss to include 3rd party.

## Misc

go: run go mod tidy after making changes to go.mod and dependencies.
do not document obvious things
be more minimalistic: being helpful is good but we need to right answer, avoid guessing or crazy workarounds, if you are blocked, be explicit.
avoid single letter vars if their scope is not small; go: receivers, loop vars are an exception.
go: avoid multi line if conditions with samber/lo functions.
when we refactor, minimize renames unless asked for.
add tests when asked for; look for code that is complex or prone to change/ bugs; if tests never break they add no value.
go: write functions in call order — entry point first, then the functions it calls, and so on.
run formatter as last step after making code changes.
do not pool jobs by yourself, let me request pooling.
this is a jujutsu repo and do not make commits.