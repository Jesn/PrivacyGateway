import { defineConfig } from 'vite'
import { resolve } from 'path'
import legacy from '@vitejs/plugin-legacy'

export default defineConfig({
  // 基础配置
  base: '/assets/',
  
  // 开发服务器配置
  server: {
    port: 3000,
    host: true,
    proxy: {
      // 代理API请求到Go后端
      '/config': {
        target: 'http://localhost:10805',
        changeOrigin: true
      },
      '/logs': {
        target: 'http://localhost:10805',
        changeOrigin: true
      },
      '/tokens': {
        target: 'http://localhost:10805',
        changeOrigin: true
      },
      '/proxy': {
        target: 'http://localhost:10805',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:10805',
        ws: true
      }
    }
  },
  
  // 构建配置
  build: {
    outDir: 'dist',
    assetsDir: '',
    sourcemap: true,
    minify: 'terser',
    target: 'es2015',
    
    // 代码分割配置
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html')
      },
      output: {
        // 分包策略
        manualChunks: {
          // 第三方库
          vendor: ['axios', 'dayjs'],
          
          // 工具函数
          utils: [
            './assets/js/utils/auth.js',
            './assets/js/utils/http.js',
            './assets/js/utils/dom.js',
            './assets/js/utils/date.js',
            './assets/js/utils/validation.js'
          ],
          
          // 状态管理
          state: [
            './assets/js/utils/StateManager.js',
            './assets/js/utils/EventBus.js',
            './assets/js/utils/ComponentManager.js'
          ],
          
          // 组件
          components: [
            './assets/js/components/Component.js',
            './assets/js/components/Modal.js',
            './assets/js/components/DataTable.js',
            './assets/js/components/Form.js'
          ],
          
          // API层
          api: [
            './assets/js/api/proxy.js',
            './assets/js/api/logs.js',
            './assets/js/api/tokens.js'
          ]
        },
        
        // 文件命名
        chunkFileNames: 'js/[name]-[hash].js',
        entryFileNames: 'js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.')
          const ext = info[info.length - 1]
          
          if (/\.(css)$/.test(assetInfo.name)) {
            return `css/[name]-[hash].${ext}`
          }
          
          if (/\.(png|jpe?g|gif|svg|webp|ico)$/.test(assetInfo.name)) {
            return `images/[name]-[hash].${ext}`
          }
          
          if (/\.(woff2?|eot|ttf|otf)$/.test(assetInfo.name)) {
            return `fonts/[name]-[hash].${ext}`
          }
          
          return `assets/[name]-[hash].${ext}`
        }
      }
    },
    
    // 压缩配置
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    }
  },
  
  // 插件配置
  plugins: [
    // 兼容性插件
    legacy({
      targets: ['defaults', 'not IE 11']
    })
  ],
  
  // 路径解析
  resolve: {
    alias: {
      '@': resolve(__dirname, './assets'),
      '@js': resolve(__dirname, './assets/js'),
      '@css': resolve(__dirname, './assets/css'),
      '@components': resolve(__dirname, './assets/js/components'),
      '@utils': resolve(__dirname, './assets/js/utils'),
      '@api': resolve(__dirname, './assets/js/api'),
      '@modules': resolve(__dirname, './assets/js/modules')
    }
  },
  
  // CSS配置
  css: {
    postcss: './postcss.config.js',
    devSourcemap: true
  },
  
  // 优化配置
  optimizeDeps: {
    include: [
      'axios',
      'dayjs'
    ]
  },
  
  // 环境变量
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '2.0.0'),
    __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
    __DEV__: JSON.stringify(process.env.NODE_ENV === 'development')
  },
  
  // 预览配置
  preview: {
    port: 4173,
    host: true
  }
})
