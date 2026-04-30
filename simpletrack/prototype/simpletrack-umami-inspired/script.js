const rangeButtons = document.querySelectorAll(".range-button");
const copyButtons = document.querySelectorAll("[data-copy-target]");
const routeButtons = document.querySelectorAll("[data-route]");

const metricsByRange = {
  "24h": {
    visitors: "1,984",
    visits: "2,741",
    views: "7,328",
    bounce: "29%",
    duration: "2m 18s",
    deltas: {
      visitors: "+4.2%",
      visits: "+3.8%",
      views: "+6.1%",
      bounce: "-1.4%",
      duration: "+2.2%",
    },
  },
  "7d": {
    visitors: "12,842",
    visits: "19,511",
    views: "51,024",
    bounce: "31%",
    duration: "2m 42s",
    deltas: {
      visitors: "+8.4%",
      visits: "+6.1%",
      views: "+12.7%",
      bounce: "-3.2%",
      duration: "+9.5%",
    },
  },
  "30d": {
    visitors: "43,906",
    visits: "68,214",
    views: "182,003",
    bounce: "34%",
    duration: "2m 09s",
    deltas: {
      visitors: "+11.9%",
      visits: "+9.7%",
      views: "+14.4%",
      bounce: "-1.9%",
      duration: "+7.1%",
    },
  },
};

function applyRange(range) {
  const metrics = metricsByRange[range];
  if (!metrics) return;

  document.querySelectorAll("[data-metric]").forEach((node) => {
    node.textContent = metrics[node.dataset.metric];
  });

  document.querySelectorAll("[data-delta]").forEach((node) => {
    node.textContent = metrics.deltas[node.dataset.delta];
  });

  rangeButtons.forEach((button) => {
    button.classList.toggle("is-active", button.dataset.range === range);
  });
}

rangeButtons.forEach((button) => {
  button.addEventListener("click", () => applyRange(button.dataset.range));
});

routeButtons.forEach((button) => {
  button.addEventListener("click", () => {
    if (button.dataset.route) {
      window.location.href = button.dataset.route;
    }
  });
});

copyButtons.forEach((button) => {
  button.addEventListener("click", async () => {
    const selector = button.dataset.copyTarget;
    const target = document.querySelector(selector);
    if (!target) return;

    const value = target.textContent.trim();
    try {
      await navigator.clipboard.writeText(value);
      const original = button.textContent;
      button.textContent = "Copied";
      setTimeout(() => {
        button.textContent = original;
      }, 1400);
    } catch (error) {
      console.error(error);
    }
  });
});

applyRange("7d");
