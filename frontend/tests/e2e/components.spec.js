/**
 * 组件功能端到端测试
 */

import { test, expect } from '@playwright/test';

test.describe('组件系统测试', () => {
  test.beforeEach(async ({ page }) => {
    // 登录到主页面
    await page.goto('/');
    await page.fill('input[name="secret"]', '12345678');
    await page.click('button[type="submit"]');
    await page.waitForSelector('#main-app', { state: 'visible' });
  });

  test('应该正确显示测试按钮', async ({ page }) => {
    // 检查测试按钮是否存在
    await expect(page.locator('button:has-text("测试 Modal 组件")')).toBeVisible();
    await expect(page.locator('button:has-text("测试 DataTable 组件")')).toBeVisible();
    await expect(page.locator('button:has-text("测试 Form 组件")')).toBeVisible();
    await expect(page.locator('button:has-text("测试状态管理")')).toBeVisible();
  });

  test('Modal组件测试', async ({ page }) => {
    // 点击Modal测试按钮
    await page.click('button:has-text("测试 Modal 组件")');
    
    // 检查Modal是否出现
    await expect(page.locator('.modal, [role="dialog"]')).toBeVisible();
    await expect(page.locator('h3:has-text("Modal 组件测试")')).toBeVisible();
    
    // 检查Modal内容
    await expect(page.locator('text=这是使用 Tailwind CSS 优化后的 Modal 组件')).toBeVisible();
    
    // 检查按钮
    await expect(page.locator('button:has-text("太棒了！")')).toBeVisible();
    await expect(page.locator('button:has-text("关闭")')).toBeVisible();
    
    // 点击确认按钮
    await page.click('button:has-text("太棒了！")');
    
    // 检查Modal是否关闭
    await expect(page.locator('.modal, [role="dialog"]')).toBeHidden();
    
    // 检查成功消息
    await expect(page.locator('text=Modal 组件测试成功')).toBeVisible();
  });

  test('DataTable组件测试', async ({ page }) => {
    // 点击DataTable测试按钮
    await page.click('button:has-text("测试 DataTable 组件")');
    
    // 检查Modal和表格是否出现
    await expect(page.locator('h3:has-text("DataTable 组件测试")')).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
    
    // 检查表格内容
    await expect(page.locator('th:has-text("ID")')).toBeVisible();
    await expect(page.locator('th:has-text("名称")')).toBeVisible();
    await expect(page.locator('th:has-text("状态")')).toBeVisible();
    
    // 检查测试数据
    await expect(page.locator('td:has-text("测试配置1")')).toBeVisible();
    await expect(page.locator('td:has-text("测试配置2")')).toBeVisible();
    await expect(page.locator('td:has-text("测试配置3")')).toBeVisible();
    
    // 检查状态显示
    await expect(page.locator('text=启用')).toBeVisible();
    await expect(page.locator('text=禁用')).toBeVisible();
    
    // 检查搜索框
    await expect(page.locator('input[placeholder*="搜索"]')).toBeVisible();
    
    // 检查复选框
    await expect(page.locator('input[type="checkbox"]')).toHaveCount(4); // 3行数据 + 1个全选
    
    // 关闭Modal
    await page.click('button:has-text("确定")');
    await expect(page.locator('h3:has-text("DataTable 组件测试")')).toBeHidden();
  });

  test('状态管理测试', async ({ page }) => {
    // 监听控制台消息
    const consoleMessages = [];
    page.on('console', msg => {
      if (msg.text().includes('计数器变化')) {
        consoleMessages.push(msg.text());
      }
    });
    
    // 点击状态管理测试按钮
    await page.click('button:has-text("测试状态管理")');
    
    // 检查开始消息
    await expect(page.locator('text=状态管理测试开始')).toBeVisible();
    
    // 等待状态管理测试完成（5秒）
    await page.waitForTimeout(6000);
    
    // 检查完成消息
    await expect(page.locator('text=状态管理测试完成')).toBeVisible();
    
    // 验证控制台消息（应该有5个计数器变化消息）
    expect(consoleMessages.length).toBeGreaterThanOrEqual(5);
  });

  test('标签页切换功能', async ({ page }) => {
    // 检查默认标签页
    await expect(page.locator('[data-tab="overview"]')).toHaveClass(/border-primary-500/);
    await expect(page.locator('#overview-tab')).toBeVisible();
    
    // 切换到代理配置标签页
    await page.click('[data-tab="proxy-config"]');
    await expect(page.locator('[data-tab="proxy-config"]')).toHaveClass(/border-primary-500/);
    await expect(page.locator('#proxy-config-tab')).toBeVisible();
    await expect(page.locator('#overview-tab')).toBeHidden();
    
    // 切换到日志标签页
    await page.click('[data-tab="logs"]');
    await expect(page.locator('[data-tab="logs"]')).toHaveClass(/border-primary-500/);
    await expect(page.locator('#logs-tab')).toBeVisible();
    await expect(page.locator('#proxy-config-tab')).toBeHidden();
    
    // 切换回概览标签页
    await page.click('[data-tab="overview"]');
    await expect(page.locator('[data-tab="overview"]')).toHaveClass(/border-primary-500/);
    await expect(page.locator('#overview-tab')).toBeVisible();
    await expect(page.locator('#logs-tab')).toBeHidden();
  });

  test('消息提示功能', async ({ page }) => {
    // 点击Modal测试按钮
    await page.click('button:has-text("测试 Modal 组件")');
    
    // 点击确认按钮触发消息提示
    await page.click('button:has-text("太棒了！")');
    
    // 检查消息提示
    const messageContainer = page.locator('#message-container');
    await expect(messageContainer).toBeVisible();
    
    const message = messageContainer.locator('text=Modal 组件测试成功');
    await expect(message).toBeVisible();
    
    // 等待消息自动消失
    await page.waitForTimeout(4000);
    await expect(message).toBeHidden();
  });

  test('响应式设计测试', async ({ page }) => {
    // 测试桌面视图
    await page.setViewportSize({ width: 1200, height: 800 });
    await expect(page.locator('.container')).toBeVisible();
    
    // 测试平板视图
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.container')).toBeVisible();
    
    // 测试手机视图
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('.container')).toBeVisible();
    
    // 检查导航在小屏幕上是否正常工作
    await page.click('[data-tab="proxy-config"]');
    await expect(page.locator('#proxy-config-tab')).toBeVisible();
  });

  test('键盘导航测试', async ({ page }) => {
    // 使用Tab键导航
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    
    // 检查焦点是否正确
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
    
    // 使用Enter键激活按钮
    await page.keyboard.press('Enter');
    
    // 检查是否有相应的反应（取决于焦点在哪个元素上）
  });

  test('错误处理测试', async ({ page }) => {
    // 监听控制台错误
    const errors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    // 执行一些操作
    await page.click('button:has-text("测试 Modal 组件")');
    await page.click('button:has-text("太棒了！")');
    
    // 检查是否有JavaScript错误
    expect(errors.length).toBe(0);
  });
});
