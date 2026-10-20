// One switch for "real backend or canned data", read by lib/auth.ts and
// lib/profile.ts. NEXT_PUBLIC_USE_MOCKS=false is the real backend, true is mocks.
// Unset: mocks in development (so the UI runs with nothing else started), the
// real backend in a production build, so a missing variable can never ship fake
// sign-ins.
const flag = process.env.NEXT_PUBLIC_USE_MOCKS;

export const USE_MOCKS = flag === 'true' || (flag !== 'false' && process.env.NODE_ENV !== 'production');
