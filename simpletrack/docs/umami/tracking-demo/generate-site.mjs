import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(fileURLToPath(import.meta.url));

const sample = {
  preset: "growth-baseline-x3",
  totalUsers: 648,
  browserUsers: 72,
  batchUsers: 576,
  sessionsPerUser: 3,
  totalSessions: 1944,
  eventRange: "15,552-23,328",
  expectedEvents: 19440,
  logicalDays: 42,
  browserPaid: 6,
  batchPaid: 48,
  paidWorkspaces: 54,
  paidMonthly: 36,
  paidAnnual: 18,
  plans: ["free", "trial", "pro_monthly", "pro_annual"],
  campaigns: [
    "producthunt_launch",
    "google_brand",
    "google_competitor",
    "docs_seo",
    "linkedin_founder",
    "email_nurture",
  ],
  cohorts: ["spring_launch", "self_serve_wave", "paid_pilot"],
};

const businessEvents = [
  "pricing_viewed",
  "compare_opened",
  "signup_started",
  "signup_completed",
  "workspace_created",
  "install_started",
  "sdk_install_completed",
  "first_event_sent",
  "dashboard_viewed",
  "filter_applied",
  "segment_opened",
  "cohort_opened",
  "checkout_started",
  "checkout_completed",
  "subscription_upgraded",
  "billing_viewed",
];

const publicNav = [
  ["Product", "index.html"],
  ["Pricing", "pricing.html"],
  ["Compare", "compare.html"],
  ["Docs", "docs.html"],
  ["Sign up", "signup.html"],
];

const appNav = [
  ["Dashboard", "app/dashboard.html"],
  ["Events", "app/events.html"],
  ["Funnels", "app/funnels.html"],
  ["Segments", "app/segments.html"],
  ["Cohorts", "app/cohorts.html"],
  ["Billing", "billing.html"],
];

const pages = [
  {
    path: "index.html",
    title: "Growth Console",
    section: "public",
    kind: "landing",
    event: "dashboard_viewed",
    eyebrow: "SimpleTrack Growth Console",
    headline: "Acquisition, activation, and revenue in one compact operating view.",
    body: "A production-like SaaS surface for feeding Umami Cloud with real pageviews, sessions, events, filters, attribution, revenue, performance, and replay evidence.",
    primary: ["Start trial", "signup.html", "signup_started"],
    secondary: ["Inspect pricing", "pricing.html", "pricing_viewed"],
  },
  {
    path: "pricing.html",
    title: "Pricing",
    section: "public",
    kind: "pricing",
    event: "pricing_viewed",
    eyebrow: "Pricing",
    headline: "Clear plans for the four buying states in the growth baseline.",
    body: "Free, trial, monthly, and annual plans are carried through every event so Compare, Breakdown, Revenue, and Attribution have usable dimensions.",
    primary: ["Start trial", "signup.html", "signup_started"],
    secondary: ["Compare platforms", "compare.html", "compare_opened"],
  },
  {
    path: "compare.html",
    title: "Compare",
    section: "public",
    kind: "compare",
    event: "compare_opened",
    eyebrow: "Competitive intent",
    headline: "Separate high-intent competitor traffic from casual discovery.",
    body: "This page creates a realistic comparison path for google_competitor, founder-led LinkedIn traffic, and returning product evaluation sessions.",
    primary: ["Choose SimpleTrack", "signup.html", "signup_started"],
    secondary: ["Read implementation docs", "docs.html", "install_started"],
  },
  {
    path: "docs.html",
    title: "Docs",
    section: "public",
    kind: "docs",
    event: "install_started",
    eyebrow: "Integration docs",
    headline: "Install, identify, track, verify, then analyze.",
    body: "Docs traffic models high-intent SEO and activation behavior without storing any account, cookie, token, or secret in the repository.",
    primary: ["Install web SDK", "install.html", "install_started"],
    secondary: ["Create workspace", "signup.html", "signup_started"],
  },
  {
    path: "signup.html",
    title: "Create Workspace",
    section: "public",
    kind: "signup",
    event: "signup_started",
    eyebrow: "Workspace setup",
    headline: "Create a workspace with the same plan, campaign, and cohort context.",
    body: "The form is intentionally lightweight so real browser personas can complete it and still generate pageviews, interactions, and identity calls.",
    primary: ["Create workspace", "signup-success.html", "signup_completed"],
    secondary: ["Review pricing", "pricing.html", "pricing_viewed"],
  },
  {
    path: "signup-success.html",
    title: "Workspace Created",
    section: "public",
    kind: "success",
    event: "workspace_created",
    eyebrow: "Workspace ready",
    headline: "The account exists. Activation now depends on first data.",
    body: "This transition gives Goals, Funnels, and Journeys a clean boundary between signup completion and installation.",
    primary: ["Start install", "install.html", "install_started"],
    secondary: ["Open dashboard", "app/dashboard.html", "dashboard_viewed"],
  },
  {
    path: "install.html",
    title: "Install",
    section: "public",
    kind: "install",
    event: "install_started",
    eyebrow: "Activation",
    headline: "Install the snippet and keep the event plan readable.",
    body: "The install path feeds activation, first-value, retention, and replay checks with realistic operator behavior.",
    primary: ["Open web install", "install-web.html", "install_started"],
    secondary: ["Open docs", "docs.html", "install_started"],
  },
  {
    path: "install-web.html",
    title: "Web Install",
    section: "public",
    kind: "code",
    event: "sdk_install_completed",
    eyebrow: "Web SDK",
    headline: "One snippet, one website id, one verification event.",
    body: "The simulated install step exposes the same public tracker contract that all pages use.",
    primary: ["Verify first event", "install-verify.html", "first_event_sent"],
    secondary: ["Open dashboard", "app/dashboard.html", "dashboard_viewed"],
  },
  {
    path: "install-verify.html",
    title: "Install Verified",
    section: "public",
    kind: "verified",
    event: "first_event_sent",
    eyebrow: "First value",
    headline: "The first event arrived. Now the workspace can explain conversion quality.",
    body: "This page marks activation and gives Retention, Funnels, and Journeys a consistent milestone.",
    primary: ["Open dashboard", "app/dashboard.html", "dashboard_viewed"],
    secondary: ["Configure funnels", "app/funnels.html", "filter_applied"],
  },
  {
    path: "checkout.html",
    title: "Checkout",
    section: "public",
    kind: "checkout",
    event: "checkout_started",
    eyebrow: "Revenue",
    headline: "Convert trial intent into attributable revenue.",
    body: "Checkout carries plan, campaign, cohort, billing cycle, currency, and revenue fields for Revenue and Attribution review.",
    primary: ["Complete checkout", "checkout-success.html", "checkout_completed"],
    secondary: ["Review pricing", "pricing.html", "pricing_viewed"],
  },
  {
    path: "checkout-success.html",
    title: "Checkout Complete",
    section: "public",
    kind: "paid",
    event: "checkout_completed",
    eyebrow: "Paid workspace",
    headline: "Revenue is attached to the same journey that produced the signup.",
    body: "The success state gives Revenue, Goals, Funnels, and Attribution a concrete conversion endpoint.",
    primary: ["View billing", "billing.html", "billing_viewed"],
    secondary: ["Open dashboard", "app/dashboard.html", "dashboard_viewed"],
  },
  {
    path: "billing.html",
    title: "Billing",
    section: "app",
    kind: "billing",
    event: "billing_viewed",
    eyebrow: "Billing",
    headline: "Plan, usage, invoice, and upgrade signals.",
    body: "Billing keeps monetization inside the product workspace while still generating subscription and revenue properties.",
    primary: ["Upgrade annual", "checkout.html", "subscription_upgraded"],
    secondary: ["Return dashboard", "app/dashboard.html", "dashboard_viewed"],
  },
  {
    path: "app/dashboard.html",
    title: "Dashboard",
    section: "app",
    kind: "dashboard",
    event: "dashboard_viewed",
    eyebrow: "Overview",
    headline: "Growth health for the last 42 logical days.",
    body: "Acquisition quality, activation rate, cohort health, and revenue movement are visible without leaving the first viewport.",
    primary: ["Apply campaign filter", "app/events.html", "filter_applied"],
    secondary: ["Start checkout", "checkout.html", "checkout_started"],
  },
  {
    path: "app/events.html",
    title: "Events",
    section: "app",
    kind: "events",
    event: "dashboard_viewed",
    eyebrow: "Event analysis",
    headline: "Events and properties with dense business context.",
    body: "Every core event carries plan, campaign, cohort, role, workspace size, logical day, and source properties.",
    primary: ["Apply segment", "app/segments.html", "filter_applied"],
    secondary: ["Open cohort", "app/cohorts.html", "cohort_opened"],
  },
  {
    path: "app/funnels.html",
    title: "Funnels",
    section: "app",
    kind: "funnels",
    event: "dashboard_viewed",
    eyebrow: "Conversion",
    headline: "Pricing to first event to paid workspace.",
    body: "The fixed path maps pricing_viewed, signup_completed, sdk_install_completed, first_event_sent, and checkout_completed.",
    primary: ["Inspect journey", "app/dashboard.html", "filter_applied"],
    secondary: ["Open billing", "billing.html", "billing_viewed"],
  },
  {
    path: "app/segments.html",
    title: "Segments",
    section: "app",
    kind: "segments",
    event: "segment_opened",
    eyebrow: "Segments",
    headline: "Reusable traffic slices for source, plan, cohort, and role.",
    body: "Segments are modeled as event properties and browser interactions so Cloud filters have high-cardinality material to inspect.",
    primary: ["Apply segment", "app/events.html", "filter_applied"],
    secondary: ["Open cohorts", "app/cohorts.html", "cohort_opened"],
  },
  {
    path: "app/cohorts.html",
    title: "Cohorts",
    section: "app",
    kind: "cohorts",
    event: "cohort_opened",
    eyebrow: "Retention",
    headline: "Three launch cohorts with returning-session signals.",
    body: "spring_launch, self_serve_wave, and paid_pilot are carried through identify, events, revenue, and filters.",
    primary: ["Apply cohort", "app/dashboard.html", "filter_applied"],
    secondary: ["Open funnels", "app/funnels.html", "filter_applied"],
  },
];

