import Link from "next/link";
import { Header } from "../components/Header";
import { PixelPanel } from "../components/ui/PixelPanel";

/**
 * Placeholder home. Once the auth branch lands this redirects to /account
 * when a session exists and /signin when it doesn't (UC-06).
 */
export default function Home() {
  return (
    <>
      <Header />
      <main className="mx-auto w-full max-w-2xl px-4 py-12">
        <PixelPanel>
          <h1
            className="font-display leading-tight"
            style={{ fontSize: "calc(var(--px) * 10)" }}
          >
            Fisher Timer
          </h1>
          <p className="mt-2 text-bark">
            The design system is in place. Screens land on the next two
            branches.
          </p>
          <Link
            href="/styleguide"
            className="mt-5 inline-block underline underline-offset-4"
          >
            View the styleguide
          </Link>
        </PixelPanel>
      </main>
    </>
  );
}
