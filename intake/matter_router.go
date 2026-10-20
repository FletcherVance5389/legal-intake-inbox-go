package intake

import (
	"context"
	"fmt"
	"html"
)

type MatterSubmission struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	MatterType     string `json:"matter_type"`
	Message        string `json:"message"`
	CaptchaToken   string `json:"captcha_token"`
	WidgetRecordID string `json:"widget_record_id"`
	IP             string `json:"-"`
	RequestID      string `json:"request_id"`
}

type captchaResult struct {
	Valid bool `json:"valid"`
}

type emailResult struct {
	MessageID string `json:"message_id"`
}

type RoutedMatter struct {
	MessageID string `json:"message_id"`
	State     string `json:"state"`
}

// RouteMatter makes bot screening and inbox delivery one explicit decision.
func (c *Client) RouteMatter(ctx context.Context, teamInbox string, form MatterSubmission) (RoutedMatter, error) {
	var captcha captchaResult
	if err := c.call(ctx, httpMethodPost, "/v1/captcha/verify", map[string]any{
		"widget_record_id": form.WidgetRecordID, "token": form.CaptchaToken, "vendor": "turnstile", "ip": form.IP, "action": "legal_contact", "score_threshold": 0.7,
	}, "", &captcha); err != nil {
		return RoutedMatter{}, err
	}
	if !captcha.Valid {
		return RoutedMatter{}, &APIError{Code: "captcha rejected", Message: "captcha rejected", Status: 422}
	}

	message := fmt.Sprintf("<h2>New legal matter</h2><p><b>Name:</b> %s</p><p><b>Email:</b> %s</p><p><b>Matter:</b> %s</p><p>%s</p>",
		html.EscapeString(form.Name), html.EscapeString(form.Email), html.EscapeString(form.MatterType), html.EscapeString(form.Message))
	var sent emailResult
	if err := c.call(ctx, httpMethodPost, "/v1/email/send", map[string]any{
		"to": teamInbox, "subject": "Matter intake: " + form.MatterType, "html": message,
	}, form.RequestID, &sent); err != nil {
		return RoutedMatter{}, err
	}
	return RoutedMatter{MessageID: sent.MessageID, State: "delivered_to_team_inbox"}, nil
}

const httpMethodPost = "POST"
