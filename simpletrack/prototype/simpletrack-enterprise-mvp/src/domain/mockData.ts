import type {
  AnalyticsEvent,
  DictionaryEvent,
  DictionaryProperty,
  GoalDefinition,
  IngestionRule,
  KpiMetric,
  LiveSignal,
  SiteConfig,
  TrafficPoint,
} from "./types";

export const siteConfig: SiteConfig = {
  id: "st_7c82d1",
  name: "Acme SaaS",
  domain: "acme.example",
  environment: "Production",
  trackerVersion: "1.0.0",
  lastSeenAt: "16 seconds ago",
};

export const trackerSnippet = `<script
  defer
  src="https://cdn.simpletrack.dev/tracker.js"
  data-website-id="st_7c82d1"
  data-domains="acme.example"
></script>`;

export const kpis: KpiMetric[] = [
  { label: "Pageviews", value: "12,842", delta: "+8.4%", trend: "up" },
  { label: "Visitors", value: "3,128", delta: "+4.1%", trend: "up" },
  { label: "Events", value: "1,904", delta: "+11.7%", trend: "up" },
  { label: "Activation goal", value: "18.6%", delta: "-1.2%", trend: "down" },
];

export const traffic: TrafficPoint[] = [
  { date: "04-22", pageviews: 920, visitors: 214, events: 133 },
  { date: "04-23", pageviews: 1104, visitors: 282, events: 188 },
  { date: "04-24", pageviews: 1290, visitors: 307, events: 224 },
  { date: "04-25", pageviews: 1188, visitors: 291, events: 207 },
  { date: "04-26", pageviews: 1362, visitors: 339, events: 241 },
  { date: "04-27", pageviews: 1516, visitors: 374, events: 286 },
  { date: "04-28", pageviews: 1714, visitors: 422, events: 335 },
];

export const liveSignals: LiveSignal[] = [
  { id: "sig-001", time: "now", type: "event", name: "first_event_sent", path: "/app/install", visitor: "v_8021", status: "accepted" },
  { id: "sig-002", time: "18s", type: "pageview", name: "Pageview", path: "/pricing", visitor: "v_8017", status: "accepted" },
  { id: "sig-003", time: "32s", type: "event", name: "signup_completed", path: "/signup/success", visitor: "v_8013", status: "accepted" },
  { id: "sig-004", time: "51s", type: "event", name: "sdk_install_completed", path: "/app/install", visitor: "v_8009", status: "accepted" },
  { id: "sig-005", time: "1m", type: "pageview", name: "Pageview", path: "/docs/install", visitor: "v_8004", status: "accepted" },
];

export const topPages = [
  { key: "/", path: "/", views: 3221, visitors: 980 },
  { key: "/pricing", path: "/pricing", views: 1882, visitors: 611 },
  { key: "/docs/install", path: "/docs/install", views: 1430, visitors: 392 },
  { key: "/compare", path: "/compare", views: 1184, visitors: 331 },
  { key: "/signup/success", path: "/signup/success", views: 801, visitors: 244 },
];

export const topReferrers = [
  { key: "direct", source: "direct", visitors: 1124, share: "35.9%" },
  { key: "google", source: "google", visitors: 884, share: "28.3%" },
  { key: "producthunt", source: "producthunt", visitors: 482, share: "15.4%" },
  { key: "linkedin", source: "linkedin", visitors: 311, share: "9.9%" },
  { key: "docs", source: "docs.simpletrack.dev", visitors: 211, share: "6.7%" },
];

