package launcher

import "sync"

// ProgressReporter — interfejs do raportowania postępu auto-launch
// z launcher do UI. UI implementuje go w ProgressWindow.
type ProgressReporter interface {
	Status(msg string)
	Progress(fraction float64)
	OnSkip() <-chan struct{}
}

// noopReporter — milcząca implementacja używana w testach i tle (CLI silent mode).
type noopReporter struct{ ch chan struct{} }

func NewNoopReporter() ProgressReporter {
	return &noopReporter{ch: make(chan struct{})}
}
func (n *noopReporter) Status(string)           {}
func (n *noopReporter) Progress(float64)        {}
func (n *noopReporter) OnSkip() <-chan struct{} { return n.ch }

// ChannelReporter — buforuje ostatnie wartości; baza dla UI ProgressWindow,
// które odczytuje LastStatus/LastProgress i odświeża widget'y.
type ChannelReporter struct {
	mu       sync.Mutex
	status   string
	progress float64
	skipCh   chan struct{}
	once     sync.Once
}

func NewChannelReporter() *ChannelReporter {
	return &ChannelReporter{skipCh: make(chan struct{})}
}

func (c *ChannelReporter) Status(msg string) {
	c.mu.Lock()
	c.status = msg
	c.mu.Unlock()
}

func (c *ChannelReporter) Progress(f float64) {
	c.mu.Lock()
	c.progress = f
	c.mu.Unlock()
}

func (c *ChannelReporter) OnSkip() <-chan struct{} { return c.skipCh }

// Skip oznacza chęć użytkownika "pomiń backup, uruchom grę". Idempotentne —
// druga próba close'a chronion przez sync.Once.
func (c *ChannelReporter) Skip() {
	c.once.Do(func() { close(c.skipCh) })
}

func (c *ChannelReporter) LastStatus() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *ChannelReporter) LastProgress() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.progress
}
