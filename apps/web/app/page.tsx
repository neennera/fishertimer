import Image, { type ImageProps } from "next/image";
import { Button } from "@repo/ui/button";
import styles from "./page.module.css";

type Props = Omit<ImageProps, "src"> & {
  srcLight: string;
  srcDark: string;
};

const ThemeImage = (props: Props) => {
  const { srcLight, srcDark, ...rest } = props;

  return (
    <>
      <Image {...rest} src={srcLight} className="imgLight" />
      <Image {...rest} src={srcDark} className="imgDark" />
    </>
  );
};

export default function Home() {
  return (
    <div className={styles.page}>
      <main className={styles.main}>
        <ThemeImage
          className={styles.logo}
          srcLight="turborepo-dark.svg"
          srcDark="turborepo-light.svg"
          alt="Turborepo logo"
          width={180}
          height={38}
          priority
        />
        <div style={{ textAlign: "center", marginBottom: "1rem" }}>
          <h1 style={{ fontSize: "2rem", margin: "0.5rem 0" }}>🎣 Fisher Timer — Student Portal</h1>
          <p style={{ color: "#888", margin: 0 }}>
            Frontend Website (Client page) — Communicating via API Gateway (Port 8000)
          </p>
        </div>

        <div style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
          gap: "1rem",
          width: "100%",
          maxWidth: "700px"
        }}>
          <div style={{ padding: "1rem", border: "1px solid rgba(255,255,255,0.1)", borderRadius: "8px" }}>
            <h3>⏱️ Study Timer</h3>
            <p style={{ fontSize: "0.85rem", color: "#aaa" }}>Focus & rest cycles synced via API Gateway</p>
          </div>
          <div style={{ padding: "1rem", border: "1px solid rgba(255,255,255,0.1)", borderRadius: "8px" }}>
            <h3>👥 Study Sessions</h3>
            <p style={{ fontSize: "0.85rem", color: "#aaa" }}>Join & create study rooms with peers</p>
          </div>
          <div style={{ padding: "1rem", border: "1px solid rgba(255,255,255,0.1)", borderRadius: "8px" }}>
            <h3>🐟 FishTank Rewards</h3>
            <p style={{ fontSize: "0.85rem", color: "#aaa" }}>Earn fish rewards upon ending sessions</p>
          </div>
          <div style={{ padding: "1rem", border: "1px solid rgba(255,255,255,0.1)", borderRadius: "8px" }}>
            <h3>🏆 Leaderboard</h3>
            <p style={{ fontSize: "0.85rem", color: "#aaa" }}>Weekly & all-time focus rankings</p>
          </div>
        </div>

        <div className={styles.ctas}>
          <a
            className={styles.primary}
            href="/admin"
          >
            🛡️ Switch to Admin Portal
          </a>
          <Button appName="web" className={styles.secondary}>
            Start Studying
          </Button>
        </div>
      </main>
      <footer className={styles.footer}>
        <a
          href="https://vercel.com/templates?search=turborepo&utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"
          target="_blank"
          rel="noopener noreferrer"
        >
          <Image
            aria-hidden
            src="/window.svg"
            alt="Window icon"
            width={16}
            height={16}
          />
          Examples
        </a>
        <a
          href="https://turborepo.dev?utm_source=create-turbo"
          target="_blank"
          rel="noopener noreferrer"
        >
          <Image
            aria-hidden
            src="/globe.svg"
            alt="Globe icon"
            width={16}
            height={16}
          />
          Go to turborepo.dev →
        </a>
      </footer>
    </div>
  );
}
