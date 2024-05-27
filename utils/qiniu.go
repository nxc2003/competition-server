package utils

import (
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/cdn"
	"github.com/qiniu/go-sdk/v7/storage"
	"time"
)

// 七牛云的访问密钥、桶名和域名
var (
	accessKey = "your_access_key"
	secretKey = "your_secret_key"
	bucket    = "your_bucket"
	domain    = "your_domain"
)

// 初始化 Mac、配置、桶管理器和 CDN 管理器
var mac = qbox.NewMac(accessKey, secretKey)
var cfg = storage.Config{}
var bucketManager = storage.NewBucketManager(mac, &cfg)
var cdnManager = cdn.NewCdnManager(mac)

// GetToken 生成上传令牌
func GetToken(name string) string {
	putPolicy := storage.PutPolicy{
		Scope:   fmt.Sprintf("%s:%s", bucket, name),
		Expires: 3600, // 令牌有效期为1小时
	}
	return putPolicy.UploadToken(mac)
}

// GetFileUrl 生成下载链接
func GetFileUrl(filename string) string {
	deadline := time.Now().Add(time.Minute).Unix() // 链接有效期为1分钟
	return storage.MakePrivateURL(mac, domain, filename, deadline)
}

// RefreshUrl 刷新 CDN 缓存
func RefreshUrl(name string) error {
	urls := []string{fmt.Sprintf("%s/%s", domain, name)}
	_, err := cdnManager.RefreshUrls(urls)
	return err
}

// FileInfo 文件信息结构体
type FileInfo struct {
	Fsize    int64
	Hash     string
	MimeType string
	PutTime  int64
	Type     int
}

// GetFileInfo 获取文件信息
func GetFileInfo(name string) (*FileInfo, error) {
	fileInfo, err := bucketManager.Stat(bucket, name)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Fsize:    fileInfo.Fsize,
		Hash:     fileInfo.Hash,
		MimeType: fileInfo.MimeType,
		PutTime:  fileInfo.PutTime,
		Type:     fileInfo.Type,
	}, nil
}

// DeleteFile 删除文件
func DeleteFile(names []string) error {
	deleteOps := make([]string, len(names))
	for i, name := range names {
		deleteOps[i] = storage.URIDelete(bucket, name)
	}

	_, err := bucketManager.Batch(deleteOps)
	return err
}
