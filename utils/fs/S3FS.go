package fs

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type S3FS struct {
	Endpoint  string
	AccessKey string
	SecretKey string

	Region string
	Bucket string
}

func (s3fs *S3FS) GetFile(path string) ([]byte, error) {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	svc := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})
	_, err = svc.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s3fs.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s3fs *S3FS) PutFile(path string, data []byte) error {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	svc := s3manager.NewUploader(sess)

	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3fs.Bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s3fs *S3FS) PutFileStream(path string, data io.ReadCloser) error {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	svc := s3manager.NewUploader(sess)

	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3fs.Bucket),
		Key:    aws.String(path),
		Body:   data,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s3fs *S3FS) DeleteFile(path string) error {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	svc := s3.New(sess)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s3fs.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s3fs *S3FS) DeleteFolder(path string) error {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	svc := s3.New(sess)

	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(s3fs.Bucket),
		Prefix: aws.String(path),
	})

	return s3manager.NewBatchDeleteWithClient(svc).Delete(context.Background(), iter)
}

func (s3fs *S3FS) DeleteList(objects []string) error {
	for _, obj := range objects {
		err := s3fs.DeleteFile(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s3fs *S3FS) DeleteFolderAsList(path string) error {
	list, err := s3fs.ListFolder(path)
	if err != nil {
		return err
	}
	err = s3fs.DeleteList(list)
	return err
}

func (s3fs *S3FS) ListFolder(path string) ([]string, error) {
	creds := credentials.NewStaticCredentials(s3fs.AccessKey, s3fs.SecretKey, "")
	cfg := aws.NewConfig().WithEndpoint(s3fs.Endpoint).WithRegion(s3fs.Region).WithCredentials(creds).WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess)

	s3Keys := make([]string, 0)

	err = svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(s3fs.Bucket),
		Prefix: aws.String(path),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			s3Keys = append(s3Keys, *obj.Key)
		}
		return true
	})
	return s3Keys, err
}

func NewS3FS(Props map[string]string) *S3FS {
	return &S3FS{
		Endpoint:  Props["endpoint"],
		AccessKey: Props["access_key"],
		SecretKey: Props["secret"],
		Region:    Props["region"],
		Bucket:    Props["bucket"],
	}
}
