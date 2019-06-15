package file

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jery1024/fserver/web"
	"github.com/jery1024/mlog"
)

var rootDistMap, fileDistMap sync.Map

func RegistRootDist(name, rootPath string) error {
	if ret, ok := fileDistMap.Load(name); ok {
		return fmt.Errorf("RegistFileDist failed with exist name, rootPath: %v", ret)
	}
	if name == "" {
		return fmt.Errorf("RegistFileDist failed with invalid name, name: %v", name)
	}
	if rootPath == "" {
		return fmt.Errorf("RegistFileDist failed with invalid rootPath, rootPath: %v", rootPath)
	}
	if !filepath.IsAbs(rootPath) {
		if ret, err := filepath.Abs(rootPath); err != nil {
			return fmt.Errorf("RegistFileDist failed with invalid rootPath, rootPath: %v, detail: %v", rootPath, err.Error())
		} else {
			rootPath = ret
		}
	}
	rootDistMap.Store(name, rootPath)
	mlog.Infof("RegistRootDist successed, name: %v, rootPath: %v", name, rootPath)
	return nil
}

func RegistFileDist(webPath, distPath string) error {
	if ret, ok := fileDistMap.Load(webPath); ok {
		return fmt.Errorf("RegistFileDist failed with exist webPath, distPath: %v", ret)
	}
	if webPath == "" {
		return fmt.Errorf("RegistFileDist failed with invalid webPath, webPath: %v", webPath)
	}
	if distPath == "" {
		return fmt.Errorf("RegistFileDist failed with invalid distPath, distPath: %v", distPath)
	}
	fileDistMap.Store(webPath, distPath)
	mlog.Infof("RegistFileDist successed, webPath: %v, distPath: %v", webPath, distPath)
	return nil
}

func GetFileDist(webPath string) string {
	if ret, ok := fileDistMap.Load(webPath); ok {
		return fmt.Sprintf("%v", ret)
	}
	return ""
}

func GetRootDist(name string) string {
	if ret, ok := rootDistMap.Load(name); ok {
		return fmt.Sprintf("%v", ret)
	}
	return ""
}

/*
Title Upload
router /upload [post]

base64Image: ""
uploadFiles: file(s)
*/
func UploadFile(w http.ResponseWriter, r *http.Request, params url.Values) error {
	helper := &resUploadHelper{params: params, request: r, respwriter: w}
	helper.verify()
	// helper.saveImage()
	helper.saveFileLocal()
	return helper.result()
}

type resUploadHelper struct {
	params     url.Values
	request    *http.Request
	respwriter http.ResponseWriter

	imageName   string
	base64Image string

	// single file
	file       multipart.File
	fileHeader *multipart.FileHeader

	// multi files
	files []*multipart.FileHeader

	fileList map[string]string

	derr error
	oerr string
}

func (p *resUploadHelper) verify() {
	if p.derr != nil {
		return
	}

	p.fileList = make(map[string]string, 0)
	// p.imageName = p.params.Get("imageName")
	// p.base64Image = p.params.Get("base64Image")
	// if p.base64Image != "" {
	// 	return
	// }

	// single file
	file, fileHeader, err := p.request.FormFile("uploadFile")
	if err != nil {
		p.derr = fmt.Errorf("verify uploadFile failed, detail: %v", err.Error())
		mlog.Errorf("MultipartForm.File: %+v", p.request.MultipartForm)
		return
	}
	p.file = file
	p.fileHeader = fileHeader
}

func (p *resUploadHelper) saveImage() {
	if p.derr != nil || p.base64Image == "" {
		return
	}
	sessID := web.SessionID(p.request)

	// 验证是否为合法的图片内容
	if !strings.Contains(p.imageName, ".") {
		p.derr = fmt.Errorf("saveImage failed with unsupport image, imageName: %v", p.imageName)
		return
	}
	imgDist := GetFileDist(p.request.URL.Path)
	imgDist = strings.Replace(imgDist, "sessionID", sessID, -1)

	filename := p.imageName
	visitPath := path.Join(imgDist, filename)

	// if ExistRes(filepath.Join(GetCurrentAbsPath(), "public", sessID, "dist", "img", filename)) {
	if ExistRes(filepath.Join(imgDist, sessID, "img", filename)) {
		p.fileList[visitPath] = filepath.Join(imgDist, sessID, "img", p.fileHeader.Filename)
		return
	}

	saveImgFile := filepath.Join(imgDist, sessID, "img", filename)
	if err := SaveForBase64(saveImgFile, p.base64Image); err != nil {
		p.derr = fmt.Errorf("saveImage failed with invalid base64 image, base64Image: %v", p.base64Image)
		return
	}
	p.fileList[visitPath] = filepath.Join(imgDist, sessID, "img", p.fileHeader.Filename)
}

func (p *resUploadHelper) saveFileLocal() {
	if p.derr != nil || p.fileHeader == nil {
		return
	}
	sessID := web.SessionID(p.request)
	server := web.GetStatus(web.C_Status_Server, p.request).(*web.TServer)

	rootDist := GetRootDist(server.Name)
	fileDist := GetFileDist(p.request.URL.Path)
	visitPath := path.Join(path.Base(rootDist), sessID, fileDist, p.fileHeader.Filename)
	rootDist = filepath.Join(rootDist, sessID)

	// if ExistRes(filepath.Join(GetCurrentAbsPath(), "public", sessID, "dsist/res", p.fileHeader.Filename)) {
	if ExistRes(filepath.Join(rootDist, fileDist, p.fileHeader.Filename)) {
		p.file.Close()
		p.fileList[visitPath] = filepath.Join(rootDist, fileDist, p.fileHeader.Filename)
		return
	}

	//create destination file making sure the path is writeable.
	dst, err := os.Create(filepath.Join(rootDist, fileDist, p.fileHeader.Filename))
	defer dst.Close()
	if err != nil {
		p.derr = fmt.Errorf("saveFileLocal failed while doing os.Create, detail: %v", err.Error())
		return
	}

	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, p.file); err != nil {
		p.derr = fmt.Errorf("saveFileLocal failed while doing io.Copy, detail: %v", err.Error())
		return
	}
	p.file.Close()
	p.fileList[visitPath] = filepath.Join(rootDist, fileDist, p.fileHeader.Filename)
}

func (p *resUploadHelper) result() error {
	ret := struct {
		Error    string
		FileList map[string]string
	}{}

	if p.derr != nil {
		mlog.Error(p.derr.Error())
		ret.Error = "UploadFile failed"
		if p.oerr != "" {
			ret.Error = p.oerr
		}
		web.ResponseError(p.respwriter, p.request, ret.Error)
		return p.derr
	}

	ret.FileList = p.fileList
	web.ResponseOk(p.respwriter, p.request, ret)
	return nil
}

func GetCurrentAbsPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err.Error()
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func ExistRes(fileOrPath string) bool {
	_, err := os.Stat(fileOrPath)
	return err == nil
}

func SaveForBase64(filePath, content string) error {
	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("SaveToFile failed, filePath:%v, detail:%v", filePath, err.Error())
	}
	defer file.Close()

	// decode by base64
	contentBuf, base64Err := base64.StdEncoding.DecodeString(content)
	if base64Err != nil {
		return base64Err
	}

	_, err = file.Write(contentBuf)
	return err
}
