/*
 * UpdateHub
 * Copyright (C) 2017
 * O.S. Systems Sofware LTDA: contato@ossystems.com.br
 *
 * SPDX-License-Identifier:     GPL-2.0
 */

package updatehub

import (
	"sync"

	"github.com/UpdateHub/updatehub/metadata"
)

// DownloadingState is the State interface implementation for the UpdateHubStateDownloading
type DownloadingState struct {
	BaseState
	CancellableState
	ReportableState
	ProgressTracker

	updateMetadata *metadata.UpdateMetadata
}

// ID returns the state id
func (state *DownloadingState) ID() UpdateHubState {
	return state.id
}

// Cancel cancels a state if it is cancellable
func (state *DownloadingState) Cancel(ok bool, nextState State) bool {
	state.CancellableState.Cancel(ok, nextState)
	return ok
}

// UpdateMetadata is the ReportableState interface implementation
func (state *DownloadingState) UpdateMetadata() *metadata.UpdateMetadata {
	return state.updateMetadata
}

// Handle for DownloadingState starts the objects downloads. It goes
// to the installing state if successfull. It goes back to the error
// state otherwise.
func (state *DownloadingState) Handle(uh *UpdateHub) (State, bool) {
	var err error
	var nextState State

	nextState = state

	progressChan := make(chan int, 10)

	m := sync.Mutex{}
	m.Lock()

	go func() {
		m.Lock()
		defer m.Unlock()

		err = uh.Controller.DownloadUpdate(state.updateMetadata, state.cancel, progressChan)
		close(progressChan)
	}()

	m.Unlock()
	for p := range progressChan {
		state.ProgressTracker.SetProgress(p)
	}

	if err != nil {
		nextState = NewErrorState(state.updateMetadata, NewTransientError(err))
	} else {
		nextState = NewDownloadedState(state.updateMetadata)
	}

	// state cancelled
	if state.NextState() != nil {
		return state.NextState(), true
	}

	return nextState, false
}

// ToMap is for the State interface implementation
func (state *DownloadingState) ToMap() map[string]interface{} {
	m := state.BaseState.ToMap()
	m["progress"] = state.ProgressTracker.GetProgress()
	return m
}

// NewDownloadingState creates a new DownloadingState from a metadata.UpdateMetadata
func NewDownloadingState(updateMetadata *metadata.UpdateMetadata, pti ProgressTracker) *DownloadingState {
	state := &DownloadingState{
		BaseState:       BaseState{id: UpdateHubStateDownloading},
		updateMetadata:  updateMetadata,
		ProgressTracker: pti,
	}

	return state
}