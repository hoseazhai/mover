package cloud

import "io"

//Uploader ... 文件上传的统一接口
type Uploader interface {
	Uploader(objectKey string, r io.Reader) (string, error)
}
