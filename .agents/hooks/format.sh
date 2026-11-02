#!/usr/bin/env bash
# PostToolUse hook: format the file an agent just edited.
# .go -> golangci-lint fmt (gofumpt + gci, as CI), falling back to gofmt when only golangci-lint v1
# is installed. .tf -> terraform fmt. Silent no-op when the tool is missing.
set -u

file=$(python3 -I -c 'import json,sys; print(json.load(sys.stdin).get("tool_input",{}).get("file_path",""))' 2>/dev/null)
[[ -n "$file" && -f "$file" ]] || exit 0

case "$file" in
  *.go)
    golangci-lint fmt "$file" >/dev/null 2>&1 || { command -v gofmt >/dev/null && gofmt -w "$file"; }
    ;;
  *.tf)
    command -v terraform >/dev/null && terraform fmt "$file" >/dev/null
    ;;
esac
exit 0
