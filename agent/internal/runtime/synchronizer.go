package runtime

import (
	"time"

	"github.com/beeosphere/bee/agent/internal/core"
	"github.com/beeosphere/bee/agent/models"
)

const (
	invalidDeploy = "INVALID_DEPLOY_RESPONSE"
	notStored     = "MODEL_NOT_STORED"
	emptyModel    = "EMPTY_MODEL"
)

const modelKey = "model"

type ModelFileMetadata struct {
	Id        string    `json:"id"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *ModelFileMetadata) IsEmpty() bool {
	return m.Id == "" && m.Hash == ""
}

type SyncedModel struct {
	ModelId   string
	ModelHash string
	Model     []byte
	Resources map[string][]byte
}

func NewSyncedModel(modelId, modelHash string) *SyncedModel {
	return &SyncedModel{
		ModelId:   modelId,
		ModelHash: modelHash,
		Resources: make(map[string][]byte),
	}
}

func (ms *SyncedModel) HasModel() bool {
	return len(ms.Model) > 0
}

func (ms *SyncedModel) HasResources() bool {
	return len(ms.Resources) > 0
}

func (ms *SyncedModel) SetModel(model []byte) {
	ms.Model = model
}

func (ms *SyncedModel) AddResource(id string, resource []byte) {
	ms.Resources[id] = resource
}

type Synchronizer interface {
	Startup() error
	SynchronizeFromStore() bool
	SynchronizeFromHive()
	OnSynchronized(func(model *SyncedModel))
}

// REMOTE DEPLOYER

type synchronizer struct {
	session         *core.Session
	log             models.Logger
	commander       *commander
	store           models.Store
	syncedCallbacks []func(model *SyncedModel)
}

func NewSynchronizer(session *core.Session, httpClient *core.HttpClient, commander *commander, store models.Store) Synchronizer {
	return &synchronizer{
		session:         session,
		log:             session.Log,
		commander:       commander,
		store:           store,
		syncedCallbacks: make([]func(model *SyncedModel), 0),
	}
}

func (s *synchronizer) Startup() error {

	s.commander.OnCommandReceived(DEPLOY_SIG, s.deploySignalReceived)
	s.commander.OnCommandReceived(DEPLOY, s.deployReceived)
	s.commander.OnCommandReceived(DEPLOY_ACK, s.deployAckReceived)
	return nil
}

func (s *synchronizer) SynchronizeFromStore() bool {
	// Check if there is a stored model. If so, load it and notify synchronized immediately
	if data, err := s.store.ReadData(modelKey); err == nil && len(data) > 0 {
		meta := s.getStoredModelMetadata()
		if !meta.IsEmpty() {

			model := NewSyncedModel(meta.Id, meta.Hash)
			model.SetModel(data)

			s.log.Infof("Model found in local store (ID: %s, Hash: %s)", meta.Id, core.ShortValue(meta.Hash)) // TODO: Review what to log here...

			s.notifySynchronized(model)
			return true
		}
		s.log.Warnf("Stored model data found but metadata is missing")
	}
	return false
}

func (s *synchronizer) SynchronizeFromHive() {
	// If no stored model, proceed to request deployment from remote deployer
	cmd := &core.DeployRequestCommand{
		AgentId:   s.session.Bee,
		PubKey:    s.session.PublicKey,
		Timestamp: time.Now(),
	}
	// 1. Check store to complete current model info if model metadata exists
	meta := s.getStoredModelMetadata()
	if !meta.IsEmpty() {

		cmd.CurrentModelId = meta.Id
		cmd.CurrentModelHash = meta.Hash
	}
	// 2. Emit deploy request periodically
	s.commander.Emit(DEPLOY_REQ, cmd, nil, 5, 3600) // Every 5 seconds for 1 hour
}

func (s *synchronizer) OnSynchronized(callback func(model *SyncedModel)) {
	s.syncedCallbacks = append(s.syncedCallbacks, callback)
}

func (s *synchronizer) deploySignalReceived(data any, headers models.BusHeaders) {
	s.SynchronizeFromHive()
}

func (s *synchronizer) deployReceived(data any, headers models.BusHeaders) {
	// Stop emitting deploy requests
	s.commander.CancelEmit(DEPLOY_REQ)

	// Prepare deploy response with default values
	r := &core.DeployedCommand{
		AgentId:   s.session.Bee,
		PubKey:    s.session.PublicKey,
		Timestamp: time.Now(),
		Failed:    false,
		Errors:    []string{},
	}

	// Process deploy response
	cmd := core.DirectDeserialize[core.DeployResponseCommand](data)
	if cmd != nil {

		r.ModelId = cmd.ModelId
		r.ModelHash = cmd.ModelHash

		if !cmd.Synced {
			if len(cmd.Model) > 0 {

				err := s.storeModel(cmd.ModelId, cmd.ModelHash, cmd.Model)
				if err == nil {
					s.log.Infof("Model synchronized (ID: %s, Hash: %s)", cmd.ModelId, core.ShortValue(cmd.ModelHash)) // TODO: Review what to log here...

					model := NewSyncedModel(cmd.ModelId, cmd.ModelHash)
					model.SetModel(cmd.Model)
					// TODO: Add resources (via downloader or included in the deploy command)

					s.notifySynchronized(model)

				} else {
					r.AddError(notStored)
				}
			} else {
				r.AddError(emptyModel)
			}
		}
	} else {
		r.AddError(invalidDeploy)
	}

	s.commander.Emit(DEPLOYED, r, nil, 5, 3600) // Every 5 seconds for 1 hour
}

func (s *synchronizer) deployAckReceived(data any, headers models.BusHeaders) {
	// Stop emitting deployed
	s.commander.CancelEmit(DEPLOYED)
}

// func (s *synchronizer) downloadResources(metadata *DeployBinding) (data *AgentData, err error) {
// 	if !metadata.IsEmpty() {

// 		// TODO: Based on metadata.ConfigHash determine if the agent's config has to be updated
// 		// To allow this, the deployer should have access to the agent's config...

// 		data, err = s.downloader.DownloadResources(metadata)
// 		if err == nil {

// 			data.ConfigId = metadata.ConfigId
// 			data.ConfigHash = metadata.ConfigHash
// 		}
// 		return
// 	}
// 	data = &AgentData{} // Empty agent data
// 	return
// }

func (s *synchronizer) notifySynchronized(model *SyncedModel) {
	for _, callback := range s.syncedCallbacks {
		go callback(model)
	}
}

func (s *synchronizer) getStoredModelMetadata() *ModelFileMetadata {
	metadata := ModelFileMetadata{}
	if err := s.store.ReadMeta(modelKey, &metadata); err == nil {
		return &metadata
	}
	return &ModelFileMetadata{
		Id:   "",
		Hash: "",
	}
}

func (s *synchronizer) storeModel(id, hash string, model []byte) error {
	metadata := ModelFileMetadata{
		Id:        id,
		Hash:      hash,
		Timestamp: time.Now(),
	}
	return s.store.Write(modelKey, model, &metadata)
}

// func (s *synchronizer) retrieveModel() ([]byte, error) {
// 	data, err := s.store.ReadData(modelKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// func cast[T any](data any) *T {
// 	var result T
// 	err := core.Deserialize(data.([]byte), &result)
// 	if err != nil {
// 		return nil
// 	}
// 	return &result
// }

// func cast[T any](data any) *T {
// 	casted, ok := data.(*T)
// 	if !ok {
// 		return nil
// 	}
// 	return casted
// }

// // LOCAL DEPLOYER

// type localDeployer struct {
// 	session        *session
// 	configDeployed chan *AgentData
// }

// func newLocalDeployer(session *session) Deployer {
// 	return &localDeployer{
// 		session:        session,
// 		configDeployed: make(chan *AgentData),
// 	}
// }

// func (d *localDeployer) Startup() error {
// 	return nil
// }

// func (d *localDeployer) Deploy(ctx context.Context) {

// 	configPath := d.session.configPath

// 	file, err := os.Open(configPath)
// 	if err != nil {
// 		log.Error("Error opening config file: ", err)
// 		return
// 	}

// 	var config BeeConfiguration
// 	decoder := json.NewDecoder(file)
// 	err = decoder.Decode(&config)
// 	file.Close()
// 	if err != nil {
// 		log.Error("Error decoding config file: ", err)
// 		return
// 	}

// 	hash, err := calculateHash(&config)
// 	if err != nil {
// 		log.Error("Error calculating config hash: ", err)
// 		return
// 	}

// 	agentData := &AgentData{
// 		ConfigId:   configPath,
// 		ConfigHash: hash,
// 		Config:     &config,
// 		Resources:  make(map[string][]byte),
// 	}

// 	d.configDeployed <- agentData

// 	// // If config.watch is true, watch for changes in the configuration file
// 	// if config.Watch {
// 	// 	// TODO: Implement file watching logic
// 	// }
// }

// func (d *localDeployer) Deployed() <-chan *AgentData {
// 	// Return the channel that is whatching for the deployed config in the local file system
// 	return d.configDeployed
// }

// func calculateHash(config *BeeConfiguration) (string, error) {
// 	configBytes, err := json.Marshal(config)
// 	if err != nil {
// 		return "", err
// 	}
// 	hash := sha256.Sum256(configBytes)
// 	return hex.EncodeToString(hash[:]), nil
// }
