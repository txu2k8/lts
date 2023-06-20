package s3worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"stress/pkg/bench"
	. "stress/pkg/logger"
	"stress/workflow"
	"stress/workflow/video"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

// Put benchmarks upload speed.
type VideoS3Workflow struct {
	workflow.Common
	video.VideoWorkflow
	// S3Client func() (cl *minio.Client, done func())
	prefixes map[string]struct{}
}

// CreateEmptyBucket will create an empty bucket
// or delete all content if it already exists.
func (u *VideoS3Workflow) CreateEmptyBucket(ctx context.Context) error {
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

// Prepare will create an empty buckets ot delete any content already there.
func (u *VideoS3Workflow) Prepare(ctx context.Context) error {
	Logger.Infof("Stage-Prepare:Create empty buckets: %s%d~%d", u.BucketPrefix, 0, u.BucketNum)
	return nil // u.CreateEmptyBucket(ctx)
}

// Start will execute the main workflow.
// Operations should begin executing when the start channel is closed.
func (u *VideoS3Workflow) Start(ctx context.Context, wait chan struct{}) (bench.Operations, error) {
	var wg sync.WaitGroup
	wg.Add(u.Concurrency)
	c := bench.NewCollector()
	if u.AutoTermDur > 0 {
		ctx = c.AutoTerm(ctx, http.MethodPut, u.AutoTermScale, bench.AutoTermCheck, bench.AutoTermSamples, u.AutoTermDur)
	}
	u.prefixes = make(map[string]struct{}, u.Concurrency)

	// Non-terminating context.
	nonTerm := context.Background()

	for i := 0; i < u.Concurrency; i++ {
		src := u.Source()
		u.prefixes[src.Prefix()] = struct{}{}
		go func(i int) {
			rcv := c.Receiver()
			defer wg.Done()
			opts := u.PutOpts
			done := ctx.Done()

			<-wait
			for {
				select {
				case <-done:
					return
				default:
				}
				obj := src.Object()
				opts.ContentType = obj.ContentType
				client, cldone := u.S3Client()
				op := bench.Operation{
					OpType:   http.MethodPut,
					Thread:   uint16(i),
					Size:     obj.Size,
					File:     obj.Name,
					ObjPerOp: 1,
					Endpoint: client.EndpointURL().String(),
				}
				op.Start = time.Now()
				res, err := client.PutObject(nonTerm, u.Bucket, obj.Name, obj.Reader, obj.Size, opts)
				op.End = time.Now()
				if err != nil {
					u.Error("upload error: ", err)
					op.Err = err.Error()
				}
				obj.VersionID = res.VersionID

				if res.Size != obj.Size && op.Err == "" {
					err := fmt.Sprint("short upload. want:", obj.Size, ", got:", res.Size)
					if op.Err == "" {
						op.Err = err
					}
					u.Error(err)
				}
				op.Size = res.Size
				cldone()
				rcv <- op
			}
		}(i)
	}
	wg.Wait()
	return c.Close(), nil
}

// Cleanup deletes everything uploaded to the bucket.
func (u *VideoS3Workflow) Cleanup(ctx context.Context) {
	var pf []string
	for p := range u.prefixes {
		pf = append(pf, p)
	}
	u.DeleteAllInBucket(ctx, pf...)
}
