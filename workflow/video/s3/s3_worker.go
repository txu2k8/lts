package s3worker

import (
	"context"
	"errors"
	"stress/workflow/video"

	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

// Put benchmarks upload speed.
type S3VideoWorkflow struct {
	video.VideoWorkflow
	S3Client func() (cl *minio.Client, done func())
}

// CreateEmptyBucket will create an empty bucket
// or delete all content if it already exists.
func (u *S3VideoWorkflow) CreateEmptyBucket(ctx context.Context) error {
	cl, done := u.S3Client()
	defer done()
	x, err := cl.BucketExists(ctx, u.Bucket)
	if err != nil {
		return err
	}

	if x && u.Locking {
		_, _, _, err := cl.GetBucketObjectLockConfig(ctx, u.Bucket)
		if err != nil {
			if !u.Clear {
				return errors.New("not allowed to clear bucket to re-create bucket with locking")
			}
			if bvc, err := cl.GetBucketVersioning(ctx, u.Bucket); err == nil {
				u.Versioned = bvc.Status == "Enabled"
			}
			console.Eraseline()
			console.Infof("\rClearing Bucket %q to enable locking...", u.Bucket)
			u.DeleteAllInBucket(ctx)
			err = cl.RemoveBucket(ctx, u.Bucket)
			if err != nil {
				return err
			}
			// Recreate bucket.
			x = false
		}
	}

	if !x {
		console.Eraseline()
		console.Infof("\rCreating Bucket %q...", u.Bucket)
		err := cl.MakeBucket(ctx, u.Bucket, minio.MakeBucketOptions{
			Region:        u.Location,
			ObjectLocking: u.Locking,
		})
		// In client mode someone else may have created it first.
		// Check if it exists now.
		// We don't test against a specific error since we might run against many different servers.
		if err != nil {
			x, err2 := cl.BucketExists(ctx, u.Bucket)
			if err2 != nil {
				return err2
			}
			if !x {
				// It still doesn't exits, return original error.
				return err
			}
		}
	}
	if bvc, err := cl.GetBucketVersioning(ctx, u.Bucket); err == nil {
		u.Versioned = bvc.Status == "Enabled"
	}

	if u.Clear {
		console.Eraseline()
		console.Infof("\rClearing Bucket %q...", u.Bucket)
		u.DeleteAllInBucket(ctx)
	}
	return nil
}
