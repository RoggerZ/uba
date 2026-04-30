import { Card } from "antd";
import type { PropsWithChildren, ReactNode } from "react";

interface SectionCardProps extends PropsWithChildren {
  title: ReactNode;
  extra?: ReactNode;
}

export function SectionCard({ title, extra, children }: SectionCardProps) {
  return (
    <Card className="section-card" title={title} extra={extra} bordered>
      {children}
    </Card>
  );
}
