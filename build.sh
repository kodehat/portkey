#!/bin/sh

TARGET_PACKAGE='github.com/kodehat/portkey/internal/build'
BUILD_TIME=$(date -u +"%Y.%m.%d_%H:%M:%S")

commit_hash=
version=dev
go_version=unknown
output=./

while getopts 'o:v:' OPTION; do
  case "$OPTION" in
    o)
      output="$OPTARG"
      ;;
    v)
      version="$OPTARG"
      ;;
    ?)
      echo "script usage: $(basename "$0") [-o output] [-v version]" >&2
      exit 1
      ;;
  esac
done
shift "$((OPTIND -1))"

latest_commit=$(git log -1 --pretty=format:%h || echo 'N/A')
if [ "$latest_commit" = "${latest_commit#fatal}" ]; then
    commit_hash=$latest_commit
else
    commit_hash=
fi

if command -v go > /dev/null 2>&1; then
    go_version=$(go version | sed -nr 's/.*([0-9]+\.[0-9]+\.[0-9]+).*/\1/p')
fi

FLAG="-X $TARGET_PACKAGE.BuildTime=$BUILD_TIME"
FLAG="$FLAG -X $TARGET_PACKAGE.CommitHash=$commit_hash"
FLAG="$FLAG -X $TARGET_PACKAGE.Version=$version"
FLAG="$FLAG -X $TARGET_PACKAGE.GoVersion=$go_version"

echo "[Go v${go_version}] Building portkey in ${output} at ${BUILD_TIME} with commit ${commit_hash:-unknown} in version ${version}."

CGO_ENABLED=0 go build -o "${output}" -ldflags "${FLAG}"