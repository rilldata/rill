# Rill Onboarding & PLG Audit — Improvement Ideas
**Role:** Senior product growth / UX / onboarding consultant review
**Date:** 2026-07-19
**Grounding:** Journey reconstructed from the codebase (`web-local` welcome/deploy routes, `web-common` add-data/welcome/onboarding/sample-data features, `cli/cmd/start` + `cli/pkg/local`, `web-admin` routes and billing features, `admin` trial/billing jobs) and from the verified code review in `audit/plg-first-user-experience.md`. Finding IDs like `plg-first-run-01` refer to that document.

**Activation definition used throughout:**
- **Local activation (aha):** first interactive dashboard viewed on data the user recognizes (own data or a chosen sample), within the first session.
- **Cloud activation (habit/monetizable):** project deployed to Rill Cloud **and** at least one other person views a dashboard in it.

**Assumptions & missing information (read first):**
- I had no access to funnel analytics, session recordings, or support tickets. Where I estimate abandonment risk, it is inferred from the code paths a user must survive, not from measured drop-off. The instrumentation gaps below (§6) are themselves a finding: today the funnel *cannot* be measured end-to-end — `AppStart` telemetry is silently dropped for any project with a connector error (`plg-first-run-07`), and there are no events between "app start" and "deploy success".
- The pre-product journey (website → docs → `curl | sh` install) was out of scope; I start at `rill start`.
- Embedded example project *content* quality could not be verified (bundled at build time); the ClickHouse GitHub-analytics example may require live network access on first run — flagged as risk, not fact.
- I assume most first-time users arrive in one of three segments: **(A) data engineer with data at hand** (a CSV/Parquet file or warehouse creds), **(B) evaluator with no data ready** (wants to see the product work in one click), **(C) invited viewer** (a teammate shared a dashboard). The current journey serves A best, B worst, and C adequately.

---

## 1. Executive summary

Rill's onboarding has an unusually strong core: `rill start` gets a first-time user from an empty directory to a real, interactive dashboard in one tool, the add-data wizard is a clean guarded state machine, AI generation degrades to deterministic generators, and the local → deploy → invite → cloud handoff exists end-to-end. The **happy path is genuinely good** — arguably best-in-class for a code-first BI tool. The three most important weaknesses are:

1. **The unhappy paths fail silently, exactly where new users live.** The splash "Generate sample data" CTA can leave a user staring at a spinner for ~30s and then strand them on an empty project with no error (`plg-first-run-01`); "Generate dashboard with AI" can spin forever on a backend error (`plg-first-run-02`); example-unpack failures are swallowed by a bare `catch` (`plg-first-run-04`); the `/deploy?org=` deep link 404s (`plg-first-run-03`). A first-time user cannot distinguish "broken product" from "my mistake" — silent failure at minute two is churn, not friction.

2. **The cloud-first entry is a dead end, and the funnel ends at deploy.** A brand-new user who signs up at ui.rilldata.com with no org sees a one-line message pointing at docs (`web-admin/src/features/home/WelcomeMessage.svelte`) — no in-product way to create an org, no demo to explore, no path back to value. And after a successful deploy, the journey stops: the invite screen is skippable with no follow-up, there is no share loop, no lifecycle messaging tied to activation state, and trial expiry is a hibernation cliff.

3. **The journey is unmeasured and un-personalized.** There are no events between app-start and deploy-success, `AppStart` itself is dropped for misconfigured projects, and every user gets the same splash regardless of whether they arrived with data, without data, or via an invite. You cannot optimize a funnel you cannot see, and you cannot shorten a journey you don't segment.

Strongest elements to protect: the three-path splash concept (example / your data / generated data), the add-data wizard's guarded step machine, AI-with-deterministic-fallback generation, and the auto-generated first dashboard. Nothing below recommends removing these; the work is making them survive failure, measuring them, and extending the journey past deploy.

---

## 2. Journey review

Each step: **Goal** (user's), **Questions** (user's likely concerns), **Works**, **Friction**, **Abandonment risk** (L/M/H with reason), **Improve**.

### Step 0 — Install & `rill start` (CLI)
- **Goal:** "Get this running and see if it's worth my time."
- **Questions:** Do I need an account? Will this touch my data? How long will this take?
- **Works:** Single binary, no signup wall, `rill start` in an empty directory goes straight to the browser splash. No-login-required is a real differentiator — keep it sacred.
- **Friction:** Terminal output at first run doesn't set expectations for what's about to open or what "good" looks like. If any connector fails to analyze, the `AppStart` activation event is silently dropped (`plg-first-run-07`, `cli/pkg/local/app.go:544`) — invisible to the user but blinds the funnel.
- **Abandonment risk:** Low (users who got this far are motivated).
- **Improve:** Fix `emitStartEvent`; print a one-line orientation in the terminal ("Opening Rill at localhost:9009 — pick an example or connect your data").

