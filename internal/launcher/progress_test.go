package launcher

import (
	"testing"
	"time"
)

func TestNoopReporter_ImplementsInterface(t *testing.T) {
	var r ProgressReporter = NewNoopReporter()
	r.Status("hello")
	r.Progress(0.5)
	select {
	case <-r.OnSkip():
		t.Error("noop reporter should never close OnSkip")
	case <-time.After(20 * time.Millisecond):
		// ok
	}
}

func TestChannelReporter_RecordsStatusAndProgress(t *testing.T) {
	r := NewChannelReporter()
	r.Status("creating backup")
	r.Progress(0.3)
	if r.LastStatus() != "creating backup" {
		t.Errorf("status = %q", r.LastStatus())
	}
	if r.LastProgress() != 0.3 {
		t.Errorf("progress = %f", r.LastProgress())
	}
}

func TestChannelReporter_Skip_ClosesChannel(t *testing.T) {
	r := NewChannelReporter()
	r.Skip()
	select {
	case <-r.OnSkip():
		// ok
	case <-time.After(50 * time.Millisecond):
		t.Error("expected OnSkip channel to be closed after Skip()")
	}
}

func TestChannelReporter_Skip_Idempotent(t *testing.T) {
	r := NewChannelReporter()
	r.Skip()
	r.Skip() // calling again must not panic on close-of-closed-channel
}
