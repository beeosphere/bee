package runtime

import (
	"github.com/beeosphere/bee/agent/models"
)

type Processor interface {
	Process(model *models.Model) error
	Dispose() error
}
