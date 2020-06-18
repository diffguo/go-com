package tools

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/diffguo/gocom/log"
	"io"
)

type OssSignParam struct {
	Method      string
	ContentMd5  string
	ContentType string
	Date        string
	CHeader     map[string]string
	CResource   string
}

var AccessKeyID = ""
var AccessKeySecret = ""
var OssHouseBucket *oss.Bucket = nil

func InitOss(accessKeyID, accessKeySecret string) {
	AccessKeyID = accessKeyID
	AccessKeySecret = accessKeySecret
}

func UploadToTWNoExpireOss(resourcePath string, contentType string, reader io.Reader) bool {
	if OssHouseBucket == nil {
		client, err := oss.New("oss-cn-chengdu.aliyuncs.com", AccessKeyID, AccessKeySecret)
		if err != nil {
			log.Errorf("init oss client error: %s", err.Error())
			return false
		}

		OssHouseBucket, err = client.Bucket("tw-erp-no-expire")
		if err != nil {
			log.Errorf("init oss bucket error: %s", err.Error())
			return false
		}
	}

	options := []oss.Option{
		oss.ContentType(contentType),
		oss.CacheControl("max-age=31536000"), /*缓存365天*/
	}

	signedURL, err := OssHouseBucket.SignURL(resourcePath, oss.HTTPPut, 60, options...)
	if err != nil {
		if err != nil {
			log.Errorf("init oss sign url error: %s", err.Error())
			return false
		}
	}

	err = OssHouseBucket.PutObjectWithURL(signedURL, reader, options...)
	if err != nil {
		log.Errorf("upload house res err: %s", err.Error())
		return false
	}

	return true
}

func DeleteTWOssRes(resourcePath string) bool {
	if OssHouseBucket == nil {
		client, err := oss.New("https://oss-cn-hangzhou.aliyuncs.com", AccessKeyID, AccessKeySecret)
		if err != nil {
			log.Errorf("init oss client error: %s", err.Error())
			return false
		}

		OssHouseBucket, err = client.Bucket("yjl-house-res")
		if err != nil {
			log.Errorf("init oss bucket error: %s", err.Error())
			return false
		}
	}

	err := OssHouseBucket.DeleteObject(resourcePath)
	if err != nil {
		log.Errorf("delete house res err: %s", err.Error())
		return false
	}

	return true
}
