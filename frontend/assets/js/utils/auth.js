/**
 * 认证状态管理模块
 * 负责管理用户认证状态、密钥存储和会话控制
 */

class AuthManager {
    constructor() {
        this.adminSecret = '';
        this.sessionKey = 'adminSecret';
        this.expiryKey = 'sessionExpiry';
        this.defaultExpiry = 24 * 60 * 60 * 1000; // 24小时
        
        // 初始化时恢复会话
        this.restoreSession();
    }

    /**
     * 获取管理员密钥
     * @returns {string} 管理员密钥
     */
    getAdminSecret() {
        // 优先返回内存中的密钥，然后检查sessionStorage
        return this.adminSecret || this.getStoredSecret() || '';
    }

    /**
     * 设置管理员密钥
     * @param {string} secret - 管理员密钥
     * @param {number} expiryHours - 过期时间（小时），默认24小时
     */
    setAdminSecret(secret, expiryHours = 24) {
        if (!secret || typeof secret !== 'string') {
            throw new Error('密钥必须是非空字符串');
        }

        this.adminSecret = secret;
        
        // 计算过期时间
        const expiryTime = Date.now() + (expiryHours * 60 * 60 * 1000);
        
        // 存储到sessionStorage
        try {
            sessionStorage.setItem(this.sessionKey, secret);
            sessionStorage.setItem(this.expiryKey, expiryTime.toString());
        } catch (error) {
            console.warn('无法保存会话信息到sessionStorage:', error);
        }

        // 触发状态变化事件
        this.notifyStateChange('login', { secret, expiryTime });
    }

    /**
     * 清除管理员密钥
     */
    clearAdminSecret() {
        const wasLoggedIn = this.isLoggedIn();
        
        this.adminSecret = '';
        
        // 清除sessionStorage
        try {
            sessionStorage.removeItem(this.sessionKey);
            sessionStorage.removeItem(this.expiryKey);
        } catch (error) {
            console.warn('无法清除sessionStorage:', error);
        }

        // 如果之前已登录，触发登出事件
        if (wasLoggedIn) {
            this.notifyStateChange('logout', {});
        }
    }

    /**
     * 检查是否已登录
     * @returns {boolean} 是否已登录
     */
    isLoggedIn() {
        const secret = this.getAdminSecret();
        
        if (!secret || secret.length === 0) {
            return false;
        }

        // 检查会话是否过期
        if (this.isSessionExpired()) {
            this.clearAdminSecret();
            return false;
        }

        return true;
    }

    /**
     * 检查会话是否过期
     * @returns {boolean} 会话是否过期
     */
    isSessionExpired() {
        try {
            const expiryTime = sessionStorage.getItem(this.expiryKey);
            if (!expiryTime) {
                return false; // 没有过期时间设置，认为未过期
            }
            
            return Date.now() > parseInt(expiryTime);
        } catch (error) {
            console.warn('检查会话过期时出错:', error);
            return false;
        }
    }

    /**
     * 获取会话剩余时间（毫秒）
     * @returns {number} 剩余时间，如果已过期返回0
     */
    getSessionRemainingTime() {
        try {
            const expiryTime = sessionStorage.getItem(this.expiryKey);
            if (!expiryTime) {
                return this.defaultExpiry;
            }
            
            const remaining = parseInt(expiryTime) - Date.now();
            return Math.max(0, remaining);
        } catch (error) {
            console.warn('获取会话剩余时间出错:', error);
            return 0;
        }
    }

    /**
     * 延长会话时间
     * @param {number} hours - 延长的小时数
     */
    extendSession(hours = 24) {
        if (!this.isLoggedIn()) {
            throw new Error('用户未登录，无法延长会话');
        }

        const newExpiryTime = Date.now() + (hours * 60 * 60 * 1000);
        
        try {
            sessionStorage.setItem(this.expiryKey, newExpiryTime.toString());
            this.notifyStateChange('sessionExtended', { expiryTime: newExpiryTime });
        } catch (error) {
            console.warn('延长会话时出错:', error);
        }
    }

    /**
     * 从存储中获取密钥
     * @private
     */
    getStoredSecret() {
        try {
            return sessionStorage.getItem(this.sessionKey);
        } catch (error) {
            console.warn('无法从sessionStorage获取密钥:', error);
            return null;
        }
    }

    /**
     * 恢复会话
     * @private
     */
    restoreSession() {
        const storedSecret = this.getStoredSecret();
        
        if (storedSecret && !this.isSessionExpired()) {
            this.adminSecret = storedSecret;
            this.notifyStateChange('sessionRestored', { secret: storedSecret });
        } else if (storedSecret && this.isSessionExpired()) {
            // 会话已过期，清除存储
            this.clearAdminSecret();
        }
    }

    /**
     * 状态变化通知
     * @private
     */
    notifyStateChange(event, data) {
        // 触发自定义事件
        if (typeof window !== 'undefined') {
            window.dispatchEvent(new CustomEvent('authStateChange', {
                detail: { event, data, timestamp: Date.now() }
            }));
        }

        // 调用回调函数
        if (this.onStateChange) {
            this.onStateChange(event, data);
        }
    }

    /**
     * 设置状态变化回调
     * @param {Function} callback - 回调函数
     */
    onStateChange(callback) {
        if (typeof callback === 'function') {
            this.onStateChange = callback;
        }
    }

    /**
     * 验证密钥格式
     * @param {string} secret - 要验证的密钥
     * @returns {boolean} 是否有效
     */
    static validateSecret(secret) {
        if (!secret || typeof secret !== 'string') {
            return false;
        }

        // 基本长度检查
        if (secret.length < 6) {
            return false;
        }

        // 可以添加更多验证规则
        return true;
    }

    /**
     * 生成会话信息摘要
     * @returns {Object} 会话信息
     */
    getSessionInfo() {
        return {
            isLoggedIn: this.isLoggedIn(),
            hasSecret: !!this.getAdminSecret(),
            remainingTime: this.getSessionRemainingTime(),
            isExpired: this.isSessionExpired()
        };
    }
}

// 创建全局实例
export const authManager = new AuthManager();

// 兼容旧的函数接口
export function getAdminSecret() {
    return authManager.getAdminSecret();
}

export function setAdminSecret(secret) {
    return authManager.setAdminSecret(secret);
}

export function clearAdminSecret() {
    return authManager.clearAdminSecret();
}

export function isLoggedIn() {
    return authManager.isLoggedIn();
}

// 导出类以便扩展
export { AuthManager };
