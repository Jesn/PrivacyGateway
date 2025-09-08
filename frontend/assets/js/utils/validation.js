/**
 * 数据验证工具模块
 * 提供表单验证、数据格式验证等功能
 */

class ValidationUtils {
    /**
     * 验证是否为空
     * @param {*} value - 要验证的值
     * @returns {boolean} 是否为空
     */
    static isEmpty(value) {
        if (value === null || value === undefined) return true;
        if (typeof value === 'string') return value.trim() === '';
        if (Array.isArray(value)) return value.length === 0;
        if (typeof value === 'object') return Object.keys(value).length === 0;
        return false;
    }

    /**
     * 验证字符串长度
     * @param {string} value - 字符串值
     * @param {number} min - 最小长度
     * @param {number} max - 最大长度
     * @returns {boolean} 是否符合长度要求
     */
    static isValidLength(value, min = 0, max = Infinity) {
        if (typeof value !== 'string') return false;
        const length = value.trim().length;
        return length >= min && length <= max;
    }

    /**
     * 验证URL格式
     * @param {string} url - URL字符串
     * @returns {boolean} 是否为有效URL
     */
    static isValidURL(url) {
        if (!url || typeof url !== 'string') return false;
        
        try {
            new URL(url);
            return true;
        } catch (error) {
            return false;
        }
    }

    /**
     * 验证HTTP URL格式
     * @param {string} url - URL字符串
     * @returns {boolean} 是否为有效的HTTP URL
     */
    static isValidHTTPURL(url) {
        if (!this.isValidURL(url)) return false;
        
        try {
            const urlObj = new URL(url);
            return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
        } catch (error) {
            return false;
        }
    }

