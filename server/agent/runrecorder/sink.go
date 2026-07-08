// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

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
