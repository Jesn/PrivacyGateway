/**
 * 日志API模块
 * 负责访问日志的查询、筛选、详情查看等功能
 */

import { httpClient } from '../utils/http.js';

class LogsAPI {
    constructor() {
        this.baseURL = '/logs/api';
    }

    /**
     * 获取访问日志列表
     * @param {Object} options - 查询选项
     * @returns {Promise<Object>} 日志列表响应
     */
    async getLogs(options = {}) {
        try {
            const {
                page = 1,
                limit = 20,
                method = '',
                status = '',
                domain = '',
                search = '',
                startTime = '',
                endTime = ''
            } = options;

            const params = { page, limit };
            
            // 添加筛选条件
            if (method) params.method = method;
            if (status) params.status = status;
            if (domain) params.domain = domain;
            if (search) params.search = search;
            if (startTime) params.start_time = startTime;
            if (endTime) params.end_time = endTime;

            const response = await httpClient.get(this.baseURL, params);
            return {
                logs: response.data.logs || [],
                total: response.data.total || 0,
                page: response.data.page || page,
                totalPages: Math.ceil((response.data.total || 0) / limit),
                stats: response.data.stats || {}
            };
        } catch (error) {
            console.error('获取访问日志失败:', error);
            throw new Error('获取访问日志失败: ' + error.message);
        }
    }

    /**
     * 获取日志详情
     * @param {string} id - 日志ID
     * @returns {Promise<Object>} 日志详情
     */
    async getLogDetail(id) {
        try {
            const response = await httpClient.get(this.baseURL, { id });
            // 返回第一条日志记录，因为API返回的是数组格式
            if (response.data.logs && response.data.logs.length > 0) {
                return response.data.logs[0];
            } else {
                throw new Error('未找到指定的日志记录');
            }
        } catch (error) {
            console.error('获取日志详情失败:', error);
            throw new Error('获取日志详情失败: ' + error.message);
        }
    }

    /**
     * 获取日志统计信息
     * @param {Object} options - 统计选项
     * @returns {Promise<Object>} 统计数据
     */
    async getLogStats(options = {}) {
        try {
            const {
                period = '24h', // 24h, 7d, 30d
                groupBy = 'hour' // hour, day, status, method
            } = options;

            const params = { period, group_by: groupBy };
            const response = await httpClient.get(`${this.baseURL}/stats`, params);
            return response.data;
        } catch (error) {
            console.error('获取日志统计失败:', error);
            throw new Error('获取日志统计失败: ' + error.message);
        }
    }

    /**
     * 获取统计信息 (getStats别名方法，兼容BusinessManager调用)
     * @param {Object} options - 统计选项
     * @returns {Promise<Object>} 统计数据
     */
    async getStats(options = {}) {
        return this.getLogStats(options);
    }

    /**
     * 导出日志
     * @param {Object} options - 导出选项
     * @returns {Promise<Blob>} 导出文件
     */
    async exportLogs(options = {}) {
        try {
            const {
                format = 'csv', // csv, json
                method = '',
                status = '',
                domain = '',
                search = '',
                startTime = '',
                endTime = '',
                limit = 10000
            } = options;

            const params = { format, limit };
            
            // 添加筛选条件
            if (method) params.method = method;
            if (status) params.status = status;
            if (domain) params.domain = domain;
            if (search) params.search = search;
            if (startTime) params.start_time = startTime;
            if (endTime) params.end_time = endTime;

            const response = await httpClient.get(`${this.baseURL}/export`, params, {
                headers: {
                    'Accept': format === 'csv' ? 'text/csv' : 'application/json'
                }
            });

            // 创建下载链接
            const blob = new Blob([response.data], {
                type: format === 'csv' ? 'text/csv' : 'application/json'
            });
            
            return blob;
        } catch (error) {
            console.error('导出日志失败:', error);
            throw new Error('导出日志失败: ' + error.message);
        }
    }

    /**
     * 清理日志
     * @param {Object} options - 清理选项
     * @returns {Promise<Object>} 清理结果
     */
    async cleanLogs(options = {}) {
        try {
            const {
                beforeDate = '', // 清理此日期之前的日志
                keepDays = 30,   // 保留最近N天的日志
                status = ''      // 只清理特定状态的日志
            } = options;

            const data = { keep_days: keepDays };
            if (beforeDate) data.before_date = beforeDate;
            if (status) data.status = status;

            const response = await httpClient.post(`${this.baseURL}/clean`, data);
            return response.data;
        } catch (error) {
            console.error('清理日志失败:', error);
            throw new Error('清理日志失败: ' + error.message);
        }
    }

