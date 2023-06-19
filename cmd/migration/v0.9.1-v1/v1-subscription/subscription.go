package v1subscription

import "encoding/json"

type Subscription struct {
	ID       string `json:"id"`
	Event    string `json:"event"`
	Endpoint string `json:"endpoint"`
	Secret   string `json:"secret"`
}

func (s *Subscription) Serialize() []byte {
	b, _ := json.Marshal(*s)
	return b
}
