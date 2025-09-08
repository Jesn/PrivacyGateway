/**
 * 事件总线系统
 * 提供组件间的事件通信和解耦
 */

class EventBus {
    constructor() {
        this.events = new Map();
        this.onceEvents = new Map();
        this.middleware = [];
        this.maxListeners = 100;
        this.debug = false;
    }

    /**
     * 监听事件
     * @param {string} eventName - 事件名
     * @param {Function} listener - 监听器函数
     * @param {Object} options - 选项
     * @returns {Function} 取消监听函数
     */
    on(eventName, listener, options = {}) {
        const { priority = 0, context = null } = options;
        
        if (typeof listener !== 'function') {
            throw new Error('监听器必须是函数');
        }

        if (!this.events.has(eventName)) {
            this.events.set(eventName, []);
        }

        const listeners = this.events.get(eventName);
        
        // 检查监听器数量限制
        if (listeners.length >= this.maxListeners) {
            console.warn(`事件 ${eventName} 的监听器数量已达到上限 ${this.maxListeners}`);
        }

        const listenerObj = {
            listener,
            priority,
            context,
            id: this.generateId()
        };

        // 按优先级插入
        const insertIndex = listeners.findIndex(l => l.priority < priority);
        if (insertIndex === -1) {
            listeners.push(listenerObj);
        } else {
            listeners.splice(insertIndex, 0, listenerObj);
        }

        if (this.debug) {
            console.log(`[EventBus] 添加监听器: ${eventName}`, listenerObj);
        }

        // 返回取消监听函数
        return () => this.off(eventName, listener);
    }

    /**
     * 监听一次性事件
     * @param {string} eventName - 事件名
     * @param {Function} listener - 监听器函数
     * @param {Object} options - 选项
     * @returns {Function} 取消监听函数
     */
    once(eventName, listener, options = {}) {
        const onceListener = (...args) => {
            this.off(eventName, onceListener);
            listener.apply(options.context || null, args);
        };

        return this.on(eventName, onceListener, options);
    }

    /**
     * 取消监听事件
     * @param {string} eventName - 事件名
     * @param {Function} listener - 监听器函数
     */
    off(eventName, listener) {
        if (!this.events.has(eventName)) return;

        const listeners = this.events.get(eventName);
        const index = listeners.findIndex(l => l.listener === listener);
        
        if (index > -1) {
            const removed = listeners.splice(index, 1)[0];
            
            if (this.debug) {
                console.log(`[EventBus] 移除监听器: ${eventName}`, removed);
            }
            
            // 如果没有监听器了，删除事件
            if (listeners.length === 0) {
                this.events.delete(eventName);
            }
        }
    }

    /**
     * 取消所有监听器
     * @param {string} eventName - 事件名，不提供则清除所有
     */
    offAll(eventName) {
        if (eventName) {
            this.events.delete(eventName);
            if (this.debug) {
                console.log(`[EventBus] 清除所有监听器: ${eventName}`);
            }
        } else {
            this.events.clear();
            if (this.debug) {
                console.log('[EventBus] 清除所有事件监听器');
            }
        }
    }

    /**
     * 触发事件
     * @param {string} eventName - 事件名
     * @param {...*} args - 事件参数
     * @returns {boolean} 是否有监听器处理了事件
     */
    emit(eventName, ...args) {
        // 应用中间件
        const eventData = {
            name: eventName,
            args,
            timestamp: Date.now(),
            preventDefault: false,
            stopPropagation: false
        };

        const processedEvent = this.applyMiddleware(eventData);
        if (!processedEvent || processedEvent.preventDefault) {
            return false;
        }

        const listeners = this.events.get(eventName);
        if (!listeners || listeners.length === 0) {
            if (this.debug) {
                console.log(`[EventBus] 没有监听器处理事件: ${eventName}`);
            }
            return false;
        }

        if (this.debug) {
            console.log(`[EventBus] 触发事件: ${eventName}`, processedEvent.args);
        }

        let handled = false;
        
        // 按优先级顺序执行监听器
        for (const listenerObj of listeners) {
            try {
                const result = listenerObj.listener.apply(
                    listenerObj.context, 
                    processedEvent.args
                );
                
                handled = true;
                
                // 如果监听器返回 false，停止传播
                if (result === false || processedEvent.stopPropagation) {
                    break;
                }
            } catch (error) {
                console.error(`[EventBus] 监听器执行错误 (${eventName}):`, error);
            }
        }

        return handled;
    }

