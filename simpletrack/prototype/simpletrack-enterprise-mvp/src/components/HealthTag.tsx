import { Tag } from "antd";
import type { Health } from "../domain/types";

interface HealthTagProps {
  value: Health | "accepted" | "quarantined" | "enforced";
}

export function HealthTag({ value }: HealthTagProps) {
  const color =
    value === "healthy" || value === "active" || value === "accepted" || value === "enforced"
      ? "success"
      : value === "review" || value === "draft" || value === "reserved"
        ? "warning"
        : "default";

  return <Tag color={color}>{value}</Tag>;
}
