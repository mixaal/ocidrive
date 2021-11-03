package sync

import (
	"encoding/json"
	"log"
	"os"

	"mikc.net/ocidrive/fs"
)

func DumpMap(filename string, m map[string]fs.FsInfo) {
	jsonString, err := json.Marshal(m)
	if err != nil {
		log.Println(err.Error())
	} else {
		fs.LocalWriteFile(filename, []byte(jsonString))
	}
}

func LoadMap(filename string) map[string]fs.FsInfo {
	m := make(map[string]fs.FsInfo)

	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err.Error())
		return m
	}
	err = json.Unmarshal(content, &m)
	if err != nil {
		log.Println(err.Error())
	}
	return m
}

func meta_diff(s_info, d_info fs.FsInfo) bool {
	return s_info.Size != d_info.Size
}

func upload(local_info, os_info fs.FsInfo) bool {
	if meta_diff(local_info, os_info) && local_info.LastModifiedUTC > os_info.LastModifiedUTC {
		return true
	}

	return false
}

func download(os_info, local_info fs.FsInfo) bool {
	if meta_diff(os_info, local_info) && os_info.LastModifiedUTC > local_info.LastModifiedUTC {
		return true
	}

	return false
}

func ToDownload(object_store, local_store map[string]fs.FsInfo) []string {
	o := make([]string, 0)
	for k := range object_store {
		//log.Println("Examining ", k)
		os_info := object_store[k]

		if local_info, ok := local_store[k]; !ok || download(os_info, local_info) {
			o = append(o, k)
		}

		//log.Printf("src_size=%d dst_size=%d\n", src_size, dst_size)
	}

	return o
}

func ToUpload(local_store, object_store map[string]fs.FsInfo) []string {
	o := make([]string, 0)
	for k := range local_store {
		//log.Println("Examining ", k)
		local_info := local_store[k]

		if os_info, ok := object_store[k]; !ok || upload(local_info, os_info) {
			o = append(o, k)
		}

		//log.Printf("src_size=%d dst_size=%d\n", src_size, dst_size)
	}

	return o
}

func MissingOnly(src, dst map[string]fs.FsInfo) []string {
	o := make([]string, 0)
	for k := range src {
		if _, ok := dst[k]; !ok {
			o = append(o, k)
		}
	}

	return o
}
