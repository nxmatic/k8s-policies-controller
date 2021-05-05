# content from git...
set -a
test -z $VERSION && VERSION=$(git describe --tags | sed 's/^v//')
eval KANIKO_FLAGS=\"$KANIKO_FLAGS\"
set +a
