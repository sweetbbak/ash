default:
  just build
build:
  go build -o "./bin/readline" ./cmd/readline
release:
  go build -o "./bin/readline" -ldflags='-s -w' ./cmd/readline
static:
  CC="zig cc -target x86_64-linux-musl" go build ./cmd/readline
