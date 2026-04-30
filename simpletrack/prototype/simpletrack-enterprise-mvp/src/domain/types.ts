export type Health = "healthy" | "review" | "draft" | "reserved" | "active";

export interface SiteConfig {
  id: string;
  name: string;
  domain: string;
  environment: "Production" | "Staging";
  trackerVersion: string;
  lastSeenAt: string;
}

export interface KpiMetric {
  label: string;
  value: string;
  delta: string;
  trend: "up" | "down";
}

export interface TrafficPoint {
  date: string;
  pageviews: number;
  visitors: number;
  events: number;
}

export interface LiveSignal {
  id: string;
  time: string;
  type: "pageview" | "event";
  name: string;
  path: string;
  visitor: string;
  status: "accepted" | "quarantined";
}

export interface AnalyticsEvent {
  key: string;
  name: string;
  description: string;
  count: number;
  visitors: number;
  lastSeen: string;
  health: Health;
  properties: Record<string, Record<string, number>>;
}

export interface GoalDefinition {
  id: string;
  name: string;
  type: "event" | "page";
  rule: string;
  denominator: string;
  conversions: number;
  population: number;
  rate: string;
  status: "active" | "draft";
}

export interface DictionaryEvent {
  name: string;
  status: Health;
  required: string;
}

export interface DictionaryProperty {
  key: string;
  type: "enum" | "number" | "string" | "boolean";
  values: string;
}

export interface IngestionRule {
  rule: string;
  detail: string;
  mode: "enforced" | "review";
}
