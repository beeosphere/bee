package core

// import "context"

// type synchronizer struct {
// 	session    *session
// 	httpClient *HttpClient
// 	publisher  *publisher
// 	syncData   *SyncData
// }

// func newSynchronizer(session *session, httpClient *HttpClient) *synchronizer {
// 	return &synchronizer{
// 		session:    session,
// 		httpClient: httpClient,
// 	}
// }

// func (s *synchronizer) Startup() error {
// 	// Instantiates publisher
// 	s.publisher = newCommandPublisher() // TODO: Redundant. It is executed from controller's startup method.

// 	// TODO: Initialize resource manager (downloader)

// 	return nil
// }

// func (s *synchronizer) Synchronize() (*BeeConfiguration, error) {
// 	// Request sync data
// 	var syncData SyncData
// 	if err := s.publisher.Request(SyncRequestTopic(s.session.hub, s.session.bee), struct{}{}, &syncData); err != nil {
// 		return nil, err
// 	}

// 	// TODO: Request only new version resources

// 	// Request bee configuration
// 	var config BeeConfiguration
// 	if err := s.httpClient.Get(context.TODO(), syncData.Config.Uri, &config); err != nil {
// 		return nil, err
// 	}

// 	s.syncData = &syncData
// 	return &config, nil
// }
