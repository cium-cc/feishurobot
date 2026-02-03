package feishubot

// MsgType represents the type of message to send.
type MsgType string

const (
	// MsgTypeText represents a plain text message.
	MsgTypeText MsgType = "text"

	// MsgTypePost represents a rich text (post) message.
	MsgTypePost MsgType = "post"

	// MsgTypeImage represents an image message.
	MsgTypeImage MsgType = "image"

	// MsgTypeShareChat represents a share chat (group card) message.
	MsgTypeShareChat MsgType = "share_chat"

	// MsgTypeInteractive represents an interactive card message.
	MsgTypeInteractive MsgType = "interactive"
)

// Language represents the language for post messages.
type Language string

const (
	// LanguageZhCN represents Simplified Chinese.
	LanguageZhCN Language = "zh_cn"

	// LanguageEnUS represents English.
	LanguageEnUS Language = "en_us"

	// LanguageJa represents Japanese.
	LanguageJa Language = "ja"
)

// Message represents a message to be sent to Feishu webhook.
type Message struct {
	MsgType   MsgType        `json:"msg_type"`
	Content   map[string]any `json:"content,omitempty"`
	Card      map[string]any `json:"card,omitempty"`
	Timestamp int64          `json:"timestamp,omitempty"`
	Sign      string         `json:"sign,omitempty"`
}

// PostContent represents the content of a rich text (post) message.
type PostContent struct {
	Title   string      `json:"title,omitempty"`
	Content []Paragraph `json:"content"`
}

// PostLanguageContent represents post content for a specific language.
type PostLanguageContent struct {
	Language Language    `json:"-"` // Not serialized directly, used as map key
	Content  PostContent `json:"-"` // Not serialized directly
}

// Paragraph represents a paragraph in a post message.
// A paragraph is an array of elements (text, link, @mention, image, etc.).
type Paragraph []Element

// Element represents an element within a paragraph.
// Valid element types: TextElement, LinkElement, AtElement, ImageElement.
type Element map[string]any

// NewTextElement creates a plain text element.
func NewTextElement(text string) Element {
	return map[string]any{
		"tag":  "text",
		"text": text,
	}
}

// NewLinkElement creates a hyperlink element.
// The href must be a valid URL, otherwise the message will fail to send.
func NewLinkElement(text, href string) Element {
	return map[string]any{
		"tag":  "a",
		"text": text,
		"href": href,
	}
}

// NewAtElement creates an @ mention element.
// For external groups, only Open ID is supported for @ single user.
// For @ all, use "all" as user_id.
func NewAtElement(userID, userName string) Element {
	return map[string]any{
		"tag":       "at",
		"user_id":   userID,
		"user_name": userName,
	}
}

// NewImageElement creates an image element within a paragraph.
// The imageKey must be obtained from Feishu image upload API.
func NewImageElement(imageKey string) Element {
	return map[string]any{
		"tag":       "img",
		"image_key": imageKey,
	}
}

// NewEmoticonElement creates an emoticon/emoji element.
func NewEmoticonElement(emojiKey string) Element {
	return map[string]any{
		"tag":       "emotion",
		"emoji_key": emojiKey,
	}
}

// NewParagraph creates a paragraph from elements.
func NewParagraph(elements ...Element) Paragraph {
	return Paragraph(elements)
}

// NewPostContent creates a PostContent with title and paragraphs.
func NewPostContent(title string, paragraphs ...Paragraph) *PostContent {
	return &PostContent{
		Title:   title,
		Content: paragraphs,
	}
}

// NewPostLanguageContent creates post content for a specific language.
func NewPostLanguageContent(lang Language, content *PostContent) PostLanguageContent {
	return PostLanguageContent{
		Language: lang,
		Content:  *content,
	}
}

// NewPostMessage creates a new rich text (post) message.
//
// The language parameter specifies which language version to include.
// For multi-language support, use NewPostMessageMultiLanguage.
//
// Example:
//
//	paragraph := feishubot.NewParagraph(
//		feishubot.NewTextElement("Project update: "),
//		feishubot.NewLinkElement("View", "https://example.com"),
//	)
//	content := feishubot.NewPostContent("Title", paragraph)
//	message := feishubot.NewPostMessage(feishubot.LanguageZhCN, content)
func NewPostMessage(lang Language, content *PostContent) *Message {
	return &Message{
		MsgType: MsgTypePost,
		Content: map[string]any{
			"post": map[string]any{
				string(lang): map[string]any{
					"title":   content.Title,
					"content": convertParagraphs(content.Content),
				},
			},
		},
	}
}

// NewPostMessageMultiLanguage creates a rich text message with multiple language versions.
// At least one language must be provided.
//
// Example:
//
//	paragraph := feishubot.NewParagraph(feishubot.NewTextElement("Hello"))
//	content := feishubot.NewPostContent("Title", paragraph)
//	message := feishubot.NewPostMessageMultiLanguage(
//		feishubot.NewPostLanguageContent(feishubot.LanguageZhCN, content),
//		feishubot.NewPostLanguageContent(feishubot.LanguageEnUS, content),
//	)
func NewPostMessageMultiLanguage(langContents ...PostLanguageContent) *Message {
	postData := make(map[string]any)
	for _, lc := range langContents {
		postData[string(lc.Language)] = map[string]any{
			"title":   lc.Content.Title,
			"content": convertParagraphs(lc.Content.Content),
		}
	}

	return &Message{
		MsgType: MsgTypePost,
		Content: map[string]any{
			"post": postData,
		},
	}
}

