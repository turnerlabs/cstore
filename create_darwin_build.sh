read -p 'Git Tag: ' VERSION_NUM
read -p 'GitHub Actions Build #: ' BUILD_NUM

cd cli 

go build -ldflags "-X main.version=$VERSION_NUM.$BUILD_NUM" -o ../cstore_darwin_amd64 