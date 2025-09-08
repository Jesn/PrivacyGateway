/**
 * 认证模块
 * 负责登录页面的显示、登录处理、状态切换等功能
 */

import { authManager } from '../utils/auth.js';
import { httpClient } from '../utils/http.js';
import { DOMUtils } from '../utils/dom.js';

class AuthModule {
    constructor() {
        this.loginForm = null;
        this.loginPage = null;
        this.mainApp = null;
        
        this.init();
    }

    /**
     * 初始化认证模块
     */
    init() {
        this.bindElements();
        this.bindEvents();
        this.checkInitialState();
    }

    /**
     * 绑定DOM元素
     */
    bindElements() {
        this.loginForm = DOMUtils.getById('login-form');
        this.loginPage = DOMUtils.getById('login-page');
        this.mainApp = DOMUtils.getById('main-app');
    }

    /**
     * 绑定事件
     */
    bindEvents() {
        // 登录表单提交事件
        if (this.loginForm) {
            DOMUtils.on(this.loginForm, 'submit', this.handleLogin.bind(this));
        }

        // 监听认证状态变化
        window.addEventListener('authStateChange', this.handleAuthStateChange.bind(this));

        // 监听HTTP错误
        window.addEventListener('httpError', this.handleHttpError.bind(this));

        // 页面可见性变化时检查会话
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
                this.checkSessionValidity();
            }
        });
    }

    /**
     * 检查初始状态
     */
    checkInitialState() {
        if (authManager.isLoggedIn()) {
            this.showMainApp();
        } else {
            this.showLoginPage();
        }
    }

    /**
     * 显示登录页面
     */
    showLoginPage() {
        if (this.loginPage) {
            DOMUtils.show(this.loginPage, 'flex');
        }
        if (this.mainApp) {
            DOMUtils.hide(this.mainApp);
        }

        // 聚焦到密钥输入框
        const secretInput = DOMUtils.getById('admin-secret');
        if (secretInput) {
            setTimeout(() => secretInput.focus(), 100);
        }

        // 触发页面显示事件
        this.notifyPageChange('login');
    }

    /**
     * 显示主应用页面
     */
    showMainApp() {
        if (this.loginPage) {
            DOMUtils.hide(this.loginPage);
        }
        if (this.mainApp) {
            DOMUtils.show(this.mainApp, 'block');
        }

        // 触发页面显示事件
        this.notifyPageChange('main');
    }

    /**
     * 处理登录
     * @param {Event} event - 表单提交事件
     */
    async handleLogin(event) {
        event.preventDefault();

        const formData = new FormData(event.target);
        const secret = formData.get('secret');

        if (!secret) {
            this.showError('请输入管理员密钥');
            return;
        }

        // 验证密钥格式
        if (!authManager.constructor.validateSecret(secret)) {
            this.showError('密钥格式不正确，至少需要6个字符');
            return;
        }

        try {
            this.showLoading('验证中...');

            // 测试密钥是否有效
            const response = await httpClient.get('/config/proxy', {}, {
                headers: { 'X-Log-Secret': secret },
                skipAuth: true // 跳过自动认证，使用提供的密钥
            });

            if (response.status === 200) {
                // 登录成功
                authManager.setAdminSecret(secret);
                this.showMainApp();
                this.showSuccess('登录成功');
                
                // 清空表单
                event.target.reset();
            } else {
                this.showError('密钥验证失败');
            }

        } catch (error) {
            console.error('登录失败:', error);
            
            if (error.status === 401 || error.status === 403) {
                this.showError('管理员密钥错误，请检查后重试');
            } else if (error.status === 0) {
                this.showError('无法连接到服务器，请检查网络连接');
            } else {
                this.showError('登录失败: ' + error.message);
            }
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 退出登录
     */
    logout() {
        authManager.clearAdminSecret();
        this.showLoginPage();
        this.showSuccess('已退出登录');
    }

    /**
     * 处理认证状态变化
     * @param {CustomEvent} event - 状态变化事件
     */
    handleAuthStateChange(event) {
        const { event: authEvent, data } = event.detail;

        switch (authEvent) {
            case 'login':
                console.log('用户登录:', data);
                break;
            case 'logout':
                console.log('用户退出');
                this.showLoginPage();
                break;
            case 'sessionExpired':
                this.showError('会话已过期，请重新登录');
                this.showLoginPage();
                break;
            case 'sessionRestored':
                console.log('会话已恢复');
                break;
        }
    }

    /**
     * 处理HTTP错误
     * @param {CustomEvent} event - HTTP错误事件
     */
    handleHttpError(event) {
        const { message, type } = event.detail;

        if (type === 'auth') {
            this.showLoginPage();
        }

        this.showError(message);
    }

    /**
     * 检查会话有效性
     */
    checkSessionValidity() {
        if (!authManager.isLoggedIn()) {
            this.showLoginPage();
            return;
        }

        // 检查会话剩余时间
        const remainingTime = authManager.getSessionRemainingTime();
        const oneHour = 60 * 60 * 1000; // 1小时

        if (remainingTime < oneHour && remainingTime > 0) {
            // 会话即将过期，提醒用户
            this.showWarning(`会话将在 ${Math.ceil(remainingTime / (60 * 1000))} 分钟后过期`);
        }
    }

    /**
     * 显示加载状态
     * @param {string} message - 加载消息
     */
    showLoading(message = '加载中...') {
        // 禁用登录按钮
        const submitBtn = this.loginForm?.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = true;
            submitBtn.textContent = message;
        }
    }

    /**
     * 隐藏加载状态
     */
    hideLoading() {
        // 恢复登录按钮
        const submitBtn = this.loginForm?.querySelector('button[type="submit"]');
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.innerHTML = '<span class="icon">🔑</span> 登录';
        }
    }

    /**
     * 显示成功消息
     * @param {string} message - 消息内容
     */
    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    /**
     * 显示错误消息
     * @param {string} message - 消息内容
     */
    showError(message) {
        this.showMessage(message, 'error');
    }

    /**
     * 显示警告消息
     * @param {string} message - 消息内容
     */
    showWarning(message) {
        this.showMessage(message, 'warning');
    }

    /**
     * 显示消息
     * @param {string} message - 消息内容
     * @param {string} type - 消息类型
     */
    showMessage(message, type = 'info') {
        // 触发消息事件，由消息组件处理
        window.dispatchEvent(new CustomEvent('showMessage', {
            detail: { message, type }
        }));
    }

    /**
     * 页面变化通知
     * @param {string} page - 页面名称
     */
    notifyPageChange(page) {
        window.dispatchEvent(new CustomEvent('pageChange', {
            detail: { page, timestamp: Date.now() }
        }));
    }

    /**
     * 获取认证状态信息
     * @returns {Object} 认证状态
     */
    getAuthStatus() {
        return authManager.getSessionInfo();
    }

    /**
     * 延长会话
     * @param {number} hours - 延长小时数
     */
    extendSession(hours = 24) {
        try {
            authManager.extendSession(hours);
            this.showSuccess(`会话已延长 ${hours} 小时`);
        } catch (error) {
            this.showError('延长会话失败: ' + error.message);
        }
    }
}

// 创建全局实例
export const authModule = new AuthModule();

// 兼容旧的函数接口
export function showLoginPage() {
    authModule.showLoginPage();
}

export function showMainApp() {
    authModule.showMainApp();
}

export function logout() {
    authModule.logout();
}

// 导出类以便扩展
export { AuthModule };
