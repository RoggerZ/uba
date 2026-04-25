import { existsSync, readdirSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
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
  websiteId: readFlag("website-id", ""),
  baseUrl: readFlag("base-url", "http://localhost:4173/site").replace(/\/$/, ""),
  scriptUrl: readFlag("script-url", "https://cloud.umami.is/script.js"),
  recorderUrl: readFlag("recorder-url", ""),
  userAgent: readFlag(
    "user-agent",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
  ),
  users: Number(readFlag("users", "72")),
  sessionsPerUser: Number(readFlag("sessions-per-user", "3")),
  headless: !hasFlag("headed"),
  slowMo: Number(readFlag("slow-mo", "0")),
  settleMs: Number(readFlag("settle-ms", "1200")),
  blockUmami: hasFlag("block-umami"),
  useDefaultUserAgent: hasFlag("use-default-user-agent"),
  useCustomUserAgent: hasFlag("use-custom-user-agent"),
  dryRun: hasFlag("dry-run"),
};

const campaigns = [
  { campaign: "producthunt_launch", source: "producthunt", medium: "launch" },
  { campaign: "google_brand", source: "google", medium: "cpc" },
  { campaign: "google_competitor", source: "google", medium: "cpc" },
  { campaign: "docs_seo", source: "google", medium: "organic" },
  { campaign: "linkedin_founder", source: "linkedin", medium: "social" },
  { campaign: "email_nurture", source: "email", medium: "lifecycle" },
];

const cohorts = ["spring_launch", "self_serve_wave", "paid_pilot"];
const roles = ["founder", "growth_lead", "engineer", "product_manager"];
const workspaceSizes = ["1-10", "11-50", "51-200"];

const sessionPlans = [
  {
    name: "acquisition",
    entry: "index.html",
    clicks: ["Inspect pricing", "Compare platforms", "Create workspace"],
    expectedEvents: 8,
  },
  {
    name: "activation",
    entry: "signup.html",
    clicks: ["Create workspace", "Start install", "Open web install", "Verify first event", "Open dashboard", "Campaign", "Cohort"],
    expectedEvents: 10,
  },
  {
    name: "analysis-and-revenue",
    entry: "app/dashboard.html",
    clicks: ["Campaign", "Cohort", "Plan", "producthunt_launch", "activated_trials", "paid_pilot", "Start checkout", "Complete checkout", "View billing"],
    expectedEvents: 12,
  },
];

function pad(number, size) {
  return String(number).padStart(size, "0");
}

function planForPersona(index) {
  if (index >= 66) return "pro_monthly";
  if (index >= 48) return "trial";
  if (index >= 24) return "trial";
  return "free";
}

function syntheticDate(logicalDay) {
  const date = new Date(Date.UTC(2026, 2, 15));
  date.setUTCDate(date.getUTCDate() + logicalDay - 1);
  return date.toISOString().slice(0, 10);
}

function persona(index, sessionIndex) {
  const traffic = campaigns[(index + sessionIndex) % campaigns.length];
  const logicalDay = ((index * 3 + sessionIndex) % 42) + 1;
  const plan = planForPersona(index);
  return {
    id: `browser-user-${pad(index + 1, 3)}`,
    session: sessionIndex + 1,
    sessionName: sessionPlans[sessionIndex].name,
    plan,
    stage: index >= 66 ? "paid" : index >= 48 ? "activated_trial" : index >= 24 ? "registered" : "top_funnel",
    cohort: cohorts[index % cohorts.length],
    role: roles[index % roles.length],
    workspaceSize: workspaceSizes[index % workspaceSizes.length],
    logicalDay,
    syntheticDate: syntheticDate(logicalDay),
    ...traffic,
  };
}

function urlFor(entry, profile) {
  const params = new URLSearchParams({
    websiteId: config.websiteId,
    scriptUrl: config.scriptUrl,
    utm_source: profile.source,
    utm_medium: profile.medium,
    utm_campaign: profile.campaign,
    cohort: profile.cohort,
    plan: profile.plan,
    persona: profile.id,
    session: String(profile.session),
    role: profile.role,
    workspace_size: profile.workspaceSize,
    stage: profile.stage,
    logical_day: String(profile.logicalDay),
    synthetic_date: profile.syntheticDate,
    performance: "1",
  });
  if (config.recorderUrl) {
    params.set("recorderUrl", config.recorderUrl);
    params.set("replays", "1");
  }
  return `${config.baseUrl}/${entry}?${params.toString()}`;
}

async function clickByText(page, text) {
  try {
    const clicked = await page.evaluate((label) => {
      const candidates = [...document.querySelectorAll("a,button")];
      const target = candidates.find((element) => element.textContent?.trim().replace(/\s+/g, " ") === label);
      if (!target) return false;
      target.click();
      return true;
    }, text);
    if (!clicked) return false;
    await page.waitForLoadState("domcontentloaded", { timeout: 3500 }).catch(() => {});
    await page.waitForLoadState("networkidle", { timeout: 5000 }).catch(() => {});
    await page.waitForTimeout(config.settleMs);
    return true;
  } catch {
    return false;
  }
}

