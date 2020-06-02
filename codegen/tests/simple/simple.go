package simple

import (
	"context"
)

func GetRawList(ctx context.Context, req *GetRawListRequest, client GetRawListClient) (*Str, error) {
	var s Str = "raw"

	return &s, nil
}

func GetRawIntList(ctx context.Context, req *GetRawIntListRequest, client GetRawIntListClient) (*Integer, error) {
	var s Integer = 123

	return &s, nil
}

func GetRawIdStatesList(ctx context.Context, req *GetRawIdStatesListRequest, client GetRawIdStatesListClient) (*Str, error) {
	var s Str = "raw"

	return &s, nil
}

func GetStuffList(ctx context.Context, req *GetStuffListRequest, client GetStuffListClient) (*Stuff, error) {
	s := Stuff{
		InnerStuff: "response",
		ResponseStuff: Response{
			Data: ItemSet{
				M: map[string]Item{
					"John": {
						A1:   "CollinsSt",
						A2:   "LonasDaleSt",
						Name: "John",
					},
					"James": {
						A1:   "SpencerSt",
						A2:   "CollinsSt",
						Name: "James",
					},
				},
			},
		},
	}

	return &s, nil
}
