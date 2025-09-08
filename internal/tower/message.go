package tower

import "github.com/nanthakumaran-s/golancer/internal/config"

type ConfigUpdated struct {
	NewConfig *config.Config
}
type Shutdown struct{}
