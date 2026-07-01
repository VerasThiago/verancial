package validator

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeUrlValues(t *testing.T) {
	t.Run("returns values for present keys in key order", func(t *testing.T) {
		values := url.Values{}
		values.Set("name", "Jane")
		values.Set("email", "jane@example.com")

		merged := MergeUrlValues([]string{"name", "email"}, values)

		assert.Equal(t, []string{"Jane", "jane@example.com"}, merged)
	})

	t.Run("skips keys not present in values", func(t *testing.T) {
		values := url.Values{}
		values.Set("name", "Jane")

		merged := MergeUrlValues([]string{"name", "missing"}, values)

		assert.Equal(t, []string{"Jane"}, merged)
	})

	t.Run("empty keys returns nil merged slice", func(t *testing.T) {
		values := url.Values{}
		values.Set("name", "Jane")

		merged := MergeUrlValues([]string{}, values)

		assert.Nil(t, merged)
	})

	t.Run("empty values with keys returns nil merged slice", func(t *testing.T) {
		merged := MergeUrlValues([]string{"name"}, url.Values{})
		assert.Nil(t, merged)
	})

	t.Run("key present but empty string value is still included", func(t *testing.T) {
		values := url.Values{}
		values.Set("name", "")

		merged := MergeUrlValues([]string{"name"}, values)

		assert.Equal(t, []string{""}, merged)
	})
}

func TestGetRulesKey(t *testing.T) {
	t.Run("returns all keys from the rules map", func(t *testing.T) {
		rule := map[string][]string{
			"email": {"required", "email"},
			"name":  {"required"},
		}

		keys := GetRulesKey(rule)

		assert.ElementsMatch(t, []string{"email", "name"}, keys)
		assert.Len(t, keys, 2)
	})

	t.Run("empty map returns empty non-nil slice", func(t *testing.T) {
		keys := GetRulesKey(map[string][]string{})

		assert.NotNil(t, keys)
		assert.Empty(t, keys)
	})

	t.Run("nil map returns empty slice", func(t *testing.T) {
		keys := GetRulesKey(nil)

		assert.NotNil(t, keys)
		assert.Empty(t, keys)
	})
}
