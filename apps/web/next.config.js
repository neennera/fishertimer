/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    const authUrl = process.env.AUTH_SERVICE_URL || 'http://localhost:8081';
    const accountUrl = process.env.ACCOUNT_SERVICE_URL || 'http://localhost:8082';
    const sessionUrl = process.env.SESSION_SERVICE_URL || 'http://localhost:8083';
    const timerUrl = process.env.TIMER_SERVICE_URL || 'http://localhost:8084';
    const rewardUrl = process.env.REWARD_SERVICE_URL || 'http://localhost:8085';
    const leaderboardUrl = process.env.LEADERBOARD_SERVICE_URL || 'http://localhost:8086';
    const adminUrl = process.env.ADMIN_SERVICE_URL || 'http://localhost:8087';

    return [
      {
        source: '/api/auth/:path*',
        destination: `${authUrl}/api/v1/auth/:path*`,
      },
      {
        source: '/api/account/:path*',
        destination: `${accountUrl}/api/v1/account/:path*`,
      },
      {
        source: '/api/session/:path*',
        destination: `${sessionUrl}/api/v1/study-session/:path*`,
      },
      {
        source: '/api/timer/:path*',
        destination: `${timerUrl}/api/v1/study-timer/:path*`,
      },
      {
        source: '/api/reward/:path*',
        destination: `${rewardUrl}/api/v1/reward/:path*`,
      },
      {
        source: '/api/leaderboard/:path*',
        destination: `${leaderboardUrl}/api/v1/leaderboard/:path*`,
      },
      {
        source: '/api/admin/:path*',
        destination: `${adminUrl}/api/v1/admin/:path*`,
      },
    ];
  },
};

export default nextConfig;
