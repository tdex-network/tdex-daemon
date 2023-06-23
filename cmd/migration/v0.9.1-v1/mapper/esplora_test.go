package mapper

import "testing"

func TestGetUnspentStatus(t *testing.T) {
	t.Skip()

	mapperSvc := NewService(nil, "http://localhost:3001")
	r, err := mapperSvc.GetUnspentStatus(
		"e7001a9d4e30c43608c37a379ee34677f71d6ef7a8cfda665a5a9b9aa720701f", 2,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(r.Spent)
	t.Log(r.Status.Confirmed)
	t.Log(r.Status.BlockHeight)
	t.Log(r.Status.BlockHash)
	t.Log(r.Status.BlockTime)
}
