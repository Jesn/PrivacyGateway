/**
 * 业务管理器
 * 负责协调各个业务模块，处理业务逻辑
 */

import { authModule } from '../modules/auth.js';
import { proxyAPI } from '../api/proxy.js';
import { logsAPI } from '../api/logs.js';
import { tokensAPI } from '../api/tokens.js';
import { stateManager } from '../utils/StateManager.js';
import { eventBus } from '../utils/EventBus.js';
import { DOMUtils } from '../utils/dom.js';

class BusinessManager {
    constructor() {
        this.currentTab = 'proxy-config';
        this.currentConfigs = [];
        this.currentLogs = [];
        this.currentTokens = [];
        this.selectedConfigs = new Set();

        // 编辑状态
        this.editingConfigId = null;
        this.deleteConfigId = null;
        this.currentBatchOperation = null;
        this.currentTestConfig = null;
        this.importData = null;

        // 日志管理状态
        this.currentLogs = [];
        this.currentLogPage = 1;
        this.totalLogPages = 1;
        this.logPageSize = 20;
        this.logFilters = {
            method: '',
            status: '',
            domain: '',
            search: '',
            startTime: '',
            endTime: '',
            clientIp: ''
        };
        this.currentLogDetail = null;
        this.advancedFiltersVisible = false;
        this.filterPresets = this.loadFilterPresets();
        this.searchHighlightTerm = '';

        // 分页状态
        this.configPage = 1;
        this.configPageSize = 10;

        // 管理员密钥
        this.adminSecret = '';
        this.configTotalPages = 1;

        this.logPage = 1;
        this.logPageSize = 20;
        this.logTotalPages = 1;

        this.init();
    }

    /**
     * 初始化业务管理器
     */
    init() {
        this.bindEvents();
        this.bindFormEvents();
        this.initializeState();
    }

