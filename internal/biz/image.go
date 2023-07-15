/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-02 14:56:55
 * @LastEditTime: 2023-07-15 11:56:10
 */
package biz

import (
	"context"
	"gpt-meeting-service/internal/conf"
	"net/http"
	"net/url"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type ImageUsecase struct {
	conf *conf.Data
	log  *log.Helper
}

func NewImageUsecase(conf *conf.Data, logger log.Logger) *ImageUsecase {
	return &ImageUsecase{
		conf: conf,
		log:  log.NewHelper(logger),
	}
}

func (i *ImageUsecase) UploadImage(fd *os.File, name string) string {

	u, _ := url.Parse(i.conf.TxCos.Url)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  i.conf.TxCos.SecretId,  // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
			SecretKey: i.conf.TxCos.SecretKey, // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
		},
	})
	path := "meeting/" + name

	_, err := c.Object.PutFromFile(context.Background(), path, name, nil)
	if err != nil {
		i.log.Errorf("failed upload image. " + err.Error())
	}
	return c.Object.GetObjectURL(path).String()
}
