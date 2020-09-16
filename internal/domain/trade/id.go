package trade

type id struct {
	request  string
	accept   string
	complete string
}

func (i id) id() []string {
	ids := []string{i.request}
	if len(i.accept) > 0 {
		ids = append(ids, i.accept)
	}
	if len(i.complete) > 0 {
		ids = append(ids, i.complete)
	}
	return ids
}
