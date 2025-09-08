import { defineConfig } from 'vitest/config'
import { resolve } from 'path'

export default defineConfig({
  test: {
    // 测试环境配置
    environment: 'happy-dom', // 或者使用 'jsdom'
    
    // 全局设置
    globals: true,
    
    // 测试文件匹配模式
    include: [
      'tests/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}',
      'assets/js/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'
    ],

    // 排除文件
    exclude: [
      'node_modules',
      'dist',
      '.git',
      '.cache',
      'tests/e2e/**/*' // 排除E2E测试
    ],
    
    // 覆盖率配置
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      reportsDirectory: './coverage',
      exclude: [
        'node_modules/',
        'tests/',
        'dist/',
        '**/*.config.js',
        '**/*.config.ts'
      ],
      thresholds: {
        global: {
          branches: 70,
          functions: 70,
          lines: 70,
          statements: 70
        }
      }
    },
    
    // 测试超时
    testTimeout: 10000,
    hookTimeout: 10000,
    
    // 并发设置
    threads: true,
    maxThreads: 4,
    minThreads: 1,
    
    // 监听模式配置
    watch: {
      ignore: ['node_modules/**', 'dist/**']
    },
    
    // 设置文件
    setupFiles: ['./tests/setup.js'],
    
    // 模拟配置
    clearMocks: true,
    restoreMocks: true,
    
    // 报告器
    reporter: ['verbose', 'json', 'html'],
    outputFile: {
      json: './test-results.json',
      html: './test-results.html'
    }
  },
  
  // 路径解析（与主配置保持一致）
  resolve: {
    alias: {
      '@': resolve(__dirname, './assets'),
      '@js': resolve(__dirname, './assets/js'),
      '@css': resolve(__dirname, './assets/css'),
      '@components': resolve(__dirname, './assets/js/components'),
      '@utils': resolve(__dirname, './assets/js/utils'),
      '@api': resolve(__dirname, './assets/js/api'),
      '@modules': resolve(__dirname, './assets/js/modules'),
      '@tests': resolve(__dirname, './tests')
    }
  },
  
  // 环境变量
  define: {
    __TEST__: true,
    __DEV__: false
  }
})
