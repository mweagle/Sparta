package sparta

import (
	"sync"
)

// workResult is the result from a worker task
type workResult interface {
	Result() interface{}
	Error() error
}

// taskResult is a convenience type for a task poll return value
type taskResult struct {
	result interface{}
	err    error
}

func (tr *taskResult) Result() interface{} {
	return tr.result
}
func (tr *taskResult) Error() error {
	return tr.err
}

func newTaskResult(taskValue interface{}, err error) workResult {
	return &taskResult{
		result: taskValue,
		err:    err,
	}
}

type taskFunc func() workResult

// workTask encapsulates a work item that should go in a work pool.
type workTask struct {
	// Result is the result of the work action
	Result workResult
	task   taskFunc
}

// Run runs a Task and does appropriate accounting via a given sync.WorkGroup.
func (t *workTask) Run(wg *sync.WaitGroup) {
	t.Result = t.task()
	wg.Done()
}

// newWorkTask initializes a new task based on a given work function.
func newWorkTask(f taskFunc) *workTask {
	return &workTask{task: f}
}

// workerPool is a worker group that runs a number of tasks at a configured
// concurrency.
type workerPool struct {
	Tasks []*workTask

	concurrency int
	tasksChan   chan *workTask
	wg          sync.WaitGroup
}

// newWorkerPool initializes a new pool with the given tasks and at the given
// concurrency.
func newWorkerPool(tasks []*workTask, concurrency int) *workerPool {
	return &workerPool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *workTask),
	}
}

// HasErrors indicates whether there were any errors from tasks run. Its result
// is only meaningful after Run has been called.
func (p *workerPool) workResults() ([]interface{}, []error) {
	result := []interface{}{}
	errors := []error{}

	for _, eachResult := range p.Tasks {
		if eachResult.Result.Error() != nil {
			errors = append(errors, eachResult.Result.Error())
		} else {
			result = append(result, eachResult.Result.Result())
		}
	}
	return result, errors
}

// Run runs all work within the pool and blocks until it's finished.
func (p *workerPool) Run() ([]interface{}, []error) {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}

	// all workers return
	close(p.tasksChan)

	p.wg.Wait()
	return p.workResults()
}

// The work loop for any single goroutine.
func (p *workerPool) work() {
	for task := range p.tasksChan {
		task.Run(&p.wg)
	}
}
