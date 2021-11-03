package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/example/helpers"
	"github.com/oracle/oci-go-sdk/v49/objectstorage"

	"github.com/oracle/oci-go-sdk/v49/common/auth"

	"mikc.net/ocidrive/fs"
	"mikc.net/ocidrive/sync"
)

func getenv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		panic(fmt.Errorf("Environment variable %s not set!", name))
	}
	return val
}

func envAndCompare(name, value string) bool {
	val := os.Getenv(name)
	return val == value
}

func getConfigProvider() common.ConfigurationProvider {
	use_ip := envAndCompare("OCI_DRIVE_AUTH_IP", "true")
	if use_ip {
		provider, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("Using instance principal to authenticate.")
		return provider
	} else {
		return common.DefaultConfigProvider()
	}

}

func main() {

	compartmend_id := getenv("OCI_DRIVE_COMPARTMENT_ID")
	bucket_name := getenv("OCI_DRIVE_BUCKET_NAME")
	local_fs := getenv("OCI_DRIVE_LOCAL_FS")
	drive_id := getenv("OCI_DRIVE_ID")

	calibration_delta := int64(60000) // calibration time tolerance in millis

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	local_snap_file := filepath.Join(home, ".oci", drive_id, "local_snap_file")
	remote_snap_file := filepath.Join(home, ".oci", drive_id, "remote_snap_file")

	fs.LocalCreateDir(local_fs)

	c, clerr := objectstorage.NewObjectStorageClientWithConfigurationProvider(getConfigProvider())
	helpers.FatalIfError(clerr)

	ctx := context.Background()
	namespace := fs.OsGetNamespace(ctx, c)
	log.Printf("Logged under namespace: %s\n", namespace)

	fs.OsFindOrCreateBucket(ctx, c, namespace, compartmend_id, bucket_name)

	sync.Calibration(ctx, c, namespace, bucket_name, local_fs, calibration_delta)

	// last_fs_snaphot := make(map[string]int64)
	// last_os_snapshot := make(map[string]int64)

	last_fs_snaphot := sync.LoadMap(local_snap_file)
	last_os_snapshot := sync.LoadMap(remote_snap_file)

	for {
		sync.RemoveEmptyDirectories(local_fs)
		object_store, err := fs.OsListObjects(ctx, c, namespace, bucket_name)
		if err != nil {
			log.Println("Remote OS: FIX ME, not comparing anything!")
			continue
		}
		local_store, err := fs.LocalListTree(local_fs)
		if err != nil {
			log.Println("LocalFS: FIX ME, not comparing anything!")
			continue
		}

		cont := false
		if len(last_fs_snaphot) > 0 {
			local_diff := sync.MissingOnly(last_fs_snaphot, local_store)
			log.Println("local_diff: ", local_diff)

			if len(local_diff) > 0 {
				sync.RemoveFilesOnObjectStore(ctx, c, namespace, bucket_name, local_fs, local_diff)
				last_fs_snaphot = local_store
				cont = true
			}
		}
		if len(last_os_snapshot) > 0 {
			remote_diff := sync.MissingOnly(last_os_snapshot, object_store)
			log.Println("remote_diff: ", remote_diff)
			if len(remote_diff) > 0 {
				fs.LocalRemoveFiles(local_fs, remote_diff)
				cont = true
				last_os_snapshot = object_store
			}

		}
		if cont {
			continue
		}
		last_fs_snaphot = local_store
		last_os_snapshot = object_store

		log.Println("OS: ", object_store)
		log.Println("LFS: ", local_store)
		toDownload := sync.ToDownload(object_store, local_store)
		toUpload := sync.ToUpload(local_store, object_store)
		log.Println("missing or modified on local store: ", toDownload)
		log.Println("missing or modified on remote store: ", toUpload)

		sync.UploadFilesToObjectStore(ctx, c, namespace, bucket_name, local_fs, toUpload)
		sync.DownloadFilesToLocalStore(ctx, c, namespace, bucket_name, local_fs, toDownload)

		sync.DumpMap(local_snap_file, last_fs_snaphot)
		sync.DumpMap(remote_snap_file, last_os_snapshot)

		time.Sleep(3500 * time.Millisecond)
	}

	// for i := 0; i < len(all_objects); i++ {
	// 	log.Printf("%s %s %d\n", *all_objects[i].Name, all_objects[i].TimeModified.String(), *all_objects[i].Size)
	// }

	//content := getObject(ctx, c, namespace, bucket_name, "main.go")
	//log.Printf("Writing %d bytes ...\n", len(content))
	//localWriteFile("source/prefix/ahoj", content)
	//log.Println("Content written.")

}
