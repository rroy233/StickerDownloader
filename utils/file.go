package utils

import (
	"archive/zip"
	"bufio"
	"crypto/md5"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/rroy233/logger.v2"
	"io"
	"os"
	"strings"
)

const MB = 1 << 20

type UploadFile struct {
	ZipPath    string
	FolderPath string
	CleanList  []string
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
		_, err = SendFileByPath(update, fmt.Sprintf("%s_part-%d.zip", f.FolderPath, i))
		if err != nil {
			logger.Error.Println("UploadFragment SendFile error", err)
		}
	}

	return err
}

func (f *UploadFile) UploadSingle(update *tgbotapi.Update) error {
	_, err := SendFileByPath(update, f.ZipPath)
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
