/**
 * Playwright全局设置
 * 在所有测试开始前执行
 */

import { chromium } from '@playwright/test';

async function globalSetup(config) {
  console.log('🚀 开始端到端测试设置...');

  // 等待服务器启动
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // 检查服务器是否可用
    console.log('⏳ 等待服务器启动...');
    const baseURL = config.use?.baseURL || 'http://localhost:10805';
    await page.goto(baseURL, {
      waitUntil: 'networkidle',
      timeout: 60000
    });

    console.log('✅ 服务器已启动');

    // 预热应用
    console.log('🔥 预热应用...');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);

    console.log('✅ 应用预热完成');

  } catch (error) {
    console.error('❌ 全局设置失败:', error);
    throw error;
  } finally {
    await browser.close();
  }

  console.log('✅ 端到端测试设置完成');
}

export default globalSetup;
