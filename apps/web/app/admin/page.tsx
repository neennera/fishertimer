import Link from 'next/link';
import styles from '../page.module.css';

export default function AdminPage() {
  return (
    <div className={styles.page}>
      <main className={styles.main}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <span style={{ fontSize: '2.5rem' }}>🛡️</span>
          <div>
            <h1 style={{ margin: 0, fontSize: '1.8rem' }}>
              Fisher Timer — Admin Moderation Portal
            </h1>
            <p style={{ margin: '0.25rem 0 0 0', color: '#888' }}>
              Frontend Website (Admin Page) — Direct connection to Admin Service (Port 8087)
            </p>
          </div>
        </div>

        <div
          style={{
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '8px',
            padding: '1.5rem',
            width: '100%',
            maxWidth: '650px',
            display: 'flex',
            flexDirection: 'column',
            gap: '1rem',
          }}
        >
          <div style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.1)', paddingBottom: '0.75rem' }}>
            <h3 style={{ margin: 0 }}>Direct Service Communication</h3>
            <p style={{ margin: '0.25rem 0 0 0', fontSize: '0.9rem', color: '#aaa' }}>
              As designed in the architecture diagram, the Admin Website does <strong>not</strong> go through the API Gateway. It connects directly to <code>services/admin</code> (Port 8087).
            </p>
          </div>

          <div>
            <h4 style={{ margin: '0 0 0.5rem 0' }}>Admin Operations</h4>
            <ul style={{ margin: 0, paddingLeft: '1.25rem', lineHeight: '1.8' }}>
              <li><strong>ViewActiveSessions:</strong> Inspect live study room states.</li>
              <li><strong>ViewParticipants:</strong> List participants currently in an active room.</li>
              <li><strong>MonitorTimerStatus:</strong> Real-time timer inspection across participants.</li>
              <li><strong>ReviewReport & BanUser:</strong> Moderate users and update ban flags directly in Account Service.</li>
              <li><strong>KickParticipant:</strong> Command Study Session service to kick participants.</li>
            </ul>
          </div>
        </div>

        <div className={styles.ctas}>
          <Link href="/" className={styles.secondary}>
            ← Return to Client Portal
          </Link>
        </div>
      </main>
    </div>
  );
}