const featureMatrix = [
  ["Session", "72 browser personas and 576 batch users, 3 sessions each", "pageview, identify, tag, persona context", "P07-S01"],
  ["RealTime", "browser flows plus CTA events", "live pageviews and track() calls", "P07-S02"],
  ["Performance", "data-performance enabled by URL flag", "real browser page loads", "P07-S03"],
  ["Compare", "plan, campaign, cohort, source dimensions", "batch event density", "P07-S04"],
  ["BreakDown", "source, medium, role, plan, cohort properties", "batch event density", "P07-S05"],
  ["Goals", "signup, first event, checkout endpoints", "Cloud goal configuration", "P07-S06"],
  ["Filter", "plan, campaign, cohort, role, workspace size", "fields and named slices", "P07-S07"],
  ["Funnels", "pricing -> signup -> install -> first_event -> checkout", "browser and batch event order", "P08-S01"],
  ["Journeys", "landing, compare, docs, dashboard, checkout paths", "real multi-page navigation", "P08-S02"],
  ["Retention", "3 cohort labels over 42 logical days", "synthetic_date and logical_day", "P08-S03"],
  ["Replays", "browser personas only", "recorder availability and account capability", "P08-S04"],
  ["Segments", "named traffic slices", "filter_applied and segment_opened", "P08-S05"],
  ["Cohorts", sample.cohorts.join(" / "), "identify and event properties", "P08-S06"],
  ["UTM", sample.campaigns.join(" / "), "URL params and event properties", "P08-S07"],
  ["Revenue", "54 paid workspaces", "checkout_completed revenue fields", "P08-S08"],
  ["Attribution", "campaign -> signup -> checkout", "UTM plus revenue fields", "P08-S09"],
];

function write(relativePath, content) {
  const target = join(root, relativePath);
  mkdirSync(dirname(target), { recursive: true });
  writeFileSync(target, `${content.trim()}\n`, "utf8");
}

function md(lines) {
  return lines.join("\n");
}

function relPrefix(pagePath) {
  const depth = pagePath.split("/").length - 1;
  return depth === 0 ? "./" : "../".repeat(depth);
}

function pageLink(currentPath, targetPath) {
  if (targetPath.startsWith("#") || targetPath.startsWith("http")) {
    return targetPath;
  }
  const depth = currentPath.split("/").length - 1;
  return `${depth === 0 ? "" : "../".repeat(depth)}${targetPath}`;
}

function navLinks(page, items) {
  return items
    .map(([label, target]) => {
      const active = page.path === target ? ' aria-current="page"' : "";
      return `<a${active} href="${pageLink(page.path, target)}">${label}</a>`;
    })
    .join("\n");
}

function lineChart() {
  return `<svg class="line-chart" viewBox="0 0 720 220" role="img" aria-label="42 day growth trend">
    <defs>
      <linearGradient id="lineFill" x1="0" x2="0" y1="0" y2="1">
        <stop offset="0" stop-color="#0b6f6a" stop-opacity="0.18" />
        <stop offset="1" stop-color="#0b6f6a" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path class="grid" d="M36 42H684M36 92H684M36 142H684M36 192H684" />
    <path class="area" d="M38 178C94 150 120 122 168 132C220 144 248 86 302 96C358 106 382 58 440 64C512 72 548 34 682 44L682 198L38 198Z" />
    <path class="line" d="M38 178C94 150 120 122 168 132C220 144 248 86 302 96C358 106 382 58 440 64C512 72 548 34 682 44" />
  </svg>`;
}

function barChart() {
  return `<div class="bar-chart" role="img" aria-label="Campaign distribution">
    <span style="--h:58%"></span><span style="--h:84%"></span><span style="--h:46%"></span>
    <span style="--h:72%"></span><span style="--h:52%"></span><span style="--h:96%"></span>
    <span style="--h:66%"></span><span style="--h:38%"></span><span style="--h:77%"></span>
  </div>`;
}

function metricStrip() {
  return `<section class="metric-strip" aria-label="Growth baseline metrics">
    <div><span>Users</span><strong>648</strong><small>216 x 3 baseline</small></div>
    <div><span>Sessions</span><strong>1,944</strong><small>3 per user</small></div>
    <div><span>Events</span><strong>~19.4k</strong><small>8/10/12 templates</small></div>
    <div><span>Revenue</span><strong>$6.3k</strong><small>54 paid workspaces</small></div>
  </section>`;
}

