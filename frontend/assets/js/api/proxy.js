/**
 * 代理配置API模块
 * 负责代理配置的CRUD操作、批量操作、导入导出等功能
 */

import { httpClient } from '../utils/http.js';

class ProxyAPI {
    constructor() {
        this.baseURL = '/config/proxy';
    }

    /**
     * 获取代理配置列表
     * @param {number} page - 页码
     * @param {number} limit - 每页数量
     * @param {string} search - 搜索关键词
     * @returns {Promise<Object>} 配置列表响应
     */
    async getConfigs(page = 1, limit = 10, search = '') {
        try {
            const params = { page, limit };
            if (search) {
                params.search = search;
            }

            const response = await httpClient.get(this.baseURL, params);
            return {
                configs: response.data.configs || [],
                total: response.data.total || 0,
                page: response.data.page || page,
                totalPages: Math.ceil((response.data.total || 0) / limit)
            };
        } catch (error) {
            console.error('获取代理配置失败:', error);
            throw new Error('获取代理配置失败: ' + error.message);
        }
    }

    /**
     * 获取单个代理配置
     * @param {string} id - 配置ID
     * @returns {Promise<Object>} 配置详情
     */
    async getConfig(id) {
        try {
            const response = await httpClient.get(`${this.baseURL}/${id}`);
            return response.data;
        } catch (error) {
            console.error('获取配置详情失败:', error);
            throw new Error('获取配置详情失败: ' + error.message);
        }
    }

    /**
     * 创建代理配置
     * @param {Object} config - 配置数据
     * @returns {Promise<Object>} 创建结果
     */
    async createConfig(config) {
        try {
            // 验证必填字段
            this.validateConfig(config);

            const response = await httpClient.post(this.baseURL, config);
            return response.data;
        } catch (error) {
            console.error('创建配置失败:', error);
            throw new Error('创建配置失败: ' + error.message);
        }
    }

    /**
     * 更新代理配置
     * @param {string} id - 配置ID
     * @param {Object} config - 配置数据
     * @returns {Promise<Object>} 更新结果
     */
    async updateConfig(id, config) {
        try {
            // 验证必填字段
            this.validateConfig(config);

            const response = await httpClient.put(`${this.baseURL}?id=${id}`, config);
            return response.data;
        } catch (error) {
            console.error('更新配置失败:', error);
            throw new Error('更新配置失败: ' + error.message);
        }
    }

    /**
     * 删除代理配置
     * @param {string} id - 配置ID
     * @returns {Promise<Object>} 删除结果
     */
    async deleteConfig(id) {
        try {
            const response = await httpClient.delete(`${this.baseURL}?id=${id}`);
            return response.data;
        } catch (error) {
            console.error('删除配置失败:', error);
            throw new Error('删除配置失败: ' + error.message);
        }
    }

    /**
     * 批量删除代理配置
     * @param {Array<string>} ids - 配置ID数组
     * @returns {Promise<Object>} 删除结果
     */
    async batchDeleteConfigs(ids) {
        try {
            if (!Array.isArray(ids) || ids.length === 0) {
                throw new Error('请选择要删除的配置');
            }

            const response = await httpClient.post(`${this.baseURL}/batch-delete`, { ids });
            return response.data;
        } catch (error) {
            console.error('批量删除配置失败:', error);
            throw new Error('批量删除配置失败: ' + error.message);
        }
    }

    /**
     * 批量更新配置
     * @param {Array<string>} ids - 配置ID数组
     * @param {Object} updates - 更新的字段
     * @returns {Promise<Object>} 更新结果
     */
    async batchUpdateConfigs(ids, updates) {
        try {
            if (!Array.isArray(ids) || ids.length === 0) {
                throw new Error('请选择要更新的配置');
            }

            const response = await httpClient.post(`${this.baseURL}/batch-update`, {
                ids,
                updates
            });
            return response.data;
        } catch (error) {
            console.error('批量更新配置失败:', error);
            throw new Error('批量更新配置失败: ' + error.message);
        }
    }

    /**
     * 批量启用/禁用配置
     * @param {Array<string>} ids - 配置ID数组
     * @param {boolean} enabled - 是否启用
     * @returns {Promise<Object>} 操作结果
     */
    async batchToggleConfigs(ids, enabled) {
        try {
            if (!Array.isArray(ids) || ids.length === 0) {
                throw new Error('请选择要操作的配置');
            }

            const response = await httpClient.post(`${this.baseURL}/batch-toggle`, { 
                ids, 
                enabled 
            });
            return response.data;
        } catch (error) {
            console.error('批量操作配置失败:', error);
            throw new Error('批量操作配置失败: ' + error.message);
        }
    }

