package tools

import (
	"log"
	"os"
	"testing"
)

// go test -v oss_test.go oss.go -test.run TestOssUpload

func TestOssUpload(t *testing.T) {
	file, err := os.Open("oss.go")
	if err != nil {
		log.Fatal(err)
	}

	ret := UploadToTWNoExpireOss("oss.go", "text", file)
	if !ret {
		log.Fatal("fail")
	}
}