function productPreview() {
  const rows = [
    ["producthunt_launch", "trial", "spring_launch", "signup_completed", "18.7%"],
    ["google_competitor", "pro_monthly", "paid_pilot", "checkout_completed", "$29"],
    ["docs_seo", "free", "self_serve_wave", "first_event_sent", "41.3%"],
    ["linkedin_founder", "pro_annual", "paid_pilot", "subscription_upgraded", "$290"],
  ]
    .map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`)
    .join("\n");

  return `<div class="product-preview" aria-label="SimpleTrack analytics preview">
    <div class="preview-head">
      <span>Growth baseline x3</span>
      <strong>42 logical days</strong>
    </div>
    ${metricStrip()}
    <div class="preview-grid">
      <div class="panel panel-wide">
        <div class="panel-head"><span>Activation trend</span><strong>first_event_sent</strong></div>
        ${lineChart()}
      </div>
      <div class="panel">
        <div class="panel-head"><span>UTM mix</span><strong>6 campaigns</strong></div>
        ${barChart()}
      </div>
    </div>
    <div class="table-panel compact-table">
      <table>
        <thead><tr><th>Campaign</th><th>Plan</th><th>Cohort</th><th>Event</th><th>Value</th></tr></thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
  </div>`;
}

function surfaceForKind(page) {
  if (page.kind === "pricing") {
    return `<div class="pricing-grid">
      <article><span>Free</span><strong>$0</strong><p>Top-funnel and evaluation workspaces.</p><small>324 users in the default mix</small></article>
      <article class="selected"><span>Trial</span><strong>14 days</strong><p>Signup and install behavior before first value.</p><small>270 users before paid conversion</small></article>
      <article><span>Pro Monthly</span><strong>$29</strong><p>Monthly subscription revenue events.</p><small>36 paid workspaces</small></article>
      <article><span>Pro Annual</span><strong>$290</strong><p>Annual revenue and attribution checks.</p><small>18 paid workspaces</small></article>
    </div>`;
  }

  if (page.kind === "compare") {
    return `<div class="compare-board">
      <div class="compare-row head"><span>Signal</span><span>SimpleTrack</span><span>Generic analytics</span></div>
      <div class="compare-row"><span>Attribution</span><strong>Campaign to checkout</strong><em>Traffic only</em></div>
      <div class="compare-row"><span>Activation</span><strong>First event milestone</strong><em>Manual review</em></div>
      <div class="compare-row"><span>Workspace</span><strong>Dense operator view</strong><em>Card dashboard</em></div>
      <div class="compare-row"><span>Research fit</span><strong>Umami module coverage</strong><em>Single demo path</em></div>
    </div>`;
  }

  if (page.kind === "docs" || page.kind === "install" || page.kind === "code" || page.kind === "verified") {
    return `<div class="docs-board">
      <div><span>1</span><strong>Load tracker</strong><p>initTracker({ websiteId, scriptUrl, enablePerformance, enableReplays })</p></div>
      <div><span>2</span><strong>Identify workspace</strong><p>identifyBusinessUser(profile) carries plan, campaign, cohort, role, and stage.</p></div>
      <div><span>3</span><strong>Track business events</strong><p>trackBusinessEvent(name, data) keeps event names fixed and properties dense.</p></div>
      <div><span>4</span><strong>Verify first value</strong><p>first_event_sent becomes the activation milestone for goals and funnels.</p></div>
    </div>`;
  }

  if (page.kind === "signup") {
    return `<form class="signup-form" data-st-event="signup_completed">
      <label>Work email<input name="email" type="email" value="founder@example.com" autocomplete="email" /></label>
      <label>Workspace size<select name="workspaceSize"><option>1-10</option><option selected>11-50</option><option>51-200</option></select></label>
      <label>Primary goal<select name="goal"><option>Activation</option><option selected>Revenue attribution</option><option>Retention</option></select></label>
      <button class="button primary" type="submit">Create workspace</button>
    </form>`;
  }

  if (page.kind === "checkout" || page.kind === "paid") {
    return `<div class="checkout-panel">
      <div class="invoice-line"><span>Plan</span><strong data-plan-label>Pro Monthly</strong></div>
      <div class="invoice-line"><span>Campaign</span><strong data-campaign-label>producthunt_launch</strong></div>
      <div class="invoice-line"><span>Cohort</span><strong data-cohort-label>spring_launch</strong></div>
      <div class="invoice-line total"><span>Revenue</span><strong data-revenue-label>$29 USD</strong></div>
    </div>`;
  }

  return productPreview();
}

function actionRow(page) {
  const primaryProps = page.primary[2] === "checkout_completed"
    ? ' data-st-prop-currency="USD" data-st-prop-revenue-source="checkout"'
    : "";
  const secondaryProps = page.secondary[2] === "billing_viewed"
    ? ' data-st-prop-section="billing"'
    : "";
  return `<div class="action-row">
    <a class="button primary" data-st-event="${page.primary[2]}"${primaryProps} href="${pageLink(page.path, page.primary[1])}">${page.primary[0]}</a>
    <a class="button secondary" data-st-event="${page.secondary[2]}"${secondaryProps} href="${pageLink(page.path, page.secondary[1])}">${page.secondary[0]}</a>
  </div>`;
}

function publicPage(page) {
  return `<header class="topbar">
    <a class="brand" href="${pageLink(page.path, "index.html")}"><span></span>SimpleTrack</a>
    <nav>${navLinks(page, publicNav)}</nav>
    <a class="button small" data-st-event="signup_started" href="${pageLink(page.path, "signup.html")}">Start</a>
  </header>
  <main class="marketing-shell">
    <section class="hero-band">
      <div class="hero-copy">
        <span class="eyebrow">${page.eyebrow}</span>
        <h1>${page.headline}</h1>
        <p>${page.body}</p>
        ${actionRow(page)}
      </div>
      ${surfaceForKind(page)}
    </section>
    <section class="feature-band" aria-label="Simulation coverage">
      <article><span>Traffic</span><strong>72 browser personas</strong><p>Real page transitions for sessions, realtime, performance, and replay checks.</p></article>
      <article><span>Density</span><strong>576 batch users</strong><p>High-volume business events for breakdowns, funnels, UTM, and attribution.</p></article>
      <article><span>Revenue</span><strong>54 paid workspaces</strong><p>Checkout events carry revenue, currency, plan, campaign, and cohort.</p></article>
    </section>
  </main>`;
}

function appRows() {
  return [
    ["producthunt_launch", "trial", "spring_launch", "signup_completed", "18.7%", "goal"],
    ["google_brand", "trial", "self_serve_wave", "first_event_sent", "43.1%", "activation"],
    ["google_competitor", "pro_monthly", "paid_pilot", "checkout_completed", "$29", "revenue"],
    ["docs_seo", "free", "self_serve_wave", "install_started", "31.8%", "journey"],
    ["linkedin_founder", "pro_annual", "paid_pilot", "subscription_upgraded", "$290", "upgrade"],
  ]
    .map((row) => `<tr>${row.map((cell) => `<td>${cell}</td>`).join("")}</tr>`)
    .join("\n");
}

function appPage(page) {
  return `<main class="workspace">
    <aside class="sidebar">
      <a class="brand" href="${pageLink(page.path, "index.html")}"><span></span>SimpleTrack</a>
      <nav>${navLinks(page, appNav)}</nav>
      <div class="side-note">
        <span>Preset</span>
        <strong>${sample.preset}</strong>
        <small>${sample.totalSessions} sessions / ~${sample.expectedEvents.toLocaleString("en-US")} events</small>
      </div>
    </aside>
    <section class="workpane">
      <header class="toolbar">
        <div>
          <span class="eyebrow">${page.eyebrow}</span>
          <h1>${page.headline}</h1>
          <p>${page.body}</p>
        </div>
        <div class="toolbar-actions" aria-label="Filters">
          <button data-st-event="filter_applied" data-st-prop-field="campaign">Campaign</button>
          <button data-st-event="filter_applied" data-st-prop-field="cohort">Cohort</button>
          <button data-st-event="filter_applied" data-st-prop-field="plan">Plan</button>
        </div>
      </header>
      ${metricStrip()}
      <section class="filter-bar" aria-label="Active filters">
        <button data-st-event="filter_applied" data-st-prop-campaign="producthunt_launch">producthunt_launch</button>
        <button data-st-event="segment_opened" data-st-prop-segment="activated_trials">activated_trials</button>
        <button data-st-event="cohort_opened" data-st-prop-cohort="paid_pilot">paid_pilot</button>
        <button data-st-event="filter_applied" data-st-prop-plan="pro_monthly">pro_monthly</button>
      </section>
      <section class="analysis-grid">
        <div class="panel panel-wide">
          <div class="panel-head"><span>42 day signal</span><strong>${page.title}</strong></div>
          ${lineChart()}
        </div>
        <div class="panel">
          <div class="panel-head"><span>Campaign mix</span><strong>6 groups</strong></div>
          ${barChart()}
        </div>
      </section>
      <section class="table-panel">
        <div class="panel-head"><span>Breakdown</span><strong>Plan x Campaign x Cohort</strong></div>
        <table>
          <thead><tr><th>Campaign</th><th>Plan</th><th>Cohort</th><th>Event</th><th>Value</th><th>Use</th></tr></thead>
          <tbody>${appRows()}</tbody>
        </table>
      </section>
      ${actionRow(page)}
    </section>
  </main>`;
}

function html(page) {
  const prefix = relPrefix(page.path);
  const shell = page.section === "app" ? appPage(page) : publicPage(page);
  return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>${page.title} | SimpleTrack</title>
    <link rel="stylesheet" href="${prefix}styles.css" />
    <script type="module" src="${prefix}main.js"></script>
  </head>
  <body data-page="${page.kind}" data-page-event="${page.event}">
    ${shell}
  </body>
</html>`;
}

