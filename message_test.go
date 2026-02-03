package feishubot

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewTextMessage(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    *Message
		wantErr bool
	}{
		{
			name: "valid text message",
			text: "Hello, world!",
			want: &Message{
				MsgType: MsgTypeText,
				Content: map[string]any{
					"text": "Hello, world!",
				},
			},
			wantErr: false,
		},
		{
			name: "empty text message",
			text: "",
			want: &Message{
				MsgType: MsgTypeText,
				Content: map[string]any{
					"text": "",
				},
			},
			wantErr: false,
		},
		{
			name: "text with @ mention",
			text: `<at user_id="ou_xxx">Tom</at> new update`,
			want: &Message{
				MsgType: MsgTypeText,
				Content: map[string]any{
					"text": `<at user_id="ou_xxx">Tom</at> new update`,
				},
			},
			wantErr: false,
		},
		{
			name: "text with @ all",
			text: `<at user_id="all">所有人</at> announcement`,
			want: &Message{
				MsgType: MsgTypeText,
				Content: map[string]any{
					"text": `<at user_id="all">所有人</at> announcement`,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTextMessage(tt.text)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("NewTextMessage() diff = %v", cmp.Diff(tt.want, got))
			}
			if got.MsgType != tt.want.MsgType {
				t.Errorf("NewTextMessage() MsgType = %v, want %v", got.MsgType, tt.want.MsgType)
			}
			if got.Content["text"] != tt.want.Content["text"] {
				t.Errorf("NewTextMessage() Content[text] = %v, want %v", got.Content["text"], tt.want.Content["text"])
			}
		})
	}
}

