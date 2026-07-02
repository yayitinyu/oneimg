package telegram

import "testing"

func TestValidateWebhookSecret(t *testing.T) {
	secret := WebhookSecret("123456:bot-token")
	if !ValidateWebhookSecret("123456:bot-token", secret) {
		t.Fatal("expected derived webhook secret to validate")
	}
	if ValidateWebhookSecret("123456:bot-token", "wrong") {
		t.Fatal("wrong webhook secret unexpectedly validated")
	}
	if ValidateWebhookSecret("", secret) || ValidateWebhookSecret("123456:bot-token", "") {
		t.Fatal("empty token or secret unexpectedly validated")
	}
}
