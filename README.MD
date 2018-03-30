# workerpool
--
    import "github.com/gopot/workerpool"


## Usage

```go
const (
	// Represents default value for worker pool capacity
	DEFAULT_MAX_WORKERS = 1
)
```
Represents WorkerPool default values.

#### type WorkerPool

```go
type WorkerPool struct {
}
```

Represents Worker Pool sateful implementation with lazy initialization. All
instancies of this type should be better initialized by [Factory
Method](#func--newworkerpool). However, any instance created not by factory will
be initialized with defaulst upon calling [Do](#func-workerpool-do) or
[Close](#func-workerpool-close)

#### func  NewWorkerPool

```go
func NewWorkerPool(maxWorkers int) *WorkerPool
```
Instantiates a new worker pool with up to `maxWorkers` worker capacity.

#### func (*WorkerPool) Close

```go
func (this *WorkerPool) Close()
```
Gracefully closes worker pool. It waits until all currently executing work units
are complete.

#### func (*WorkerPool) Do

```go
func (this *WorkerPool) Do(ctx context.Context, workUnit func()) error
```
Schedules a workUnit to be asynchronusly executed by any free worker in the pool
unless `ctx` is canceled or [Close()](#func-workerpoll-close) is invoced. For
convinience `Do` returns when goroutine with `workUnit` is scheduled and started
to run. Returns the following errors:

* `nil` - if job was successfully scheduled.

* of `interface{ AwaitingCancelledError() }` behaviour in case context was
cancelled while awaiting execution slot.

* of `interface{ WorkerPoolClosingError() }` behaviour in case `WorkerPool`
closing started while it was waiting execution slot.