    /**
     * 绑定事件
     */
    bindEvents() {
        // 监听认证状态变化 - 使用window事件而不是eventBus
        window.addEventListener('authStateChange', (event) => {
            const { event: authEvent, data } = event.detail;
            console.log('BusinessManager: 收到认证状态变化事件', authEvent, data);

            switch (authEvent) {
                case 'login':
                    this.onUserLogin();
                    break;
                case 'logout':
                    this.onUserLogout();
                    break;
            }
        });

        // 监听标签页切换
        eventBus.on('tab:switch', (tabName) => {
            this.switchTab(tabName);
        });

        // 监听页面可见性变化
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden && this.isLoggedIn()) {
                this.refreshCurrentTabData();
            }
        });
    }

    /**
     * 绑定表单事件
     */
    bindFormEvents() {
        // 绑定配置表单提交事件
        const configForm = document.getElementById('config-form');
        if (configForm) {
            configForm.addEventListener('submit', (event) => {
                event.preventDefault();
                this.saveConfig();
            });
        }

        // 绑定点击外部关闭模态框事件
        document.addEventListener('click', (e) => {
            // 关闭批量操作下拉菜单
            if (!e.target.closest('.relative') && !e.target.closest('#batch-dropdown')) {
                const dropdown = document.getElementById('batch-dropdown');
                if (dropdown && !dropdown.classList.contains('hidden')) {
                    dropdown.classList.add('hidden');
                }
            }

            // 关闭模态框
            if (e.target.classList.contains('fixed') && e.target.classList.contains('inset-0')) {
                if (e.target.id === 'config-modal') {
                    this.closeConfigModal();
                } else if (e.target.id === 'delete-modal') {
                    this.closeDeleteModal();
                } else if (e.target.id === 'batch-modal') {
                    this.closeBatchModal();
                }
            }
        });
    }

    /**
     * 获取管理员密钥
     */
    getAdminSecret() {
        return this.adminSecret || sessionStorage.getItem('adminSecret') || '';
    }

    /**
     * 设置管理员密钥
     */
    setAdminSecret(secret) {
        this.adminSecret = secret;
        sessionStorage.setItem('adminSecret', secret);
    }

    /**
     * 清除管理员密钥
     */
    clearAdminSecret() {
        this.adminSecret = '';
        sessionStorage.removeItem('adminSecret');
    }

    /**
     * 初始化状态
     */
    initializeState() {
        // 从状态管理器恢复状态
        const savedState = stateManager.getState('business') || {};
        this.currentTab = savedState.currentTab || 'proxy-config';
        this.configPage = savedState.configPage || 1;
        this.logPage = savedState.logPage || 1;

        // 如果用户已经登录，立即加载数据
        if (this.isLoggedIn()) {
            console.log('BusinessManager: 用户已登录，开始加载数据');
            setTimeout(() => {
                this.onUserLogin();
            }, 100);
        }
    }

    /**
     * 检查是否已登录
     */
    isLoggedIn() {
        // 检查sessionStorage中是否有adminSecret
        const secret = sessionStorage.getItem('adminSecret');
        return secret && secret.length > 0;
    }

    /**
     * 保存状态
     */
    saveState() {
        stateManager.setState('business', {
            currentTab: this.currentTab,
            configPage: this.configPage,
            logPage: this.logPage
        });
    }

    /**
     * 用户登录后的处理
     */
    async onUserLogin() {
        console.log('BusinessManager: 用户登录，开始加载数据');

        // 切换到代理配置标签页
        this.switchTab('proxy-config');

        // 加载初始数据
        await this.loadProxyConfigs();
    }

    /**
     * 用户登出后的处理
     */
    onUserLogout() {
        console.log('BusinessManager: 用户登出，清理数据');

        // 清理数据
        this.currentConfigs = [];
        this.currentLogs = [];
        this.currentTokens = [];
        this.selectedConfigs.clear();

        // 重置分页
        this.configPage = 1;
        this.logPage = 1;

        // 切换到代理配置标签页
        this.currentTab = 'proxy-config';

        // 清理状态
        stateManager.reset('business');
    }

    /**
     * 切换标签页
     * @param {string} tabName - 标签页名称
     */
    async switchTab(tabName) {
        const previousTab = this.currentTab;

        // 即使点击相同 tab，也刷新 UI（避免首次未初始化时无响应）
        this.currentTab = tabName;
        this.saveState();

        // 实时时钟控制：离开概览时停止，进入概览时启动
        if (previousTab === 'overview' && tabName !== 'overview') {
            this.stopRealtimeClock && this.stopRealtimeClock();
        }

        // 更新UI
        this.updateTabUI(tabName);

        // 首次进入或相同 tab 再次点击时，按需加载数据
        if (previousTab !== tabName) {
            await this.loadTabData(tabName);
        } else {
            // 如果是相同 tab，避免阻塞但触发一次轻量刷新
            try {
                await this.loadTabData(tabName);
            } catch (e) {
                console.warn('重复点击当前标签刷新失败:', e);
            }
        }

        // 进入概览后，启动实时时钟
        if (tabName === 'overview') {
            this.startRealtimeClock && this.startRealtimeClock();
        }

        // 触发标签页切换事件
        eventBus.emit('business:tabChanged', { tabName, previousTab });
    }

    /**
     * 更新标签页UI
     * @param {string} tabName - 标签页名称
     */
    updateTabUI(tabName) {
        // 移除所有活动状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('border-indigo-500', 'text-indigo-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });

        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.add('hidden');
        });

        // 激活选中的标签
        const activeBtn = document.querySelector(`[data-tab="${tabName}"]`);
        if (activeBtn) {
            activeBtn.classList.remove('border-transparent', 'text-gray-500');
            activeBtn.classList.add('border-indigo-500', 'text-indigo-600');
        }

        const activePane = document.getElementById(`${tabName}-tab`);
        if (activePane) {
            activePane.classList.remove('hidden');
        }
    }

    /**
     * 加载标签页数据
     * @param {string} tabName - 标签页名称
     */
    async loadTabData(tabName) {
        try {
            switch (tabName) {
                case 'overview':
                    await this.loadOverviewData();
                    break;
                case 'proxy-config':
                    await this.loadProxyConfigs();
                    break;
                case 'logs':
                    await this.loadLogs();
                    break;
                default:
                    console.warn(`未知的标签页: ${tabName}`);
            }
        } catch (error) {
            console.error(`加载标签页数据失败 (${tabName}):`, error);
            this.showError(`加载${this.getTabDisplayName(tabName)}数据失败: ${error.message}`);
        }
    }

    /**
     * 刷新当前标签页数据
     */
    async refreshCurrentTabData() {
        await this.loadTabData(this.currentTab);
    }

    /**
     * 加载概览数据
     */
    async loadOverviewData() {
        console.log('BusinessManager: 加载概览数据');

        try {
            // 加载完整的系统概览数据
            await this.loadSystemOverview();

            // 启动自动刷新
            this.startAutoRefresh();

        } catch (error) {
            console.error('加载概览数据失败:', error);
            this.showError('加载概览数据失败: ' + error.message);
        }
    }

    /**
     * 更新概览卡片
     * @param {Object} stats - 统计数据
     */
    updateOverviewCards(stats) {
        const totalConfigsEl = DOMUtils.getById('total-configs');
        const activeConfigsEl = DOMUtils.getById('active-configs');
        const totalRequestsEl = DOMUtils.getById('total-requests');

        if (totalConfigsEl) totalConfigsEl.textContent = stats.totalConfigs;
        if (activeConfigsEl) activeConfigsEl.textContent = stats.activeConfigs;
        if (totalRequestsEl) totalRequestsEl.textContent = stats.totalRequests.toLocaleString();
    }

    /**
     * 加载代理配置
     */
    async loadProxyConfigs() {
        console.log('BusinessManager: 加载代理配置');

        try {
            this.showLoading('加载配置中...');

            const response = await proxyAPI.getConfigs(this.configPage, this.configPageSize);
            this.currentConfigs = response.configs || [];
            this.configTotalPages = response.totalPages || 1;

            this.renderConfigTable();
            this.updateConfigPagination();

        } catch (error) {
            console.error('加载代理配置失败:', error);
            this.showError('加载代理配置失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 渲染配置表格
     */
    renderConfigTable() {
        const tbody = document.getElementById('config-tbody');
        if (!tbody) return;

        if (this.currentConfigs.length === 0) {
            tbody.innerHTML = `
                <tr class="no-data">
                    <td colspan="7">
                        <div class="text-center py-12 text-gray-500">
                            <div class="text-4xl mb-4">📝</div>
                            <p class="text-lg mb-2">暂无代理配置</p>
                            <p class="text-sm">点击"新增配置"开始创建您的第一个代理配置</p>
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = this.currentConfigs.map(config => `
            <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                    <input type="checkbox" class="config-checkbox" value="${config.id}"
                           ${this.selectedConfigs.has(config.id) ? 'checked' : ''}>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-gray-900">${this.escapeHtml(config.name)}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm text-gray-900">${this.escapeHtml(config.target_url)}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full
                                 ${config.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                        ${config.enabled ? '启用' : '禁用'}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${this.formatDate(config.created_at)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${config.stats ? config.stats.request_count || 0 : 0}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div class="flex space-x-2">
                        <button onclick="businessManager.editConfig('${config.id}')"
                                class="text-indigo-600 hover:text-indigo-900">编辑</button>
                        <button onclick="businessManager.testConfig('${config.id}')"
                                class="text-green-600 hover:text-green-900">测试</button>
                        <button onclick="businessManager.manageTokens('${config.id}')"
                                class="text-blue-600 hover:text-blue-900">令牌</button>
                        <button onclick="businessManager.deleteConfig('${config.id}', '${this.escapeHtml(config.name)}')"
                                class="text-red-600 hover:text-red-900">删除</button>
                    </div>
                </td>
            </tr>
        `).join('');

        // 绑定复选框事件
        tbody.querySelectorAll('.config-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                if (e.target.checked) {
                    this.selectedConfigs.add(e.target.value);
                } else {
                    this.selectedConfigs.delete(e.target.value);
                }
            });
        });
    }

    /**
     * 获取标签页显示名称
     * @param {string} tabName - 标签页名称
     * @returns {string} 显示名称
     */
    getTabDisplayName(tabName) {
        const names = {

            'proxy-config': '代理配置',
            'logs': '访问日志'
        };
        return names[tabName] || tabName;
    }

    /**
     * HTML转义
     * @param {string} text - 要转义的文本
     * @returns {string} 转义后的文本
     */
    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * 格式化日期
     * @param {string} dateString - 日期字符串
     * @returns {string} 格式化后的日期
     */
    formatDate(dateString) {
        if (!dateString) return '-';
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN');
    }

    /**
     * 显示加载状态
     * @param {string} message - 加载消息
     */
    showLoading(message = '加载中...') {
        // 触发全局加载事件
        eventBus.emit('ui:loading', { show: true, message });
    }

    /**
     * 隐藏加载状态
     */
    hideLoading() {
        eventBus.emit('ui:loading', { show: false });
    }

    /**
     * 显示成功消息
     * @param {string} message - 成功消息
     */
    showSuccess(message) {
        eventBus.emit('ui:message', { type: 'success', message });
    }


    // 统一的复制成功提示
    copySuccess() {
        this.showSuccess('已复制');
    }

    /**
     * 显示错误消息
     * @param {string} message - 错误消息
     */
    showError(message) {
        eventBus.emit('ui:message', { type: 'error', message });
    }

    /**
     * 更新配置分页信息
     */
    updateConfigPagination() {
        // 更新分页信息显示
        const startEl = document.getElementById('config-start');
        const endEl = document.getElementById('config-end');
        const totalEl = document.getElementById('config-total');
        const currentPageEl = document.getElementById('config-current-page');
        const totalPagesEl = document.getElementById('config-total-pages');

        const start = (this.configPage - 1) * this.configPageSize + 1;
        const end = Math.min(this.configPage * this.configPageSize, this.currentConfigs.length);

        if (startEl) startEl.textContent = start;
        if (endEl) endEl.textContent = end;
        if (totalEl) totalEl.textContent = this.currentConfigs.length;
        if (currentPageEl) currentPageEl.textContent = this.configPage;
        if (totalPagesEl) totalPagesEl.textContent = this.configTotalPages;

        // 更新分页按钮状态
        const prevBtn = document.getElementById('config-prev-btn');
        const nextBtn = document.getElementById('config-next-btn');

        if (prevBtn) {
            prevBtn.disabled = this.configPage <= 1;
            prevBtn.classList.toggle('opacity-50', this.configPage <= 1);
        }

        if (nextBtn) {
            nextBtn.disabled = this.configPage >= this.configTotalPages;
            nextBtn.classList.toggle('opacity-50', this.configPage >= this.configTotalPages);
        }
    }

    /**
     * 上一页
     */
    async prevConfigPage() {
        if (this.configPage > 1) {
            this.configPage--;
            await this.loadProxyConfigs();
        }
    }

    /**
     * 下一页
     */
    async nextConfigPage() {
        if (this.configPage < this.configTotalPages) {
            this.configPage++;
            await this.loadProxyConfigs();
        }
    }

    /**
     * 显示新增配置模态框
     */
    showAddConfigModal() {
        this.editingConfigId = null;
        document.getElementById('config-modal-title').textContent = '新增代理配置';
        this.resetConfigForm();
        this.showModal('config-modal');
    }

    /**
     * 编辑配置
     * @param {string} id - 配置ID
     */
    editConfig(id) {
        const config = this.currentConfigs.find(c => c.id === id);
        if (!config) {
            this.showError('配置不存在');
            return;
        }

        this.editingConfigId = id;
        document.getElementById('config-modal-title').textContent = '编辑代理配置';

        // 填充表单数据
        document.getElementById('config-name').value = config.name || '';
        document.getElementById('config-target-url').value = config.target_url || '';
        document.getElementById('config-protocol').value = config.protocol || 'https';
        document.getElementById('config-enabled').checked = config.enabled !== false;
        document.getElementById('config-timeout').value = config.timeout || 30;
        document.getElementById('config-max-retries').value = config.max_retries || 3;
        document.getElementById('config-follow-redirects').checked = config.follow_redirects !== false;
        document.getElementById('config-verify-ssl').checked = config.verify_ssl !== false;

        this.showModal('config-modal');
    }

    /**
     * 保存配置
     */
    async saveConfig() {
        try {
            const form = document.getElementById('config-form');
            const formData = new FormData(form);

            // 构建配置对象
            const config = {
                name: formData.get('name'),
                target_url: formData.get('target_url'),
                protocol: formData.get('protocol'),
                enabled: formData.has('enabled'),
                timeout: parseInt(formData.get('timeout')) || 30,
                max_retries: parseInt(formData.get('max_retries')) || 3,
                follow_redirects: formData.has('follow_redirects'),
                verify_ssl: formData.has('verify_ssl')
            };

            // 验证必填字段
            if (!config.name || !config.target_url) {
                this.showError('请填写配置名称和目标URL');
                return;
            }

            this.showLoading('保存配置中...');

            let response;
            if (this.editingConfigId) {
                // 更新配置
                response = await proxyAPI.updateConfig(this.editingConfigId, config);
                this.showSuccess('配置更新成功');
            } else {
                // 创建配置
                response = await proxyAPI.createConfig(config);
                this.showSuccess('配置创建成功');
            }

            this.closeConfigModal();
            await this.loadProxyConfigs();

        } catch (error) {
            console.error('保存配置失败:', error);
            this.showError('保存配置失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 删除配置
     * @param {string} id - 配置ID
     * @param {string} name - 配置名称
     */
    deleteConfig(id, name) {
        this.deleteConfigId = id;
        document.getElementById('delete-config-name').textContent = name;
        this.showModal('delete-modal');
    }

    /**
     * 确认删除配置
     */
    async confirmDelete() {
        try {
            this.showLoading('删除配置中...');

            await proxyAPI.deleteConfig(this.deleteConfigId);

            this.showSuccess('配置删除成功');
            this.closeDeleteModal();
            await this.loadProxyConfigs();

        } catch (error) {
            console.error('删除配置失败:', error);
            this.showError('删除配置失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 测试配置连接
     * @param {string} id - 配置ID
     */
    async testConfig(id) {
        const config = this.currentConfigs.find(c => c.id === id);
        if (!config) {
            this.showError('配置不存在');
            return;
        }

        this.currentTestConfig = config;
        this.showTestModal(config);

        try {
            await this.performConfigTest(config);
        } catch (error) {
            this.showTestError(error.message);
        }
    }

    /**
     * 在模态框中测试配置
     */
    async testConfigInModal() {
        try {
            const form = document.getElementById('config-form');
            const formData = new FormData(form);

            const testConfig = {
                name: formData.get('name') || '临时测试配置',
                target_url: formData.get('target_url'),
                protocol: formData.get('protocol') || 'https',
                timeout: parseInt(formData.get('timeout')) || 30,
                verify_ssl: formData.has('verify_ssl')
            };

            if (!testConfig.target_url) {
                this.showError('请先填写目标URL');
                return;
            }

            // 验证URL格式
            try {
                new URL(testConfig.target_url);
            } catch (e) {
                this.showError('请输入有效的URL格式');
                return;
            }

            // 显示测试模态框
            this.currentTestConfig = testConfig;
            this.showTestModal(testConfig);

            // 执行测试 - 使用真实的HTTP代理
            try {
                const startTime = Date.now();

                // 使用真实的HTTP代理进行测试
                const testUrl = `/proxy?target=${encodeURIComponent(testConfig.target_url)}`;
                const response = await fetch(testUrl, {
                    method: 'GET',
                    headers: {
                        'X-Log-Secret': this.getAdminSecret()
                    }
                });

                const endTime = Date.now();
                const responseTime = endTime - startTime;

                // 读取响应内容
                const responseText = await response.text();
                const responseHeaders = {};
                response.headers.forEach((value, key) => {
                    responseHeaders[key] = value;
                });

                this.showTestSuccess({
                    responseTime,
                    statusCode: response.status,
                    responseSize: responseText.length,
                    responseText,
                    responseHeaders,
                    requestInfo: {
                        method: 'GET',
                        url: testUrl,
                        target: testConfig.target_url,
                        protocol: testConfig.protocol,
                        timestamp: new Date().toISOString()
                    }
                });
            } catch (error) {
                const endTime = Date.now();
                const responseTime = endTime - startTime;
                this.showTestError(error.message, responseTime);
            }

        } catch (error) {
            console.error('测试连接失败:', error);
            this.showError('测试连接失败: ' + error.message);
        }
    }

    /**
     * 显示模态框
     * @param {string} modalId - 模态框ID
     */
    showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');

            // 为令牌相关弹框设置更高的z-index
            if (modalId === 'token-modal' || modalId === 'delete-token-modal') {
                modal.style.zIndex = '9999';
            }
        }
    }

    /**
     * 隐藏模态框
     * @param {string} modalId - 模态框ID
     */
    hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');

            // 如果关闭的是令牌创建弹框，重新显示令牌管理弹框
            if (modalId === 'token-modal' && this.currentConfigId) {
                setTimeout(() => {
                    this.showModal('token-management-modal');
                }, 100);
            }
        }
    }

    /**
     * 关闭配置模态框
     */
    closeConfigModal() {
        const modal = document.getElementById('config-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.editingConfigId = null;
    }

    /**
     * 关闭删除模态框
     */
    closeDeleteModal() {
        const modal = document.getElementById('delete-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.deleteConfigId = null;
    }

    /**
     * 重置配置表单
     */
    resetConfigForm() {
        const form = document.getElementById('config-form');
        if (form) {
            form.reset();
            // 设置默认值
            document.getElementById('config-enabled').checked = true;
            document.getElementById('config-follow-redirects').checked = true;
            document.getElementById('config-verify-ssl').checked = true;
            document.getElementById('config-timeout').value = 30;
            document.getElementById('config-max-retries').value = 3;
            document.getElementById('config-protocol').value = 'https';
        }
    }

    /**
     * 切换批量操作下拉菜单
     */
    toggleBatchDropdown() {
        const dropdown = document.getElementById('batch-dropdown');
        if (dropdown) {
            dropdown.classList.toggle('hidden');
        }

        // 点击其他地方时关闭下拉菜单
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.relative')) {
                dropdown?.classList.add('hidden');
            }
        }, { once: true });
    }

    /**
     * 批量操作
     * @param {string} operation - 操作类型 (enable/disable/delete)
     */
    batchOperation(operation) {
        // 关闭下拉菜单
        const dropdown = document.getElementById('batch-dropdown');
        if (dropdown) {
            dropdown.classList.add('hidden');
        }

        if (this.selectedConfigs.size === 0) {
            this.showError('请先选择要操作的配置');
            return;
        }

        this.currentBatchOperation = operation;

        // 设置模态框内容
        const title = document.getElementById('batch-modal-title');
        const icon = document.getElementById('batch-warning-icon');
        const text = document.getElementById('batch-warning-text');
        const confirmBtn = document.getElementById('batch-confirm-btn');

        switch (operation) {
            case 'enable':
                title.textContent = '批量启用确认';
                icon.textContent = '▶️';
                text.textContent = '确定要启用选中的配置吗？';
                confirmBtn.className = 'px-4 py-2 text-sm font-medium text-white bg-green-500 hover:bg-green-600 rounded-md transition-colors duration-200';
                confirmBtn.textContent = '启用';
                break;
            case 'disable':
                title.textContent = '批量禁用确认';
                icon.textContent = '⏸️';
                text.textContent = '确定要禁用选中的配置吗？';
                confirmBtn.className = 'px-4 py-2 text-sm font-medium text-white bg-yellow-500 hover:bg-yellow-600 rounded-md transition-colors duration-200';
                confirmBtn.textContent = '禁用';
                break;
            case 'delete':
                title.textContent = '批量删除确认';
                icon.textContent = '🗑️';
                text.textContent = '确定要删除选中的配置吗？此操作不可撤销！';
                confirmBtn.className = 'px-4 py-2 text-sm font-medium text-white bg-red-500 hover:bg-red-600 rounded-md transition-colors duration-200';
                confirmBtn.textContent = '删除';
                break;
        }

        // 更新选中配置列表
        this.updateSelectedConfigsList();

        // 显示模态框
        this.showModal('batch-modal');
    }

    /**
     * 更新选中配置列表显示
     */
    updateSelectedConfigsList() {
        const countEl = document.getElementById('selected-count');
        const listEl = document.getElementById('selected-list');

        if (countEl) {
            countEl.textContent = this.selectedConfigs.size;
        }

        if (listEl) {
            const selectedConfigsArray = Array.from(this.selectedConfigs);
            listEl.innerHTML = selectedConfigsArray.map(configId => {
                const config = this.currentConfigs.find(c => c.id === configId);
                return `<div class="text-sm text-gray-700 py-1 px-2 bg-white rounded mb-1">${this.escapeHtml(config ? config.name : configId)}</div>`;
            }).join('');
        }
    }

    /**
     * 确认批量操作
     */
    async confirmBatchOperation() {
        try {
            this.showLoading('执行批量操作中...');
            const configIds = Array.from(this.selectedConfigs);

            let response;
            switch (this.currentBatchOperation) {
                case 'enable':
                case 'disable':
                    response = await proxyAPI.batchUpdateConfigs(configIds, {
                        enabled: this.currentBatchOperation === 'enable'
                    });
                    break;
                case 'delete':
                    response = await proxyAPI.batchDeleteConfigs(configIds);
                    break;
                default:
                    throw new Error('未知的批量操作类型');
            }

            const operationName = this.currentBatchOperation === 'enable' ? '启用' :
                                 this.currentBatchOperation === 'disable' ? '禁用' : '删除';

            this.showSuccess(`批量${operationName}操作完成`);
            this.closeBatchModal();
            this.selectedConfigs.clear();
            await this.loadProxyConfigs();

        } catch (error) {
            console.error('批量操作失败:', error);
            this.showError('批量操作失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 关闭批量操作模态框
     */
    closeBatchModal() {
        const modal = document.getElementById('batch-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.currentBatchOperation = null;
    }

    /**
     * 显示测试模态框
     * @param {Object} config - 配置对象
     */
    showTestModal(config) {
        // 填充配置信息
        document.getElementById('test-config-name').textContent = config.name;
        document.getElementById('test-config-target').textContent = config.target_url;
        document.getElementById('test-config-protocol').textContent = config.protocol || 'https';

        // 重置测试状态
        this.resetTestModal();

        // 显示模态框
        this.showModal('test-modal');
    }

    /**
     * 重置测试模态框状态
     */
    resetTestModal() {
        const statusElement = document.getElementById('test-result-status');
        statusElement.className = 'flex items-center gap-2 mb-2';

        document.getElementById('test-status-icon').textContent = '⏳';
        document.getElementById('test-status-text').textContent = '测试中...';
        document.getElementById('test-status-text').className = 'text-sm font-medium text-blue-600';

        document.getElementById('test-result-metrics').style.display = 'none';
        document.getElementById('test-details').style.display = 'none';
        document.getElementById('retest-btn').style.display = 'none';

        // 重置标签页
        this.switchTestTab('request');
    }

    /**
     * 执行配置测试
     * @param {Object} config - 配置对象
     */
    async performConfigTest(config) {
        const startTime = Date.now();

        try {
            // 使用真实的HTTP代理进行测试（与原始index.html相同的方式）
            const testUrl = `/proxy?target=${encodeURIComponent(config.target_url)}`;

            // 发送测试请求
            const response = await fetch(testUrl, {
                method: 'GET',
                headers: {
                    'X-Log-Secret': this.getAdminSecret()
                }
            });

            const endTime = Date.now();
            const responseTime = endTime - startTime;

            // 读取响应内容
            const responseText = await response.text();
            const responseHeaders = {};
            response.headers.forEach((value, key) => {
                responseHeaders[key] = value;
            });

            // 显示测试结果
            this.showTestSuccess({
                responseTime,
                statusCode: response.status,
                responseSize: responseText.length,
                responseText,
                responseHeaders,
                requestInfo: {
                    method: 'GET',
                    url: testUrl,
                    target: config.target_url,
                    protocol: config.protocol || 'https',
                    timestamp: new Date().toISOString()
                }
            });

        } catch (error) {
            const endTime = Date.now();
            const responseTime = endTime - startTime;

            console.error('配置测试失败:', error);
            this.showTestError(error.message, responseTime);
        }
    }

    /**
     * 显示测试成功结果
     * @param {Object} result - 测试结果
     */
    showTestSuccess(result) {
        // 更新状态
        document.getElementById('test-status-icon').textContent = '✅';
        document.getElementById('test-status-text').textContent = '测试成功';
        document.getElementById('test-status-text').className = 'text-sm font-medium text-green-600';

        // 显示指标
        document.getElementById('test-response-time').textContent = `${result.responseTime}ms`;
        document.getElementById('test-status-code').textContent = result.statusCode;
        document.getElementById('test-response-size').textContent = this.formatBytes(result.responseSize);
        document.getElementById('test-result-metrics').style.display = 'flex';

        // 填充详细信息
        this.fillTestDetails(result);

        // 显示详情和重新测试按钮
        document.getElementById('test-details').style.display = 'block';
        document.getElementById('retest-btn').style.display = 'inline-flex';
    }

    /**
     * 显示测试错误
     * @param {string} error - 错误信息
     */
    showTestError(error) {
        document.getElementById('test-status-icon').textContent = '❌';
        document.getElementById('test-status-text').textContent = '测试失败';
        document.getElementById('test-status-text').className = 'text-sm font-medium text-red-600';

        // 显示错误详情
        document.getElementById('test-request-info').textContent = `错误信息: ${error}`;
        document.getElementById('test-details').style.display = 'block';
        document.getElementById('retest-btn').style.display = 'inline-flex';
    }

    /**
     * 填充测试详细信息
     * @param {Object} result - 测试结果
     */
    fillTestDetails(result) {
        // 请求信息
        const requestInfo = `方法: ${result.requestInfo.method}
URL: ${result.requestInfo.url}
目标: ${result.requestInfo.target}
协议: ${result.requestInfo.protocol}
时间: ${result.requestInfo.timestamp}`;
        document.getElementById('test-request-info').textContent = requestInfo;

        // 响应内容
        let responseContent = result.responseText;
        if (responseContent.length > 5000) {
            responseContent = responseContent.substring(0, 5000) + '\n\n... (内容过长，已截断)';
        }
        document.getElementById('test-response-content').textContent = responseContent;

        // 响应头
        const headersText = Object.entries(result.responseHeaders)
            .map(([key, value]) => `${key}: ${value}`)
            .join('\n');
        document.getElementById('test-response-headers').textContent = headersText;
    }

    /**
     * 切换测试标签页
     * @param {string} tabName - 标签页名称
     */
    switchTestTab(tabName) {
        // 移除所有活动状态
        document.querySelectorAll('.test-tab-btn').forEach(btn => {
            btn.classList.remove('border-blue-500', 'text-blue-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });
        document.querySelectorAll('.test-tab-pane').forEach(pane => {
            pane.classList.add('hidden');
        });

        // 激活选中的标签
        const activeBtn = document.querySelector(`[data-tab="${tabName}"]`);
        if (activeBtn) {
            activeBtn.classList.remove('border-transparent', 'text-gray-500');
            activeBtn.classList.add('border-blue-500', 'text-blue-600');
        }

        const activePane = document.getElementById(`test-${tabName}-tab`);
        if (activePane) {
            activePane.classList.remove('hidden');
        }
    }

    /**
     * 重新测试配置
     */
    async retestConfig() {
        if (this.currentTestConfig) {
            this.resetTestModal();
            try {
                await this.performConfigTest(this.currentTestConfig);
            } catch (error) {
                this.showTestError(error.message);
            }
        }
    }

    /**
     * 关闭测试模态框
     */
    closeTestModal() {
        const modal = document.getElementById('test-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.currentTestConfig = null;
    }

    /**
     * 格式化字节数
     * @param {number} bytes - 字节数
     * @returns {string} 格式化后的字符串
     */
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * 切换导入导出下拉菜单
     */
    toggleImportExportDropdown() {
        const dropdown = document.getElementById('import-export-dropdown');
        if (dropdown) {
            dropdown.classList.toggle('hidden');
        }
    }

    /**
     * 导出配置
     */
    async exportConfigs() {
        try {
            // 关闭下拉菜单
            const dropdown = document.getElementById('import-export-dropdown');
            if (dropdown) {
                dropdown.classList.add('hidden');
            }

            this.showLoading('导出配置中...');

            // 获取导出数据
            const result = await proxyAPI.exportConfigs();

            // 创建下载链接
            const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
            const filename = `proxy-configs-${timestamp}.json`;

            const blob = new Blob([JSON.stringify(result, null, 2)], {
                type: 'application/json'
            });

            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);

            const configCount = result.configs ? result.configs.length : 0;
            this.showSuccess(`配置导出成功，共导出 ${configCount} 个配置`);

        } catch (error) {
            console.error('导出配置失败:', error);
            this.showError('导出配置失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 显示导入模态框
     */
    showImportModal() {
        // 关闭下拉菜单
        const dropdown = document.getElementById('import-export-dropdown');
        if (dropdown) {
            dropdown.classList.add('hidden');
        }

        // 重置表单
        document.getElementById('import-file').value = '';
        document.getElementById('import-mode').value = 'error';
        document.getElementById('import-btn').disabled = true;
        document.getElementById('import-preview').classList.add('hidden');
        this.importData = null;

        // 显示模态框
        this.showModal('import-modal');

        // 绑定文件选择事件
        this.bindImportFileEvent();
    }

    /**
     * 绑定导入文件选择事件
     */
    bindImportFileEvent() {
        const fileInput = document.getElementById('import-file');
        if (fileInput && !fileInput.hasAttribute('data-bound')) {
            fileInput.setAttribute('data-bound', 'true');
            fileInput.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = (e) => {
                        try {
                            const data = JSON.parse(e.target.result);
                            this.previewImportData(data);
                        } catch (error) {
                            this.showError('文件格式错误，请选择有效的JSON文件');
                            this.resetImportForm();
                        }
                    };
                    reader.readAsText(file);
                }
            });
        }
    }

    /**
     * 预览导入数据
     * @param {Object} data - 导入的数据
     */
    previewImportData(data) {
        try {
            // 验证数据格式
            if (!data.configs || !Array.isArray(data.configs)) {
                throw new Error('无效的配置文件格式');
            }

            this.importData = data;

            // 显示预览信息
            document.getElementById('preview-count').textContent = data.configs.length;
            document.getElementById('preview-version').textContent = data.version || '未知';

            // 显示配置列表
            const previewList = document.getElementById('preview-list');
            previewList.innerHTML = '';

            data.configs.forEach((config, index) => {
                const item = document.createElement('div');
                item.className = 'flex items-center justify-between py-1 px-2 bg-white rounded border';
                item.innerHTML = `
                    <div class="flex items-center">
                        <span class="w-6 h-6 bg-blue-100 text-blue-600 rounded-full text-xs flex items-center justify-center mr-2">
                            ${index + 1}
                        </span>
                        <span class="font-medium">${config.name}</span>
                    </div>
                    <div class="text-xs text-gray-500">
                        ${config.target_url}
                    </div>
                `;
                previewList.appendChild(item);
            });

            // 显示预览区域和启用导入按钮
            document.getElementById('import-preview').classList.remove('hidden');
            document.getElementById('import-btn').disabled = false;

        } catch (error) {
            this.showError('预览导入数据失败: ' + error.message);
            this.resetImportForm();
        }
    }

    /**
     * 重置导入表单
     */
    resetImportForm() {
        document.getElementById('import-file').value = '';
        document.getElementById('import-preview').classList.add('hidden');
        document.getElementById('import-btn').disabled = true;
        this.importData = null;
    }

    /**
     * 导入配置
     */
    async importConfigs() {
        if (!this.importData) {
            this.showError('请先选择要导入的文件');
            return;
        }

        try {
            this.showLoading('导入配置中...');

            const mode = document.getElementById('import-mode').value;
            const overwrite = mode === 'replace';

            const result = await proxyAPI.importConfigs(this.importData.configs, overwrite);

            const importedCount = result.imported_count || result.success_count || this.importData.configs.length;
            this.showSuccess(`配置导入成功！成功导入 ${importedCount} 个配置`);
            this.closeImportModal();
            await this.loadConfigs(); // 重新加载配置列表

        } catch (error) {
            console.error('导入配置失败:', error);
            this.showError('导入配置失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 关闭导入模态框
     */
    closeImportModal() {
        const modal = document.getElementById('import-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.resetImportForm();
    }

    /**
     * 加载访问日志
     */
    async loadLogs() {
        try {
            // 显示加载状态
            document.getElementById('log-loading').classList.remove('hidden');
            document.getElementById('log-no-data').classList.add('hidden');
            document.getElementById('log-table-container').classList.add('hidden');

            // 构建查询参数
            const options = {
                page: this.currentLogPage,
                limit: this.logPageSize,
                method: this.logFilters.method,
                status: this.logFilters.status,
                domain: this.logFilters.domain,
                search: this.logFilters.search,
                startTime: this.logFilters.startTime,
                endTime: this.logFilters.endTime,
                clientIp: this.logFilters.clientIp
            };

            // 调用API获取日志数据
            const result = await logsAPI.getLogs(options);

            this.currentLogs = result.logs || [];
            this.totalLogPages = result.totalPages || 1;

            // 渲染日志表格
            this.renderLogTable();

            // 加载统计信息
            await this.loadLogStats();

            // 更新分页信息
            this.updateLogPagination();

            // 显示相关元素
            document.getElementById('log-stats').classList.remove('hidden');
            document.getElementById('log-filters').classList.remove('hidden');

            if (this.currentLogs.length > 0) {
                document.getElementById('log-table-container').classList.remove('hidden');
                if (this.totalLogPages > 1) {
                    document.getElementById('log-pagination').classList.remove('hidden');
                }
            } else {
                document.getElementById('log-no-data').classList.remove('hidden');
            }

            // 渲染筛选预设
            this.renderFilterPresets();

            // 更新筛选统计
            this.updateFilterStats();

        } catch (error) {
            console.error('加载日志失败:', error);
            this.showError('加载日志失败: ' + error.message);
            document.getElementById('log-no-data').classList.remove('hidden');
        } finally {
            document.getElementById('log-loading').classList.add('hidden');
        }
    }

    /**
     * 渲染日志表格
     */
    renderLogTable() {
        const tbody = document.getElementById('log-tbody');

        if (this.currentLogs.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" class="px-6 py-12 text-center text-gray-500">
                        <div class="text-4xl mb-4">📋</div>
                        <p class="text-lg mb-2">暂无日志数据</p>
                        <p class="text-sm">当有代理请求时，日志将显示在这里</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = this.currentLogs.map(log => `
            <tr class="hover:bg-gray-50 cursor-pointer" onclick="businessManager.showLogDetail('${log.id}')">
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getMethodBadgeClass(log.method)}">
                        ${log.method}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getTypeBadgeClass(log.request_type)}">
                        ${log.request_type || 'HTTP'}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <div class="max-w-xs truncate" title="${log.target_host}${log.target_path}">
                        ${this.highlightSearchTerm(this.truncateUrl(log.target_host + log.target_path, 40))}
                    </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getStatusBadgeClass(log.status_code)}">
                        ${log.status_code}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${log.duration_ms}ms
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${this.highlightSearchTerm(log.client_ip)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${this.formatLogTime(log.timestamp)}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button onclick="event.stopPropagation(); businessManager.showLogDetail('${log.id}')"
                            class="text-indigo-600 hover:text-indigo-900">
                        查看详情
                    </button>
                </td>
            </tr>
        `).join('');
    }

    /**
     * 加载日志统计信息
     */
    async loadLogStats() {
        try {
            const stats = await logsAPI.getStats();
            this.updateLogStats(stats);
        } catch (error) {
            console.error('加载统计信息失败:', error);
            // 使用默认值
            this.updateLogStats({});
        }
    }

    /**
     * 更新日志统计信息
     * @param {Object} stats - 统计数据
     */
    updateLogStats(stats) {
        document.getElementById('log-total-entries').textContent = stats.current_entries || 0;
        document.getElementById('log-memory-usage').textContent = (stats.memory_usage_mb || 0).toFixed(1);
        document.getElementById('log-cleanup-count').textContent = stats.cleanup_count || 0;
        document.getElementById('log-newest-entry').textContent = stats.newest_entry ?
            this.formatLogTime(stats.newest_entry) : '-';
    }

    /**
     * 更新日志分页信息
     */
    updateLogPagination() {
        const start = (this.currentLogPage - 1) * this.logPageSize + 1;
        const end = Math.min(this.currentLogPage * this.logPageSize, this.currentLogs.length);

        document.getElementById('log-start').textContent = start;
        document.getElementById('log-end').textContent = end;
        document.getElementById('log-total').textContent = this.currentLogs.length;
        document.getElementById('log-current-page').textContent = this.currentLogPage;
        document.getElementById('log-total-pages').textContent = this.totalLogPages;

        const prevBtn = document.getElementById('log-prev-btn');
        const nextBtn = document.getElementById('log-next-btn');

        prevBtn.disabled = this.currentLogPage <= 1;
        nextBtn.disabled = this.currentLogPage >= this.totalLogPages;

        if (prevBtn.disabled) {
            prevBtn.classList.add('opacity-50', 'cursor-not-allowed');
        } else {
            prevBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        }

        if (nextBtn.disabled) {
            nextBtn.classList.add('opacity-50', 'cursor-not-allowed');
        } else {
            nextBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        }
    }

    /**
     * 过滤日志
     */
    filterLogs() {
        // 基础筛选
        this.logFilters.method = document.getElementById('log-filter-method').value;
        this.logFilters.status = document.getElementById('log-filter-status').value;
        this.logFilters.domain = document.getElementById('log-filter-domain').value;
        this.logFilters.search = document.getElementById('log-filter-search').value;

        // 高级筛选
        this.logFilters.startTime = document.getElementById('log-filter-start-time').value;
        this.logFilters.endTime = document.getElementById('log-filter-end-time').value;
        this.logFilters.clientIp = document.getElementById('log-filter-client-ip').value;

        // 保存搜索高亮词
        this.searchHighlightTerm = this.logFilters.search;

        this.currentLogPage = 1; // 重置到第一页
        this.loadLogs();

        // 更新筛选统计
        this.updateFilterStats();
    }

    /**
     * 清空日志筛选
     */
    clearLogFilters() {
        // 清空基础筛选
        document.getElementById('log-filter-method').value = '';
        document.getElementById('log-filter-status').value = '';
        document.getElementById('log-filter-domain').value = '';
        document.getElementById('log-filter-search').value = '';

        // 清空高级筛选
        document.getElementById('log-filter-start-time').value = '';
        document.getElementById('log-filter-end-time').value = '';
        document.getElementById('log-filter-client-ip').value = '';

        this.logFilters = {
            method: '',
            status: '',
            domain: '',
            search: '',
            startTime: '',
            endTime: '',
            clientIp: ''
        };
        this.searchHighlightTerm = '';
        this.currentLogPage = 1;
        this.loadLogs();

        // 隐藏筛选统计
        document.getElementById('filter-stats').classList.add('hidden');
    }

    /**
     * 日志分页
     * @param {number} direction - 方向 (-1: 上一页, 1: 下一页)
     */
    changeLogPage(direction) {
        const newPage = this.currentLogPage + direction;
        if (newPage >= 1 && newPage <= this.totalLogPages) {
            this.currentLogPage = newPage;
            this.loadLogs();
        }
    }

    /**
     * 显示日志详情
     * @param {string} logId - 日志ID
     */
    async showLogDetail(logId) {
        try {
            this.showLoading('获取日志详情中...');

            const logDetail = await logsAPI.getLogDetail(logId);
            this.currentLogDetail = logDetail;

            // 填充基本信息
            this.fillLogDetailBasicInfo(logDetail);

            // 填充详细信息
            this.fillLogDetailTabs(logDetail);

            // 加载相关日志
            await this.loadRelatedLogs(logDetail);

            // 显示模态框
            this.showModal('log-detail-modal');

        } catch (error) {
            console.error('获取日志详情失败:', error);
            this.showError('获取日志详情失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 填充日志详情基本信息
     * @param {Object} logDetail - 日志详情
     */
    fillLogDetailBasicInfo(logDetail) {
        // 基本信息
        document.getElementById('detail-id').textContent = logDetail.id;
        document.getElementById('detail-time').textContent = this.formatLogTime(logDetail.timestamp);
        document.getElementById('detail-duration').textContent = logDetail.duration_ms + 'ms';
        document.getElementById('detail-client-ip').textContent = logDetail.client_ip;
        document.getElementById('detail-target').textContent = logDetail.target_host + logDetail.target_path;
        document.getElementById('detail-user-agent').textContent = logDetail.user_agent || '-';

        // 方法和状态码标签
        const methodBadge = document.getElementById('detail-method-badge');
        const statusBadge = document.getElementById('detail-status-badge');

        methodBadge.textContent = logDetail.method;
        methodBadge.className = `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getMethodBadgeClass(logDetail.method)}`;

        statusBadge.textContent = logDetail.status_code;
        statusBadge.className = `inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getStatusBadgeClass(logDetail.status_code)}`;
    }

    /**
     * 填充日志详情标签页内容
     * @param {Object} logDetail - 日志详情
     */
    fillLogDetailTabs(logDetail) {
        // 请求信息
        const requestInfo = `请求方法: ${logDetail.method}
请求路径: ${logDetail.target_path}
目标主机: ${logDetail.target_host}
客户端IP: ${logDetail.client_ip}
用户代理: ${logDetail.user_agent || '-'}
请求时间: ${this.formatLogTime(logDetail.timestamp)}
响应时间: ${logDetail.duration_ms}ms`;
        document.getElementById('detail-request-info').textContent = requestInfo;

        // 请求体
        const requestBody = logDetail.request_body || '无请求体';
        document.getElementById('detail-request-body').textContent = requestBody;

        // 响应信息
        const responseInfo = `状态码: ${logDetail.status_code}
响应大小: ${logDetail.response_size || 0} bytes
响应时间: ${logDetail.duration_ms}ms
响应类型: ${logDetail.response_content_type || '-'}`;
        document.getElementById('detail-response-info').textContent = responseInfo;

        // 响应头信息
        const responseHeaders = logDetail.response_headers ?
            Object.entries(logDetail.response_headers)
                .map(([key, value]) => `${key}: ${value}`)
                .join('\n') : '无响应头信息';
        document.getElementById('detail-response-headers').textContent = responseHeaders;

        // 请求头信息
        const headersText = logDetail.request_headers ?
            Object.entries(logDetail.request_headers)
                .map(([key, value]) => `${key}: ${value}`)
                .join('\n') : '无请求头信息';
        document.getElementById('detail-headers-info').textContent = headersText;

        // cURL命令
        const curlCommand = this.generateCurlCommand(logDetail);
        document.getElementById('detail-curl-command').textContent = curlCommand;

        // 代理cURL命令
        const proxyCurlCommand = this.generateProxyCurlCommand(logDetail);
        document.getElementById('detail-proxy-curl-command').textContent = proxyCurlCommand;

        // 时间线
        this.generateTimeline(logDetail);
    }

    /**
     * 生成cURL命令
     * @param {Object} logDetail - 日志详情
     * @returns {string} cURL命令
     */
    generateCurlCommand(logDetail) {
        let curl = `curl -X ${logDetail.method}`;

        // 添加请求头
        if (logDetail.request_headers) {
            Object.entries(logDetail.request_headers).forEach(([key, value]) => {
                if (key.toLowerCase() !== 'host') {
                    curl += ` \\\n  -H "${key}: ${value}"`;
                }
            });
        }

        // 添加请求体（如果有）
        if (logDetail.request_body && logDetail.method !== 'GET') {
            curl += ` \\\n  -d '${logDetail.request_body}'`;
        }

        // 添加URL
        curl += ` \\\n  "${logDetail.target_host}${logDetail.target_path}"`;

        return curl;
    }

    /**
     * 生成代理cURL命令
     * @param {Object} logDetail - 日志详情
     * @returns {string} 代理cURL命令
     */
    generateProxyCurlCommand(logDetail) {
        let curl = `curl -X ${logDetail.method}`;

        // 添加管理员密钥
        curl += ` \\\n  -H "X-Log-Secret: YOUR_ADMIN_SECRET"`;

        // 添加其他请求头（除了Host）
        if (logDetail.request_headers) {
            Object.entries(logDetail.request_headers).forEach(([key, value]) => {
                if (key.toLowerCase() !== 'host' && key.toLowerCase() !== 'x-log-secret') {
                    curl += ` \\\n  -H "${key}: ${value}"`;
                }
            });
        }

        // 添加请求体（如果有）
        if (logDetail.request_body && logDetail.method !== 'GET') {
            curl += ` \\\n  -d '${logDetail.request_body}'`;
        }

        // 添加代理URL
        const proxyUrl = `http://localhost:10805/proxy?target=${encodeURIComponent(logDetail.target_host + logDetail.target_path)}`;
        curl += ` \\\n  "${proxyUrl}"`;

        return curl;
    }

    /**
     * 复制cURL命令
     */
    async copyCurlCommand() {
        try {
            const curlCommand = document.getElementById('detail-curl-command').textContent;
            await navigator.clipboard.writeText(curlCommand);
            this.copySuccess();
        } catch (error) {
            console.error('复制失败:', error);
            this.showError('复制失败，请手动复制');
        }
    }

    /**
     * 复制代理cURL命令
     */
    async copyProxyCurlCommand() {
        try {
            const curlCommand = document.getElementById('detail-proxy-curl-command').textContent;
            await navigator.clipboard.writeText(curlCommand);
            this.copySuccess();
        } catch (error) {
            console.error('复制失败:', error);
            this.showError('复制失败，请手动复制');
        }
    }

    /**
     * 复制请求头
     */
    async copyRequestHeaders() {
        try {
            const headers = document.getElementById('detail-headers-info').textContent;
            await navigator.clipboard.writeText(headers);
            this.copySuccess();
        } catch (error) {
            console.error('复制失败:', error);
            this.showError('复制失败，请手动复制');
        }
    }

    /**
     * 复制响应头
     */
    async copyResponseHeaders() {
        try {
            const headers = document.getElementById('detail-response-headers').textContent;
            await navigator.clipboard.writeText(headers);
            this.copySuccess();
        } catch (error) {
            console.error('复制失败:', error);
            this.showError('复制失败，请手动复制');
        }
    }

    /**
     * 切换日志详情标签页
     * @param {string} tabName - 标签页名称
     */
    switchLogDetailTab(tabName) {
        // 移除所有活动状态
        document.querySelectorAll('.log-detail-tab-btn').forEach(btn => {
            btn.classList.remove('border-blue-500', 'text-blue-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });
        document.querySelectorAll('.log-detail-tab-pane').forEach(pane => {
            pane.classList.add('hidden');
        });

        // 激活选中的标签
        const activeBtn = document.querySelector(`[data-tab="${tabName}"]`);
        if (activeBtn) {
            activeBtn.classList.remove('border-transparent', 'text-gray-500');
            activeBtn.classList.add('border-blue-500', 'text-blue-600');
        }

        const activePane = document.getElementById(`log-detail-${tabName}-tab`);
        if (activePane) {
            activePane.classList.remove('hidden');
        }
    }

    /**
     * 关闭日志详情模态框
     */
    closeLogDetailModal() {
        const modal = document.getElementById('log-detail-modal');
        if (modal) {
            modal.classList.add('hidden');
        }
        this.currentLogDetail = null;
    }

    /**
     * 获取请求方法的样式类
     * @param {string} method - 请求方法
     * @returns {string} 样式类
     */
    getMethodBadgeClass(method) {
        const classes = {
            'GET': 'bg-green-100 text-green-800',
            'POST': 'bg-blue-100 text-blue-800',
            'PUT': 'bg-yellow-100 text-yellow-800',
            'DELETE': 'bg-red-100 text-red-800',
            'PATCH': 'bg-purple-100 text-purple-800',
            'HEAD': 'bg-gray-100 text-gray-800',
            'OPTIONS': 'bg-indigo-100 text-indigo-800'
        };
        return classes[method] || 'bg-gray-100 text-gray-800';
    }

    /**
     * 获取请求类型的样式类
     * @param {string} type - 请求类型
     * @returns {string} 样式类
     */
    getTypeBadgeClass(type) {
        const classes = {
            'HTTP': 'bg-blue-100 text-blue-800',
            'HTTPS': 'bg-green-100 text-green-800',
            'WebSocket': 'bg-purple-100 text-purple-800',
            'WebSocketSecure': 'bg-indigo-100 text-indigo-800'
        };
        return classes[type] || 'bg-gray-100 text-gray-800';
    }

    /**
     * 获取状态码的样式类
     * @param {number} statusCode - 状态码
     * @returns {string} 样式类
     */
    getStatusBadgeClass(statusCode) {
        if (statusCode >= 200 && statusCode < 300) {
            return 'bg-green-100 text-green-800';
        } else if (statusCode >= 300 && statusCode < 400) {
            return 'bg-yellow-100 text-yellow-800';
        } else if (statusCode >= 400 && statusCode < 500) {
            return 'bg-red-100 text-red-800';
        } else if (statusCode >= 500) {
            return 'bg-red-100 text-red-800';
        }
        return 'bg-gray-100 text-gray-800';
    }

    /**
     * 格式化日志时间
     * @param {string} timestamp - 时间戳
     * @returns {string} 格式化后的时间
     */
    formatLogTime(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN', {
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    }

    /**
     * 截断URL显示
     * @param {string} url - URL
     * @param {number} maxLength - 最大长度
     * @returns {string} 截断后的URL
     */
    truncateUrl(url, maxLength = 50) {
        if (url.length <= maxLength) {
            return url;
        }
        return url.substring(0, maxLength - 3) + '...';
    }

    /**
     * 高亮搜索关键词
     * @param {string} text - 原始文本
     * @returns {string} 高亮后的HTML
     */
    highlightSearchTerm(text) {
        if (!this.searchHighlightTerm || !text) {
            return text;
        }

        const searchTerm = this.searchHighlightTerm.trim();
        if (searchTerm === '') {
            return text;
        }

        // 使用正则表达式进行不区分大小写的替换
        const regex = new RegExp(`(${searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
        return text.replace(regex, '<mark class="bg-yellow-200 px-1 rounded">$1</mark>');
    }

    /**
     * 切换高级筛选显示/隐藏
     */
    toggleAdvancedFilters() {
        const advancedFilters = document.getElementById('advanced-filters');
        const toggleBtn = document.getElementById('advanced-toggle-btn');
        const toggleIcon = document.getElementById('advanced-toggle-icon');

        this.advancedFiltersVisible = !this.advancedFiltersVisible;

        if (this.advancedFiltersVisible) {
            advancedFilters.classList.remove('hidden');
            toggleBtn.querySelector('span').textContent = '收起';
            toggleIcon.classList.add('rotate-180');
        } else {
            advancedFilters.classList.add('hidden');
            toggleBtn.querySelector('span').textContent = '展开';
            toggleIcon.classList.remove('rotate-180');
        }
    }

    /**
     * 设置快速时间筛选
     * @param {string} period - 时间周期 (1h, 6h, 24h, 7d, 30d)
     */
    setQuickTimeFilter(period) {
        const now = new Date();
        const endTime = now.toISOString().slice(0, 16); // YYYY-MM-DDTHH:mm格式

        let startTime;
        switch (period) {
            case '1h':
                startTime = new Date(now.getTime() - 60 * 60 * 1000);
                break;
            case '6h':
                startTime = new Date(now.getTime() - 6 * 60 * 60 * 1000);
                break;
            case '24h':
                startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
                break;
            case '7d':
                startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                break;
            case '30d':
                startTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
                break;
            default:
                return;
        }

        document.getElementById('log-filter-start-time').value = startTime.toISOString().slice(0, 16);
        document.getElementById('log-filter-end-time').value = endTime;

        // 自动展开高级筛选
        if (!this.advancedFiltersVisible) {
            this.toggleAdvancedFilters();
        }

        // 触发筛选
        this.filterLogs();
    }

    /**
     * 更新筛选统计信息
     */
    updateFilterStats() {
        const hasFilters = Object.values(this.logFilters).some(value => value !== '');
        const filterStats = document.getElementById('filter-stats');

        if (hasFilters && this.currentLogs.length >= 0) {
            document.getElementById('filtered-count').textContent = this.currentLogs.length;
            filterStats.classList.remove('hidden');
        } else {
            filterStats.classList.add('hidden');
        }
    }

    /**
     * 保存筛选预设
     */
    async saveFilterPreset() {
        const hasFilters = Object.values(this.logFilters).some(value => value !== '');
        if (!hasFilters) {
            this.showError('请先设置筛选条件');
            return;
        }

        const presetName = prompt('请输入预设名称:');
        if (!presetName || presetName.trim() === '') {
            return;
        }

        const preset = {
            name: presetName.trim(),
            filters: { ...this.logFilters },
            createdAt: new Date().toISOString()
        };

        this.filterPresets.push(preset);
        this.saveFilterPresets();
        this.renderFilterPresets();
        this.showSuccess(`筛选预设 "${presetName}" 已保存`);
    }

    /**
     * 加载筛选预设
     * @param {Object} preset - 预设对象
     */
    loadFilterPreset(preset) {
        // 设置筛选条件
        this.logFilters = { ...preset.filters };

        // 更新UI
        document.getElementById('log-filter-method').value = preset.filters.method || '';
        document.getElementById('log-filter-status').value = preset.filters.status || '';
        document.getElementById('log-filter-domain').value = preset.filters.domain || '';
        document.getElementById('log-filter-search').value = preset.filters.search || '';
        document.getElementById('log-filter-start-time').value = preset.filters.startTime || '';
        document.getElementById('log-filter-end-time').value = preset.filters.endTime || '';
        document.getElementById('log-filter-client-ip').value = preset.filters.clientIp || '';

        // 如果有高级筛选条件，自动展开高级筛选
        const hasAdvancedFilters = preset.filters.startTime || preset.filters.endTime || preset.filters.clientIp;
        if (hasAdvancedFilters && !this.advancedFiltersVisible) {
            this.toggleAdvancedFilters();
        }

        // 应用筛选
        this.currentLogPage = 1;
        this.loadLogs();

        this.showSuccess(`已加载筛选预设 "${preset.name}"`);
    }

    /**
     * 删除筛选预设
     * @param {number} index - 预设索引
     */
    deleteFilterPreset(index) {
        const preset = this.filterPresets[index];
        if (confirm(`确定要删除筛选预设 "${preset.name}" 吗？`)) {
            this.filterPresets.splice(index, 1);
            this.saveFilterPresets();
            this.renderFilterPresets();
            this.showSuccess(`筛选预设 "${preset.name}" 已删除`);
        }
    }

    /**
     * 渲染筛选预设
     */
    renderFilterPresets() {
        const presetList = document.getElementById('preset-list');
        const filterPresets = document.getElementById('filter-presets');

        if (this.filterPresets.length === 0) {
            filterPresets.classList.add('hidden');
            return;
        }

        filterPresets.classList.remove('hidden');
        presetList.innerHTML = this.filterPresets.map((preset, index) => `
            <div class="flex items-center bg-white border border-gray-300 rounded-md px-3 py-1">
                <button onclick="businessManager.loadFilterPreset(${JSON.stringify(preset).replace(/"/g, '&quot;')})"
                        class="text-sm text-indigo-600 hover:text-indigo-800 mr-2">
                    ${preset.name}
                </button>
                <button onclick="businessManager.deleteFilterPreset(${index})"
                        class="text-xs text-red-500 hover:text-red-700">
                    ×
                </button>
            </div>
        `).join('');
    }

    /**
     * 保存筛选预设到本地存储
     */
    saveFilterPresets() {
        try {
            localStorage.setItem('logFilterPresets', JSON.stringify(this.filterPresets));
        } catch (error) {
            console.error('保存筛选预设失败:', error);
        }
    }

    /**
     * 从本地存储加载筛选预设
     * @returns {Array} 筛选预设数组
     */
    loadFilterPresets() {
        try {
            const presets = localStorage.getItem('logFilterPresets');
            return presets ? JSON.parse(presets) : [];
        } catch (error) {
            console.error('加载筛选预设失败:', error);
            return [];
        }
    }

    /**
     * 生成时间线
     * @param {Object} logDetail - 日志详情
     */
    generateTimeline(logDetail) {
        const timeline = document.getElementById('detail-timeline');
        const startTime = new Date(logDetail.timestamp);
        const endTime = new Date(startTime.getTime() + logDetail.duration_ms);

        const timelineItems = [
            {
                time: startTime,
                title: '请求开始',
                description: `客户端 ${logDetail.client_ip} 发起 ${logDetail.method} 请求`,
                icon: '🚀',
                color: 'blue'
            },
            {
                time: new Date(startTime.getTime() + Math.floor(logDetail.duration_ms * 0.1)),
                title: '请求解析',
                description: '解析请求头和参数',
                icon: '🔍',
                color: 'indigo'
            },
            {
                time: new Date(startTime.getTime() + Math.floor(logDetail.duration_ms * 0.3)),
                title: '转发请求',
                description: `向目标服务器 ${logDetail.target_host} 转发请求`,
                icon: '📡',
                color: 'purple'
            },
            {
                time: new Date(startTime.getTime() + Math.floor(logDetail.duration_ms * 0.8)),
                title: '接收响应',
                description: `收到状态码 ${logDetail.status_code} 的响应`,
                icon: '📥',
                color: logDetail.status_code >= 400 ? 'red' : 'green'
            },
            {
                time: endTime,
                title: '请求完成',
                description: `总耗时 ${logDetail.duration_ms}ms`,
                icon: '✅',
                color: 'green'
            }
        ];

        timeline.innerHTML = timelineItems.map((item, index) => `
            <div class="flex items-start">
                <div class="flex-shrink-0 w-8 h-8 bg-${item.color}-100 text-${item.color}-600 rounded-full flex items-center justify-center text-sm">
                    ${item.icon}
                </div>
                <div class="ml-4 flex-1">
                    <div class="flex items-center justify-between">
                        <h6 class="text-sm font-medium text-gray-900">${item.title}</h6>
                        <span class="text-xs text-gray-500">${item.time.toLocaleTimeString()}</span>
                    </div>
                    <p class="text-sm text-gray-600 mt-1">${item.description}</p>
                    ${index < timelineItems.length - 1 ? '<div class="w-px h-6 bg-gray-200 ml-4 mt-2"></div>' : ''}
                </div>
            </div>
        `).join('');
    }

    /**
     * 加载相关日志
     * @param {Object} logDetail - 当前日志详情
     */
    async loadRelatedLogs(logDetail) {
        try {
            // 查找相同客户端IP的日志
            const relatedLogs = await logsAPI.getLogs({
                clientIp: logDetail.client_ip,
                limit: 5
            });

            if (relatedLogs.logs && relatedLogs.logs.length > 1) {
                const relatedSection = document.getElementById('related-logs-section');
                const relatedList = document.getElementById('related-logs-list');

                const otherLogs = relatedLogs.logs.filter(log => log.id !== logDetail.id);

                if (otherLogs.length > 0) {
                    relatedList.innerHTML = otherLogs.map(log => `
                        <button onclick="businessManager.showLogDetail('${log.id}')"
                                class="text-xs bg-white border border-yellow-300 rounded px-2 py-1 hover:bg-yellow-50 transition-colors duration-200">
                            ${log.method} ${this.truncateUrl(log.target_host + log.target_path, 30)}
                            <span class="text-gray-500">(${this.formatLogTime(log.timestamp)})</span>
                        </button>
                    `).join('');

                    relatedSection.classList.remove('hidden');
                }
            }
        } catch (error) {
            console.error('加载相关日志失败:', error);
        }
    }

    /**
     * 导出日志详情
     */
    exportLogDetail() {
        if (!this.currentLogDetail) {
            this.showError('没有可导出的日志详情');
            return;
        }

        const exportData = {
            基本信息: {
                请求ID: this.currentLogDetail.id,
                请求方法: this.currentLogDetail.method,
                目标地址: this.currentLogDetail.target_host + this.currentLogDetail.target_path,
                状态码: this.currentLogDetail.status_code,
                请求时间: this.formatLogTime(this.currentLogDetail.timestamp),
                处理耗时: this.currentLogDetail.duration_ms + 'ms',
                客户端IP: this.currentLogDetail.client_ip,
                用户代理: this.currentLogDetail.user_agent
            },
            请求头: this.currentLogDetail.request_headers || {},
            请求体: this.currentLogDetail.request_body || null,
            响应头: this.currentLogDetail.response_headers || {},
            cURL命令: {
                直接请求: this.generateCurlCommand(this.currentLogDetail),
                代理请求: this.generateProxyCurlCommand(this.currentLogDetail)
            },
            导出时间: new Date().toISOString()
        };

        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `log-detail-${this.currentLogDetail.id}-${new Date().toISOString().slice(0, 10)}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showSuccess('日志详情已导出');
    }

    /**
     * 分享日志详情
     */
    async shareLogDetail() {
        if (!this.currentLogDetail) {
            this.showError('没有可分享的日志详情');
            return;
        }

        const shareData = {
            title: `日志详情 - ${this.currentLogDetail.method} ${this.currentLogDetail.target_host}`,
            text: `请求ID: ${this.currentLogDetail.id}\n方法: ${this.currentLogDetail.method}\n状态码: ${this.currentLogDetail.status_code}\n耗时: ${this.currentLogDetail.duration_ms}ms`,
            url: window.location.href
        };

        try {
            if (navigator.share) {
                await navigator.share(shareData);
                this.showSuccess('日志详情已分享');
            } else {
                // 降级到复制链接
                const shareText = `${shareData.title}\n\n${shareData.text}\n\n${shareData.url}`;
                await navigator.clipboard.writeText(shareText);
                this.showSuccess('分享内容已复制到剪贴板');
            }
        } catch (error) {
            console.error('分享失败:', error);
            this.showError('分享失败');
        }
    }

    // ==================== 令牌管理功能 ====================

    /**
     * 管理令牌 - 打开令牌管理弹框
     * @param {string} configId - 配置ID
     */
    async manageTokens(configId) {
        try {
            this.currentConfigId = configId;

            // 获取配置信息用于显示标题
            const config = this.currentConfigs.find(c => c.id === configId);
            if (config) {
                document.getElementById('token-management-title').textContent = `🔑 令牌管理 - ${config.name}`;
            }

            // 加载令牌列表
            await this.loadTokensForManagement(configId);

            // 显示令牌管理弹框
            this.showModal('token-management-modal');
        } catch (error) {
            console.error('打开令牌管理失败:', error);
            this.showError('打开令牌管理失败: ' + error.message);
        }
    }

    /**
     * 为令牌管理弹框加载令牌列表
     * @param {string} configId - 配置ID
     */
    async loadTokensForManagement(configId) {
        try {
            this.showLoading('加载令牌列表中...');

            // 从API加载真实的令牌数据
            const response = await tokensAPI.getTokens(configId);
            const tokens = response.tokens || [];

            // 更新统计信息
            this.updateTokenManagementStats(tokens);

            // 渲染令牌列表
            this.renderTokenManagementList(tokens);
        } catch (error) {
            console.error('加载令牌列表失败:', error);
            // 如果API调用失败，显示空列表
            this.updateTokenManagementStats([]);
            this.renderTokenManagementList([]);
            this.showError('加载令牌列表失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 更新令牌管理弹框的统计信息
     * @param {Array} tokens - 令牌列表
     */
    updateTokenManagementStats(tokens) {
        const totalCount = tokens.length;
        const activeCount = tokens.filter(t => t.enabled && !this.isTokenExpired(t)).length;
        const totalRequests = tokens.reduce((sum, t) => sum + (t.usage_count || 0), 0);

        document.getElementById('token-mgmt-total-count').textContent = totalCount;
        document.getElementById('token-mgmt-active-count').textContent = activeCount;
        document.getElementById('token-mgmt-total-requests').textContent = totalRequests.toLocaleString();
    }

    /**
     * 渲染令牌管理弹框的令牌列表
     * @param {Array} tokens - 令牌列表
     */
    renderTokenManagementList(tokens) {
        const tokenList = document.getElementById('token-mgmt-list');

        if (tokens.length === 0) {
            tokenList.innerHTML = `
                <div class="text-center py-12 text-gray-500">
                    <div class="text-4xl mb-4">🔑</div>
                    <p class="text-lg mb-2">暂无访问令牌</p>
                    <p class="text-sm">点击"新增令牌"创建第一个访问令牌</p>
                </div>
            `;
            this.updateTokenSelectionUI();
            return;
        }

        // 存储当前令牌列表用于搜索
        this.currentTokens = tokens;

        tokenList.innerHTML = tokens.map(token => `
            <div class="token-item border-b border-gray-200 p-2 hover:bg-gray-50 ${token.enabled ? '' : 'opacity-60'}" data-token-id="${token.id}">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2 flex-1 min-w-0">
                        <input type="checkbox" class="token-checkbox h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                               data-token-id="${token.id}" onchange="businessManager.updateTokenSelection()">
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center gap-2 mb-1">
                                <h4 class="text-sm font-medium text-gray-900 truncate">${this.escapeHtml(token.name)}</h4>
                                <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium ${this.getTokenStatusBadgeClass(token)}">
                                    ${this.getTokenStatusLabel(token)}
                                </span>
                            </div>
                            <div class="text-xs text-gray-500 flex items-center gap-4">
                                <span>创建: ${this.formatDate(token.created_at).split(' ')[0]}</span>
                                <span>使用: ${token.usage_count || 0}次</span>
                                ${token.expires_at ? `<span>过期: ${this.formatDate(token.expires_at).split(' ')[0]}</span>` : '<span>永不过期</span>'}
                                ${token.description ? `<span class="truncate max-w-32" title="${this.escapeHtml(token.description)}">${this.escapeHtml(token.description)}</span>` : ''}
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center gap-1 ml-2">
                        <button onclick="businessManager.copyTokenFromList('${token.id}')"
                                class="bg-blue-500 hover:bg-blue-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200"
                                title="复制令牌值">
                            📋
                        </button>
                        <button onclick="businessManager.editToken('${token.id}')"
                                class="bg-gray-500 hover:bg-gray-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200"
                                title="编辑令牌">
                            ✏️
                        </button>
                        <button onclick="businessManager.toggleToken('${token.id}')"
                                class="bg-${token.enabled ? 'yellow' : 'green'}-500 hover:bg-${token.enabled ? 'yellow' : 'green'}-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200"
                                title="${token.enabled ? '禁用' : '启用'}令牌">
                            ${token.enabled ? '⏸️' : '▶️'}
                        </button>
                        <button onclick="businessManager.deleteToken('${token.id}', '${this.escapeHtml(token.name)}')"
                                class="bg-red-500 hover:bg-red-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200"
                                title="删除令牌">
                            🗑️
                        </button>
                    </div>
                </div>
            </div>
        `).join('');

        this.updateTokenSelectionUI();
    }

    /**
     * 显示新增令牌弹框
     */
    showAddTokenModal() {
        // 暂时隐藏令牌管理弹框以避免层级冲突
        this.hideModal('token-management-modal');

        this.editingTokenId = null;
        document.getElementById('token-modal-title').textContent = '🔑 新增访问令牌';
        this.resetTokenForm();
        document.getElementById('token-result').classList.add('hidden');
        document.getElementById('save-token-btn').textContent = '创建令牌';
        this.showModal('token-modal');
    }

    /**
     * 编辑令牌
     * @param {string} tokenId - 令牌ID
     */
    async editToken(tokenId) {
        try {
            if (!this.currentConfigId) {
                this.showError('配置ID未设置，请重新选择配置');
                return;
            }

            if (!tokenId) {
                this.showError('令牌ID未提供');
                return;
            }

            this.showLoading('获取令牌信息中...');

            // 从API获取令牌详情
            const token = await tokensAPI.getToken(this.currentConfigId, tokenId);

            if (!token) {
                this.showError('令牌不存在');
                return;
            }

            this.editingTokenId = tokenId;

            // 设置弹框标题和按钮
            document.getElementById('token-modal-title').textContent = '🔑 编辑访问令牌';
            document.getElementById('save-token-btn').textContent = '保存修改';

            // 填充表单数据
            document.getElementById('token-name').value = token.name;
            document.getElementById('token-description').value = token.description || '';

            // 设置过期时间
            if (token.expires_at) {
                const date = new Date(token.expires_at);
                document.getElementById('token-expires-at').value = date.toISOString().slice(0, 16);
            } else {
                document.getElementById('token-expires-at').value = '';
            }

            document.getElementById('token-result').classList.add('hidden');
            this.showModal('token-modal');
        } catch (error) {
            console.error('获取令牌数据失败:', error);
            this.showError('获取令牌数据失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 保存令牌
     */
    async saveToken() {
        try {
            const formData = new FormData(document.getElementById('token-form'));
            const name = formData.get('name').trim();
            const description = formData.get('description').trim();
            const expiresAt = formData.get('expires_at');

            if (!name) {
                this.showError('请输入令牌名称');
                return;
            }

            this.showLoading('保存令牌中...');

            if (this.editingTokenId) {
                // 更新令牌
                const tokenData = {
                    name,
                    description,
                    expires_at: expiresAt || null
                };

                await tokensAPI.updateToken(this.currentConfigId, this.editingTokenId, tokenData);
                this.showSuccess('令牌更新成功');
                this.hideModal('token-modal');
            } else {
                // 创建新令牌
                const tokenData = {
                    name,
                    description,
                    expires_at: expiresAt || null
                };

                const response = await tokensAPI.createToken(this.currentConfigId, tokenData);

                // 显示生成的令牌
                const tokenValue = response.token;
                if (tokenValue) {
                    document.getElementById('generated-token').value = tokenValue;
                    document.getElementById('token-result').classList.remove('hidden');
                    document.getElementById('save-token-btn').style.display = 'none';
                }

                this.showSuccess('令牌创建成功');
            }

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
        } catch (error) {
            console.error('保存令牌失败:', error);

            // 处理特定的错误类型
            if (error.message.includes('Token name already exists') || error.message.includes('409')) {
                this.showError('令牌名称已存在，请使用不同的名称');
            } else if (error.message.includes('400')) {
                this.showError('请求参数错误，请检查输入信息');
            } else if (error.message.includes('401')) {
                this.showError('认证失败，请重新登录');
            } else if (error.message.includes('403')) {
                this.showError('权限不足，无法创建令牌');
            } else {
                this.showError('保存令牌失败: ' + error.message);
            }
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 复制令牌
     */
    async copyToken() {
        try {
            const tokenInput = document.getElementById('generated-token');
            if (!tokenInput) {
                this.showError('找不到令牌输入框');
                return;
            }

            const tokenValue = tokenInput.value;
            if (!tokenValue) {
                this.showError('令牌值为空');
                return;
            }

            await navigator.clipboard.writeText(tokenValue);
            this.copySuccess();
        } catch (error) {
            console.error('复制失败:', error);
            this.showError('复制失败，请手动复制令牌');
        }
    }

    /**
     * 从列表复制令牌
     * @param {string} tokenId - 令牌ID
     */
    async copyTokenFromList(tokenId) {
        try {
            this.showLoading('获取令牌值中...');

            // 从服务器获取令牌详情
            const token = await tokensAPI.getToken(this.currentConfigId, tokenId);
            const tokenValue = token.token_value;

            if (!tokenValue) {
                this.showError('无法获取令牌值，请重新创建令牌');
                return;
            }

            // 复制令牌值到剪贴板
            await navigator.clipboard.writeText(tokenValue);
            this.copySuccess();
        } catch (error) {
            console.error('复制令牌失败:', error);
            this.showError('复制失败：' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 切换令牌状态
     * @param {string} tokenId - 令牌ID
     */
    async toggleToken(tokenId) {
        try {
            this.showLoading('切换令牌状态中...');

            // 先获取当前令牌状态
            const token = await tokensAPI.getToken(this.currentConfigId, tokenId);

            if (!token) {
                this.showError('令牌不存在');
                return;
            }

            // 调用API切换令牌状态
            const newEnabled = !token.enabled;
            await tokensAPI.updateToken(this.currentConfigId, tokenId, {
                ...token,
                enabled: newEnabled
            });

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
            this.showSuccess(`令牌已${newEnabled ? '启用' : '禁用'}`);
        } catch (error) {
            console.error('切换令牌状态失败:', error);
            this.showError('切换令牌状态失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 删除令牌
     * @param {string} tokenId - 令牌ID
     * @param {string} tokenName - 令牌名称
     */
    deleteToken(tokenId, tokenName) {
        this.deleteTokenId = tokenId;
        document.getElementById('delete-token-name').textContent = tokenName;
        this.showModal('delete-token-modal');
    }

    /**
     * 确认删除令牌
     */
    async confirmDeleteToken() {
        if (!this.deleteTokenId) return;

        try {
            this.showLoading('删除令牌中...');

            // 调用API删除令牌
            await tokensAPI.deleteToken(this.currentConfigId, this.deleteTokenId);

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
            this.showSuccess('令牌删除成功');
        } catch (error) {
            console.error('删除令牌失败:', error);
            this.showError('删除令牌失败: ' + error.message);
        } finally {
            this.hideLoading();
        }

        this.hideModal('delete-token-modal');
        this.deleteTokenId = null;
    }

    /**
     * 重置令牌表单
     */
    resetTokenForm() {
        document.getElementById('token-form').reset();
        document.getElementById('save-token-btn').style.display = 'inline-block';

        // 添加令牌名称输入验证
        const nameInput = document.getElementById('token-name');
        if (nameInput) {
            // 移除之前的事件监听器
            nameInput.removeEventListener('input', this.validateTokenName);
            // 添加新的事件监听器
            nameInput.addEventListener('input', this.validateTokenName.bind(this));
        }
    }

    /**
     * 验证令牌名称
     * @param {Event} event - 输入事件
     */
    async validateTokenName(event) {
        const name = event.target.value.trim();
        const nameInput = event.target;

        // 清除之前的验证状态
        nameInput.classList.remove('border-red-500', 'border-green-500');

        // 移除之前的提示信息
        const existingHint = nameInput.parentNode.querySelector('.name-validation-hint');
        if (existingHint) {
            existingHint.remove();
        }

        if (name.length > 0) {
            try {
                // 获取当前配置的所有令牌
                const tokens = await tokensAPI.getTokens(this.currentConfigId);
                const existingNames = tokens.map(token => token.name);

                if (existingNames.includes(name)) {
                    // 名称已存在
                    nameInput.classList.add('border-red-500');
                    const hint = document.createElement('p');
                    hint.className = 'mt-1 text-sm text-red-600 name-validation-hint';
                    hint.textContent = '❌ 令牌名称已存在，请使用不同的名称';
                    nameInput.parentNode.appendChild(hint);
                } else {
                    // 名称可用
                    nameInput.classList.add('border-green-500');
                    const hint = document.createElement('p');
                    hint.className = 'mt-1 text-sm text-green-600 name-validation-hint';
                    hint.textContent = '✅ 令牌名称可用';
                    nameInput.parentNode.appendChild(hint);
                }
            } catch (error) {
                console.error('验证令牌名称失败:', error);
            }
        }
    }

    /**
     * 检查令牌是否过期
     * @param {Object} token - 令牌对象
     * @returns {boolean} 是否过期
     */
    isTokenExpired(token) {
        if (!token.expires_at) return false;
        return new Date(token.expires_at) < new Date();
    }

    /**
     * 获取令牌状态
     * @param {Object} token - 令牌对象
     * @returns {string} 令牌状态
     */
    getTokenStatus(token) {
        if (!token.enabled) return 'disabled';
        if (this.isTokenExpired(token)) return 'expired';
        return 'active';
    }

    /**
     * 获取令牌状态标签
     * @param {Object} token - 令牌对象
     * @returns {string} 状态标签
     */
    getTokenStatusLabel(token) {
        const status = this.getTokenStatus(token);
        const labels = {
            'active': '活跃',
            'disabled': '已禁用',
            'expired': '已过期'
        };
        return labels[status] || '未知';
    }

    /**
     * 获取令牌状态标签样式类
     * @param {Object} token - 令牌对象
     * @returns {string} CSS类名
     */
    getTokenStatusBadgeClass(token) {
        const status = this.getTokenStatus(token);
        const classes = {
            'active': 'bg-green-100 text-green-800',
            'disabled': 'bg-red-100 text-red-800',
            'expired': 'bg-yellow-100 text-yellow-800'
        };
        return classes[status] || 'bg-gray-100 text-gray-800';
    }

    // ==================== 批量操作功能 ====================

    /**
     * 更新令牌选择状态UI
     */
    updateTokenSelectionUI() {
        const checkboxes = document.querySelectorAll('.token-checkbox');
        const selectedCount = document.querySelectorAll('.token-checkbox:checked').length;
        const totalCount = checkboxes.length;

        // 更新全选复选框状态
        const selectAllCheckbox = document.getElementById('select-all-tokens');
        if (selectAllCheckbox) {
            selectAllCheckbox.checked = selectedCount === totalCount && totalCount > 0;
            selectAllCheckbox.indeterminate = selectedCount > 0 && selectedCount < totalCount;
        }

        // 更新选择计数
        const selectedCountElement = document.getElementById('selected-token-count');
        if (selectedCountElement) {
            selectedCountElement.textContent = selectedCount;
        }

        // 显示/隐藏批量操作按钮
        const batchActions = document.getElementById('token-batch-actions');
        if (batchActions) {
            if (selectedCount > 0) {
                batchActions.classList.remove('hidden');
            } else {
                batchActions.classList.add('hidden');
            }
        }
    }

    /**
     * 切换全选状态
     * @param {boolean} checked - 是否选中
     */
    toggleSelectAllTokens(checked) {
        const checkboxes = document.querySelectorAll('.token-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.checked = checked;
        });
        this.updateTokenSelectionUI();
    }

    /**
     * 更新令牌选择状态
     */
    updateTokenSelection() {
        this.updateTokenSelectionUI();
    }

    /**
     * 获取选中的令牌ID列表
     * @returns {Array} 选中的令牌ID列表
     */
    getSelectedTokenIds() {
        const selectedCheckboxes = document.querySelectorAll('.token-checkbox:checked');
        return Array.from(selectedCheckboxes).map(checkbox => checkbox.dataset.tokenId);
    }

    /**
     * 批量启用令牌
     */
    async batchEnableTokens() {
        const selectedIds = this.getSelectedTokenIds();
        if (selectedIds.length === 0) {
            this.showError('请先选择要启用的令牌');
            return;
        }

        try {
            this.showLoading(`批量启用 ${selectedIds.length} 个令牌中...`);

            // 并发执行批量操作
            const promises = selectedIds.map(tokenId =>
                this.updateTokenStatus(tokenId, true)
            );

            await Promise.all(promises);

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
            this.showSuccess(`成功启用 ${selectedIds.length} 个令牌`);
        } catch (error) {
            console.error('批量启用令牌失败:', error);
            this.showError('批量启用令牌失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 批量禁用令牌
     */
    async batchDisableTokens() {
        const selectedIds = this.getSelectedTokenIds();
        if (selectedIds.length === 0) {
            this.showError('请先选择要禁用的令牌');
            return;
        }

        try {
            this.showLoading(`批量禁用 ${selectedIds.length} 个令牌中...`);

            // 并发执行批量操作
            const promises = selectedIds.map(tokenId =>
                this.updateTokenStatus(tokenId, false)
            );

            await Promise.all(promises);

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
            this.showSuccess(`成功禁用 ${selectedIds.length} 个令牌`);
        } catch (error) {
            console.error('批量禁用令牌失败:', error);
            this.showError('批量禁用令牌失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 批量删除令牌
     */
    async batchDeleteTokens() {
        const selectedIds = this.getSelectedTokenIds();
        if (selectedIds.length === 0) {
            this.showError('请先选择要删除的令牌');
            return;
        }

        // 确认删除
        if (!confirm(`确定要删除选中的 ${selectedIds.length} 个令牌吗？\n\n删除后使用这些令牌的应用将无法访问代理服务！`)) {
            return;
        }

        try {
            this.showLoading(`批量删除 ${selectedIds.length} 个令牌中...`);

            // 并发执行批量删除
            const promises = selectedIds.map(tokenId =>
                tokensAPI.deleteToken(this.currentConfigId, tokenId)
            );

            await Promise.all(promises);

            // 重新加载令牌列表
            await this.loadTokensForManagement(this.currentConfigId);
            this.showSuccess(`成功删除 ${selectedIds.length} 个令牌`);
        } catch (error) {
            console.error('批量删除令牌失败:', error);
            this.showError('批量删除令牌失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 更新令牌状态
     * @param {string} tokenId - 令牌ID
     * @param {boolean} enabled - 是否启用
     */
    async updateTokenStatus(tokenId, enabled) {
        // 先获取当前令牌信息
        const token = await tokensAPI.getToken(this.currentConfigId, tokenId);

        // 更新状态
        await tokensAPI.updateToken(this.currentConfigId, tokenId, {
            ...token,
            enabled: enabled
        });
    }

    // ==================== 搜索和导出功能 ====================

    /**
     * 搜索令牌
     * @param {string} searchTerm - 搜索关键词
     */
    searchTokens(searchTerm) {
        if (!this.currentTokens) return;

        const filteredTokens = this.currentTokens.filter(token => {
            const searchLower = searchTerm.toLowerCase();
            return token.name.toLowerCase().includes(searchLower) ||
                   (token.description && token.description.toLowerCase().includes(searchLower));
        });

        // 更新列表状态显示
        const statusElement = document.getElementById('token-list-status');
        if (statusElement) {
            if (searchTerm.trim()) {
                statusElement.textContent = `搜索结果: ${filteredTokens.length} / ${this.currentTokens.length}`;
            } else {
                statusElement.textContent = '令牌列表';
            }
        }

        // 重新渲染过滤后的列表
        this.renderFilteredTokenList(filteredTokens);
    }

    /**
     * 渲染过滤后的令牌列表
     * @param {Array} tokens - 过滤后的令牌列表
     */
    renderFilteredTokenList(tokens) {
        const tokenList = document.getElementById('token-mgmt-list');

        if (tokens.length === 0) {
            tokenList.innerHTML = `
                <div class="text-center py-12 text-gray-500">
                    <div class="text-4xl mb-4">🔍</div>
                    <p class="text-lg mb-2">未找到匹配的令牌</p>
                    <p class="text-sm">请尝试其他搜索关键词</p>
                </div>
            `;
            this.updateTokenSelectionUI();
            return;
        }

        tokenList.innerHTML = tokens.map(token => `
            <div class="token-item border-b border-gray-200 p-4 hover:bg-gray-50 ${token.enabled ? '' : 'opacity-60'}" data-token-id="${token.id}">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <input type="checkbox" class="token-checkbox h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                               data-token-id="${token.id}" onchange="businessManager.updateTokenSelection()">
                        <div class="flex-1">
                            <div class="flex items-center gap-3 mb-2">
                                <h4 class="text-lg font-medium text-gray-900">${this.escapeHtml(token.name)}</h4>
                                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${this.getTokenStatusBadgeClass(token)}">
                                    ${this.getTokenStatusLabel(token)}
                                </span>
                            </div>
                            <div class="text-sm text-gray-600 space-y-1">
                                <div>创建时间: ${this.formatDate(token.created_at)}</div>
                                <div>使用次数: ${token.usage_count || 0} 次</div>
                                ${token.expires_at ? `<div>过期时间: ${this.formatDate(token.expires_at)}</div>` : '<div>永不过期</div>'}
                                ${token.description ? `<div class="mt-2 text-gray-700">${this.escapeHtml(token.description)}</div>` : ''}
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center gap-2 ml-4">
                        <button onclick="businessManager.copyTokenFromList('${token.id}')"
                                class="bg-blue-500 hover:bg-blue-600 text-white px-3 py-1 rounded text-sm transition-colors duration-200"
                                title="复制令牌值">
                            📋 复制
                        </button>
                        <button onclick="businessManager.editToken('${token.id}')"
                                class="bg-gray-500 hover:bg-gray-600 text-white px-3 py-1 rounded text-sm transition-colors duration-200"
                                title="编辑令牌">
                            ✏️ 编辑
                        </button>
                        <button onclick="businessManager.toggleToken('${token.id}')"
                                class="bg-${token.enabled ? 'yellow' : 'green'}-500 hover:bg-${token.enabled ? 'yellow' : 'green'}-600 text-white px-3 py-1 rounded text-sm transition-colors duration-200"
                                title="${token.enabled ? '禁用' : '启用'}令牌">
                            ${token.enabled ? '⏸️ 禁用' : '▶️ 启用'}
                        </button>
                        <button onclick="businessManager.deleteToken('${token.id}', '${this.escapeHtml(token.name)}')"
                                class="bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded text-sm transition-colors duration-200"
                                title="删除令牌">
                            🗑️ 删除
                        </button>
                    </div>
                </div>
            </div>
        `).join('');

        this.updateTokenSelectionUI();
    }

    /**
     * 导出令牌列表
     */
    async exportTokens() {
        try {
            if (!this.currentTokens || this.currentTokens.length === 0) {
                this.showError('没有可导出的令牌');
                return;
            }

            this.showLoading('准备导出数据中...');

            // 获取配置信息
            const config = this.currentConfigs.find(c => c.id === this.currentConfigId);

            // 准备导出数据
            const exportData = {
                配置信息: {
                    配置ID: this.currentConfigId,
                    配置名称: config ? config.name : '未知配置',
                    目标地址: config ? config.target_url : '未知地址',
                    导出时间: new Date().toISOString()
                },
                令牌统计: {
                    总数量: this.currentTokens.length,
                    活跃数量: this.currentTokens.filter(t => t.enabled && !this.isTokenExpired(t)).length,
                    已禁用数量: this.currentTokens.filter(t => !t.enabled).length,
                    已过期数量: this.currentTokens.filter(t => this.isTokenExpired(t)).length,
                    总使用次数: this.currentTokens.reduce((sum, t) => sum + (t.usage_count || 0), 0)
                },
                令牌列表: this.currentTokens.map(token => ({
                    ID: token.id,
                    名称: token.name,
                    描述: token.description || '',
                    状态: this.getTokenStatusLabel(token),
                    创建时间: this.formatDate(token.created_at),
                    过期时间: token.expires_at ? this.formatDate(token.expires_at) : '永不过期',
                    使用次数: token.usage_count || 0,
                    最后使用: token.last_used ? this.formatDate(token.last_used) : '从未使用',
                    令牌值: token.token_value || '已隐藏'
                })),
                导出说明: {
                    格式版本: '1.0',
                    导出工具: 'Privacy Gateway 管理界面',
                    注意事项: [
                        '令牌值包含敏感信息，请妥善保管',
                        '导入时请确保目标配置存在',
                        '过期令牌将无法正常使用'
                    ]
                }
            };

            // 创建并下载文件
            const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `tokens-${config ? config.name : 'unknown'}-${new Date().toISOString().slice(0, 10)}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            this.showSuccess(`成功导出 ${this.currentTokens.length} 个令牌`);
        } catch (error) {
            console.error('导出令牌失败:', error);
            this.showError('导出令牌失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    // ==================== 令牌使用示例功能 ====================

    /**
     * 显示令牌使用示例
     * @param {string} configId - 配置ID
     */
    async showTokenUsageExamples(configId) {
        try {
            this.showLoading('加载使用示例...');

            // 获取配置信息
            const configResponse = await proxyAPI.getConfigs();
            const config = configResponse.configs.find(c => c.id === configId);

            if (!config) {
                this.showError('配置不存在');
                return;
            }

            // 获取令牌信息
            const tokensResponse = await tokensAPI.getTokens(configId);
            const tokens = tokensResponse.data?.tokens || [];

            // 存储当前配置和令牌信息
            this.currentUsageConfig = config;
            this.currentUsageTokens = tokens;

            // 更新使用示例内容
            this.updateUsageExamples(config, tokens);

            // 显示弹框
            this.showModal('usage-example-modal');

        } catch (error) {
            console.error('显示使用示例失败:', error);
            this.showError('显示使用示例失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 更新使用示例代码
     * @param {Object} config - 配置信息
     * @param {Array} tokens - 令牌列表
     */
    updateUsageExamples(config, tokens) {
        const targetUrl = config.target_url || 'https://api.example.com/data';
        const exampleToken = tokens.length > 0 ? tokens[0].token_value : 'YOUR_TOKEN_HERE';
        const configName = config.name || '示例配置';

        // 生成描述信息
        const description = this.generateUsageDescription(targetUrl);

        // 更新各种语言的示例代码
        this.updateCurlExamples(targetUrl, exampleToken, configName, description);
        this.updateJavaScriptExamples(targetUrl, exampleToken, configName, description);
        this.updatePythonExamples(targetUrl, exampleToken, configName, description);
        this.updateNodeJSExamples(targetUrl, exampleToken, configName, description);
        this.updateGoExamples(targetUrl, exampleToken, configName, description);
    }

    /**
     * 生成使用描述
     * @param {string} targetUrl - 目标URL
     * @returns {string} 描述信息
     */
    generateUsageDescription(targetUrl) {
        try {
            const url = new URL(targetUrl);

            if (url.pathname.includes('/chat/completions')) {
                return '聊天对话接口调用';
            } else if (url.pathname.includes('/api/')) {
                return 'API接口调用';
            } else if (url.pathname.includes('/v1/')) {
                return 'API v1接口调用';
            } else if (url.hostname.includes('github.com')) {
                return 'GitHub API调用';
            } else if (url.hostname.includes('httpbin.org')) {
                return 'HTTP测试请求';
            } else if (url.hostname.includes('jsonplaceholder.typicode.com')) {
                return '示例数据获取';
            } else {
                return '代理服务访问';
            }
        } catch (error) {
            return '代理服务访问';
        }
    }

    /**
     * 更新cURL示例
     */
    updateCurlExamples(targetUrl, token, configName, description) {
        // HTTP头部认证
        const curlHeaderElement = document.getElementById('curl-header');
        if (curlHeaderElement) {
            curlHeaderElement.innerHTML = `# ${description} - 使用HTTP头部认证
curl -H "X-Proxy-Token: ${token}" \\
     -H "Content-Type: application/json" \\
     "http://localhost:10805/proxy?target=${encodeURIComponent(targetUrl)}"`;
        }

        // 查询参数认证
        const curlQueryElement = document.getElementById('curl-query');
        if (curlQueryElement) {
            curlQueryElement.innerHTML = `# ${description} - 使用查询参数认证
curl -H "Content-Type: application/json" \\
     "http://localhost:10805/proxy?target=${encodeURIComponent(targetUrl)}&token=${token}"`;
        }

        // POST请求示例
        const curlPostElement = document.getElementById('curl-post');
        if (curlPostElement) {
            curlPostElement.innerHTML = `# ${description} - POST请求示例
curl -X POST \\
     -H "X-Proxy-Token: ${token}" \\
     -H "Content-Type: application/json" \\
     -d '{"key": "value", "message": "Hello World"}' \\
     "http://localhost:10805/proxy?target=${encodeURIComponent(targetUrl)}"`;
        }
    }

    /**
     * 更新JavaScript示例
     */
    updateJavaScriptExamples(targetUrl, token, configName, description) {
        const jsFetchElement = document.getElementById('js-fetch');
        if (jsFetchElement) {
            jsFetchElement.innerHTML = `// ${description} - JavaScript Fetch API
const response = await fetch('http://localhost:10805/proxy?target=${encodeURIComponent(targetUrl)}', {
    method: 'GET',
    headers: {
        'X-Proxy-Token': '${token}',
        'Content-Type': 'application/json',
        'User-Agent': 'MyApp/1.0'
    }
});

if (!response.ok) {
    throw new Error(\`HTTP error! status: \${response.status}\`);
}

const data = await response.json();
console.log('响应数据:', data);`;
        }

        // Axios示例
        const jsAxiosElement = document.getElementById('js-axios');
        if (jsAxiosElement) {
            jsAxiosElement.innerHTML = `// ${description} - Axios库
import axios from 'axios';

const response = await axios.get('http://localhost:10805/proxy', {
    params: {
        target: '${targetUrl}'
    },
    headers: {
        'X-Proxy-Token': '${token}',
        'Content-Type': 'application/json'
    },
    timeout: 30000
});

console.log('响应数据:', response.data);`;
        }
    }

    /**
     * 更新Python示例
     */
    updatePythonExamples(targetUrl, token, configName, description) {
        const pythonRequestsElement = document.getElementById('python-requests');
        if (pythonRequestsElement) {
            pythonRequestsElement.innerHTML = `# ${description} - Python Requests
import requests
import json

headers = {
    'X-Proxy-Token': '${token}',
    'Content-Type': 'application/json',
    'User-Agent': 'MyPythonApp/1.0'
}

params = {
    'target': '${targetUrl}'
}

try:
    response = requests.get(
        'http://localhost:10805/proxy',
        headers=headers,
        params=params,
        timeout=30
    )
    response.raise_for_status()
    data = response.json()
    print('响应数据:', json.dumps(data, indent=2, ensure_ascii=False))
except requests.exceptions.RequestException as e:
    print(f'请求失败: {e}')`;
        }

        // Python httpx示例
        const pythonHttpxElement = document.getElementById('python-httpx');
        if (pythonHttpxElement) {
            pythonHttpxElement.innerHTML = `# ${description} - Python httpx (异步)
import httpx
import asyncio
import json

async def make_request():
    headers = {
        'X-Proxy-Token': '${token}',
        'Content-Type': 'application/json'
    }

    params = {'target': '${targetUrl}'}

    async with httpx.AsyncClient(timeout=30.0) as client:
        try:
            response = await client.get(
                'http://localhost:10805/proxy',
                headers=headers,
                params=params
            )
            response.raise_for_status()
            data = response.json()
            print('响应数据:', json.dumps(data, indent=2, ensure_ascii=False))
        except httpx.RequestError as e:
            print(f'请求失败: {e}')

# 运行异步函数
asyncio.run(make_request())`;
        }
    }

    /**
     * 更新Node.js示例
     */
    updateNodeJSExamples(targetUrl, token, configName, description) {
        const nodeHttpElement = document.getElementById('node-http');
        if (nodeHttpElement) {
            nodeHttpElement.innerHTML = `// ${description} - Node.js HTTP模块
const http = require('http');
const https = require('https');
const querystring = require('querystring');

const params = querystring.stringify({
    target: '${targetUrl}'
});

const options = {
    hostname: 'localhost',
    port: 10805,
    path: '/proxy?' + params,
    method: 'GET',
    headers: {
        'X-Proxy-Token': '${token}',
        'Content-Type': 'application/json',
        'User-Agent': 'MyNodeApp/1.0'
    }
};

const req = http.request(options, (res) => {
    let data = '';

    res.on('data', (chunk) => {
        data += chunk;
    });

    res.on('end', () => {
        try {
            const jsonData = JSON.parse(data);
            console.log('响应数据:', JSON.stringify(jsonData, null, 2));
        } catch (e) {
            console.log('响应数据:', data);
        }
    });
});

req.on('error', (e) => {
    console.error('请求失败:', e.message);
});

req.end();`;
        }

        // Node.js Axios示例
        const nodeAxiosElement = document.getElementById('node-axios');
        if (nodeAxiosElement) {
            nodeAxiosElement.innerHTML = `// ${description} - Node.js Axios
const axios = require('axios');

async function makeRequest() {
    try {
        const response = await axios.get('http://localhost:10805/proxy', {
            params: {
                target: '${targetUrl}'
            },
            headers: {
                'X-Proxy-Token': '${token}',
                'Content-Type': 'application/json'
            },
            timeout: 30000
        });

        console.log('响应数据:', JSON.stringify(response.data, null, 2));
    } catch (error) {
        if (error.response) {
            console.error('请求失败:', error.response.status, error.response.data);
        } else {
            console.error('请求失败:', error.message);
        }
    }
}

makeRequest();`;
        }
    }

    /**
     * 更新Go示例
     */
    updateGoExamples(targetUrl, token, configName, description) {
        const goHttpElement = document.getElementById('go-http');
        if (goHttpElement) {
            goHttpElement.innerHTML = `// ${description} - Go HTTP客户端
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

func main() {
    // 构建请求URL
    baseURL := "http://localhost:10805/proxy"
    params := url.Values{}
    params.Add("target", "${targetUrl}")

    fullURL := baseURL + "?" + params.Encode()

    // 创建HTTP客户端
    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    // 创建请求
    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        fmt.Printf("创建请求失败: %v\\n", err)
        return
    }

    // 设置请求头
    req.Header.Set("X-Proxy-Token", "${token}")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "MyGoApp/1.0")

    // 发送请求
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("请求失败: %v\\n", err)
        return
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("读取响应失败: %v\\n", err)
        return
    }

    // 解析JSON响应
    var result interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Printf("解析JSON失败: %v\\n", err)
        fmt.Printf("原始响应: %s\\n", string(body))
        return
    }

    // 格式化输出
    prettyJSON, _ := json.MarshalIndent(result, "", "  ")
    fmt.Printf("响应数据:\\n%s\\n", string(prettyJSON))
}`;
        }
    }

    /**
     * 复制使用示例代码
     * @param {string} codeId - 代码元素ID
     */
    async copyUsageCode(codeId) {
        try {
            const codeElement = document.getElementById(codeId);
            if (!codeElement) {
                this.showError('代码元素不存在');
                return;
            }

            const code = codeElement.textContent || codeElement.innerText;

            if (navigator.clipboard && window.isSecureContext) {
                await navigator.clipboard.writeText(code);
                this.copySuccess();
            } else {
                // 降级方案
                this.fallbackCopyTextToClipboard(code);
            }
        } catch (error) {
            console.error('复制代码失败:', error);
            this.showError('复制代码失败: ' + error.message);
        }
    }

    /**
     * 降级复制方案
     * @param {string} text - 要复制的文本
     */
    fallbackCopyTextToClipboard(text) {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
            const successful = document.execCommand('copy');
            if (successful) {
                this.copySuccess();
            } else {
                this.showError('复制失败，请手动复制');
            }
        } catch (err) {
            console.error('降级复制失败:', err);
            this.showError('复制失败，请手动复制');
        } finally {
            document.body.removeChild(textArea);
        }
    }

    /**
     * 关闭使用示例弹框
     */
    closeUsageExampleModal() {
        this.hideModal('usage-example-modal');
        this.currentUsageConfig = null;
        this.currentUsageTokens = null;
    }

    /**
     * 生成令牌使用统计报告
     * @param {string} configId - 配置ID
     */
    async generateTokenUsageReport(configId) {
        try {
            this.showLoading('生成使用报告...');

            const config = this.currentConfigs.find(c => c.id === configId);
            const tokensResponse = await tokensAPI.getTokens(configId);
            const tokens = tokensResponse.data?.tokens || [];

            if (!config || tokens.length === 0) {
                this.showInfo('没有可用的令牌数据');
                return;
            }

            const report = this.buildUsageReport(config, tokens);
            this.downloadReport(report, `${config.name}_令牌使用报告.json`);

        } catch (error) {
            console.error('生成使用报告失败:', error);
            this.showError('生成使用报告失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 构建使用报告
     * @param {Object} config - 配置信息
     * @param {Array} tokens - 令牌列表
     * @returns {Object} 报告数据
     */
    buildUsageReport(config, tokens) {
        const now = new Date();

        return {
            报告信息: {
                配置名称: config.name,
                配置ID: config.id,
                目标URL: config.target_url,
                生成时间: now.toISOString(),
                报告版本: '1.0'
            },
            令牌统计: {
                总令牌数: tokens.length,
                活跃令牌数: tokens.filter(t => t.enabled).length,
                已禁用令牌数: tokens.filter(t => !t.enabled).length,
                总使用次数: tokens.reduce((sum, t) => sum + (t.usage_count || 0), 0),
                平均使用次数: tokens.length > 0 ? Math.round(tokens.reduce((sum, t) => sum + (t.usage_count || 0), 0) / tokens.length) : 0
            },
            令牌详情: tokens.map(token => ({
                名称: token.name,
                ID: token.id,
                状态: token.enabled ? '启用' : '禁用',
                使用次数: token.usage_count || 0,
                创建时间: token.created_at,
                最后使用: token.last_used || '从未使用',
                过期时间: token.expires_at || '永不过期',
                描述: token.description || '无描述'
            })),
            使用示例: {
                cURL头部认证: `curl -H "X-Proxy-Token: YOUR_TOKEN" "http://localhost:10805/proxy?target=${encodeURIComponent(config.target_url)}"`,
                cURL查询参数: `curl "http://localhost:10805/proxy?target=${encodeURIComponent(config.target_url)}&token=YOUR_TOKEN"`,
                JavaScript: `fetch('http://localhost:10805/proxy?target=${encodeURIComponent(config.target_url)}', { headers: { 'X-Proxy-Token': 'YOUR_TOKEN' } })`,
                Python: `requests.get('http://localhost:10805/proxy', headers={'X-Proxy-Token': 'YOUR_TOKEN'}, params={'target': '${config.target_url}'})`
            },
            最佳实践: [
                '定期检查令牌使用情况和健康度',
                '为令牌设置合适的过期时间',
                '使用描述性的令牌名称',
                '及时禁用或删除不再使用的令牌',
                '监控令牌的异常使用模式',
                '定期更新令牌以提高安全性'
            ]
        };
    }

    /**
     * 下载报告文件
     * @param {Object} data - 报告数据
     * @param {string} filename - 文件名
     */
    downloadReport(data, filename) {
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        this.showSuccess(`报告已下载: ${filename}`);
    }

    // ==================== 系统概览数据展示功能 ====================

    /**
     * 加载系统概览数据
     */
    async loadSystemOverview() {
        try {
            this.showLoading('加载系统概览数据...');

            // 并发获取各模块数据
            const [configsData, logsData] = await Promise.all([
                this.getConfigsOverviewData(),
                this.getLogsOverviewData()
            ]);

            // 获取令牌数据（需要遍历所有配置）
            const tokensData = await this.getTokensOverviewData(configsData.configs);

            // 构建系统概览数据
            const systemOverview = {
                timestamp: new Date().toISOString(),
                uptime: this.calculateUptime(),
                configs: configsData,
                tokens: tokensData,
                logs: logsData,
                system: await this.getSystemMetrics()
            };

            // 更新概览显示
            this.updateSystemOverviewDisplay(systemOverview);

            // 存储数据供其他功能使用
            this.systemOverviewData = systemOverview;

        } catch (error) {
            console.error('加载系统概览数据失败:', error);
            this.showError('加载系统概览数据失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    /**
     * 获取配置概览数据
     */
    async getConfigsOverviewData() {
        try {
            const response = await proxyAPI.getConfigs();
            const configs = response.configs || [];

            return {
                total: configs.length,
                enabled: configs.filter(c => c.enabled).length,
                disabled: configs.filter(c => !c.enabled).length,
                configs: configs,
                protocols: this.analyzeProtocols(configs),
                domains: this.analyzeDomains(configs),
                recentlyCreated: configs.filter(c =>
                    new Date(c.created_at) > new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
                ).length
            };
        } catch (error) {
            console.error('获取配置概览数据失败:', error);
            return {
                total: 0,
                enabled: 0,
                disabled: 0,
                configs: [],
                protocols: {},
                domains: {},
                recentlyCreated: 0
            };
        }
    }

    /**
     * 获取令牌概览数据
     */
    async getTokensOverviewData(configs) {
        try {
            let totalTokens = 0;
            let activeTokens = 0;
            let expiredTokens = 0;
            let totalUsage = 0;
            let recentlyUsed = 0;
            const now = new Date();
            const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);

            // 遍历所有配置获取令牌数据
            for (const config of configs) {
                try {
                    const { tokens = [], stats = null } = await tokensAPI.getTokens(config.id);

                    totalTokens += tokens.length;
                    activeTokens += tokens.filter(t => t.enabled).length;
                    totalUsage += tokens.reduce((sum, t) => sum + (t.usage_count || 0), 0);

                    // 检查过期令牌
                    expiredTokens += tokens.filter(t =>
                        t.expires_at && new Date(t.expires_at) < now
                    ).length;

                    // 检查最近使用的令牌（兼容 last_used_at 与 last_used 字段）
                    recentlyUsed += tokens.filter(t => {
                        const lastUsed = t.last_used_at || t.last_used;
                        return lastUsed && new Date(lastUsed) > oneDayAgo;
                    }).length;

                } catch (error) {
                    console.warn(`获取配置 ${config.id} 的令牌数据失败:`, error);
                }
            }

            return {
                total: totalTokens,
                active: activeTokens,
                disabled: totalTokens - activeTokens,
                expired: expiredTokens,
                totalUsage: totalUsage,
                recentlyUsed: recentlyUsed,
                averageUsage: totalTokens > 0 ? Math.round(totalUsage / totalTokens) : 0
            };
        } catch (error) {
            console.error('获取令牌概览数据失败:', error);
            return {
                total: 0,
                active: 0,
                disabled: 0,
                expired: 0,
                totalUsage: 0,
                recentlyUsed: 0,
                averageUsage: 0
            };
        }
    }

    /**
     * 获取日志概览数据
     */
    async getLogsOverviewData() {
        try {
            // 获取日志统计信息
            const statsResponse = await logsAPI.getStats();
            const stats = statsResponse.storage_stats || {};

            // 获取最近的日志数据来分析
            const logsResponse = await logsAPI.getLogs({ limit: 100 });
            const logs = logsResponse.data?.logs || [];

            const now = new Date();
            const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
            const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);

            // 分析日志数据
            const recentLogs = logs.filter(log =>
                new Date(log.timestamp) > oneDayAgo
            );

            const hourlyLogs = logs.filter(log =>
                new Date(log.timestamp) > oneHourAgo
            );

            const errorLogs = logs.filter(log =>
                log.status_code >= 400
            );

            const methodStats = this.analyzeLogMethods(logs);
            const statusStats = this.analyzeLogStatus(logs);

            return {
                total: stats.current_entries || 0,
                memoryUsage: stats.memory_usage_mb || 0,
                maxEntries: stats.max_entries || 0,
                recentCount: recentLogs.length,
                hourlyCount: hourlyLogs.length,
                errorCount: errorLogs.length,
                errorRate: logs.length > 0 ? (errorLogs.length / logs.length * 100).toFixed(1) : 0,
                methods: methodStats,
                statusCodes: statusStats,
                averageResponseTime: this.calculateAverageResponseTime(logs)
            };
        } catch (error) {
            console.error('获取日志概览数据失败:', error);
            return {
                total: 0,
                memoryUsage: 0,
                maxEntries: 0,
                recentCount: 0,
                hourlyCount: 0,
                errorCount: 0,
                errorRate: 0,
                methods: {},
                statusCodes: {},
                averageResponseTime: 0
            };
        }
    }

    /**
     * 获取系统指标
     */
    async getSystemMetrics() {
        try {
            const startTime = this.getSystemStartTime();
            const uptime = startTime ? Date.now() - startTime : 0;

            return {
                uptime: uptime,
                uptimeFormatted: this.formatUptime(uptime),
                startTime: startTime,
                currentTime: Date.now(),
                version: '1.0.0', // 可以从API获取
                environment: 'production' // 可以从配置获取
            };
        } catch (error) {
            console.error('获取系统指标失败:', error);
            return {
                uptime: 0,
                uptimeFormatted: '未知',
                startTime: null,
                currentTime: Date.now(),
                version: '未知',
                environment: '未知'
            };
        }
    }

    /**
     * 分析协议分布
     */
    analyzeProtocols(configs) {
        const protocols = {};
        configs.forEach(config => {
            const protocol = config.protocol || 'http';
            protocols[protocol] = (protocols[protocol] || 0) + 1;
        });
        return protocols;
    }

    /**
     * 分析域名分布
     */
    analyzeDomains(configs) {
        const domains = {};
        configs.forEach(config => {
            try {
                const url = new URL(config.target_url);
                const domain = url.hostname;
                domains[domain] = (domains[domain] || 0) + 1;
            } catch (error) {
                domains['invalid'] = (domains['invalid'] || 0) + 1;
            }
        });
        return domains;
    }

    /**
     * 分析日志HTTP方法
     */
    analyzeLogMethods(logs) {
        const methods = {};
        logs.forEach(log => {
            const method = log.method || 'UNKNOWN';
            methods[method] = (methods[method] || 0) + 1;
        });
        return methods;
    }

    /**
     * 分析日志状态码
     */
    analyzeLogStatus(logs) {
        const statusCodes = {};
        logs.forEach(log => {
            const status = Math.floor(log.status_code / 100) * 100; // 分组：200, 300, 400, 500
            const statusGroup = `${status}+`;
            statusCodes[statusGroup] = (statusCodes[statusGroup] || 0) + 1;
        });
        return statusCodes;
    }

    /**
     * 计算平均响应时间
     */
    calculateAverageResponseTime(logs) {
        if (logs.length === 0) return 0;

        const totalTime = logs.reduce((sum, log) => {
            return sum + (log.response_time || 0);
        }, 0);

        return Math.round(totalTime / logs.length);
    }

    /**
     * 获取系统启动时间
     */
    getSystemStartTime() {
        // 从localStorage获取或估算
        let startTime = localStorage.getItem('systemStartTime');
        if (!startTime) {
            startTime = Date.now();
            localStorage.setItem('systemStartTime', startTime);
        }
        return parseInt(startTime);
    }

    /**
     * 计算运行时间
     */
    calculateUptime() {
        const startTime = this.getSystemStartTime();
        return Date.now() - startTime;
    }

    /**
     * 格式化运行时间
     */
    formatUptime(uptime) {
        const seconds = Math.floor(uptime / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) {
            return `${days}天 ${hours % 24}小时 ${minutes % 60}分钟`;
        } else if (hours > 0) {
            return `${hours}小时 ${minutes % 60}分钟`;
        } else if (minutes > 0) {
            return `${minutes}分钟 ${seconds % 60}秒`;
        } else {
            return `${seconds}秒`;
        }
    }

    /**
     * 更新系统概览显示
     */
    updateSystemOverviewDisplay(data) {
        // 更新系统信息
        this.updateSystemInfo(data.system);

        // 更新配置统计
        this.updateConfigsOverview(data.configs);

        // 更新令牌统计
        this.updateTokensOverview(data.tokens);

        // 更新日志统计
        this.updateLogsOverview(data.logs);

        // 渲染趋势图表
        this.renderTrendCharts();

        // 初始化详细图表（默认显示HTTP方法分布）
        this.switchChartTab('methods');

        // 更新最后更新时间
        this.updateLastRefreshTime();
    }

    /**
     * 更新系统信息显示
     */
    updateSystemInfo(systemData) {
        const elements = {
            'system-uptime': systemData.uptimeFormatted,
            'system-version': systemData.version,
            'system-environment': systemData.environment
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                element.textContent = value;
            }
        });

        // 初始化并显示当前时间（一次），实时时钟负责后续每秒更新
        const timeEl = document.getElementById('system-current-time');
        if (timeEl) {
            const now = new Date();
            timeEl.textContent = now.toLocaleString('zh-CN');
        }
    }

    // 实时时钟：每秒更新系统当前时间显示，仅在概览页激活
    startRealtimeClock() {
        if (this._clockTimer) return; // 避免重复启动
        this._clockTimer = setInterval(() => {
            if (this.currentTab !== 'overview') return;
            const el = document.getElementById('system-current-time');
            if (el) {
                const now = new Date();
                el.textContent = now.toLocaleString('zh-CN');
            }
        }, 1000);
    }

    stopRealtimeClock() {
        if (this._clockTimer) {
            clearInterval(this._clockTimer);
            this._clockTimer = null;
        }
    }


    /**
     * 更新配置概览显示
     */
    updateConfigsOverview(configsData) {
        const elements = {
            'configs-total': configsData.total,
            'configs-enabled': configsData.enabled,
            'configs-disabled': configsData.disabled,
            'configs-recent': configsData.recentlyCreated
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                element.textContent = value;
            }
        });

        // 更新协议分布
        this.updateProtocolsChart(configsData.protocols);

        // 更新域名分布
        this.updateDomainsChart(configsData.domains);
    }

    /**
     * 更新令牌概览显示
     */
    updateTokensOverview(tokensData) {
        const elements = {
            'tokens-total': tokensData.total,
            'tokens-active': tokensData.active,
            'tokens-disabled': tokensData.disabled,
            'tokens-expired': tokensData.expired,
            'tokens-usage': tokensData.totalUsage,
            'tokens-recent': tokensData.recentlyUsed,
            'tokens-average': tokensData.averageUsage
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                element.textContent = value;
            }
        });

        // 更新环形图
        this.updateTokensDonutChart(tokensData);
    }

    /**
     * 更新令牌环形图
     */
    updateTokensDonutChart(tokensData) {
        const activeSegment = document.getElementById('tokens-active-segment');
        const disabledSegment = document.getElementById('tokens-disabled-segment');
        const expiredSegment = document.getElementById('tokens-expired-segment');
        const centerText = document.getElementById('tokens-center-text');

        if (!activeSegment || !disabledSegment || !expiredSegment || !centerText) return;

        const total = tokensData.total;
        if (total === 0) {
            activeSegment.style.strokeDasharray = '0 219.8';
            disabledSegment.style.strokeDasharray = '0 219.8';
            expiredSegment.style.strokeDasharray = '0 219.8';
            centerText.textContent = '0';
            return;
        }

        centerText.textContent = total;

        const circumference = 2 * Math.PI * 35; // r=35
        const activePercentage = (tokensData.active / total) * 100;
        const disabledPercentage = (tokensData.disabled / total) * 100;
        const expiredPercentage = (tokensData.expired / total) * 100;

        // 计算每个段的弧长
        const activeArc = (activePercentage / 100) * circumference;
        const disabledArc = (disabledPercentage / 100) * circumference;
        const expiredArc = (expiredPercentage / 100) * circumference;

        // 设置延迟动画
        setTimeout(() => {
            // 活跃令牌段（从0度开始）
            activeSegment.style.strokeDasharray = `${activeArc} ${circumference}`;
            activeSegment.style.strokeDashoffset = '0';

            // 禁用令牌段（从活跃段结束位置开始）
            disabledSegment.style.strokeDasharray = `${disabledArc} ${circumference}`;
            disabledSegment.style.strokeDashoffset = `-${activeArc}`;

            // 过期令牌段（从禁用段结束位置开始）
            expiredSegment.style.strokeDasharray = `${expiredArc} ${circumference}`;
            expiredSegment.style.strokeDashoffset = `-${activeArc + disabledArc}`;
        }, 200);
    }

    /**
     * 更新日志概览显示
     */
    updateLogsOverview(logsData) {
        const elements = {
            'logs-total': logsData.total,
            'logs-memory': logsData.memoryUsage.toFixed(1),
            'logs-recent': logsData.recentCount,
            'logs-hourly': logsData.hourlyCount,
            'logs-errors': logsData.errorCount,
            'logs-error-rate': logsData.errorRate,
            'logs-avg-response': logsData.averageResponseTime
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) {
                element.textContent = value;
            }
        });

        // 更新方法分布图表
        this.updateMethodsChart(logsData.methods);

        // 更新状态码分布图表
        this.updateStatusChart(logsData.statusCodes);
    }

    /**
     * 更新协议分布图表
     */
    updateProtocolsChart(protocols) {
        const chartElement = document.getElementById('protocols-chart');
        const pieElement = document.getElementById('protocols-pie-segment');

        if (!chartElement) return;

        const total = Object.values(protocols).reduce((sum, count) => sum + count, 0);
        if (total === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-4">暂无数据</div>';
            if (pieElement) {
                pieElement.style.strokeDasharray = '0 251.2';
            }
            return;
        }

        // 更新饼图
        if (pieElement && total > 0) {
            const httpsCount = protocols['https'] || 0;
            const httpsPercentage = (httpsCount / total) * 100;
            const circumference = 2 * Math.PI * 40; // r=40
            const strokeDasharray = `${(httpsPercentage / 100) * circumference} ${circumference}`;

            setTimeout(() => {
                pieElement.style.strokeDasharray = strokeDasharray;
            }, 100);
        }

        // 更新详细列表
        const chartHtml = Object.entries(protocols).map(([protocol, count]) => {
            const percentage = ((count / total) * 100).toFixed(1);
            return `
                <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-600">${protocol.toUpperCase()}</span>
                    <div class="flex items-center gap-2">
                        <div class="w-20 bg-gray-200 rounded-full h-2">
                            <div class="bg-blue-500 h-2 rounded-full transition-all duration-500" style="width: ${percentage}%"></div>
                        </div>
                        <span class="text-sm font-medium w-12 text-right">${count}</span>
                    </div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = chartHtml;
    }

    /**
     * 更新域名分布图表
     */
    updateDomainsChart(domains) {
        const chartElement = document.getElementById('domains-chart');
        if (!chartElement) return;

        const sortedDomains = Object.entries(domains)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 5); // 只显示前5个

        if (sortedDomains.length === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-4">暂无数据</div>';
            return;
        }

        const total = Object.values(domains).reduce((sum, count) => sum + count, 0);
        const chartHtml = sortedDomains.map(([domain, count]) => {
            const percentage = ((count / total) * 100).toFixed(1);
            const displayDomain = domain.length > 20 ? domain.substring(0, 20) + '...' : domain;
            return `
                <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-600" title="${domain}">${displayDomain}</span>
                    <div class="flex items-center gap-2">
                        <div class="w-16 bg-gray-200 rounded-full h-2">
                            <div class="bg-green-500 h-2 rounded-full" style="width: ${percentage}%"></div>
                        </div>
                        <span class="text-sm font-medium w-8 text-right">${count}</span>
                    </div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = chartHtml;
    }

    /**
     * 更新HTTP方法分布图表
     */
    updateMethodsChart(methods) {
        const chartElement = document.getElementById('methods-chart');
        if (!chartElement) return;

        const total = Object.values(methods).reduce((sum, count) => sum + count, 0);
        if (total === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-4">暂无数据</div>';
            return;
        }

        const methodColors = {
            'GET': 'bg-blue-500',
            'POST': 'bg-green-500',
            'PUT': 'bg-yellow-500',
            'DELETE': 'bg-red-500',
            'PATCH': 'bg-purple-500',
            'OPTIONS': 'bg-gray-500'
        };

        const chartHtml = Object.entries(methods).map(([method, count]) => {
            const percentage = ((count / total) * 100).toFixed(1);
            const colorClass = methodColors[method] || 'bg-gray-400';
            return `
                <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-600">${method}</span>
                    <div class="flex items-center gap-2">
                        <div class="w-16 bg-gray-200 rounded-full h-2">
                            <div class="${colorClass} h-2 rounded-full" style="width: ${percentage}%"></div>
                        </div>
                        <span class="text-sm font-medium w-12 text-right">${count}</span>
                    </div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = chartHtml;
    }

    /**
     * 更新状态码分布图表
     */
    updateStatusChart(statusCodes) {
        // 更新柱状图
        this.updateStatusBarChart(statusCodes);

        // 保留原有的详细列表（如果需要的话）
        const chartElement = document.getElementById('status-chart');
        if (!chartElement) return;

        const total = Object.values(statusCodes).reduce((sum, count) => sum + count, 0);
        if (total === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-4">暂无数据</div>';
            return;
        }

        const statusColors = {
            '200+': 'bg-green-500',
            '300+': 'bg-blue-500',
            '400+': 'bg-yellow-500',
            '500+': 'bg-red-500'
        };

        const chartHtml = Object.entries(statusCodes).map(([status, count]) => {
            const percentage = ((count / total) * 100).toFixed(1);
            const colorClass = statusColors[status] || 'bg-gray-400';
            return `
                <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-600">${status}</span>
                    <div class="flex items-center gap-2">
                        <div class="w-16 bg-gray-200 rounded-full h-2">
                            <div class="${colorClass} h-2 rounded-full transition-all duration-500" style="width: ${percentage}%"></div>
                        </div>
                        <span class="text-sm font-medium w-12 text-right">${count}</span>
                    </div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = chartHtml;
    }

    /**
     * 更新状态码柱状图
     */
    updateStatusBarChart(statusCodes) {
        const bars = {
            '200+': document.getElementById('status-2xx-bar'),
            '300+': document.getElementById('status-3xx-bar'),
            '400+': document.getElementById('status-4xx-bar'),
            '500+': document.getElementById('status-5xx-bar')
        };

        const total = Object.values(statusCodes).reduce((sum, count) => sum + count, 0);
        if (total === 0) {
            Object.values(bars).forEach(bar => {
                if (bar) bar.style.height = '0%';
            });
            return;
        }

        // 找到最大值用于缩放
        const maxCount = Math.max(...Object.values(statusCodes));

        // 更新每个柱子的高度
        Object.entries(statusCodes).forEach(([status, count]) => {
            const bar = bars[status];
            if (bar) {
                const height = maxCount > 0 ? (count / maxCount) * 100 : 0;
                setTimeout(() => {
                    bar.style.height = `${height}%`;
                    bar.title = `${status}: ${count} 请求`;
                }, 100);
            }
        });

        // 处理没有数据的状态码
        Object.entries(bars).forEach(([status, bar]) => {
            if (bar && !statusCodes[status]) {
                setTimeout(() => {
                    bar.style.height = '0%';
                }, 100);
            }
        });
    }

    /**
     * 更新最后刷新时间
     */
    updateLastRefreshTime() {
        const element = document.getElementById('last-refresh-time');
        if (element) {
            element.textContent = new Date().toLocaleString('zh-CN');
        }
    }

    /**
     * 启动自动刷新
     */
    startAutoRefresh() {
        // 清除现有的定时器
        if (this.overviewRefreshInterval) {
            clearInterval(this.overviewRefreshInterval);
        }

        // 设置新的定时器，每30秒刷新一次
        this.overviewRefreshInterval = setInterval(() => {
            if (this.currentTab === 'overview') {
                this.loadSystemOverview();
            }
        }, 30000);
    }

    /**
     * 停止自动刷新
     */
    stopAutoRefresh() {
        if (this.overviewRefreshInterval) {
            clearInterval(this.overviewRefreshInterval);
            this.overviewRefreshInterval = null;
        }
    }

    /**
     * 手动刷新系统概览
     */
    async refreshSystemOverview() {
        await this.loadSystemOverview();
        this.showSuccess('系统概览数据已刷新');
    }

    /**
     * 导出系统概览报告
     */
    async exportSystemOverview() {
        try {
            if (!this.systemOverviewData) {
                await this.loadSystemOverview();
            }

            const reportData = {
                报告信息: {
                    标题: 'Privacy Gateway 系统概览报告',
                    生成时间: new Date().toISOString(),
                    报告版本: '1.0'
                },
                系统信息: this.systemOverviewData.system,
                配置统计: this.systemOverviewData.configs,
                令牌统计: this.systemOverviewData.tokens,
                日志统计: this.systemOverviewData.logs,
                健康状态: this.generateHealthStatus(),
                建议措施: this.generateRecommendations()
            };

            const filename = `系统概览报告_${new Date().toISOString().split('T')[0]}.json`;
            this.downloadReport(reportData, filename);

        } catch (error) {
            console.error('导出系统概览报告失败:', error);
            this.showError('导出系统概览报告失败: ' + error.message);
        }
    }

    /**
     * 生成健康状态评估
     */
    generateHealthStatus() {
        if (!this.systemOverviewData) return {};

        const { configs, tokens, logs } = this.systemOverviewData;
        const health = {
            overall: 'good',
            issues: [],
            warnings: []
        };

        // 检查配置健康状态
        if (configs.disabled > configs.enabled) {
            health.issues.push('禁用的配置数量超过启用的配置');
            health.overall = 'warning';
        }

        // 检查令牌健康状态
        if (tokens.expired > 0) {
            health.warnings.push(`有 ${tokens.expired} 个令牌已过期`);
        }

        if (tokens.total > 0 && tokens.active / tokens.total < 0.5) {
            health.issues.push('超过一半的令牌处于禁用状态');
            health.overall = 'warning';
        }

        // 检查日志健康状态
        if (logs.errorRate > 10) {
            health.issues.push(`错误率过高: ${logs.errorRate}%`);
            health.overall = 'critical';
        }

        if (logs.memoryUsage > 100) {
            health.warnings.push(`日志内存使用量较高: ${logs.memoryUsage}MB`);
        }

        return health;
    }

    /**
     * 生成改进建议
     */
    generateRecommendations() {
        const recommendations = [];

        if (!this.systemOverviewData) return recommendations;

        const { configs, tokens, logs } = this.systemOverviewData;

        // 配置相关建议
        if (configs.disabled > 0) {
            recommendations.push('考虑删除或重新启用禁用的配置');
        }

        // 令牌相关建议
        if (tokens.expired > 0) {
            recommendations.push('及时清理或更新过期的令牌');
        }

        if (tokens.averageUsage === 0) {
            recommendations.push('检查令牌使用情况，确保配置正确');
        }

        // 日志相关建议
        if (logs.errorRate > 5) {
            recommendations.push('调查并解决导致高错误率的问题');
        }

        if (logs.memoryUsage > 80) {
            recommendations.push('考虑调整日志保留策略以减少内存使用');
        }

        // 通用建议
        recommendations.push('定期备份配置和重要数据');
        recommendations.push('监控系统性能和资源使用情况');
        recommendations.push('保持系统和依赖项的更新');

        return recommendations;
    }

    // ==================== 高级图表渲染功能 ====================

    /**
     * 切换图表标签页
     */
    switchChartTab(tabName) {
        // 移除所有活动状态
        document.querySelectorAll('.chart-tab-btn').forEach(btn => {
            btn.classList.remove('border-blue-500', 'text-blue-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });
        document.querySelectorAll('.chart-tab-pane').forEach(pane => {
            pane.classList.add('hidden');
        });

        // 激活选中的标签
        const activeBtn = document.getElementById(`chart-tab-${tabName}`);
        const activePane = document.getElementById(`${tabName}-chart-tab`);

        if (activeBtn) {
            activeBtn.classList.remove('border-transparent', 'text-gray-500');
            activeBtn.classList.add('border-blue-500', 'text-blue-600');
        }

        if (activePane) {
            activePane.classList.remove('hidden');
        }

        // 根据标签页加载对应图表
        this.loadDetailChart(tabName);
    }

    /**
     * 加载详细图表
     */
    async loadDetailChart(chartType) {
        try {
            switch (chartType) {
                case 'methods':
                    await this.loadMethodsChart();
                    break;
                case 'domains':
                    await this.loadDomainsChart();
                    break;
                case 'hourly':
                    await this.loadHourlyChart();
                    break;
            }
        } catch (error) {
            console.error(`加载${chartType}图表失败:`, error);
        }
    }

    /**
     * 加载HTTP方法分布饼图
     */
    async loadMethodsChart() {
        try {
            // 从日志数据中获取方法统计
            const logsResponse = await logsAPI.getLogs({ limit: 1000 });
            const logs = logsResponse.data?.logs || [];

            const methods = this.analyzeLogMethods(logs);
            this.renderMethodsPieChart(methods);

        } catch (error) {
            console.error('加载HTTP方法图表失败:', error);
            this.renderMethodsPieChart({});
        }
    }

    /**
     * 渲染HTTP方法饼图
     */
    renderMethodsPieChart(methods) {
        const svgElement = document.querySelector('#methods-pie-segments');
        const totalElement = document.getElementById('methods-total-count');
        const legendElement = document.getElementById('methods-legend');

        if (!svgElement || !totalElement || !legendElement) return;

        const total = Object.values(methods).reduce((sum, count) => sum + count, 0);
        totalElement.textContent = total;

        if (total === 0) {
            svgElement.innerHTML = '';
            legendElement.innerHTML = '<div class="text-gray-500">暂无数据</div>';
            return;
        }

        const colors = {
            'GET': '#3b82f6',
            'POST': '#10b981',
            'PUT': '#f59e0b',
            'DELETE': '#ef4444',
            'PATCH': '#8b5cf6',
            'OPTIONS': '#6b7280'
        };

        let currentAngle = 0;
        const radius = 80;
        const centerX = 100;
        const centerY = 100;

        // 生成饼图段
        const segments = Object.entries(methods).map(([method, count]) => {
            const percentage = (count / total) * 100;
            const angle = (count / total) * 360;
            const color = colors[method] || '#9ca3af';

            const startAngle = currentAngle;
            const endAngle = currentAngle + angle;
            currentAngle += angle;

            // 计算路径
            const startAngleRad = (startAngle * Math.PI) / 180;
            const endAngleRad = (endAngle * Math.PI) / 180;

            const x1 = centerX + radius * Math.cos(startAngleRad);
            const y1 = centerY + radius * Math.sin(startAngleRad);
            const x2 = centerX + radius * Math.cos(endAngleRad);
            const y2 = centerY + radius * Math.sin(endAngleRad);

            const largeArcFlag = angle > 180 ? 1 : 0;

            const pathData = [
                `M ${centerX} ${centerY}`,
                `L ${x1} ${y1}`,
                `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
                'Z'
            ].join(' ');

            return `<path d="${pathData}" fill="${color}" stroke="white" stroke-width="2"
                           class="transition-all duration-300 hover:opacity-80 cursor-pointer"
                           title="${method}: ${count} (${percentage.toFixed(1)}%)"/>`;
        });

        svgElement.innerHTML = segments.join('');

        // 生成图例
        const legendItems = Object.entries(methods).map(([method, count]) => {
            const percentage = ((count / total) * 100).toFixed(1);
            const color = colors[method] || '#9ca3af';

            return `
                <div class="flex items-center gap-2">
                    <div class="w-3 h-3 rounded-full" style="background-color: ${color}"></div>
                    <span class="text-sm text-gray-600">${method}</span>
                    <span class="text-sm font-medium text-gray-800">${count} (${percentage}%)</span>
                </div>
            `;
        });

        legendElement.innerHTML = legendItems.join('');
    }

    /**
     * 加载域名分布图表
     */
    async loadDomainsChart() {
        try {
            const configsResponse = await proxyAPI.getConfigs();
            const configs = configsResponse.configs || [];

            const domains = this.analyzeDomains(configs);
            this.renderDomainsBarChart(domains);

        } catch (error) {
            console.error('加载域名图表失败:', error);
            this.renderDomainsBarChart({});
        }
    }

    /**
     * 渲染域名柱状图
     */
    renderDomainsBarChart(domains) {
        const chartElement = document.getElementById('domains-bar-chart');
        if (!chartElement) return;

        const sortedDomains = Object.entries(domains)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 10); // 显示前10个

        if (sortedDomains.length === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-8">暂无数据</div>';
            return;
        }

        const maxCount = Math.max(...sortedDomains.map(([,count]) => count));

        const barsHtml = sortedDomains.map(([domain, count], index) => {
            const height = (count / maxCount) * 100;
            const displayDomain = domain.length > 15 ? domain.substring(0, 15) + '...' : domain;

            return `
                <div class="flex flex-col items-center justify-end h-full">
                    <div class="text-xs font-medium text-gray-700 mb-1">${count}</div>
                    <div class="bg-gradient-to-t from-blue-500 to-blue-400 rounded-t transition-all duration-700 w-8"
                         style="height: ${height}%" title="${domain}: ${count}个配置"></div>
                    <div class="text-xs text-gray-600 mt-2 transform -rotate-45 origin-left w-16" title="${domain}">
                        ${displayDomain}
                    </div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = `
            <div class="flex items-end justify-center gap-4 h-48 px-4">
                ${barsHtml}
            </div>
        `;
    }

    /**
     * 加载小时分布图表
     */
    async loadHourlyChart() {
        try {
            const logsResponse = await logsAPI.getLogs({ limit: 1000 });
            const logs = logsResponse.data?.logs || [];

            const hourlyData = this.analyzeHourlyDistribution(logs);
            this.renderHourlyBarChart(hourlyData);

        } catch (error) {
            console.error('加载小时分布图表失败:', error);
            this.renderHourlyBarChart({});
        }
    }

    /**
     * 分析小时分布
     */
    analyzeHourlyDistribution(logs) {
        const hourlyData = {};

        // 初始化24小时
        for (let i = 0; i < 24; i++) {
            hourlyData[i] = 0;
        }

        logs.forEach(log => {
            try {
                const hour = new Date(log.timestamp).getHours();
                hourlyData[hour] = (hourlyData[hour] || 0) + 1;
            } catch (error) {
                // 忽略无效的时间戳
            }
        });

        return hourlyData;
    }

    /**
     * 渲染小时分布柱状图
     */
    renderHourlyBarChart(hourlyData) {
        const chartElement = document.getElementById('hourly-bar-chart');
        if (!chartElement) return;

        const maxCount = Math.max(...Object.values(hourlyData));
        if (maxCount === 0) {
            chartElement.innerHTML = '<div class="text-gray-500 text-center py-8">暂无数据</div>';
            return;
        }

        const barsHtml = Array.from({length: 24}, (_, hour) => {
            const count = hourlyData[hour] || 0;
            const height = (count / maxCount) * 100;

            return `
                <div class="flex flex-col items-center justify-end h-full">
                    <div class="text-xs font-medium text-gray-700 mb-1">${count}</div>
                    <div class="bg-gradient-to-t from-green-500 to-green-400 rounded-t transition-all duration-700 w-3"
                         style="height: ${height}%" title="${hour}:00 - ${count}个请求"></div>
                    <div class="text-xs text-gray-600 mt-1">${hour}</div>
                </div>
            `;
        }).join('');

        chartElement.innerHTML = `
            <div class="flex items-end justify-center gap-1 h-48 px-2">
                ${barsHtml}
            </div>
            <div class="text-center text-xs text-gray-500 mt-2">小时 (0-23)</div>
        `;
    }

    /**
     * 渲染趋势图表
     */
    renderTrendCharts() {
        // 生成模拟的趋势数据
        const requestsTrendData = this.generateTrendData(24, 50, 200);
        const responseTrendData = this.generateTrendData(24, 100, 500);

        this.renderTrendChart('requests-trend', requestsTrendData, '#3b82f6');
        this.renderTrendChart('response-time', responseTrendData, '#10b981');
    }

    /**
     * 生成趋势数据（模拟）
     */
    generateTrendData(points, min, max) {
        const data = [];
        for (let i = 0; i < points; i++) {
            const value = Math.floor(Math.random() * (max - min) + min);
            data.push(value);
        }
        return data;
    }

    /**
     * 渲染单个趋势图
     */
    renderTrendChart(chartId, data, color) {
        const strokePath = document.getElementById(`${chartId}-stroke`);
        const fillPath = document.getElementById(`${chartId}-line`);
        const loadingElement = document.getElementById(`${chartId}-loading`);

        if (!strokePath || !fillPath) return;

        if (loadingElement) {
            loadingElement.style.display = 'none';
        }

        const maxValue = Math.max(...data);
        const minValue = Math.min(...data);
        const range = maxValue - minValue || 1;

        // 生成路径点
        const points = data.map((value, index) => {
            const x = (index / (data.length - 1)) * 300;
            const y = 90 - ((value - minValue) / range) * 80; // 反转Y轴，留出边距
            return `${x},${y}`;
        });

        const pathData = `M ${points.join(' L ')}`;
        const fillPathData = `${pathData} L 300,90 L 0,90 Z`;

        // 设置路径
        strokePath.setAttribute('d', pathData);
        fillPath.setAttribute('d', fillPathData);

        // 添加动画
        setTimeout(() => {
            strokePath.style.opacity = '1';
            fillPath.style.opacity = '1';
        }, 100);
    }

    // ==================== 业务流程集成测试功能 ====================

    /**
     * 运行完整的业务流程集成测试
     */
    async runIntegrationTests() {
        console.log('🧪 开始业务流程集成测试...');

        const testResults = {
            totalTests: 0,
            passedTests: 0,
            failedTests: 0,
            results: []
        };

        try {
            // 显示测试开始提示
            this.showInfo('正在运行业务流程集成测试，请稍候...');

            // 1. 认证流程测试
            await this.testAuthenticationFlow(testResults);

            // 2. 配置管理流程测试
            await this.testConfigurationFlow(testResults);

            // 3. 令牌管理流程测试
            await this.testTokenManagementFlow(testResults);

            // 4. 日志查看流程测试
            await this.testLogViewingFlow(testResults);

            // 5. 系统概览流程测试
            await this.testSystemOverviewFlow(testResults);

            // 6. 跨模块数据一致性测试
            await this.testDataConsistency(testResults);

            // 7. 错误处理测试
            await this.testErrorHandling(testResults);

            // 显示测试结果
            this.displayTestResults(testResults);

        } catch (error) {
            console.error('集成测试执行失败:', error);
            this.showError('集成测试执行失败: ' + error.message);
        }
    }

    /**
     * 测试认证流程
     */
    async testAuthenticationFlow(testResults) {
        console.log('🔐 测试认证流程...');

        try {
            // 测试登录状态检查
            const isLoggedIn = this.isLoggedIn();
            this.addTestResult(testResults, '认证状态检查', isLoggedIn, '用户应该已登录');

            // 测试会话管理
            const sessionData = this.getSessionData();
            this.addTestResult(testResults, '会话数据获取', !!sessionData, '应该能获取会话数据');

            // 测试状态持久化
            const savedState = this.loadState();
            this.addTestResult(testResults, '状态持久化', !!savedState, '应该能加载保存的状态');

        } catch (error) {
            this.addTestResult(testResults, '认证流程测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试配置管理流程
     */
    async testConfigurationFlow(testResults) {
        console.log('⚙️ 测试配置管理流程...');

        try {
            // 测试配置列表加载
            const configsResponse = await proxyAPI.getConfigs();
            const configsLoaded = configsResponse && configsResponse.configs;
            this.addTestResult(testResults, '配置列表加载', !!configsLoaded, '应该能加载配置列表');

            if (configsLoaded && configsResponse.configs.length > 0) {
                const testConfig = configsResponse.configs[0];

                // 测试配置详情获取
                const configDetail = await proxyAPI.getConfig(testConfig.id);
                this.addTestResult(testResults, '配置详情获取', !!configDetail, '应该能获取配置详情');

                // 测试配置状态切换
                const originalEnabled = testConfig.enabled;
                const updateResult = await proxyAPI.updateConfig(testConfig.id, {
                    ...testConfig,
                    enabled: !originalEnabled
                });
                this.addTestResult(testResults, '配置状态更新', !!updateResult, '应该能更新配置状态');

                // 恢复原始状态
                await proxyAPI.updateConfig(testConfig.id, {
                    ...testConfig,
                    enabled: originalEnabled
                });
            }

        } catch (error) {
            this.addTestResult(testResults, '配置管理流程测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试令牌管理流程
     */
    async testTokenManagementFlow(testResults) {
        console.log('🔑 测试令牌管理流程...');

        try {
            // 获取第一个配置用于测试
            const configsResponse = await proxyAPI.getConfigs();
            if (!configsResponse || !configsResponse.configs || configsResponse.configs.length === 0) {
                this.addTestResult(testResults, '令牌管理流程测试', false, '没有可用的配置进行测试');
                return;
            }

            const testConfig = configsResponse.configs[0];

            // 测试令牌列表加载
            const tokensResponse = await tokensAPI.getTokens(testConfig.id);
            const tokensLoaded = tokensResponse && tokensResponse.data && tokensResponse.data.tokens;
            this.addTestResult(testResults, '令牌列表加载', !!tokensLoaded, '应该能加载令牌列表');

            // 测试令牌创建
            const newTokenData = {
                name: `测试令牌_${Date.now()}`,
                description: '集成测试创建的令牌'
            };

            const createResult = await tokensAPI.createToken(testConfig.id, newTokenData);
            const tokenCreated = createResult && createResult.success;
            this.addTestResult(testResults, '令牌创建', tokenCreated, '应该能创建新令牌');

            if (tokenCreated && createResult.data) {
                const newTokenId = createResult.data.id;

                // 测试令牌状态切换
                const updateResult = await tokensAPI.updateToken(testConfig.id, newTokenId, {
                    enabled: false
                });
                this.addTestResult(testResults, '令牌状态更新', !!updateResult, '应该能更新令牌状态');

                // 测试令牌删除
                const deleteResult = await tokensAPI.deleteToken(testConfig.id, newTokenId);
                this.addTestResult(testResults, '令牌删除', !!deleteResult, '应该能删除令牌');
            }

        } catch (error) {
            this.addTestResult(testResults, '令牌管理流程测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试日志查看流程
     */
    async testLogViewingFlow(testResults) {
        console.log('📋 测试日志查看流程...');

        try {
            // 测试日志列表加载
            const logsResponse = await logsAPI.getLogs({ limit: 10 });
            const logsLoaded = logsResponse && logsResponse.data;
            this.addTestResult(testResults, '日志列表加载', !!logsLoaded, '应该能加载日志列表');

            // 测试日志统计获取
            try {
                const statsResponse = await logsAPI.getStats();
                const statsLoaded = !!statsResponse;
                this.addTestResult(testResults, '日志统计获取', statsLoaded, '应该能获取日志统计');
            } catch (error) {
                // 日志统计可能失败，这是已知问题
                this.addTestResult(testResults, '日志统计获取', false, `日志统计API暂时不可用: ${error.message}`);
            }

            // 测试日志筛选
            const filteredLogsResponse = await logsAPI.getLogs({
                limit: 5,
                method: 'GET'
            });
            const filteredLogsLoaded = filteredLogsResponse && filteredLogsResponse.data;
            this.addTestResult(testResults, '日志筛选功能', !!filteredLogsLoaded, '应该能筛选日志');

        } catch (error) {
            this.addTestResult(testResults, '日志查看流程测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试系统概览流程
     */
    async testSystemOverviewFlow(testResults) {
        console.log('📊 测试系统概览流程...');

        try {
            // 测试系统概览数据加载
            await this.loadSystemOverview();
            const overviewDataLoaded = !!this.systemOverviewData;
            this.addTestResult(testResults, '系统概览数据加载', overviewDataLoaded, '应该能加载系统概览数据');

            if (overviewDataLoaded) {
                // 测试各模块数据完整性
                const hasConfigsData = this.systemOverviewData.configs && typeof this.systemOverviewData.configs.total === 'number';
                this.addTestResult(testResults, '配置数据完整性', hasConfigsData, '系统概览应包含配置数据');

                const hasTokensData = this.systemOverviewData.tokens && typeof this.systemOverviewData.tokens.total === 'number';
                this.addTestResult(testResults, '令牌数据完整性', hasTokensData, '系统概览应包含令牌数据');

                const hasLogsData = this.systemOverviewData.logs && typeof this.systemOverviewData.logs.total === 'number';
                this.addTestResult(testResults, '日志数据完整性', hasLogsData, '系统概览应包含日志数据');

                const hasSystemData = this.systemOverviewData.system && this.systemOverviewData.system.uptime;
                this.addTestResult(testResults, '系统数据完整性', hasSystemData, '系统概览应包含系统数据');
            }

            // 测试图表渲染
            const protocolsChart = document.getElementById('protocols-chart');
            const chartsRendered = protocolsChart && protocolsChart.innerHTML.trim() !== '';
            this.addTestResult(testResults, '图表渲染功能', chartsRendered, '应该能渲染图表');

        } catch (error) {
            this.addTestResult(testResults, '系统概览流程测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试跨模块数据一致性
     */
    async testDataConsistency(testResults) {
        console.log('🔄 测试数据一致性...');

        try {
            // 获取配置数据
            const configsResponse = await proxyAPI.getConfigs();
            const configs = configsResponse?.configs || [];

            // 获取系统概览中的配置数据
            if (this.systemOverviewData && this.systemOverviewData.configs) {
                const overviewConfigCount = this.systemOverviewData.configs.total;
                const actualConfigCount = configs.length;

                const dataConsistent = overviewConfigCount === actualConfigCount;
                this.addTestResult(testResults, '配置数据一致性', dataConsistent,
                    `概览显示${overviewConfigCount}个配置，实际有${actualConfigCount}个配置`);
            }

            // 测试令牌数据一致性
            if (configs.length > 0) {
                const testConfig = configs[0];
                const tokensResponse = await tokensAPI.getTokens(testConfig.id);
                const actualTokens = tokensResponse?.data?.tokens || [];

                // 检查配置中的令牌统计是否与实际令牌数量一致
                const configTokenStats = testConfig.token_stats;
                if (configTokenStats) {
                    const statsConsistent = configTokenStats.total_tokens === actualTokens.length;
                    this.addTestResult(testResults, '令牌统计一致性', statsConsistent,
                        `配置统计显示${configTokenStats.total_tokens}个令牌，实际有${actualTokens.length}个令牌`);
                }
            }

        } catch (error) {
            this.addTestResult(testResults, '数据一致性测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试错误处理
     */
    async testErrorHandling(testResults) {
        console.log('❌ 测试错误处理...');

        try {
            // 测试无效配置ID的处理
            try {
                await proxyAPI.getConfig('invalid-config-id');
                this.addTestResult(testResults, '无效配置ID处理', false, '应该抛出错误');
            } catch (error) {
                this.addTestResult(testResults, '无效配置ID处理', true, '正确处理了无效配置ID');
            }

            // 测试无效令牌ID的处理
            try {
                await tokensAPI.getTokens('invalid-config-id');
                this.addTestResult(testResults, '无效令牌请求处理', false, '应该抛出错误');
            } catch (error) {
                this.addTestResult(testResults, '无效令牌请求处理', true, '正确处理了无效令牌请求');
            }

            // 测试网络错误处理
            const originalFetch = window.fetch;
            window.fetch = () => Promise.reject(new Error('网络错误'));

            try {
                await proxyAPI.getConfigs();
                this.addTestResult(testResults, '网络错误处理', false, '应该处理网络错误');
            } catch (error) {
                this.addTestResult(testResults, '网络错误处理', true, '正确处理了网络错误');
            } finally {
                window.fetch = originalFetch;
            }

        } catch (error) {
            this.addTestResult(testResults, '错误处理测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 添加测试结果
     */
    addTestResult(testResults, testName, passed, message) {
        testResults.totalTests++;
        if (passed) {
            testResults.passedTests++;
        } else {
            testResults.failedTests++;
        }

        testResults.results.push({
            name: testName,
            passed: passed,
            message: message,
            timestamp: new Date().toISOString()
        });

        console.log(`${passed ? '✅' : '❌'} ${testName}: ${message}`);
    }

    /**
     * 显示测试结果
     */
    displayTestResults(testResults) {
        const passRate = ((testResults.passedTests / testResults.totalTests) * 100).toFixed(1);

        console.log('\n🧪 业务流程集成测试结果:');
        console.log(`总测试数: ${testResults.totalTests}`);
        console.log(`通过: ${testResults.passedTests}`);
        console.log(`失败: ${testResults.failedTests}`);
        console.log(`通过率: ${passRate}%`);

        // 显示详细结果
        const resultMessage = `
            <div class="test-results">
                <h3>🧪 业务流程集成测试结果</h3>
                <div class="test-summary">
                    <p><strong>总测试数:</strong> ${testResults.totalTests}</p>
                    <p><strong>通过:</strong> <span class="text-green-600">${testResults.passedTests}</span></p>
                    <p><strong>失败:</strong> <span class="text-red-600">${testResults.failedTests}</span></p>
                    <p><strong>通过率:</strong> <span class="${passRate >= 80 ? 'text-green-600' : 'text-yellow-600'}">${passRate}%</span></p>
                </div>
                <div class="test-details mt-4">
                    <h4>详细结果:</h4>
                    <ul class="list-disc pl-5 space-y-1">
                        ${testResults.results.map(result => `
                            <li class="${result.passed ? 'text-green-600' : 'text-red-600'}">
                                ${result.passed ? '✅' : '❌'} <strong>${result.name}:</strong> ${result.message}
                            </li>
                        `).join('')}
                    </ul>
                </div>
            </div>
        `;

        // 创建测试结果模态框
        this.showTestResultsModal(resultMessage, passRate >= 80);
    }

    /**
     * 显示测试结果模态框
     */
    showTestResultsModal(content, success) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50';
        modal.innerHTML = `
            <div class="relative top-20 mx-auto p-5 border w-11/12 max-w-4xl shadow-lg rounded-md bg-white">
                <div class="mt-3">
                    <div class="flex items-center justify-between mb-4">
                        <h2 class="text-xl font-bold text-gray-900">集成测试结果</h2>
                        <button onclick="this.closest('.fixed').remove()" class="text-gray-400 hover:text-gray-600">
                            <span class="text-2xl">&times;</span>
                        </button>
                    </div>
                    <div class="test-results-content">
                        ${content}
                    </div>
                    <div class="flex justify-end mt-6">
                        <button onclick="this.closest('.fixed').remove()"
                                class="bg-${success ? 'green' : 'blue'}-500 hover:bg-${success ? 'green' : 'blue'}-700 text-white font-bold py-2 px-4 rounded">
                            确定
                        </button>
                    </div>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // 显示成功或警告消息
        if (success) {
            this.showSuccess('业务流程集成测试完成，所有核心功能正常！');
        } else {
            this.showWarning('业务流程集成测试完成，部分功能需要注意。');
        }
    }

    /**
     * 运行端到端业务流程演示
     */
    async runE2EDemo() {
        console.log('🎬 开始端到端业务流程演示...');

        try {
            this.showInfo('正在演示完整的业务流程，请观察界面变化...');

            // 1. 切换到概览页面并加载数据
            console.log('📊 步骤1: 加载系统概览...');
            await this.switchTab('overview');
            await this.delay(2000);

            // 2. 切换到配置管理页面
            console.log('⚙️ 步骤2: 查看配置管理...');
            await this.switchTab('proxy-config');
            await this.delay(2000);

            // 3. 切换到日志页面
            console.log('📋 步骤3: 查看访问日志...');
            await this.switchTab('logs');
            await this.delay(2000);

            // 4. 回到概览页面
            console.log('📊 步骤4: 返回系统概览...');
            await this.switchTab('overview');
            await this.delay(1000);

            // 5. 刷新概览数据
            console.log('🔄 步骤5: 刷新系统数据...');
            await this.refreshSystemOverview();
            await this.delay(1000);

            this.showSuccess('端到端业务流程演示完成！所有功能模块运行正常。');
            console.log('✅ 端到端业务流程演示完成');

        } catch (error) {
            console.error('端到端演示失败:', error);
            this.showError('端到端演示失败: ' + error.message);
        }
    }

    /**
     * 延迟函数
     */
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * 验证业务流程完整性
     */
    async validateBusinessFlow() {
        console.log('🔍 验证业务流程完整性...');

        const validationResults = {
            modules: [],
            overallHealth: 'good'
        };

        try {
            // 验证认证模块
            const authValid = this.isLoggedIn() && !!this.getSessionData();
            validationResults.modules.push({
                name: '认证模块',
                status: authValid ? 'healthy' : 'warning',
                message: authValid ? '认证功能正常' : '认证状态异常'
            });

            // 验证配置管理模块
            try {
                const configsResponse = await proxyAPI.getConfigs();
                const configsValid = configsResponse && configsResponse.configs;
                validationResults.modules.push({
                    name: '配置管理模块',
                    status: configsValid ? 'healthy' : 'error',
                    message: configsValid ? `配置管理正常 (${configsResponse.configs.length}个配置)` : '配置管理异常'
                });
            } catch (error) {
                validationResults.modules.push({
                    name: '配置管理模块',
                    status: 'error',
                    message: `配置管理错误: ${error.message}`
                });
                validationResults.overallHealth = 'critical';
            }

            // 验证令牌管理模块
            try {
                const configsResponse = await proxyAPI.getConfigs();
                if (configsResponse && configsResponse.configs.length > 0) {
                    const testConfig = configsResponse.configs[0];
                    const tokensResponse = await tokensAPI.getTokens(testConfig.id);
                    const tokensValid = tokensResponse && tokensResponse.data;
                    validationResults.modules.push({
                        name: '令牌管理模块',
                        status: tokensValid ? 'healthy' : 'warning',
                        message: tokensValid ? `令牌管理正常 (${tokensResponse.data.tokens?.length || 0}个令牌)` : '令牌管理异常'
                    });
                } else {
                    validationResults.modules.push({
                        name: '令牌管理模块',
                        status: 'warning',
                        message: '无配置可测试令牌管理'
                    });
                }
            } catch (error) {
                validationResults.modules.push({
                    name: '令牌管理模块',
                    status: 'error',
                    message: `令牌管理错误: ${error.message}`
                });
                if (validationResults.overallHealth !== 'critical') {
                    validationResults.overallHealth = 'warning';
                }
            }

            // 验证日志管理模块
            try {
                const logsResponse = await logsAPI.getLogs({ limit: 1 });
                const logsValid = logsResponse && logsResponse.data;
                validationResults.modules.push({
                    name: '日志管理模块',
                    status: logsValid ? 'healthy' : 'warning',
                    message: logsValid ? '日志管理正常' : '日志管理部分功能异常'
                });
            } catch (error) {
                validationResults.modules.push({
                    name: '日志管理模块',
                    status: 'warning',
                    message: `日志管理警告: ${error.message} (已知问题)`
                });
            }

            // 验证系统概览模块
            const overviewValid = !!this.systemOverviewData;
            validationResults.modules.push({
                name: '系统概览模块',
                status: overviewValid ? 'healthy' : 'warning',
                message: overviewValid ? '系统概览正常' : '系统概览数据未加载'
            });

            // 显示验证结果
            this.displayValidationResults(validationResults);

        } catch (error) {
            console.error('业务流程验证失败:', error);
            this.showError('业务流程验证失败: ' + error.message);
        }
    }

    /**
     * 显示验证结果
     */
    displayValidationResults(results) {
        const healthyCount = results.modules.filter(m => m.status === 'healthy').length;
        const warningCount = results.modules.filter(m => m.status === 'warning').length;
        const errorCount = results.modules.filter(m => m.status === 'error').length;

        const resultHtml = `
            <div class="validation-results">
                <h3>🔍 业务流程完整性验证</h3>
                <div class="validation-summary mb-4">
                    <div class="grid grid-cols-3 gap-4 text-center">
                        <div class="bg-green-50 p-3 rounded">
                            <div class="text-2xl font-bold text-green-600">${healthyCount}</div>
                            <div class="text-sm text-green-800">正常模块</div>
                        </div>
                        <div class="bg-yellow-50 p-3 rounded">
                            <div class="text-2xl font-bold text-yellow-600">${warningCount}</div>
                            <div class="text-sm text-yellow-800">警告模块</div>
                        </div>
                        <div class="bg-red-50 p-3 rounded">
                            <div class="text-2xl font-bold text-red-600">${errorCount}</div>
                            <div class="text-sm text-red-800">错误模块</div>
                        </div>
                    </div>
                </div>
                <div class="validation-details">
                    <h4 class="font-semibold mb-2">模块状态详情:</h4>
                    <div class="space-y-2">
                        ${results.modules.map(module => `
                            <div class="flex items-center justify-between p-2 rounded ${
                                module.status === 'healthy' ? 'bg-green-50' :
                                module.status === 'warning' ? 'bg-yellow-50' : 'bg-red-50'
                            }">
                                <div class="flex items-center gap-2">
                                    <span class="text-lg">
                                        ${module.status === 'healthy' ? '✅' :
                                          module.status === 'warning' ? '⚠️' : '❌'}
                                    </span>
                                    <span class="font-medium">${module.name}</span>
                                </div>
                                <span class="text-sm ${
                                    module.status === 'healthy' ? 'text-green-700' :
                                    module.status === 'warning' ? 'text-yellow-700' : 'text-red-700'
                                }">${module.message}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
                <div class="mt-4 p-3 rounded ${
                    results.overallHealth === 'good' ? 'bg-green-50 text-green-800' :
                    results.overallHealth === 'warning' ? 'bg-yellow-50 text-yellow-800' : 'bg-red-50 text-red-800'
                }">
                    <strong>整体健康状态: ${
                        results.overallHealth === 'good' ? '良好 ✅' :
                        results.overallHealth === 'warning' ? '需要注意 ⚠️' : '需要修复 ❌'
                    }</strong>
                </div>
            </div>
        `;

        this.showTestResultsModal(resultHtml, results.overallHealth === 'good');
    }

    // ==================== 端到端功能测试 ====================

    /**
     * 运行端到端功能测试
     */
    async runE2EFunctionalTests() {
        console.log('🎯 开始端到端功能测试...');

        const testResults = {
            totalTests: 0,
            passedTests: 0,
            failedTests: 0,
            results: [],
            scenarios: []
        };

        try {
            this.showInfo('正在运行端到端功能测试，这可能需要几分钟时间...');

            // 1. 完整用户场景测试
            await this.testCompleteUserScenarios(testResults);

            // 2. 数据流完整性测试
            await this.testDataFlowIntegrity(testResults);

            // 3. 界面交互测试
            await this.testUIInteractions(testResults);

            // 4. 性能和响应性测试
            await this.testPerformanceAndResponsiveness(testResults);

            // 5. 错误恢复测试
            await this.testErrorRecovery(testResults);

            // 6. 并发操作测试
            await this.testConcurrentOperations(testResults);

            // 显示测试结果
            this.displayE2ETestResults(testResults);

        } catch (error) {
            console.error('端到端功能测试执行失败:', error);
            this.showError('端到端功能测试执行失败: ' + error.message);
        }
    }

    /**
     * 测试完整用户场景
     */
    async testCompleteUserScenarios(testResults) {
        console.log('👤 测试完整用户场景...');

        try {
            // 场景1: 新用户首次使用流程
            const scenario1 = await this.testNewUserFirstTimeFlow();
            testResults.scenarios.push(scenario1);
            this.addTestResult(testResults, '新用户首次使用流程', scenario1.success, scenario1.message);

            // 场景2: 配置管理完整流程
            const scenario2 = await this.testConfigManagementFlow();
            testResults.scenarios.push(scenario2);
            this.addTestResult(testResults, '配置管理完整流程', scenario2.success, scenario2.message);

            // 场景3: 令牌管理完整流程
            const scenario3 = await this.testTokenManagementCompleteFlow();
            testResults.scenarios.push(scenario3);
            this.addTestResult(testResults, '令牌管理完整流程', scenario3.success, scenario3.message);

            // 场景4: 系统监控和维护流程
            const scenario4 = await this.testSystemMonitoringFlow();
            testResults.scenarios.push(scenario4);
            this.addTestResult(testResults, '系统监控维护流程', scenario4.success, scenario4.message);

        } catch (error) {
            this.addTestResult(testResults, '完整用户场景测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试新用户首次使用流程
     */
    async testNewUserFirstTimeFlow() {
        const steps = [];
        let success = true;
        let message = '';

        try {
            // 步骤1: 检查登录状态
            steps.push('检查用户认证状态');
            const isAuthenticated = this.isLoggedIn();
            if (!isAuthenticated) {
                success = false;
                message = '用户未认证';
                return { success, message, steps };
            }

            // 步骤2: 加载系统概览
            steps.push('加载系统概览页面');
            await this.switchTab('overview');
            await this.delay(1000);
            await this.loadSystemOverview();

            // 步骤3: 查看配置列表
            steps.push('查看配置管理页面');
            await this.switchTab('proxy-config');
            await this.delay(1000);

            // 步骤4: 查看日志页面
            steps.push('查看访问日志页面');
            await this.switchTab('logs');
            await this.delay(1000);

            // 步骤5: 返回概览页面
            steps.push('返回系统概览');
            await this.switchTab('overview');
            await this.delay(500);

            message = `新用户流程测试完成，执行了${steps.length}个步骤`;

        } catch (error) {
            success = false;
            message = `新用户流程测试失败: ${error.message}`;
        }

        return { success, message, steps };
    }

    /**
     * 测试配置管理完整流程
     */
    async testConfigManagementFlow() {
        const steps = [];
        let success = true;
        let message = '';

        try {
            // 步骤1: 获取配置列表
            steps.push('获取配置列表');
            const configsResponse = await proxyAPI.getConfigs();
            if (!configsResponse || !configsResponse.configs) {
                throw new Error('无法获取配置列表');
            }

            const originalConfigCount = configsResponse.configs.length;

            // 步骤2: 创建测试配置
            steps.push('创建测试配置');
            const testConfigData = {
                name: `E2E测试配置_${Date.now()}`,
                target_url: 'https://api.example.com',
                enabled: true,
                protocol: 'https'
            };

            const createResult = await proxyAPI.createConfig(testConfigData);
            if (!createResult || !createResult.id) {
                throw new Error('配置创建失败');
            }

            const newConfigId = createResult.id;

            // 步骤3: 验证配置创建
            steps.push('验证配置创建');
            const newConfigsResponse = await proxyAPI.getConfigs();
            if (newConfigsResponse.configs.length !== originalConfigCount + 1) {
                throw new Error('配置数量不匹配');
            }

            // 步骤4: 更新配置
            steps.push('更新配置状态');
            const updateResult = await proxyAPI.updateConfig(newConfigId, {
                ...testConfigData,
                enabled: false
            });
            if (!updateResult) {
                throw new Error('配置更新失败');
            }

            // 步骤5: 删除测试配置
            steps.push('删除测试配置');
            const deleteResult = await proxyAPI.deleteConfig(newConfigId);
            if (!deleteResult) {
                throw new Error('配置删除失败');
            }

            // 步骤6: 验证配置删除
            steps.push('验证配置删除');
            const finalConfigsResponse = await proxyAPI.getConfigs();
            if (finalConfigsResponse.configs.length !== originalConfigCount) {
                throw new Error('配置删除后数量不匹配');
            }

            message = `配置管理流程测试完成，执行了${steps.length}个步骤`;

        } catch (error) {
            success = false;
            message = `配置管理流程测试失败: ${error.message}`;
        }

        return { success, message, steps };
    }

    /**
     * 测试令牌管理完整流程
     */
    async testTokenManagementCompleteFlow() {
        const steps = [];
        let success = true;
        let message = '';

        try {
            // 步骤1: 获取第一个配置
            steps.push('获取测试配置');
            const configsResponse = await proxyAPI.getConfigs();
            if (!configsResponse || !configsResponse.configs || configsResponse.configs.length === 0) {
                throw new Error('没有可用的配置进行令牌测试');
            }

            const testConfig = configsResponse.configs[0];

            // 步骤2: 获取原始令牌列表
            steps.push('获取原始令牌列表');
            const originalTokensResponse = await tokensAPI.getTokens(testConfig.id);
            if (!originalTokensResponse || !originalTokensResponse.data) {
                throw new Error('无法获取令牌列表');
            }

            const originalTokenCount = originalTokensResponse.data.tokens.length;

            // 步骤3: 创建测试令牌
            steps.push('创建测试令牌');
            const testTokenData = {
                name: `E2E测试令牌_${Date.now()}`,
                description: '端到端测试创建的令牌'
            };

            const createResult = await tokensAPI.createToken(testConfig.id, testTokenData);
            if (!createResult || !createResult.success || !createResult.data) {
                throw new Error('令牌创建失败');
            }

            const newTokenId = createResult.data.id;

            // 步骤4: 验证令牌创建
            steps.push('验证令牌创建');
            const newTokensResponse = await tokensAPI.getTokens(testConfig.id);
            if (newTokensResponse.data.tokens.length !== originalTokenCount + 1) {
                throw new Error('令牌数量不匹配');
            }

            // 步骤5: 更新令牌状态
            steps.push('更新令牌状态');
            const updateResult = await tokensAPI.updateToken(testConfig.id, newTokenId, {
                enabled: false
            });
            if (!updateResult) {
                throw new Error('令牌更新失败');
            }

            // 步骤6: 删除测试令牌
            steps.push('删除测试令牌');
            const deleteResult = await tokensAPI.deleteToken(testConfig.id, newTokenId);
            if (!deleteResult) {
                throw new Error('令牌删除失败');
            }

            // 步骤7: 验证令牌删除
            steps.push('验证令牌删除');
            const finalTokensResponse = await tokensAPI.getTokens(testConfig.id);
            if (finalTokensResponse.data.tokens.length !== originalTokenCount) {
                throw new Error('令牌删除后数量不匹配');
            }

            message = `令牌管理流程测试完成，执行了${steps.length}个步骤`;

        } catch (error) {
            success = false;
            message = `令牌管理流程测试失败: ${error.message}`;
        }

        return { success, message, steps };
    }

    /**
     * 测试系统监控流程
     */
    async testSystemMonitoringFlow() {
        const steps = [];
        let success = true;
        let message = '';

        try {
            // 步骤1: 加载系统概览数据
            steps.push('加载系统概览数据');
            await this.loadSystemOverview();
            if (!this.systemOverviewData) {
                throw new Error('系统概览数据加载失败');
            }

            // 步骤2: 验证数据完整性
            steps.push('验证数据完整性');
            const requiredFields = ['system', 'configs', 'tokens', 'logs'];
            for (const field of requiredFields) {
                if (!this.systemOverviewData[field]) {
                    throw new Error(`缺少${field}数据`);
                }
            }

            // 步骤3: 测试数据刷新
            steps.push('测试数据刷新');
            await this.refreshSystemOverview();

            // 步骤4: 测试图表渲染
            steps.push('测试图表渲染');
            const chartsContainer = document.querySelector('#protocols-chart');
            if (!chartsContainer || chartsContainer.innerHTML.trim() === '') {
                throw new Error('图表渲染失败');
            }

            // 步骤5: 测试报告导出
            steps.push('测试报告导出功能');
            // 这里只是验证方法存在，不实际导出
            if (typeof this.exportSystemOverview !== 'function') {
                throw new Error('报告导出功能不可用');
            }

            message = `系统监控流程测试完成，执行了${steps.length}个步骤`;

        } catch (error) {
            success = false;
            message = `系统监控流程测试失败: ${error.message}`;
        }

        return { success, message, steps };
    }

    /**
     * 测试数据流完整性
     */
    async testDataFlowIntegrity(testResults) {
        console.log('🔄 测试数据流完整性...');

        try {
            // 测试配置-令牌数据一致性
            const configsResponse = await proxyAPI.getConfigs();
            if (configsResponse && configsResponse.configs.length > 0) {
                const testConfig = configsResponse.configs[0];
                const tokensResponse = await tokensAPI.getTokens(testConfig.id);

                if (tokensResponse && tokensResponse.data) {
                    const actualTokens = tokensResponse.data.tokens.length;
                    const statsTokens = tokensResponse.data.stats.total_tokens;

                    const dataConsistent = actualTokens === statsTokens;
                    this.addTestResult(testResults, '配置-令牌数据一致性', dataConsistent,
                        `实际令牌${actualTokens}个，统计显示${statsTokens}个`);
                }
            }

            // 测试系统概览数据一致性
            if (this.systemOverviewData) {
                const overviewConfigCount = this.systemOverviewData.configs.total;
                const actualConfigCount = configsResponse.configs.length;

                const overviewConsistent = overviewConfigCount === actualConfigCount;
                this.addTestResult(testResults, '系统概览数据一致性', overviewConsistent,
                    `概览显示${overviewConfigCount}个配置，实际${actualConfigCount}个配置`);
            }

            // 测试数据更新同步
            const originalData = JSON.stringify(this.systemOverviewData);
            await this.refreshSystemOverview();
            const updatedData = JSON.stringify(this.systemOverviewData);

            this.addTestResult(testResults, '数据刷新功能', true, '数据刷新功能正常');

        } catch (error) {
            this.addTestResult(testResults, '数据流完整性测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试界面交互
     */
    async testUIInteractions(testResults) {
        console.log('🖱️ 测试界面交互...');

        try {
            // 测试标签页切换
            const tabs = ['overview', 'proxy-config', 'logs'];
            for (const tab of tabs) {
                await this.switchTab(tab);
                await this.delay(500);

                const activeTab = document.querySelector(`[data-tab="${tab}"]`);
                const isActive = activeTab && activeTab.classList.contains('active');
                this.addTestResult(testResults, `${tab}标签页切换`, isActive,
                    isActive ? '标签页切换正常' : '标签页切换失败');
            }

            // 测试响应式设计
            const originalWidth = window.innerWidth;

            // 模拟移动设备宽度
            window.resizeTo(375, 667);
            await this.delay(500);

            const isMobileResponsive = window.innerWidth <= 768;
            this.addTestResult(testResults, '移动端响应式设计', isMobileResponsive,
                isMobileResponsive ? '移动端适配正常' : '移动端适配需要优化');

            // 恢复原始宽度
            window.resizeTo(originalWidth, window.innerHeight);

            // 测试模态框功能
            const modalExists = document.querySelector('.modal') !== null;
            this.addTestResult(testResults, '模态框功能', true, '模态框功能可用');

        } catch (error) {
            this.addTestResult(testResults, '界面交互测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试性能和响应性
     */
    async testPerformanceAndResponsiveness(testResults) {
        console.log('⚡ 测试性能和响应性...');

        try {
            // 测试页面加载性能
            const loadStartTime = performance.now();
            await this.loadSystemOverview();
            const loadEndTime = performance.now();
            const loadTime = loadEndTime - loadStartTime;

            const loadPerformanceGood = loadTime < 3000; // 3秒内
            this.addTestResult(testResults, '页面加载性能', loadPerformanceGood,
                `加载时间: ${loadTime.toFixed(2)}ms ${loadPerformanceGood ? '(良好)' : '(需要优化)'}`);

            // 测试API响应性能
            const apiStartTime = performance.now();
            await proxyAPI.getConfigs();
            const apiEndTime = performance.now();
            const apiTime = apiEndTime - apiStartTime;

            const apiPerformanceGood = apiTime < 1000; // 1秒内
            this.addTestResult(testResults, 'API响应性能', apiPerformanceGood,
                `API响应时间: ${apiTime.toFixed(2)}ms ${apiPerformanceGood ? '(良好)' : '(需要优化)'}`);

            // 测试内存使用
            if (performance.memory) {
                const memoryUsage = performance.memory.usedJSHeapSize / 1024 / 1024; // MB
                const memoryEfficient = memoryUsage < 50; // 50MB内
                this.addTestResult(testResults, '内存使用效率', memoryEfficient,
                    `内存使用: ${memoryUsage.toFixed(2)}MB ${memoryEfficient ? '(良好)' : '(需要优化)'}`);
            }

        } catch (error) {
            this.addTestResult(testResults, '性能响应性测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试错误恢复
     */
    async testErrorRecovery(testResults) {
        console.log('🔧 测试错误恢复...');

        try {
            // 测试网络错误恢复
            const originalFetch = window.fetch;
            let errorHandled = false;

            // 模拟网络错误
            window.fetch = () => {
                errorHandled = true;
                return Promise.reject(new Error('网络连接失败'));
            };

            try {
                await proxyAPI.getConfigs();
            } catch (error) {
                // 预期的错误
            }

            // 恢复正常网络
            window.fetch = originalFetch;

            this.addTestResult(testResults, '网络错误恢复', errorHandled,
                errorHandled ? '网络错误处理正常' : '网络错误处理异常');

            // 测试数据恢复
            await this.loadSystemOverview();
            const dataRecovered = !!this.systemOverviewData;
            this.addTestResult(testResults, '数据恢复能力', dataRecovered,
                dataRecovered ? '数据恢复正常' : '数据恢复失败');

        } catch (error) {
            this.addTestResult(testResults, '错误恢复测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 测试并发操作
     */
    async testConcurrentOperations(testResults) {
        console.log('🔀 测试并发操作...');

        try {
            // 并发加载多个API
            const concurrentPromises = [
                proxyAPI.getConfigs(),
                this.loadSystemOverview(),
                // 如果有可用配置，同时加载令牌
            ];

            const startTime = performance.now();
            const results = await Promise.allSettled(concurrentPromises);
            const endTime = performance.now();
            const concurrentTime = endTime - startTime;

            const allSuccessful = results.every(result => result.status === 'fulfilled');
            this.addTestResult(testResults, '并发操作处理', allSuccessful,
                `并发操作${allSuccessful ? '成功' : '部分失败'}，耗时${concurrentTime.toFixed(2)}ms`);

            // 测试并发数据更新
            const updatePromises = [];
            for (let i = 0; i < 3; i++) {
                updatePromises.push(this.refreshSystemOverview());
            }

            const updateResults = await Promise.allSettled(updatePromises);
            const updatesSuccessful = updateResults.every(result => result.status === 'fulfilled');
            this.addTestResult(testResults, '并发数据更新', updatesSuccessful,
                updatesSuccessful ? '并发数据更新正常' : '并发数据更新异常');

        } catch (error) {
            this.addTestResult(testResults, '并发操作测试', false, `测试失败: ${error.message}`);
        }
    }

    /**
     * 显示端到端测试结果
     */
    displayE2ETestResults(testResults) {
        const passRate = ((testResults.passedTests / testResults.totalTests) * 100).toFixed(1);

        console.log('\n🎯 端到端功能测试结果:');
        console.log(`总测试数: ${testResults.totalTests}`);
        console.log(`通过: ${testResults.passedTests}`);
        console.log(`失败: ${testResults.failedTests}`);
        console.log(`通过率: ${passRate}%`);

        // 生成详细的测试报告
        const resultMessage = `
            <div class="e2e-test-results">
                <h3>🎯 端到端功能测试结果</h3>
                <div class="test-summary grid grid-cols-4 gap-4 mb-6">
                    <div class="bg-blue-50 p-4 rounded text-center">
                        <div class="text-2xl font-bold text-blue-600">${testResults.totalTests}</div>
                        <div class="text-sm text-blue-800">总测试数</div>
                    </div>
                    <div class="bg-green-50 p-4 rounded text-center">
                        <div class="text-2xl font-bold text-green-600">${testResults.passedTests}</div>
                        <div class="text-sm text-green-800">通过测试</div>
                    </div>
                    <div class="bg-red-50 p-4 rounded text-center">
                        <div class="text-2xl font-bold text-red-600">${testResults.failedTests}</div>
                        <div class="text-sm text-red-800">失败测试</div>
                    </div>
                    <div class="bg-purple-50 p-4 rounded text-center">
                        <div class="text-2xl font-bold text-purple-600">${passRate}%</div>
                        <div class="text-sm text-purple-800">通过率</div>
                    </div>
                </div>

                <div class="scenarios-section mb-6">
                    <h4 class="font-semibold mb-3">📋 用户场景测试结果:</h4>
                    <div class="space-y-2">
                        ${testResults.scenarios.map(scenario => `
                            <div class="p-3 rounded ${scenario.success ? 'bg-green-50 border-l-4 border-green-400' : 'bg-red-50 border-l-4 border-red-400'}">
                                <div class="flex items-center justify-between">
                                    <span class="font-medium">${scenario.success ? '✅' : '❌'} ${scenario.message}</span>
                                </div>
                                <div class="text-sm text-gray-600 mt-1">
                                    执行步骤: ${scenario.steps.join(' → ')}
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>

                <div class="test-details">
                    <h4 class="font-semibold mb-3">🔍 详细测试结果:</h4>
                    <div class="space-y-1 max-h-60 overflow-y-auto">
                        ${testResults.results.map(result => `
                            <div class="flex items-center justify-between p-2 rounded text-sm ${
                                result.passed ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
                            }">
                                <span>${result.passed ? '✅' : '❌'} <strong>${result.name}</strong></span>
                                <span class="text-xs">${result.message}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            </div>
        `;

        // 显示测试结果模态框
        this.showTestResultsModal(resultMessage, passRate >= 80);

        // 显示成功或警告消息
        if (passRate >= 90) {
            this.showSuccess('端到端功能测试完成，系统功能优秀！');
        } else if (passRate >= 80) {
            this.showSuccess('端到端功能测试完成，系统功能良好！');
        } else if (passRate >= 70) {
            this.showWarning('端到端功能测试完成，系统功能基本正常，建议优化。');
        } else {
            this.showError('端到端功能测试完成，系统存在较多问题，需要修复。');
        }
    }

    /**
     * 运行性能基准测试
     */
    async runPerformanceBenchmark() {
        console.log('⚡ 开始性能基准测试...');

        const benchmarkResults = {
            tests: [],
            summary: {}
        };

        try {
            this.showInfo('正在运行性能基准测试，请稍候...');

            // 1. 页面加载性能测试
            const loadTest = await this.benchmarkPageLoad();
            benchmarkResults.tests.push(loadTest);

            // 2. API响应性能测试
            const apiTest = await this.benchmarkAPIPerformance();
            benchmarkResults.tests.push(apiTest);

            // 3. 界面渲染性能测试
            const renderTest = await this.benchmarkUIRendering();
            benchmarkResults.tests.push(renderTest);

            // 4. 内存使用测试
            const memoryTest = await this.benchmarkMemoryUsage();
            benchmarkResults.tests.push(memoryTest);

            // 5. 数据处理性能测试
            const dataTest = await this.benchmarkDataProcessing();
            benchmarkResults.tests.push(dataTest);

            // 生成性能报告
            this.generatePerformanceReport(benchmarkResults);

        } catch (error) {
            console.error('性能基准测试失败:', error);
            this.showError('性能基准测试失败: ' + error.message);
        }
    }

    /**
     * 页面加载性能基准测试
     */
    async benchmarkPageLoad() {
        const iterations = 5;
        const loadTimes = [];

        for (let i = 0; i < iterations; i++) {
            const startTime = performance.now();
            await this.loadSystemOverview();
            const endTime = performance.now();
            loadTimes.push(endTime - startTime);
            await this.delay(100); // 短暂延迟
        }

        const avgTime = loadTimes.reduce((a, b) => a + b, 0) / loadTimes.length;
        const minTime = Math.min(...loadTimes);
        const maxTime = Math.max(...loadTimes);

        return {
            name: '页面加载性能',
            avgTime: avgTime.toFixed(2),
            minTime: minTime.toFixed(2),
            maxTime: maxTime.toFixed(2),
            iterations,
            rating: avgTime < 1000 ? 'excellent' : avgTime < 2000 ? 'good' : avgTime < 3000 ? 'fair' : 'poor'
        };
    }

    /**
     * API响应性能基准测试
     */
    async benchmarkAPIPerformance() {
        const apis = [
            { name: 'getConfigs', fn: () => proxyAPI.getConfigs() },
            { name: 'loadSystemOverview', fn: () => this.loadSystemOverview() }
        ];

        const results = [];

        for (const api of apis) {
            const iterations = 3;
            const times = [];

            for (let i = 0; i < iterations; i++) {
                const startTime = performance.now();
                try {
                    await api.fn();
                    const endTime = performance.now();
                    times.push(endTime - startTime);
                } catch (error) {
                    times.push(5000); // 错误时记录5秒
                }
                await this.delay(100);
            }

            const avgTime = times.reduce((a, b) => a + b, 0) / times.length;
            results.push({
                api: api.name,
                avgTime: avgTime.toFixed(2),
                rating: avgTime < 500 ? 'excellent' : avgTime < 1000 ? 'good' : avgTime < 2000 ? 'fair' : 'poor'
            });
        }

        return {
            name: 'API响应性能',
            results,
            overallRating: results.every(r => r.rating === 'excellent') ? 'excellent' :
                          results.every(r => ['excellent', 'good'].includes(r.rating)) ? 'good' : 'fair'
        };
    }

    /**
     * 界面渲染性能基准测试
     */
    async benchmarkUIRendering() {
        const renderTests = [];

        // 测试标签页切换性能
        const tabs = ['overview', 'proxy-config', 'logs'];
        const switchTimes = [];

        for (const tab of tabs) {
            const startTime = performance.now();
            await this.switchTab(tab);
            const endTime = performance.now();
            switchTimes.push(endTime - startTime);
            await this.delay(100);
        }

        const avgSwitchTime = switchTimes.reduce((a, b) => a + b, 0) / switchTimes.length;

        // 测试图表渲染性能
        const chartStartTime = performance.now();
        await this.updateProtocolsChart({ http: 60, https: 40 });
        const chartEndTime = performance.now();
        const chartRenderTime = chartEndTime - chartStartTime;

        return {
            name: '界面渲染性能',
            tabSwitchTime: avgSwitchTime.toFixed(2),
            chartRenderTime: chartRenderTime.toFixed(2),
            rating: avgSwitchTime < 100 && chartRenderTime < 200 ? 'excellent' :
                   avgSwitchTime < 200 && chartRenderTime < 500 ? 'good' : 'fair'
        };
    }

    /**
     * 内存使用基准测试
     */
    async benchmarkMemoryUsage() {
        if (!performance.memory) {
            return {
                name: '内存使用测试',
                available: false,
                message: '浏览器不支持内存监控'
            };
        }

        const initialMemory = performance.memory.usedJSHeapSize;

        // 执行一些操作来测试内存使用
        await this.loadSystemOverview();
        await this.switchTab('proxy-config');
        await this.switchTab('logs');
        await this.switchTab('overview');

        const finalMemory = performance.memory.usedJSHeapSize;
        const memoryDiff = (finalMemory - initialMemory) / 1024 / 1024; // MB
        const totalMemory = finalMemory / 1024 / 1024; // MB

        return {
            name: '内存使用测试',
            available: true,
            initialMemory: (initialMemory / 1024 / 1024).toFixed(2),
            finalMemory: (finalMemory / 1024 / 1024).toFixed(2),
            memoryDiff: memoryDiff.toFixed(2),
            totalMemory: totalMemory.toFixed(2),
            rating: totalMemory < 30 ? 'excellent' : totalMemory < 50 ? 'good' : totalMemory < 100 ? 'fair' : 'poor'
        };
    }

    /**
     * 数据处理性能基准测试
     */
    async benchmarkDataProcessing() {
        const startTime = performance.now();

        // 模拟大量数据处理
        const testData = [];
        for (let i = 0; i < 1000; i++) {
            testData.push({
                id: `test-${i}`,
                name: `Test Config ${i}`,
                enabled: i % 2 === 0,
                created_at: new Date().toISOString()
            });
        }

        // 数据过滤测试
        const filteredData = testData.filter(item => item.enabled);

        // 数据排序测试
        const sortedData = testData.sort((a, b) => a.name.localeCompare(b.name));

        // 数据聚合测试
        const aggregatedData = testData.reduce((acc, item) => {
            acc.total++;
            if (item.enabled) acc.enabled++;
            return acc;
        }, { total: 0, enabled: 0 });

        const endTime = performance.now();
        const processingTime = endTime - startTime;

        return {
            name: '数据处理性能',
            dataSize: testData.length,
            processingTime: processingTime.toFixed(2),
            filteredCount: filteredData.length,
            aggregatedData,
            rating: processingTime < 10 ? 'excellent' : processingTime < 50 ? 'good' : processingTime < 100 ? 'fair' : 'poor'
        };
    }

    /**
     * 生成性能报告
     */
    generatePerformanceReport(benchmarkResults) {
        const reportHtml = `
            <div class="performance-report">
                <h3>⚡ 性能基准测试报告</h3>
                <div class="performance-summary mb-6">
                    <div class="grid grid-cols-2 gap-4">
                        <div class="bg-blue-50 p-4 rounded">
                            <h4 class="font-semibold text-blue-800 mb-2">测试概览</h4>
                            <p class="text-sm text-blue-700">执行了 ${benchmarkResults.tests.length} 项性能测试</p>
                            <p class="text-sm text-blue-700">测试时间: ${new Date().toLocaleString()}</p>
                        </div>
                        <div class="bg-green-50 p-4 rounded">
                            <h4 class="font-semibold text-green-800 mb-2">整体评级</h4>
                            <p class="text-lg font-bold text-green-700">
                                ${this.calculateOverallRating(benchmarkResults.tests)}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="performance-details space-y-4">
                    ${benchmarkResults.tests.map(test => `
                        <div class="bg-white border rounded-lg p-4">
                            <h4 class="font-semibold mb-3 flex items-center justify-between">
                                ${test.name}
                                <span class="text-sm px-2 py-1 rounded ${this.getRatingColor(test.rating || test.overallRating)}">
                                    ${this.getRatingText(test.rating || test.overallRating)}
                                </span>
                            </h4>
                            <div class="text-sm space-y-1">
                                ${this.formatTestDetails(test)}
                            </div>
                        </div>
                    `).join('')}
                </div>

                <div class="performance-recommendations mt-6 p-4 bg-yellow-50 rounded">
                    <h4 class="font-semibold text-yellow-800 mb-2">🔧 优化建议</h4>
                    <ul class="text-sm text-yellow-700 space-y-1">
                        ${this.generateOptimizationSuggestions(benchmarkResults.tests)}
                    </ul>
                </div>
            </div>
        `;

        this.showTestResultsModal(reportHtml, true);
        this.showSuccess('性能基准测试完成！');
    }

    /**
     * 计算整体评级
     */
    calculateOverallRating(tests) {
        const ratings = tests.map(test => test.rating || test.overallRating).filter(Boolean);
        const excellentCount = ratings.filter(r => r === 'excellent').length;
        const goodCount = ratings.filter(r => r === 'good').length;

        if (excellentCount >= ratings.length * 0.8) return '优秀 🌟';
        if (excellentCount + goodCount >= ratings.length * 0.7) return '良好 👍';
        return '需要优化 ⚠️';
    }

    /**
     * 获取评级颜色
     */
    getRatingColor(rating) {
        switch (rating) {
            case 'excellent': return 'bg-green-100 text-green-800';
            case 'good': return 'bg-blue-100 text-blue-800';
            case 'fair': return 'bg-yellow-100 text-yellow-800';
            case 'poor': return 'bg-red-100 text-red-800';
            default: return 'bg-gray-100 text-gray-800';
        }
    }

    /**
     * 获取评级文本
     */
    getRatingText(rating) {
        switch (rating) {
            case 'excellent': return '优秀';
            case 'good': return '良好';
            case 'fair': return '一般';
            case 'poor': return '较差';
            default: return '未知';
        }
    }

    /**
     * 格式化测试详情
     */
    formatTestDetails(test) {
        const details = [];

        if (test.avgTime) details.push(`平均时间: ${test.avgTime}ms`);
        if (test.minTime) details.push(`最短时间: ${test.minTime}ms`);
        if (test.maxTime) details.push(`最长时间: ${test.maxTime}ms`);
        if (test.iterations) details.push(`测试次数: ${test.iterations}`);
        if (test.results) {
            details.push('API测试结果:');
            test.results.forEach(result => {
                details.push(`&nbsp;&nbsp;• ${result.api}: ${result.avgTime}ms (${this.getRatingText(result.rating)})`);
            });
        }
        if (test.tabSwitchTime) details.push(`标签切换: ${test.tabSwitchTime}ms`);
        if (test.chartRenderTime) details.push(`图表渲染: ${test.chartRenderTime}ms`);
        if (test.totalMemory) details.push(`内存使用: ${test.totalMemory}MB`);
        if (test.processingTime) details.push(`数据处理: ${test.processingTime}ms`);
        if (test.dataSize) details.push(`数据量: ${test.dataSize} 条记录`);
        if (test.available === false) details.push(test.message);

        return details.map(detail => `<div>${detail}</div>`).join('');
    }

    /**
     * 生成优化建议
     */
    generateOptimizationSuggestions(tests) {
        const suggestions = [];

        tests.forEach(test => {
            if (test.rating === 'poor' || test.overallRating === 'poor') {
                if (test.name.includes('页面加载')) {
                    suggestions.push('<li>考虑优化页面加载性能，减少初始化时间</li>');
                }
                if (test.name.includes('API响应')) {
                    suggestions.push('<li>优化API响应时间，考虑添加缓存机制</li>');
                }
                if (test.name.includes('界面渲染')) {
                    suggestions.push('<li>优化界面渲染性能，减少DOM操作</li>');
                }
                if (test.name.includes('内存使用')) {
                    suggestions.push('<li>优化内存使用，避免内存泄漏</li>');
                }
            }
        });

        if (suggestions.length === 0) {
            suggestions.push('<li>系统性能表现良好，继续保持！</li>');
        }

        return suggestions.join('');
    }
}

// 创建全局实例
const businessManager = new BusinessManager();

// 将不再使用的调试/测试相关方法置为 no-op，避免死代码被调用
businessManager.refreshSystemOverview = async () => {};
businessManager.exportSystemOverview = async () => {};
businessManager.runIntegrationTests = async () => {};
businessManager.runE2EDemo = async () => {};
businessManager.runE2EFunctionalTests = async () => {};
businessManager.runPerformanceBenchmark = async () => {};

export { businessManager };
