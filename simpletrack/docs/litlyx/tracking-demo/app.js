const statusLog = document.getElementById("status-log");
const workspaceIdInput = document.getElementById("workspace-id");
const scriptUrlInput = document.getElementById("script-url");
const eventANameInput = document.getElementById("event-a-name");
const eventAMetadataInput = document.getElementById("event-a-metadata");
const eventBNameInput = document.getElementById("event-b-name");
const eventBMetadataInput = document.getElementById("event-b-metadata");

function log(message, data) {
  const stamp = new Date().toISOString();
  const suffix = data === undefined ? "" : `\n${JSON.stringify(data, null, 2)}`;
  statusLog.textContent = `[${stamp}] ${message}${suffix}\n\n${statusLog.textContent}`;
}

function parseJson(label, value) {
  try {
    return JSON.parse(value);
  } catch (error) {
    log(`${label} JSON parse failed`, { error: String(error) });
    throw error;
  }
}

function getLit() {
  if (!window.Lit || typeof window.Lit.event !== "function") {
    log("Lit.event is not available yet.");
    return null;
  }
  return window.Lit;
}

function loadTracker() {
  const workspaceId = workspaceIdInput.value.trim();
  const scriptUrl = scriptUrlInput.value.trim();

  if (!workspaceId || !scriptUrl) {
    log("Missing workspace ID or script URL.");
    return;
  }

  const existing = document.getElementById("litlyx-script");
  if (existing) {
    existing.remove();
  }

  const script = document.createElement("script");
  script.id = "litlyx-script";
  script.defer = true;
  script.dataset.workspace = workspaceId;
  script.src = scriptUrl;
  script.onload = () => {
    log("Litlyx browser script loaded.", {
      workspaceId,
      scriptUrl,
      autoPageviewExpected: true,
    });
  };
  script.onerror = () => {
    log("Litlyx browser script failed to load.", { workspaceId, scriptUrl });
  };

  document.head.appendChild(script);
}

function fireCustomEvent(nameInput, metadataInput, label) {
  const Lit = getLit();
  if (!Lit) return;

  const eventName = nameInput.value.trim();
  const metadata = parseJson(label, metadataInput.value);
  Lit.event(eventName, { metadata });
  log(`${label} fired`, { eventName, metadata });
}

document.getElementById("load-tracker").addEventListener("click", loadTracker);
document.getElementById("fire-event-a").addEventListener("click", () => {
  fireCustomEvent(eventANameInput, eventAMetadataInput, "event-a");
});
document.getElementById("fire-event-b").addEventListener("click", () => {
  fireCustomEvent(eventBNameInput, eventBMetadataInput, "event-b");
});

log("Demo page ready.", {
  href: window.location.href,
  note: "Paste workspace ID at runtime. Do not hardcode it into repo files.",
});
