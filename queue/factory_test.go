package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/session"
	"github.com/zhiyunliu/golibs/xtypes"
)

// mockMQP is a mock implementation of IMQP for testing
type mockMQP struct {
	name            string
	pushError       error
	batchPushError  error
	delayPushError  error
	closeError      error
	pushCalled      bool
	batchPushCalled bool
	delayPushCalled bool
}

func (m *mockMQP) Name() string {
	return m.name
}

func (m *mockMQP) Push(ctx context.Context, key string, value Message) error {
	m.pushCalled = true
	return m.pushError
}

func (m *mockMQP) BatchPush(ctx context.Context, key string, value ...Message) error {
	m.batchPushCalled = true
	return m.batchPushError
}

func (m *mockMQP) DelayPush(ctx context.Context, key string, value Message, delaySeconds int64) error {
	m.delayPushCalled = true
	return m.delayPushError
}

func (m *mockMQP) Close() error {
	return m.closeError
}

// mockConfig is a mock implementation of config.Config for testing
type mockConfig struct{}

func (m *mockConfig) Get(path string) config.Config {
	return m
}

func (m *mockConfig) Value(key string) config.Value {
	return &mockValue{}
}

func (m *mockConfig) Load() error {
	return nil
}
func (m *mockConfig) Source(sources ...config.Source) error {
	return nil
}

func (m *mockConfig) Scan(v interface{}) error {
	return nil
}

func (m *mockConfig) ScanTo(v interface{}) error {
	return nil
}

func (m *mockConfig) Watch(key string, o config.Observer) error {
	return nil
}
func (m *mockConfig) Close() error {
	return nil
}
func (m *mockConfig) Path() string {
	return ""
}

func (m *mockConfig) Root() config.Config {
	return m
}

// mockValue is a mock implementation of config.Value for testing
type mockValue struct{}

func (m *mockValue) Bool() (bool, error) {
	return false, nil
}

func (m *mockValue) Int() (int64, error) {
	return 0, nil
}

func (m *mockValue) Int64() int64 {
	return 0
}

func (m *mockValue) Float64() float64 {
	return 0
}
func (m *mockValue) Float() (float64, error) {
	return 0, nil
}

func (m *mockValue) String() string {
	return ""
}

func (m *mockValue) Duration() (time.Duration, error) {
	return 0, nil
}

func (m *mockValue) Unmarshal(val interface{}) error {
	return nil
}

func (m *mockValue) Scan(val interface{}) error {
	return m.ScanTo(val)
}
func (m *mockValue) ScanTo(val interface{}) error {
	return nil
}

func (m *mockValue) Close() bool {
	return false
}

func (v *mockValue) Exists() bool {
	return true
}
func (v *mockValue) Load() interface{} {
	return true
}

func (v *mockValue) Store(interface{}) {

}

func (v *mockValue) Map() (map[string]config.Value, error) {
	return map[string]config.Value{}, nil
}
func (v *mockValue) Slice() ([]config.Value, error) {
	return []config.Value{}, nil
}
func TestNewQueue(t *testing.T) {
	// Save original mqpResolvers
	originalResolvers := make(map[string]MqpResolver)
	for k, v := range mqpResolvers {
		originalResolvers[k] = v
	}
	defer func() {
		// Restore original mqpResolvers
		mqpResolvers = originalResolvers
	}()

	// Register a mock resolver
	mqpResolvers["mock"] = &mockMQPResolver{}

	cfg := &mockConfig{}
	q, err := newQueue("mock", cfg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if q == nil {
		t.Error("Expected queue to be created")
	}
}

func TestNewQueueWithError(t *testing.T) {
	// Save original mqpResolvers
	originalResolvers := make(map[string]MqpResolver)
	for k, v := range mqpResolvers {
		originalResolvers[k] = v
	}
	defer func() {
		// Restore original mqpResolvers
		mqpResolvers = originalResolvers
	}()

	cfg := &mockConfig{}
	_, err := newQueue("unknown", cfg)
	if err == nil {
		t.Error("Expected error for unknown protocol")
	}
}

func TestQueueSend(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.Send(ctx, "test-key", "test-value")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockMQP.pushCalled {
		t.Error("Expected Push to be called")
	}
}

func TestQueueSendWithEmptyKey(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.Send(ctx, "", "test-value")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestQueueSendWithNilValue(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.Send(ctx, "test-key", nil)
	if err == nil {
		t.Error("Expected error for nil value")
	}
}

func TestQueueBatchSend(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.BatchSend(ctx, "test-key", "value1", "value2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockMQP.batchPushCalled {
		t.Error("Expected BatchPush to be called")
	}
}

func TestQueueBatchSendWithEmptyKey(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.BatchSend(ctx, "", "value1", "value2")
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestQueueDelaySend(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.DelaySend(ctx, "test-key", "test-value", 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockMQP.delayPushCalled {
		t.Error("Expected DelayPush to be called")
	}
}

func TestQueueDelaySendWithEmptyKey(t *testing.T) {
	mockMQP := &mockMQP{name: "mock"}
	q := &queue{q: mockMQP}

	ctx := context.Background()
	err := q.DelaySend(ctx, "", "test-value", 10)
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestQueueBuildMessage(t *testing.T) {
	q := &queue{}

	ctx := context.Background()
	msg, err := q.buildMessage(ctx, "test-value")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if msg == nil {
		t.Error("Expected message to be created")
	}
}

func TestQueueBuildMessageWithNil(t *testing.T) {
	q := &queue{}

	ctx := context.Background()
	_, err := q.buildMessage(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil value")
	}
}

func TestQueueBuildMessageWithContextSession(t *testing.T) {
	q := &queue{}

	ctx := context.Background()
	ctx = session.WithContext(ctx, "test-session-id")
	msg, err := q.buildMessage(ctx, "test-value")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if msg == nil {
		t.Error("Expected message to be created")
	}
	if msg.Header()[constants.HeaderRequestId] != "test-session-id" {
		t.Error("Expected request ID to be set from context")
	}
}

func TestQueueClose(t *testing.T) {
	mockMQP := &mockMQP{name: "mock", closeError: errors.New("close error")}
	q := &queue{q: mockMQP}

	err := q.Close()
	if err == nil {
		t.Error("Expected close error")
	}
}

func TestQueueMessageStructure(t *testing.T) {
	// Test that message has correct structure
	msg, err := NewMsg("test body", WithHeader("test-key", "test-value"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if msg.Header()["test-key"] != "test-value" {
		t.Error("Expected header to be set")
	}

	body := msg.Body()
	var result string = string(body)
	if result != "test body" {
		t.Error("Expected body to match")
	}
}

func TestQueueMessageWithXTypes(t *testing.T) {
	data := xtypes.XMap{"key": "value"}
	msg, err := NewMsg(data)
	if msg == nil || err != nil {
		t.Error("Expected message to be created from xtypes")
	}
}

// mockMQPResolver is a mock implementation of MqpResolver for testing
type mockMQPResolver struct{}

func (m *mockMQPResolver) Name() string {
	return "mock"
}

func (m *mockMQPResolver) Resolve(setting config.Config, opts ...Option) (IMQP, error) {
	return &mockMQP{name: "mock"}, nil
}
