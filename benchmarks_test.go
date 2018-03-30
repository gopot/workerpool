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
	"testing"

	wp "github.com/gopot/workerpool"
)

func Benchmark_DoScheduleEmptyJob(b *testing.B) {

	testCases := []struct {
		Alias       string
		Parallelism int
		WorkerPool  *wp.WorkerPool
	}{
		{
			Alias:       `Sequental on default wp`,
			Parallelism: 1,
			WorkerPool:  &wp.WorkerPool{},
		},
		{
			Alias:       `4 on default wp`,
			Parallelism: 4,
			WorkerPool:  &wp.WorkerPool{},
		},
		{
			Alias:       `8 on default wp`,
			Parallelism: 8,
			WorkerPool:  &wp.WorkerPool{},
		},
		{
			Alias:       `4 on wp with 4 slots`,
			Parallelism: 4,
			WorkerPool:  wp.NewWorkerPool(4),
		},
		{
			Alias:       `8 on wp with 8 slots`,
			Parallelism: 8,
			WorkerPool:  wp.NewWorkerPool(8),
		},
		{
			Alias:       `16 on wp with 8 slots`,
			Parallelism: 16,
			WorkerPool:  wp.NewWorkerPool(8),
		},
		{
			Alias:       `16 on wp with 16 slots`,
			Parallelism: 16,
			WorkerPool:  wp.NewWorkerPool(16),
		},
	}

	for _, tCase := range testCases {

		b.SetParallelism(tCase.Parallelism)

		benchFn := func(pb *testing.PB) {
			for pb.Next() {
				tCase.WorkerPool.Do(context.Background(), func() {})
			}
		}

		testFn := func(b *testing.B) {
			b.RunParallel(benchFn)
		}

		b.Run(tCase.Alias, testFn)
	}

}
