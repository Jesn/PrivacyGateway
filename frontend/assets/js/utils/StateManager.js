/**
 * 状态管理器
 * 提供全局状态管理、状态订阅、状态持久化等功能
 */

class StateManager {
    constructor() {
        this.state = {};
        this.subscribers = new Map();
        this.middleware = [];
        this.history = [];
        this.maxHistorySize = 50;
        
        // 初始化默认状态
        this.initializeState();
    }

    /**
     * 初始化默认状态
     */
    initializeState() {
        this.state = {
            // 认证状态
            auth: {
                isLoggedIn: false,
                user: null,
                sessionExpiry: null
            },
            
            // UI状态
            ui: {
                currentTab: 'overview',
                loading: false,
                notifications: []
            },
            
            // 数据状态
            data: {
                configs: [],
                logs: [],
                tokens: {},
                stats: {}
            },
            
            // 应用设置
            settings: {
                theme: 'light',
                language: 'zh-CN',
                pageSize: 20
            }
        };
    }

    /**
     * 获取状态
     * @param {string} path - 状态路径，如 'auth.isLoggedIn'
     * @returns {*} 状态值
     */
    getState(path) {
        if (!path) return this.state;
        
        return path.split('.').reduce((obj, key) => {
            return obj && obj[key] !== undefined ? obj[key] : undefined;
        }, this.state);
    }

    /**
     * 设置状态
     * @param {string} path - 状态路径
     * @param {*} value - 新值
     * @param {Object} options - 选项
     */
    setState(path, value, options = {}) {
        const { silent = false, merge = false } = options;
        
        // 记录历史
        this.addToHistory(path, this.getState(path), value);
        
        // 应用中间件
        const action = { type: 'SET_STATE', path, value, options };
        const processedAction = this.applyMiddleware(action);
        
        if (processedAction) {
            // 更新状态
            this.updateState(processedAction.path, processedAction.value, merge);
            
            // 通知订阅者
            if (!silent) {
                this.notifySubscribers(processedAction.path, processedAction.value);
            }
        }
    }

    /**
     * 更新状态
     * @private
     */
    updateState(path, value, merge = false) {
        const keys = path.split('.');
        const lastKey = keys.pop();
        
        // 获取父对象
        const parent = keys.reduce((obj, key) => {
            if (!obj[key] || typeof obj[key] !== 'object') {
                obj[key] = {};
            }
            return obj[key];
        }, this.state);
        
        // 设置值
        if (merge && typeof parent[lastKey] === 'object' && typeof value === 'object') {
            parent[lastKey] = { ...parent[lastKey], ...value };
        } else {
            parent[lastKey] = value;
        }
    }

    /**
     * 订阅状态变化
     * @param {string} path - 状态路径
     * @param {Function} callback - 回调函数
     * @returns {Function} 取消订阅函数
     */
    subscribe(path, callback) {
        if (!this.subscribers.has(path)) {
            this.subscribers.set(path, new Set());
        }
        
        this.subscribers.get(path).add(callback);
        
        // 返回取消订阅函数
        return () => {
            const pathSubscribers = this.subscribers.get(path);
            if (pathSubscribers) {
                pathSubscribers.delete(callback);
                if (pathSubscribers.size === 0) {
                    this.subscribers.delete(path);
                }
            }
        };
    }

    /**
     * 通知订阅者
     * @private
     */
    notifySubscribers(path, value) {
        // 通知精确路径的订阅者
        const exactSubscribers = this.subscribers.get(path);
        if (exactSubscribers) {
            exactSubscribers.forEach(callback => {
                try {
                    callback(value, path);
                } catch (error) {
                    console.error('状态订阅回调执行错误:', error);
                }
            });
        }
        
        // 通知父路径的订阅者
        const pathParts = path.split('.');
        for (let i = pathParts.length - 1; i > 0; i--) {
            const parentPath = pathParts.slice(0, i).join('.');
            const parentSubscribers = this.subscribers.get(parentPath);
            if (parentSubscribers) {
                const parentValue = this.getState(parentPath);
                parentSubscribers.forEach(callback => {
                    try {
                        callback(parentValue, parentPath);
                    } catch (error) {
                        console.error('状态订阅回调执行错误:', error);
                    }
                });
            }
        }
        
        // 通知全局订阅者
        const globalSubscribers = this.subscribers.get('*');
        if (globalSubscribers) {
            globalSubscribers.forEach(callback => {
                try {
                    callback(this.state, '*');
                } catch (error) {
                    console.error('状态订阅回调执行错误:', error);
                }
            });
        }
    }

