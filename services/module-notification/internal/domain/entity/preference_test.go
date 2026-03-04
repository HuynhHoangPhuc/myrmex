package entity_test

import (
	"testing"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/domain/entity"
)

func TestDefaultPreferences_Count(t *testing.T) {
	prefs := entity.DefaultPreferences()
	want := len(entity.AllEventTypes) * len(entity.AllChannels) // 12 × 2 = 24
	if len(prefs) != want {
		t.Errorf("DefaultPreferences() = %d items, want %d", len(prefs), want)
	}
}

func TestDefaultPreferences_AllEnabled(t *testing.T) {
	for _, p := range entity.DefaultPreferences() {
		if !p.Enabled {
			t.Errorf("default preference %s:%s should be enabled", p.EventType, p.Channel)
		}
	}
}

func TestDefaultPreferences_CoversAllCombinations(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range entity.DefaultPreferences() {
		seen[p.EventType+":"+p.Channel] = true
	}
	for _, et := range entity.AllEventTypes {
		for _, ch := range entity.AllChannels {
			if !seen[et+":"+ch] {
				t.Errorf("missing combination %s:%s in DefaultPreferences", et, ch)
			}
		}
	}
}
