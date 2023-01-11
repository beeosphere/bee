package core

type ErrorData struct {
	Error string `json:"error,omitempty"`
}

type SyncConfigInfo struct {
	Uri     string `json:"uri"`
	Version int    `json:"version"`
}
type SyncResourceInfo struct {
	Id      string `json:"id"`
	Uri     string `json:"uri"`
	Hash    string `json:"hash"`
	Type    string `json:"type"`
	Version int    `json:"version"`
}

type SyncData struct {
	ErrorData

	Config    *SyncConfigInfo     `json:"config"`
	Resources []*SyncResourceInfo `json:"resources"`
}
