const fs = require("fs/promises");
const path = require("path");
const { chromium } = require("playwright");

const ROOT = "C:/Users/admin/Documents/src/uba/simpletrack/docs/litlyx";
const EMAIL = process.env.LITLYX_EMAIL;
const PASSWORD = process.env.LITLYX_PASSWORD;
const SEO_URL = process.env.LITLYX_SEO_URL || "https://dashboard.litlyx.com/seo";
const CONTENT_CLIP = { x: 260, y: 60, width: 1180, height: 1040 };

if (!EMAIL || !PASSWORD) {
  throw new Error("Set LITLYX_EMAIL and LITLYX_PASSWORD before running this capture script.");
}

async function ensureDir(filePath) {
  await fs.mkdir(path.dirname(filePath), { recursive: true });
}

async function waitForSettled(page, timeout = 30000) {
  await page.waitForLoadState("domcontentloaded", { timeout }).catch(() => {});
  await page.waitForLoadState("networkidle", { timeout }).catch(() => {});
}

async function redactVisibleSensitiveText(page) {
  await page
    .evaluate(() => {
      const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
      const nodes = [];
      while (walker.nextNode()) {
        nodes.push(walker.currentNode);
      }
      for (const node of nodes) {
        node.nodeValue = node.nodeValue
          .replace(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi, "account@example.com")
          .replace(/\b[a-f0-9]{24}\b/gi, "workspace_id");
      }
    })
    .catch(() => {});
}

async function shoot(page, relativePath, options = {}) {
  await redactVisibleSensitiveText(page);
  const filePath = path.join(ROOT, relativePath);
  await ensureDir(filePath);
  await page.screenshot({
    path: filePath,
    clip: options.clip,
    fullPage: options.fullPage ?? true,
  });
}

async function shootContent(page, relativePath) {
  await shoot(page, relativePath, { clip: CONTENT_CLIP, fullPage: false });
}

async function clickIfVisible(locator) {
  if (await locator.count()) {
    await locator.first().click();
    return true;
  }
  return false;
}