    /**
     * 测试代理配置
     * @param {string} id - 配置ID
     * @param {Object} testData - 测试数据
     * @returns {Promise<Object>} 测试结果
     */
    async testConfig(id, testData = {}) {
        try {
            const response = await httpClient.post(`${this.baseURL}/test`, {
                id,
                name: testData.name || '配置测试',
                target_url: testData.target_url,
                protocol: testData.protocol || 'https',
                method: testData.method || 'GET',
                path: testData.path || '/',
                headers: testData.headers || {},
                body: testData.body || ''
            });
            return response.data;
        } catch (error) {
            console.error('测试配置失败:', error);
            throw new Error('测试配置失败: ' + error.message);
        }
    }

    /**
     * 测试连接（不需要配置ID，直接测试URL）
     * @param {Object} config - 配置对象
     * @returns {Promise<Object>} 测试结果
     */
    async testConnection(config) {
        try {
            const response = await httpClient.post(`${this.baseURL}/test-connection`, {
                name: config.name || '临时测试配置',
                target_url: config.target_url,
                protocol: config.protocol || 'https',
                timeout: config.timeout || 30,
                verify_ssl: config.verify_ssl !== false
            });
            return response.data;
        } catch (error) {
            console.error('测试连接失败:', error);
            throw new Error('测试连接失败: ' + error.message);
        }
    }

    /**
     * 导出配置
     * @param {Array<string>} ids - 配置ID数组，为空则导出全部
     * @returns {Promise<Object>} 导出数据
     */
    async exportConfigs(ids = []) {
        try {
            const params = ids.length > 0 ? { ids: ids.join(',') } : {};
            const response = await httpClient.get(`${this.baseURL}/export`, params);
            return response.data;
        } catch (error) {
            console.error('导出配置失败:', error);
            throw new Error('导出配置失败: ' + error.message);
        }
    }

    /**
     * 导入配置
     * @param {Array<Object>} configs - 配置数据数组
     * @param {boolean} overwrite - 是否覆盖同名配置
     * @returns {Promise<Object>} 导入结果
     */
    async importConfigs(configs, overwrite = false) {
        try {
            if (!Array.isArray(configs) || configs.length === 0) {
                throw new Error('导入数据不能为空');
            }

            // 验证每个配置
            configs.forEach((config, index) => {
                try {
                    this.validateConfig(config);
                } catch (error) {
                    throw new Error(`第 ${index + 1} 个配置验证失败: ${error.message}`);
                }
            });

            const response = await httpClient.post(`${this.baseURL}/import`, {
                configs,
                overwrite
            });
            return response.data;
        } catch (error) {
            console.error('导入配置失败:', error);
            throw new Error('导入配置失败: ' + error.message);
        }
    }

    /**
     * 获取配置统计信息
     * @returns {Promise<Object>} 统计数据
     */
    async getStats() {
        try {
            const response = await httpClient.get(`${this.baseURL}/stats`);
            return response.data;
        } catch (error) {
            console.error('获取统计信息失败:', error);
            throw new Error('获取统计信息失败: ' + error.message);
        }
    }

    /**
     * 验证配置数据
     * @private
     * @param {Object} config - 配置数据
     */
    validateConfig(config) {
        if (!config) {
            throw new Error('配置数据不能为空');
        }

        if (!config.name || config.name.trim() === '') {
            throw new Error('配置名称不能为空');
        }

        if (!config.target_url || config.target_url.trim() === '') {
            throw new Error('目标URL不能为空');
        }

        // 验证URL格式
        try {
            new URL(config.target_url);
        } catch (error) {
            throw new Error('目标URL格式不正确');
        }

        // 验证路径格式
        if (config.path && !config.path.startsWith('/')) {
            throw new Error('路径必须以 / 开头');
        }

        // 验证HTTP方法
        const validMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];
        if (config.allowed_methods && Array.isArray(config.allowed_methods)) {
            const invalidMethods = config.allowed_methods.filter(method => 
                !validMethods.includes(method.toUpperCase())
            );
            if (invalidMethods.length > 0) {
                throw new Error(`不支持的HTTP方法: ${invalidMethods.join(', ')}`);
            }
        }

        // 验证超时时间
        if (config.timeout && (isNaN(config.timeout) || config.timeout < 0)) {
            throw new Error('超时时间必须是非负数');
        }

        // 验证重试次数
        if (config.retry_count && (isNaN(config.retry_count) || config.retry_count < 0)) {
            throw new Error('重试次数必须是非负数');
        }
    }

    /**
     * 格式化配置数据用于显示
     * @param {Object} config - 原始配置数据
     * @returns {Object} 格式化后的配置数据
     */
    formatConfigForDisplay(config) {
        return {
            ...config,
            allowed_methods_display: Array.isArray(config.allowed_methods) 
                ? config.allowed_methods.join(', ') 
                : 'ALL',
            status_display: config.enabled ? '启用' : '禁用',
            created_at_display: config.created_at 
                ? new Date(config.created_at).toLocaleString('zh-CN')
                : '-',
            updated_at_display: config.updated_at 
                ? new Date(config.updated_at).toLocaleString('zh-CN')
                : '-'
        };
    }
}

// 创建全局实例
export const proxyAPI = new ProxyAPI();

// 导出类以便扩展
export { ProxyAPI };
