package sparta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mweagle/Sparta/system"

	gocf "github.com/mweagle/go-cloudformation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const mainSysInfoSample = `
package main

import (
	"fmt"

	"github.com/zcalusic/sysinfo"
)

func main() {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	fmt.Printf("%v", si)
}
`

// describeInfoValue is a utility function that accepts
// some type of dynamic gocf value and transforms it into
// something that is `describe` output compatible
func describeInfoValue(dynamicValue interface{}) string {
	switch typedArn := dynamicValue.(type) {
	case string:
		return typedArn
	case gocf.Stringable:
		data, dataErr := json.Marshal(typedArn)
		if dataErr != nil {
			data = []byte(fmt.Sprintf("%v", typedArn))
		}
		return string(data)
	default:
		panic(fmt.Sprintf("Unsupported dynamic value type for `describe`: %+v", typedArn))
	}
}

// Create a stable temporary filename in the current working
// directory
func temporaryFile(name string) (*os.File, error) {
	workingDir, err := os.Getwd()
	if nil != err {
		return nil, err
	}

	// Use a stable temporary name
	temporaryPath := filepath.Join(workingDir, ScratchDirectory, name)
	buildDir := filepath.Dir(temporaryPath)
	mkdirErr := os.MkdirAll(buildDir, os.ModePerm)
	if nil != mkdirErr {
		return nil, mkdirErr
	}

	tmpFile, err := os.Create(temporaryPath)
	if err != nil {
		return nil, errors.New("Failed to create temporary file: " + err.Error())
	}

	return tmpFile, nil
}

// relativePath returns the relative path of logPath if it's relative to the current
// workint directory
func relativePath(logPath string) string {
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil {
		relPath := strings.TrimPrefix(logPath, cwd)
		if relPath != logPath {
			logPath = fmt.Sprintf(".%s", relPath)
		}
	}
	return logPath
}

func buildSysInfoSample(logger *logrus.Logger) error {
	workingDir, workingDirErr := os.Getwd()
	if workingDirErr != nil {
		return errors.Wrapf(workingDirErr, "Failed to determine working directory")
	}
	temporaryDir := filepath.Join(workingDir,
		ScratchDirectory,
		"buildTest")

	mkdirErr := os.MkdirAll(temporaryDir, os.ModePerm)
	if mkdirErr != nil {
		return errors.Wrapf(mkdirErr, "Failed to create tempdir for build")
	}
	defer os.RemoveAll(temporaryDir)
	mainOutputSource := filepath.Join(temporaryDir, "main.go")
	writeErr := ioutil.WriteFile(mainOutputSource,
		[]byte(mainSysInfoSample),
		0644)

	if writeErr != nil {
		return errors.Wrapf(writeErr, "Failed to create main file")
	}
	// Build it...
	buildArgs := []string{
		"build",
	}
	if logger.Level == logrus.DebugLevel {
		buildArgs = append(buildArgs, "-v")
	}
	buildArgs = append(buildArgs, "main.go")
	cmd := exec.Command("go", buildArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
	cmd.Dir = temporaryDir
	logger.Debug("Verifying sysinfo package")
	cmdError := system.RunOSCommand(cmd, logger)
	if cmdError != nil {
		return errors.Wrapf(cmdError, "Failed")
	}
	return nil
}

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
