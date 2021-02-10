package restlib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlEncodeStructOfStrings(t *testing.T) {
	type BananaRequest struct {
		Banana     string
		BananaType string
	}

	req := &BananaRequest{
		Banana:     "ripe",
		BananaType: "wrapped",
	}
	data, err := urlencode(req)
	require.NoError(t, err)
	expectedData := []byte(`Banana=ripe&BananaType=wrapped`)
	require.Equal(t, expectedData, data)
}

func TestUrlEncodeStructWithNullableStrings(t *testing.T) {
	type BananaRequest struct {
		Banana     string
		BananaType *string
	}

	req := &BananaRequest{
		Banana: "ripe",
	}
	data, err := urlencode(req)
	require.NoError(t, err)
	expectedData := []byte(`Banana=ripe&BananaType=`)
	require.Equal(t, expectedData, data)
}

func TestUrlEncodeStructWithInteger(t *testing.T) {
	type BananaRequest struct {
		Banana       string
		Multiplicity int
	}

	req := &BananaRequest{
		Banana:       "ripe",
		Multiplicity: 6,
	}
	data, err := urlencode(req)
	require.NoError(t, err)
	expectedData := []byte(`Banana=ripe&Multiplicity=6`)
	require.Equal(t, expectedData, data)
}
