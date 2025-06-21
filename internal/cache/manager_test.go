package cache

import (
	"testing"
	"time"
)

func TestManager_SetAndGet(t *testing.T) {
	m := NewManager()

	// Test basic set and get
	key := "test_key"
	value := "test_value"
	m.Set(key, value, time.Minute)

	retrieved, found := m.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}

	if retrieved != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
}

func TestManager_GetString(t *testing.T) {
	m := NewManager()

	// Test string retrieval
	key := "string_key"
	value := "string_value"
	m.Set(key, value, time.Minute)

	retrieved, found := m.GetString(key)
	if !found {
		t.Error("Expected to find cached string value")
	}

	if retrieved != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}

	// Test non-string value
	m.Set("non_string", 123, time.Minute)
	_, found = m.GetString("non_string")
	if found {
		t.Error("Expected not to find non-string value as string")
	}
}

func TestManager_SetAndGetJSON(t *testing.T) {
	m := NewManager()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestStruct{
		Name:  "test",
		Value: 42,
	}

	// Test JSON set and get
	key := "json_key"
	err := m.SetJSON(key, original, time.Minute)
	if err != nil {
		t.Fatalf("Failed to set JSON: %v", err)
	}

	var retrieved TestStruct
	found := m.GetJSON(key, &retrieved)
	if !found {
		t.Error("Expected to find cached JSON value")
	}

	if retrieved.Name != original.Name || retrieved.Value != original.Value {
		t.Errorf("Expected %+v, got %+v", original, retrieved)
	}
}

func TestManager_Delete(t *testing.T) {
	m := NewManager()

	key := "delete_key"
	value := "delete_value"
	m.Set(key, value, time.Minute)

	// Verify it exists
	_, found := m.Get(key)
	if !found {
		t.Error("Expected to find value before deletion")
	}

	// Delete and verify it's gone
	m.Delete(key)
	_, found = m.Get(key)
	if found {
		t.Error("Expected value to be deleted")
	}
}

func TestManager_Flush(t *testing.T) {
	m := NewManager()

	// Add multiple items
	m.Set("key1", "value1", time.Minute)
	m.Set("key2", "value2", time.Minute)
	m.Set("key3", "value3", time.Minute)

	if m.ItemCount() != 3 {
		t.Errorf("Expected 3 items, got %d", m.ItemCount())
	}

	// Flush all
	m.Flush()

	if m.ItemCount() != 0 {
		t.Errorf("Expected 0 items after flush, got %d", m.ItemCount())
	}
}

func TestManager_CacheKeys(t *testing.T) {
	m := NewManager()

	// Test currency list key
	key := m.GetCurrencyListKey()
	expected := string(CurrencyListKey)
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}

	// Test currency detail key
	currencyID := 1
	key = m.GetCurrencyDetailKey(currencyID)
	expected = "currency:detail:1"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}

	// Test wiki search key
	query := "test query"
	key = m.GetWikiSearchKey(query)
	expected = "wiki:search:test query"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}

	// Test wiki page key
	pageTitle := "Test Page"
	key = m.GetWikiPageKey(pageTitle)
	expected = "wiki:page:Test Page"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}

	// Test wallet key
	apiKeyHash := "abcd1234"
	key = m.GetWalletKey(apiKeyHash)
	expected = "wallet:abcd1234"
	if key != expected {
		t.Errorf("Expected %s, got %s", expected, key)
	}
}

func TestManager_TTLExpiration(t *testing.T) {
	m := NewManager()

	key := "ttl_key"
	value := "ttl_value"
	ttl := 100 * time.Millisecond

	m.Set(key, value, ttl)

	// Should exist immediately
	_, found := m.Get(key)
	if !found {
		t.Error("Expected to find value immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = m.Get(key)
	if found {
		t.Error("Expected value to be expired")
	}
}
