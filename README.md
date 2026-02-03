# Feishu Custom Bot SDK

A Go SDK toolkit for Feishu (Lark) custom bot webhook API, supporting message sending with optional security verification (signatures).

## Features

- Send text messages with @mentions
- Send rich text (post) messages with flexible element builders
- Send image messages
- Send share chat (group card) messages
- Send interactive card messages with builder pattern
- Multi-language support for post messages
- Optional signature verification (HmacSHA256 + Base64)
- Context support for request cancellation
- Full test coverage

## Installation

```bash
go get github.com/cium-cc/feishurobot
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/cium-cc/feishurobot"
)

func main() {
    // Create client with webhook URL
    client := feishubot.NewClient(
        "https://open.feishu.cn/open-apis/bot/v2/hook/your_webhook_id",
        "", // No secret for now
    )

    // Send text message
    message := feishubot.NewTextMessage("Hello, World!")
    resp, err := client.Send(context.Background(), message)
    if err != nil {
        log.Fatalf("Failed to send message: %v", err)
    }

    log.Printf("Message sent! Code: %d, Msg: %s", resp.Code, resp.Msg)
}
```

### With Signature Verification

```go
// Create client with webhook URL and secret
client := feishubot.NewClient(
    "https://open.feishu.cn/open-apis/bot/v2/hook/your_webhook_id",
    "your_secret_here", // Enable signature verification
)

// Messages will automatically include timestamp and signature
message := feishubot.NewTextMessage("Secure message!")
resp, err := client.Send(context.Background(), message)
```

## Message Types

### Text Message

```go
// Simple text
message := feishubot.NewTextMessage("Hello, World!")

// @ single user (replace with actual user ID)
message := feishubot.NewTextMessage(`<at user_id="ou_xxx">Tom</at> notification for you.`)

// @ all users
message := feishubot.NewTextMessage(`<at user_id="all">所有人</at> announcement.`)
```

### Rich Text (Post) Message

#### Simple Post

```go
// Build content with a simple paragraph
content := feishubot.NewPostContent(
    "Project Update",
    feishubot.NewParagraph(
        feishubot.NewTextElement("Project has been updated successfully!"),
    ),
)

message := feishubot.NewPostMessage(feishubot.LanguageZhCN, content)
```

#### Post with Multiple Elements

```go
// Build content with multiple elements in a paragraph
content := feishubot.NewPostContent(
    "Daily Report",
    // First paragraph: text + link + @mention
    feishubot.NewParagraph(
        feishubot.NewTextElement("Project update: "),
        feishubot.NewLinkElement("View Details", "https://example.com/project/123"),
        feishubot.NewTextElement(" - Assigned to: "),
        feishubot.NewAtElement("ou_xxx", "Tom"),
    ),
    // Second paragraph: text only
    feishubot.NewParagraph(
        feishubot.NewTextElement("Please review the changes by end of day."),
    ),
)

message := feishubot.NewPostMessage(feishubot.LanguageZhCN, content)
```

#### Post with Multiple Paragraphs and Element Types

```go
// Available element types:
// - NewTextElement(text) - plain text
// - NewLinkElement(text, href) - hyperlink
// - NewAtElement(userID, userName) - @ mention
// - NewImageElement(imageKey) - inline image
// - NewEmoticonElement(emojiKey) - emoji

// Build complex post with multiple paragraphs
content := feishubot.NewPostContent(
    "Complex Report",
    // First paragraph: text + link + @mention + image
    feishubot.NewParagraph(
        feishubot.NewTextElement("Project update: "),
        feishubot.NewLinkElement("View", "https://example.com"),
        feishubot.NewTextElement(" - "),
        feishubot.NewAtElement("ou_xxx", "Tom"),
        feishubot.NewImageElement("img_key_123"),
    ),
    // Second paragraph: text only
    feishubot.NewParagraph(
        feishubot.NewTextElement("Please review by EOD."),
    ),
)

message := feishubot.NewPostMessage(feishubot.LanguageZhCN, content)
```

