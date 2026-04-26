import { existsSync, readdirSync } from "node:fs";
import { mkdir } from "node:fs/promises";
import { homedir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";

const require = createRequire(import.meta.url);
const args = process.argv.slice(2);
const scriptDir = dirname(fileURLToPath(import.meta.url));
const defaultOutputRoot = resolve(scriptDir, "..", "snapshots");

function readFlag(name, fallback = "") {
  const index = args.findIndex((arg) => arg === `--${name}`);
  if (index === -1 || index === args.length - 1) return fallback;
  return args[index + 1];
}

function hasFlag(name) {
  return args.includes(`--${name}`);
}

const config = {
  websiteId: readFlag("website-id"),
  region: readFlag("region", "us"),
  outputRoot: resolve(readFlag("output-root", defaultOutputRoot)),
  headless: !hasFlag("headed"),
  loginWaitMs: Number(readFlag("login-wait-ms", "0")),
  settleMs: Number(readFlag("settle-ms", "6500")),
  storageState: readFlag("storage-state"),
  userDataDir: readFlag("user-data-dir"),
  slowMo: Number(readFlag("slow-mo", "0")),
  only: readFlag("only"),
  redact: !hasFlag("no-redact"),
};

if (!config.websiteId) {
  console.error("Missing required flag: --website-id");
  process.exit(1);
}

const cloudBase = `https://cloud.umami.is/analytics/${config.region}/websites/${config.websiteId}`;

const targets = [
  ["P07-S01", "phase-07-traffic-and-behavior-insights", "sessions", ["Sessions"]],
  ["P07-S02", "phase-07-traffic-and-behavior-insights", "realtime", ["Realtime", "Real-time"]],
  ["P07-S03", "phase-07-traffic-and-behavior-insights", "performance", ["Performance"]],
  ["P07-S04", "phase-07-traffic-and-behavior-insights", "compare", ["Compare"]],
  ["P07-S05", "phase-07-traffic-and-behavior-insights", "breakdown", ["Breakdown", "BreakDown"]],
  ["P07-S06", "phase-07-traffic-and-behavior-insights", "goals", ["Goals"]],
  ["P07-S07", "phase-07-traffic-and-behavior-insights", "filter", ["Filter", "Filters"]],
  ["P07-S08", "phase-07-traffic-and-behavior-insights", "filter-segment-applied", ["Compare"], "apply-producthunt-segment-filter"],
  ["P08-S01", "phase-08-growth-and-monetization-insights", "funnels", ["Funnels"]],
  ["P08-S02", "phase-08-growth-and-monetization-insights", "journeys", ["Journeys"]],
  ["P08-S03", "phase-08-growth-and-monetization-insights", "retention", ["Retention"]],
  ["P08-S04", "phase-08-growth-and-monetization-insights", "replays", ["Replays"]],
  ["P08-S05", "phase-08-growth-and-monetization-insights", "segments", ["Segments"]],
  ["P08-S05A", "phase-08-growth-and-monetization-insights", "segment-config", ["Segments"], "open-producthunt-segment"],
  ["P08-S06", "phase-08-growth-and-monetization-insights", "cohorts", ["Cohorts"]],
  ["P08-S06A", "phase-08-growth-and-monetization-insights", "cohort-config", ["Cohorts"], "open-paid-checkout-cohort"],
  ["P08-S07", "phase-08-growth-and-monetization-insights", "utm", ["UTM"]],
  ["P08-S08", "phase-08-growth-and-monetization-insights", "revenue", ["Revenue"]],
  ["P08-S09", "phase-08-growth-and-monetization-insights", "attribution", ["Attribution"], "select-checkout-attribution"],
];

const selectedTargets = (() => {
  if (!config.only) return targets;
  const selected = new Set(
    config.only
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
  );
  return targets.filter(([id, , name]) => selected.has(id) || selected.has(name));
})();

if (selectedTargets.length === 0) {
  console.error(`No matching capture targets for --only ${config.only}`);
  process.exit(1);
}

async function loadPlaywright() {
  try {
    return await import("playwright");
  } catch (error) {
    const candidates = [];
    if (process.env.PLAYWRIGHT_MODULE_DIR) candidates.push(process.env.PLAYWRIGHT_MODULE_DIR);
    candidates.push(join(process.cwd(), "node_modules", "playwright"));
    candidates.push(join(homedir(), "AppData", "Local", "npm-cache", "_npx"));

    for (const candidate of candidates) {
      if (!existsSync(candidate)) continue;
      if (candidate.endsWith("_npx")) {
        for (const entry of readdirSync(candidate, { withFileTypes: true })) {
          if (!entry.isDirectory()) continue;
          const playwrightDir = join(candidate, entry.name, "node_modules", "playwright");
          if (existsSync(playwrightDir)) return require(playwrightDir);
        }
      } else if (existsSync(candidate)) {
        return require(candidate);
      }
    }

    throw new Error(`Unable to load Playwright. Run with: npx --yes -p playwright node capture-cloud-snapshots.mjs ...\nOriginal error: ${error}`);
  }
}

async function waitForLoginIfNeeded(page) {
  await page.goto(cloudBase, { waitUntil: "domcontentloaded", timeout: 30000 });
  await page.waitForTimeout(2500);
  if (!(await isLoginPage(page))) return true;
  return waitForManualLogin(page);
}

async function waitForManualLogin(page) {
  if (!config.loginWaitMs || config.headless) return false;
  console.log(`Cloud login required. Waiting ${config.loginWaitMs}ms for headed browser login...`);
  const deadline = Date.now() + config.loginWaitMs;
  while (Date.now() < deadline) {
    await page.waitForTimeout(1000);
    if (!(await isLoginPage(page))) return true;
  }
  return false;
}

async function isLoginPage(page) {
  if (page.url().includes("/login")) return true;
  return page.evaluate(() => {
    const text = document.body?.innerText || "";
    const hasPasswordInput = Boolean(document.querySelector('input[type="password"]'));
    return hasPasswordInput && text.includes("Email Address") && text.includes("Log in");
  }).catch(() => false);
}

async function clickLabel(page, labels) {
  for (const label of labels) {
    const locator = page.getByText(label, { exact: true }).first();
    if (await locator.count().catch(() => 0)) {
      await locator.click({ timeout: 5000 }).catch(() => null);
      await settle(page);
      return true;
    }
  }
  return false;
}

async function clickButtonByText(page, text) {
  return page.evaluate((text) => {
    const buttons = [...document.querySelectorAll("button")];
    const target = buttons.reverse().find((button) => (button.textContent || "").trim() === text && !button.disabled);
    if (!target) return false;
    target.click();
    return true;
  }, text);
}

async function clickEditInRow(page, rowName) {
  return page.evaluate((rowName) => {
    const text = [...document.querySelectorAll("a,span,div,td")].find((element) => (element.textContent || "").trim() === rowName);
    const row = text?.closest("tr") || text?.parentElement?.parentElement;
    const buttons = row ? [...row.querySelectorAll("button")] : [];
    const target = buttons[0] || text;
    if (!target) return false;
    target.click();
    return true;
  }, rowName);
}

async function prepareTarget(page, prepare) {
  if (!prepare) return false;

  if (prepare === "open-producthunt-segment") {
    const opened = await clickEditInRow(page, "Producthunt Launch Segment");
    await settle(page);
    return opened;
  }

  if (prepare === "open-paid-checkout-cohort") {
    const opened = await clickEditInRow(page, "Paid Checkout Cohort");
    await settle(page);
    return opened;
  }

  if (prepare === "apply-producthunt-segment-filter") {
    const opened = await clickButtonByText(page, "Filter");
    if (!opened) return false;
    await page.waitForTimeout(800);
    await page.getByRole("tab", { name: "Segments" }).click().catch(() => null);
    await page.waitForTimeout(700);
    await page.getByRole("option", { name: "Producthunt Launch Segment" }).click().catch(() => null);
    await page.waitForTimeout(700);
    const applied = await clickButtonByText(page, "Apply");
    await settle(page);
    const text = await page.evaluate(() => document.body?.innerText || "").catch(() => "");
    return applied && text.includes("Producthunt Launch Segment");
  }

  if (prepare === "select-checkout-attribution") {
    const selectedType = await page.locator("select").nth(3).selectOption("event").then(() => true).catch(() => false);
    await page.waitForTimeout(1000);
    const searchInputs = page.locator('input[type="search"]');
    const count = await searchInputs.count().catch(() => 0);
    let selectedStep = false;
    if (count > 0) {
      const target = searchInputs.nth(count - 1);
      await target.fill("checkout_completed").catch(() => null);
      await page.waitForTimeout(1000);
      await target.press("Enter").catch(() => null);
      selectedStep = (await target.inputValue().catch(() => "")) === "checkout_completed";
    }
    await settle(page);
    return selectedType && selectedStep;
  }

  return false;
}

function redactionMasks(page) {
  if (!config.redact) return [];
  return [
    page.locator("button").filter({ hasText: /@/ }),
    page.locator("text=/[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}/"),
  ];
}

async function settle(page) {
  await page.waitForLoadState("domcontentloaded", { timeout: 5000 }).catch(() => null);
  await page.waitForLoadState("networkidle", { timeout: 10000 }).catch(() => null);
  await page.waitForTimeout(config.settleMs);
}

async function captureTarget(page, target) {
  const [id, phase, name, labels, prepare] = target;
  const dir = join(config.outputRoot, phase);
  await mkdir(dir, { recursive: true });

  await page.goto(cloudBase, { waitUntil: "domcontentloaded", timeout: 30000 });
  await settle(page);
  if (await isLoginPage(page)) {
    const loggedIn = await waitForManualLogin(page);
    if (!loggedIn) {
      throw new Error(`Cloud login required before capturing ${id}. No formal screenshots were written.`);
    }
    await page.goto(cloudBase, { waitUntil: "domcontentloaded", timeout: 30000 });
    await settle(page);
  }
  const clicked = await clickLabel(page, labels);
  const prepared = await prepareTarget(page, prepare);
  if (prepare && !prepared) {
    throw new Error(`Unable to prepare ${id} with action ${prepare}. Run configure-cloud-reports.mjs first if the required report object is missing.`);
  }
  if (await isLoginPage(page)) {
    throw new Error(`Cloud redirected to login while capturing ${id}. No formal screenshots were written.`);
  }
  const filename = `${id}-${name}.png`;
  const output = join(dir, filename);
  await page.screenshot({
    path: output,
    fullPage: true,
    mask: redactionMasks(page),
    maskColor: "#f8fafc",
  });

  return {
    id,
    page: name,
    clicked,
    prepared,
    redacted: config.redact,
    url: page.url(),
    output,
    loginRedirect: page.url().includes("/login"),
  };
}

const { chromium } = await loadPlaywright();
const contextOptions = {
  viewport: { width: 1440, height: 1000 },
  locale: "en-US",
};
if (config.storageState) contextOptions.storageState = config.storageState;

let browser;
let context;
if (config.userDataDir) {
  context = await chromium.launchPersistentContext(config.userDataDir, {
    ...contextOptions,
    headless: config.headless,
    slowMo: config.slowMo,
  });
} else {
  browser = await chromium.launch({ headless: config.headless, slowMo: config.slowMo });
  context = await browser.newContext(contextOptions);
}
const page = await context.newPage();

try {
  const loggedIn = await waitForLoginIfNeeded(page);
  if (!loggedIn) {
    console.error("Cloud login required. Re-run with --headed --login-wait-ms 120000, or provide a storage state outside the repo via --storage-state.");
    process.exitCode = 2;
  } else {
    const results = [];
    let failed = false;
    try {
      for (const target of selectedTargets) {
        results.push(await captureTarget(page, target));
      }
    } catch (error) {
      console.error(error.message);
      process.exitCode = 2;
      failed = true;
    }
    if (!failed) {
      console.log(JSON.stringify({ captured: results.length, results }, null, 2));
    }
  }
} finally {
  await context.close();
  if (browser) await browser.close();
}
