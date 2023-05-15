package video

import (
	"context"
	"fmt"
	"net/http"
	"s3stress/pkg/bench"
	"s3stress/pkg/workflow"
	"sync"
	"time"
)

// Put benchmarks upload speed.
type VideoWorkflow struct {
	workflow.Common
	VideoInfo
	prefixes map[string]struct{}
}

// Init will create empty buckets
func (u *VideoWorkflow) Init(ctx context.Context) error {
	// 输入参数计算分析
	u.CalcData()

	fmt.Printf("Stage-Init:Create empty buckets: %s%d~%d", u.BucketPrefix, 0, u.BucketNum)
	return nil // u.CreateEmptyBucket(ctx)
}

// Prepare will create an empty bucket ot delete any content already there.
func (u *VideoWorkflow) Prepare(ctx context.Context) error {
	u.CalcData()

	fmt.Printf("Stage-Init:Create empty buckets: %s%d~%d", u.BucketPrefix, 0, u.BucketNum)
	return nil
}

// Start will execute the main workflow.
// Operations should begin executing when the start channel is closed.
func (u *VideoWorkflow) Start(ctx context.Context, wait chan struct{}) (bench.Operations, error) {
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
				client, cldone := u.Client()
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
func (u *VideoWorkflow) Cleanup(ctx context.Context) {
	var pf []string
	for p := range u.prefixes {
		pf = append(pf, p)
	}
	u.DeleteAllInBucket(ctx, pf...)
}
