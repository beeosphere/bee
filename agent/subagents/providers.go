package subagents

import (
	"fmt"

	"github.com/beeosphere/bee/agent/models"
	"github.com/tidwall/gjson"
)

// PathSubmodelProvider returns a SubmodelProvider that extracts a JSON sub-object
// from the model's Data field at the given gjson path, returning the raw JSON bytes.
func PathSubmodelProvider(path string) SubmodelResolver {
	return func(name string, model models.Model) ([]byte, error) {
		json := model.Data
		result := gjson.GetBytes(json, path)
		if !result.Exists() {
			return nil, fmt.Errorf("path '%q' not found in model data", path)
		}
		var raw []byte
		if result.Index > 0 {
			raw = json[result.Index : result.Index+len(result.Raw)]
		} else {
			raw = []byte(result.Raw)
		}
		return raw, nil
	}
}
