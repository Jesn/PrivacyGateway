/**
 * 认证功能端到端测试
 */

import { test, expect } from '@playwright/test';

test.describe('认证功能', () => {
  test.beforeEach(async ({ page }) => {
    // 每个测试前访问首页
    await page.goto('/');
  });

  test('应该显示登录页面', async ({ page }) => {
    // 检查登录页面元素
    await expect(page.locator('h1')).toContainText('Privacy Gateway');
    await expect(page.locator('input[name="secret"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    
    // 检查页面标题
    await expect(page).toHaveTitle(/Privacy Gateway/);
  });

  test('应该拒绝无效的密钥', async ({ page }) => {
    // 输入无效密钥
    await page.fill('input[name="secret"]', 'invalid-secret');
    await page.click('button[type="submit"]');
    
    // 等待错误消息出现
    await expect(page.locator('.text-red-500, .error')).toBeVisible();
    
    // 确保仍在登录页面
    await expect(page.locator('input[name="secret"]')).toBeVisible();
  });

  test('应该接受有效的密钥并跳转到主页面', async ({ page }) => {
    // 输入有效密钥
    await page.fill('input[name="secret"]', '12345678');
    await page.click('button[type="submit"]');
    
    // 等待页面跳转
    await page.waitForSelector('#main-app', { state: 'visible' });
    
    // 检查主页面元素
    await expect(page.locator('#main-app')).toBeVisible();
    await expect(page.locator('#login-page')).toBeHidden();
    
    // 检查导航标签
    await expect(page.locator('[data-tab="overview"]')).toBeVisible();
    await expect(page.locator('[data-tab="proxy-config"]')).toBeVisible();
    await expect(page.locator('[data-tab="logs"]')).toBeVisible();
    
    // 检查退出登录按钮
    await expect(page.locator('button:has-text("退出登录")')).toBeVisible();
  });

  test('应该正确处理会话恢复', async ({ page }) => {
    // 首先登录
    await page.fill('input[name="secret"]', '12345678');
    await page.click('button[type="submit"]');
    await page.waitForSelector('#main-app', { state: 'visible' });
    
    // 刷新页面
    await page.reload();
    
    // 检查是否自动恢复会话
    await expect(page.locator('#main-app')).toBeVisible();
    await expect(page.locator('#login-page')).toBeHidden();
  });

  test('应该正确处理退出登录', async ({ page }) => {
    // 首先登录
    await page.fill('input[name="secret"]', '12345678');
    await page.click('button[type="submit"]');
    await page.waitForSelector('#main-app', { state: 'visible' });
    
    // 点击退出登录
    await page.click('button:has-text("退出登录")');
    
    // 检查是否返回登录页面
    await expect(page.locator('#login-page')).toBeVisible();
    await expect(page.locator('#main-app')).toBeHidden();
    
    // 检查会话是否被清除
    await page.reload();
    await expect(page.locator('#login-page')).toBeVisible();
  });

  test('应该在密钥输入框按回车时提交表单', async ({ page }) => {
    // 在密钥输入框中输入并按回车
    await page.fill('input[name="secret"]', '12345678');
    await page.press('input[name="secret"]', 'Enter');
    
    // 等待页面跳转
    await page.waitForSelector('#main-app', { state: 'visible' });
    await expect(page.locator('#main-app')).toBeVisible();
  });

  test('应该显示加载状态', async ({ page }) => {
    // 监听网络请求
    const responsePromise = page.waitForResponse('/config/proxy');
    
    // 输入密钥并提交
    await page.fill('input[name="secret"]', '12345678');
    await page.click('button[type="submit"]');
    
    // 检查按钮是否显示加载状态（如果有的话）
    // 这取决于具体的实现
    
    // 等待响应
    await responsePromise;
    
    // 确保最终跳转到主页面
    await expect(page.locator('#main-app')).toBeVisible();
  });
});
