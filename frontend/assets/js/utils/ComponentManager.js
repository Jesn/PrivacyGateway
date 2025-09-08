/**
 * 组件管理器
 * 统一管理组件的创建、销毁、查找等操作
 */

import { eventBus } from './EventBus.js';
import { stateManager } from './StateManager.js';

class ComponentManager {
    constructor() {
        this.components = new Map();
        this.componentTypes = new Map();
        this.instances = new Map();
        this.autoId = 0;
        
        // 监听组件生命周期事件
        this.bindLifecycleEvents();
    }

    /**
     * 注册组件类型
     * @param {string} name - 组件名称
     * @param {Class} ComponentClass - 组件类
     */
    register(name, ComponentClass) {
        this.componentTypes.set(name, ComponentClass);
        
        // 为组件类添加静态方法
        ComponentClass.create = (container, options) => {
            return this.create(name, container, options);
        };
        
        ComponentClass.find = (selector) => {
            return this.find(name, selector);
        };
        
        ComponentClass.findAll = (selector) => {
            return this.findAll(name, selector);
        };
    }

    /**
     * 创建组件实例
     * @param {string} type - 组件类型
     * @param {Element|string} container - 容器元素或选择器
     * @param {Object} options - 组件选项
     * @returns {Component} 组件实例
     */
    create(type, container, options = {}) {
        const ComponentClass = this.componentTypes.get(type);
        if (!ComponentClass) {
            throw new Error(`未注册的组件类型: ${type}`);
        }

        // 解析容器
        const containerElement = typeof container === 'string' 
            ? document.querySelector(container) 
            : container;
            
        if (!containerElement) {
            throw new Error('组件容器不存在');
        }

        // 创建组件实例
        const instance = new ComponentClass(containerElement, options);
        
        // 生成唯一ID
        const id = options.id || this.generateId(type);
        instance.managerId = id;
        instance.type = type;
        
        // 注册实例
        this.instances.set(id, instance);
        
        // 按类型分组
        if (!this.components.has(type)) {
            this.components.set(type, new Set());
        }
        this.components.get(type).add(instance);
        
        // 触发创建事件
        eventBus.emit('component:created', {
            type,
            id,
            instance,
            container: containerElement
        });
        
        return instance;
    }

    /**
     * 销毁组件
     * @param {string|Component} target - 组件ID或组件实例
     */
    destroy(target) {
        let instance;
        let id;
        
        if (typeof target === 'string') {
            id = target;
            instance = this.instances.get(id);
        } else {
            instance = target;
            id = instance.managerId;
        }
        
        if (!instance) return;
        
        // 从注册表中移除
        this.instances.delete(id);
        
        const typeInstances = this.components.get(instance.type);
        if (typeInstances) {
            typeInstances.delete(instance);
            if (typeInstances.size === 0) {
                this.components.delete(instance.type);
            }
        }
        
        // 销毁组件
        instance.destroy();
        
        // 触发销毁事件
        eventBus.emit('component:destroyed', {
            type: instance.type,
            id,
            instance
        });
    }

    /**
     * 查找组件实例
     * @param {string} type - 组件类型
     * @param {string} selector - 选择器或ID
     * @returns {Component|null} 组件实例
     */
    find(type, selector) {
        if (selector.startsWith('#')) {
            // 按ID查找
            const id = selector.substring(1);
            const instance = this.instances.get(id);
            return instance && instance.type === type ? instance : null;
        } else {
            // 按容器选择器查找
            const element = document.querySelector(selector);
            if (!element) return null;
            
            const typeInstances = this.components.get(type);
            if (!typeInstances) return null;
            
            for (const instance of typeInstances) {
                if (instance.container === element) {
                    return instance;
                }
            }
        }
        
        return null;
    }

    /**
     * 查找所有匹配的组件实例
     * @param {string} type - 组件类型
     * @param {string} selector - 选择器
     * @returns {Array<Component>} 组件实例数组
     */
    findAll(type, selector) {
        const typeInstances = this.components.get(type);
        if (!typeInstances) return [];
        
        if (!selector) {
            return Array.from(typeInstances);
        }
        
        const elements = document.querySelectorAll(selector);
        const result = [];
        
        for (const instance of typeInstances) {
            for (const element of elements) {
                if (instance.container === element) {
                    result.push(instance);
                    break;
                }
            }
        }
        
        return result;
    }

    /**
     * 获取组件实例
     * @param {string} id - 组件ID
     * @returns {Component|null} 组件实例
     */
    getInstance(id) {
        return this.instances.get(id) || null;
    }

    /**
     * 获取所有组件实例
     * @param {string} type - 组件类型，可选
     * @returns {Array<Component>} 组件实例数组
     */
    getAllInstances(type) {
        if (type) {
            const typeInstances = this.components.get(type);
            return typeInstances ? Array.from(typeInstances) : [];
        }
        
        return Array.from(this.instances.values());
    }

