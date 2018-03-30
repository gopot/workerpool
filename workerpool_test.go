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

package workerpool_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	wp "github.com/gopot/workerpool"
)

func Test_WorkerPoolInterface(t *testing.T) {

	type workerPool interface {
		Do(context.Context, func()) error

		Close()
	}

	var _ workerPool = &wp.WorkerPool{}

}

func Test_WorkerPoolOneJob(t *testing.T) {

	t.Parallel()

	testCases := []struct {
		TestAlias string
		// Having
		WorkerPool *wp.WorkerPool
		// When
		AlreadyRunningJobs int
		Context            context.Context
		// Then
		JobExecuted                  bool
		ReturnNilError               bool
		ReturnAwaitingCancelledError bool
	}{
		{
			TestAlias:          `One slot zero preload one job`,
			WorkerPool:         wp.NewWorkerPool(1),
			AlreadyRunningJobs: 0,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
		{
			TestAlias:          `Ten slots zero preload one job`,
			WorkerPool:         wp.NewWorkerPool(10),
			AlreadyRunningJobs: 0,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
		{
			TestAlias:          `Ten slots nine preload one job`,
			WorkerPool:         wp.NewWorkerPool(10),
			AlreadyRunningJobs: 9,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
		{
			TestAlias:          `100 slots one free one job`,
			WorkerPool:         wp.NewWorkerPool(100),
			AlreadyRunningJobs: 99,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
		{
			TestAlias:          `Ten slots none free one job cancelled context`,
			WorkerPool:         wp.NewWorkerPool(10),
			AlreadyRunningJobs: 10,
			Context: func() context.Context {
				ctx, cancelFn := context.WithCancel(context.Background())
				cancelFn()
				return ctx
			}(),
			JobExecuted:                  false,
			ReturnAwaitingCancelledError: true,
		},
	}

	for _, tCase := range testCases {

		testFn := func(t *testing.T) {

			const jobFuncMockName = "jobFunc"

			// Preload WorkerPool with running jobs
			runningWorkersWaitFor := make(chan struct{})
			defer close(runningWorkersWaitFor)

			for i := 0; i < tCase.AlreadyRunningJobs; i++ {
				running := make(chan struct{})

				job := func() { close(running); <-runningWorkersWaitFor }
				tCase.WorkerPool.Do(context.Background(), job)
				// Ensure job started
				<-running
			}

			jobMock := &mock.Mock{}

			// Set up mock call expectations
			if tCase.JobExecuted {
				jobMock.On(jobFuncMockName).Once()
			}

			mockedJob := func() { jobMock.MethodCalled(jobFuncMockName) }

			actualError := tCase.WorkerPool.Do(tCase.Context, mockedJob)

			if (actualError == nil) != tCase.ReturnNilError {
				t.Errorf("Failed with unexpected error %+v", actualError)
			}

			if actualError != nil {

				type AwaitingCancelledError interface {
					AwaitingCancelledError()
				}

				if _, ok := actualError.(AwaitingCancelledError); ok != tCase.ReturnAwaitingCancelledError {
					t.Errorf("Returned error %#v does not implement expectd interface AwaitingCancelledError", actualError)
				}
			}

			jobMock.AssertExpectations(t)

		}

		t.Run(tCase.TestAlias, testFn)
	}

}

func Test_WorkerPoolAllJobsExecuted(t *testing.T) {

	t.Parallel()

	testCases := []struct {
		Alias        string
		WorkerPool   *wp.WorkerPool
		NumberOfJobs int
	}{
		{
			Alias:        `WP with 1 slot executes 1K jobs`,
			WorkerPool:   wp.NewWorkerPool(1),
			NumberOfJobs: 1000,
		},
		{
			Alias:        `WP with 10 slots executes 1K jobs`,
			WorkerPool:   wp.NewWorkerPool(10),
			NumberOfJobs: 1000,
		},
		{
			Alias:        `WP with 100 slots executes 1K jobs`,
			WorkerPool:   wp.NewWorkerPool(100),
			NumberOfJobs: 1000,
		},
		{
			Alias:        `WP with 1K slots executes 5K jobs`,
			WorkerPool:   wp.NewWorkerPool(1000),
			NumberOfJobs: 5000,
		},
		{
			Alias:        `WP with 100 slots executes 10 jobs`,
			WorkerPool:   wp.NewWorkerPool(100),
			NumberOfJobs: 10,
		},
	}

	for _, tCase := range testCases {

		testFn := func(t *testing.T) {

			const jobFuncMockName = "jobFunc"

			jobMock := &mock.Mock{}
			jobMock.On(jobFuncMockName)

			wg := &sync.WaitGroup{}
			for i := 0; i < tCase.NumberOfJobs; i++ {
				wg.Add(1)
				go func() {

					tCase.WorkerPool.Do(context.Background(), func() { defer wg.Done(); jobMock.MethodCalled(jobFuncMockName) })
				}()
			}

			wg.Wait()

			jobMock.AssertNumberOfCalls(t, jobFuncMockName, tCase.NumberOfJobs)

		}

		t.Run(tCase.Alias, testFn)
	}

}

func Test_WorkerPoolClosing(t *testing.T) {

	t.Parallel()

	testCases := []struct {
		TestAlias string
		// Having
		WorkerPool *wp.WorkerPool
		// When
		AlreadyRunningJobs int
		Context            context.Context
		// Then
		JobExecuted                  bool
		ReturnNilError               bool
		ReturnAwaitingCancelledError bool
		ReturnWorkerPoolClosingError bool
	}{
		{
			TestAlias:          `One slot zero preload one job`,
			WorkerPool:         wp.NewWorkerPool(1),
			AlreadyRunningJobs: 0,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
		{
			TestAlias:          `Ten slots none free one job cancelled context`,
			WorkerPool:         wp.NewWorkerPool(10),
			AlreadyRunningJobs: 10,
			Context: func() context.Context {
				ctx, cancelFn := context.WithCancel(context.Background())
				cancelFn()
				return ctx
			}(),
			JobExecuted:                  false,
			ReturnAwaitingCancelledError: true,
		},
		{
			TestAlias:                    `Ten slots none free one normal job`,
			WorkerPool:                   wp.NewWorkerPool(10),
			AlreadyRunningJobs:           10,
			Context:                      context.Background(),
			JobExecuted:                  false,
			ReturnWorkerPoolClosingError: true,
		},
		{
			TestAlias:          `Incorrect initialized wp and one normal job`,
			WorkerPool:         &wp.WorkerPool{},
			AlreadyRunningJobs: 0,
			Context:            context.Background(),
			JobExecuted:        true,
			ReturnNilError:     true,
		},
	}

	for _, tCase := range testCases {

		tCaseTestAlias := tCase.TestAlias
		tCaseWorkerPool := tCase.WorkerPool
		tCaseAlreadyRunningJobs := tCase.AlreadyRunningJobs
		tCaseContext := tCase.Context
		tCaseJobExecuted := tCase.JobExecuted
		tCaseReturnNilError := tCase.ReturnNilError
		tCaseReturnAwaitingCancelledError := tCase.ReturnAwaitingCancelledError
		tCaseReturnWorkerPoolClosingError := tCase.ReturnWorkerPoolClosingError

		testFn := func(t *testing.T) {

			const jobFuncMockName = "jobFunc"

			// Preload WorkerPool with running jobs
			runningWorkersWaitFor := make(chan struct{})
			runningWorkerDone := make(chan struct{})

			for i := 0; i < tCaseAlreadyRunningJobs; i++ {
				running := make(chan struct{})

				job := func() {
					defer func() { runningWorkerDone <- struct{}{} }()
					close(running)
					<-runningWorkersWaitFor
				}

				tCaseWorkerPool.Do(context.Background(), job)

				// Ensure job has started
				<-running
			}

			jobMock := &mock.Mock{}

			// Set up mock call expectations
			if tCaseJobExecuted {
				jobMock.On(jobFuncMockName).Once()
			}

			mockedJob := func() { jobMock.MethodCalled(jobFuncMockName) }

			wpDoStarting := make(chan struct{})
			wpDoCompleted := make(chan struct{})

			// Test workerpool.Do
			go func() {

				defer close(wpDoCompleted)

				// Notify that Do starting
				close(wpDoStarting)

				actualError := tCaseWorkerPool.Do(tCaseContext, mockedJob)

				if (actualError == nil) != tCaseReturnNilError {
					t.Errorf("Failed with unexpected error %+v", actualError)
				}

				if actualError != nil {

					type AwaitingCancelledError interface {
						AwaitingCancelledError()
					}

					type WorkerPoolClosingError interface {
						WorkerPoolClosingError()
					}

					if _, ok := actualError.(AwaitingCancelledError); ok != tCaseReturnAwaitingCancelledError {
						t.Errorf("Returned error %#v does not implement expectd interface AwaitingCancelledError", actualError)
					}

					if _, ok := actualError.(WorkerPoolClosingError); ok != tCaseReturnWorkerPoolClosingError {
						t.Errorf("Returned error %#v does not implement expectd interface WorkerPoolClosingError", actualError)
					}
				}

				jobMock.AssertExpectations(t)

			}()

			// Do not proceed unless Do goroutine is scheduled
			<-wpDoStarting

			wpCloseStarting := make(chan struct{})
			wpCloseReturned := make(chan struct{})

			// Close worker pool
			go func() {

				// Notify when wp.Close returned. It should not happen before all jobs are done.
				defer close(wpCloseReturned)

				// Notify that goroutine is scheduled
				close(wpCloseStarting)

				// Close worker pool
				tCaseWorkerPool.Close()

			}()

			allRunningJobsAreDone := make(chan struct{})

			// Test that worker pool .Close does not return unless all jobs are done
			go func() {
				defer close(allRunningJobsAreDone)

				// Collect signals from all preloaded jobs
				for i := 0; i < tCaseAlreadyRunningJobs; i++ {

					select {
					case <-runningWorkerDone:
					case <-wpCloseReturned:
						// Prevent infinite loop
						wpCloseReturned = nil

						// If wpIsClosed inside this loop == some jobs are not done yet
						t.Error("Worker pool closed before all jobs were done")
					}
				}
			}()

			// Wait untill closing wp goroutine is scheduled.
			<-wpCloseStarting

			// Wait for Do goroutine is complete
			<-wpDoCompleted

			t.Log("Releasing workers")
			close(runningWorkersWaitFor)

			// Wait for all preload job signals arecollected
			<-allRunningJobsAreDone

		}

		t.Run(tCaseTestAlias, testFn)
	}

}
