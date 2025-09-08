import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright端到端测试配置
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  // 测试目录
  testDir: './tests/e2e',
  
  // 全局设置
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  
  // 报告器
  reporter: [
    ['html'],
    ['json', { outputFile: 'test-results/e2e-results.json' }],
    ['junit', { outputFile: 'test-results/e2e-results.xml' }]
  ],
  
  // 全局测试配置
  use: {
    // 基础URL
    baseURL: 'http://localhost:10805',
    
    // 浏览器配置
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    
    // 等待配置
    actionTimeout: 10000,
    navigationTimeout: 30000,
    
    // 忽略HTTPS错误
    ignoreHTTPSErrors: true,
    
    // 用户代理
    userAgent: 'Playwright E2E Tests'
  },

  // 项目配置 - 不同浏览器
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    
    // 移动端测试
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  // 开发服务器配置
  webServer: {
    command: 'cd .. && ADMIN_SECRET=12345678 go run main.go',
    url: 'http://localhost:10805',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
  
  // 输出目录
  outputDir: 'test-results/',
  
  // 全局设置和清理
  globalSetup: './tests/e2e/global-setup.js',
  globalTeardown: './tests/e2e/global-teardown.js',
});