    /**
     * 销毁所有组件
     * @param {string} type - 组件类型，可选
     */
    destroyAll(type) {
        if (type) {
            const typeInstances = this.components.get(type);
            if (typeInstances) {
                Array.from(typeInstances).forEach(instance => this.destroy(instance));
            }
        } else {
            Array.from(this.instances.values()).forEach(instance => this.destroy(instance));
        }
    }

    /**
     * 生成唯一ID
     * @private
     */
    generateId(type) {
        return `${type}-${++this.autoId}-${Date.now()}`;
    }

    /**
     * 绑定生命周期事件
     * @private
     */
    bindLifecycleEvents() {
        // 监听组件挂载事件
        eventBus.on('component:mounted', (data) => {
            const { component } = data.detail;
            if (component.managerId) {
                stateManager.setState(`components.${component.managerId}.mounted`, true);
            }
        });
        
        // 监听组件销毁事件
        eventBus.on('component:destroyed', (data) => {
            const { id } = data.detail;
            stateManager.setState(`components.${id}`, undefined);
        });
        
        // 监听页面卸载，清理所有组件
        window.addEventListener('beforeunload', () => {
            this.destroyAll();
        });
    }

    /**
     * 自动初始化页面中的组件
     * @param {Element} container - 容器元素，默认为document
     */
    autoInit(container = document) {
        const elements = container.querySelectorAll('[data-component]');
        
        elements.forEach(element => {
            const type = element.dataset.component;
            const optionsStr = element.dataset.componentOptions;
            
            let options = {};
            if (optionsStr) {
                try {
                    options = JSON.parse(optionsStr);
                } catch (error) {
                    console.warn('组件选项解析失败:', error);
                }
            }
            
            // 避免重复初始化
            if (!element.dataset.componentId) {
                try {
                    const instance = this.create(type, element, options);
                    element.dataset.componentId = instance.managerId;
                } catch (error) {
                    console.error(`自动初始化组件失败 (${type}):`, error);
                }
            }
        });
    }

    /**
     * 创建组件工厂函数
     * @param {string} type - 组件类型
     * @returns {Function} 工厂函数
     */
    factory(type) {
        return (container, options) => this.create(type, container, options);
    }

    /**
     * 批量注册组件
     * @param {Object} components - 组件映射对象
     */
    registerBatch(components) {
        Object.entries(components).forEach(([name, ComponentClass]) => {
            this.register(name, ComponentClass);
        });
    }

    /**
     * 获取组件统计信息
     * @returns {Object} 统计信息
     */
    getStats() {
        const stats = {
            totalInstances: this.instances.size,
            typeCount: this.components.size,
            types: {}
        };
        
        this.components.forEach((instances, type) => {
            stats.types[type] = instances.size;
        });
        
        return stats;
    }

    /**
     * 调试信息
     */
    debug() {
        console.group('[ComponentManager] Debug Info');
        console.log('Registered Types:', Array.from(this.componentTypes.keys()));
        console.log('Total Instances:', this.instances.size);
        console.log('Stats:', this.getStats());
        
        this.components.forEach((instances, type) => {
            console.log(`${type}:`, Array.from(instances).map(i => i.managerId));
        });
        
        console.groupEnd();
    }

    /**
     * 清理无效的组件实例
     */
    cleanup() {
        const toRemove = [];
        
        this.instances.forEach((instance, id) => {
            // 检查容器是否还在DOM中
            if (!document.contains(instance.container)) {
                toRemove.push(id);
            }
        });
        
        toRemove.forEach(id => {
            console.log(`清理无效组件: ${id}`);
            this.destroy(id);
        });
        
        return toRemove.length;
    }

    /**
     * 监听DOM变化，自动清理组件
     */
    enableAutoCleanup() {
        if (this.observer) return;
        
        this.observer = new MutationObserver((mutations) => {
            let needsCleanup = false;
            
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList' && mutation.removedNodes.length > 0) {
                    needsCleanup = true;
                }
            });
            
            if (needsCleanup) {
                // 延迟清理，避免频繁操作
                clearTimeout(this.cleanupTimer);
                this.cleanupTimer = setTimeout(() => {
                    this.cleanup();
                }, 1000);
            }
        });
        
        this.observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    }

    /**
     * 禁用自动清理
     */
    disableAutoCleanup() {
        if (this.observer) {
            this.observer.disconnect();
            this.observer = null;
        }
        
        if (this.cleanupTimer) {
            clearTimeout(this.cleanupTimer);
            this.cleanupTimer = null;
        }
    }
}

// 创建全局组件管理器实例
export const componentManager = new ComponentManager();

// 导出类以便扩展
export { ComponentManager };
