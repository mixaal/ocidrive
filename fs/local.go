package fs

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func LocalListTree(directory string) (map[string]FsInfo, error) {
	abs_dir, _ := filepath.Abs(directory)
	if !strings.HasSuffix(abs_dir, string(os.PathSeparator)) {
		abs_dir = abs_dir + string(os.PathSeparator)
	}
	log.Printf("Resolved path %s to absolute %s", directory, abs_dir)
	file_map := make(map[string]FsInfo)
	err := filepath.Walk(abs_dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				trimmed := strings.TrimPrefix(path, abs_dir)
				log.Println(trimmed, info.Size(), info.ModTime())
				file_map[trimmed] = FsInfo{
					Size:            info.Size(),
					LastModifiedUTC: info.ModTime().UTC().UnixMilli(),
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err.Error())
		return file_map, err
	}
	return file_map, nil
}

func LocalWriteFile(filename string, content []byte) {
	basename := filepath.Base(filename)
	log.Println("LocalWriteFile: base=", basename)
	log.Println("LocalWriteFile: file=", filename)

	if basename != filename {
		dirname := filepath.Dir(filename)
		log.Println("LocalWriteFile: dir=", dirname)
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			log.Println(err.Error())
		}
	}

	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		log.Println(err.Error())
	}
}

func LocalCreateDir(filename string) {
	_, err := os.Stat(filename)
	if err == nil {
		return
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, 0755)
	}
}

func LocalRemoveFiles(local_fs string, files []string) {
	if files == nil {
		return
	}

	for _, file := range files {
		abs_path := filepath.Join(local_fs, file)
		err := os.Remove(abs_path)
		if err != nil {
			log.Println(err.Error())
		}
	}

}
