/**
 * Playwright全局清理
 * 在所有测试结束后执行
 */

async function globalTeardown(config) {
  console.log('🧹 开始端到端测试清理...');

  // 这里可以添加清理逻辑
  // 例如：清理测试数据、关闭服务等

  console.log('✅ 端到端测试清理完成');
}

export default globalTeardown;
