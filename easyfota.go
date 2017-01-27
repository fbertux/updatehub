package main

import (
	"errors"
	"fmt"
	"path"
	"time"

	"bitbucket.org/ossystems/agent/client"
	"bitbucket.org/ossystems/agent/metadata"
)

type EasyFota struct {
	Controller

	state        State
	pollInterval int
	timeStep     time.Duration
	api          *client.ApiClient
	updater      client.Updater
}

type Controller interface {
	CheckUpdate() (*metadata.Metadata, int)
	FetchUpdate(*metadata.Metadata) error
}

func (fota *EasyFota) CheckUpdate() (*metadata.Metadata, int) {
	updateMetadata, extraPoll, err := fota.updater.CheckUpdate(fota.api.Request())
	if err != nil {
		return nil, 0
	}

	return updateMetadata.(*metadata.Metadata), extraPoll
}

func (fota *EasyFota) FetchUpdate(updateMetadata *metadata.Metadata) error {
	// For now, we installs the first object
	// FIXME: What object I should to install?
	obj := updateMetadata.Objects[0][0]

	if obj == nil {
		return errors.New("object not found")
	}

	// FIXME: read product uid from firmaware metadata
	productUID := "1"

	packageUID, err := updateMetadata.Checksum()
	if err != nil {
		return err
	}

	objectUID := obj.GetObjectData().Sha256sum

	uri := "/"
	uri = path.Join(uri, productUID)
	uri = path.Join(uri, packageUID)
	uri = path.Join(uri, objectUID)

	body, contentLength, err := fota.updater.FetchUpdate(fota.api.Request(), uri)
	if err != nil {
		return err
	}

	fmt.Println(body)
	fmt.Println(contentLength)

	return nil
}

func (fota *EasyFota) MainLoop() {
	for {
		fmt.Println("Handling state:", StateToString(fota.state.Id()))

		state, cancelled := fota.state.Handle(fota)

		if state.Id() == EasyFotaStateError {
			if es, ok := state.(*ErrorState); ok {
				// FIXME: log error
				fmt.Println(es.cause)
			}
		}

		if cancelled {
			fmt.Println("State cancelled")
		}

		fota.state = state
	}
}