package bybit_test

import (
	"encoding/json"
	"testing"

	"github.com/lesomnus/tiny-short/bybit"
	"github.com/stretchr/testify/require"
)

func TestTransferIdJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		require := require.New(t)

		v := bybit.TransferId{
			0x12, 0x34, 0x56, 0x78,
			0x9a, 0xbc, 0xde, 0xf1,
			0x23, 0x45, 0x67, 0x89,
			0xab, 0xcd, 0xef, 0x12,
		}
		data, err := json.Marshal(&v)
		require.NoError(err)
		require.Equal(`"12345678-9abc-def1-2345-6789abcdef12"`, string(data))
	})

	t.Run("unmarshal", func(t *testing.T) {
		require := require.New(t)

		var v bybit.TransferId
		err := json.Unmarshal([]byte(`"12345678-9abc-def1-2345-6789abcdef12"`), &v)
		require.NoError(err)
		require.Equal(bybit.TransferId{
			0x12, 0x34, 0x56, 0x78,
			0x9a, 0xbc, 0xde, 0xf1,
			0x23, 0x45, 0x67, 0x89,
			0xab, 0xcd, 0xef, 0x12,
		}, v)
	})

	t.Run("generate random if nil when marshal", func(t *testing.T) {
		require := require.New(t)

		var v bybit.TransferId
		data, err := json.Marshal(&v)
		require.NoError(err)
		require.NotEqual(`"00000000-0000-0000-0000-000000000000"`, string(data))
	})

	t.Run("marshal in struct", func(t *testing.T) {
		require := require.New(t)

		s := struct {
			Foo bybit.TransferId `json:"foo"`
		}{
			Foo: bybit.TransferId{
				0x12, 0x34, 0x56, 0x78,
				0x9a, 0xbc, 0xde, 0xf1,
				0x23, 0x45, 0x67, 0x89,
				0xab, 0xcd, 0xef, 0x12,
			},
		}
		data, err := json.Marshal(s)
		require.NoError(err)
		require.Equal(`{"foo":"12345678-9abc-def1-2345-6789abcdef12"}`, string(data))
	})

	t.Run("unmarshal from struct", func(t *testing.T) {
		require := require.New(t)

		s := struct {
			Foo bybit.TransferId `json:"foo"`
		}{
			Foo: bybit.TransferId{},
		}
		err := json.Unmarshal([]byte(`{"foo":"12345678-9abc-def1-2345-6789abcdef12"}`), &s)
		require.NoError(err)
		require.Equal(bybit.TransferId{
			0x12, 0x34, 0x56, 0x78,
			0x9a, 0xbc, 0xde, 0xf1,
			0x23, 0x45, 0x67, 0x89,
			0xab, 0xcd, 0xef, 0x12,
		}, s.Foo)
	})
}

func TestAmountUnmarshalJSON(t *testing.T) {
	t.Run("string literal", func(t *testing.T) {
		require := require.New(t)

		amount := bybit.Amount(0)
		err := json.Unmarshal([]byte("\"3.14\""), &amount)
		require.NoError(err)
		require.Equal(bybit.Amount(3.14), amount)
	})

	t.Run("struct field", func(t *testing.T) {
		require := require.New(t)

		s := struct {
			Amount bybit.Amount `json:"foo"`
		}{
			Amount: 0,
		}
		err := json.Unmarshal([]byte("{\"foo\": \"3.14\"}"), &s)
		require.NoError(err)
		require.Equal(bybit.Amount(3.14), s.Amount)
	})
}
