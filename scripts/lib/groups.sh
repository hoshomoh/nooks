#!/usr/bin/env bash
# Which of ci.sh's groups a set of changed paths could possibly break.
#
# Sourced by preflight.sh, and exercised on its own by groups.test.sh. It is a pure
# function of the paths: no git, no filesystem, so the mapping can be tested by naming
# files that do not exist.
#
# The rule is one-way. Naming a group that did not need running costs minutes; missing
# one costs a red CI after a green local run, which is the whole thing preflight exists
# to prevent. So anything unrecognised runs everything, and every doubt resolves that
# way.

ALL_GROUPS="go proto shared web website"

# groupsFor prints the groups the given paths could affect, in ci.sh's own order.
groupsFor() {
  local wanted="" path

  for path in "$@"; do
    case "$path" in
      # The Go server and its dependencies.
      cmd/* | internal/* | server/* | store/* | go.mod | go.sum)
        wanted="$wanted go" ;;

      # The contract every client is generated from. A change here reaches the Go
      # server, the generated TypeScript client, and everything built on it.
      proto/*)
        wanted="$wanted $ALL_GROUPS" ;;

      # The generated client. Both apps import it; the Go side does not.
      packages/api/*)
        wanted="$wanted shared web website" ;;

      # Shared source both apps compile.
      packages/shared/*)
        wanted="$wanted shared web website" ;;

      # Tokens and the mark, imported by both apps.
      packages/design/*)
        wanted="$wanted web website" ;;

      apps/web/*)
        wanted="$wanted web" ;;

      apps/website/*)
        wanted="$wanted website" ;;

      # The changelog page reads this file at build time.
      CHANGELOG.md)
        wanted="$wanted website" ;;

      # Prose nothing builds from. The three files that are not read by any build step
      # are named rather than matched, so a new root file gets everything by default.
      README.md | CONTRIBUTING.md | AGENTS.md | CONTEXT.md | DESIGN.md | STANDARDS.md | LICENCE | LICENSE)
        ;;

      # Lockfiles, workspace layout, the checks themselves, and anything unrecognised.
      *)
        wanted="$wanted $ALL_GROUPS" ;;
    esac
  done

  # Deduplicated, and in ci.sh's order rather than the order paths happened to arrive.
  local group out=""
  for group in $ALL_GROUPS; do
    case " $wanted " in *" $group "*) out="$out $group" ;; esac
  done
  echo "${out# }"
}