export const analyticsEvents: AnalyticsEvent[] = [
  {
    key: "first_event_sent",
    name: "first_event_sent",
    description: "User sent the first verified business event",
    count: 416,
    visitors: 382,
    lastSeen: "16 seconds ago",
    health: "healthy",
    properties: {
      plan: { free: 148, trial: 169, pro_monthly: 78, pro_annual: 21 },
      campaign: { producthunt_launch: 121, docs_seo: 102, google_brand: 91, email_nurture: 62, linkedin_founder: 40 },
      role: { founder: 144, engineer: 108, growth_lead: 88, product_manager: 51, ops: 25 },
    },
  },
  {
    key: "sdk_install_completed",
    name: "sdk_install_completed",
    description: "Tracker snippet passed client-side verification",
    count: 389,
    visitors: 357,
    lastSeen: "51 seconds ago",
    health: "healthy",
    properties: {
      plan: { free: 140, trial: 155, pro_monthly: 72, pro_annual: 22 },
      surface: { onboarding: 277, settings: 112 },
      workspaceSize: { "1-10": 196, "11-50": 121, "51-200": 54, "201-500": 18 },
    },
  },
  {
    key: "signup_completed",
    name: "signup_completed",
    description: "Signup finished after account verification",
    count: 302,
    visitors: 296,
    lastSeen: "32 seconds ago",
    health: "healthy",
    properties: {
      plan: { free: 118, trial: 129, pro_monthly: 42, pro_annual: 13 },
      campaign: { google_brand: 86, producthunt_launch: 75, docs_seo: 58, linkedin_founder: 44, email_nurture: 39 },
      cohort: { spring_launch: 121, self_serve_wave: 109, paid_pilot: 72 },
    },
  },
  {
    key: "install_started",
    name: "install_started",
    description: "User opened install instructions",
    count: 621,
    visitors: 544,
    lastSeen: "2 minutes ago",
    health: "healthy",
    properties: {
      surface: { onboarding: 471, settings: 150 },
      plan: { free: 271, trial: 229, pro_monthly: 94, pro_annual: 27 },
      role: { founder: 211, engineer: 198, growth_lead: 112, product_manager: 71, ops: 29 },
    },
  },
  {
    key: "checkout_completed",
    name: "checkout_completed",
    description: "Paid conversion confirmed by backend",
    count: 49,
    visitors: 49,
    lastSeen: "47 minutes ago",
    health: "reserved",
    properties: {
      plan: { pro_monthly: 31, pro_annual: 18 },
      currency: { USD: 49 },
      campaign: { producthunt_launch: 17, google_brand: 14, linkedin_founder: 9, email_nurture: 9 },
    },
  },
];

export const goals: GoalDefinition[] = [
  {
    id: "goal_first_event",
    name: "Activation: first event sent",
    type: "event",
    rule: "event.name = first_event_sent",
    denominator: "unique visitors with tracker loaded",
    conversions: 382,
    population: 2054,
    rate: "18.6%",
    status: "active",
  },
  {
    id: "goal_signup",
    name: "Signup completed",
    type: "event",
    rule: "event.name = signup_completed",
    denominator: "unique visitors",
    conversions: 296,
    population: 3128,
    rate: "9.5%",
    status: "draft",
  },
];

export const dictionaryEvents: DictionaryEvent[] = [
  { name: "signup_started", status: "active", required: "plan, campaign, cohort" },
  { name: "signup_completed", status: "active", required: "plan, campaign, cohort" },
  { name: "install_started", status: "active", required: "surface, plan" },
  { name: "sdk_install_completed", status: "active", required: "surface, plan" },
  { name: "first_event_sent", status: "active", required: "surface, plan" },
  { name: "checkout_completed", status: "reserved", required: "revenue, currency, plan, campaign, cohort" },
];

export const dictionaryProperties: DictionaryProperty[] = [
  { key: "plan", type: "enum", values: "free, trial, pro_monthly, pro_annual" },
  { key: "campaign", type: "enum", values: "producthunt_launch, google_brand, docs_seo, linkedin_founder, email_nurture" },
  { key: "cohort", type: "enum", values: "spring_launch, self_serve_wave, paid_pilot" },
  { key: "role", type: "enum", values: "founder, growth_lead, engineer, product_manager, ops" },
  { key: "workspaceSize", type: "enum", values: "1-10, 11-50, 51-200, 201-500" },
  { key: "currency", type: "enum", values: "USD" },
];

export const ingestionRules: IngestionRule[] = [
  { rule: "Reject PII keys", detail: "email, phone, name, token, cookie", mode: "enforced" },
  { rule: "Flatten properties", detail: "nested JSON is rejected", mode: "enforced" },
  { rule: "Domain allowlist", detail: "acme.example, app.acme.example", mode: "enforced" },
  { rule: "Unknown event handling", detail: "accepted into quarantine for review", mode: "review" },
];

export const backendPhases = [
  { phase: "P0", name: "Contract", scope: "Event schema, property dictionary, privacy baseline, UTM rules" },
  { phase: "P1", name: "Trust loop", scope: "Collect API, validator, event store, realtime read model, events query, simple goal" },
  { phase: "P2", name: "Diagnosis", scope: "Breakdown, compare, sessions, segments, funnels, journeys" },
  { phase: "P3", name: "Value", scope: "Cohorts, retention, revenue, attribution, links, pixels" },
  { phase: "P4", name: "Expansion", scope: "Teams, share URLs, API keys, performance, replays" },
];
