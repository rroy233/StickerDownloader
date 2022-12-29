package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rroy233/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type UploadFile struct {
	FilePath  string
	UploadRes *fileHostUploadResp
	InfoRes   *fileHostInfoResp
}

type fileHostUploadResp struct {
	Status bool `json:"status"`
	Data   struct {
		File struct {
			Url struct {
				Full  string `json:"full"`
				Short string `json:"short"`
			} `json:"url"`
			Metadata struct {
				Id   string `json:"id"`
				Name string `json:"name"`
				Size struct {
					Bytes    interface{} `json:"bytes"`
					Readable string      `json:"readable"`
				} `json:"size"`
			} `json:"metadata"`
		} `json:"file"`
	} `json:"data,omitempty"`
	Errors struct {
		File []string `json:"file"`
	} `json:"errors,omitempty"`
}

type fileHostInfoResp struct {
	Status bool `json:"status"`
	Data   struct {
		File struct {
			Url struct {
				Full  string `json:"full"`
				Short string `json:"short"`
			} `json:"url"`
			Metadata struct {
				Id   string `json:"id"`
				Name string `json:"name"`
				Size struct {
					Bytes    interface{} `json:"bytes"`
					Readable string      `json:"readable"`
				} `json:"size"`
			} `json:"metadata"`
		} `json:"file"`
	} `json:"data,omitempty"`
	Errors struct {
		File string `json:"file"`
	} `json:"errors,omitempty"`
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveFile(path string) {
	_ = os.Remove(path)
	return
}

func NewUploadFile(filePath string) *UploadFile {
	return &UploadFile{
		FilePath: filePath,
	}
}

func (f *UploadFile) Upload2FileHost() error {

	file, _ := os.Open(f.FilePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", f.FilePath)
	if err != nil {
		logger.Error.Println("Upload2FileHost-CreateFormFile failed", err)
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Error.Println("Upload2FileHost-io.Copy failed", err)
		return err
	}
	writer.Close()

	r, _ := http.NewRequest("POST", "https://api.anonfiles.com/upload", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(r)
	if err != nil {
		logger.Error.Println("Upload2FileHost-http request err:", err)
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error.Println("Upload2FileHost-failed to read http body:", err)
		return err
	}

	logger.Info.Println("Upload2FileHost-api return:", string(data))

	uploadApiRes := new(fileHostUploadResp)
	err = json.Unmarshal(data, uploadApiRes)
	if err != nil {
		logger.Error.Println("Upload2FileHost-failed to parse body:", err)
		return err
	}
	if uploadApiRes.Status == false {
		logger.Error.Println("Upload2FileHost-api return error:", uploadApiRes.Errors.File)
		return err
	}

	f.UploadRes = uploadApiRes

	//获取信息
	if err := f.getInfo(); err != nil {
		logger.Error.Println("Upload2FileHost-getInfo error:", err)
		return err
	}

	return nil
}

func (f *UploadFile) getInfo() error {
	if f.UploadRes == nil {
		return errors.New("f.UploadRes nil")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.anonfiles.com/v2/file/%s/info", f.UploadRes.Data.File.Metadata.Id), nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Info.Println("Upload2FileHost-api return:", string(data))

	infoApiRes := new(fileHostInfoResp)
	err = json.Unmarshal(data, infoApiRes)
	if err != nil {
		return err
	}
	if infoApiRes.Status == false {
		return err
	}
	f.InfoRes = infoApiRes
	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// Compress 压缩文件
// files 文件数组，可以是不同dir下的文件或者文件夹
// dest 压缩文件存放地址
func Compress(src, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	err = compress(f, "", w)
	if err != nil {
		return err
	}
	return nil
}

func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func GetFileExtName(fileName string) string {
	part := strings.Split(fileName, ".")
	return part[len(part)-1]
}

func CleanTmp() {
	err := os.RemoveAll("./storage/tmp")
	if err != nil {
		logger.Error.Println("failed to clean temp files", err)
	}
}
