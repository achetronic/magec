// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runrecorder

// Sink persists finalized run records. Implementations must be safe for
// concurrent use since AfterRunCallback fires from whatever goroutine ADK
// runs the invocation on, and MarkRunError can race with a flush.
type Sink interface {
	// SaveRun persists a complete run record (run metadata plus all events).
	SaveRun(record RunRecord) error

	// SetRunError updates the status/error of an already-persisted run.
	// Used when MarkRunError is called after the run has already flushed.
	SetRunError(runID, message string) error
}