async function runSession(browser, userIndex, sessionIndex) {
  const profile = persona(userIndex, sessionIndex);
  const plan = sessionPlans[sessionIndex];
  const contextOptions = {
    viewport: sessionIndex === 1 ? { width: 390, height: 844 } : { width: 1440, height: 920 },
    locale: "en-US",
  };
  if (config.useCustomUserAgent) {
    contextOptions.userAgent = `SimpleTrackPersona/${profile.id} ${plan.name}`;
  } else if (!config.useDefaultUserAgent) {
    contextOptions.userAgent = config.userAgent;
  }
  const context = await browser.newContext(contextOptions);

  if (config.blockUmami) {
    await context.route("**/*", (route) => {
      const hostname = new URL(route.request().url()).hostname;
      if (hostname === "cloud.umami.is" || hostname === "api.umami.is") {
        return route.abort();
      }
      return route.continue();
    });
  }

  const page = await context.newPage();
  const errors = [];
  page.on("pageerror", (error) => errors.push(String(error)));
  page.on("console", (message) => {
    if (message.type() === "error") {
      const text = message.text();
      const location = message.location();
      if (location.url.endsWith("/favicon.ico")) {
        return;
      }
      if (config.blockUmami && text.includes("Failed to load resource: net::ERR_FAILED")) {
        return;
      }
      errors.push(text);
    }
  });

  await page.goto(urlFor(plan.entry, profile), { waitUntil: "domcontentloaded", timeout: 15000 });
  await page.waitForLoadState("networkidle", { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(config.settleMs);

  let clicked = 0;
  for (const text of plan.clicks) {
    if (await clickByText(page, text)) {
      clicked += 1;
    }
  }

  await page.waitForTimeout(config.settleMs);
  await context.close();

  return {
    persona: profile.id,
    session: profile.session,
    flow: plan.name,
    plan: profile.plan,
    campaign: profile.campaign,
    cohort: profile.cohort,
    clicked,
    expectedEvents: plan.expectedEvents,
    errors,
  };
}

function dryRunSummary() {
  const rows = [];
  for (let userIndex = 0; userIndex < config.users; userIndex += 1) {
    for (let sessionIndex = 0; sessionIndex < config.sessionsPerUser; sessionIndex += 1) {
      const profile = persona(userIndex, sessionIndex);
      rows.push({
        persona: profile.id,
        session: profile.session,
        flow: sessionPlans[sessionIndex].name,
        entry: sessionPlans[sessionIndex].entry,
        plan: profile.plan,
        campaign: profile.campaign,
        cohort: profile.cohort,
        expectedEvents: sessionPlans[sessionIndex].expectedEvents,
      });
    }
  }
  return {
    browserUsers: config.users,
    sessions: config.users * config.sessionsPerUser,
    expectedEvents: rows.reduce((sum, row) => sum + row.expectedEvents, 0),
    paidBrowserConversions: Math.max(0, config.users - 66),
    sample: rows.slice(0, 6),
  };
}

async function loadPlaywright() {
  try {
    return await import("playwright");
  } catch (error) {
    const candidates = [];
    if (process.env.PLAYWRIGHT_MODULE_DIR) {
      candidates.push(process.env.PLAYWRIGHT_MODULE_DIR);
    }
    candidates.push(join(process.cwd(), "node_modules", "playwright"));
    candidates.push(join(homedir(), "AppData", "Local", "npm-cache", "_npx"));

    for (const candidate of candidates) {
      if (!existsSync(candidate)) continue;
      if (candidate.endsWith("_npx")) {
        for (const entry of readdirSync(candidate, { withFileTypes: true })) {
          if (!entry.isDirectory()) continue;
          const playwrightDir = join(candidate, entry.name, "node_modules", "playwright");
          if (existsSync(playwrightDir)) {
            return require(playwrightDir);
          }
        }
      } else if (existsSync(candidate)) {
        return require(candidate);
      }
    }

    throw new Error(
      `Unable to load Playwright. Run with: npx --yes -p playwright node run-browser-flows.mjs ...\nOriginal error: ${error}`,
    );
  }
}

if (config.dryRun) {
  console.log(JSON.stringify(dryRunSummary(), null, 2));
  process.exit(0);
}

const { chromium } = await loadPlaywright();
const browser = await chromium.launch({ headless: config.headless, slowMo: config.slowMo });
const results = [];

try {
  for (let userIndex = 0; userIndex < config.users; userIndex += 1) {
    for (let sessionIndex = 0; sessionIndex < config.sessionsPerUser; sessionIndex += 1) {
      const result = await runSession(browser, userIndex, sessionIndex);
      results.push(result);
      if (results.length % 24 === 0) {
        console.log(`Completed browser sessions: ${results.length}/${config.users * config.sessionsPerUser}`);
      }
    }
  }
} finally {
  await browser.close();
}

const summary = {
  browserUsers: config.users,
  sessions: results.length,
  expectedEvents: results.reduce((sum, row) => sum + row.expectedEvents, 0),
  paidBrowserConversions: Math.max(0, config.users - 66),
  totalClicks: results.reduce((sum, row) => sum + row.clicked, 0),
  errors: results.flatMap((row) => row.errors.map((error) => ({ persona: row.persona, session: row.session, error }))).slice(0, 20),
};

console.log(JSON.stringify(summary, null, 2));

if (summary.errors.length > 0) {
  process.exit(1);
}