// convertParagraphs converts Paragraph type to the expected JSON structure.
func convertParagraphs(paragraphs []Paragraph) [][]map[string]any {
	result := make([][]map[string]any, 0, len(paragraphs))
	for _, p := range paragraphs {
		paragraphArray := make([]map[string]any, 0, len(p))
		for _, e := range p {
			paragraphArray = append(paragraphArray, e)
		}
		result = append(result, paragraphArray)
	}
	return result
}

// NewTextMessage creates a new text message.
//
// The text can include @ mentions using HTML-style tags:
//   - @ single user: <at user_id="ou_xxx">Name</at>
//   - @ all: <at user_id="all">所有人</at>
//
// The user_id must be a valid Open ID or User ID of a group member.
func NewTextMessage(text string) *Message {
	return &Message{
		MsgType: MsgTypeText,
		Content: map[string]any{
			"text": text,
		},
	}
}

// NewImageMessage creates a new image message.
//
// The imageKey must be obtained from Feishu image upload API.
// See: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/image/create
func NewImageMessage(imageKey string) *Message {
	return &Message{
		MsgType: MsgTypeImage,
		Content: map[string]any{
			"image_key": imageKey,
		},
	}
}

// NewShareChatMessage creates a new share chat (group card) message.
//
// The bot can only share the group it belongs to.
// The shareChatID is the group ID (oc_xxxxxxxxx).
// See: https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/chat-id-description
func NewShareChatMessage(shareChatID string) *Message {
	return &Message{
		MsgType: MsgTypeShareChat,
		Content: map[string]any{
			"share_chat_id": shareChatID,
		},
	}
}

// Card represents an interactive card.
type Card struct {
	Schema string         `json:"schema"`
	Config map[string]any `json:"config,omitempty"`
	Body   *CardBody      `json:"body,omitempty"`
	Header *CardHeader    `json:"header,omitempty"`
}

// CardBody represents the body section of a card.
type CardBody struct {
	Direction string        `json:"direction,omitempty"`
	Padding   string        `json:"padding,omitempty"`
	Elements  []CardElement `json:"elements"`
}

// CardHeader represents the header section of a card.
type CardHeader struct {
	Title     *CardTitle `json:"title"`
	Subtitle  *CardTitle `json:"subtitle,omitempty"`
	Template  string     `json:"template,omitempty"`
	UiElement *CardTitle `json:"ui_element,omitempty"` // New API field
}

// CardTitle represents a title element (can be plain_text or lark_md).
type CardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// NewCardTitle creates a plain text title element.
func NewCardTitle(content string) *CardTitle {
	return &CardTitle{
		Tag:     "plain_text",
		Content: content,
	}
}

// NewCardMarkdownTitle creates a markdown title element.
func NewCardMarkdownTitle(content string) *CardTitle {
	return &CardTitle{
		Tag:     "lark_md",
		Content: content,
	}
}

// CardElement represents an element in the card body.
type CardElement map[string]any

// NewMarkdownElement creates a markdown text element.
func NewMarkdownElement(content string) CardElement {
	return map[string]any{
		"tag":     "markdown",
		"content": content,
	}
}

// NewDivElement creates a div element.
func NewDivElement(text *CardTitle) CardElement {
	return map[string]any{
		"tag":  "div",
		"text": text,
	}
}

// NewButtonElement creates a button element.
func NewButtonElement(text, buttonType string, url string) CardElement {
	return map[string]any{
		"tag": "button",
		"text": map[string]any{
			"tag":     "plain_text",
			"content": text,
		},
		"type": buttonType,
		"url":  url,
	}
}

// NewCard creates a Card with the given parameters.
func NewCard(schema string) *Card {
	return &Card{
		Schema: schema,
	}
}

// SetConfig sets the config for the card.
func (c *Card) SetConfig(config map[string]any) *Card {
	c.Config = config
	return c
}

// SetBody sets the body for the card.
func (c *Card) SetBody(body *CardBody) *Card {
	c.Body = body
	return c
}

// SetHeader sets the header for the card.
func (c *Card) SetHeader(header *CardHeader) *Card {
	c.Header = header
	return c
}

// ToMap converts the Card to a map for JSON serialization.
func (c *Card) ToMap() map[string]any {
	result := map[string]any{
		"schema": c.Schema,
	}

	if c.Config != nil {
		result["config"] = c.Config
	}
	if c.Body != nil {
		result["body"] = c.Body
	}
	if c.Header != nil {
		result["header"] = c.Header
	}

	return result
}

// NewInteractiveMessage creates a new interactive card message from a Card.
//
// Example:
//
//	card := feishubot.NewCard("2.0").
//		SetHeader(&feishubot.CardHeader{
//			Title:   feishubot.NewCardTitle("Card Title"),
//			Template: "blue",
//		}).
//		SetBody(&feishubot.CardBody{
//			Elements: []feishubot.CardElement{
//				feishubot.NewMarkdownElement("Hello!"),
//			},
//		})
//	message := feishubot.NewInteractiveMessage(card)
func NewInteractiveMessage(card *Card) *Message {
	return &Message{
		MsgType: MsgTypeInteractive,
		Card:    card.ToMap(),
	}
}

// NewInteractiveMessageFromMap creates a new interactive card message from a raw map.
// This is useful when you have the card structure from Feishu card builder tool.
//
// Use Feishu card builder tool to design cards and get the structure:
// https://open.feishu.cn/document/uAjLw4CM/ukzMukzMukzM/feishu-cards/feishu-card-cardkit/feishu-cardkit-overview
//
// Note: Interactive cards sent via custom bot only support URL navigation,
// not request callback interactions.
func NewInteractiveMessageFromMap(card map[string]any) *Message {
	return &Message{
		MsgType: MsgTypeInteractive,
		Card:    card,
	}
}
