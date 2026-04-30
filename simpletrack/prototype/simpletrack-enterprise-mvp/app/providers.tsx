"use client";

import { App as AntApp, ConfigProvider, theme } from "antd";
import type { ReactNode } from "react";

export function Providers({ children }: { children: ReactNode }) {
  return (
    <ConfigProvider
      theme={{
        algorithm: theme.compactAlgorithm,
        token: {
          colorPrimary: "#0f766e",
          borderRadius: 4,
          fontFamily:
            'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
          colorBgLayout: "#f5f6f8",
          colorText: "#17202e",
        },
        components: {
          Card: {
            borderRadiusLG: 4,
            headerBg: "#fbfcfd",
          },
          Layout: {
            siderBg: "#fbfcfd",
            headerBg: "#ffffff",
          },
          Table: {
            headerBg: "#f6f7f9",
            borderColor: "#d8dde5",
          },
        },
      }}
    >
      <AntApp>{children}</AntApp>
    </ConfigProvider>
  );
}
