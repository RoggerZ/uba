const statusLog = document.getElementById("status-log");
const websiteIdInput = document.getElementById("website-id");
const scriptUrlInput = document.getElementById("script-url");
const trackNameInput = document.getElementById("track-name");
const trackDataInput = document.getElementById("track-data");
const identifyIdInput = document.getElementById("identify-id");
const identifyDataInput = document.getElementById("identify-data");

function log(message, data) {
  const stamp = new Date().toISOString();
  const suffix = data === undefined ? "" : `\n${JSON.stringify(data, null, 2)}`;
  statusLog.textContent = `[${stamp}] ${message}${suffix}\n\n${statusLog.textContent}`;
}

window.umamiBeforeSendHandler = function umamiBeforeSendHandler(type, payload) {
  log("beforeSend payload", { type, payload });
  return payload;
};

function parseJson(label, value) {
  try {
    return JSON.parse(value);
  } catch (error) {
    log(`${label} JSON 解析失败`, { error: String(error) });
    throw error;
  }
}

function loadTracker() {
  const websiteId = websiteIdInput.value.trim();
  const scriptUrl = scriptUrlInput.value.trim();

  if (!websiteId || !scriptUrl) {
    log("缺少 website id 或 script url");
    return;
  }

  const existing = document.getElementById("umami-script");
  if (existing) {
    existing.remove();
  }

  const script = document.createElement("script");
  script.id = "umami-script";
  script.async = true;
  script.src = scriptUrl;
  script.dataset.websiteId = websiteId;
  script.dataset.beforeSend = "umamiBeforeSendHandler";
  script.onload = () => log("tracker script loaded", { websiteId, scriptUrl });
  script.onerror = () => log("tracker script load failed", { websiteId, scriptUrl });
  document.head.appendChild(script);
}

function manualPageview() {
  if (!window.umami || typeof window.umami.track !== "function") {
    log("window.umami.track 不可用");
    return;
  }

  window.umami.track();
  log("manual pageview fired", { url: window.location.href });
}

function fireTrack() {
  if (!window.umami || typeof window.umami.track !== "function") {
    log("window.umami.track 不可用");
    return;
  }

  const eventName = trackNameInput.value.trim();
  const data = parseJson("track data", trackDataInput.value);
  window.umami.track(eventName, data);
  log("track() fired", { eventName, data });
}

function fireIdentify() {
  if (!window.umami || typeof window.umami.identify !== "function") {
    log("window.umami.identify 不可用");
    return;
  }

  const identity = identifyIdInput.value.trim();
  const data = parseJson("identify data", identifyDataInput.value);
  window.umami.identify(identity, data);
  log("identify() fired", { identity, data });
}

document.getElementById("load-tracker").addEventListener("click", loadTracker);
document.getElementById("track-pageview").addEventListener("click", manualPageview);
document.getElementById("fire-track").addEventListener("click", fireTrack);
document.getElementById("fire-identify").addEventListener("click", fireIdentify);

log("demo page ready", { origin: window.location.origin, href: window.location.href });
