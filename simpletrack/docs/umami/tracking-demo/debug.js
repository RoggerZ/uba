import {
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
  log.textContent = [new Date().toISOString(), message, payload].filter(Boolean).join("\n");
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
});
