/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-02 14:08:03
 * @LastEditTime: 2023-07-12 00:33:02
 */
package service

import (
	"gpt-meeting-service/internal/biz"
	"gpt-meeting-service/internal/conf"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type ImageService struct {
	iu   *biz.ImageUsecase
	conf *conf.Data
	log  *log.Helper
}

func NewImageService(iu *biz.ImageUsecase, conf *conf.Data, logger log.Logger) *ImageService {
	return &ImageService{
		iu:   iu,
		conf: conf,
		log:  log.NewHelper(logger),
	}
}

func (s *ImageService) UploadFile(ctx http.Context) error {
	req := ctx.Request()

	file, header, err := req.FormFile("file")
	if err != nil {
		return ctx.JSON(200, Resp(400, "file not found", nil))
	}
	defer file.Close()
	// 检查文件大小
	if header.Size > (10 << 20) { // 限制文件大小为10MB
		s.log.Error("file size exceeds the limit")
		return ctx.JSON(200, Resp(400, "file size exceeds the limit", nil))
	}

	// 检查文件类型
	ext := path.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" { // 限制文件类型为jpg和png
		s.log.Error("unsupported file type")
		return ctx.JSON(200, Resp(400, "unsupported file type", nil))
	}
	name := strconv.Itoa(int(time.Now().UnixMicro())) + ext
	// 创建一个新的文件
	// dst, err: = os.(header.Filename)
	imgPath := "/image/" + name
	filePath := s.conf.AssetsPath + imgPath
	dst, err := os.Create(filePath)
	if err != nil {
		s.log.Error(err.Error())
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	defer dst.Close()

	// 将上传的文件内容复制到目标文件中
	if _, err := io.Copy(dst, file); err != nil {
		s.log.Error(err.Error())
		return ctx.JSON(200, Resp(500, err.Error(), nil))
	}
	// imgPath := s.iu.UploadImage(dst, name)
	// if err := os.Remove(name); err != nil {
	// 	s.log.Error(err.Error())
	// }
	// 使用本地存储文件
	s.log.Debugf("imgPath: %s", imgPath)
	return ctx.JSON(200, Resp(200, "success", map[string]string{
		"imageUrl": imgPath,
	}))
}

func Resp(code int, msg string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	}
}
