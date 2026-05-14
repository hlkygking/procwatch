package supervisor

import (
	"context"
	"testing"
	"time"
)

func TestProcessTimeoutStore_SetAndGetPolicy(t *testing.T) {
	s := NewProcessTimeoutStore()
	p := ProcessTimeoutPolicy{MaxRuntime: 5 * time.Second, GracePeriod: time.Second}
	s.SetPolicy("web", p)

	got, ok := s.GetPolicy("web")
	if !ok {
		t.Fatal("expected policy to exist")
	}
	if got.MaxRuntime != p.MaxRuntime {
		t.Errorf("MaxRuntime: got %s, want %s", got.MaxRuntime, p.MaxRuntime)
	}
}

func TestProcessTimeoutStore_GetMissing(t *testing.T) {
	s := NewProcessTimeoutStore()
	_, ok := s.GetPolicy("ghost")
	if ok {
		t.Error("expected no policy for unknown process")
	}
}

func TestProcessTimeoutStore_ArmFiresCancel(t *testing.T) {
	s := NewProcessTimeoutStore()
	s.SetPolicy("worker", ProcessTimeoutPolicy{MaxRuntime: 50 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Arm(ctx, "worker", cancel)

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled by timeout")
	}
}

func TestProcessTimeoutStore_ArmNoOpWhenZeroRuntime(t *testing.T) {
	s := NewProcessTimeoutStore()
	s.SetPolicy("idle", ProcessTimeoutPolicy{MaxRuntime: 0})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Arm(ctx, "idle", cancel)

	select {
	case <-ctx.Done():
		t.Fatal("context should not have been cancelled")
	case <-time.After(100 * time.Millisecond):
		// expected: timer did not fire
	}
}

func TestProcessTimeoutStore_DisarmPreventsCancel(t *testing.T) {
	s := NewProcessTimeoutStore()
	s.SetPolicy("svc", ProcessTimeoutPolicy{MaxRuntime: 80 * time.Millisecond})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Arm(ctx, "svc", cancel)
	s.Disarm("svc")

	select {
	case <-ctx.Done():
		t.Fatal("context should not have been cancelled after disarm")
	case <-time.After(200 * time.Millisecond):
		// expected
	}
}

func TestProcessTimeoutPolicy_String(t *testing.T) {
	cases := []struct {
		p    ProcessTimeoutPolicy
		want string
	}{
		{ProcessTimeoutPolicy{}, "no timeout"},
		{ProcessTimeoutPolicy{MaxRuntime: 10 * time.Second, GracePeriod: 2 * time.Second}, "max=10s grace=2s"},
	}
	for _, tc := range cases {
		got := tc.p.String()
		if got != tc.want {
			t.Errorf("String() = %q, want %q", got, tc.want)
		}
	}
}
