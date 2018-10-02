read -p 'Git Tag: ' VERSION_NUM
read -p 'CircleCI Build #: ' CIRCLE_BUILD_NUM

go build -ldflags "-X main.version=$VERSION_NUM.$CIRCLE_BUILD_NUM" -o cstore_darwin_amd64 