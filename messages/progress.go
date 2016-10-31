package messages

import (
	"encoding/json"
	"strconv"
	"errors"
)

const (
	TOPIC                  = "coffeepot.progress"
	STAGE_FAILED           = -1;
	STAGE_PENDING          = 1;
	STAGE_BOILING_WATTER   = 2;
	STAGE_BREWING_COFFEE   = 3;
	STAGE_ADDING_ADDITIONS = 4;
	STAGE_FINISHED         = 5;
)

var RelatedTaskIdNotFound = errors.New("Related task not found in message")

type Progress struct {
	Version string `json:"_version"`
	Payload map[string]string `json:"payload"`
	Topic string `json:"topic"`
}

func (p Progress) is(stage int) bool {
	return strconv.Itoa(stage) == p.Payload["stage"]
}

func (p Progress) GetRelatedTaskId() (string, error) {
	if v, ok := p.Payload["taskID"]; ok == true {
		return v, nil
	}

	return "", RelatedTaskIdNotFound
}

func FromSerialized(serialized []byte) (Progress, error) {
	p := Progress{}
	if err := json.Unmarshal(serialized, &p); err == nil {
		return p, nil
	} else {
		return Progress{}, err
	}
}
