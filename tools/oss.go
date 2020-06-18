package tools

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/diffguo/gocom/log"
	"io"
)

type OssBucket struct {
	bucket *oss.Bucket
}

func InitOssBucket(endPoint, accessKeyID, accessKeySecret, bucketName string) (*OssBucket, error) {
	client, err := oss.New(endPoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Errorf("init oss client error: %s", err.Error())
		return nil, err
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Errorf("init oss bucket error: %s", err.Error())
		return nil, err
	}

	return &OssBucket{bucket: bucket}, nil
}

func (bucket *OssBucket) UploadToTWNoExpireOss(resourcePath string, contentType string, reader io.Reader) bool {
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.CacheControl("max-age=31536000"), /*缓存365天*/
	}

	signedURL, err := bucket.bucket.SignURL(resourcePath, oss.HTTPPut, 60, options...)
	if err != nil {
		if err != nil {
			log.Errorf("init oss sign url error: %s", err.Error())
			return false
		}
	}

	err = bucket.bucket.PutObjectWithURL(signedURL, reader, options...)
	if err != nil {
		log.Errorf("upload house res err: %s", err.Error())
		return false
	}

	return true
}

func (bucket *OssBucket) DeleteTWOssRes(resourcePath string) bool {
	err := bucket.bucket.DeleteObject(resourcePath)
	if err != nil {
		log.Errorf("delete house res err: %s", err.Error())
		return false
	}

	return true
}
