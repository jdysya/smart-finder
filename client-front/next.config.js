/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  async rewrites() {
    return [
      {
        source: '/api/:path*',   // 前端访问：/api/xxx
        destination: 'http://localhost:8964/api/:path*',  // 实际代理到后端服务
      }
    ];
  },
};

module.exports = nextConfig;
