# Dalista API — Agent Instructions

Golang backend for Dalista, a shopping-list / shop-catalog sharing app with item comparison and an AI agent. This file is the standing context for any Claude Code session working in this repo — read it before starting any ticket.

## Reference project — copy its conventions exactly

`/Users/mina/Documents/williamsburg/lumina/lumina-be` is a real, in-production Golang API on the same stack. Dalista's backend follows its conventions file-for-file, not just in spirit. When in doubt about how something should be structured, open the equivalent file in Lumina first.

Stack: Gin (`gin-gonic/gin`), GORM + `pgx` Postgres driver, Goose SQL migrations (never GORM auto-migrate), `golang-jwt/jwt/v5`, `go-playground/validator` via `ShouldBindJSON`, `caarlos0/env` + `godotenv` for config, Sentry for error tracking, `swaggo` for OpenAPI docs, Go `testing` + `testify` against a real Postgres CI service container.

## Layering (every feature follows this)

```
main.go                              # entrypoint: server.Initialize() + server.Run()
config/                              # env-driven config, single Config struct, GetConfig()
api/
├── forms/<feature>Forms.go          # request DTOs + validation tags + ToModel()
├── resources/<feature>Resource.go   # response DTOs + FromModel()
├── responses/responses.go           # shared error/response envelopes
└── server/
    ├── handlers/<feature>Handlers/  # bind form → call service → map to resource → JSON
    ├── middlewares/                 # auth, permission (AbortIfCant), rate limit, CORS
    ├── routes/api.go                # route groups per feature, middleware attached per group
    ├── logger/                      # Sentry-backed structured logging
    └── initializer.go / runner.go
app/
├── repository/
│   ├── database.go                 # GORM/Postgres connection singleton, GetDB()/SetDB()
│   ├── migrations/*.sql             # Goose, timestamp-prefixed, additive only, -- +goose Up/Down
│   └── models/<feature>Models/*.go  # one GORM struct per entity, explicit TableName()
├── services/<feature>Services/
│   ├── <entity>Service.go           # business logic, DB via repository.GetDB()
│   └── permissions.go               # permission string constants + AllPermissions slice
└── utils/                            # hashing, JWT, encryption, validation helpers
integrations/                         # third-party clients
cmd/                                  # one-off executables (createadmin, syncpermissions, ...)
test/<feature>Tests/{handlers,services,resources}/
```

Handler pattern, exactly as in Lumina: (1) `middlewares.AbortIfCant(c, services.PermX)`, (2) bind+validate into a `forms.XForm`, (3) `form.ToModel()`, (4) call the feature's service package, (5) map result via `resources.XFromModel(...)`, (6) return JSON. Every handler gets a Swagger doc comment — the OpenAPI spec this generates is what the mobile app and dashboard integrate against, so never skip it.

Permissions: each feature's `services/<feature>Services/permissions.go` defines its own permission strings (e.g. `lists:create`, `lists:share`, `public_lists:star`, `admin:send_sms`) plus an `AllPermissions` slice. A top-level `permissionRegistry.go` aggregates all of them and defines `RoleDefaults()`. Use Dalista's own permission strings — do not reuse Lumina's literal ones.

Models: UUID primary keys (`google/uuid`), explicit `gorm` tags, explicit `TableName()`, plain `time.Time` timestamp fields. One file per entity, grouped into `<feature>Models` packages.

Migrations: hand-written SQL via Goose only. One file per logical schema change, timestamp-prefixed, never edited after merge. Postgres native `ENUM` types created idempotently (`DO $$ ... EXCEPTION WHEN duplicate_object THEN NULL; END $$`).

Tests live under `test/<feature>Tests/{handlers,services,resources}/`, mirroring the production package layout — not beside the source files. CI spins up a real Postgres service container, runs Goose migrations, then `go test ./test/...`.

## Module → package mapping

Each module from the module breakdown doc becomes one `<feature>Models` + `<feature>Services` + `<feature>Handlers` trio:

- `userModels/`, `authServices/`, `authHandlers/` — Auth & Identity, Users & Profiles
- `listModels/` (GroupedList, List, ListVariant, Category), `itemModels/`, `listServices/`, `listHandlers/` — Lists, Grouped Lists & Categories; Periodic Lists & Today's List
- `sharingModels/` (ListShare, PendingInvite), `sharingServices/`, `sharingHandlers/` — Sharing & Invites
- `publicListModels/` (PublicListStar), `publicListServices/`, `publicListHandlers/` — Public Lists & Discovery
- `comparisonModels/` (ComparisonSession), `comparisonServices/`, `comparisonHandlers/` — Item Comparison
- `activityModels/` (Comment, ActivityEvent), `activityServices/`, `activityHandlers/` — Activity, Comments & Timeline
- `agentModels/` (AgentActionLog), `agentServices/`, `agentHandlers/` — AI Agent
- `billingModels/` (Subscription [provider-agnostic], AppleIAPReceipt, GooglePlayPurchase, StorageUsage), `billingServices/`, `billingHandlers/` + `appleIAPHandlers/` + `playBillingHandlers/` — Billing & Subscriptions
- `notificationServices/`, `notificationHandlers/` — Notifications (push + admin SMS)

`integrations/` additions not present in Lumina: `fcmSender.go` (Firebase Cloud Messaging — default notification channel), `smsMisrSender.go` (SMS/OTP, used sparingly — OTP, admin-initiated SMS, unregistered-invite reach), `appstore/` (Apple App Store Server API + Server Notifications V2), `playbilling/` (Play Developer API + Real-time Developer Notifications), `oauth/` (Google + Facebook token verification), `smsOtpAuth/` (phone-as-sign-in-method, OTP issuance + verification).

**No `stripe/` package.** v1 billing is Apple IAP + Google Play Billing only. The `Subscription.billing_provider` enum has `apple_iap` and `google_play` populated and `stripe` reserved for a future value — don't build Stripe integration, just leave the enum room for it.

**No email integration.** Notifications are push (FCM) by default and SMS (SMS Misr) only where necessary. No Mailgun, no email-dependent flows.

## Auth model

Four sign-in methods, all first-class: Google OAuth, Facebook OAuth, email/password, and **phone number + OTP as its own standalone method** (no email required on that path). `User.email` is optional; `User.phone` is required and unique. Every account needs a verified phone number eventually (required for invite-by-phone auto-attach), regardless of which method was used to sign up — for social/email signups this is a post-signup verification step, not blocking initial signup.

## Billing

Native app store billing only in v1 — Apple IAP (StoreKit) on iOS, Google Play Billing on Android. Apple App Store Server Notifications V2 payloads are signed with JWS, not a shared secret like Stripe — verify accordingly. Google notifications arrive via Pub/Sub (Real-time Developer Notifications). The client never gets trusted on its own purchase confirmation — raw receipt/purchase token goes to the backend, which validates server-side before activating a tier.

## Known open items (don't block on these, but don't silently resolve them either — flag and ask)

1. Real-time sync mechanism for collaborative lists (websockets/SSE/polling) — not yet decided.
2. AI agent LLM provider + invocation model (sync vs. background job) — not yet decided.
3. Exact notification event list (which actions trigger push vs. SMS) — not yet finalized.
4. Dashboard metric definitions ("interested in a category," usage metrics, time granularity) — not yet finalized.

## Where the full spec lives

The complete requirements digest, module breakdown, technical architecture doc, and phased build plan live in the "Claude outputs" folder of the Dalista project, and the full ticket backlog is in Linear (team: Dalista, projects: API / Mobile App / Dashboard, labeled Phase 0 through Phase 12). When a ticket's one-line description isn't enough context, check those docs before guessing.
