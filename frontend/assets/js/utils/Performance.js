/**
 * 性能监控和优化工具
 * 提供性能监控、懒加载、防抖节流等功能
 */

class Performance {
    constructor() {
        this.metrics = new Map();
        this.observers = new Map();
        this.timers = new Map();
        this.isSupported = this.checkSupport();
        
        if (this.isSupported) {
            this.initPerformanceObserver();
        }
    }

    /**
     * 检查浏览器支持
     * @private
     */
    checkSupport() {
        return typeof window !== 'undefined' && 
               'performance' in window && 
               'PerformanceObserver' in window;
    }

    /**
     * 初始化性能观察器
     * @private
     */
    initPerformanceObserver() {
        try {
            // 监控导航性能
            const navObserver = new PerformanceObserver((list) => {
                const entries = list.getEntries();
                entries.forEach(entry => {
                    this.recordMetric('navigation', {
                        type: entry.type,
                        duration: entry.duration,
                        startTime: entry.startTime,
                        loadEventEnd: entry.loadEventEnd,
                        domContentLoadedEventEnd: entry.domContentLoadedEventEnd
                    });
                });
            });
            navObserver.observe({ entryTypes: ['navigation'] });
            this.observers.set('navigation', navObserver);

            // 监控资源加载性能
            const resourceObserver = new PerformanceObserver((list) => {
                const entries = list.getEntries();
                entries.forEach(entry => {
                    this.recordMetric('resource', {
                        name: entry.name,
                        duration: entry.duration,
                        size: entry.transferSize,
                        type: this.getResourceType(entry.name)
                    });
                });
            });
            resourceObserver.observe({ entryTypes: ['resource'] });
            this.observers.set('resource', resourceObserver);

            // 监控用户交互性能
            if ('PerformanceEventTiming' in window) {
                const eventObserver = new PerformanceObserver((list) => {
                    const entries = list.getEntries();
                    entries.forEach(entry => {
                        this.recordMetric('event', {
                            name: entry.name,
                            duration: entry.duration,
                            startTime: entry.startTime,
                            processingStart: entry.processingStart,
                            processingEnd: entry.processingEnd
                        });
                    });
                });
                eventObserver.observe({ entryTypes: ['event'] });
                this.observers.set('event', eventObserver);
            }
        } catch (error) {
            console.warn('性能监控初始化失败:', error);
        }
    }

    /**
     * 记录性能指标
     * @param {string} type - 指标类型
     * @param {Object} data - 指标数据
     */
    recordMetric(type, data) {
        if (!this.metrics.has(type)) {
            this.metrics.set(type, []);
        }
        
        this.metrics.get(type).push({
            ...data,
            timestamp: Date.now()
        });
        
        // 限制存储数量
        const maxEntries = 100;
        const entries = this.metrics.get(type);
        if (entries.length > maxEntries) {
            entries.splice(0, entries.length - maxEntries);
        }
    }

    /**
     * 获取资源类型
     * @private
     */
    getResourceType(url) {
        if (url.includes('.js')) return 'script';
        if (url.includes('.css')) return 'stylesheet';
        if (url.match(/\.(png|jpg|jpeg|gif|svg|webp)$/)) return 'image';
        if (url.match(/\.(woff|woff2|ttf|eot)$/)) return 'font';
        return 'other';
    }

    /**
     * 开始性能计时
     * @param {string} name - 计时器名称
     */
    startTimer(name) {
        this.timers.set(name, performance.now());
    }

    /**
     * 结束性能计时
     * @param {string} name - 计时器名称
     * @returns {number} 耗时（毫秒）
     */
    endTimer(name) {
        const startTime = this.timers.get(name);
        if (!startTime) {
            console.warn(`计时器 ${name} 不存在`);
            return 0;
        }
        
        const duration = performance.now() - startTime;
        this.timers.delete(name);
        
        this.recordMetric('custom', {
            name,
            duration,
            type: 'timer'
        });
        
        return duration;
    }

    /**
     * 测量函数执行时间
     * @param {string} name - 测量名称
     * @param {Function} fn - 要测量的函数
     * @returns {*} 函数返回值
     */
    measure(name, fn) {
        this.startTimer(name);
        const result = fn();
        this.endTimer(name);
        return result;
    }

    /**
     * 测量异步函数执行时间
     * @param {string} name - 测量名称
     * @param {Function} fn - 要测量的异步函数
     * @returns {Promise<*>} 函数返回值
     */
    async measureAsync(name, fn) {
        this.startTimer(name);
        try {
            const result = await fn();
            this.endTimer(name);
            return result;
        } catch (error) {
            this.endTimer(name);
            throw error;
        }
    }

    /**
     * 防抖函数
     * @param {Function} func - 要防抖的函数
     * @param {number} wait - 等待时间（毫秒）
     * @param {boolean} immediate - 是否立即执行
     * @returns {Function} 防抖后的函数
     */
    static debounce(func, wait, immediate = false) {
        let timeout;
        
        return function executedFunction(...args) {
            const later = () => {
                timeout = null;
                if (!immediate) func.apply(this, args);
            };
            
            const callNow = immediate && !timeout;
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
            
            if (callNow) func.apply(this, args);
        };
    }

