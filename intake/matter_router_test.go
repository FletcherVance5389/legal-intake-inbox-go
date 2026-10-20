package intake

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteMatter(t *testing.T) {
	tests := []struct {
		name          string
		captchaReply  string
		wantErr       bool
		wantEmailCall bool
	}{
		{name: "accepted contact reaches the inbox", captchaReply: `{"ok":true,"data":{"valid":true}}`, wantEmailCall: true},
		{name: "rejected captcha does not send email", captchaReply: `{"ok":true,"data":{"valid":false}}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/captcha/verify":
					if r.Method != http.MethodPost {
						t.Fatalf("captcha method = %s", r.Method)
					}
					var payload map[string]any
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Fatal(err)
					}
					if payload["widget_record_id"] != "widget-42" {
						t.Fatalf("widget_record_id = %v", payload["widget_record_id"])
					}
					io.WriteString(w, tt.captchaReply)
				case "/v1/email/send":
					emailCalls++
					if r.Header.Get("Idempotency-Key") != "contact-42" {
						t.Fatal("missing idempotency key")
					}
					io.WriteString(w, `{"ok":true,"data":{"message_id":"msg-42"}}`)
				default:
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "test-key", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			got, err := client.RouteMatter(context.Background(), "intake@example.test", MatterSubmission{Name: "Ava", Email: "ava@example.test", MatterType: "signed document", Message: "Please review", CaptchaToken: "token", WidgetRecordID: "widget-42", RequestID: "contact-42"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v", err)
			}
			if tt.wantEmailCall && (got.MessageID != "msg-42" || got.State != "delivered_to_team_inbox") {
				t.Fatalf("result = %#v", got)
			}
			if (emailCalls > 0) != tt.wantEmailCall {
				t.Fatalf("email calls = %d", emailCalls)
			}
			if strings.Contains(got.State, " ") {
				t.Fatal("state must remain machine-readable")
			}
		})
	}
}
