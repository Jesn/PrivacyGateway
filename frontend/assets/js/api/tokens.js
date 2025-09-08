/**
 * 令牌API模块
 * 负责令牌的CRUD操作、状态管理等功能
 */

import { httpClient } from '../utils/http.js';

class TokensAPI {
    constructor() {
        this.baseURL = '/config/proxy';
    }

    /**
     * 获取令牌列表
     * @param {string} configId - 配置ID
     * @returns {Promise<Object>} 令牌响应数据，包含tokens和stats
     */
    async getTokens(configId) {
        try {
            if (!configId) {
                throw new Error('配置ID不能为空');
            }

            const response = await httpClient.get(`${this.baseURL}/${configId}/tokens`);
            // httpClient返回 { data: {...}, status: 200, ... }
            // 后端API返回 { success: true, data: { tokens: [...], stats: {...} }, status: 200 }
            // 所以需要访问 response.data.data
            return response.data?.data || { tokens: [], stats: null };
        } catch (error) {
            console.error('获取令牌列表失败:', error);
            throw new Error('获取令牌列表失败: ' + error.message);
        }
    }

    /**
     * 获取令牌详情
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @returns {Promise<Object>} 令牌详情
     */
    async getToken(configId, tokenId) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const response = await httpClient.get(`${this.baseURL}/${configId}/tokens/${tokenId}`);
            return response.data?.data || response.data;
        } catch (error) {
            console.error('获取令牌详情失败:', error);
            throw new Error('获取令牌详情失败: ' + error.message);
        }
    }

    /**
     * 创建令牌
     * @param {string} configId - 配置ID
     * @param {Object} tokenData - 令牌数据
     * @returns {Promise<Object>} 创建结果
     */
    async createToken(configId, tokenData) {
        try {
            if (!configId) {
                throw new Error('配置ID不能为空');
            }

            // 验证令牌数据
            this.validateTokenData(tokenData);

            const response = await httpClient.post(`${this.baseURL}/${configId}/tokens`, tokenData);
            return response.data;
        } catch (error) {
            console.error('创建令牌失败:', error);
            throw new Error('创建令牌失败: ' + error.message);
        }
    }

    /**
     * 更新令牌
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @param {Object} tokenData - 令牌数据
     * @returns {Promise<Object>} 更新结果
     */
    async updateToken(configId, tokenId, tokenData) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            // 验证令牌数据
            this.validateTokenData(tokenData);

            const response = await httpClient.put(`${this.baseURL}/${configId}/tokens/${tokenId}`, tokenData);
            return response.data;
        } catch (error) {
            console.error('更新令牌失败:', error);
            throw new Error('更新令牌失败: ' + error.message);
        }
    }

    /**
     * 删除令牌
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @returns {Promise<Object>} 删除结果
     */
    async deleteToken(configId, tokenId) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const response = await httpClient.delete(`${this.baseURL}/${configId}/tokens/${tokenId}`);
            return response.data;
        } catch (error) {
            console.error('删除令牌失败:', error);
            throw new Error('删除令牌失败: ' + error.message);
        }
    }

    /**
     * 切换令牌状态
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @param {boolean} enabled - 是否启用
     * @returns {Promise<Object>} 操作结果
     */
    async toggleToken(configId, tokenId, enabled) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const response = await httpClient.post(`${this.baseURL}/${configId}/tokens/${tokenId}/toggle`, {
                enabled
            });
            return response.data;
        } catch (error) {
            console.error('切换令牌状态失败:', error);
            throw new Error('切换令牌状态失败: ' + error.message);
        }
    }

    /**
     * 重新生成令牌
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @returns {Promise<Object>} 新令牌数据
     */
    async regenerateToken(configId, tokenId) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const response = await httpClient.post(`${this.baseURL}/${configId}/tokens/${tokenId}/regenerate`);
            return response.data;
        } catch (error) {
            console.error('重新生成令牌失败:', error);
            throw new Error('重新生成令牌失败: ' + error.message);
        }
    }

    /**
     * 获取令牌使用统计
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @param {Object} options - 统计选项
     * @returns {Promise<Object>} 使用统计
     */
    async getTokenStats(configId, tokenId, options = {}) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const {
                period = '7d', // 7d, 30d, 90d
                groupBy = 'day' // hour, day
            } = options;

            const params = { period, group_by: groupBy };
            const response = await httpClient.get(`${this.baseURL}/${configId}/tokens/${tokenId}/stats`, params);
            return response.data;
        } catch (error) {
            console.error('获取令牌统计失败:', error);
            throw new Error('获取令牌统计失败: ' + error.message);
        }
    }

    /**
     * 获取令牌使用示例
     * @param {string} configId - 配置ID
     * @param {string} tokenId - 令牌ID
     * @returns {Promise<Object>} 使用示例
     */
    async getTokenUsageExamples(configId, tokenId) {
        try {
            if (!configId || !tokenId) {
                throw new Error('配置ID和令牌ID不能为空');
            }

            const response = await httpClient.get(`${this.baseURL}/${configId}/tokens/${tokenId}/examples`);
            return response.data;
        } catch (error) {
            console.error('获取使用示例失败:', error);
            throw new Error('获取使用示例失败: ' + error.message);
        }
    }

    /**
     * 验证令牌数据
     * @private
     * @param {Object} tokenData - 令牌数据
     */
    validateTokenData(tokenData) {
        if (!tokenData) {
            throw new Error('令牌数据不能为空');
        }

        if (!tokenData.name || tokenData.name.trim() === '') {
            throw new Error('令牌名称不能为空');
        }

        // 验证过期时间
        if (tokenData.expires_at) {
            const expiryDate = new Date(tokenData.expires_at);
            if (isNaN(expiryDate.getTime())) {
                throw new Error('过期时间格式不正确');
            }
            if (expiryDate <= new Date()) {
                throw new Error('过期时间不能早于当前时间');
            }
        }

        // 验证速率限制
        if (tokenData.rate_limit) {
            const rateLimit = parseInt(tokenData.rate_limit);
            if (isNaN(rateLimit) || rateLimit < 0) {
                throw new Error('速率限制必须是非负数');
            }
        }

        // 验证IP白名单
        if (tokenData.allowed_ips && Array.isArray(tokenData.allowed_ips)) {
            tokenData.allowed_ips.forEach(ip => {
                if (!this.isValidIP(ip)) {
                    throw new Error(`无效的IP地址: ${ip}`);
                }
            });
        }
    }

    /**
     * 验证IP地址格式
     * @private
     * @param {string} ip - IP地址
     * @returns {boolean} 是否有效
     */
    isValidIP(ip) {
        // 简单的IP地址验证
        const ipRegex = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/;
        if (!ipRegex.test(ip)) {
            return false;
        }

        const parts = ip.split('/')[0].split('.');
        return parts.every(part => {
            const num = parseInt(part);
            return num >= 0 && num <= 255;
        });
    }

    /**
     * 格式化令牌数据用于显示
     * @param {Object} token - 原始令牌数据
     * @returns {Object} 格式化后的令牌数据
     */
    formatTokenForDisplay(token) {
        return {
            ...token,
            status_display: token.enabled ? '启用' : '禁用',
            status_class: token.enabled ? 'status-enabled' : 'status-disabled',
            created_at_display: token.created_at 
                ? new Date(token.created_at).toLocaleString('zh-CN')
                : '-',
            expires_at_display: token.expires_at 
                ? new Date(token.expires_at).toLocaleString('zh-CN')
                : '永不过期',
            last_used_display: token.last_used_at 
                ? new Date(token.last_used_at).toLocaleString('zh-CN')
                : '从未使用',
            rate_limit_display: token.rate_limit 
                ? `${token.rate_limit}/小时`
                : '无限制',
            allowed_ips_display: Array.isArray(token.allowed_ips) && token.allowed_ips.length > 0
                ? token.allowed_ips.join(', ')
                : '不限制',
            usage_count_display: token.usage_count || 0,
            token_preview: token.token 
                ? `${token.token.substring(0, 8)}...${token.token.substring(token.token.length - 4)}`
                : '-'
        };
    }

    /**
     * 生成令牌使用示例代码
     * @param {Object} config - 配置信息
     * @param {Object} token - 令牌信息
     * @returns {Object} 使用示例代码
     */
    generateUsageExamples(config, token) {
        const baseURL = window.location.origin;
        const proxyPath = config.path || '/';
        const tokenValue = token.token;

        return {
            curl: `curl -X GET "${baseURL}${proxyPath}" \\
  -H "Authorization: Bearer ${tokenValue}" \\
  -H "Content-Type: application/json"`,
            
            javascript: `fetch('${baseURL}${proxyPath}', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ${tokenValue}',
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => console.log(data))
.catch(error => console.error('Error:', error));`,
            
            python: `import requests

url = "${baseURL}${proxyPath}"
headers = {
    "Authorization": "Bearer ${tokenValue}",
    "Content-Type": "application/json"
}

response = requests.get(url, headers=headers)
print(response.json())`,
            
            nodejs: `const axios = require('axios');

const config = {
  method: 'get',
  url: '${baseURL}${proxyPath}',
  headers: {
    'Authorization': 'Bearer ${tokenValue}',
    'Content-Type': 'application/json'
  }
};

axios(config)
.then(response => {
  console.log(response.data);
})
.catch(error => {
  console.log(error);
});`
        };
    }
}

// 创建全局实例
export const tokensAPI = new TokensAPI();

// 导出类以便扩展
export { TokensAPI };
