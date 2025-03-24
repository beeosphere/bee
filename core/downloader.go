package core

import (
	"context"
	"fmt"
)

type Downloader interface {
	DownloadResources(info *DeployBinding) (*AgentData, error)
}

type remoteDownloader struct {
	httpClient *HttpClient
}

type AgentData struct {
	ConfigId   string
	ConfigHash string
	Config     *BeeConfiguration
	Resources  map[string][]byte
}

func newAgentData() *AgentData {
	return &AgentData{
		Resources: make(map[string][]byte),
	}
}

func (a *AgentData) HasConfig() bool {
	return a.Config != nil
}

func (a *AgentData) AddResource(id string, resource []byte) {
	fmt.Printf("Resource: %s %d\n", id, len(resource))
	a.Resources[id] = resource
}

func newRemoteDownloader(httpClient *HttpClient) Downloader {
	return &remoteDownloader{
		httpClient: httpClient,
	}
}

func (d *remoteDownloader) DownloadResources(info *DeployBinding) (*AgentData, error) {
	data := newAgentData()

	// TODO: Implement retry policy...?

	var config BeeConfiguration
	if err := d.httpClient.Get(context.TODO(), fmt.Sprintf("api/configs/%s/snapshot", info.ConfigId), &config); err != nil {
		fmt.Println("Error downloading config: ", err)
		return nil, err
	}
	data.Config = &config

	// Download resources if necessary
	for _, resource := range info.Resources {
		bytes, err := d.httpClient.GetBytes(context.TODO(), fmt.Sprintf("api/resources/%s/download", resource.Id))
		if err != nil {
			fmt.Println("Error downloading resource: ", err)
			return nil, err
		}
		data.AddResource(resource.Id, bytes)
	}

	return data, nil
}