    /**
     * 添加中间件
     * @param {Function} middleware - 中间件函数
     */
    addMiddleware(middleware) {
        this.middleware.push(middleware);
    }

    /**
     * 应用中间件
     * @private
     */
    applyMiddleware(action) {
        return this.middleware.reduce((acc, middleware) => {
            return acc ? middleware(acc, this) : acc;
        }, action);
    }

    /**
     * 添加到历史记录
     * @private
     */
    addToHistory(path, oldValue, newValue) {
        this.history.push({
            timestamp: Date.now(),
            path,
            oldValue,
            newValue
        });
        
        // 限制历史记录大小
        if (this.history.length > this.maxHistorySize) {
            this.history.shift();
        }
    }

    /**
     * 获取历史记录
     * @returns {Array} 历史记录
     */
    getHistory() {
        return [...this.history];
    }

    /**
     * 清除历史记录
     */
    clearHistory() {
        this.history = [];
    }

    /**
     * 批量更新状态
     * @param {Object} updates - 更新对象
     * @param {Object} options - 选项
     */
    batchUpdate(updates, options = {}) {
        const { silent = false } = options;
        
        Object.entries(updates).forEach(([path, value]) => {
            this.setState(path, value, { silent: true });
        });
        
        // 批量通知
        if (!silent) {
            Object.keys(updates).forEach(path => {
                this.notifySubscribers(path, this.getState(path));
            });
        }
    }

    /**
     * 重置状态
     * @param {string} path - 状态路径，不提供则重置全部
     */
    reset(path) {
        if (path) {
            const defaultValue = this.getDefaultValue(path);
            this.setState(path, defaultValue);
        } else {
            this.initializeState();
            this.notifySubscribers('*', this.state);
        }
    }

    /**
     * 获取默认值
     * @private
     */
    getDefaultValue(path) {
        // 这里可以根据路径返回相应的默认值
        const defaults = {
            'auth.isLoggedIn': false,
            'auth.user': null,
            'ui.loading': false,
            'ui.currentTab': 'overview',
            'data.configs': [],
            'data.logs': []
        };
        
        return defaults[path];
    }

    /**
     * 持久化状态到localStorage
     * @param {Array} paths - 要持久化的状态路径
     */
    persist(paths = ['settings']) {
        const persistData = {};
        
        paths.forEach(path => {
            const value = this.getState(path);
            if (value !== undefined) {
                persistData[path] = value;
            }
        });
        
        try {
            localStorage.setItem('app-state', JSON.stringify(persistData));
        } catch (error) {
            console.warn('状态持久化失败:', error);
        }
    }

    /**
     * 从localStorage恢复状态
     * @param {Array} paths - 要恢复的状态路径
     */
    restore(paths = ['settings']) {
        try {
            const persistData = JSON.parse(localStorage.getItem('app-state') || '{}');
            
            paths.forEach(path => {
                if (persistData[path] !== undefined) {
                    this.setState(path, persistData[path], { silent: true });
                }
            });
        } catch (error) {
            console.warn('状态恢复失败:', error);
        }
    }

    /**
     * 创建计算属性
     * @param {string} name - 计算属性名
     * @param {Function} getter - 计算函数
     * @param {Array} dependencies - 依赖的状态路径
     */
    computed(name, getter, dependencies = []) {
        const computedValue = () => getter(this.state);
        
        // 初始计算
        this.setState(`computed.${name}`, computedValue(), { silent: true });
        
        // 监听依赖变化
        dependencies.forEach(dep => {
            this.subscribe(dep, () => {
                this.setState(`computed.${name}`, computedValue());
            });
        });
    }

    /**
     * 创建动作
     * @param {string} name - 动作名
     * @param {Function} action - 动作函数
     */
    action(name, action) {
        this[name] = (...args) => {
            return action(this, ...args);
        };
    }

    /**
     * 调试信息
     */
    debug() {
        console.group('StateManager Debug Info');
        console.log('Current State:', this.state);
        console.log('Subscribers:', this.subscribers);
        console.log('History:', this.history);
        console.log('Middleware:', this.middleware);
        console.groupEnd();
    }
}

// 创建全局状态管理器实例
export const stateManager = new StateManager();

// 导出类以便扩展
export { StateManager };