#### Multi-Language Post

```go
// Create content once, use for multiple languages
content := feishubot.NewPostContent(
    "Title",
    feishubot.NewParagraph(feishubot.NewTextElement("Content")),
)

message := feishubot.NewPostMessageMultiLanguage(
    feishubot.NewPostLanguageContent(feishubot.LanguageZhCN, content),
    feishubot.NewPostLanguageContent(feishubot.LanguageEnUS, content),
)
```

### Image Message

```go
message := feishubot.NewImageMessage("img_ecffc3b9-8f14-400f-a014-05eca1a4310g")
```

The imageKey must be obtained from Feishu image upload API.

### Share Chat (Group Card) Message

```go
message := feishubot.NewShareChatMessage("oc_f5b1a7eb27ae2****339ff")
```

The bot can only share the group it belongs to.

### Interactive Card Message

#### Simple Card

```go
// Build card with header and body
card := feishubot.NewCard("2.0").
    SetHeader(&feishubot.CardHeader{
        Title:   feishubot.NewCardTitle("Welcome"),
        Template: "blue",
    }).
    SetBody(&feishubot.CardBody{
        Elements: []feishubot.CardElement{
            feishubot.NewMarkdownElement("Hello! This is an interactive card message."),
        },
    })

message := feishubot.NewInteractiveMessage(card)
```

#### Card with Markdown and Buttons

```go
// Available card element types:
// - NewMarkdownElement(content) - markdown text
// - NewDivElement(text) - div with text
// - NewButtonElement(text, type, url) - button

// Build a card with markdown body and buttons
card := feishubot.NewCard("2.0").
    SetConfig(map[string]any{
        "wide_screen_mode": true,
    }).
    SetHeader(&feishubot.CardHeader{
        Title:   feishubot.NewCardTitle("Task Alert"),
        Template: "red",
    }).
    SetBody(&feishubot.CardBody{
        Direction: "vertical",
        Padding:   "12px 12px 12px 12px",
        Elements: []feishubot.CardElement{
            feishubot.NewMarkdownElement("**High Priority Task**\n\nPlease complete this task by end of day."),
            feishubot.NewButtonElement("View Details", "primary", "https://example.com/task/123"),
            feishubot.NewButtonElement("Dismiss", "default", "https://example.com/dismiss"),
        },
    })

message := feishubot.NewInteractiveMessage(card)
```

#### Card from Map (for Card Builder Tool)

```go
// Use card from Feishu card builder tool
cardMap := map[string]any{
    "schema": "2.0",
    "body": map[string]any{
        "elements": []map[string]any{
            {"tag": "markdown", "content": "Your card content here..."},
        },
    },
}

message := feishubot.NewInteractiveMessageFromMap(cardMap)
```

## API Reference

### Client

```go
type Client struct {
    WebhookURL string
    Secret     string    // Optional secret for signature verification
    HTTPClient HTTPClient // HTTP client interface
}
```

#### NewClient

```go
func NewClient(webhookURL string, secret string) *Client
```

Creates a new Feishu bot client.

- `webhookURL`: The full webhook URL for your custom bot
- `secret`: Optional secret for signature verification. If empty, no signature will be sent.

The default HTTP client has a 30 second timeout. For custom timeout settings, use SetHTTPClient after creating client.

#### Send

```go
func (c *Client) Send(ctx context.Context, msg *Message) (*Response, error)
```

Sends a message to Feishu webhook.

- `ctx`: Context for request, can be used for cancellation
- `msg`: The message to send

Returns:
- `*Response`: The API response
- `error`: An error if request fails or returns a non-zero code

#### SetHTTPClient

```go
func (c *Client) SetHTTPClient(client HTTPClient)
```

Sets a custom HTTP client for the bot client. This is useful for testing or for custom timeout configurations.

### Message Types

```go
type MsgType string

const (
    MsgTypeText        MsgType = "text"
    MsgTypePost        MsgType = "post"
    MsgTypeImage       MsgType = "image"
    MsgTypeShareChat   MsgType = "share_chat"
    MsgTypeInteractive MsgType = "interactive"
)
```

