package gateway

import (
	"errors"
)

type EXTERNAL_MissingType struct {
}

func unmarshalJSONWithValidationEXTERNAL_MissingType(_ []byte) (*EXTERNAL_MissingType, bool, error) {
	return nil, false, errors.New("always fail for MissingType")
}
