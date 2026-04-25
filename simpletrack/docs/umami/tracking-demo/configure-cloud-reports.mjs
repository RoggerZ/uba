import { existsSync, readdirSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const args = process.argv.slice(2);

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
  headless: !hasFlag("headed"),
  loginWaitMs: Number(readFlag("login-wait-ms", "0")),
  settleMs: Number(readFlag("settle-ms", "2500")),
  storageState: readFlag("storage-state"),
  userDataDir: readFlag("user-data-dir"),
  slowMo: Number(readFlag("slow-mo", "0")),
  only: readFlag("only"),
  dryRun: hasFlag("dry-run"),
};

if (!config.websiteId) {
  console.error("Missing required flag: --website-id");
  process.exit(1);
}

const cloudBase = `https://cloud.umami.is/analytics/${config.region}/websites/${config.websiteId}`;

const reportSpecs = [
  {
    key: "goal",
    route: "goals",
    button: "Goal",
    name: "Checkout Completed Goal",
    create: createCheckoutGoal,
  },
  {
    key: "funnel",
    route: "funnels",
    button: "Funnel",
    name: "Growth Baseline Checkout Funnel",
    create: createCheckoutFunnel,
  },
  {
    key: "segment",
    route: "segments",
    button: "Segment",
    name: "Producthunt Launch Segment",
    create: createProducthuntSegment,
  },
  {
    key: "cohort",
    route: "cohorts",
    button: "Cohort",
    name: "Paid Checkout Cohort",
    create: createPaidCheckoutCohort,
  },
];

const selectedSpecs = (() => {
  if (!config.only) return reportSpecs;
  const selected = new Set(
    config.only
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
  );
  return reportSpecs.filter((spec) => selected.has(spec.key) || selected.has(spec.name));
})();

if (selectedSpecs.length === 0) {
  console.error(`No matching report objects for --only ${config.only}`);
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

    throw new Error(`Unable to load Playwright. Run with: npx --yes -p playwright node configure-cloud-reports.mjs ...\nOriginal error: ${error}`);
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

async function settle(page) {
  await page.waitForLoadState("domcontentloaded", { timeout: 5000 }).catch(() => null);
  await page.waitForLoadState("networkidle", { timeout: 10000 }).catch(() => null);
  await page.waitForTimeout(config.settleMs);
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

async function clickExactText(page, text) {
  return page.evaluate((text) => {
    const candidates = [...document.querySelectorAll("button,[role=button],[role=option],div,span,li")];
    const target = candidates.find((element) => (element.textContent || "").trim() === text);
    if (!target) return false;
    target.click();
    return true;
  }, text);
}

async function pageText(page) {
  return page.evaluate(() => document.body?.innerText || "");
}

async function selectOptionByText(page, optionText, index = 0) {
  const selected = await page.evaluate(({ optionText, index }) => {
    const selects = [...document.querySelectorAll("select")].filter((select) =>
      [...select.options].some((option) => option.textContent.trim() === optionText || option.value === optionText),
    );
    const select = selects[index];
    if (!select) return false;
    const option = [...select.options].find((item) => item.textContent.trim() === optionText || item.value === optionText);
    if (!option) return false;
    select.value = option.value;
    select.dispatchEvent(new Event("change", { bubbles: true }));
    return true;
  }, { optionText, index });
  await page.waitForTimeout(500);
  return selected;
}

async function selectStepType(page, stepIndex, optionText) {
  const selected = await page.evaluate(({ stepIndex, optionText }) => {
    const selects = [...document.querySelectorAll("select")].filter((select) =>
      [...select.options].some((option) => option.textContent.trim() === optionText || option.value === optionText),
    );
    const select = selects[stepIndex];
    if (!select) return false;
    const option = [...select.options].find((item) => item.textContent.trim() === optionText || item.value === optionText);
    if (!option) return false;
    select.value = option.value;
    select.dispatchEvent(new Event("change", { bubbles: true }));
    return true;
  }, { stepIndex, optionText });
  await page.waitForTimeout(500);
  return selected;
}

async function ensureReport(page, spec) {
  await page.goto(`${cloudBase}/${spec.route}`, { waitUntil: "domcontentloaded", timeout: 60000 });
  await settle(page);
  if (await isLoginPage(page)) {
    throw new Error("Cloud login required before configuring report objects.");
  }

  const before = await pageText(page);
  if (before.includes(spec.name)) {
    return { key: spec.key, name: spec.name, status: "exists" };
  }

  if (config.dryRun) {
    return { key: spec.key, name: spec.name, status: "missing" };
  }

  const opened = await clickButtonByText(page, spec.button);
  if (!opened) {
    return { key: spec.key, name: spec.name, status: "failed", reason: "open-button-not-found" };
  }
  await page.waitForTimeout(1000);

  await spec.create(page);
  const saved = await clickButtonByText(page, "Save");
  if (!saved) {
    return { key: spec.key, name: spec.name, status: "failed", reason: "save-button-not-found" };
  }
  await settle(page);

  const after = await pageText(page);
  if (!after.includes(spec.name)) {
    return { key: spec.key, name: spec.name, status: "failed", reason: "object-not-visible-after-save" };
  }

  return { key: spec.key, name: spec.name, status: "created" };
}

async function createCheckoutGoal(page) {
  await page.locator('input[name="name"]').fill("Checkout Completed Goal");
  await selectOptionByText(page, "Triggered event");
  await page.locator('input[name="parameters.value"]').fill("checkout_completed");
}

async function createCheckoutFunnel(page) {
  await page.locator('input[name="name"]').fill("Growth Baseline Checkout Funnel");
  await page.locator('input[name="window"]').fill("60");
  const steps = ["pricing_viewed", "checkout_started", "checkout_completed"];

  for (let index = 0; index < steps.length; index += 1) {
    if (index > 0) {
      await clickButtonByText(page, "Add");
      await page.waitForTimeout(500);
    }
    await selectStepType(page, index, "Triggered event");
    await page.locator(`input[name="steps.${index}.value"]`).fill(steps[index]);
  }
}

async function createProducthuntSegment(page) {
  await page.locator('input[name="name"]').fill("Producthunt Launch Segment");
  await clickExactText(page, "Campaign");
  await page.waitForTimeout(700);
  await clickButtonByText(page, "Select an item");
  await page.waitForTimeout(1000);
  await clickExactText(page, "producthunt_launch");
}

async function createPaidCheckoutCohort(page) {
  await page.locator('input[name="name"]').fill("Paid Checkout Cohort");
  await selectOptionByText(page, "Triggered event");
  await page.locator('input[name="parameters.action.value"]').fill("checkout_completed");
  await page.locator('select[name="parameters.dateRange"]').selectOption("90day");
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
    for (const spec of selectedSpecs) {
      results.push(await ensureReport(page, spec));
    }
    const failed = results.filter((result) => result.status === "failed");
    console.log(JSON.stringify({ dryRun: config.dryRun, configured: results }, null, 2));
    if (failed.length > 0) process.exitCode = 1;
  }
} finally {
  await context.close();
  if (browser) await browser.close();
}
