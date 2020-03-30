package aliyun

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
)

//AliOSS 结构体定义
type AliOSS struct {
	Endpoint, AccessKey, AccessSecret string
	Bucket                            *oss.Bucket
}

// NewAliOSS ... 新建阿里云OSS 对象
func NewAliOSS(bucket, endpoint, key, secret string) (*AliOSS, error) {
	client, err := oss.New(endpoint, key, secret)
	if err != nil {
		return nil, err
	}
	bucketObject, err := client.Bucket(bucket)
	if err != nil {
		return nil, err
	}
	return &AliOSS{
		Endpoint:     endpoint,
		AccessKey:    key,
		AccessSecret: secret,
		Bucket:       bucketObject,
	}, nil
}

// 实现Upoad ... 实现upload方法
func (ali *AliOSS) Uploader(objectKey string, r io.Reader) (string, error) {
	err := ali.Bucket.PutObject(objectKey, r)
	if err != nil {
		return "", err
	}
	return objectKey, nil
}