    /**
     * 节流函数
     * @param {Function} func - 要节流的函数
     * @param {number} limit - 限制时间（毫秒）
     * @returns {Function} 节流后的函数
     */
    static throttle(func, limit) {
        let inThrottle;
        
        return function executedFunction(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    /**
     * 懒加载图片
     * @param {string} selector - 图片选择器
     * @param {Object} options - 选项
     */
    static lazyLoadImages(selector = 'img[data-src]', options = {}) {
        const defaultOptions = {
            rootMargin: '50px',
            threshold: 0.1,
            ...options
        };

        if (!('IntersectionObserver' in window)) {
            // 降级处理
            document.querySelectorAll(selector).forEach(img => {
                if (img.dataset.src) {
                    img.src = img.dataset.src;
                    img.removeAttribute('data-src');
                }
            });
            return;
        }

        const imageObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    if (img.dataset.src) {
                        img.src = img.dataset.src;
                        img.removeAttribute('data-src');
                        img.classList.add('loaded');
                    }
                    imageObserver.unobserve(img);
                }
            });
        }, defaultOptions);

        document.querySelectorAll(selector).forEach(img => {
            imageObserver.observe(img);
        });
    }

    /**
     * 懒加载模块
     * @param {Function} importFn - 动态导入函数
     * @param {string} fallback - 降级处理
     * @returns {Promise} 模块Promise
     */
    static async lazyLoadModule(importFn, fallback = null) {
        try {
            return await importFn();
        } catch (error) {
            console.error('模块懒加载失败:', error);
            if (fallback) {
                return fallback;
            }
            throw error;
        }
    }

    /**
     * 预加载资源
     * @param {Array} urls - 资源URL数组
     * @param {string} type - 资源类型
     */
    static preloadResources(urls, type = 'script') {
        urls.forEach(url => {
            const link = document.createElement('link');
            link.rel = 'preload';
            link.href = url;
            link.as = type;
            document.head.appendChild(link);
        });
    }

    /**
     * 获取页面性能指标
     * @returns {Object} 性能指标
     */
    getPageMetrics() {
        if (!this.isSupported) {
            return { error: '浏览器不支持性能API' };
        }

        const navigation = performance.getEntriesByType('navigation')[0];
        const paint = performance.getEntriesByType('paint');
        
        const metrics = {
            // 页面加载时间
            pageLoadTime: navigation ? navigation.loadEventEnd - navigation.fetchStart : 0,
            
            // DOM解析时间
            domParseTime: navigation ? navigation.domContentLoadedEventEnd - navigation.domLoading : 0,
            
            // 资源加载时间
            resourceLoadTime: navigation ? navigation.loadEventEnd - navigation.domContentLoadedEventEnd : 0,
            
            // 首次绘制时间
            firstPaint: 0,
            firstContentfulPaint: 0,
            
            // 内存使用情况
            memory: performance.memory ? {
                used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
                total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
                limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
            } : null
        };

        // 处理绘制时间
        paint.forEach(entry => {
            if (entry.name === 'first-paint') {
                metrics.firstPaint = entry.startTime;
            } else if (entry.name === 'first-contentful-paint') {
                metrics.firstContentfulPaint = entry.startTime;
            }
        });

        return metrics;
    }

    /**
     * 获取资源性能统计
     * @returns {Object} 资源统计
     */
    getResourceStats() {
        const resources = this.metrics.get('resource') || [];
        const stats = {
            total: resources.length,
            totalSize: 0,
            avgDuration: 0,
            types: {}
        };

        resources.forEach(resource => {
            stats.totalSize += resource.size || 0;
            
            if (!stats.types[resource.type]) {
                stats.types[resource.type] = { count: 0, size: 0, duration: 0 };
            }
            
            stats.types[resource.type].count++;
            stats.types[resource.type].size += resource.size || 0;
            stats.types[resource.type].duration += resource.duration || 0;
        });

        if (resources.length > 0) {
            stats.avgDuration = resources.reduce((sum, r) => sum + (r.duration || 0), 0) / resources.length;
        }

        return stats;
    }

    /**
     * 生成性能报告
     * @returns {Object} 性能报告
     */
    generateReport() {
        return {
            timestamp: new Date().toISOString(),
            page: this.getPageMetrics(),
            resources: this.getResourceStats(),
            custom: this.metrics.get('custom') || [],
            browser: {
                userAgent: navigator.userAgent,
                language: navigator.language,
                platform: navigator.platform,
                cookieEnabled: navigator.cookieEnabled,
                onLine: navigator.onLine
            }
        };
    }

    /**
     * 清理性能数据
     */
    cleanup() {
        this.metrics.clear();
        this.timers.clear();
        
        this.observers.forEach(observer => {
            observer.disconnect();
        });
        this.observers.clear();
    }

    /**
     * 监控长任务
     * @param {Function} callback - 回调函数
     */
    monitorLongTasks(callback) {
        if (!('PerformanceLongTaskTiming' in window)) {
            console.warn('浏览器不支持长任务监控');
            return;
        }

        const observer = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            entries.forEach(entry => {
                callback({
                    duration: entry.duration,
                    startTime: entry.startTime,
                    name: entry.name,
                    attribution: entry.attribution
                });
            });
        });

        observer.observe({ entryTypes: ['longtask'] });
        this.observers.set('longtask', observer);
    }

    /**
     * 检测内存泄漏
     * @returns {Object} 内存信息
     */
    checkMemoryLeak() {
        if (!performance.memory) {
            return { error: '浏览器不支持内存监控' };
        }

        const memory = performance.memory;
        const usage = memory.usedJSHeapSize / memory.jsHeapSizeLimit;
        
        return {
            used: Math.round(memory.usedJSHeapSize / 1024 / 1024),
            total: Math.round(memory.totalJSHeapSize / 1024 / 1024),
            limit: Math.round(memory.jsHeapSizeLimit / 1024 / 1024),
            usage: Math.round(usage * 100),
            warning: usage > 0.8 ? '内存使用率过高，可能存在内存泄漏' : null
        };
    }
}

// 创建全局性能监控实例
export const performanceMonitor = new Performance();

// 导出工具函数
export const { debounce, throttle, lazyLoadImages, lazyLoadModule, preloadResources } = Performance;

// 导出类以便扩展
export { Performance };
