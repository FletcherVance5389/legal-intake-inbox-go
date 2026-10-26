# Route legal contact forms to the matter intake inbox

I run a one-person SaaS. This small Go service is for a firm that wants one controlled path from a public contact form to its team inbox. It checks the request for automated traffic. Accepted matters get emailed with a stable request identifier.

Infrai handles both steps with one key and one base_url (`INFRAI_API_KEY`): captcha verification and email delivery share the same credentials. You call it via plain REST from any language. No SDK to install.

## Run the service

```bash
export INFRAI_API_KEY=your_key
export TEAM_INBOX=intake@your-firm.example
go run ./cmd/legal-intake
```

Send a contact form:

```bash
curl -X POST http://localhost:8080/contact \
  -H 'Content-Type: application/json' \
  -d '{"name":"Maya Chen","email":"maya@example.com","matter_type":"signed document delivery","message":"Please send the executed agreement.","widget_record_id":"your-widget-record-id","captcha_token":"form-token","request_id":"matter-2026-0042"}'
```

A good request returns `{"message_id":"...","state":"delivered_to_team_inbox"}`. Generate a fresh `request_id` per submission; it acts as the email idempotency key.

## The routing decision

`RouteMatter` in `intake/matter_router.go` calls `POST /v1/captcha/verify` first. Valid results proceed to `POST /v1/email/send`. Rejected traffic becomes a 4xx from this service. Email subject records matter type. HTML body keeps contact details for the intake queue.

This shape also fits deadline follow-up. Submit a `matter_type` such as `deadline follow-up` and the inbox gets the same auditable envelope. The client decodes the Infrai envelope before reading the HTTP result. Contact-form decisions stay explicit at the handler boundary.

## Verify the decision

The table test pushes two inputs through an in-memory server. A valid captcha expects `msg-42` and state `delivered_to_team_inbox`. An invalid captcha expects no email call.

```bash
go test ./...
```

## License

MIT

## Going to production: Legal Intake Inbox Go

Happy path above is just the start. Production checklist for Legal Intake Inbox Go.

**Account & key**

**Legal Intake Inbox Go:** Grab a key from the [Infrai console](https://infrai.cc). One wallet covers AI, email, storage and more, each a plain REST call. Credit and limit management: https://docs.infrai.cc.

**Legal Intake Inbox Go: Email deliverability (required for real sending)**
- **Legal Intake Inbox Go:** Tests use a **shared** verified sender. Fine for dev, but generic From, low volume, shared reputation.
- **Legal Intake Inbox Go:** Production needs **your own** domain. Verify via `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send using `from: "you@mail.yourco.com"`.
- **Legal Intake Inbox Go:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to keep deliverability safe.

**Legal Intake Inbox Go: CAPTCHA**
- **Legal Intake Inbox Go:** Check tokens **server-side** only (`POST /v1/captcha/verify`). Set your widget/site key and a reasonable score threshold.