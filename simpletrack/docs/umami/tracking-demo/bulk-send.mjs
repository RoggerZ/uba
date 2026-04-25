const args = process.argv.slice(2);

function readFlag(name, fallback = "") {
  const index = args.findIndex((arg) => arg === `--${name}`);
  if (index === -1 || index === args.length - 1) {
    return fallback;
  }
  return args[index + 1];
}

function hasFlag(name) {
  return args.includes(`--${name}`);
}

const config = {
  preset: readFlag("preset", "growth-baseline-x3"),
  website: readFlag("website-id"),
  hostname: readFlag("hostname", "localhost"),
  baseUrl: readFlag("base-url", "http://localhost:4173/site").replace(/\/$/, ""),
  apiUrl: readFlag("api-url", "https://api-gateway.umami.dev/api/send"),
  users: Number(readFlag("users", "576")),
  sessionsPerUser: Number(readFlag("sessions-per-user", "3")),
  dryRun: hasFlag("dry-run"),
  concurrency: Math.max(1, Number(readFlag("concurrency", "8"))),
  limit: Number(readFlag("limit", "0")),
  timeoutMs: Math.max(1000, Number(readFlag("timeout-ms", "12000"))),
  retries: Math.max(0, Number(readFlag("retries", "2"))),
  userAgent: readFlag(
    "user-agent",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
  ),
  runId: readFlag("run-id", `growth-x3-${new Date().toISOString().replace(/[:.]/g, "-")}`),
};

if (config.preset !== "growth-baseline-x3") {
  console.error(`Unsupported preset: ${config.preset}`);
  process.exit(1);
}

if (!config.website) {
  console.error("Missing required flag: --website-id");
  process.exit(1);
}

const campaigns = [
  { campaign: "producthunt_launch", source: "producthunt", medium: "launch", referrer: "https://www.producthunt.com/" },
  { campaign: "google_brand", source: "google", medium: "cpc", referrer: "https://www.google.com/" },
  { campaign: "google_competitor", source: "google", medium: "cpc", referrer: "https://www.google.com/search?q=mixpanel+alternative" },
  { campaign: "docs_seo", source: "google", medium: "organic", referrer: "https://www.google.com/search?q=product+analytics+docs" },
  { campaign: "linkedin_founder", source: "linkedin", medium: "social", referrer: "https://www.linkedin.com/" },
  { campaign: "email_nurture", source: "email", medium: "lifecycle", referrer: "https://mail.google.com/" },
];

const cohorts = ["spring_launch", "self_serve_wave", "paid_pilot"];
const roles = ["founder", "growth_lead", "engineer", "product_manager", "ops"];
const workspaceSizes = ["1-10", "11-50", "51-200", "201-500"];
const screens = ["1440x900", "1536x960", "1728x1117", "390x844", "430x932"];

const pages = {
  index: { path: "/index.html", title: "Growth Console | SimpleTrack" },
  pricing: { path: "/pricing.html", title: "Pricing | SimpleTrack" },
  compare: { path: "/compare.html", title: "Compare | SimpleTrack" },
  docs: { path: "/docs.html", title: "Docs | SimpleTrack" },
  signup: { path: "/signup.html", title: "Create Workspace | SimpleTrack" },
  signupSuccess: { path: "/signup-success.html", title: "Workspace Created | SimpleTrack" },
  install: { path: "/install.html", title: "Install | SimpleTrack" },
  installWeb: { path: "/install-web.html", title: "Web Install | SimpleTrack" },
  installVerify: { path: "/install-verify.html", title: "Install Verified | SimpleTrack" },
  dashboard: { path: "/app/dashboard.html", title: "Dashboard | SimpleTrack" },
  events: { path: "/app/events.html", title: "Events | SimpleTrack" },
  funnels: { path: "/app/funnels.html", title: "Funnels | SimpleTrack" },
  segments: { path: "/app/segments.html", title: "Segments | SimpleTrack" },
  cohorts: { path: "/app/cohorts.html", title: "Cohorts | SimpleTrack" },
  checkout: { path: "/checkout.html", title: "Checkout | SimpleTrack" },
  checkoutSuccess: { path: "/checkout-success.html", title: "Checkout Complete | SimpleTrack" },
  billing: { path: "/billing.html", title: "Billing | SimpleTrack" },
};

const sessionTemplates = [
  [
    ["pricing_viewed", "pricing"],
    ["compare_opened", "compare"],
    ["signup_started", "signup"],
    ["signup_completed", "signup"],
    ["workspace_created", "signupSuccess"],
    ["install_started", "install"],
    ["sdk_install_completed", "installWeb"],
    ["first_event_sent", "installVerify"],
  ],
  [
    ["dashboard_viewed", "dashboard"],
    ["filter_applied", "dashboard"],
    ["segment_opened", "segments"],
    ["cohort_opened", "cohorts"],
    ["dashboard_viewed", "events"],
    ["filter_applied", "events"],
    ["dashboard_viewed", "funnels"],
    ["filter_applied", "funnels"],
    ["billing_viewed", "billing"],
    ["subscription_upgraded", "billing"],
  ],
  [
    ["pricing_viewed", "pricing"],
    ["checkout_started", "checkout"],
    ["checkout_completed", "checkoutSuccess"],
    ["billing_viewed", "billing"],
    ["dashboard_viewed", "dashboard"],
    ["filter_applied", "dashboard"],
    ["segment_opened", "segments"],
    ["cohort_opened", "cohorts"],
    ["dashboard_viewed", "events"],
    ["filter_applied", "events"],
    ["dashboard_viewed", "funnels"],
    ["billing_viewed", "billing"],
  ],
];

