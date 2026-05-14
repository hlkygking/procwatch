package supervisor

import (
	"testing"
	"time"
)

func TestProcessCooldown_NotCooledImmediately(t *testing.T) {
	cd := NewProcessCooldown(CooldownPolicy{StableDuration: 10 * time.Second})
	cd.RecordStart("svc")
	if cd.IsCooled("svc") {
		t.Fatal("expected process to not be cooled immediately after start")
	}
}

func TestProcessCooldown_CooledAfterDuration(t *testing.T) {
	cd := NewProcessCooldown(CooldownPolicy{StableDuration: 5 * time.Second})
	now := time.Now()
	cd.nowFunc = func() time.Time { return now }
	cd.RecordStart("svc")

	// Advance time past stable duration
	cd.nowFunc = func() time.Time { return now.Add(6 * time.Second) }
	stable := cd.RecordStable("svc")

	if !stable {
		t.Fatal("expected RecordStable to return true after duration elapsed")
	}
	if !cd.IsCooled("svc") {
		t.Fatal("expected process to be cooled after stable duration")
	}
}

func TestProcessCooldown_NotCooledBeforeDuration(t *testing.T) {
	cd := NewProcessCooldown(CooldownPolicy{StableDuration: 10 * time.Second})
	now := time.Now()
	cd.nowFunc = func() time.Time { return now }
	cd.RecordStart("svc")

	cd.nowFunc = func() time.Time { return now.Add(3 * time.Second) }
	stable := cd.RecordStable("svc")

	if stable {
		t.Fatal("expected RecordStable to return false before duration elapsed")
	}
	if cd.IsCooled("svc") {
		t.Fatal("expected process to not be cooled before stable duration")
	}
}

func TestProcessCooldown_RecordStartResetsCooled(t *testing.T) {
	cd := NewProcessCooldown(CooldownPolicy{StableDuration: 1 * time.Second})
	now := time.Now()
	cd.nowFunc = func() time.Time { return now }
	cd.RecordStart("svc")
	cd.nowFunc = func() time.Time { return now.Add(2 * time.Second) }
	cd.RecordStable("svc")

	if !cd.IsCooled("svc") {
		t.Fatal("expected cooled after stable duration")
	}

	// Restart resets cooldown
	cd.RecordStart("svc")
	if cd.IsCooled("svc") {
		t.Fatal("expected cooldown to reset after new start")
	}
}

func TestProcessCooldown_Reset(t *testing.T) {
	cd := NewProcessCooldown(DefaultCooldownPolicy())
	cd.RecordStart("svc")
	cd.Reset("svc")

	if cd.IsCooled("svc") {
		t.Fatal("expected IsCooled to be false after Reset")
	}
	if cd.RecordStable("svc") {
		t.Fatal("expected RecordStable to return false for unknown process after Reset")
	}
}

func TestProcessCooldown_DefaultPolicy(t *testing.T) {
	p := DefaultCooldownPolicy()
	if p.StableDuration != 30*time.Second {
		t.Fatalf("expected 30s stable duration, got %v", p.StableDuration)
	}
}

func TestProcessCooldown_UnknownProcess(t *testing.T) {
	cd := NewProcessCooldown(DefaultCooldownPolicy())
	if cd.IsCooled("ghost") {
		t.Fatal("expected IsCooled to be false for unknown process")
	}
}
