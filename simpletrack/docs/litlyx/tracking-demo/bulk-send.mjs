const DEFAULT_ENDPOINT = "https://broker.litlyx.com/event";

function parseArgs(argv) {
  const args = {};
  for (let i = 0; i < argv.length; i += 1) {
    const item = argv[i];
    if (!item.startsWith("--")) continue;
    const key = item.slice(2);
    const value = argv[i + 1];
    if (value && !value.startsWith("--")) {
      args[key] = value;
      i += 1;
    } else {
      args[key] = "true";
    }
  }
  return args;
}

async function postEvent(endpoint, payload) {
  const response = await fetch(endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const text = await response.text();
  return { ok: response.ok, status: response.status, text };
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (!args.pid) {
    console.log("Usage: node bulk-send.mjs --pid WORKSPACE_ID [--users 6] [--sessions-per-user 2]");
    process.exit(1);
  }

  const endpoint = args.endpoint || DEFAULT_ENDPOINT;
  const users = Number(args.users || 6);
  const sessionsPerUser = Number(args["sessions-per-user"] || 2);
  const baseUrl = args["base-url"] || "http://localhost:4174";
  const website = args.website || "localhost";

  const eventNames = [
    "demo_signup_click",
    "demo_checkout_started",
    "demo_report_requested",
    "demo_upgrade_viewed",
  ];

  let accepted = 0;
  const errors = [];

  for (let u = 0; u < users; u += 1) {
    for (let s = 0; s < sessionsPerUser; s += 1) {
      for (let e = 0; e < eventNames.length; e += 1) {
        const payload = {
          pid: args.pid,
          name: eventNames[e],
          metadata: JSON.stringify({
            source: "bulk-send",
            user: `demo-user-${String(u + 1).padStart(2, "0")}`,
            session: s + 1,
            step: e + 1,
            plan: e % 2 === 0 ? "starter" : "pro",
            page: `${baseUrl}/index.html?user=${u + 1}&session=${s + 1}`,
          }),
          website,
          userAgent: "litlyx-bulk-send/1.0",
        };

        const result = await postEvent(endpoint, payload);
        if (result.ok) {
          accepted += 1;
        } else {
          errors.push({ payload, result });
        }
      }
    }
  }

  console.log(JSON.stringify({
    accepted,
    attempted: users * sessionsPerUser * eventNames.length,
    errors: errors.slice(0, 5),
  }, null, 2));

  if (errors.length) {
    process.exit(1);
  }
}

main().catch((error) => {
  console.error(String(error));
  process.exit(1);
});
