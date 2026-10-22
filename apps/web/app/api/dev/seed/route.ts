import { spawn } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { NextResponse } from 'next/server';

// Dev-only: runs services/reward/database/schemas/004_seed_current_user_month.js
// inside the reward MongoDB container (docker exec ... mongosh), the same
// command documented in the seed's header. Refuses to run in production, and
// only for a signed-in admin, who can only seed their own account.

export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';

const SEED_FILE = 'services/reward/database/schemas/004_seed_current_user_month.js';
const CONTAINER = process.env.REWARD_DB_CONTAINER || 'fishertimer-reward-db';
const MONGO_USER = process.env.REWARD_DB_USER || 'mongoadmin';
const MONGO_PASSWORD = process.env.REWARD_DB_PASSWORD || 'mongopassword';

async function readSeed(): Promise<string> {
  // `next dev` runs from apps/web; tolerate being started from the repo root.
  const candidates = [
    path.resolve(process.cwd(), '../..', SEED_FILE),
    path.resolve(process.cwd(), SEED_FILE),
  ];
  for (const file of candidates) {
    try {
      return await readFile(file, 'utf8');
    } catch {
      // try the next one
    }
  }
  throw new Error(`Seed script not found (${SEED_FILE})`);
}

function runMongosh(script: string, userId: string, displayName: string) {
  return new Promise<{ code: number | null; output: string }>((resolve, reject) => {
    // No shell: every value is its own argv entry.
    const child = spawn(
      'docker',
      [
        'exec',
        '-i',
        '-e',
        `SEED_USER_ID=${userId}`,
        '-e',
        `SEED_DISPLAY_NAME=${displayName}`,
        CONTAINER,
        'mongosh',
        '-u',
        MONGO_USER,
        '-p',
        MONGO_PASSWORD,
        '--authenticationDatabase',
        'admin',
        '--quiet',
      ],
      { windowsHide: true },
    );
    let output = '';
    child.stdout.on('data', (d) => (output += d));
    child.stderr.on('data', (d) => (output += d));
    child.on('error', reject);
    child.on('close', (code) => resolve({ code, output: output.trim() }));
    child.stdin.end(script);
  });
}

const GATEWAY_URL = process.env.API_GATEWAY_URL || 'http://localhost:8000';

/** The caller's account, asked of the Account service with their own cookie; null if not signed in. */
async function sessionUser(request: Request): Promise<{ user_id?: string; role?: string } | null> {
  const cookie = request.headers.get('cookie');
  if (!cookie) {
    return null;
  }
  try {
    const res = await fetch(`${GATEWAY_URL}/api/auth/me`, {
      headers: { cookie },
      cache: 'no-store',
      signal: AbortSignal.timeout(5000),
    });
    return res.ok ? ((await res.json()) as { user_id?: string; role?: string }) : null;
  } catch {
    return null;
  }
}

export async function POST(request: Request) {
  if (process.env.NODE_ENV === 'production') {
    return NextResponse.json({ error: 'Not available in production.' }, { status: 404 });
  }

  const caller = await sessionUser(request);
  if (!caller) {
    return NextResponse.json({ error: 'Sign in first.' }, { status: 401 });
  }
  if (caller.role !== 'ADMIN') {
    return NextResponse.json({ error: 'Admins only.' }, { status: 403 });
  }

  let body: { user_id?: unknown; display_name?: unknown };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: 'Invalid JSON body.' }, { status: 400 });
  }

  const userId = typeof body.user_id === 'string' ? body.user_id.trim() : '';
  const displayName = typeof body.display_name === 'string' ? body.display_name.trim() : '';
  if (!/^[A-Za-z0-9_-]{1,64}$/.test(userId)) {
    return NextResponse.json({ error: 'Missing or invalid user_id.' }, { status: 400 });
  }

  if (userId !== caller.user_id) {
    return NextResponse.json({ error: 'You can only seed your own account.' }, { status: 403 });
  }

  try {
    const script = await readSeed();
    const { code, output } = await runMongosh(script, userId, displayName.slice(0, 64) || 'You');
    if (code !== 0) {
      return NextResponse.json(
        { error: 'Seed failed.', output },
        { status: 500 },
      );
    }
    return NextResponse.json({ ok: true, output });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : 'Seed failed.' },
      { status: 500 },
    );
  }
}
