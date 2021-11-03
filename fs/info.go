package fs

type FsInfo struct {
	OsMd5           *string
	FsMd5           *string
	LastModifiedUTC int64
	Size            int64
}
