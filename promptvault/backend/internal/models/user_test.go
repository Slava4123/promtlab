package models

import (
	"encoding/json"
	"testing"
)

func TestUser_LegacyLimit_EmptyJSON(t *testing.T) {
	u := &User{}
	if v, ok := u.LegacyLimit("max_prompts"); ok || v != 0 {
		t.Errorf("expected (0,false) for empty JSON, got (%d,%v)", v, ok)
	}
}

func TestUser_LegacyLimit_ExistingField(t *testing.T) {
	u := &User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50}`)}
	v, ok := u.LegacyLimit("max_prompts")
	if !ok || v != 50 {
		t.Errorf("expected (50,true), got (%d,%v)", v, ok)
	}
}

func TestUser_LegacyLimit_MissingField(t *testing.T) {
	u := &User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50}`)}
	if v, ok := u.LegacyLimit("max_collections"); ok || v != 0 {
		t.Errorf("expected (0,false) for missing field, got (%d,%v)", v, ok)
	}
}

func TestUser_LegacyLimit_MultipleFields(t *testing.T) {
	u := &User{LegacyQuotas: json.RawMessage(`{"max_prompts": 50, "max_ext_uses_daily": 100, "max_mcp_uses_daily": 100}`)}

	cases := map[string]int{
		"max_prompts":        50,
		"max_ext_uses_daily": 100,
		"max_mcp_uses_daily": 100,
	}
	for field, want := range cases {
		got, ok := u.LegacyLimit(field)
		if !ok || got != want {
			t.Errorf("%s: expected (%d,true), got (%d,%v)", field, want, got, ok)
		}
	}
}

func TestUser_LegacyLimit_InvalidJSON(t *testing.T) {
	u := &User{LegacyQuotas: json.RawMessage(`not valid json`)}
	if _, ok := u.LegacyLimit("max_prompts"); ok {
		t.Errorf("expected ok=false for invalid JSON, got true (panic protection failed)")
	}
}

func TestUser_LegacyLimit_NonNumeric(t *testing.T) {
	u := &User{LegacyQuotas: json.RawMessage(`{"max_prompts": "fifty"}`)}
	if _, ok := u.LegacyLimit("max_prompts"); ok {
		t.Errorf("expected ok=false for non-numeric value, got true")
	}
}
