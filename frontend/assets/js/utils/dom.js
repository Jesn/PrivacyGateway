/**
 * DOM操作工具模块
 * 提供常用的DOM操作方法，简化DOM操作代码
 */

class DOMUtils {
    /**
     * 获取单个元素
     * @param {string} selector - CSS选择器
     * @param {Element} context - 查找上下文，默认为document
     * @returns {Element|null} 找到的元素
     */
    static get(selector, context = document) {
        return context.querySelector(selector);
    }

    /**
     * 获取多个元素
     * @param {string} selector - CSS选择器
     * @param {Element} context - 查找上下文，默认为document
     * @returns {NodeList} 找到的元素列表
     */
    static getAll(selector, context = document) {
        return context.querySelectorAll(selector);
    }

    /**
     * 根据ID获取元素
     * @param {string} id - 元素ID
     * @returns {Element|null} 找到的元素
     */
    static getById(id) {
        return document.getElementById(id);
    }

    /**
     * 创建元素
     * @param {string} tag - 标签名
     * @param {Object} attributes - 属性对象
     * @param {string} content - 内容
     * @returns {Element} 创建的元素
     */
    static create(tag, attributes = {}, content = '') {
        const element = document.createElement(tag);
        
        // 设置属性
        Object.entries(attributes).forEach(([key, value]) => {
            if (key === 'className' || key === 'class') {
                element.className = value;
            } else if (key === 'innerHTML') {
                element.innerHTML = value;
            } else if (key === 'textContent') {
                element.textContent = value;
            } else if (key.startsWith('data-')) {
                element.setAttribute(key, value);
            } else {
                element[key] = value;
            }
        });
        
        // 设置内容
        if (content) {
            element.innerHTML = content;
        }
        
        return element;
    }

