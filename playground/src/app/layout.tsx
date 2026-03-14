import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "driftabot playground",
  description: "Interactively compare OpenAPI, GraphQL, and gRPC schemas to detect breaking and non-breaking changes.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="antialiased">{children}</body>
    </html>
  );
}
