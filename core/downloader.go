package core

import (
	"context"
	"fmt"
)

type downloader struct {
	httpClient *HttpClient
}

type agentData struct {
	Config    *BeeConfiguration
	Resources map[string][]byte
}

func newAgentData() *agentData {
	return &agentData{
		Resources: make(map[string][]byte),
	}
}

func (a *agentData) HasConfig() bool {
	return a.Config != nil
}

func (a *agentData) AddResource(id string, resource []byte) {
	a.Resources[id] = resource
}

func newDownloader(httpClient *HttpClient) *downloader {
	return &downloader{
		httpClient: httpClient,
	}
}

func (d *downloader) DownloadResources(info *DeployBinding) (*agentData, error) {
	data := newAgentData()

	// TODO: Implement retry policy...?

	var config BeeConfiguration
	if err := d.httpClient.Get(context.TODO(), fmt.Sprintf("api/configs/%s/snapshot", info.ConfigId), &config); err != nil {
		fmt.Println("Error downloading config: ", err)
		return nil, err
	}
	data.Config = &config

	// Download resources if necessary
	// data.AddResource("res1", resource1)
	// data.AddResource("res2", resource2)...

	return data, nil
}