    /**
     * 添加CSS类
     * @param {Element|string} element - 元素或选择器
     * @param {string|Array} className - 类名或类名数组
     */
    static addClass(element, className) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        const classes = Array.isArray(className) ? className : [className];
        el.classList.add(...classes);
    }

    /**
     * 移除CSS类
     * @param {Element|string} element - 元素或选择器
     * @param {string|Array} className - 类名或类名数组
     */
    static removeClass(element, className) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        const classes = Array.isArray(className) ? className : [className];
        el.classList.remove(...classes);
    }

    /**
     * 切换CSS类
     * @param {Element|string} element - 元素或选择器
     * @param {string} className - 类名
     * @returns {boolean} 切换后是否包含该类
     */
    static toggleClass(element, className) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return false;
        
        return el.classList.toggle(className);
    }

    /**
     * 检查是否包含CSS类
     * @param {Element|string} element - 元素或选择器
     * @param {string} className - 类名
     * @returns {boolean} 是否包含该类
     */
    static hasClass(element, className) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return false;
        
        return el.classList.contains(className);
    }

    /**
     * 显示元素
     * @param {Element|string} element - 元素或选择器
     * @param {string} display - 显示方式，默认为'block'
     */
    static show(element, display = 'block') {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.style.display = display;
    }

    /**
     * 隐藏元素
     * @param {Element|string} element - 元素或选择器
     */
    static hide(element) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.style.display = 'none';
    }

    /**
     * 切换元素显示/隐藏
     * @param {Element|string} element - 元素或选择器
     * @param {string} display - 显示方式，默认为'block'
     */
    static toggle(element, display = 'block') {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        if (el.style.display === 'none') {
            this.show(el, display);
        } else {
            this.hide(el);
        }
    }

    /**
     * 设置元素文本内容
     * @param {Element|string} element - 元素或选择器
     * @param {string} text - 文本内容
     */
    static setText(element, text) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.textContent = text;
    }

    /**
     * 设置元素HTML内容
     * @param {Element|string} element - 元素或选择器
     * @param {string} html - HTML内容
     */
    static setHTML(element, html) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.innerHTML = html;
    }

    /**
     * 获取元素文本内容
     * @param {Element|string} element - 元素或选择器
     * @returns {string} 文本内容
     */
    static getText(element) {
        const el = typeof element === 'string' ? this.get(element) : element;
        return el ? el.textContent : '';
    }

    /**
     * 获取元素HTML内容
     * @param {Element|string} element - 元素或选择器
     * @returns {string} HTML内容
     */
    static getHTML(element) {
        const el = typeof element === 'string' ? this.get(element) : element;
        return el ? el.innerHTML : '';
    }

    /**
     * 设置元素属性
     * @param {Element|string} element - 元素或选择器
     * @param {string|Object} attr - 属性名或属性对象
     * @param {*} value - 属性值
     */
    static setAttr(element, attr, value) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        if (typeof attr === 'object') {
            Object.entries(attr).forEach(([key, val]) => {
                el.setAttribute(key, val);
            });
        } else {
            el.setAttribute(attr, value);
        }
    }

    /**
     * 获取元素属性
     * @param {Element|string} element - 元素或选择器
     * @param {string} attr - 属性名
     * @returns {string|null} 属性值
     */
    static getAttr(element, attr) {
        const el = typeof element === 'string' ? this.get(element) : element;
        return el ? el.getAttribute(attr) : null;
    }

    /**
     * 移除元素属性
     * @param {Element|string} element - 元素或选择器
     * @param {string} attr - 属性名
     */
    static removeAttr(element, attr) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.removeAttribute(attr);
    }

    /**
     * 添加事件监听器
     * @param {Element|string} element - 元素或选择器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     * @param {Object} options - 事件选项
     */
    static on(element, event, handler, options = {}) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.addEventListener(event, handler, options);
    }

    /**
     * 移除事件监听器
     * @param {Element|string} element - 元素或选择器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     */
    static off(element, event, handler) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.removeEventListener(event, handler);
    }

    /**
     * 委托事件监听
     * @param {Element|string} container - 容器元素或选择器
     * @param {string} selector - 目标元素选择器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     */
    static delegate(container, selector, event, handler) {
        const containerEl = typeof container === 'string' ? this.get(container) : container;
        if (!containerEl) return;
        
        containerEl.addEventListener(event, (e) => {
            const target = e.target.closest(selector);
            if (target && containerEl.contains(target)) {
                handler.call(target, e);
            }
        });
    }

    /**
     * 等待DOM加载完成
     * @param {Function} callback - 回调函数
     */
    static ready(callback) {
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', callback);
        } else {
            callback();
        }
    }

    /**
     * 平滑滚动到元素
     * @param {Element|string} element - 元素或选择器
     * @param {Object} options - 滚动选项
     */
    static scrollTo(element, options = {}) {
        const el = typeof element === 'string' ? this.get(element) : element;
        if (!el) return;
        
        el.scrollIntoView({
            behavior: 'smooth',
            block: 'start',
            inline: 'nearest',
            ...options
        });
    }

    /**
     * 复制文本到剪贴板
     * @param {string} text - 要复制的文本
     * @returns {Promise<boolean>} 是否成功
     */
    static async copyToClipboard(text) {
        try {
            if (navigator.clipboard && navigator.clipboard.writeText) {
                await navigator.clipboard.writeText(text);
                return true;
            } else {
                // 降级方案
                const textArea = this.create('textarea', {
                    value: text,
                    style: 'position: absolute; left: -9999px; top: -9999px;'
                });
                
                document.body.appendChild(textArea);
                textArea.select();
                const success = document.execCommand('copy');
                document.body.removeChild(textArea);
                
                return success;
            }
        } catch (error) {
            console.warn('复制到剪贴板失败:', error);
            return false;
        }
    }
}

// 简化的全局函数
export const $ = DOMUtils.get.bind(DOMUtils);
export const $$ = DOMUtils.getAll.bind(DOMUtils);

// 导出工具类
export { DOMUtils };
