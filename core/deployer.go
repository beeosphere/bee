package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
	// log "github.com/sirupsen/logrus"
)

type Deployer interface {
	Startup() error
	Deploy(ctx context.Context)
	Deployed() <-chan *AgentData
}

// REMOTE DEPLOYER

type remoteDeployer struct {
	session *session
	// httpClient       *HttpClient
	downloader       Downloader
	publisher        Publisher
	deploySubscriber Subscriber
	deployMetadata   chan *DeployBinding
	configDeployed   chan *AgentData
}

func newRemoteDeployer(session *session, httpClient *HttpClient) Deployer {
	return &remoteDeployer{
		session:        session,
		downloader:     newRemoteDownloader(httpClient),
		deployMetadata: make(chan *DeployBinding),
		configDeployed: make(chan *AgentData),
	}
}

func (d *remoteDeployer) Startup() error {
	d.publisher = newCommandPublisher()

	d.deploySubscriber = newCommandSubscriber(d.session.bee, "DEPLOY", d)
	return d.deploySubscriber.Subscribe()
}

func (d *remoteDeployer) Deployed() <-chan *AgentData {
	return d.configDeployed
}

// messageHandler implementation: Start
func (d *remoteDeployer) processMessage(msg *DataMessage) error {
	return nil
}
func (d *remoteDeployer) executeCommand(cmd *CommandMessage) error {
	switch cmd.Cmd {
	case Deploy:

		var metadata DeployBinding
		if err := json.Unmarshal(cmd.Data, &metadata); err == nil {
			d.deployMetadata <- &metadata
		}

		return nil
	}
	return nil
}

// messageHandler implementation: End

func (d *remoteDeployer) Deploy(ctx context.Context) {

	// TODO: Implement deploy request to be called from the controller. Implement goroutine inside the method...
	go func() {
		d.deployRequest(ctx) // It should be synchronous. Not called as goroutine

		for {
			select {
			case <-ctx.Done():
				return

			case metadata := <-d.deployMetadata:
				agentData, err := d.downloadResources(metadata)
				if agentData != nil {
					d.configDeployed <- agentData
				}
				if err != nil {
					// TODO: log error
					fmt.Println("Error downloading resources: ", err)
				}
			}
		}
	}()
}

func (d *remoteDeployer) deployRequest(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	var agentData *AgentData
	var err error

	// Publishes a deploy request
	d.publisher.Publish(cmd_DEPLOY_REQ, &DeployRequest{AgentId: d.session.bee, Timestamp: time.Now()})

	// var metadata *DeployBinding
	for agentData == nil {
		select {
		case metadata := <-d.deployMetadata:
			ticker.Stop() // Stop the ticker

			agentData, err = d.downloadResources(metadata)
			if agentData != nil {
				d.configDeployed <- agentData
				return
			}
			if err != nil {
				// TODO: log error
				fmt.Println("Error downloading resources: ", err)
			}
			// Reinitialize the ticker
			ticker = time.NewTicker(5 * time.Second)

		case <-ticker.C:
			// Publishes a deploy request again
			d.publisher.Publish(cmd_DEPLOY_REQ, &DeployRequest{AgentId: d.session.bee, Timestamp: time.Now()})

		case <-ctx.Done():
			return
		}
	}
}

func (d *remoteDeployer) downloadResources(metadata *DeployBinding) (data *AgentData, err error) {
	if !metadata.IsEmpty() {

		// TODO: Based on metadata.ConfigHash determine if the agent's config has to be updated
		// To allow this, the deployer should have access to the agent's config...

		data, err = d.downloader.DownloadResources(metadata)
		if err == nil {

			data.ConfigId = metadata.ConfigId
			data.ConfigHash = metadata.ConfigHash
		}
		return
	}
	data = &AgentData{} // Empty agent data
	return
}

// LOCAL DEPLOYER

type localDeployer struct {
	session        *session
	configDeployed chan *AgentData
}

func newLocalDeployer(session *session) Deployer {
	return &localDeployer{
		session:        session,
		configDeployed: make(chan *AgentData),
	}
}

func (d *localDeployer) Startup() error {
	return nil
}

func (d *localDeployer) Deploy(ctx context.Context) {

	configPath := d.session.configPath

	file, err := os.Open(configPath)
	if err != nil {
		log.Error("Error opening config file: ", err)
		return
	}

	var config BeeConfiguration
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	file.Close()
	if err != nil {
		log.Error("Error decoding config file: ", err)
		return
	}

	hash, err := calculateHash(&config)
	if err != nil {
		log.Error("Error calculating config hash: ", err)
		return
	}

	agentData := &AgentData{
		ConfigId:   configPath,
		ConfigHash: hash,
		Config:     &config,
		Resources:  make(map[string][]byte),
	}

	d.configDeployed <- agentData

	// // If config.watch is true, watch for changes in the configuration file
	// if config.Watch {
	// 	// TODO: Implement file watching logic
	// }
}

func (d *localDeployer) Deployed() <-chan *AgentData {
	// Return the channel that is whatching for the deployed config in the local file system
	return d.configDeployed
}

func calculateHash(config *BeeConfiguration) (string, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(configBytes)
	return hex.EncodeToString(hash[:]), nil
}
