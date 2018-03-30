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

/*
	The file contains defenitions of errors by behaviour.
	All error types are private in order to avoid direct error usage.

	Errors should be checked for corresponding interfaces.
*/

package workerpool

// Represents basic error
type basicError struct {
	msg string
}

// Implements golang `error` interface
func (be *basicError) Error() string {
	return be.msg
}

// Private factory for `basicError`
func newBasicError(msg string) basicError {
	return basicError{msg: msg}
}

// Represents error returned in case of awaiting execution slot was cancelled
type awaitingCanelledError struct {
	basicError
}

// Defines interface for `AwaitingCancelledError`
func (e *awaitingCanelledError) AwaitingCancelledError() {}

// Private factory for `awaitingCanelledError`
func newAwaitingCancelledError(msg string) error {

	return &awaitingCanelledError{basicError: newBasicError(msg)}
}

// Represents error returned in case of `WorkerPool` is closing during request or awaiting execution slot
type workerPoolClosingError struct {
	basicError
}

// Defines interface for `WorkerPoolClosingError`
func (e *workerPoolClosingError) WorkerPoolClosingError() {}

// Private factory for `workerPoolClosingError`
func newWorkerPoolClosingError(msg string) error {
	return &workerPoolClosingError{basicError: newBasicError(msg)}
}