function styles() {
  return `:root {
  color-scheme: light;
  --bg: #f6f7f9;
  --surface: #ffffff;
  --surface-2: #eef2f5;
  --ink: #111827;
  --muted: #5d6675;
  --soft: #8a94a6;
  --line: #d8dee7;
  --line-strong: #c4ccd8;
  --accent: #0b6f6a;
  --accent-ink: #064d49;
  --accent-soft: #e6f3f1;
  --blue: #1b3f5f;
  --good: #147654;
  --warn: #a26313;
  --danger: #a82b20;
  --radius: 10px;
  --shadow: 0 20px 60px rgba(17, 24, 39, 0.08);
  font-family: "Aptos", "IBM Plex Sans", "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
}

* { box-sizing: border-box; }

html { min-width: 0; }

body {
  min-width: 0;
  margin: 0;
  overflow-x: hidden;
  background:
    radial-gradient(circle at 18% 12%, rgba(11, 111, 106, 0.08), transparent 28%),
    linear-gradient(180deg, #fbfcfd 0%, var(--bg) 45%, #edf1f5 100%);
  color: var(--ink);
  font-size: 15px;
}

a { color: inherit; text-decoration: none; }
button, input, select { font: inherit; }

.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-weight: 760;
  letter-spacing: -0.02em;
}

.brand span {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  background: linear-gradient(135deg, var(--accent), #8bbab4);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.72);
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 10;
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr) auto;
  align-items: center;
  gap: 24px;
  min-height: 68px;
  padding: 0 clamp(18px, 4vw, 42px);
  border-bottom: 1px solid rgba(216, 222, 231, 0.86);
  background: rgba(246, 247, 249, 0.9);
  backdrop-filter: blur(16px);
}

.topbar nav,
.sidebar nav {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.topbar nav a,
.sidebar nav a {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  border-radius: 8px;
  padding: 0 12px;
  color: var(--muted);
  font-size: 14px;
  transition: background 180ms ease, color 180ms ease;
}

.topbar nav a:hover,
.sidebar nav a:hover,
.topbar nav a[aria-current="page"],
.sidebar nav a[aria-current="page"] {
  background: var(--accent-soft);
  color: var(--accent-ink);
}

.button,
button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 44px;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  padding: 0 15px;
  background: var(--surface);
  color: var(--ink);
  cursor: pointer;
  font-weight: 680;
  transition: transform 160ms ease, border-color 160ms ease, background 160ms ease;
}

.button:hover,
button:hover {
  transform: translateY(-1px);
  border-color: #9aa6b8;
}

.button.primary {
  border-color: var(--accent);
  background: var(--accent);
  color: #fff;
}

.button.secondary {
  background: transparent;
}

.button.small {
  min-height: 38px;
  padding: 0 13px;
}

.marketing-shell {
  min-height: calc(100svh - 68px);
}

.hero-band {
  display: grid;
  grid-template-columns: minmax(280px, 0.62fr) minmax(460px, 1fr);
  align-items: center;
  gap: clamp(28px, 5vw, 62px);
  min-height: calc(100svh - 68px);
  padding: clamp(34px, 6vw, 68px) clamp(18px, 4vw, 48px) 34px;
}

.hero-copy {
  max-width: 620px;
  animation: rise 520ms ease both;
}

.eyebrow {
  display: inline-flex;
  margin-bottom: 14px;
  color: var(--accent-ink);
  font-size: 12px;
  font-weight: 780;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  max-width: 820px;
  font-family: "Aptos Display", "Aptos", "Segoe UI", sans-serif;
  font-size: clamp(36px, 5vw, 72px);
  line-height: 0.94;
  letter-spacing: -0.06em;
}

.toolbar h1 {
  font-size: clamp(26px, 3vw, 44px);
  letter-spacing: -0.045em;
}

p {
  margin: 16px 0 0;
  max-width: 650px;
  color: var(--muted);
  line-height: 1.62;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}

.product-preview,
.pricing-grid,
.compare-board,
.docs-board,
.signup-form,
.checkout-panel {
  min-width: 0;
  border: 1px solid var(--line);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: var(--shadow);
  animation: rise 620ms 80ms ease both;
}

.product-preview {
  padding: 14px;
}

.preview-head,
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 42px;
  color: var(--muted);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.preview-head strong,
.panel-head strong {
  color: var(--ink);
  font-size: 13px;
  letter-spacing: 0;
  text-transform: none;
}

.metric-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--line);
}

.metric-strip div {
  min-width: 0;
  padding: 16px;
  background: var(--surface);
}

.metric-strip span,
.metric-strip small {
  display: block;
  color: var(--muted);
  font-size: 12px;
}

.metric-strip strong {
  display: block;
  margin: 4px 0 2px;
  font-family: "Cascadia Mono", "SFMono-Regular", monospace;
  font-size: clamp(22px, 2.8vw, 34px);
  letter-spacing: -0.05em;
}

.preview-grid,
.analysis-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(260px, 0.6fr);
  gap: 12px;
  margin-top: 12px;
}

.panel,
.table-panel {
  min-width: 0;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--surface);
  padding: 14px;
}

.line-chart {
  width: 100%;
  height: auto;
  min-height: 160px;
}

.line-chart .grid {
  stroke: #e5e9ef;
  stroke-width: 1;
}

.line-chart .area {
  fill: url(#lineFill);
}

.line-chart .line {
  fill: none;
  stroke: var(--accent);
  stroke-width: 5;
  stroke-linecap: round;
}

.bar-chart {
  display: grid;
  grid-template-columns: repeat(9, minmax(10px, 1fr));
  align-items: end;
  gap: 7px;
  min-height: 182px;
  padding-top: 18px;
}

.bar-chart span {
  height: var(--h);
  min-height: 28px;
  border-radius: 8px 8px 3px 3px;
  background: linear-gradient(180deg, #174966, #0b6f6a);
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

th,
td {
  padding: 12px 10px;
  border-bottom: 1px solid #edf0f4;
  text-align: left;
  white-space: nowrap;
}

th {
  color: var(--muted);
  font-size: 11px;
  font-weight: 780;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

td {
  color: #273142;
  font-family: "Cascadia Mono", "SFMono-Regular", monospace;
}

.feature-band {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  margin: 0 clamp(18px, 4vw, 48px) 56px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 16px;
  background: var(--line);
}

.feature-band article {
  min-width: 0;
  padding: 24px;
  background: rgba(255, 255, 255, 0.72);
}

.feature-band span,
.pricing-grid span,
.docs-board span {
  color: var(--accent-ink);
  font-size: 12px;
  font-weight: 780;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.feature-band strong,
.pricing-grid strong,
.docs-board strong,
.compare-row strong {
  display: block;
  margin-top: 8px;
  font-size: 20px;
  letter-spacing: -0.03em;
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  overflow: hidden;
  background: var(--line);
}

.pricing-grid article {
  min-width: 0;
  padding: 24px;
  background: var(--surface);
}

.pricing-grid article.selected {
  background: #f2faf8;
}

.pricing-grid small {
  display: block;
  margin-top: 18px;
  color: var(--muted);
}

.compare-board,
.docs-board,
.signup-form,
.checkout-panel {
  padding: 18px;
}

.compare-row {
  display: grid;
  grid-template-columns: minmax(120px, 0.7fr) minmax(160px, 1fr) minmax(130px, 0.85fr);
  gap: 14px;
  align-items: center;
  min-height: 58px;
  border-bottom: 1px solid #edf0f4;
}

.compare-row.head {
  color: var(--muted);
  font-size: 12px;
  font-weight: 780;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.compare-row em {
  color: var(--muted);
  font-style: normal;
}

.docs-board {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.docs-board div {
  min-width: 0;
  padding: 20px;
  background: var(--surface);
}

.signup-form {
  display: grid;
  gap: 14px;
}

.signup-form label {
  display: grid;
  gap: 8px;
  color: var(--muted);
  font-size: 13px;
  font-weight: 680;
}

.signup-form input,
.signup-form select {
  min-height: 46px;
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  padding: 0 12px;
  background: var(--surface);
  color: var(--ink);
}

.checkout-panel {
  display: grid;
  gap: 1px;
  background: var(--line);
}

.invoice-line {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  padding: 18px;
  background: var(--surface);
}

.invoice-line.total strong {
  color: var(--accent-ink);
  font-size: 26px;
}

.workspace {
  display: grid;
  grid-template-columns: 246px minmax(0, 1fr);
  min-height: 100svh;
}

.sidebar {
  position: sticky;
  top: 0;
  align-self: start;
  display: grid;
  gap: 18px;
  height: 100svh;
  padding: 20px;
  border-right: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(18px);
}

.sidebar nav {
  align-items: stretch;
  flex-direction: column;
}

.side-note {
  align-self: end;
  display: grid;
  gap: 5px;
  border-top: 1px solid var(--line);
  padding-top: 16px;
  color: var(--muted);
  font-size: 12px;
}

.side-note strong {
  color: var(--ink);
}

.workpane {
  min-width: 0;
  padding: 24px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 18px;
}

.toolbar p {
  max-width: 760px;
}

.toolbar-actions,
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.toolbar-actions button,
.filter-bar button {
  min-height: 38px;
  border-color: var(--line);
  background: var(--surface);
  color: var(--muted);
  font-size: 13px;
}

.filter-bar {
  margin: 14px 0 0;
}

.workpane > .metric-strip,
.workpane > .analysis-grid,
.workpane > .table-panel {
  margin-top: 14px;
}

@keyframes rise {
  from { opacity: 0; transform: translateY(14px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.001ms !important;
    scroll-behavior: auto !important;
    transition-duration: 0.001ms !important;
  }
}

@media (max-width: 980px) {
  .topbar {
    grid-template-columns: 1fr auto;
  }

  .topbar nav {
    grid-column: 1 / -1;
    overflow: hidden;
    flex-wrap: wrap;
  }

  .hero-band,
  .workspace,
  .analysis-grid,
  .preview-grid,
  .feature-band {
    grid-template-columns: 1fr;
  }

  .hero-band {
    min-height: auto;
  }

  .sidebar {
    position: relative;
    height: auto;
    border-right: 0;
    border-bottom: 1px solid var(--line);
  }

  .sidebar nav {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .toolbar {
    display: grid;
  }
}

@media (max-width: 640px) {
  body {
    font-size: 14px;
  }

  .topbar {
    padding: 8px 14px;
  }

  .topbar nav a,
  .sidebar nav a {
    min-height: 36px;
    padding: 0 10px;
  }

  .hero-band,
  .workpane {
    padding: 22px 14px;
  }

  h1 {
    font-size: clamp(34px, 12vw, 48px);
    line-height: 0.98;
  }

  .metric-strip,
  .pricing-grid,
  .docs-board {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  th:nth-child(3),
  td:nth-child(3),
  th:nth-child(6),
  td:nth-child(6) {
    display: none;
  }

  th,
  td {
    white-space: normal;
    overflow-wrap: anywhere;
    padding: 10px 7px;
    font-size: 12px;
  }

  .compare-row {
    grid-template-columns: 1fr;
    gap: 4px;
    padding: 12px 0;
  }
}`;
}

