/**
 * HTTP请求工具模块
 * 提供统一的API请求接口，包含认证、错误处理、请求拦截等功能
 */

import { authManager } from './auth.js';

class HttpClient {
    constructor(baseURL = '/') {
        this.baseURL = baseURL;
        this.defaultHeaders = {
            'Content-Type': 'application/json'
        };
        this.requestInterceptors = [];
        this.responseInterceptors = [];
        this.errorHandlers = [];
        
        // 添加默认的认证拦截器
        this.addRequestInterceptor(this.authInterceptor.bind(this));
        
        // 添加默认的错误处理器
        this.addErrorHandler(this.defaultErrorHandler.bind(this));
    }

    /**
     * 认证拦截器
     * @private
     */
    authInterceptor(config) {
        const secret = authManager.getAdminSecret();
        if (secret) {
            config.headers = {
                ...config.headers,
                'X-Log-Secret': secret
            };
        }
        return config;
    }

    /**
     * 默认错误处理器
     * @private
     */
    defaultErrorHandler(error, config) {
        console.error('HTTP请求错误:', error);
        
        // 401错误自动清除认证信息
        if (error.status === 401) {
            authManager.clearAdminSecret();
            this.notifyError('认证失败，请重新登录', 'auth');
        }
        
        // 网络错误
        if (!error.status) {
            this.notifyError('网络连接失败，请检查网络设置', 'network');
        }
        
        return error;
    }

    /**
     * 添加请求拦截器
     * @param {Function} interceptor - 拦截器函数
     */
    addRequestInterceptor(interceptor) {
        if (typeof interceptor === 'function') {
            this.requestInterceptors.push(interceptor);
        }
    }

    /**
     * 添加响应拦截器
     * @param {Function} interceptor - 拦截器函数
     */
    addResponseInterceptor(interceptor) {
        if (typeof interceptor === 'function') {
            this.responseInterceptors.push(interceptor);
        }
    }

    /**
     * 添加错误处理器
     * @param {Function} handler - 错误处理函数
     */
    addErrorHandler(handler) {
        if (typeof handler === 'function') {
            this.errorHandlers.push(handler);
        }
    }

    /**
     * 应用请求拦截器
     * @private
     */
    applyRequestInterceptors(config) {
        return this.requestInterceptors.reduce((acc, interceptor) => {
            return interceptor(acc) || acc;
        }, config);
    }

    /**
     * 应用响应拦截器
     * @private
     */
    applyResponseInterceptors(response, config) {
        return this.responseInterceptors.reduce((acc, interceptor) => {
            return interceptor(acc, config) || acc;
        }, response);
    }

    /**
     * 应用错误处理器
     * @private
     */
    applyErrorHandlers(error, config) {
        this.errorHandlers.forEach(handler => {
            handler(error, config);
        });
        return error;
    }

    /**
     * 通用请求方法
     * @param {string} url - 请求URL
     * @param {Object} options - 请求选项
     */
    async request(url, options = {}) {
        // 检查认证状态
        if (!authManager.isLoggedIn() && !options.skipAuth) {
            throw new Error('需要管理员密钥才能访问此功能');
        }

        // 构建完整URL
        const fullURL = url.startsWith('http') ? url : this.baseURL + url.replace(/^\//, '');
        
        // 构建请求配置
        let config = {
            method: 'GET',
            headers: { ...this.defaultHeaders },
            ...options,
            headers: {
                ...this.defaultHeaders,
                ...options.headers
            }
        };

        // 应用请求拦截器
        config = this.applyRequestInterceptors(config);

        try {
            // 发送请求
            const response = await fetch(fullURL, config);
            
            // 检查响应状态
            if (!response.ok) {
                const errorText = await response.text();
                const error = new Error(`HTTP ${response.status}: ${errorText}`);
                error.status = response.status;
                error.response = response;
                throw error;
            }

            // 解析响应
            let data;
            const contentType = response.headers.get('content-type');
            
            if (contentType && contentType.includes('application/json')) {
                data = await response.json();
            } else {
                data = await response.text();
            }

            // 构建响应对象
            const responseObj = {
                data,
                status: response.status,
                statusText: response.statusText,
                headers: response.headers,
                config
            };

            // 应用响应拦截器
            return this.applyResponseInterceptors(responseObj, config);

        } catch (error) {
            // 应用错误处理器
            this.applyErrorHandlers(error, config);
            throw error;
        }
    }

    /**
     * GET请求
     * @param {string} url - 请求URL
     * @param {Object} params - 查询参数
     * @param {Object} options - 请求选项
     */
    async get(url, params = {}, options = {}) {
        // 构建查询字符串
        const queryString = new URLSearchParams(params).toString();
        const fullURL = queryString ? `${url}?${queryString}` : url;
        
        return this.request(fullURL, {
            ...options,
            method: 'GET'
        });
    }

    /**
     * POST请求
     * @param {string} url - 请求URL
     * @param {*} data - 请求数据
     * @param {Object} options - 请求选项
     */
    async post(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'POST',
            body: this.serializeData(data, options.headers)
        });
    }

    /**
     * PUT请求
     * @param {string} url - 请求URL
     * @param {*} data - 请求数据
     * @param {Object} options - 请求选项
     */
    async put(url, data, options = {}) {
        return this.request(url, {
            ...options,
            method: 'PUT',
            body: this.serializeData(data, options.headers)
        });
    }

    /**
     * DELETE请求
     * @param {string} url - 请求URL
     * @param {Object} options - 请求选项
     */
    async delete(url, options = {}) {
        return this.request(url, {
            ...options,
            method: 'DELETE'
        });
    }

    /**
     * 序列化请求数据
     * @private
     */
    serializeData(data, headers = {}) {
        if (!data) return undefined;
        
        const contentType = headers['Content-Type'] || headers['content-type'] || this.defaultHeaders['Content-Type'];
        
        if (contentType.includes('application/json')) {
            return JSON.stringify(data);
        }
        
        if (data instanceof FormData) {
            return data;
        }
        
        if (typeof data === 'string') {
            return data;
        }
        
        return JSON.stringify(data);
    }

    /**
     * 错误通知
     * @private
     */
    notifyError(message, type = 'error') {
        // 触发自定义事件
        if (typeof window !== 'undefined') {
            window.dispatchEvent(new CustomEvent('httpError', {
                detail: { message, type, timestamp: Date.now() }
            }));
        }
    }

    /**
     * 创建带有特定配置的新实例
     * @param {Object} config - 配置选项
     */
    static create(config = {}) {
        return new HttpClient(config.baseURL);
    }
}

// 创建默认实例
export const httpClient = new HttpClient();

// 兼容旧的API请求函数
export async function apiRequest(url, options = {}) {
    try {
        const response = await httpClient.request(url, options);
        return response.data;
    } catch (error) {
        throw error;
    }
}

// 导出类以便扩展
export { HttpClient };
