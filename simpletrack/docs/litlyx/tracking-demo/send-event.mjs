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

function usage() {
  console.log(`Usage:
  node send-event.mjs --pid WORKSPACE_ID --name EVENT_NAME [options]

Options:
  --website      Website/hostname label, default: localhost
  --user-agent   User agent label, default: litlyx-research-demo/1.0
  --metadata     JSON string, default: {"source":"tracking-demo"}
  --endpoint     Override endpoint, default: ${DEFAULT_ENDPOINT}
`);
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help || !args.pid || !args.name) {
    usage();
    process.exit(args.help ? 0 : 1);
  }

  let metadata = { source: "tracking-demo" };
  if (args.metadata) {
    metadata = JSON.parse(args.metadata);
  }

  const payload = {
    pid: args.pid,
    name: args.name,
    metadata: JSON.stringify(metadata),
    website: args.website || "localhost",
    userAgent: args["user-agent"] || "litlyx-research-demo/1.0",
  };

  const response = await fetch(args.endpoint || DEFAULT_ENDPOINT, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const text = await response.text();
  console.log(JSON.stringify({
    ok: response.ok,
    status: response.status,
    body: text,
    payload,
  }, null, 2));

  if (!response.ok) {
    process.exit(1);
  }
}

main().catch((error) => {
  console.error(String(error));
  process.exit(1);
});
