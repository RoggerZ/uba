import { create } from "zustand";
import { persist } from "zustand/middleware";
import { goals, liveSignals, siteConfig } from "../domain/mockData";
import type { GoalDefinition, LiveSignal, SiteConfig } from "../domain/types";

interface ConsoleState {
  site: SiteConfig;
  installVerified: boolean;
  livePaused: boolean;
  selectedEventKey: string;
  selectedProperty: string;
  selectedGoalId: string;
  allowedDomains: string[];
  liveSignals: LiveSignal[];
  goals: GoalDefinition[];
  setSite: (site: SiteConfig) => void;
  setInstallVerified: (verified: boolean) => void;
  setLivePaused: (paused: boolean) => void;
  selectEvent: (eventKey: string, property?: string) => void;
  selectProperty: (property: string) => void;
  selectGoal: (goalId: string) => void;
  updateGoal: (goalId: string, patch: Partial<GoalDefinition>) => void;
  addAllowedDomain: (domain: string) => void;
  pushLiveSignal: () => void;
  reset: () => void;
}

const initialState = {
  site: siteConfig,
  installVerified: true,
  livePaused: false,
  selectedEventKey: "first_event_sent",
  selectedProperty: "plan",
  selectedGoalId: "goal_first_event",
  allowedDomains: ["acme.example", "app.acme.example"],
  liveSignals,
  goals,
};

export const useConsoleStore = create<ConsoleState>()(
  persist(
    (set, get) => ({
      ...initialState,
      setSite: (site) => set({ site }),
      setInstallVerified: (installVerified) => set({ installVerified }),
      setLivePaused: (livePaused) => set({ livePaused }),
      selectEvent: (selectedEventKey, selectedProperty = "plan") => set({ selectedEventKey, selectedProperty }),
      selectProperty: (selectedProperty) => set({ selectedProperty }),
      selectGoal: (selectedGoalId) => set({ selectedGoalId }),
      updateGoal: (goalId, patch) =>
        set({
          goals: get().goals.map((goal) => (goal.id === goalId ? { ...goal, ...patch } : goal)),
        }),
      addAllowedDomain: (domain) => {
        const allowedDomains = get().allowedDomains;
        if (allowedDomains.includes(domain)) return;
        set({ allowedDomains: [...allowedDomains, domain] });
      },
      pushLiveSignal: () => {
        if (get().livePaused) return;
        const samples: Omit<LiveSignal, "id" | "time">[] = [
          { type: "pageview", name: "Pageview", path: "/docs/install", visitor: "v_9041", status: "accepted" },
          { type: "event", name: "install_started", path: "/app/install", visitor: "v_9102", status: "accepted" },
          { type: "event", name: "first_event_sent", path: "/app/install", visitor: "v_9268", status: "accepted" },
        ];
        const sample = samples[Math.floor(Math.random() * samples.length)];
        const next: LiveSignal = {
          id: `sig-${Date.now()}`,
          time: "now",
          ...sample,
        };
        set({ liveSignals: [next, ...get().liveSignals.slice(0, 6)] });
      },
      reset: () => set(initialState),
    }),
    {
      name: "simpletrack-enterprise-console",
    },
  ),
);
