package sync

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/oracle/oci-go-sdk/v49/objectstorage"

	"mikc.net/ocidrive/fs"
)

func Calibration(ctx context.Context, c objectstorage.ObjectStorageClient, ns, bucket, local_fs string, delta int64) {
	log.Println("Calibration begins ...")
	cal_key := ".calibration"
	calibration_file := filepath.Join(local_fs, cal_key)
	fs.LocalWriteFile(calibration_file, []byte("example content"))
	local_modified, err := fs.LocalListTree(local_fs)
	if err != nil {
		log.Fatal(err.Error())
	}
	calibration_info, ok := local_modified[cal_key]
	if !ok {
		log.Fatal("can't find calibration file on local system")
	}
	local_modified_utc := calibration_info.LastModifiedUTC
	log.Printf("Local file last modified UTC: %d\n", local_modified_utc)
	UploadFilesToObjectStore(ctx, c, ns, bucket, local_fs, []string{cal_key})
	os_modified, err := fs.OsListObjects(ctx, c, ns, bucket)
	if err != nil {
		fs.LocalRemoveFiles(local_fs, []string{cal_key})
		log.Fatal(err.Error())
	}
	calibration_info, ok = os_modified[cal_key]
	if !ok {
		fs.LocalRemoveFiles(local_fs, []string{cal_key})
		log.Fatal("can't find calibration file on object store")
	}
	os_modified_utc := calibration_info.LastModifiedUTC
	log.Printf("Object Store file last modified UTC: %d\n", os_modified_utc)

	time_diff := math.Abs(float64(os_modified_utc) - float64(local_modified_utc))
	if time_diff > float64(delta) {
		fs.LocalRemoveFiles(local_fs, []string{cal_key})
		log.Printf("time difference measured: %f (tolerance=%d)\n", time_diff, delta)
		log.Fatal("local and remote time differ too much, please sync the time first")
	}
	fs.LocalRemoveFiles(local_fs, []string{cal_key})
	err = fs.OsDeleteObject(ctx, c, ns, bucket, cal_key)
	if err != nil {
		log.Println(err.Error())
	}
}

func UploadFilesToObjectStore(ctx context.Context, c objectstorage.ObjectStorageClient, ns, bucket, local_fs string, files []string) {
	if files == nil {
		return
	}
	for _, file := range files {
		log.Println("Uploading ", file)
		abs_path := filepath.Join(local_fs, file)
		content, err := os.ReadFile(abs_path)
		if err == nil {
			l := int64(len(content))
			rc := io.NopCloser(bytes.NewReader(content))
			fs.OsPutObject(ctx, c, ns, bucket, fs.ConvertPathToSlash(file), l, rc, nil)
		}
	}
}

func DownloadFilesToLocalStore(ctx context.Context, c objectstorage.ObjectStorageClient, ns, bucket, local_fs string, files []string) {
	if files == nil {
		return
	}
	for _, file := range files {
		//abs_path := filepath.Join(local_fs, fs.ConvertPathToBackSlash(file))
		abs_path := filepath.Join(local_fs, file)
		content := fs.OsGetObject(ctx, c, ns, bucket, fs.ConvertPathToSlash(file))
		if content != nil {
			fs.LocalWriteFile(abs_path, content)
		}
	}
}

func RemoveFilesOnObjectStore(ctx context.Context, c objectstorage.ObjectStorageClient, ns, bucket, local_fs string, files []string) {
	if files == nil {
		return
	}
	for _, file := range files {
		log.Println("Deleting ", file)
		err := fs.OsDeleteObject(ctx, c, ns, bucket, file)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func processEmptyDir(path string, info os.FileInfo) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("Checking directory ", path)
	log.Println("Checking directory files ", files)

	if len(files) != 0 {
		return
	}

	now := time.Now().UTC()
	modified := info.ModTime().UTC()
	if time.Duration(now.Sub(modified).Seconds()) < 60 {
		// Empty directory with last modification time in less than 1 min ago, return for now
		return
	}

	err = os.Remove(path)
	if err != nil {
		log.Println(err.Error())
	}

	log.Println("Removed empty directory ", path)
}

func RemoveEmptyDirectories(local_fs string) {
	err := filepath.Walk(local_fs,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path != local_fs {
					processEmptyDir(path, info)
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err.Error())
	}
}