function trackerJs() {
  return `const DEFAULT_SCRIPT_URL = "https://cloud.umami.is/script.js";
const storageKey = "simpletrack.research.profile";

window.simpleTrackLog = window.simpleTrackLog || [];

function params() {
  return new URLSearchParams(window.location.search);
}

function readStoredProfile() {
  try {
    return JSON.parse(window.localStorage.getItem(storageKey) || "{}");
  } catch {
    return {};
  }
}

function writeStoredProfile(profile) {
  window.localStorage.setItem(storageKey, JSON.stringify(profile));
}

function boolParam(name, fallback = false) {
  const value = params().get(name);
  if (value == null) return fallback;
  return ["1", "true", "yes", "on"].includes(value.toLowerCase());
}

function revenueForPlan(plan) {
  if (plan === "pro_annual") return 290;
  if (plan === "pro_monthly") return 29;
  return 0;
}

function profileFromUrl() {
  const query = params();
  const campaign = query.get("utm_campaign") || query.get("campaign") || "producthunt_launch";
  const source = query.get("utm_source") || campaign.split("_")[0] || "direct";
  const medium = query.get("utm_medium") || "referral";
  const plan = query.get("plan") || "trial";
  return {
    preset: "growth-baseline-x3",
    persona: query.get("persona") || query.get("user") || "browser-persona-preview",
    sessionIndex: Number(query.get("session") || "1"),
    plan,
    billingCycle: plan === "pro_annual" ? "annual" : plan === "pro_monthly" ? "monthly" : "none",
    campaign,
    source,
    medium,
    content: query.get("utm_content") || "simulation_site",
    term: query.get("utm_term") || "analytics_simulation",
    cohort: query.get("cohort") || "spring_launch",
    role: query.get("role") || "founder",
    workspaceSize: query.get("workspace_size") || "11-50",
    accountStage: query.get("stage") || "trial",
    logicalDay: Number(query.get("logical_day") || "42"),
    syntheticDate: query.get("synthetic_date") || new Date().toISOString().slice(0, 10),
    currency: "USD",
    revenue: revenueForPlan(plan),
  };
}

export function applySessionProfile(sessionProfile = {}) {
  const next = {
    ...readStoredProfile(),
    ...profileFromUrl(),
    ...window.simpleTrackSession,
    ...sessionProfile,
  };
  window.simpleTrackSession = next;
  writeStoredProfile(next);
  return next;
}

window.SimpleTrackBeforeSend = (type, payload) => {
  const profile = window.simpleTrackSession || applySessionProfile();
  return {
    ...payload,
    data: {
      ...(payload && payload.data ? payload.data : {}),
      ...profile,
      eventTransport: "browser",
      beforeSendType: type,
    },
  };
};

export function initTracker(options = {}) {
  const query = params();
  const stored = readStoredProfile();
  const websiteId = options.websiteId || query.get("websiteId") || query.get("website") || stored.websiteId;
  const scriptUrl = options.scriptUrl || query.get("scriptUrl") || stored.scriptUrl || DEFAULT_SCRIPT_URL;
  const recorderUrl = options.recorderUrl || query.get("recorderUrl") || stored.recorderUrl;
  const enablePerformance = options.enablePerformance ?? boolParam("performance", stored.enablePerformance ?? true);
  const enableReplays = options.enableReplays ?? boolParam("replays", stored.enableReplays ?? false);

  applySessionProfile({ websiteId, scriptUrl, recorderUrl, enablePerformance, enableReplays });

  if (!websiteId) {
    document.documentElement.dataset.tracker = "missing-website-id";
    window.simpleTrackLog.unshift({ at: new Date().toISOString(), type: "tracker", status: "missing website id" });
    return { loaded: false, websiteId, scriptUrl, enablePerformance, enableReplays };
  }

  if (!document.getElementById("umami-script")) {
    const script = document.createElement("script");
    script.id = "umami-script";
    script.defer = true;
    script.src = scriptUrl;
    script.dataset.websiteId = websiteId;
    script.dataset.beforeSend = "SimpleTrackBeforeSend";
    if (enablePerformance) script.dataset.performance = "true";
    document.head.appendChild(script);
  }

  if (enableReplays && recorderUrl && !document.getElementById("umami-recorder-script")) {
    const recorder = document.createElement("script");
    recorder.id = "umami-recorder-script";
    recorder.defer = true;
    recorder.src = recorderUrl;
    document.head.appendChild(recorder);
  }

  document.documentElement.dataset.tracker = "loading";
  return { loaded: true, websiteId, scriptUrl, enablePerformance, enableReplays };
}

export function identifyBusinessUser(profile = {}) {
  const session = applySessionProfile(profile);
  const id = profile.id || profile.persona || session.persona;
  const data = {
    distinctId: id,
    emailDomain: profile.emailDomain || "example.com",
    role: profile.role || session.role,
    workspaceSize: profile.workspaceSize || session.workspaceSize,
    plan: profile.plan || session.plan,
    cohort: profile.cohort || session.cohort,
    campaign: profile.campaign || session.campaign,
    accountStage: profile.accountStage || session.accountStage,
    preset: "growth-baseline-x3",
  };

  if (window.umami && typeof window.umami.identify === "function") {
    window.umami.identify(id, data);
  } else {
    window.simpleTrackLog.unshift({ at: new Date().toISOString(), type: "identify:queued", id, data });
  }
  return { id, data };
}

export function trackBusinessEvent(name, data = {}) {
  const session = applySessionProfile();
  const revenue = name === "checkout_completed" || name === "subscription_upgraded"
    ? revenueForPlan(data.plan || session.plan)
    : data.revenue;
  const payload = {
    ...session,
    ...data,
    ...(revenue ? { revenue, currency: data.currency || session.currency || "USD" } : {}),
    eventName: name,
    preset: "growth-baseline-x3",
  };

  if (window.umami && typeof window.umami.track === "function") {
    window.umami.track(name, payload);
  } else {
    window.simpleTrackLog.unshift({ at: new Date().toISOString(), type: "event:queued", name, payload });
  }
  return payload;
}

export function runPersonaStep(stepName, metadata = {}) {
  return trackBusinessEvent(stepName, {
    stepName,
    ...metadata,
  });
}

window.SimpleTrackResearch = {
  initTracker,
  applySessionProfile,
  identifyBusinessUser,
  trackBusinessEvent,
  runPersonaStep,
};

initTracker();`;
}

