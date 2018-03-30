//   Copyright © 2015-2017 Ivan Kostko (github.com/ivan-kostko; github.com/gopot)

//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at

//       http://www.apache.org/licenses/LICENSE-2.0

//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package workerpool // import "github.com/gopot/workerpool"

import (
	"context"
	"sync"
	"time"
)

// Represents WorkerPool default values.
const (
	// Represents default value for worker pool capacity
	DEFAULT_MAX_WORKERS = 1
)

// Represents Worker Pool sateful implementation with lazy initialization.
// All instancies of this type should be better initialized by [Factory Method](#func--newworkerpool).
// However, any instance created not by factory will be initialized with defaulst upon calling [Do](#func-workerpool-do) or [Close](#func-workerpool-close)
type WorkerPool struct {
	doOnce      sync.Once
	workersChan chan struct{}
	closingChan chan struct{}
}

// Instantiates a new worker pool with up to `maxWorkers` worker capacity.
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	wp := &WorkerPool{}
	wp.init(maxWorkers)
	return wp
}

// Schedules a workUnit to be asynchronusly executed by any free worker in the pool unless `ctx` is canceled or [Close()](#func-workerpoll-close) is invoced.
// For convinience `Do` returns when goroutine with `workUnit` is scheduled and started to run.
// Returns the following errors:
//
// * `nil` - if job was successfully scheduled.
//
// * of `interface{ AwaitingCancelledError() }` behaviour in case context was cancelled while awaiting execution slot.
//
// * of `interface{ WorkerPoolClosingError() }` behaviour in case `WorkerPool` closing started while it was waiting execution slot.
func (this *WorkerPool) Do(ctx context.Context, workUnit func()) error {

	this.onEveryCall()

	select {
	case <-ctx.Done():
		// Return AwaitingCancelled error
		return newAwaitingCancelledError("Context was cancelled during awaiting execution slot")
	case <-this.closingChan:
		// Return PoolClosing error
		return newWorkerPoolClosingError("Worker pool is closing")
	case this.workersChan <- struct{}{}:
		started := make(chan struct{})
		go func() {
			defer func() { <-this.workersChan }()
			close(started)

			workUnit()
		}()
		<-started
	}
	return nil

}

// Gracefully closes worker pool.
// It waits until all currently executing work units are complete.
func (this *WorkerPool) Close() {

	this.onEveryCall()

	close(this.closingChan)

	for len(this.workersChan) != 0 {
		<-time.After(1)
	}
	return
}

func (this *WorkerPool) onEveryCall() {
	this.init(DEFAULT_MAX_WORKERS)
}

func (this *WorkerPool) init(maxWorkers int) {
	this.doOnce.Do(func() {
		this.workersChan = make(chan struct{}, maxWorkers)
		this.closingChan = make(chan struct{})
	})
}
