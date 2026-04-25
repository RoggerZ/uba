const DEFAULT_SCRIPT_URL = "https://cloud.umami.is/script.js";
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

initTracker();