func TestNewPostMessage(t *testing.T) {
	tests := []struct {
		name    string
		lang    Language
		content *PostContent
		want    *Message
		wantErr bool
	}{
		{
			name: "valid post message with title",
			lang: LanguageZhCN,
			content: NewPostContent(
				"Project Update",
				NewParagraph(NewTextElement("Project has been updated successfully.")),
			),
			want: &Message{
				MsgType: MsgTypePost,
				Content: map[string]any{
					"post": map[string]any{
						"zh_cn": map[string]any{
							"title": "Project Update",
							"content": [][]map[string]any{
								{
									{"tag": "text", "text": "Project has been updated successfully."},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid post message without title",
			lang: LanguageEnUS,
			content: NewPostContent(
				"",
				NewParagraph(NewTextElement("Simple post content")),
			),
			want: &Message{
				MsgType: MsgTypePost,
				Content: map[string]any{
					"post": map[string]any{
						"en_us": map[string]any{
							"title": "",
							"content": [][]map[string]any{
								{
									{"tag": "text", "text": "Simple post content"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "post message with link",
			lang: LanguageZhCN,
			content: NewPostContent(
				"Title",
				NewParagraph(
					NewTextElement("Project has been updated: "),
					NewLinkElement("View", "http://www.example.com/"),
				),
			),
			want: &Message{
				MsgType: MsgTypePost,
				Content: map[string]any{
					"post": map[string]any{
						"zh_cn": map[string]any{
							"title": "Title",
							"content": [][]map[string]any{
								{
									{"tag": "text", "text": "Project has been updated: "},
									{"tag": "a", "text": "View", "href": "http://www.example.com/"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "post message with @ mention",
			lang: LanguageZhCN,
			content: NewPostContent(
				"Notify",
				NewParagraph(
					NewTextElement("Notification for "),
					NewAtElement("ou_xxx", "Tom"),
				),
			),
			want: &Message{
				MsgType: MsgTypePost,
				Content: map[string]any{
					"post": map[string]any{
						"zh_cn": map[string]any{
							"title": "Notify",
							"content": [][]map[string]any{
								{
									{"tag": "text", "text": "Notification for "},
									{"tag": "at", "user_id": "ou_xxx", "user_name": "Tom"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multi-language post message",
			lang: LanguageZhCN,
			content: NewPostContent(
				"Title",
				NewParagraph(NewTextElement("Content")),
			),
			want: &Message{
				MsgType: MsgTypePost,
				Content: map[string]any{
					"post": map[string]any{
						"zh_cn": map[string]any{
							"title": "Title",
							"content": [][]map[string]any{
								{{"tag": "text", "text": "Content"}},
							},
						},
						"en_us": map[string]any{
							"title": "Title",
							"content": [][]map[string]any{
								{{"tag": "text", "text": "Content"}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Message
			if tt.name == "multi-language post message" {
				content := NewPostContent("Title", NewParagraph(NewTextElement("Content")))
				got = NewPostMessageMultiLanguage(
					NewPostLanguageContent(LanguageZhCN, content),
					NewPostLanguageContent(LanguageEnUS, content),
				)
			} else {
				got = NewPostMessage(tt.lang, tt.content)
			}

			if got.MsgType != tt.want.MsgType {
				t.Errorf("NewPostMessage() MsgType = %v, want %v", got.MsgType, tt.want.MsgType)
			}
			// Compare nested structures
			gotBytes, _ := json.Marshal(got.Content)
			wantBytes, _ := json.Marshal(tt.want.Content)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("NewPostMessage() Content = %s, want %s", gotBytes, wantBytes)
			}
		})
	}
}

func TestNewImageMessage(t *testing.T) {
	tests := []struct {
		name     string
		imageKey string
		want     *Message
		wantErr  bool
	}{
		{
			name:     "valid image message",
			imageKey: "img_ecffc3b9-8f14-400f-a014-05eca1a4310g",
			want: &Message{
				MsgType: MsgTypeImage,
				Content: map[string]any{
					"image_key": "img_ecffc3b9-8f14-400f-a014-05eca1a4310g",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewImageMessage(tt.imageKey)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("NewImageMessage() diff = %v", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestNewShareChatMessage(t *testing.T) {
	tests := []struct {
		name        string
		shareChatID string
		want        *Message
		wantErr     bool
	}{
		{
			name:        "valid share chat message",
			shareChatID: "oc_f5b1a7eb27ae2****339ff",
			want: &Message{
				MsgType: MsgTypeShareChat,
				Content: map[string]any{
					"share_chat_id": "oc_f5b1a7eb27ae2****339ff",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewShareChatMessage(tt.shareChatID)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("NewShareChatMessage() diff = %v", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestNewInteractiveMessage(t *testing.T) {
	tests := []struct {
		name string
		card map[string]any
		want *Message
	}{
		{
			name: "interactive message from Card",
			card: map[string]any{
				"schema": "2.0",
				"body": map[string]any{
					"elements": []map[string]any{
						{"tag": "markdown", "content": "Hello!"},
					},
				},
				"header": map[string]any{
					"title":    map[string]any{"tag": "plain_text", "content": "Card Title"},
					"template": "blue",
				},
			},
			want: &Message{
				MsgType: MsgTypeInteractive,
				Card: map[string]any{
					"schema": "2.0",
					"body": map[string]any{
						"elements": []map[string]any{
							{"tag": "markdown", "content": "Hello!"},
						},
					},
					"header": map[string]any{
						"title":    map[string]any{"tag": "plain_text", "content": "Card Title"},
						"template": "blue",
					},
				},
			},
		},
		{
			name: "interactive message with nil card",
			card: nil,
			want: &Message{MsgType: MsgTypeInteractive, Card: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInteractiveMessageFromMap(tt.card)
			if got.MsgType != tt.want.MsgType {
				t.Errorf("NewInteractiveMessage() MsgType = %v, want %v", got.MsgType, tt.want.MsgType)
			}
			if !cmp.Equal(got.Card, tt.want.Card) {
				t.Errorf("NewInteractiveMessage() Card diff = %v", cmp.Diff(tt.want.Card, got.Card))
			}
		})
	}
}

func TestPostElements(t *testing.T) {
	tests := []struct {
		name string
		elem Element
		want []string
	}{
		{
			name: "text element",
			elem: NewTextElement("Hello"),
			want: []string{`"tag":"text"`, `"text":"Hello"`},
		},
		{
			name: "link element",
			elem: NewLinkElement("View", "https://example.com"),
			want: []string{`"tag":"a"`, `"text":"View"`, `"href":"https://example.com"`},
		},
		{
			name: "at element",
			elem: NewAtElement("ou_xxx", "Tom"),
			want: []string{`"tag":"at"`, `"user_id":"ou_xxx"`, `"user_name":"Tom"`},
		},
		{
			name: "image element",
			elem: NewImageElement("img_key_123"),
			want: []string{`"tag":"img"`, `"image_key":"img_key_123"`},
		},
		{
			name: "emoticon element",
			elem: NewEmoticonElement("emoji_key_123"),
			want: []string{`"tag":"emotion"`, `"emoji_key":"emoji_key_123"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.elem)
			got := string(data)
			for _, wantContains := range tt.want {
				if !containsString(got, wantContains) {
					t.Errorf("Element = %v, missing %q", got, wantContains)
				}
			}
		})
	}
}

func TestCardElements(t *testing.T) {
	tests := []struct {
		name string
		elem CardElement
		want []string
	}{
		{
			name: "markdown element",
			elem: NewMarkdownElement("Hello **World**"),
			want: []string{`"tag":"markdown"`, `"content":"Hello **World**"`},
		},
		{
			name: "button element",
			elem: NewButtonElement("Click Me", "primary", "https://example.com"),
			want: []string{`"tag":"button"`, `"tag":"plain_text"`, `"content":"Click Me"`, `"type":"primary"`, `"url":"https://example.com"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.elem)
			got := string(data)
			for _, wantContains := range tt.want {
				if !containsString(got, wantContains) {
					t.Errorf("CardElement = %v, missing %q", got, wantContains)
				}
			}
		})
	}
}

func TestCardBuilder(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		want []string
	}{
		{
			name: "simple card",
			card: NewCard("2.0").
				SetBody(&CardBody{
					Elements: []CardElement{
						NewMarkdownElement("Hello, World!"),
					},
				}),
			want: []string{`"schema":"2.0"`, `"tag":"markdown"`, `"content":"Hello, World!"`},
		},
		{
			name: "card with header",
			card: NewCard("2.0").
				SetHeader(&CardHeader{
					Title:    NewCardTitle("Title"),
					Template: "blue",
				}).
				SetBody(&CardBody{
					Elements: []CardElement{
						NewMarkdownElement("Content"),
					},
				}),
			want: []string{`"schema":"2.0"`, `"tag":"plain_text"`, `"content":"Title"`, `"template":"blue"`, `"content":"Content"`},
		},
		{
			name: "card with config",
			card: NewCard("2.0").
				SetConfig(map[string]any{
					"wide_screen_mode": true,
				}).
				SetBody(&CardBody{
					Elements: []CardElement{
						NewMarkdownElement("Content"),
					},
				}),
			want: []string{`"schema":"2.0"`, `"wide_screen_mode":true`, `"content":"Content"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.card.ToMap()
			data, _ := json.Marshal(got)
			gotStr := string(data)
			for _, wantContains := range tt.want {
				if !containsString(gotStr, wantContains) {
					t.Errorf("Card.ToMap() got %v, missing %q", gotStr, wantContains)
				}
			}
		})
	}
}

func TestMessageJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		message  *Message
		contains []string
	}{
		{
			name:     "text message serialization",
			message:  NewTextMessage("test message"),
			contains: []string{`"msg_type":"text"`, `"text":"test message"`},
		},
		{
			name: "post message serialization",
			message: NewPostMessage(
				LanguageZhCN,
				NewPostContent("Title", NewParagraph(NewTextElement("Content"))),
			),
			contains: []string{`"msg_type":"post"`, `"title":"Title"`, `"tag":"text"`, `"text":"Content"`},
		},
		{
			name: "post message with link serialization",
			message: NewPostMessage(
				LanguageZhCN,
				NewPostContent(
					"Update",
					NewParagraph(
						NewTextElement("View: "),
						NewLinkElement("Details", "https://example.com"),
					),
				),
			),
			contains: []string{`"msg_type":"post"`, `"tag":"a"`, `"href":"https://example.com"`},
		},
		{
			name:     "image message serialization",
			message:  NewImageMessage("img_key_123"),
			contains: []string{`"msg_type":"image"`, `"image_key":"img_key_123"`},
		},
		{
			name:     "share chat message serialization",
			message:  NewShareChatMessage("oc_12345"),
			contains: []string{`"msg_type":"share_chat"`, `"share_chat_id":"oc_12345"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			jsonStr := string(data)
			for _, wantContains := range tt.contains {
				if !containsString(jsonStr, wantContains) {
					t.Errorf("json.Marshal() result does not contain %q\ngot: %s", wantContains, jsonStr)
				}
			}
		})
	}
}

func TestMessageWithSignature(t *testing.T) {
	tests := []struct {
		name      string
		message   *Message
		timestamp int64
		sign      string
		contains  []string
	}{
		{
			name:      "message with timestamp and sign",
			message:   NewTextMessage("test"),
			timestamp: 1599360473,
			sign:      "abc123",
			contains:  []string{`"timestamp":1599360473`, `"sign":"abc123"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.message.Timestamp = tt.timestamp
			tt.message.Sign = tt.sign

			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			jsonStr := string(data)
			for _, wantContains := range tt.contains {
				if !containsString(jsonStr, wantContains) {
					t.Errorf("json.Marshal() result does not contain %q\ngot: %s", wantContains, jsonStr)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if i+len(substr) > len(s) {
			return false
		}
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMsgTypeString(t *testing.T) {
	tests := []struct {
		name string
		t    MsgType
		want string
	}{
		{"text type", MsgTypeText, "text"},
		{"post type", MsgTypePost, "post"},
		{"image type", MsgTypeImage, "image"},
		{"share_chat type", MsgTypeShareChat, "share_chat"},
		{"interactive type", MsgTypeInteractive, "interactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.t)
			if got != tt.want {
				t.Errorf("MsgType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLanguageString(t *testing.T) {
	tests := []struct {
		name string
		l    Language
		want string
	}{
		{"Simplified Chinese", LanguageZhCN, "zh_cn"},
		{"English", LanguageEnUS, "en_us"},
		{"Japanese", LanguageJa, "ja"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.l)
			if got != tt.want {
				t.Errorf("Language.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexPostMessage(t *testing.T) {
	// Test a complex post message with multiple paragraphs and element types
	message := NewPostMessage(
		LanguageZhCN,
		NewPostContent(
			"Daily Report",
			// First paragraph: text + link + @mention
			NewParagraph(
				NewTextElement("Project update: "),
				NewLinkElement("View Details", "https://example.com"),
				NewTextElement(" - "),
				NewAtElement("ou_xxx", "Tom"),
			),
			// Second paragraph: image + text
			NewParagraph(
				NewImageElement("img_key_123"),
				NewTextElement(" Screenshot attached"),
			),
			// Third paragraph: text only
			NewParagraph(
				NewTextElement("Please review by EOD."),
			),
		),
	)

	data, err := json.Marshal(message.Content)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	jsonStr := string(data)

	// Verify content structure
	requiredStrings := []string{
		`"title":"Daily Report"`,
		`"tag":"text"`,
		`"tag":"a"`,
		`"tag":"at"`,
		`"tag":"img"`,
		`"user_id":"ou_xxx"`,
		`"href":"https://example.com"`,
		`"image_key":"img_key_123"`,
	}

	for _, s := range requiredStrings {
		if !containsString(jsonStr, s) {
			t.Errorf("Complex post message missing %q\ngot: %s", s, jsonStr)
		}
	}
}

func TestInteractiveCardMessage(t *testing.T) {
	// Test building an interactive card with header, body, and buttons
	card := NewCard("2.0").
		SetConfig(map[string]any{
			"wide_screen_mode": true,
		}).
		SetHeader(&CardHeader{
			Title:    NewCardTitle("Task Alert"),
			Template: "red",
		}).
		SetBody(&CardBody{
			Direction: "vertical",
			Padding:   "12px 12px 12px 12px",
			Elements: []CardElement{
				NewMarkdownElement("**High Priority Task**\n\nPlease complete this task by end of day."),
				NewButtonElement("View Details", "primary", "https://example.com/task/123"),
				NewButtonElement("Dismiss", "default", "https://example.com/dismiss"),
			},
		})

	message := NewInteractiveMessage(card)

	data, err := json.Marshal(message.Card)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	jsonStr := string(data)

	requiredStrings := []string{
		`"schema":"2.0"`,
		`"wide_screen_mode":true`,
		`"template":"red"`,
		`"tag":"markdown"`,
		`"tag":"button"`,
		`"type":"primary"`,
		`"type":"default"`,
	}

	for _, s := range requiredStrings {
		if !containsString(jsonStr, s) {
			t.Errorf("Interactive card message missing %q\ngot: %s", s, jsonStr)
		}
	}
}