function mainJs() {
  return `import {
  applySessionProfile,
  identifyBusinessUser,
  trackBusinessEvent,
} from "./tracker.js";

function propsFromDataset(dataset) {
  const props = {};
  for (const [key, value] of Object.entries(dataset)) {
    if (key.startsWith("stProp")) {
      const propName = key.replace("stProp", "");
      props[propName.charAt(0).toLowerCase() + propName.slice(1)] = value;
    }
  }
  return props;
}

function pageDefaultProps() {
  return {
    page: document.body.dataset.page || "unknown",
    title: document.title,
    path: window.location.pathname,
  };
}

function decorateDynamicLabels() {
  const session = window.simpleTrackSession || {};
  const labels = {
    "[data-plan-label]": session.plan || "trial",
    "[data-campaign-label]": session.campaign || "producthunt_launch",
    "[data-cohort-label]": session.cohort || "spring_launch",
    "[data-revenue-label]": session.plan === "pro_annual" ? "$290 USD" : "$29 USD",
  };
  for (const [selector, value] of Object.entries(labels)) {
    const node = document.querySelector(selector);
    if (node) node.textContent = value;
  }
}

function trackElement(element) {
  const eventName = element.dataset.stEvent;
  if (!eventName) return;
  trackBusinessEvent(eventName, {
    ...pageDefaultProps(),
    label: element.textContent.trim().replace(/\\s+/g, " "),
    ...propsFromDataset(element.dataset),
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const session = applySessionProfile({
    page: document.body.dataset.page || "unknown",
  });
  decorateDynamicLabels();

  identifyBusinessUser({
    id: session.persona,
    persona: session.persona,
    plan: session.plan,
    campaign: session.campaign,
    cohort: session.cohort,
    role: session.role,
    workspaceSize: session.workspaceSize,
    accountStage: session.accountStage,
  });

  const pageEvent = document.body.dataset.pageEvent;
  if (pageEvent) {
    trackBusinessEvent(pageEvent, pageDefaultProps());
  }

  document.querySelectorAll("[data-st-event]").forEach((element) => {
    element.addEventListener("click", (event) => {
      trackElement(element);
      if (element instanceof HTMLAnchorElement && element.href && element.origin === window.location.origin) {
        event.preventDefault();
        window.setTimeout(() => {
          window.location.href = element.href;
        }, 80);
      }
    });
  });

  document.querySelectorAll("form[data-st-event]").forEach((form) => {
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      trackElement(form);
      window.setTimeout(() => {
        window.location.href = new URL("signup-success.html", window.location.href).href;
      }, 120);
    });
  });
});`;
}

function debugHtml() {
  return `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>SimpleTrack Umami Debug</title>
    <link rel="stylesheet" href="./site/styles.css" />
    <script type="module" src="./debug.js"></script>
  </head>
  <body data-page="debug" data-page-event="debug_viewed">
    <header class="topbar">
      <a class="brand" href="./site/index.html"><span></span>SimpleTrack</a>
      <nav>
        <a href="./site/index.html">SaaS site</a>
        <a aria-current="page" href="./debug.html">Debug</a>
        <a href="./index.html">Legacy lab</a>
      </nav>
    </header>
    <main class="marketing-shell">
      <section class="hero-band">
        <div class="hero-copy">
          <span class="eyebrow">Umami Cloud research</span>
          <h1>Tracker debug lab</h1>
          <p>验证 initTracker、identifyBusinessUser、trackBusinessEvent 和 runPersonaStep。website id 只通过页面输入或 URL 参数传入，不写入仓库。</p>
          <form class="signup-form" id="debug-form">
            <label>Website ID<input id="website-id" name="websiteId" placeholder="paste website id for local verification" autocomplete="off" /></label>
            <label>Script URL<input id="script-url" name="scriptUrl" value="https://cloud.umami.is/script.js" /></label>
            <label>Persona<input id="persona-id" name="persona" value="debug-founder-001" /></label>
            <button class="button primary" type="submit">Initialize tracker</button>
          </form>
          <div class="action-row">
            <button class="button secondary" id="identify-button" type="button">identify()</button>
            <button class="button secondary" id="track-button" type="button">track()</button>
            <a class="button primary" href="./site/index.html">Open SaaS simulation</a>
          </div>
        </div>
        <div class="product-preview">
          <div class="preview-head"><span>Debug output</span><strong>queued calls appear here without website id</strong></div>
          <pre class="debug-log" id="debug-log">Waiting for interaction...</pre>
        </div>
      </section>
    </main>
  </body>
</html>`;
}

function debugJs() {
  return `import {
  initTracker,
  applySessionProfile,
  identifyBusinessUser,
  trackBusinessEvent,
  runPersonaStep,
} from "./site/tracker.js";

const log = document.getElementById("debug-log");
const websiteInput = document.getElementById("website-id");
const scriptInput = document.getElementById("script-url");
const personaInput = document.getElementById("persona-id");

function render(message, value) {
  const payload = value ? JSON.stringify(value, null, 2) : "";
  log.textContent = [new Date().toISOString(), message, payload].filter(Boolean).join("\\n");
}

document.getElementById("debug-form").addEventListener("submit", (event) => {
  event.preventDefault();
  const result = initTracker({
    websiteId: websiteInput.value.trim(),
    scriptUrl: scriptInput.value.trim() || "https://cloud.umami.is/script.js",
    enablePerformance: true,
  });
  applySessionProfile({
    persona: personaInput.value.trim() || "debug-founder-001",
    campaign: "debug_manual",
    cohort: "spring_launch",
    plan: "trial",
    accountStage: "trial",
  });
  render("initTracker()", result);
});

document.getElementById("identify-button").addEventListener("click", () => {
  const result = identifyBusinessUser({
    id: personaInput.value.trim() || "debug-founder-001",
    role: "founder",
    workspaceSize: "11-50",
    plan: "trial",
  });
  render("identifyBusinessUser()", result);
});

document.getElementById("track-button").addEventListener("click", () => {
  const result = trackBusinessEvent("filter_applied", {
    field: "campaign",
    campaign: "debug_manual",
    source: "debug",
  });
  runPersonaStep("dashboard_viewed", { debug: true });
  render("trackBusinessEvent() + runPersonaStep()", result);
});`;
}

