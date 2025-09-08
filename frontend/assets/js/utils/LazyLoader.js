/**
 * 懒加载管理器
 * 提供模块懒加载、组件懒加载、资源懒加载等功能
 */

class LazyLoader {
    constructor() {
        this.loadedModules = new Map();
        this.loadingPromises = new Map();
        this.componentCache = new Map();
        this.observers = new Map();
        
        this.initIntersectionObserver();
    }

    /**
     * 初始化交叉观察器
     * @private
     */
    initIntersectionObserver() {
        if (!('IntersectionObserver' in window)) {
            console.warn('浏览器不支持 IntersectionObserver');
            return;
        }

        // 创建默认观察器
        this.defaultObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const element = entry.target;
                    const lazyType = element.dataset.lazyType;
                    const lazyTarget = element.dataset.lazyTarget;
                    
                    if (lazyType && lazyTarget) {
                        this.handleLazyLoad(element, lazyType, lazyTarget);
                    }
                }
            });
        }, {
            rootMargin: '50px',
            threshold: 0.1
        });
    }

    /**
     * 处理懒加载
     * @private
     */
    async handleLazyLoad(element, type, target) {
        try {
            switch (type) {
                case 'component':
                    await this.loadComponent(target, element);
                    break;
                case 'module':
                    await this.loadModule(target);
                    break;
                case 'image':
                    this.loadImage(element, target);
                    break;
                case 'iframe':
                    this.loadIframe(element, target);
                    break;
                default:
                    console.warn(`未知的懒加载类型: ${type}`);
            }
            
            // 停止观察
            this.defaultObserver.unobserve(element);
        } catch (error) {
            console.error(`懒加载失败 (${type}:${target}):`, error);
        }
    }

    /**
     * 懒加载模块
     * @param {string} modulePath - 模块路径
     * @returns {Promise<Object>} 模块对象
     */
    async loadModule(modulePath) {
        // 检查缓存
        if (this.loadedModules.has(modulePath)) {
            return this.loadedModules.get(modulePath);
        }

        // 检查是否正在加载
        if (this.loadingPromises.has(modulePath)) {
            return this.loadingPromises.get(modulePath);
        }

        // 开始加载
        const loadPromise = this.doLoadModule(modulePath);
        this.loadingPromises.set(modulePath, loadPromise);

        try {
            const module = await loadPromise;
            this.loadedModules.set(modulePath, module);
            this.loadingPromises.delete(modulePath);
            return module;
        } catch (error) {
            this.loadingPromises.delete(modulePath);
            throw error;
        }
    }

    /**
     * 执行模块加载
     * @private
     */
    async doLoadModule(modulePath) {
        // 支持多种模块路径格式
        const normalizedPath = this.normalizePath(modulePath);
        
        try {
            const module = await import(normalizedPath);
            console.log(`模块加载成功: ${modulePath}`);
            return module;
        } catch (error) {
            console.error(`模块加载失败: ${modulePath}`, error);
            throw new Error(`Failed to load module: ${modulePath}`);
        }
    }

    /**
     * 标准化路径
     * @private
     */
    normalizePath(path) {
        // 如果是相对路径，添加基础路径
        if (path.startsWith('./') || path.startsWith('../')) {
            return path;
        }
        
        // 如果是别名路径，转换为实际路径
        const aliases = {
            '@': './assets',
            '@js': './assets/js',
            '@components': './assets/js/components',
            '@utils': './assets/js/utils',
            '@api': './assets/js/api',
            '@modules': './assets/js/modules'
        };
        
        for (const [alias, realPath] of Object.entries(aliases)) {
            if (path.startsWith(alias + '/')) {
                return path.replace(alias, realPath);
            }
        }
        
        // 默认添加 .js 扩展名
        if (!path.includes('.')) {
            return path + '.js';
        }
        
        return path;
    }

    /**
     * 懒加载组件
     * @param {string} componentName - 组件名称
     * @param {Element} container - 容器元素
     * @param {Object} options - 组件选项
     * @returns {Promise<Component>} 组件实例
     */
    async loadComponent(componentName, container, options = {}) {
        // 检查组件缓存
        const cacheKey = `${componentName}:${container.id || 'anonymous'}`;
        if (this.componentCache.has(cacheKey)) {
            return this.componentCache.get(cacheKey);
        }

        try {
            // 动态加载组件模块
            const componentPath = `@components/${componentName}.js`;
            const module = await this.loadModule(componentPath);
            
            // 获取组件类
            const ComponentClass = module[componentName] || module.default;
            if (!ComponentClass) {
                throw new Error(`组件 ${componentName} 未找到`);
            }

            // 创建组件实例
            const component = new ComponentClass(container, options);
            
            // 缓存组件实例
            this.componentCache.set(cacheKey, component);
            
            console.log(`组件加载成功: ${componentName}`);
            return component;
        } catch (error) {
            console.error(`组件加载失败: ${componentName}`, error);
            throw error;
        }
    }

    /**
     * 懒加载图片
     * @param {Element} img - 图片元素
     * @param {string} src - 图片源
     */
    loadImage(img, src) {
        const actualSrc = src || img.dataset.src;
        if (!actualSrc) return;

        // 创建新图片对象预加载
        const newImg = new Image();
        
        newImg.onload = () => {
            img.src = actualSrc;
            img.classList.add('loaded');
            img.removeAttribute('data-src');
            
            // 添加淡入动画
            img.style.opacity = '0';
            img.style.transition = 'opacity 0.3s ease';
            setTimeout(() => {
                img.style.opacity = '1';
            }, 10);
        };
        
        newImg.onerror = () => {
            img.classList.add('error');
            console.error(`图片加载失败: ${actualSrc}`);
        };
        
        newImg.src = actualSrc;
    }

    /**
     * 懒加载iframe
     * @param {Element} iframe - iframe元素
     * @param {string} src - iframe源
     */
    loadIframe(iframe, src) {
        const actualSrc = src || iframe.dataset.src;
        if (!actualSrc) return;

        iframe.src = actualSrc;
        iframe.classList.add('loaded');
        iframe.removeAttribute('data-src');
    }

    /**
     * 注册懒加载元素
     * @param {Element|string} element - 元素或选择器
     * @param {Object} options - 选项
     */
    observe(element, options = {}) {
        const elements = typeof element === 'string' 
            ? document.querySelectorAll(element)
            : [element];

        elements.forEach(el => {
            if (el && !el.dataset.lazyObserved) {
                el.dataset.lazyObserved = 'true';
                this.defaultObserver.observe(el);
            }
        });
    }

    /**
     * 取消观察元素
     * @param {Element} element - 元素
     */
    unobserve(element) {
        if (element && element.dataset.lazyObserved) {
            delete element.dataset.lazyObserved;
            this.defaultObserver.unobserve(element);
        }
    }

    /**
     * 预加载模块
     * @param {Array<string>} modules - 模块路径数组
     * @returns {Promise<Array>} 加载结果
     */
    async preloadModules(modules) {
        const promises = modules.map(module => this.loadModule(module));
        return Promise.allSettled(promises);
    }

    /**
     * 预加载组件
     * @param {Array<string>} components - 组件名称数组
     * @returns {Promise<Array>} 加载结果
     */
    async preloadComponents(components) {
        const promises = components.map(async (componentName) => {
            try {
                const componentPath = `@components/${componentName}.js`;
                return await this.loadModule(componentPath);
            } catch (error) {
                console.warn(`预加载组件失败: ${componentName}`, error);
                return null;
            }
        });
        
        return Promise.allSettled(promises);
    }

    /**
     * 批量注册懒加载图片
     * @param {string} selector - 选择器
     * @param {Object} options - 选项
     */
    registerImages(selector = 'img[data-src]', options = {}) {
        const images = document.querySelectorAll(selector);
        
        images.forEach(img => {
            img.dataset.lazyType = 'image';
            if (!img.dataset.lazyTarget && img.dataset.src) {
                img.dataset.lazyTarget = img.dataset.src;
            }
        });
        
        this.observe(images, options);
    }

    /**
     * 批量注册懒加载组件
     * @param {string} selector - 选择器
     * @param {Object} options - 选项
     */
    registerComponents(selector = '[data-lazy-component]', options = {}) {
        const elements = document.querySelectorAll(selector);
        
        elements.forEach(el => {
            const componentName = el.dataset.lazyComponent;
            if (componentName) {
                el.dataset.lazyType = 'component';
                el.dataset.lazyTarget = componentName;
            }
        });
        
        this.observe(elements, options);
    }

    /**
     * 路由级别的代码分割
     * @param {Object} routes - 路由配置
     * @returns {Object} 懒加载路由
     */
    createLazyRoutes(routes) {
        const lazyRoutes = {};
        
        Object.entries(routes).forEach(([path, config]) => {
            if (typeof config === 'string') {
                // 简单的模块路径
                lazyRoutes[path] = () => this.loadModule(config);
            } else if (config.component) {
                // 组件配置
                lazyRoutes[path] = async () => {
                    const module = await this.loadModule(config.component);
                    return {
                        ...config,
                        component: module.default || module[config.name]
                    };
                };
            } else {
                lazyRoutes[path] = config;
            }
        });
        
        return lazyRoutes;
    }

    /**
     * 获取加载统计
     * @returns {Object} 统计信息
     */
    getStats() {
        return {
            loadedModules: this.loadedModules.size,
            loadingModules: this.loadingPromises.size,
            cachedComponents: this.componentCache.size,
            moduleList: Array.from(this.loadedModules.keys()),
            componentList: Array.from(this.componentCache.keys())
        };
    }

    /**
     * 清理缓存
     * @param {string} type - 清理类型: 'modules', 'components', 'all'
     */
    clearCache(type = 'all') {
        switch (type) {
            case 'modules':
                this.loadedModules.clear();
                break;
            case 'components':
                this.componentCache.clear();
                break;
            case 'all':
                this.loadedModules.clear();
                this.componentCache.clear();
                this.loadingPromises.clear();
                break;
        }
    }

    /**
     * 销毁懒加载器
     */
    destroy() {
        // 断开所有观察器
        this.observers.forEach(observer => observer.disconnect());
        this.observers.clear();
        
        if (this.defaultObserver) {
            this.defaultObserver.disconnect();
        }
        
        // 清理缓存
        this.clearCache();
    }
}

// 创建全局懒加载器实例
export const lazyLoader = new LazyLoader();

// 导出类以便扩展
export { LazyLoader };

// 便捷函数
export const loadModule = (path) => lazyLoader.loadModule(path);
export const loadComponent = (name, container, options) => lazyLoader.loadComponent(name, container, options);
export const observeLazy = (element, options) => lazyLoader.observe(element, options);