### Step 1 — Welcome splash (`web-local/src/routes/(misc)/welcome/+page.svelte`)
- **Goal:** "Show me what this is, fast."
- **Questions:** Which of these three options is right for me? What happens if I click one — is it reversible?
- **Works:** The three-path structure (examples / connect your data / generate sample data) maps exactly onto the three arrival segments. Quick-pick connector shortcuts reduce a step for segment A.
- **Friction:** No value proposition or expectation-setting copy on the splash itself; three equal-weight choices with no guidance on which fits whom. The quick-pick grid silently disappears if the connectors query fails (`plg-first-run-11`, explicit `TODO: add error state`). Example cards are a hardcoded frontend list that can drift from the backend embed (`plg-first-run-10`).
- **Abandonment risk:** Medium — a choice screen with no recommendation invites hesitation, and segment B (no data) has two options that both look risky.
- **Improve:** Add one sentence of orientation + a "recommended for you" nudge (see §3 idea 13); add the missing error state; drive cards from `ListExamples`.

### Step 2a — Example project path (`ProjectCards.svelte`)
- **Goal:** "One click to a working dashboard I can poke at."
- **Questions:** Is this real data? Can I see how it's built?
- **Works:** Conceptually the fastest aha for segment B; examples unpack into real, inspectable project files — a perfect "show the code-first model" moment.
- **Friction:** Every local selection pays a **fixed 5-second `setTimeout` then a full-page hard reload** (`plg-first-run-05`, self-described "legacy fix… TODO"). Unpack failures are swallowed with zero feedback (`plg-first-run-04`); a mid-loop write failure leaves a broken half-project (`plg-first-run-13`). Examples that need live network/credentials aren't labeled.
- **Abandonment risk:** Medium — 5 guaranteed dead seconds at the moment of highest curiosity; a silent failure here is fatal ("I clicked and nothing happened").
- **Improve:** Reconciliation-driven redirect (the pattern already exists in `OnboardingWorkspace`'s `waitUntil(fileArtifacts.hasFileArtifact(...))`); surface unpack errors with retry; make unpack atomic; badge examples needing network.

### Step 2b — "Connect your data" wizard (`AddDataStateManager`, `add-data/manager/steps/*`)
- **Goal:** "Get MY data in and see something."
- **Questions:** Are my credentials safe/local? Did the connection actually work? What are all these files it's creating?
- **Works:** Clean guarded step machine; import steps **await** generation so hard errors surface (`GenerateDashboardStatus` has try/catch and an always-enabled "Skip and view project" escape); AI generation falls back to deterministic simple generators.
- **Friction:** Credentials are written to YAML and only validated by attempting reconcile — no upfront "test connection" or table preview, so a typo becomes a post-import reconcile error. The connector-validate wait can hang forever at a stateVersion digit boundary (string compare `"10" > "9"` is false, `plg-first-run-08`) with no timeout. Default flow generates **four resources per source** — model + metrics view + explore + canvas, two near-identically-named dashboards (`plg-first-run-06`) — doubling the conceptual surface at the exact moment the user is trying to form a mental model. An all-numeric table yields a dimension-less KPI-only "dashboard" with no explanation (`plg-first-run-12`).
- **Abandonment risk:** High for warehouse connectors (credential friction × no pre-validation × possible hang); Low-Medium for local files.
- **Improve:** Test-connection + table preview before writing YAML; numeric version compare + overall timeout; default to a single explore dashboard; hint when no dimensions/time column detected.

### Step 2c — "Generate sample data" (`OnboardingGenerateSampleData.svelte`, `generate-sample-data.ts`)
- **Goal:** "I have no data handy — just show me the product."
- **Questions:** What should I type? How long will this take? Is it working?
- **Works:** The idea is exactly right for segment B: describe a domain, get a dataset and dashboard.
- **Friction:** The entire flow drives through the developer chat agent with **no fallback and unbounded waits**. Flag off → ~30s spinner, then stranded on an empty project with no error. AI unreachable (not logged in / offline) → chat error, empty project. The splash card is ungated while its home-screen twin is gated behind `$developerChat` (`plg-first-run-01`, HIGH). This is the single worst gap: the CTA aimed at the most fragile segment is the least robust.
- **Abandonment risk:** **High** — a silent dead-end on a primary splash CTA is a churned evaluator.
- **Improve:** Gate the CTA identically to its twin; add a deterministic offline generator fallback (DuckDB `range()`/`random()` templates); bounded waits with visible error + retry.

### Step 3 — First dashboard & workspace (post-generation)
- **Goal:** "Confirm this shows my data correctly; understand what I'm looking at."
- **Questions:** Which of these generated files matters? What do I do next? Are these numbers right?
- **Works:** Landing on a live, filterable explore dashboard is the aha; the file tree makes the code-first model tangible.
- **Friction:** No orientation after landing — no "here's what we generated and why", no suggested next action, no checklist. The four generated resources (2b) amplify the "what am I looking at" problem. Nothing marks the moment as progress ("you just built a dashboard").
- **Abandonment risk:** Medium — users who reach here mostly stay for the session, but without a next-step scaffold the session ends without deploy intent forming.
- **Improve:** Lightweight persistent checklist (Import → Dashboard ✓ → Customize → Deploy → Share); an AI "explain this dashboard" one-shot narrative; a visible Deploy nudge once a dashboard reconciles healthy.

### Step 4 — Deploy entry & login (`web-local .../deploy/+page.svelte`; `rill deploy` CLI)
- **Goal:** "Put this somewhere my team can see."
- **Questions:** What does deploying cost? Where does my data go? Do I need GitHub?
- **Works:** The dispatcher handles login/org/project branching sensibly; CLI mirrors it; telemetry exists here (`LoginStart/LoginSuccess/DeployIntent/DeploySuccess`).
- **Friction:** ~2s of queries behind a bare spinner; login round-trips through the browser to Auth0 with no explanation of what's about to happen or what it costs (trial terms appear only after the fact as a banner). The `/deploy?org=` deep link 307s to a nonexistent `fresh-deploy` route → 404 (`plg-first-run-03`) — any docs/campaign link using it is a broken funnel entry.
- **Abandonment risk:** Medium — cost/GitHub uncertainty at the exact moment of commitment; the 404 is total loss for that entry point.
- **Improve:** Fix the redirect; add a pre-login interstitial: "Free trial, no card required, your project uploads to Rill's managed infra — [Deploy] [What happens when I deploy?]".

### Step 5 — Org create/select (`deploy/organization/*`, `CreateNewOrgForm.svelte`)
- **Goal:** "Get past this form."
- **Questions:** What's an organization? Can I rename it later? Personal name or company name?
- **Works:** Auto-slugify from display name; friendly duplicate-name error mapping; sensible validation.
- **Friction:** A naming decision with unexplained permanence-feel, sitting between the user and their deploy. The mandatory field is the URL *name* while display name is optional — inverted from what a new user expects.
- **Abandonment risk:** Low — but it's pure cognitive load that could be zero.
- **Improve:** Default the org from the user's email domain or name ("acme" from jane@acme.com), one editable field, "you can rename later" microcopy. Ideally skip the screen entirely on first deploy (auto-create, rename later).

### Step 6 — Project create & upload (`deploy/project/create/+page.svelte`; GitHub variant)
- **Goal:** "Just ship it."
- **Questions:** (managed) nothing — it auto-fires; (GitHub) why do I need to install an app?
- **Works:** Managed upload is the default and auto-fires with no form — excellent. `DeployError.svelte` handles quota (→ upgrade CTA) and GitHub-no-access cases.
- **Friction:** The GitHub path detours through app-install authorization mid-flow; the interactive CLI asks "Enable automatic deploys from GitHub?" defaulting No — good default, but the question itself is premature for a first deploy.
- **Abandonment risk:** Low (managed) / Medium (GitHub detour).
- **Improve:** Keep managed as the untouched default; move the GitHub question post-first-deploy ("Connect GitHub for CI deploys" as a settings/checklist item).

### Step 7 — Invite screen (`[project]/-/invite/+page.svelte`)
- **Goal:** Mixed — many users' real goal right now is "see my deployed project", not "invite people".
- **Questions:** Can I do this later? What will invitees see?
- **Works:** Right idea at the right place in principle; email invites + copy-link + domain-access switch cover the mechanisms; "Skip for now" exists.
- **Friction:** It interrupts before the user has *seen* their deployed project — asking for referral before re-experiencing value. One-shot: skipping it has no follow-up anywhere (no checklist item, no email nudge), so the collaboration step — the retention driver — silently evaporates for skippers.
- **Abandonment risk:** Low for the session, **High for the share loop** (skip = the loop never happens).
- **Improve:** Reorder: land on `/-/deploying` → dashboard first, then prompt share in-context ("Your dashboard is live — share it") + persistent checklist item + day-2 email nudge for solo orgs.

### Step 8 — Deploying / build screen (`/-/deploying`, `DashboardBuilding.svelte`)
- **Goal:** "Tell me it's working."
- **Questions:** How long? Is it stuck?
- **Works:** Polls dashboards and auto-redirects when ready; failure emits a notification.
- **Friction:** "Hang tight!" with no step-level progress (upload → provision → ingest → reconcile) and no time expectation; a long ClickHouse/warehouse ingest looks identical to a hang. Failure notification lacks a recovery path.
- **Abandonment risk:** Medium for slow ingests.
- **Improve:** Reconcile-status-driven step list with per-resource state, plus "we'll email you when it's ready" for long builds.

### Step 9 — Cloud landing, trial & billing (`web-admin` org routes, `features/billing`)
- **Goal:** "Confirm my team can use this; understand what it costs."
- **Questions:** How long is my trial? What happens when it ends?
- **Works:** Trial banner copy is clear and staged (welcome → expiring → expired); quota errors route to a Choose Plan dialog; payment via Stripe portal.
- **Friction:** Trial starts via async job — the user's first sight of trial terms is a banner after deploy, never a proactive explanation. Trial expiry is a **cliff**: projects hibernate (banner: "this org's projects are now hibernating") — the strongest possible churn signal delivered as a hard gate. Duplicate "trial started" emails are possible in a repair race (`abp-13`); a credit-depletion job retry can leave some projects un-hibernated while telling the user everything is off (`abp-03`).
- **Abandonment risk:** High at expiry (by design, but harsher than needed).
- **Improve:** Progressive expiry (read-only dashboards for N days before hibernation); pre-expiry email with the org's own usage recap ("your team viewed 214 dashboards this month"); single owner for trial emails.

### Step 10 — Cloud-first signup, no org (`web-admin` root → `WelcomeMessage.svelte`)
- **Goal:** "I signed up on the website — now show me the product."
- **Questions:** Where is everything? How do I start?
- **Works:** Auth and org-routing logic are solid for users *with* orgs.
- **Friction:** A user with zero orgs gets one line — "Check out our Quickstart to create your first project" — linking to docs. **No in-product org creation, no demo project, no path to value without leaving the product.** The only supported route into Rill Cloud is via the local CLI deploy. This silently discards every website-first signup.
- **Abandonment risk:** **High** — this is a dead end, full stop.
- **Improve:** Real empty state: create org in-product → choose "explore a live demo project" (read-only shared demo org) or "deploy from your machine" (copy-paste `curl … && rill start` instructions with the org pre-linked via a *working* deep link — see Step 4's 404).

### Step 11 — Invited user (email → Auth0 → org/project landing)
- **Goal:** "Open the dashboard my teammate sent."
- **Questions:** Do I need to learn Rill to look at this?
- **Works:** Redirect logic lands invitees on the org/project; access-request flow exists for the uninvited; default Viewer role is sane.
- **Friction:** No viewer-oriented first-touch: an invitee lands on a full dashboard with no 10-second orientation (filters, time ranges, drill). Invitees are also the cheapest future *creators*, and nothing invites them to become one.
- **Abandonment risk:** Low for the visit; the missed opportunity is conversion of viewers → creators.
- **Improve:** One-time 3-tooltip orientation for first-time viewers; a subtle "build your own view of this data" affordance (saved-view/bookmark as the on-ramp).

---

## 3. Prioritized improvement ideas

Legend: Effort L/M/H; Priority Now/Next/Later. Ideas 1–7 are largely fixes to existing code paths (quick wins); 8–12 UX/content and product-led guidance; 13–14 personalization; 15–16 lifecycle/automation; 17–19 ambitious experiments.

| # | Idea | User problem addressed | Recommended change | Expected user impact | Expected business impact | Effort | Priority | Success metric |
|---|------|------------------------|--------------------|----------------------|--------------------------|--------|----------|----------------|
| 1 | Make "Generate sample data" un-killable | Splash CTA silently strands no-data users on an empty project (`plg-01`) | Gate splash card like its home twin (`$developerChat`/`$ai`); add deterministic offline DuckDB mock-data fallback; bounded waits + visible error/retry | Segment B reliably reaches a dashboard, logged-in or not | Rescues the most fragile evaluator segment; protects word-of-mouth | M | **Now** | % of sample-data attempts reaching a rendered dashboard (target >95%) |
| 2 | Kill the 5-second splash delay | Guaranteed 5s spinner + hard reload after picking an example (`plg-05`) | Replace `setTimeout` with reconcile/artifact-driven `goto` (pattern exists in `OnboardingWorkspace`) | Time-to-first-dashboard drops ~4s for every example user | Directly moves the core PLG metric | L | **Now** | Median splash-click → dashboard-render time |
| 3 | Surface every onboarding failure | Unpack, generate, and quick-pick failures are silent (`plg-02/04/11/13`) | Standard error affordance (message + Retry / Retry-without-AI / Edit YAML) on all generate/unpack/import surfaces; timeouts on `pollForFileCreation` and reconcile waits; numeric stateVersion compare (`plg-08`) | "Broken product" moments become recoverable errors | Fewer day-0 churns misattributed to product quality | M | **Now** | Rate of sessions ending within 60s of a failed onboarding action |
| 4 | One dashboard per source, not two | Import creates 4 resources incl. twin explore+canvas dashboards (`plg-06`) | Default to explore only; offer "Add canvas" as a follow-up action; clearer generated names | Halves the conceptual surface at the mental-model-forming moment | Higher wizard completion comprehension; fewer "what are these files" tickets | L | **Now** | Post-import dashboard-open rate; qualitative confusion mentions |
| 5 | Fix the `/deploy?org=` 404 | External deep links into the deploy funnel dead-end (`plg-03`) | Redirect to `/deploy/organization/select` (carries org through); add routing smoke test | Docs/campaign links work | Un-breaks a paid/marketing funnel entry | L | **Now** | 404 rate on `/deploy*` routes → 0 |
| 6 | Fix funnel telemetry at the source | `AppStart` dropped for misconfigured projects (`plg-07`); no events between start and deploy | Return/log `c.Err` properly; add events: `welcome_viewed`, `path_chosen{example\|data\|generated}`, `source_added`, `dashboard_viewed`, `deploy_clicked` | None directly | Makes the entire funnel measurable and unskewed — prerequisite for all experimentation in §5 | L | **Now** | Event coverage: every §2 step emits; funnel dashboards exist |
| 7 | Test connection + table preview before import | Warehouse credentials fail only after YAML is written and reconcile errors | In `ConnectorForm`, validate creds and list tables pre-import; select tables to import | Errors become inline and correctable; users import the right tables | Higher completion for the highest-intent segment (warehouse users) | M | **Now** | Connector-path wizard completion rate |
| 8 | Persistent onboarding checklist | After the first dashboard, users have no next action; deploy intent never forms | Home-screen tracker: Add data ✓ → View dashboard ✓ → Customize → Deploy → Invite, live off resource/deploy state; skipped invite resurfaces here | Orientation + a visible path from aha to activation | Lifts local→deploy conversion; resurrects the skipped share step | M | **Now** | Checklist-step completion rates; local→deploy conversion |
| 9 | Real cloud empty state | Zero-org web signups hit a docs link and leave (Step 10) | Replace `WelcomeMessage` with: create org in-product → "Explore live demo" (shared read-only demo org) or "Deploy from your machine" (copy-paste CLI + working org-linked deep link) | Website-first signups get a product path to value | Recovers an entire acquisition channel currently discarded | M | **Now** | Zero-org signup → org-created or demo-viewed within 24h |
| 10 | Dashboard-first, then share | Invite screen interrupts before the user sees their deployed project (Step 7) | Land on `/-/deploying` → live dashboard, then contextual "Your dashboard is live — invite your team / copy link" prompt | Share request arrives at peak pride, not before it | Higher invite-send rate → the retention driver | L | Next | Invite-send rate per first deploy; % orgs ≥2 members at day 7 |
| 11 | Legible build progress | "Hang tight!" spinners hide multi-minute ingests (Steps 3, 8) | Reconcile-status-driven step list (upload → provision → ingest per source → dashboards), elapsed time, "email me when ready" for long builds | Waiting feels like progress; long ingests stop looking like hangs | Fewer abandoned first deploys on big datasets | M | Next | Deploy-started → first-cloud-dashboard-view conversion |
| 12 | Dimension-less dashboard explainer | All-numeric tables yield a KPI-only "empty-feeling" dashboard with no explanation (`plg-12`) | Detect zero dims/time at generation; banner: "No categorical or time columns found — add dimensions to slice this data" + one-click add-dimension helper | "AI produced a bad dashboard" becomes a guided fix | Protects perceived AI quality | L | Next | Retention of users whose first table was all-numeric |
| 13 | Intent-aware splash | Three equal choices, no guidance; segments B and C mis-route (Step 1) | One-question segmentation ("I have data / just exploring") or infer from arrival context; reorder and annotate the three cards accordingly; recommend one | Less hesitation; each segment lands on its best path | Higher splash→path conversion, esp. segment B | L | Next | Splash → path-chosen rate; per-segment time-to-dashboard |
| 14 | Backend-driven, badged example gallery | Hardcoded example list drifts; network-dependent examples can land users on errored dashboards (`plg-10`) | Drive cards from `ListExamples` (extend proto with connector + entry file); reachability pre-flight; "requires network" badges | No broken-example dead-ends | Gallery stays honest as examples evolve | M | Next | Example-click → healthy-dashboard rate |
| 15 | Activation-state lifecycle emails | Nothing follows a skipped invite, a stalled setup, or an approaching trial end (Steps 7, 9) | Triggered sequence keyed to funnel state: day-0 "your dashboard link", day-2 "invite your team" (solo orgs only), stalled-deploy nudge, pre-expiry email with the org's own usage recap; consolidate trial-email ownership (`abp-13`) | Timely, relevant nudges instead of silence | Recovers skipped share steps; softens the trial cliff; trial→paid lift | M | Next | Email→action rate; solo-org → multi-member conversion |
| 16 | CLI deploy nudge | Local users reach a working dashboard and the terminal says nothing | When `rill start` detects a healthy dashboard + N views, print a one-time clickable "Share it: rill deploy" hint | Deploy intent forms where the developer actually lives | Cheap lift on local→cloud conversion | L | Next | rill-deploy runs attributed to nudge |
| 17 | "Explain this dashboard" AI narrative | First generated dashboard lands with zero orientation (Step 3) | One-shot AI summary on first view: what the data is, what each panel shows, two suggested questions to click into; deterministic template fallback when AI is off | The aha compounds — users understand *and* trust what they see | Differentiated first-run; deeper first sessions | M | Later | First-session interaction depth (filters/drills per session) |
| 18 | Progressive trial expiry | Expiry hibernates everything instantly — a churn cliff (Step 9) | Expired orgs get N days of read-only dashboards (viewers keep access, editors see upgrade wall) before hibernation; fix depletion-retry leaving projects half-off (`abp-03`) | Teams keep experiencing value while deciding | Higher trial→paid; ends "it just went dark" churn | M | Later | Trial→paid conversion; reactivation rate post-expiry |
| 19 | Viewer → creator on-ramp | Invitees (cheapest future creators) get no orientation and no creation path (Step 11) | First-visit 3-tooltip tour for viewers; "save your own view" bookmark affordance with a path to "build your own dashboard" | Viewers become self-serve analysts | Seat expansion inside existing orgs — the cheapest growth loop | M | Later | Viewer → first-bookmark rate; viewer → creator conversion |

---

## 4. Proposed future-state journey

Design principle: **one recommended path per segment, failure is always visible and recoverable, and the journey continues past deploy.** Steps marked ✂ are removed/postponed relative to today.

### Phase I — First session (local)

**F1. `rill start` → orientation line (unchanged + copy).**
User sees: terminal prints "Opening Rill at localhost:9009 — pick an example or connect your data." Action: none. Why: sets expectations for the browser handoff. Moves to value: primes the choice they're about to make. *Nothing skipped.*

**F2. Splash with a recommendation.**
User sees: same three cards, plus one line of value prop ("From raw data to a live dashboard in minutes — all as code") and a recommendation chip driven by a single implicit/explicit intent signal ("Just exploring? Start with an example"). Action: click one card. Why necessary: this is the segment router; without it, B-users gamble. Skipped/postponed: nothing — but the quick-pick grid gains an error state so it never silently vanishes. Moves to value: each segment enters its shortest path.

**F3a. Example path (segment B, default recommendation).**
User sees: card click → live progress ("Unpacking… Parsing… Opening dashboard") driven by reconcile state, landing on the example dashboard in ~1s. Action: none. Why: fastest possible aha. ✂ Removed: the 5s timeout + hard reload; network-dependent examples are badged or excluded. Moves to value: this *is* the value.

**F3b. Your-data path (segment A).**
User sees: connector picker → credentials form with **Test connection** → table list preview → pick table(s) → import progress → dashboard. Action: paste creds, pick a table. Why each field: creds and table choice are irreducible; everything else is generated. ✂ Postponed: canvas dashboard (explore only by default), source/model naming (auto-named, rename later). Moves to value: validation-before-write turns the riskiest step into a guided one.

**F3c. Generated-data path (segment B, no-login fallback guaranteed).**
User sees: prompt box; if AI available, agent-generated dataset; if not, template picker (e-commerce / web analytics / ad spend) backed by the deterministic DuckDB generator. Action: one prompt or one template click. Why: no-data users must never dead-end. ✂ Removed: unbounded waits; the AI-only dependency. Moves to value: a dashboard on plausible data within a minute, offline.

**F4. First dashboard + moment marking.**
User sees: their dashboard, a one-time "You just built a dashboard from `orders.csv` — here's what Rill generated" note (AI narrative when available, template otherwise), and the checklist appears: Add data ✓ · View dashboard ✓ · Customize · Deploy · Invite. Action: explore; optionally follow a suggested question. Why: converts aha into orientation and forms deploy intent. Moves to value: bridges first value → shared value.

### Phase II — Deploy & share

**F5. Deploy with expectations set.**
User sees: on "Deploy": one interstitial — "Free trial, no credit card. Your project uploads to Rill-managed infrastructure. [Deploy] [Learn more]" → browser login. Action: one click + auth. Why: the commitment moment deserves one honest sentence, not zero. ✂ Removed: the broken `?org=` redirect (fixed); the GitHub question (moved post-deploy). Moves to value: uncertainty is the enemy at this step.

**F6. Zero-decision org & project creation.**
User sees: org pre-filled from email domain/name, single editable field, "rename anytime"; project auto-named from directory; managed upload auto-fires. Action: confirm (or just watch). Why kept at all: the org slug must exist; one confirmation avoids surprise URLs. ✂ Postponed: display-name/name distinction, org selection ceremony for first-time users. Moves to value: the distance from intent to deployed is one click.

**F7. Build progress → dashboard first.**
User sees: step list (Uploading ✓ → Provisioning ✓ → Ingesting orders (2m) → Building dashboards), then **their live cloud dashboard**. Action: wait (with "email me when ready" for long ingests). Why: legible waiting; landing on the dashboard re-delivers value in the new context. ✂ Moved: the invite screen no longer precedes this. Moves to value: proof the deploy worked.

**F8. Share at peak pride.**
User sees: on the live dashboard, a contextual prompt: "Your dashboard is live — invite your team or copy the link", with the domain-access toggle. Action: send invites or skip. Why: the share ask lands right after re-experiencing value. If skipped: the checklist keeps "Invite" open and a day-2 email nudges solo orgs. Moves to value: this is the activation event's second half (another human sees it).

### Phase III — Early retention (days 1–30)

**F9. Cloud-first signups get a real start.** Zero-org users see: create-org form + "Explore a live demo" (shared read-only demo org) + "Deploy from your machine" with copy-paste CLI and a working org-linked deep link. Why: no acquisition channel dead-ends.

**F10. Invited users get a 10-second orientation.** Three tooltips (filters, time range, drill) on first visit; a "save your own view" bookmark affordance. Why: viewers are tomorrow's creators.

**F11. Lifecycle keyed to state, not time alone.** Day-0 dashboard link; stalled-deploy nudge at 48h; day-2 invite nudge (solo orgs only); weekly digest of dashboard views once the org is multi-member; pre-expiry email with the org's own usage recap; expiry → read-only grace before hibernation. Why: every message is triggered by what the user *hasn't* done or is about to lose, never generic.

---

## 5. Experiment backlog

Prerequisite for all: idea 6 (funnel instrumentation). Guardrails abbreviated: *TTFD* = time to first dashboard; *D7* = day-7 retention.

| # | Hypothesis | Variation | Primary metric | Guardrails | Segment | Effort | Expected impact |
|---|------------|-----------|----------------|-----------|---------|--------|-----------------|
| E1 | Removing the 5s delay + reload increases example-path completion | Reconcile-driven redirect vs current | Splash-click → dashboard-render conversion | Error rate on example open | All local first-runs | L | High (certain UX win; experiment mainly sizes it) |
| E2 | A recommended path on the splash lifts path-selection and TTFD | "Recommended for you" chip + one-line intent question vs neutral 3 cards | Splash → path-chosen rate; TTFD | Per-path completion (don't just shift clicks to a worse path) | First-run, esp. no-data | L | Medium-High |
| E3 | Offline sample-data fallback rescues logged-out no-data users | Deterministic template generator when AI unavailable vs current (hidden/failing CTA) | Sample-data attempt → dashboard rendered | Session length (ensure generated data isn't a dead-end itself) | Logged-out / AI-off first-runs | M | High |
| E4 | One dashboard (explore only) beats two on comprehension and progression | Explore-only + "Add canvas" follow-up vs 4-resource default | Post-import dashboard-open rate; next-action rate (customize/deploy) | Canvas adoption long-term | Add-data wizard users | L | Medium |
| E5 | Test-connection + table preview lifts warehouse wizard completion | Pre-validate + table picker vs write-YAML-then-reconcile | Connector wizard completion | Wizard duration (must not balloon) | Warehouse-connector users | M | High for segment A |
| E6 | Dashboard-first ordering increases invite sends | Deploy → dashboard → contextual share prompt vs current invite-screen-first | Invites sent per first deploy | Deploy completion; time to first cloud view | First-time deployers | L | High for share loop |
| E7 | A persistent checklist lifts local→deploy conversion | Home checklist (live state) vs none | Local→deploy conversion within 7 days | Checklist dismissal rate; annoyance signals | Post-first-dashboard locals | M | Medium-High |
| E8 | A real cloud empty state converts zero-org signups | Org-create + demo-project offer vs docs-link WelcomeMessage | Zero-org signup → org-created or demo-viewed (24h) | Support tickets from confused signups | Web-first signups | M | High (channel currently ~0) |
| E9 | Day-2 invite email converts solo orgs to multi-member | State-triggered nudge (solo orgs with ≥1 deploy) vs no email | Solo → ≥2-member orgs by day 7 | Unsubscribe rate | Solo-org deployers | L | Medium |
| E10 | Read-only grace beats hibernation cliff for trial→paid | 7-day read-only period post-expiry vs immediate hibernation | Trial→paid conversion | Infra cost of grace compute; abuse rate | Expiring trials | M | Medium-High |
| E11 | AI dashboard narrative deepens first sessions | One-shot "explain this dashboard" on first view vs none | Interactions (filters/drills) in first session; D7 | TTFD (narrative must not block render); AI cost | First dashboard viewers | M | Medium |
| E12 | Viewer orientation converts viewers to engaged users | 3-tooltip first-visit tour + bookmark affordance vs none | Viewer return-visit rate; first-bookmark rate | Tour dismissal/annoyance | First-time invitees | M | Medium (compounding) |

---

## 6. Measurement framework

**Leading indicators** predict activation/revenue and are actionable within days; **lagging indicators** confirm outcomes. Today only login/deploy events exist (`LoginStart/LoginSuccess/DeployIntent/DeploySuccess`) and `AppStart` is dropped for misconfigured projects — everything marked ★ must be newly instrumented.

| Stage | Metric | Type | Notes |
|-------|--------|------|-------|
| Signup / start | `rill start` first-runs (AppStart, fixed) | Leading | Fix `plg-first-run-07` first — current data is skewed toward healthy projects |
| | Cloud signups; % arriving with zero orgs ★ | Leading | Sizes the Step-10 dead-end |
| Onboarding completion | Splash → path-chosen rate, by path ★ | Leading | Hesitation measure for Step 1 |
| | Path completion: example-click → dashboard; wizard start → dashboard; sample-data attempt → dashboard ★ | Leading | The three funnels of §2 Steps 2a–2c; segment before comparing |
| | Onboarding error rate (surfaced vs silent) ★ | Leading | Silent-failure count should trend to zero as idea 3 lands |
| Time to first value | Median TTFD: `rill start` → first dashboard render, per path ★ | Leading | The core PLG metric; E1/E2/E3 all move it |
| | Median time: deploy-click → first cloud dashboard view ★ | Leading | Step 6–8 latency incl. ingest |
| Activation | **Local activation:** first-session dashboard view with ≥1 interaction ★ | Leading | Interaction distinguishes render from aha |
| | **Cloud activation:** deployed project + ≥1 non-creator dashboard view within 7 days ★ | Lagging | The definition to optimize toward |
| Step conversion | Full funnel: start → path → source added → dashboard viewed → deploy intent → deploy success → invite sent → invite accepted ★ | Leading | Per-step drop-off is the experiment targeting map |
| Early engagement | D1/D7 return rate (local sessions and cloud visits) | Lagging | Split creator vs viewer |
| | First-session depth: filters/drills/files-opened per session ★ | Leading | E11's primary |
| | % solo orgs → ≥2 members by day 7 ★ | Leading | The share-loop health metric (E6/E9) |
| Retention | Week-4 org activity (any dashboard view); creator retention (any project edit, week 4) | Lagging | |
| Trial → paid | Trial→paid conversion; time-to-payment-method; reactivation rate within 30 days of expiry ★ | Lagging | E10's primary; reactivation measures cliff harshness |
| Satisfaction / qualitative | Post-onboarding micro-survey (1 question after first dashboard: "Did this show what you expected?") ★; support tickets tagged onboarding; session recordings on splash + wizard | Mixed | The micro-survey is a leading proxy for perceived data-correctness — the trust variable none of the quantitative metrics capture |

**Instrumentation priorities (in order):** (1) fix `emitStartEvent`; (2) the seven ★ funnel events from start → invite-accepted; (3) per-path TTFD timers; (4) error-surface events (every failure state emits, with silent/surfaced tag); (5) the micro-survey.

---

## Appendix — Candid list of weak or unnecessary journey elements
- The **invite screen's current position** (before seeing the deployed project) is the clearest sequencing mistake — right feature, wrong moment.
- The **org create/select ceremony** on first deploy is pure overhead; nothing about it needs user input beyond one confirmable default.
- The **canvas dashboard in the default import** is speculative value at the worst possible time; it earns its place later, not at minute three.
- The **`WelcomeMessage` docs-link empty state** is not an onboarding step; it is the absence of one.
- The **fixed 5s splash delay** is acknowledged in-code as a hack and pays negative dividends on the single most important metric.
- The **GitHub deploy question** in the CLI's interactive first deploy answers a question no first-time deployer has yet asked.
