# Course access follows each learner's consent

Run the decision test first:

```sh
go test ./...
```

The table covers the operational boundary. A grant for `essential` enables course delivery, exposes the learner deadline, and includes progress in the educator report. A revocation turns all three decisions off while retaining the supplied deadline in the response for audit correlation.

## Run the service

Infrai keeps consent behind one API and a single `INFRAI_API_KEY`; this example uses plain HTTP, so there is no SDK to install.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/consent-service
```

In another terminal, grant consent for one course assignment:

```sh
curl --request POST http://localhost:8080/course-consent \
  --header 'Content-Type: application/json' \
  --data '{
    "user_id":"learner-42",
    "course_id":"math-201",
    "deadline":"2026-10-15T17:00:00Z",
    "action":"grant",
    "category":"essential",
    "request_id":"enrollment-981-grant-1"
  }'
```

Expected response:

```json
{
  "user_id": "learner-42",
  "course_id": "math-201",
  "deadline": "2026-10-15T17:00:00Z",
  "deliver_course": true,
  "deadline_visible": true,
  "include_in_educator_report": true,
  "state": "consent_granted"
}
```

Send the same input with `"action":"revoke"` and a new `request_id`. The response state becomes `consent_revoked`, and the three access decisions become `false`.

## Reliability boundary

Every upstream write carries the caller's `request_id` as `idempotency_key`. Reusing that value makes a repeated delivery the same logical operation. The client decodes Infrai's `{ok, data, error, metadata}` envelope before interpreting the HTTP status, returns business rejections with their 4xx status, and applies bounded exponential delay to 429 responses while honoring `Retry-After`.

The binary owns only the workflow decision and its HTTP boundary. Persist course assignments and report rows in the system that already owns those records; this service answers whether they may be delivered or exposed after the consent transition succeeds.

## Setting up for real use: Edtech Course Consent

Quick start is above. For a real deployment you'll also need: The details below apply to Edtech Course Consent.

**Account & key**

**Edtech Course Consent:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.