function trackingReadme() {
  return md([
    "# Umami Tracking Demo",
    "",
    "这个目录现在包含两层验证资产：",
    "",
    "- `debug.html` / `index.html`：保留 tracker、track()、identify() 的单点调试能力。",
    "- `site/`：高品质 SimpleTrack SaaS 仿真站，通过真实页面跳转、按钮点击和业务事件喂入 Umami Cloud。",
    "",
    "## Files",
    "",
    "- `site/`：多页 SaaS 仿真站。",
    "- `site/tracker.js`：共享上报 helper，暴露 `initTracker`、`identifyBusinessUser`、`trackBusinessEvent`、`applySessionProfile`、`runPersonaStep`。",
    "- `site/main.js`：页面识别、点击事件绑定、persona 初始化和表单跳转。",
    "- `debug.html` / `debug.js`：单独验证 tracker、track()、identify()。",
    "- `bulk-send.mjs`：`growth-baseline-x3` 批量事件发送脚本。",
    "- `run-browser-flows.mjs`：使用真实浏览器跑 72 个 persona 的三段式流量。",
    "- `send-event.mjs`：单事件服务端发送脚本。",
    "- `generate-site.mjs`：仿真站和归档文档生成器。",
    "",
    "## Start",
    "",
    "```powershell",
    "cd C:\\Users\\admin\\Documents\\src\\uba\\simpletrack\\docs\\umami\\tracking-demo",
    "python -m http.server 4173",
    "```",
    "",
    "打开：",
    "",
    "```text",
    "http://localhost:4173/site/index.html?websiteId=YOUR_WEBSITE_ID&utm_source=producthunt&utm_medium=launch&utm_campaign=producthunt_launch&cohort=spring_launch&plan=trial&performance=1",
    "```",
    "",
    "## Growth Baseline X3",
    "",
    `- ${sample.totalUsers} 个逻辑用户`,
    `- ${sample.totalSessions} 个 session`,
    `- ${sample.browserUsers} 个真实浏览器 persona`,
    `- ${sample.batchUsers} 个批量事件用户`,
    "- 每 session 使用 8/10/12 个事件模板之一",
    `- ${sample.campaigns.length} 组 campaign、${sample.cohorts.length} 组 cohort、${sample.plans.length} 个 plan`,
    `- ${sample.paidWorkspaces} 条收入转化，其中浏览器 ${sample.browserPaid} 条，批量 ${sample.batchPaid} 条`,
    "",
    "## Browser Flows",
    "",
    "```powershell",
    "npx --yes -p playwright node run-browser-flows.mjs --website-id YOUR_WEBSITE_ID --base-url http://localhost:4173/site",
    "```",
    "",
    "## Bulk Send",
    "",
    "先 dry run：",
    "",
    "```powershell",
    "node bulk-send.mjs --website-id YOUR_WEBSITE_ID --dry-run",
    "```",
    "",
    "真实发送：",
    "",
    "```powershell",
    "node bulk-send.mjs --website-id YOUR_WEBSITE_ID --hostname localhost --base-url http://localhost:4173/site",
    "```",
    "",
    "## Notes",
    "",
    "- 不要把账号、cookie、token 写入仓库。",
    "- `website id` 只通过 URL 参数、localStorage 或命令行参数传入。",
    "- 公开 `/api/send` 不承诺支持历史时间戳回填；42 天跨度使用 `logical_day`、`synthetic_date`、`cohort` 等业务属性表达。",
  ]);
}

function docRealData() {
  return md([
    "# 真实业务数据方案",
    "",
    "## 目标",
    "",
    "`tracking-demo/site/` 不是资讯站，也不是普通 demo 单页，而是 SimpleTrack 自己的 SaaS 产品站加产品内工作台。它用于向 Umami Cloud 注入接近真实早期增长型 SaaS 的页面、事件、来源、收入、cohort 和过滤字段。",
    "",
    "## 默认样本",
    "",
    "| 项目 | 数值 |",
    "| --- | --- |",
    `| 数据预设 | ${sample.preset} |`,
    `| 逻辑用户 | ${sample.totalUsers} |`,
    `| 真实浏览器用户 | ${sample.browserUsers} |`,
    `| 批量事件用户 | ${sample.batchUsers} |`,
    `| 每用户 session | ${sample.sessionsPerUser} |`,
    `| 总 session | ${sample.totalSessions} |`,
    "| 每 session 事件模板 | 8 / 10 / 12 |",
    `| 预计业务事件量 | ${sample.eventRange}，默认约 ${sample.expectedEvents.toLocaleString("en-US")} |`,
    `| 逻辑时间跨度 | ${sample.logicalDays} 天 |`,
    `| 付费 workspace | ${sample.paidWorkspaces} |`,
    "",
    "用户分布：324 个顶部漏斗访客、162 个注册未激活用户、108 个 trial 激活用户、54 个付费用户。",
    "",
    "收入分布：54 个付费 workspace，其中 36 个 `pro_monthly`，18 个 `pro_annual`。浏览器流量默认贡献 6 个真实 checkout 转化，批量事件贡献 48 个收入转化，合计 54。",
    "",
    "说明：Umami Cloud 的公开 `/api/send` 接口不承诺支持回填历史时间戳，所以 42 天跨度用 `logical_day`、`synthetic_date`、`cohort` 等业务属性表达。Cloud 的真实入库时间仍以请求到达时间为准。",
    "",
    "## 页面语义",
    "",
    "| 页面组 | 页面 | 用途 |",
    "| --- | --- | --- |",
    "| 获客 | `site/index.html`、`pricing.html`、`compare.html`、`docs.html` | UTM、Compare、Breakdown、Attribution |",
    "| 转化 | `signup.html`、`signup-success.html` | Goals、Funnels、Journeys |",
    "| 激活 | `install.html`、`install-web.html`、`install-verify.html` | activation、first value、Retention |",
    "| 产品内 | `app/dashboard.html`、`app/events.html`、`app/funnels.html`、`app/segments.html`、`app/cohorts.html` | Sessions、Filter、Segments、Cohorts |",
    "| 收入 | `checkout.html`、`checkout-success.html`、`billing.html` | Revenue、Attribution、Goals |",
    "",
    "## 固定维度",
    "",
    `- Plan: ${sample.plans.map((item) => `\`${item}\``).join(" / ")}`,
    `- Campaign: ${sample.campaigns.map((item) => `\`${item}\``).join(" / ")}`,
    `- Cohort: ${sample.cohorts.map((item) => `\`${item}\``).join(" / ")}`,
    "- Currency: `USD`",
    "",
    "## 固定事件",
    "",
    businessEvents.map((item) => `\`${item}\``).join("、"),
    "",
    "## 收入口径",
    "",
    "`checkout_completed` 必带 `revenue`、`currency=USD`、`plan`、`campaign`、`cohort`。Revenue 与 Attribution 复验时优先按 campaign、source、medium、plan 和 cohort 拆分。",
  ]);
}

function docDesign() {
  return md([
    "# 高品质仿真站设计规范",
    "",
    "## Visual Thesis",
    "",
    "SimpleTrack 仿真站采用浅色、克制、专业的分析工作台风格：高对比文字、紧凑指标、强表格和图表层级，接近成熟分析类 SaaS，而不是旧原型式普通后台。",
    "",
    "## 信息架构原则",
    "",
    "- 不复用 `simpletrack/prototype/simpletrack-umami-inspired/` 的页面壳、视觉、布局或信息架构。",
    "- 公共页保留产品级质感，用真实产品界面截图式布局承接转化，不做夸张营销 hero。",
    "- 产品内页第一屏就是可操作分析界面：过滤器、指标、趋势、细分表、事件流。",
    "- 所有页面必须是静态多页跳转，不依赖 SPA 路由，以便触发真实 pageview 和 session。",
    "",
    "## 视觉原则",
    "",
    "- 主体为浅色高对比界面，背景保持安静，少用装饰性渐变。",
    "- 色彩以黑白灰、青绿和深蓝为功能色，不使用大面积紫色或暗色。",
    "- 半径控制在 8px 到 16px，少用厚阴影，卡片只服务于图表、表格和工具区。",
    "- 数字使用等宽字体，表格必须有表头、分隔线和清晰扫描路径。",
    "- 移动端隐藏次要列，保证页面不出现整体横向滚动。",
    "",
    "## 组件规范",
    "",
    "| 组件 | 用法 |",
    "| --- | --- |",
    "| Topbar | 公共页主导航，保持 4 到 5 个入口 |",
    "| Sidebar | 产品内导航，覆盖 Dashboard / Events / Funnels / Segments / Cohorts / Billing |",
    "| Metric strip | 关键指标区，桌面四列，移动两列 |",
    "| Panel | 仅用于图表、表格、真实工具区，不做泛卡片堆砌 |",
    "| Table | 必须使用清晰表头、分隔线和 tabular 数字 |",
    "| Button | 所有主要动作带 `data-st-event`，同时保持 44px 以上点击高度 |",
    "",
    "## Interaction Thesis",
    "",
    "- 页面真实跳转优先，让 Umami pageview 和 session 自然形成。",
    "- CTA 点击先上报业务事件，再通过普通链接完成跳转。",
    "- Filter、Segment、Cohort 按钮是轻交互，重点是稳定生成可分析事件和属性。",
    "- 支持 `prefers-reduced-motion`，动效只用于进入层级和点击反馈。",
  ]);
}

