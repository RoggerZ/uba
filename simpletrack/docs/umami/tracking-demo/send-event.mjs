const args = process.argv.slice(2);

function readFlag(name, fallback = "") {
  const index = args.findIndex((arg) => arg === `--${name}`);
  if (index === -1 || index === args.length - 1) {
    return fallback;
  }
  return args[index + 1];
}

const website = readFlag("website-id");
const hostname = readFlag("hostname", "localhost");
const url = readFlag("url", "http://localhost:4173/index.html");
const title = readFlag("title", "Tracking Demo");
const name = readFlag("name", "server_demo_event");
const dataText = readFlag("data", "{\"source\":\"node-script\"}");
const apiUrl = readFlag("api-url", "https://api-gateway.umami.dev/api/send");
const userAgent = readFlag(
  "user-agent",
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
);

if (!website) {
  console.error("Missing required flag: --website-id");
  process.exit(1);
}

let data;
try {
  data = JSON.parse(dataText);
} catch (error) {
  console.error("Invalid --data JSON:", error);
  process.exit(1);
}

const payload = {
  type: "event",
  payload: {
    website,
    hostname,
    language: "zh-CN",
    referrer: "",
    screen: "1440x900",
    title,
    url,
    name,
    data,
  },
};

const response = await fetch(apiUrl, {
  method: "POST",
  headers: {
    "content-type": "application/json",
    "user-agent": userAgent,
  },
  body: JSON.stringify(payload),
});

const text = await response.text();
console.log("Status:", response.status);
console.log("Body:", text);

if (!response.ok) {
  process.exit(1);
}
