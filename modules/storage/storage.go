// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"io"

	"code.gitea.io/gitea/modules/setting"
)

// ObjectStorage represents an object storage to handle a bucket and files
type ObjectStorage interface {
	Save(path string, r io.Reader) (int64, error)
	Open(path string) (io.ReadCloser, error)
	Delete(path string) error
}

// Copy copys a file from source ObjectStorage to dest ObjectStorage
func Copy(dstStorage ObjectStorage, dstPath string, srcStorage ObjectStorage, srcPath string) (int64, error) {
	f, err := srcStorage.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return dstStorage.Save(dstPath, f)
}

var (
	// Attachments represents attachments storage
	Attachments ObjectStorage
)

// Init init the stoarge
func Init() error {
	var err error
	switch setting.Attachment.StoreType {
	case "local":
		Attachments, err = NewLocalStorage(setting.Attachment.Path)
	case "minio":
		minio := setting.Attachment.Minio
		Attachments, err = NewMinioStorage(
			minio.Endpoint,
			minio.AccessKeyID,
			minio.SecretAccessKey,
			minio.Bucket,
			minio.Location,
			minio.BasePath,
			minio.UseSSL,
		)
	default:
		return fmt.Errorf("Unsupported attachment store type: %s", setting.Attachment.StoreType)
	}

	if err != nil {
		return err
	}

	return nil
}