#### Language Constants

```go
type Language string

const (
    LanguageZhCN Language = "zh_cn"   // Simplified Chinese
    LanguageEnUS Language = "en_us"   // English
    LanguageJa    Language = "ja"       // Japanese
)
```

### Message Constructors

#### Text

```go
func NewTextMessage(text string) *Message
```

#### Post (Rich Text)

```go
type Paragraph []Element
type Element map[string]any

func NewPostContent(title string, paragraphs ...Paragraph) *PostContent
func NewPostMessage(lang Language, content *PostContent) *Message
func NewPostMessageMultiLanguage(langContents ...PostLanguageContent) *Message
```

#### Post Element Builders

```go
func NewTextElement(text string) Element
func NewLinkElement(text, href string) Element
func NewAtElement(userID, userName string) Element
func NewImageElement(imageKey string) Element
func NewEmoticonElement(emojiKey string) Element
func NewParagraph(elements ...Element) Paragraph
```

#### Image

```go
func NewImageMessage(imageKey string) *Message
```

#### Share Chat

```go
func NewShareChatMessage(shareChatID string) *Message
```

#### Interactive Card

```go
func NewInteractiveMessage(card *Card) *Message
func NewInteractiveMessageFromMap(card map[string]any) *Message
```

#### Card Builder

```go
type Card struct {
    Schema string
    Config map[string]any
    Body   *CardBody
    Header *CardHeader
}

type CardBody struct {
    Direction string
    Padding   string
    Elements  []CardElement
}

type CardHeader struct {
    Title    *CardTitle
    Subtitle *CardTitle
    Template string
    UiElement *CardTitle // New API field
}

type CardTitle struct {
    Tag     string
    Content string
}

func NewCard(schema string) *Card
func (c *Card) SetConfig(config map[string]any) *Card
func (c *Card) SetBody(body *CardBody) *Card
func (c *Card) SetHeader(header *CardHeader) *Card

func NewCardTitle(content string) *CardTitle
func NewCardMarkdownTitle(content string) *CardTitle

func NewMarkdownElement(content string) CardElement
func NewDivElement(text *CardTitle) CardElement
func NewButtonElement(text, buttonType string, url string) CardElement
```

### Response

```go
type Response struct {
    Code         int         `json:"code"`          // Response code, 0 means success
    Msg          string      `json:"msg"`           // Response message
    Data         interface{} `json:"data"`         // Response data
    StatusCode   int         `json:"StatusCode,omitempty"`   // Deprecated
    StatusMessage string      `json:"StatusMessage,omitempty"` // Deprecated
}
```

## Rate Limits

Feishu custom bots have the following rate limits:
- 100 requests per minute per bot
- 5 requests per second per bot

Please avoid sending messages at times like 10:00, 17:30, etc. to avoid rate limiting errors.

## Request Size Limit

The request body size must not exceed 20 KB.

## Error Codes

| Code  | Message                  | Description |
|-------|-------------------------|-------------|
| 0     | success                 | Success |
| 9499  | Bad Request             | Invalid request format |
| 19022 | Ip Not Allowed           | IP not in whitelist |
| 19024 | Key Words Not Found     | Keywords not found in message |
| 19021 | sign match fail or timestamp is not within one hour from current time | Signature verification failed |

## Development

### Running Tests

```bash
go test ./...
```

### Running Example

```bash
export FEISHU_WEBHOOK_URL="your_webhook_url"
export FEISHU_SECRET="your_secret"  # Optional
go run cmd/example/main.go
```

## Examples

See `cmd/example/main.go` for comprehensive examples of all message types.

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## References

- [Feishu Custom Bot Documentation](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/bot-v3/add-custom-bot)
- [Feishu Card Builder](https://open.feishu.cn/document/uAjLw4CM/ukzMukzMukzM/feishu-cards/feishu-card-cardkit/feishu-cardkit-overview)
