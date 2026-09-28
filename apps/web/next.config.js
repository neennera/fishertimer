/* global process */
/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    const gatewayUrl = process.env.API_GATEWAY_URL || 'http://localhost:8000';
    const adminUrl = process.env.ADMIN_SERVICE_URL || 'http://localhost:8087';

    return [
      // Client-side routes routed through the API Gateway
      {
        source: '/api/auth',
        destination: `${gatewayUrl}/api/auth`,
      },
      {
        source: '/api/auth/:path*',
        destination: `${gatewayUrl}/api/auth/:path*`,
      },
      {
        source: '/api/session',
        destination: `${gatewayUrl}/api/session`,
      },
      {
        source: '/api/session/:path*',
        destination: `${gatewayUrl}/api/session/:path*`,
      },
      {
        source: '/api/timer',
        destination: `${gatewayUrl}/api/timer`,
      },
      {
        source: '/api/timer/:path*',
        destination: `${gatewayUrl}/api/timer/:path*`,
      },
      {
        source: '/api/reward',
        destination: `${gatewayUrl}/api/reward`,
      },
      {
        source: '/api/reward/:path*',
        destination: `${gatewayUrl}/api/reward/:path*`,
      },
      {
        source: '/api/leaderboard',
        destination: `${gatewayUrl}/api/leaderboard`,
      },
      {
        source: '/api/leaderboard/:path*',
        destination: `${gatewayUrl}/api/leaderboard/:path*`,
      },
      // Admin-side routes call Admin Service directly
      {
        source: '/api/admin/:path*',
        destination: `${adminUrl}/api/v1/admin/:path*`,
      },
    ];
  },
};

export default nextConfig;
