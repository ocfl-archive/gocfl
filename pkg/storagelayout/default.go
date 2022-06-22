package storagelayout

import (
	"errors"
	"fmt"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/ocfl"
)

type Default struct{}

func (sl *Default) Name() string { return "default" }
func (sl *Default) ID2Path(id string) (string, error) {
	if len(id) > MAX_DIR_LEN {
		return "", errors.New(fmt.Sprintf("%s to long (max. %v)", id, MAX_DIR_LEN))
	}
	return ocfl.FixFilename(id), nil
}