    /**
     * 获取热门域名
     * @param {number} limit - 返回数量限制
     * @returns {Promise<Array>} 热门域名列表
     */
    async getTopDomains(limit = 10) {
        try {
            const response = await httpClient.get(`${this.baseURL}/top-domains`, { limit });
            return response.data.domains || [];
        } catch (error) {
            console.error('获取热门域名失败:', error);
            throw new Error('获取热门域名失败: ' + error.message);
        }
    }

    /**
     * 获取错误日志
     * @param {Object} options - 查询选项
     * @returns {Promise<Object>} 错误日志列表
     */
    async getErrorLogs(options = {}) {
        try {
            const {
                page = 1,
                limit = 20,
                minStatus = 400, // 最小状态码
                search = ''
            } = options;

            const params = { 
                page, 
                limit, 
                min_status: minStatus 
            };
            
            if (search) params.search = search;

            const response = await httpClient.get(`${this.baseURL}/errors`, params);
            return {
                logs: response.data.logs || [],
                total: response.data.total || 0,
                page: response.data.page || page,
                totalPages: Math.ceil((response.data.total || 0) / limit)
            };
        } catch (error) {
            console.error('获取错误日志失败:', error);
            throw new Error('获取错误日志失败: ' + error.message);
        }
    }

    /**
     * 格式化日志数据用于显示
     * @param {Object} log - 原始日志数据
     * @returns {Object} 格式化后的日志数据
     */
    formatLogForDisplay(log) {
        return {
            ...log,
            timestamp_display: log.timestamp 
                ? new Date(log.timestamp).toLocaleString('zh-CN')
                : '-',
            method_display: log.method || '-',
            status_display: this.getStatusDisplay(log.status),
            status_class: this.getStatusClass(log.status),
            duration_display: log.duration 
                ? `${log.duration}ms`
                : '-',
            size_display: log.response_size 
                ? this.formatBytes(log.response_size)
                : '-',
            url_display: this.formatURL(log.url),
            user_agent_display: this.formatUserAgent(log.user_agent)
        };
    }

    /**
     * 获取状态码显示文本
     * @private
     */
    getStatusDisplay(status) {
        if (!status) return '-';
        
        const statusTexts = {
            200: '200 OK',
            201: '201 Created',
            204: '204 No Content',
            301: '301 Moved',
            302: '302 Found',
            304: '304 Not Modified',
            400: '400 Bad Request',
            401: '401 Unauthorized',
            403: '403 Forbidden',
            404: '404 Not Found',
            405: '405 Method Not Allowed',
            429: '429 Too Many Requests',
            500: '500 Internal Error',
            502: '502 Bad Gateway',
            503: '503 Service Unavailable',
            504: '504 Gateway Timeout'
        };

        return statusTexts[status] || `${status}`;
    }

    /**
     * 获取状态码CSS类名
     * @private
     */
    getStatusClass(status) {
        if (!status) return 'status-unknown';
        
        if (status >= 200 && status < 300) return 'status-success';
        if (status >= 300 && status < 400) return 'status-redirect';
        if (status >= 400 && status < 500) return 'status-client-error';
        if (status >= 500) return 'status-server-error';
        
        return 'status-unknown';
    }

    /**
     * 格式化字节大小
     * @private
     */
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }

    /**
     * 格式化URL显示
     * @private
     */
    formatURL(url) {
        if (!url) return '-';
        
        // 如果URL太长，截断显示
        if (url.length > 50) {
            return url.substring(0, 47) + '...';
        }
        
        return url;
    }

    /**
     * 格式化User-Agent显示
     * @private
     */
    formatUserAgent(userAgent) {
        if (!userAgent) return '-';
        
        // 简化User-Agent显示
        if (userAgent.includes('Chrome')) return 'Chrome';
        if (userAgent.includes('Firefox')) return 'Firefox';
        if (userAgent.includes('Safari')) return 'Safari';
        if (userAgent.includes('Edge')) return 'Edge';
        if (userAgent.includes('curl')) return 'curl';
        if (userAgent.includes('Postman')) return 'Postman';
        
        // 如果太长，截断显示
        if (userAgent.length > 20) {
            return userAgent.substring(0, 17) + '...';
        }
        
        return userAgent;
    }
}

// 创建全局实例
export const logsAPI = new LogsAPI();

// 导出类以便扩展
export { LogsAPI };
