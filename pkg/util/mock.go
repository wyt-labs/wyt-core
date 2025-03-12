package util

import (
	"runtime"

	"github.com/tryvium-travels/memongo"
)

// MockMongoServer mock a mongo server, but remember to stop the server
func MockMongoServer() (*memongo.Server, error) {
	opts := &memongo.Options{
		MongoVersion: "6.0.1",
	}
	if runtime.GOARCH == "arm64" {
		if runtime.GOOS == "darwin" {
			// Only set the custom url as workaround for arm64 macs
			opts.DownloadURL = "https://fastdl.mongodb.org/osx/mongodb-macos-x86_64-5.0.0.tgz"
		}
	}
	return memongo.StartWithOptions(opts)
}
