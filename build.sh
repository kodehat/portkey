#!/bin/sh

TARGET_PACKAGE='github.com/kodehat/portkey/internal/build'
BUILD_TIME=$(date -u +"%Y.%m.%d_%H:%M:%S")

commit_hash=
version=${1:-dev}
go_version=unknown

latest_commit=$(git log -1 --pretty=format:%h || echo 'N/A')
if [[ latest_commit =~ 'fatal' ]];
then
    commit_hash=
else
    commit_hash=$latest_commit
fi

if [[ $(go version) =~ [0-9]+\.[0-9]+\.[0-9]+ ]];
then
    go_version=${BASH_REMATCH[0]}
fi

FLAG="-X $TARGET_PACKAGE.BuildTime=$BUILD_TIME"
FLAG="$FLAG -X $TARGET_PACKAGE.CommitHash=$commit_hash"
FLAG="$FLAG -X $TARGET_PACKAGE.Version=$version"
FLAG="$FLAG -X $TARGET_PACKAGE.GoVersion=$go_version"

echo "[Go v${go_version}] Building portkey at ${BUILD_TIME} with commit ${commit_hash:-unknown} in version ${version}."

CGO_ENABLED=0 go build -ldflags "${FLAG}"