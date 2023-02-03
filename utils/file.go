package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const MB = 1 << 20

type UploadFile struct {
	ZipPath    string
	FolderPath string
	UploadRes  *fileHostUploadResp
	InfoRes    *fileHostInfoResp
	CleanList  []string
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

func NewUploadFile(zipPath, folderPath string) *UploadFile {
	return &UploadFile{
		ZipPath:    zipPath,
		FolderPath: folderPath,
	}
}

func (f *UploadFile) CheckAvailable() bool {
	req, err := http.NewRequest(http.MethodHead, "https://api.anonfiles.com", nil)
	if err != nil {
		logger.Error.Println("CheckAvailable - NewRequest error", err)
		return false
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error.Println("CheckAvailable - client.Do error", err)
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}

func (f *UploadFile) Upload2FileHost() error {

	file, _ := os.Open(f.ZipPath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", f.ZipPath)
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

	if resp.StatusCode != 200 {
		logger.Error.Println("Upload2FileHost-api return:", string(data))
		logger.Error.Println("Upload2FileHost-http request Status:", resp.StatusCode)
		return errors.New("http request Status " + resp.Status)
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

func (f *UploadFile) UploadFragment(update *tgbotapi.Update) error {
	dir, err := os.ReadDir(f.FolderPath)
	if err != nil {
		return err
	}

	//创建0号目录
	folderIndex := 0
	err = os.Mkdir(fmt.Sprintf("%s_%d", f.FolderPath, folderIndex), 0755)
	if err != nil {
		logger.Error.Println("UploadFragment os.Mkdir error", err)
		return err
	}
	f.CleanList = append(f.CleanList, fmt.Sprintf("%s_%d", f.FolderPath, folderIndex))

	//获取所有文件的大小
	sizes := make([]int64, len(dir))
	for i, entry := range dir {
		info, err := entry.Info()
		if err != nil {
			logger.Error.Println("UploadFragment entry.Info() error", err)
			return err
		}
		sizes[i] = info.Size()
	}

	//操作文件
	sizeSum := int64(0)
	for i, entry := range dir {
		if sizeSum+sizes[i] > 49*MB {
			sizeSum = 0
			folderIndex++
			err = os.Mkdir(fmt.Sprintf("%s_%d", f.FolderPath, folderIndex), 0755)
			if err != nil {
				logger.Error.Println("UploadFragment os.Mkdir error", err)
				return err
			}
			f.CleanList = append(f.CleanList, fmt.Sprintf("%s_%d", f.FolderPath, folderIndex))
		}
		err = CopyFile(fmt.Sprintf("%s/%s", f.FolderPath, entry.Name()), fmt.Sprintf("%s_%d/%s", f.FolderPath, folderIndex, entry.Name()))
		if err != nil {
			logger.Error.Println("UploadFragment CopyFile error", err)
			continue
		}
		sizeSum += sizes[i]
	}

	//压缩并上传
	for i := 0; i <= folderIndex; i++ {
		err = Compress(fmt.Sprintf("%s_%d", f.FolderPath, i), fmt.Sprintf("%s_part-%d.zip", f.FolderPath, i))
		if err != nil {
			logger.Error.Println("UploadFragment Compress error", err)
		}
		f.CleanList = append(f.CleanList, fmt.Sprintf("%s_part-%d.zip", f.FolderPath, i))

		SendAction(GetChatID(update), ChatActionSendDocument)
		err = SendFile(update, fmt.Sprintf("%s_part-%d.zip", f.FolderPath, i))
		if err != nil {
			logger.Error.Println("UploadFragment SendFile error", err)
		}
	}

	return err
}

func (f *UploadFile) Clean() {
	for _, s := range f.CleanList {
		err := os.RemoveAll(s)
		if err != nil {
			logger.Error.Println("UploadFile.Clean error", err)
		}
	}
	return
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

func CopyFile(src, dst string) error {
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

func MD5File(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	//bufferSize = 65536
	for buf, reader := make([]byte, 65536), bufio.NewReader(file); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		hash.Write(buf[:n])
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum, nil
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

func GetChatID(update *tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return -1
}