async function login(page) {
  await page.goto("https://dashboard.litlyx.com/login", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-01-login-and-workspace/P01-S01-login-page.png");

  const emailInput = page.locator("#email");
  const passwordInput = page.locator('input[type="password"]');
  await emailInput.fill("redacted@example.com");
  await passwordInput.fill("********");
  await shoot(page, "snapshots/phase-01-login-and-workspace/P01-S02-login-filled.png", {
    fullPage: false,
  });
  await emailInput.fill(EMAIL);
  await passwordInput.fill(PASSWORD);

  await Promise.all([
    page.waitForURL((url) => !url.toString().includes("/login"), {
      timeout: 60000,
    }),
    page.getByRole("button", { name: "Log in" }).click(),
  ]);
  await waitForSettled(page, 60000);
}

async function capturePhase01(page) {
  await page.goto("https://dashboard.litlyx.com/", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-01-login-and-workspace/P01-S03-web-install-landing.png");

  await page.goto("https://dashboard.litlyx.com/workspaces", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-01-login-and-workspace/P01-S04-workspaces-list.png");
}

async function capturePhase02(page) {
  await page.goto("https://dashboard.litlyx.com/", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-02-web-overview-and-onboarding/P02-S01-install-script-tab.png");

  await page.getByText("Tag (GTM)", { exact: true }).click();
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-02-web-overview-and-onboarding/P02-S02-install-gtm-tab.png", {
    fullPage: false,
  });

  await page.goto("https://dashboard.litlyx.com/settings?tab=general", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-02-web-overview-and-onboarding/P02-S03-settings-general.png");

  await page.getByText("Domains", { exact: true }).last().click();
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-02-web-overview-and-onboarding/P02-S04-settings-domains.png");
  await shootContent(page, "snapshots/phase-02-web-overview-and-onboarding/P02-S05-settings-domain-sanitization.png");
}

async function capturePhase03(page) {
  await page.goto("https://dashboard.litlyx.com/events", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-03-product-and-events/P03-S01-product-empty-state.png");
  await page.getByText("Funnel Analysis", { exact: true }).scrollIntoViewIfNeeded().catch(() => {});
  await shootContent(page, "snapshots/phase-03-product-and-events/P03-S04-product-lower-half-empty-state.png");

  await page.getByText("Raw Data", { exact: true }).click();
  await page.waitForTimeout(1200);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-03-product-and-events/P03-S03-product-raw-data.png", {
    fullPage: false,
  });

  await page.goto("https://dashboard.litlyx.com/events", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  const setupPopup = page.waitForEvent("popup", { timeout: 8000 }).catch(() => null);
  await page.getByText("Setup events", { exact: true }).click();
  const customEventsDocs = await setupPopup;
  if (customEventsDocs) {
    await customEventsDocs.waitForLoadState("domcontentloaded", { timeout: 30000 }).catch(() => {});
    await customEventsDocs.waitForLoadState("networkidle", { timeout: 30000 }).catch(() => {});
    await shoot(
      customEventsDocs,
      "snapshots/phase-03-product-and-events/P03-S07-setup-events-docs-target.png",
      { fullPage: true },
    );
    await customEventsDocs.close();
  }

  await page.goto("https://dashboard.litlyx.com/events", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await page.getByText("Show test data", { exact: true }).click();
  await page.waitForTimeout(1500);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-03-product-and-events/P03-S02-product-with-test-data.png");
}

async function capturePhase04(page) {
  await page.goto("https://dashboard.litlyx.com/marketing", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S01-marketing-empty-state.png");

  await page.getByText("Show test data", { exact: true }).click();
  await page.waitForTimeout(1500);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S02-marketing-with-test-data.png");

  await page.goto("https://dashboard.litlyx.com/marketing", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await page.getByText("Generate UTM link", { exact: true }).click();
  await page.waitForTimeout(1200);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S03-generate-utm-dialog.png", {
    fullPage: false,
  });

  await page.goto("https://dashboard.litlyx.com/reports", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S04-report-catalog.png");
  await shootContent(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S07-report-generation-disabled.png");

  const sampleButton = page.locator('button:has-text("Sample")');
  if (await sampleButton.count()) {
    await sampleButton.first().click();
    await page.waitForTimeout(1500);
    await waitForSettled(page);
    await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S05-report-sample-preview.png", {
      fullPage: false,
    });
  }

  await page.goto(SEO_URL, {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-04-marketing-seo-and-reports/P04-S06-seo-premium-gate.png");
}

async function capturePhase05(page) {
  await page.goto("https://dashboard.litlyx.com/shields", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S01-shields-domains.png");

  await page.getByText("IP addresses", { exact: true }).click();
  await page.waitForTimeout(800);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S02-shields-ip-addresses.png");

  await page.getByText("Bot traffic", { exact: true }).click();
  await page.waitForTimeout(800);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S03-shields-bot-traffic.png");

  await page.getByText("Domains", { exact: true }).last().click();
  await page.waitForTimeout(500);
  await page.getByRole("button", { name: "Add domain" }).click();
  await page.waitForTimeout(800);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S04-shields-add-domain-dialog.png", {
    fullPage: false,
  });

  await page.goto("https://dashboard.litlyx.com/ai", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S05-ai-assistant.png");

  await page.getByRole("button", { name: "Examples" }).click();
  await page.waitForTimeout(600);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S06-ai-examples-expanded.png", {
    fullPage: false,
  });

  await page.goto("https://dashboard.litlyx.com/plans", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S07-plans-personal.png");

  await page.getByRole("button", { name: "Business" }).click();
  await page.waitForTimeout(800);
  await waitForSettled(page);
  await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S08-plans-business.png", {
    fullPage: false,
  });

  const faqButton = page.getByRole("button", {
    name: "What plan is active during my free trial?",
  });
  if (await faqButton.count()) {
    await faqButton.click();
    await page.waitForTimeout(500);
    await waitForSettled(page);
    await shoot(page, "snapshots/phase-05-members-settings-and-ai/P05-S09-plans-faq-expanded.png", {
      fullPage: false,
    });
  }

  await page.goto("https://dashboard.litlyx.com/shareable_links", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await shootContent(page, "snapshots/phase-05-members-settings-and-ai/P05-S10-shareable-links-empty.png");

  await page.goto("https://dashboard.litlyx.com/members", {
    waitUntil: "domcontentloaded",
    timeout: 60000,
  });
  await waitForSettled(page);
  await page.waitForTimeout(10000);
  await shootContent(page, "snapshots/phase-05-members-settings-and-ai/P05-S11-members-loading-limited.png");
}

async function main() {
  const browser = await chromium.launch({ headless: true, channel: "msedge" });
  const page = await browser.newPage({ viewport: { width: 1440, height: 1400 } });

  await login(page);
  await capturePhase01(page);
  await capturePhase02(page);
  await capturePhase03(page);
  await capturePhase04(page);
  await capturePhase05(page);

  await browser.close();
}

main().catch((error) => {
  console.error(error && error.stack ? error.stack : String(error));
  process.exit(1);
});
