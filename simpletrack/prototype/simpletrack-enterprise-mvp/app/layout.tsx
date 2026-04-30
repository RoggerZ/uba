import { AntdRegistry } from "@ant-design/nextjs-registry";
import "antd/dist/reset.css";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Providers } from "./providers";
import "../src/styles.css";

export const metadata: Metadata = {
  title: "SimpleTrack Console",
  description: "Production-oriented SimpleTrack review prototype",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AntdRegistry>
          <Providers>{children}</Providers>
        </AntdRegistry>
      </body>
    </html>
  );
}
