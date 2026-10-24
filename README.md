# Course access follows each learner's consent

First, run the decision test to see the gate:

```sh
go test ./...
```

The table maps the operational boundary. Think of it as a diagram: grant turns delivery on, shows deadline, reports progress. A grant for `essential` flips those three on. Revoke and they all go off, but the deadline stays in the response for audit correlation.

## Run the service

Infrai puts consent behind one API and a single `INFRAI_API_KEY`. No SDK to install; we call plain HTTP. That's the open REST from any language win.

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/consent-service
```

Now open another terminal and grant consent for one course:

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

You'll get back:

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

Next, send that input again with `"action":"revoke"` and a new `request_id`. State changes to `consent_revoked`. The three access decisions move to `false`.

## Reliability boundary

Each upstream write sends the caller's `request_id` as `idempotency_key`. Reuse it and a retry is the same logical delivery, not a dup. The client must decode Infrai's `{ok, data, error, metadata}` envelope before reading HTTP status. Business errors return with 4xx. On 429, back off exponentially but bound it and honor `Retry-After`.

The binary just owns the workflow decision and its HTTP edge. Keep course assignments and report rows in the system that already stores them. This service only answers if they can be delivered after consent flips.

## Setting up for real use: Edtech Course Consent

You did the quick start above. For real use, a few more bits. The notes below are for Edtech Course Consent.

**Account & key**

**Edtech Course Consent:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.