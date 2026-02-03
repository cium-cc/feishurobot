package feishubot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MockHTTPClient is a mock HTTP client for testing.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

// TestNewClient tests the NewClient constructor.
func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		webhook   string
		secret    string
		wantError bool
	}{
		{
			name:      "valid client with secret",
			webhook:   "https://open.feishu.cn/open-apis/bot/v2/hook/abc123",
			secret:    "my_secret",
			wantError: false,
		},
		{
			name:      "valid client without secret",
			webhook:   "https://open.feishu.cn/open-apis/bot/v2/hook/abc123",
			secret:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.webhook, tt.secret)
			if client == nil {
				t.Fatal("NewClient() returned nil")
			}
			if client.WebhookURL != tt.webhook {
				t.Errorf("NewClient() WebhookURL = %v, want %v", client.WebhookURL, tt.webhook)
			}
			if client.Secret != tt.secret {
				t.Errorf("NewClient() Secret = %v, want %v", client.Secret, tt.secret)
			}
			if client.HTTPClient == nil {
				t.Error("NewClient() HTTPClient should not be nil")
			}
		})
	}
}

// TestSend tests the Send method with various scenarios.
func TestSend(t *testing.T) {
	tests := []struct {
		name       string
		webhook    string
		secret     string
		message    *Message
		handler    http.HandlerFunc
		wantError  bool
		wantCode   int
		wantFields []string
	}{
		{
			name:    "successful send without signature",
			webhook: "/webhook",
			secret:  "",
			message: NewTextMessage("test message"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var msg Message
				err := json.NewDecoder(r.Body).Decode(&msg)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(SuccessResponse)
			},
			wantError:  false,
			wantCode:   0,
			wantFields: []string{`"msg_type":"text"`, `"text":"test message"`},
		},
		{
			name:    "successful send with signature",
			webhook: "/webhook",
			secret:  "test_secret",
			message: NewTextMessage("test message"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var msg Message
				err := json.NewDecoder(r.Body).Decode(&msg)
				require.NoError(t, err)

				// Verify timestamp and sign are present
				require.NotZero(t, msg.Timestamp)
				require.NotEmpty(t, msg.Sign)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(SuccessResponse)
			},
			wantError:  false,
			wantCode:   0,
			wantFields: []string{`"timestamp":`, `"sign":"`},
		},
		{
			name:    "send post message",
			webhook: "/webhook",
			secret:  "",
			message: NewPostMessage(
				LanguageZhCN,
				NewPostContent("Title", NewParagraph(NewTextElement("Content"))),
			),
			handler: func(w http.ResponseWriter, r *http.Request) {
				var msg Message
				err := json.NewDecoder(r.Body).Decode(&msg)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(SuccessResponse)
			},
			wantError:  false,
			wantCode:   0,
			wantFields: []string{`"msg_type":"post"`},
		},
		{
			name:    "server returns error",
			webhook: "/webhook",
			secret:  "",
			message: NewTextMessage("test"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{
					Code: 9499,
					Msg:  "Bad Request",
				})
			},
			wantError: true,
			wantCode:  9499,
		},
		{
			name:    "keyword not found error",
			webhook: "/webhook",
			secret:  "",
			message: NewTextMessage("test"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(Response{
					Code: 19024,
					Msg:  "Key Words Not Found",
				})
			},
			wantError: true,
			wantCode:  19024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL+tt.webhook, tt.secret)

			resp, err := client.Send(context.Background(), tt.message)

			if tt.wantError {
				require.Error(t, err)
				if resp != nil {
					require.Equal(t, tt.wantCode, resp.Code)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tt.wantCode, resp.Code)
			}
		})
	}
}

// TestSendWithContext tests the Send method with context cancellation.
func TestSendWithContext(t *testing.T) {
	tests := []struct {
		name        string
		cancelFunc  context.CancelFunc
		wantError   bool
		errorSubstr string
	}{
		{
			name:        "context cancelled",
			wantError:   true,
			errorSubstr: "context",
		},
		{
			name:        "context timeout",
			wantError:   true,
			errorSubstr: "deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a slow server that will take longer than the context timeout
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(SuccessResponse)
			}))
			defer server.Close()

			client := NewClient(server.URL+"/webhook", "")

			var ctx context.Context
			var cancel context.CancelFunc

			if tt.name == "context cancelled" {
				ctx, cancel = context.WithCancel(context.Background())
				cancel() // Immediately cancel
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 1*time.Microsecond)
				defer cancel()
			}

			_, err := client.Send(ctx, NewTextMessage("test"))

			if tt.wantError {
				require.Error(t, err)
			}
		})
	}
}

// TestWithHTTPClient tests setting a custom HTTP client.
func TestWithHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse)
	}))
	defer server.Close()

	mockClient := &MockHTTPClient{}
	client := NewClient(server.URL+"/webhook", "")
	client.SetHTTPClient(mockClient)

	require.Same(t, mockClient, client.HTTPClient)
}

// TestSignatureGeneration tests that signatures are correctly generated when sending.
func TestSignatureGeneration(t *testing.T) {
	secret := "test_secret"
	var capturedMsg Message

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		require.NoError(t, err)
		capturedMsg = msg

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/webhook", secret)
	_, err := client.Send(context.Background(), NewTextMessage("test"))

	require.NoError(t, err)
	require.NotZero(t, capturedMsg.Timestamp)
	require.NotEmpty(t, capturedMsg.Sign)

	// Verify the signature is correct
	expectedSign, err := GenSign(secret, capturedMsg.Timestamp)
	require.NoError(t, err)
	require.Equal(t, expectedSign, capturedMsg.Sign)
}

// TestRequestHeaders tests that correct headers are sent.
func TestRequestHeaders(t *testing.T) {
	var capturedContentType string
	var capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedMethod = r.Method

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL+"/webhook", "")
	_, err := client.Send(context.Background(), NewTextMessage("test"))

	require.NoError(t, err)
	require.Equal(t, "application/json", capturedContentType)
	require.Equal(t, "POST", capturedMethod)
}

// TestMessageTypes tests sending all supported message types.
func TestMessageTypes(t *testing.T) {
	messageTypes := []struct {
		name    string
		message *Message
	}{
		{"text", NewTextMessage("text message")},
		{
			"post",
			NewPostMessage(
				LanguageZhCN,
				NewPostContent("Title", NewParagraph(NewTextElement("Content"))),
			),
		},
		{"image", NewImageMessage("img_key_123")},
		{"share_chat", NewShareChatMessage("oc_12345")},
		{"interactive", NewInteractiveMessageFromMap(map[string]any{"schema": "2.0"})},
	}

	for _, mt := range messageTypes {
		t.Run(mt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var msg Message
				err := json.NewDecoder(r.Body).Decode(&msg)
				require.NoError(t, err)
				require.Equal(t, mt.message.MsgType, msg.MsgType)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(SuccessResponse)
			}))
			defer server.Close()

			client := NewClient(server.URL+"/webhook", "")
			_, err := client.Send(context.Background(), mt.message)

			require.NoError(t, err)
		})
	}
}

// SuccessResponse is a standard success response from Feishu API.
var SuccessResponse = Response{
	Code: 0,
	Msg:  "success",
	Data: make(map[string]any),
}
