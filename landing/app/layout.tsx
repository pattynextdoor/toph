import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Space_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const spaceMono = Space_Mono({
  variable: "--font-space-mono",
  subsets: ["latin"],
  weight: ["400", "700"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://toph.dev"),
  title: "toph — btop for AI agents",
  description:
    "A terminal dashboard for AI coding agents. See what your agents are doing. Real-time activity feed, token tracking, cost estimation. Zero config.",
  keywords: [
    "terminal",
    "dashboard",
    "AI",
    "coding agents",
    "Claude Code",
    "btop",
    "TUI",
    "Go",
    "Bubble Tea",
  ],
  authors: [{ name: "pattynextdoor" }],
  openGraph: {
    title: "toph — btop for AI agents",
    description:
      "A terminal dashboard for AI coding agents. See what your agents are doing.",
    type: "website",
    url: "https://toph.dev",
    images: [{ url: "/og.png", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "toph — btop for AI agents",
    description: "A terminal dashboard for AI coding agents.",
    images: ["/og.png"],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} ${spaceMono.variable} dark h-full antialiased`}
    >
      <body className="min-h-full bg-zinc-950 text-zinc-50">{children}</body>
    </html>
  );
}
