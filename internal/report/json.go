package report

import (
	"encoding/json"

	"github.com/DRMCGT/Squawk/internal/model"
)

// JSON renders a session as an indented JSON document.
func JSON(sess *model.Session) ([]byte, error) {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
