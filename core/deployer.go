package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beeosphere/bee/core/topics"
)

type deployer struct {
	session *session
	// httpClient       *HttpClient
	downloader       *downloader
	publisher        *publisher
	deploySubscriber *subscriber
	deployMetadata   chan *DeployBinding
	ConfigDeployed   chan *agentData
}

func newDeployer(session *session, httpClient *HttpClient) *deployer {
	return &deployer{
		session:        session,
		downloader:     newDownloader(httpClient),
		deployMetadata: make(chan *DeployBinding),
		ConfigDeployed: make(chan *agentData),
	}
}

func (d *deployer) Startup() error {
	var err error

	d.publisher = newCommandPublisher()

	if d.deploySubscriber, err = newCommandSubscriber(topics.DeployReceived(d.session.bee), d); err != nil {
		return err
	}
	return nil
}

// messageHandler implementation: Start
func (d *deployer) processMessage(msg *DataMessage) error {
	return nil
}
func (d *deployer) executeCommand(cmd *CommandMessage) error {
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

func (d *deployer) Deploy(ctx context.Context) {

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
					d.ConfigDeployed <- agentData
				}
				if err != nil {
					// TODO: log error
					fmt.Println("Error downloading resources: ", err)
				}
			}
		}
	}()
}

func (d *deployer) deployRequest(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	var agentData *agentData
	var err error

	// Publishes a deploy request
	d.publisher.Publish(topics.DeployRequest(), &DeployRequest{AgentId: d.session.bee, Timestamp: time.Now()})

	// var metadata *DeployBinding
	for agentData == nil {
		select {
		case metadata := <-d.deployMetadata:
			ticker.Stop() // Stop the ticker

			agentData, err = d.downloadResources(metadata)
			if agentData != nil {
				d.ConfigDeployed <- agentData
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
			d.publisher.Publish(topics.DeployRequest(), &DeployRequest{AgentId: d.session.bee, Timestamp: time.Now()})

		case <-ctx.Done():
			return
		}
	}
}

func (d *deployer) downloadResources(metadata *DeployBinding) (data *agentData, err error) {
	if !metadata.IsEmpty() {

		// TODO: Based on metadata.ConfigHash determine if the agent's config has to be updated
		// To allow this, the deployer should have access to the agent's config...

		data, err = d.downloader.DownloadResources(metadata)
		return
	}
	data = &agentData{} // Empty agent data
	return
}
