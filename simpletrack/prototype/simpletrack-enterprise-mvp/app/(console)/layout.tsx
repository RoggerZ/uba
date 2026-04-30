import type { ReactNode } from "react";
import { ConsoleLayout } from "../../src/components/ConsoleLayout";

export default function ConsoleRouteLayout({ children }: { children: ReactNode }) {
  return <ConsoleLayout>{children}</ConsoleLayout>;
}
