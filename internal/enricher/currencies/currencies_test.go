package currencies_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/enricher/currencies"
	"github.com/stretchr/testify/assert"
)

func TestUnknownCurrency(t *testing.T) {
	_, err := currencies.ConvertCurrency(
		context.TODO(),
		currencies.ExchangeApiMock{},
		0,
		"",
	)
	assert.Error(t, err)
}

func TestConvertCurrency(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		currency string
		expected float64
	}{
		{
			"0 EUR",
			0,
			"EUR",
			0,
		},
		{
			"0 USD",
			0,
			"USD",
			0,
		},
		{
			"10 EUR",
			10,
			"EUR",
			10,
		},
		{
			"10 USD",
			10,
			"USD",
			9.228037731600677,
		},
	}

	mockApi := currencies.ExchangeApiMock{}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf(tc.name), func(t *testing.T) {
			actual, err := currencies.ConvertCurrency(context.TODO(), mockApi, tc.value, tc.currency)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