function docMatrix() {
  return md([
    "# 功能打通矩阵",
    "",
    "| 功能 | 数据样本 | 数据来源 | 截图编号 | 当前状态 |",
    "| --- | --- | --- | --- | --- |",
    ...featureMatrix.map(([feature, sampleText, source, shot]) => `| ${feature} | ${sampleText} | ${source} | ${shot} | 待 Cloud 复验和截图 |`),
    "",
    "## 状态说明",
    "",
    "- 本地仿真站、浏览器流量脚本和批量事件脚本已经提供数据来源。",
    "- Cloud 侧仍需要真实 Umami Cloud website id 执行一轮并截图，当前不能写成已完成。",
    "- 如果账号没有 Replays、Performance、Revenue 或 Attribution 入口，必须在阶段 README 和功能分析中记录限制与证据。",
  ]);
}

function docRunbook() {
  return md([
    "# 执行与复验手册",
    "",
    "## 1. 启动本地站点",
    "",
    "```powershell",
    "cd C:\\Users\\admin\\Documents\\src\\uba\\simpletrack\\docs\\umami\\tracking-demo",
    "python -m http.server 4173",
    "```",
    "",
    "入口：",
    "",
    "- 调试页：`http://localhost:4173/debug.html`",
    "- SaaS 仿真站：`http://localhost:4173/site/index.html?websiteId=YOUR_WEBSITE_ID&utm_source=producthunt&utm_medium=launch&utm_campaign=producthunt_launch&cohort=spring_launch&plan=trial&performance=1`",
    "",
    "## 2. 跑真实浏览器流量",
    "",
    "```powershell",
    "npx --yes -p playwright node run-browser-flows.mjs --website-id YOUR_WEBSITE_ID --base-url http://localhost:4173/site",
    "```",
    "",
    "默认会跑 72 个 persona，每人 3 个独立 browser context，对应 216 个真实浏览器 session。其中 6 个 paid persona 会走到 checkout-success。",
    "",
    "## 3. 批量灌入业务事件",
    "",
    "先 dry run：",
    "",
    "```powershell",
    "node bulk-send.mjs --website-id YOUR_WEBSITE_ID --dry-run",
    "```",
    "",
    "真实发送：",
    "",
    "```powershell",
    "node bulk-send.mjs --website-id YOUR_WEBSITE_ID --hostname localhost --base-url http://localhost:4173/site",
    "```",
    "",
    "默认批量脚本会生成 576 个用户、1728 个 session、约 17,280 条业务事件和 48 条收入转化。与浏览器脚本合计后达到 648 用户、1944 session、约 19,440 条业务事件和 54 条收入转化。",
    "",
    "## 4. Cloud 页面复验顺序",
    "",
    "1. Realtime：确认当前 pageview 和 CTA 事件进入。",
    "2. Sessions：确认 persona 会话和页面路径形成。",
    "3. Performance：确认启用 performance 参数后的页面性能视图是否有数据。",
    "4. Events / Properties：确认事件名和 plan、campaign、cohort、role、workspaceSize 等属性。",
    "5. Compare / Breakdown / Filter：按 plan、campaign、cohort、source、medium 拆分。",
    "6. Goals / Funnels / Journeys：配置 signup、first event、checkout 路径。",
    "7. Segments / Cohorts / Retention：检查命名切片和三组 cohort。",
    "8. UTM / Revenue / Attribution：验证 campaign 到 checkout revenue 的链路。",
    "9. Replays：按账号能力确认真实浏览器 session 是否可回放。",
    "",
    "## 5. 截图采集顺序",
    "",
    "- Phase 07：Sessions、Realtime、Performance、Compare、Breakdown、Goals、Filter。",
    "- Phase 08：Funnels、Journeys、Retention、Replays、Segments、Cohorts、UTM、Revenue、Attribution。",
    "",
    "每张截图新增后必须同步更新 `快照索引.md`、对应 phase 的 `flow.md`、对应 phase 的 `README.md`、`快照进度.md` 和 `Umami功能深度分析.md`。",
  ]);
}

function phaseReadme(title, scope, shots) {
  return md([
    `# ${title}`,
    "",
    "## 目标",
    "",
    scope,
    "",
    "## 截图清单",
    "",
    "| 编号 | 页面 | 说明 | 当前状态 |",
    "| --- | --- | --- | --- |",
    ...shots.map(([id, page, note]) => `| ${id} | ${page} | ${note} | 待 Cloud 复验和截图 |`),
    "",
    "## 备注",
    "",
    "本阶段先归档计划和操作流。正式截图需要在真实 Umami Cloud website id、浏览器流量和批量事件都跑完后补齐。若 Cloud 账号缺少对应能力，必须记录限制和已验证证据。",
  ]);
}

function phaseFlow(title, steps) {
  return md([
    `# ${title} Flow`,
    "",
    ...steps.map((step, index) => `${index + 1}. ${step}`),
  ]);
}

for (const page of pages) {
  write(`site/${page.path}`, html(page));
}

write("site/styles.css", styles());
write("site/tracker.js", trackerJs());
write("site/main.js", mainJs());
write("debug.html", debugHtml());
write("debug.js", debugJs());
write("README.md", trackingReadme());

write("../docs/真实业务数据方案.md", docRealData());
write("../docs/高品质仿真站设计规范.md", docDesign());
write("../docs/功能打通矩阵.md", docMatrix());
write("../docs/执行与复验手册.md", docRunbook());

write(
  "../snapshots/phase-07-traffic-and-behavior-insights/README.md",
  phaseReadme("Phase 07: Traffic And Behavior Insights", "记录 Sessions、RealTime、Performance、Compare、BreakDown、Goals、Filter 在增长基线三倍样本下的页面证据。", [
    ["P07-S01", "Sessions", "真实浏览器 persona 形成会话列表"],
    ["P07-S02", "Realtime", "浏览器流量和 CTA 事件实时出现"],
    ["P07-S03", "Performance", "页面加载性能数据进入视图"],
    ["P07-S04", "Compare", "按 plan、campaign、cohort 对比"],
    ["P07-S05", "BreakDown", "按来源、计划、cohort 拆分"],
    ["P07-S06", "Goals", "signup、first event、checkout 目标"],
    ["P07-S07", "Filter", "Fields、Segments、Cohorts 过滤入口"],
  ]),
);

write(
  "../snapshots/phase-07-traffic-and-behavior-insights/flow.md",
  phaseFlow("Phase 07", [
    "启动 tracking-demo 静态服务并打开 site/index.html。",
    "携带 websiteId、UTM、plan、cohort、performance 参数运行 72 个浏览器 persona。",
    "在 Umami Cloud 中依次打开 Sessions、Realtime、Performance。",
    "运行 growth-baseline-x3 批量事件后打开 Compare、Breakdown、Goals、Filter。",
    "记录每个页面是否进入有数据状态，以及是否需要额外 Cloud 配置。",
  ]),
);

write(
  "../snapshots/phase-08-growth-and-monetization-insights/README.md",
  phaseReadme("Phase 08: Growth And Monetization Insights", "记录 Funnels、Journeys、Retention、Replays、Segments、Cohorts、UTM、Revenue、Attribution 在三倍样本下的页面证据。", [
    ["P08-S01", "Funnels", "pricing -> signup -> install -> first_event -> checkout"],
    ["P08-S02", "Journeys", "高意图流量路径图"],
    ["P08-S03", "Retention", "三组 cohort 的回访表现"],
    ["P08-S04", "Replays", "真实浏览器会话回放"],
    ["P08-S05", "Segments", "命名流量切片"],
    ["P08-S06", "Cohorts", "spring_launch / self_serve_wave / paid_pilot"],
    ["P08-S07", "UTM", "6 组 campaign 的流量表现"],
    ["P08-S08", "Revenue", "54 条收入转化"],
    ["P08-S09", "Attribution", "campaign 到 checkout revenue 的归因"],
  ]),
);

write(
  "../snapshots/phase-08-growth-and-monetization-insights/flow.md",
  phaseFlow("Phase 08", [
    "先跑浏览器 persona，保证 pageview、session、replay、performance 有真实入口。",
    "再跑 growth-baseline-x3 批量事件，补足高密度转化、收入和属性样本。",
    "配置或打开 Funnels、Goals、Segments、Cohorts 等需要对象的页面。",
    "按 UTM -> signup -> install -> first event -> checkout 顺序复验 Revenue 和 Attribution。",
    "把受 Cloud 账号能力限制的项目明确记录为限制，不写成已完成。",
  ]),
);

console.log("Generated SimpleTrack Umami SaaS simulation site and documentation.");
