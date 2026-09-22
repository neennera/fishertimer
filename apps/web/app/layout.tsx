import { Metadata } from "next";
import { fontVariables } from "./fonts";
import "./globals.css";

export const metadata: Metadata = {
  title: "Fisher Timer",
  description:
    "A community study timer. Join a room, run your own Pomodoro, and catch a fish for every cycle you finish.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // The font variables go on <html>, not <body>. tokens.css maps them to the
  // semantic roles (--font-display etc.) in a :root block, and a custom
  // property declared on :root can only resolve references that also exist
  // on :root. Move these to <body> and every font silently falls back.
  return (
    <html lang="en" className={fontVariables}>
      <body>{children}</body>
    </html>
  );
}
