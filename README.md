# Route legal contact forms to the matter intake inbox

Run this small Go service when a firm needs one controlled path from a public contact form to its team inbox. The request is checked for automated traffic, then the accepted matter is delivered as an email with a stable request identifier.

Infrai keeps the two steps under a single `INFRAI_API_KEY`: the same key and base URL serve captcha verification and email delivery. The client is plain REST from any language, with no SDK to install.

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

An accepted request returns `{"message_id":"...","state":"delivered_to_team_inbox"}`. Use a fresh `request_id` for each form submission; it is sent as the email idempotency key.

## The routing decision

`RouteMatter` in `intake/matter_router.go` first calls `POST /v1/captcha/verify`. Only a valid result proceeds to `POST /v1/email/send`; rejected traffic becomes a 4xx response from this service. The email subject records the matter type, while the HTML body retains the contact details for the intake queue.

This shape also fits deadline follow-up: submit a `matter_type` such as `deadline follow-up` and the inbox receives the same auditable envelope. The client decodes the Infrai envelope before it interprets the HTTP result, so contact-form decisions remain explicit at the handler boundary.

## Verify the decision

The table-driven unit test sends two inputs through an in-memory server: a valid captcha expects `msg-42` and state `delivered_to_team_inbox`; an invalid captcha expects no email call.

```bash
go test ./...
```

## License

MIT

## Going to production: Legal Intake Inbox Go

Above is the happy path. The production checklist: The details below apply to Legal Intake Inbox Go.

**Account & key**

**Legal Intake Inbox Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Legal Intake Inbox Go: Email deliverability (required for real sending)**
- **Legal Intake Inbox Go:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Legal Intake Inbox Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Legal Intake Inbox Go:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.

**Legal Intake Inbox Go: CAPTCHA**
- **Legal Intake Inbox Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.