const nonPaidMonetizationTemplate = [
  ["pricing_viewed", "pricing"],
  ["compare_opened", "compare"],
  ["dashboard_viewed", "dashboard"],
  ["filter_applied", "dashboard"],
  ["segment_opened", "segments"],
  ["cohort_opened", "cohorts"],
  ["dashboard_viewed", "events"],
  ["filter_applied", "events"],
  ["dashboard_viewed", "funnels"],
  ["filter_applied", "funnels"],
  ["install_started", "install"],
  ["first_event_sent", "installVerify"],
];

const nonPaidActivationTemplate = [
  ["dashboard_viewed", "dashboard"],
  ["filter_applied", "dashboard"],
  ["segment_opened", "segments"],
  ["cohort_opened", "cohorts"],
  ["dashboard_viewed", "events"],
  ["filter_applied", "events"],
  ["dashboard_viewed", "funnels"],
  ["filter_applied", "funnels"],
  ["billing_viewed", "billing"],
  ["pricing_viewed", "pricing"],
];

function pad(number, size) {
  return String(number).padStart(size, "0");
}

function accountStage(userIndex) {
  if (userIndex < 288) return "top_funnel";
  if (userIndex < 432) return "registered";
  if (userIndex < 528) return "activated_trial";
  return "paid";
}

function planForStage(stage, paidIndex) {
  if (stage === "top_funnel") return "free";
  if (stage === "registered" || stage === "activated_trial") return "trial";
  return paidIndex < 30 ? "pro_monthly" : "pro_annual";
}

function revenueForPlan(plan) {
  if (plan === "pro_annual") return 290;
  if (plan === "pro_monthly") return 29;
  return 0;
}

function syntheticDate(logicalDay) {
  const date = new Date(Date.UTC(2026, 2, 15));
  date.setUTCDate(date.getUTCDate() + logicalDay - 1);
  return date.toISOString().slice(0, 10);
}

function pageUrl(page, profile) {
  const params = new URLSearchParams({
    utm_source: profile.source,
    utm_medium: profile.medium,
    utm_campaign: profile.campaign,
    cohort: profile.cohort,
    plan: profile.plan,
    persona: profile.distinctId,
    session: String(profile.sessionNumber),
    logical_day: String(profile.logicalDay),
    synthetic_date: profile.syntheticDate,
  });
  return `${config.baseUrl}${page.path}?${params.toString()}`;
}

function eventData(eventName, profile, eventIndex, sessionEventCount) {
  const isRevenue = eventName === "checkout_completed" || eventName === "subscription_upgraded";
  return {
    preset: config.preset,
    distinctId: profile.distinctId,
    persona: profile.distinctId,
    syntheticUserIndex: profile.syntheticUserIndex,
    sessionId: profile.sessionId,
    sessionIndex: profile.sessionIndex,
    sessionNumber: profile.sessionNumber,
    sessionEventCount,
    eventIndex,
    accountStage: profile.accountStage,
    plan: profile.plan,
    billing_cycle: profile.plan === "pro_annual" ? "annual" : profile.plan === "pro_monthly" ? "monthly" : "none",
    role: profile.role,
    workspaceSize: profile.workspaceSize,
    campaign: profile.campaign,
    source: profile.source,
    medium: profile.medium,
    content: profile.content,
    term: profile.term,
    cohort: profile.cohort,
    logical_day: profile.logicalDay,
    synthetic_date: profile.syntheticDate,
    goal_name: eventName,
    segment: profile.segment,
    journey: profile.journey,
    attribution_model: "last_non_direct",
    eventTransport: "server-batch",
    run_id: config.runId,
    ...(isRevenue ? { revenue: revenueForPlan(profile.plan), currency: "USD" } : {}),
  };
}