    /**
     * 验证邮箱格式
     * @param {string} email - 邮箱地址
     * @returns {boolean} 是否为有效邮箱
     */
    static isValidEmail(email) {
        if (!email || typeof email !== 'string') return false;
        
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    /**
     * 验证IP地址格式
     * @param {string} ip - IP地址
     * @returns {boolean} 是否为有效IP
     */
    static isValidIP(ip) {
        if (!ip || typeof ip !== 'string') return false;
        
        // IPv4格式验证
        const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/;
        if (ipv4Regex.test(ip)) {
            const parts = ip.split('.');
            return parts.every(part => {
                const num = parseInt(part);
                return num >= 0 && num <= 255;
            });
        }
        
        // IPv6格式验证（简化版）
        const ipv6Regex = /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;
        return ipv6Regex.test(ip);
    }

    /**
     * 验证CIDR格式
     * @param {string} cidr - CIDR格式的IP
     * @returns {boolean} 是否为有效CIDR
     */
    static isValidCIDR(cidr) {
        if (!cidr || typeof cidr !== 'string') return false;
        
        const parts = cidr.split('/');
        if (parts.length !== 2) return false;
        
        const [ip, mask] = parts;
        if (!this.isValidIP(ip)) return false;
        
        const maskNum = parseInt(mask);
        return !isNaN(maskNum) && maskNum >= 0 && maskNum <= 32;
    }

    /**
     * 验证端口号
     * @param {number|string} port - 端口号
     * @returns {boolean} 是否为有效端口
     */
    static isValidPort(port) {
        const portNum = parseInt(port);
        return !isNaN(portNum) && portNum >= 1 && portNum <= 65535;
    }

    /**
     * 验证HTTP方法
     * @param {string} method - HTTP方法
     * @returns {boolean} 是否为有效的HTTP方法
     */
    static isValidHTTPMethod(method) {
        if (!method || typeof method !== 'string') return false;
        
        const validMethods = [
            'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 
            'HEAD', 'OPTIONS', 'TRACE', 'CONNECT'
        ];
        
        return validMethods.includes(method.toUpperCase());
    }

    /**
     * 验证JSON格式
     * @param {string} jsonString - JSON字符串
     * @returns {boolean} 是否为有效JSON
     */
    static isValidJSON(jsonString) {
        if (!jsonString || typeof jsonString !== 'string') return false;
        
        try {
            JSON.parse(jsonString);
            return true;
        } catch (error) {
            return false;
        }
    }

    /**
     * 验证数字范围
     * @param {number|string} value - 数值
     * @param {number} min - 最小值
     * @param {number} max - 最大值
     * @returns {boolean} 是否在范围内
     */
    static isInRange(value, min = -Infinity, max = Infinity) {
        const num = parseFloat(value);
        return !isNaN(num) && num >= min && num <= max;
    }

    /**
     * 验证正整数
     * @param {number|string} value - 数值
     * @returns {boolean} 是否为正整数
     */
    static isPositiveInteger(value) {
        const num = parseInt(value);
        return !isNaN(num) && num > 0 && num === parseFloat(value);
    }

    /**
     * 验证非负整数
     * @param {number|string} value - 数值
     * @returns {boolean} 是否为非负整数
     */
    static isNonNegativeInteger(value) {
        const num = parseInt(value);
        return !isNaN(num) && num >= 0 && num === parseFloat(value);
    }

    /**
     * 验证路径格式
     * @param {string} path - 路径字符串
     * @returns {boolean} 是否为有效路径
     */
    static isValidPath(path) {
        if (!path || typeof path !== 'string') return false;
        
        // 路径必须以 / 开头
        if (!path.startsWith('/')) return false;
        
        // 不能包含连续的 //
        if (path.includes('//')) return false;
        
        // 不能包含特殊字符
        const invalidChars = ['<', '>', '"', '|', '?', '*'];
        return !invalidChars.some(char => path.includes(char));
    }

    /**
     * 验证配置名称
     * @param {string} name - 配置名称
     * @returns {boolean} 是否为有效名称
     */
    static isValidConfigName(name) {
        if (!name || typeof name !== 'string') return false;
        
        const trimmed = name.trim();
        
        // 长度检查
        if (trimmed.length < 1 || trimmed.length > 100) return false;
        
        // 不能包含特殊字符
        const invalidChars = ['<', '>', '"', '\'', '&', '\\', '/'];
        return !invalidChars.some(char => trimmed.includes(char));
    }

    /**
     * 验证令牌名称
     * @param {string} name - 令牌名称
     * @returns {boolean} 是否为有效名称
     */
    static isValidTokenName(name) {
        if (!name || typeof name !== 'string') return false;
        
        const trimmed = name.trim();
        
        // 长度检查
        if (trimmed.length < 1 || trimmed.length > 50) return false;
        
        // 只允许字母、数字、下划线、连字符
        const validPattern = /^[a-zA-Z0-9_-]+$/;
        return validPattern.test(trimmed);
    }

    /**
     * 验证密钥格式
     * @param {string} secret - 密钥
     * @returns {boolean} 是否为有效密钥
     */
    static isValidSecret(secret) {
        if (!secret || typeof secret !== 'string') return false;
        
        // 最少6个字符
        if (secret.length < 6) return false;
        
        // 最多128个字符
        if (secret.length > 128) return false;
        
        return true;
    }

    /**
     * 验证超时时间
     * @param {number|string} timeout - 超时时间（秒）
     * @returns {boolean} 是否为有效超时时间
     */
    static isValidTimeout(timeout) {
        const num = parseInt(timeout);
        return !isNaN(num) && num >= 1 && num <= 300; // 1秒到5分钟
    }

    /**
     * 验证重试次数
     * @param {number|string} retryCount - 重试次数
     * @returns {boolean} 是否为有效重试次数
     */
    static isValidRetryCount(retryCount) {
        const num = parseInt(retryCount);
        return !isNaN(num) && num >= 0 && num <= 10; // 0到10次
    }

    /**
     * 验证速率限制
     * @param {number|string} rateLimit - 速率限制（每小时请求数）
     * @returns {boolean} 是否为有效速率限制
     */
    static isValidRateLimit(rateLimit) {
        const num = parseInt(rateLimit);
        return !isNaN(num) && num >= 1 && num <= 100000; // 1到10万次/小时
    }

    /**
     * 验证表单数据
     * @param {Object} data - 表单数据
     * @param {Object} rules - 验证规则
     * @returns {Object} 验证结果 {isValid: boolean, errors: Object}
     */
    static validateForm(data, rules) {
        const errors = {};
        let isValid = true;

        Object.entries(rules).forEach(([field, rule]) => {
            const value = data[field];
            const fieldErrors = [];

            // 必填验证
            if (rule.required && this.isEmpty(value)) {
                fieldErrors.push(rule.requiredMessage || `${field}不能为空`);
            }

            // 如果值为空且非必填，跳过其他验证
            if (this.isEmpty(value) && !rule.required) {
                return;
            }

            // 长度验证
            if (rule.minLength && !this.isValidLength(value, rule.minLength)) {
                fieldErrors.push(`${field}长度不能少于${rule.minLength}个字符`);
            }
            if (rule.maxLength && !this.isValidLength(value, 0, rule.maxLength)) {
                fieldErrors.push(`${field}长度不能超过${rule.maxLength}个字符`);
            }

            // 格式验证
            if (rule.type === 'url' && !this.isValidHTTPURL(value)) {
                fieldErrors.push(`${field}格式不正确`);
            }
            if (rule.type === 'email' && !this.isValidEmail(value)) {
                fieldErrors.push(`${field}格式不正确`);
            }
            if (rule.type === 'ip' && !this.isValidIP(value)) {
                fieldErrors.push(`${field}格式不正确`);
            }

            // 数值范围验证
            if (rule.min !== undefined && !this.isInRange(value, rule.min)) {
                fieldErrors.push(`${field}不能小于${rule.min}`);
            }
            if (rule.max !== undefined && !this.isInRange(value, undefined, rule.max)) {
                fieldErrors.push(`${field}不能大于${rule.max}`);
            }

            // 自定义验证函数
            if (rule.validator && typeof rule.validator === 'function') {
                const customResult = rule.validator(value, data);
                if (customResult !== true) {
                    fieldErrors.push(customResult || `${field}验证失败`);
                }
            }

            if (fieldErrors.length > 0) {
                errors[field] = fieldErrors;
                isValid = false;
            }
        });

        return { isValid, errors };
    }

    /**
     * 获取验证错误消息
     * @param {Object} errors - 错误对象
     * @returns {string} 错误消息字符串
     */
    static getErrorMessage(errors) {
        if (!errors || typeof errors !== 'object') return '';
        
        const messages = [];
        Object.entries(errors).forEach(([field, fieldErrors]) => {
            if (Array.isArray(fieldErrors)) {
                messages.push(...fieldErrors);
            } else {
                messages.push(fieldErrors);
            }
        });
        
        return messages.join('；');
    }

    /**
     * 清理和标准化输入数据
     * @param {Object} data - 输入数据
     * @returns {Object} 清理后的数据
     */
    static sanitizeInput(data) {
        if (!data || typeof data !== 'object') return data;
        
        const sanitized = {};
        
        Object.entries(data).forEach(([key, value]) => {
            if (typeof value === 'string') {
                // 去除首尾空格
                sanitized[key] = value.trim();
            } else if (Array.isArray(value)) {
                // 清理数组中的字符串
                sanitized[key] = value.map(item => 
                    typeof item === 'string' ? item.trim() : item
                ).filter(item => item !== '');
            } else {
                sanitized[key] = value;
            }
        });
        
        return sanitized;
    }
}

// 导出工具类
export { ValidationUtils };

// 兼容旧的函数接口
export function escapeHtml(text) {
    if (!text) return '';
    
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

export function validateURL(url) {
    return ValidationUtils.isValidHTTPURL(url);
}

export function validateEmail(email) {
    return ValidationUtils.isValidEmail(email);
}
