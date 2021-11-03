package fs

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/example/helpers"
	"github.com/oracle/oci-go-sdk/v49/objectstorage"
)

func ConvertPathToSlash(filename string) string {
	sep := string(os.PathSeparator)
	if sep != "/" && strings.Contains(filename, sep) {
		conv := strings.ReplaceAll(filename, sep, "/")
		log.Printf("convertPathToSlash(): %s -> %s", filename, conv)
		return conv
	} else {
		log.Printf("convertPathToSlash(): %s -> %s", filename, filename)
		return filename
	}
}

func ConvertPathToBackSlash(filename string) string {
	sep := string(os.PathSeparator)
	if sep == "\\" && strings.Contains(filename, "/") {
		conv := strings.ReplaceAll(filename, "/", sep)
		log.Printf("ConvertPathToBackSlash(): %s -> %s", filename, conv)
		return conv
	} else {
		log.Printf("ConvertPathToBackSlash(): %s -> %s", filename, filename)
		return filename
	}
}

func OsGetNamespace(ctx context.Context, c objectstorage.ObjectStorageClient) string {
	request := objectstorage.GetNamespaceRequest{}
	r, err := c.GetNamespace(ctx, request)
	helpers.FatalIfError(err)
	return *r.Value
}

func OsPutObject(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, bucketname, objectname string, contentLen int64, content io.ReadCloser, metadata map[string]string) error {
	request := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketname,
		ObjectName:    &objectname,
		ContentLength: &contentLen,
		PutObjectBody: content,
		OpcMeta:       metadata,
	}
	_, err := c.PutObject(ctx, request)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func OsDeleteObject(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, bucketname, objectname string) (err error) {
	request := objectstorage.DeleteObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketname,
		ObjectName:    &objectname,
	}
	_, err = c.DeleteObject(ctx, request)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func createBucket(ctx context.Context, c objectstorage.ObjectStorageClient, compartment_id, namespace, name string) {
	request := objectstorage.CreateBucketRequest{
		NamespaceName: &namespace,
	}
	request.CompartmentId = common.String(compartment_id)
	request.Name = &name
	request.Metadata = make(map[string]string)
	request.PublicAccessType = objectstorage.CreateBucketDetailsPublicAccessTypeNopublicaccess
	_, err := c.CreateBucket(ctx, request)
	helpers.FatalIfError(err)
}

func deleteBucket(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, name string) (err error) {
	request := objectstorage.DeleteBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &name,
	}
	_, err = c.DeleteBucket(ctx, request)
	helpers.FatalIfError(err)
	return err
}

func listBuckets(ctx context.Context, c objectstorage.ObjectStorageClient, namespace string, compartment_id string, limit int) (resp objectstorage.ListBucketsResponse) {
	opc_request_id, _ := uuid.NewRandom()
	log.Printf("opc_request_id: %s", opc_request_id.String())

	req := objectstorage.ListBucketsRequest{Page: nil,
		CompartmentId:      common.String(compartment_id),
		Fields:             []objectstorage.ListBucketsFieldsEnum{objectstorage.ListBucketsFieldsTags},
		Limit:              common.Int(limit),
		NamespaceName:      common.String(namespace),
		OpcClientRequestId: common.String(opc_request_id.String()),
	}

	// Send the request using the service client
	resp, err := c.ListBuckets(ctx, req)
	helpers.FatalIfError(err)

	// Retrieve value from the response.
	// fmt.Println(resp)
	return resp
}

func OsFindOrCreateBucket(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, compartment_id, bucket_name string) {
	found := false

	buckets := listBuckets(ctx, c, namespace, compartment_id, 1000).Items

	for i := 0; i < len(buckets); i++ {
		if *(buckets[i].Name) == bucket_name {
			found = true
			break
		}
	}
	if !found {
		log.Println("Creating bucket ...")
		createBucket(ctx, c, compartment_id, namespace, bucket_name)
		log.Println("Bucket created.")
	}
}

func OsListObjects(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, bucket_name string) (map[string]FsInfo, error) {
	opc_request_id, _ := uuid.NewRandom()
	log.Printf("opc_request_id: %s", opc_request_id.String())

	var next_start_with *string = nil

	file_map := make(map[string]FsInfo)

	for {
		req := objectstorage.ListObjectsRequest{Fields: common.String("timeCreated, timeModified, size"),
			NamespaceName:      common.String(namespace),
			OpcClientRequestId: common.String(opc_request_id.String()),
			// Prefix:             common.String("EXAMPLE-prefix-Value"),
			Start:      next_start_with,
			BucketName: common.String(bucket_name),
			// Delimiter:  common.String("/"),
			// StartAfter: common.String("EXAMPLE-startAfter-Value"),
			// End:        common.String("EXAMPLE-end-Value"),
			Limit: common.Int(1000)}

		// Send the request using the service client
		resp, err := c.ListObjects(ctx, req)
		if err != nil {
			log.Println(err.Error())
			return file_map, err
		}

		//log.Println(resp)
		objs := resp.ListObjects.Objects
		for i := 0; i < len(objs); i++ {
			conv := ConvertPathToBackSlash(*objs[i].Name)
			file_map[conv] = FsInfo{
				Size:            *objs[i].Size,
				LastModifiedUTC: objs[i].TimeModified.UTC().UnixMilli(),
				OsMd5:           objs[i].Md5,
			}
		}
		if resp.NextStartWith == nil {
			break
		}
		next_start_with = resp.NextStartWith
	}
	return file_map, nil
}

func OsGetObject(ctx context.Context, c objectstorage.ObjectStorageClient, namespace, bucket, object_name string) []byte {
	opc_request_id, _ := uuid.NewRandom()
	//log.Printf("opc_request_id: %s", opc_request_id.String())

	req := objectstorage.GetObjectRequest{BucketName: common.String(bucket),
		// HttpResponseContentDisposition: common.String("EXAMPLE-httpResponseContentDisposition-Value"),
		// IfNoneMatch:                    common.String("EXAMPLE-ifNoneMatch-Value"),
		ObjectName: common.String(object_name),
		// OpcSseCustomerKeySha256:     common.String("EXAMPLE-opcSseCustomerKeySha256-Value"),
		OpcClientRequestId: common.String(opc_request_id.String()),
		// HttpResponseContentEncoding: common.String("EXAMPLE-httpResponseContentEncoding-Value"),
		// HttpResponseContentType: common.String("application/json"),
		// HttpResponseExpires:         common.String("EXAMPLE-httpResponseExpires-Value"),
		// IfMatch:                     common.String("EXAMPLE-ifMatch-Value"),
		NamespaceName: common.String(namespace)}
	// OpcSseCustomerKey:           common.String("EXAMPLE-opcSseCustomerKey-Value"),
	// VersionId:                   common.String("ocid1.test.oc1..<unique_ID>EXAMPLE-versionId-Value"),
	// HttpResponseCacheControl:    common.String("EXAMPLE-httpResponseCacheControl-Value"),
	// HttpResponseContentLanguage: common.String("EXAMPLE-httpResponseContentLanguage-Value"),
	// OpcSseCustomerAlgorithm:     common.String("EXAMPLE-opcSseCustomerAlgorithm-Value")}

	// Send the request using the service client
	resp, err := c.GetObject(ctx, req)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	//helpers.FatalIfError(err)
	log.Println("Downloading content from ", object_name)
	// Retrieve value from the response.
	rc := resp.Content
	defer rc.Close()

	buf := new(bytes.Buffer)
	nbytes, err := buf.ReadFrom(resp.Content)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	log.Printf("Read: %d", nbytes)
	return buf.Bytes()
}