function buildPayloads() {
  const payloads = [];
  let paidIndex = 0;

  for (let userIndex = 0; userIndex < config.users; userIndex += 1) {
    const stage = accountStage(userIndex);
    const userPaidIndex = stage === "paid" ? paidIndex++ : -1;
    const plan = planForStage(stage, userPaidIndex);
    const traffic = campaigns[userIndex % campaigns.length];
    const cohort = cohorts[userIndex % cohorts.length];
    const distinctId = `batch-user-${pad(userIndex + 1, 4)}`;

    for (let sessionIndex = 0; sessionIndex < config.sessionsPerUser; sessionIndex += 1) {
      const template = stage === "paid"
        ? sessionTemplates[sessionIndex]
        : sessionIndex === 0
          ? sessionTemplates[0]
          : sessionIndex === 1
            ? nonPaidActivationTemplate
            : nonPaidMonetizationTemplate;
      const logicalDay = ((userIndex * config.sessionsPerUser + sessionIndex) % 42) + 1;
      const profile = {
        ...traffic,
        preset: config.preset,
        distinctId,
        syntheticUserIndex: userIndex + 1,
        sessionIndex,
        sessionNumber: sessionIndex + 1,
        sessionId: `${distinctId}-s${sessionIndex + 1}`,
        accountStage: stage,
        plan,
        cohort,
        role: roles[(userIndex + sessionIndex) % roles.length],
        workspaceSize: workspaceSizes[(userIndex + sessionIndex) % workspaceSizes.length],
        content: sessionIndex === 0 ? "landing_console" : sessionIndex === 1 ? "activation_workspace" : "monetization_review",
        term: stage === "paid" ? "analytics_attribution" : "product_analytics",
        logicalDay,
        syntheticDate: syntheticDate(logicalDay),
        segment: stage === "paid" ? "paid_workspaces" : stage === "activated_trial" ? "activated_trials" : "self_serve_visitors",
        journey: sessionIndex === 0 ? "acquisition" : sessionIndex === 1 ? "activation" : "monetization",
      };

      template.forEach(([eventName, pageKey], eventIndex) => {
        const page = pages[pageKey];
        payloads.push({
          type: "event",
          payload: {
            website: config.website,
            hostname: config.hostname,
            language: "en-US",
            referrer: eventIndex === 0 ? traffic.referrer : pageUrl(pages[template[Math.max(0, eventIndex - 1)][1]], profile),
            screen: screens[(userIndex + sessionIndex) % screens.length],
            title: page.title,
            url: pageUrl(page, profile),
            tag: distinctId,
            id: profile.sessionId,
            name: eventName,
            data: eventData(eventName, profile, eventIndex, template.length),
          },
        });
      });
    }
  }

  return config.limit > 0 ? payloads.slice(0, config.limit) : payloads;
}

async function send(payload) {
  let lastError;
  for (let attempt = 0; attempt <= config.retries; attempt += 1) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), config.timeoutMs);
    try {
      const response = await fetch(config.apiUrl, {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "user-agent": config.userAgent,
        },
        body: JSON.stringify(payload),
        signal: controller.signal,
      });

      const text = await response.text();
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${text}`);
      }
      return text;
    } catch (error) {
      lastError = error;
      if (attempt < config.retries) {
        await new Promise((resolve) => setTimeout(resolve, 350 * (attempt + 1)));
      }
    } finally {
      clearTimeout(timeout);
    }
  }
  throw lastError;
}

function summarize(payloads) {
  const eventCounts = new Map();
  const plans = new Map();
  let revenueEvents = 0;
  let revenue = 0;

  for (const item of payloads) {
    const eventName = item.payload.name;
    eventCounts.set(eventName, (eventCounts.get(eventName) || 0) + 1);
    const plan = item.payload.data.plan;
    plans.set(plan, (plans.get(plan) || 0) + 1);
    if (eventName === "checkout_completed") {
      revenueEvents += 1;
      revenue += item.payload.data.revenue || 0;
    }
  }

  return {
    preset: config.preset,
    runId: config.runId,
    batchUsers: config.users,
    sessions: config.users * config.sessionsPerUser,
    events: payloads.length,
    expectedWithBrowserEvents: payloads.length + 2160,
    expectedWithBrowserUsers: config.users + 72,
    expectedWithBrowserSessions: config.users * config.sessionsPerUser + 216,
    paidConversionsInBatch: revenueEvents,
    expectedPaidConversionsWithBrowser: revenueEvents + 6,
    revenueInBatch: revenue,
    eventCounts: Object.fromEntries([...eventCounts.entries()].sort()),
    planEventCounts: Object.fromEntries([...plans.entries()].sort()),
  };
}

async function runQueue(payloads) {
  let cursor = 0;
  let accepted = 0;
  const failures = [];

  async function worker() {
    while (cursor < payloads.length) {
      const index = cursor++;
      try {
        await send(payloads[index]);
        accepted += 1;
        if (accepted % 50 === 0) {
          console.log(`Accepted ${accepted}/${payloads.length}`);
        }
      } catch (error) {
        failures.push({ index, error: String(error) });
        if (failures.length >= 10) {
          throw new Error(`Stopping after 10 failures. First failure: ${failures[0].error}`);
        }
      }
    }
  }

  await Promise.all(Array.from({ length: config.concurrency }, () => worker()));
  return { accepted, failures };
}

const payloads = buildPayloads();
const summary = summarize(payloads);

if (config.dryRun) {
  console.log(JSON.stringify(summary, null, 2));
  console.log(JSON.stringify(payloads.slice(0, 2), null, 2));
  process.exit(0);
}

console.log(JSON.stringify(summary, null, 2));
const result = await runQueue(payloads);
console.log(JSON.stringify(result, null, 2));

if (result.failures.length > 0) {
  process.exit(1);
}