    /**
     * 异步触发事件
     * @param {string} eventName - 事件名
     * @param {...*} args - 事件参数
     * @returns {Promise<boolean>} 是否有监听器处理了事件
     */
    async emitAsync(eventName, ...args) {
        const listeners = this.events.get(eventName);
        if (!listeners || listeners.length === 0) {
            return false;
        }

        if (this.debug) {
            console.log(`[EventBus] 异步触发事件: ${eventName}`, args);
        }

        let handled = false;
        
        for (const listenerObj of listeners) {
            try {
                const result = await listenerObj.listener.apply(
                    listenerObj.context, 
                    args
                );
                
                handled = true;
                
                if (result === false) {
                    break;
                }
            } catch (error) {
                console.error(`[EventBus] 异步监听器执行错误 (${eventName}):`, error);
            }
        }

        return handled;
    }

    /**
     * 添加中间件
     * @param {Function} middleware - 中间件函数
     */
    use(middleware) {
        if (typeof middleware !== 'function') {
            throw new Error('中间件必须是函数');
        }
        
        this.middleware.push(middleware);
    }

    /**
     * 应用中间件
     * @private
     */
    applyMiddleware(eventData) {
        return this.middleware.reduce((data, middleware) => {
            if (!data) return data;
            
            try {
                return middleware(data) || data;
            } catch (error) {
                console.error('[EventBus] 中间件执行错误:', error);
                return data;
            }
        }, eventData);
    }

    /**
     * 获取事件监听器数量
     * @param {string} eventName - 事件名
     * @returns {number} 监听器数量
     */
    listenerCount(eventName) {
        const listeners = this.events.get(eventName);
        return listeners ? listeners.length : 0;
    }

    /**
     * 获取所有事件名
     * @returns {Array<string>} 事件名数组
     */
    eventNames() {
        return Array.from(this.events.keys());
    }

    /**
     * 设置最大监听器数量
     * @param {number} max - 最大数量
     */
    setMaxListeners(max) {
        this.maxListeners = max;
    }

    /**
     * 启用/禁用调试模式
     * @param {boolean} enabled - 是否启用
     */
    setDebug(enabled) {
        this.debug = enabled;
    }

    /**
     * 生成唯一ID
     * @private
     */
    generateId() {
        return `listener-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    /**
     * 创建命名空间
     * @param {string} namespace - 命名空间
     * @returns {Object} 命名空间对象
     */
    namespace(namespace) {
        const prefixEvent = (eventName) => `${namespace}:${eventName}`;
        
        return {
            on: (eventName, listener, options) => 
                this.on(prefixEvent(eventName), listener, options),
            
            once: (eventName, listener, options) => 
                this.once(prefixEvent(eventName), listener, options),
            
            off: (eventName, listener) => 
                this.off(prefixEvent(eventName), listener),
            
            emit: (eventName, ...args) => 
                this.emit(prefixEvent(eventName), ...args),
            
            emitAsync: (eventName, ...args) => 
                this.emitAsync(prefixEvent(eventName), ...args)
        };
    }

    /**
     * 等待事件
     * @param {string} eventName - 事件名
     * @param {number} timeout - 超时时间（毫秒）
     * @returns {Promise} Promise对象
     */
    waitFor(eventName, timeout = 5000) {
        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.off(eventName, listener);
                reject(new Error(`等待事件 ${eventName} 超时`));
            }, timeout);

            const listener = (...args) => {
                clearTimeout(timer);
                resolve(args);
            };

            this.once(eventName, listener);
        });
    }

    /**
     * 调试信息
     */
    debug() {
        console.group('[EventBus] Debug Info');
        console.log('Events:', this.events);
        console.log('Middleware:', this.middleware);
        console.log('Max Listeners:', this.maxListeners);
        
        this.events.forEach((listeners, eventName) => {
            console.log(`${eventName}: ${listeners.length} listeners`);
        });
        
        console.groupEnd();
    }
}

// 创建全局事件总线实例
export const eventBus = new EventBus();

// 导出类以便扩展
export { EventBus };
