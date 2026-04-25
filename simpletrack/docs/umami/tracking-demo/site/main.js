import {
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
    label: element.textContent.trim().replace(/\s+/g, " "),
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
});
