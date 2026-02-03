package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cium-cc/feishurobot"
)

func main() {
	// Get webhook URL from environment variable or use placeholder
	webhookURL := os.Getenv("FEISHU_WEBHOOK_URL")
	if webhookURL == "" {
		log.Println("Warning: FEISHU_WEBHOOK_URL not set, using placeholder")
		webhookURL = "https://open.feishu.cn/open-apis/bot/v2/hook/your_webhook_id"
	}

	// Optional: Get secret for signature verification
	secret := os.Getenv("FEISHU_SECRET")

	// Create client
	client := feishubot.NewClient(webhookURL, secret)

	// Example 1: Send text message
	if err := sendTextMessage(client); err != nil {
		log.Printf("Failed to send text message: %v", err)
	} else {
		log.Println("Text message sent successfully")
	}

	// Example 2: Send text message with @ mention
	if err := sendTextWithMention(client); err != nil {
		log.Printf("Failed to send text message with mention: %v", err)
	} else {
		log.Println("Text message with mention sent successfully")
	}

	// Example 3: Send rich text (post) message
	if err := sendPostMessage(client); err != nil {
		log.Printf("Failed to send post message: %v", err)
	} else {
		log.Println("Post message sent successfully")
	}

	// Example 4: Send rich text with link and @mention
	if err := sendPostMessageWithElements(client); err != nil {
		log.Printf("Failed to send post message with elements: %v", err)
	} else {
		log.Println("Post message with elements sent successfully")
	}

	// Example 5: Send interactive card message
	if err := sendInteractiveCard(client); err != nil {
		log.Printf("Failed to send interactive card: %v", err)
	} else {
		log.Println("Interactive card sent successfully")
	}

	// Example 6: Send interactive card with buttons
	if err := sendInteractiveCardWithButtons(client); err != nil {
		log.Printf("Failed to send interactive card with buttons: %v", err)
	} else {
		log.Println("Interactive card with buttons sent successfully")
	}
}

// sendTextMessage sends a simple text message.
func sendTextMessage(client *feishubot.Client) error {
	message := feishubot.NewTextMessage("Hello from Feishu Bot SDK!")
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}

// sendTextWithMention sends a text message with an @ mention.
func sendTextWithMention(client *feishubot.Client) error {
	// @ single user (replace ou_xxx with actual user ID)
	textWithMention := `<at user_id="ou_xxx">Tom</at> this is a notification for you.`

	// @ all users
	// textWithMention := `<at user_id="all">所有人</at> this is an announcement.`

	message := feishubot.NewTextMessage(textWithMention)
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}

// sendPostMessage sends a rich text (post) message.
func sendPostMessage(client *feishubot.Client) error {
	// Build content with a simple paragraph
	content := feishubot.NewPostContent(
		"Project Update",
		feishubot.NewParagraph(
			feishubot.NewTextElement("Project has been updated successfully!"),
		),
	)

	message := feishubot.NewPostMessage(feishubot.LanguageZhCN, content)
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}

// sendPostMessageWithElements sends a post message with links and @mentions.
func sendPostMessageWithElements(client *feishubot.Client) error {
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
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}

// sendInteractiveCard sends an interactive card message.
func sendInteractiveCard(client *feishubot.Client) error {
	// Build card with header and body
	card := feishubot.NewCard("2.0").
		SetHeader(&feishubot.CardHeader{
			Title:    feishubot.NewCardTitle("Welcome"),
			Template: "blue",
		}).
		SetBody(&feishubot.CardBody{
			Elements: []feishubot.CardElement{
				feishubot.NewMarkdownElement("Hello! This is an interactive card message from Feishu Bot SDK."),
			},
		})

	message := feishubot.NewInteractiveMessage(card)
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}

// sendInteractiveCardWithButtons sends an interactive card with buttons.
func sendInteractiveCardWithButtons(client *feishubot.Client) error {
	// Build a card with header, markdown body, and buttons
	card := feishubot.NewCard("2.0").
		SetConfig(map[string]any{
			"wide_screen_mode": true,
		}).
		SetHeader(&feishubot.CardHeader{
			Title:    feishubot.NewCardTitle("Task Alert"),
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
	resp, err := client.Send(context.Background(), message)
	if err != nil {
		return err
	}
	fmt.Printf("Response: Code=%d, Msg=%s\n", resp.Code, resp.Msg)
	return nil
}
