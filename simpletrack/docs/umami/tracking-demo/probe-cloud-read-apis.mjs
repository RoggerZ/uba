import { existsSync, readdirSync } from "node:fs";
import { writeFile } from "node:fs/promises";
import { homedir } from "node:os";
import { join, resolve } from "node:path";
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
  settleMs: Number(readFlag("settle-ms", "5000")),
  output: readFlag("output"),
  userDataDir: readFlag("user-data-dir"),
  storageState: readFlag("storage-state"),
  headless: !hasFlag("headed"),
};

if (!config.websiteId) {
  console.error("Missing required flag: --website-id");
  process.exit(1);
}

const routes = [
  ["overview", ""],
  ["realtime", "realtime"],
  ["events", "events"],
  ["sessions", "sessions"],
  ["performance", "performance"],
  ["compare", "compare"],
  ["breakdown", "breakdown"],
  ["goals", "goals"],
  ["funnels", "funnels"],
  ["journeys", "journeys"],
  ["retention", "retention"],
  ["replays", "replays"],
  ["segments", "segments"],
  ["cohorts", "cohorts"],
  ["utm", "utm"],
  ["revenue", "revenue"],
  ["attribution", "attribution"],
];

async function loadPlaywright() {
  try {
    return await import("playwright");
  } catch (error) {
    const candidates = [
      join(process.cwd(), "node_modules", "playwright"),
      join(homedir(), "AppData", "Local", "npm-cache", "_npx"),
    ];

    for (const candidate of candidates) {
      if (!existsSync(candidate)) continue;
      if (candidate.endsWith("_npx")) {
        for (const entry of readdirSync(candidate, { withFileTypes: true })) {
          if (!entry.isDirectory()) continue;
          const playwrightDir = join(candidate, entry.name, "node_modules", "playwright");
          if (existsSync(playwrightDir)) return require(playwrightDir);
        }
      } else {
        return require(candidate);
      }
    }

    throw new Error(
      `Unable to load Playwright. Run with: npx --yes -p playwright node probe-cloud-read-apis.mjs ...\nOriginal error: ${error}`,
    );
  }
}

function websiteUrl(path = "") {
  const base = `https://cloud.umami.is/analytics/${config.region}/websites/${config.websiteId}`;
  return path ? `${base}/${path}` : base;
}

async function settle(page) {
  await page.waitForLoadState("networkidle", { timeout: 15000 }).catch(() => {});
  await page.waitForTimeout(config.settleMs);
}

function summarizeText(text) {
  return text
    .replace(/\s+/g, " ")
    .replace(/[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/g, "[redacted-email]")
    .slice(0, 800);
}

async function collectRoute(page, routeName, path) {
  const hits = [];
  const handler = async (response) => {
    const url = response.url();
    if (!url.includes("/api/")) return;
    if (!url.includes(config.websiteId)) return;

    const hit = { status: response.status(), url };
    try {
      const contentType = response.headers()["content-type"] || "";
      if (contentType.includes("application/json")) {
        hit.body = await response.text();
      }
    } catch {}
    hits.push(hit);
  };

  page.on("response", handler);
  try {
    await page.goto(websiteUrl(path), { waitUntil: "domcontentloaded", timeout: 60000 });
    await settle(page);

    const state = await page.evaluate(() => {
      const rawText = document.body.innerText || "";
      const text = rawText.replace(/\s+/g, " ").trim();
      return {
        title: document.title,
        noData: text.includes("No data available"),
        visitorsZero: /Visitors 0/.test(text),
        visitsZero: /Visits 0/.test(text),
        viewsZero: /Views 0/.test(text),
        eventsZero: /Events 0/.test(text),
        textSample: text.slice(0, 800),
      };
    });

    return {
      route: routeName,
      finalUrl: page.url(),
      state,
      apiHits: hits,
    };
  } finally {
    page.off("response", handler);
  }
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
  });
} else {
  browser = await chromium.launch({ headless: config.headless });
  context = await browser.newContext(contextOptions);
}

const page = await context.newPage();

try {
  const results = [];
  for (const [routeName, path] of routes) {
    const item = await collectRoute(page, routeName, path);
    item.state.textSample = summarizeText(item.state.textSample);
    results.push(item);
  }

  const report = {
    generatedAt: new Date().toISOString(),
    region: config.region,
    routeCount: results.length,
    results,
  };

  if (config.output) {
    const outputPath = resolve(config.output);
    await writeFile(outputPath, JSON.stringify(report, null, 2), "utf8");
  }

  console.log(JSON.stringify(report, null, 2));
} finally {
  await context.close();
  if (browser) await browser.close();
}
