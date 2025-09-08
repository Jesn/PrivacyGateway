/**
 * 日期时间工具模块
 * 提供日期格式化、时间计算等功能
 */

class DateUtils {
    /**
     * 格式化日期时间
     * @param {Date|string|number} date - 日期对象、字符串或时间戳
     * @param {string} format - 格式字符串
     * @returns {string} 格式化后的日期字符串
     */
    static format(date, format = 'YYYY-MM-DD HH:mm:ss') {
        if (!date) return '-';
        
        const d = new Date(date);
        if (isNaN(d.getTime())) return '-';
        
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hour = String(d.getHours()).padStart(2, '0');
        const minute = String(d.getMinutes()).padStart(2, '0');
        const second = String(d.getSeconds()).padStart(2, '0');
        
        return format
            .replace('YYYY', year)
            .replace('MM', month)
            .replace('DD', day)
            .replace('HH', hour)
            .replace('mm', minute)
            .replace('ss', second);
    }

    /**
     * 格式化相对时间
     * @param {Date|string|number} date - 日期
     * @returns {string} 相对时间字符串
     */
    static formatRelative(date) {
        if (!date) return '-';
        
        const d = new Date(date);
        if (isNaN(d.getTime())) return '-';
        
        const now = new Date();
        const diff = now.getTime() - d.getTime();
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);
        
        if (seconds < 60) {
            return '刚刚';
        } else if (minutes < 60) {
            return `${minutes}分钟前`;
        } else if (hours < 24) {
            return `${hours}小时前`;
        } else if (days < 7) {
            return `${days}天前`;
        } else {
            return this.format(date, 'MM-DD HH:mm');
        }
    }

    /**
     * 格式化日志时间（简化版）
     * @param {Date|string|number} timestamp - 时间戳
     * @returns {string} 格式化后的时间
     */
    static formatLogTime(timestamp) {
        if (!timestamp) return '-';
        
        const date = new Date(timestamp);
        if (isNaN(date.getTime())) return '-';
        
        return date.toLocaleString('zh-CN', {
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    /**
     * 获取今天的开始时间
     * @returns {Date} 今天00:00:00
     */
    static getStartOfToday() {
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        return today;
    }

    /**
     * 获取今天的结束时间
     * @returns {Date} 今天23:59:59
     */
    static getEndOfToday() {
        const today = new Date();
        today.setHours(23, 59, 59, 999);
        return today;
    }

    /**
     * 获取N天前的日期
     * @param {number} days - 天数
     * @returns {Date} N天前的日期
     */
    static getDaysAgo(days) {
        const date = new Date();
        date.setDate(date.getDate() - days);
        return date;
    }

    /**
     * 获取N天后的日期
     * @param {number} days - 天数
     * @returns {Date} N天后的日期
     */
    static getDaysLater(days) {
        const date = new Date();
        date.setDate(date.getDate() + days);
        return date;
    }

    /**
     * 判断是否为今天
     * @param {Date|string|number} date - 日期
     * @returns {boolean} 是否为今天
     */
    static isToday(date) {
        if (!date) return false;
        
        const d = new Date(date);
        if (isNaN(d.getTime())) return false;
        
        const today = new Date();
        return d.toDateString() === today.toDateString();
    }

    /**
     * 判断是否为昨天
     * @param {Date|string|number} date - 日期
     * @returns {boolean} 是否为昨天
     */
    static isYesterday(date) {
        if (!date) return false;
        
        const d = new Date(date);
        if (isNaN(d.getTime())) return false;
        
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        return d.toDateString() === yesterday.toDateString();
    }

    /**
     * 计算两个日期之间的天数差
     * @param {Date|string|number} date1 - 日期1
     * @param {Date|string|number} date2 - 日期2
     * @returns {number} 天数差
     */
    static daysBetween(date1, date2) {
        const d1 = new Date(date1);
        const d2 = new Date(date2);
        
        if (isNaN(d1.getTime()) || isNaN(d2.getTime())) return 0;
        
        const timeDiff = Math.abs(d2.getTime() - d1.getTime());
        return Math.ceil(timeDiff / (1000 * 3600 * 24));
    }

    /**
     * 格式化持续时间
     * @param {number} milliseconds - 毫秒数
     * @returns {string} 格式化后的持续时间
     */
    static formatDuration(milliseconds) {
        if (!milliseconds || milliseconds < 0) return '0ms';
        
        if (milliseconds < 1000) {
            return `${Math.round(milliseconds)}ms`;
        }
        
        const seconds = milliseconds / 1000;
        if (seconds < 60) {
            return `${seconds.toFixed(1)}s`;
        }
        
        const minutes = seconds / 60;
        if (minutes < 60) {
            return `${minutes.toFixed(1)}m`;
        }
        
        const hours = minutes / 60;
        return `${hours.toFixed(1)}h`;
    }

    /**
     * 解析ISO日期字符串
     * @param {string} isoString - ISO日期字符串
     * @returns {Date|null} 解析后的日期对象
     */
    static parseISO(isoString) {
        if (!isoString) return null;
        
        const date = new Date(isoString);
        return isNaN(date.getTime()) ? null : date;
    }

    /**
     * 转换为ISO字符串
     * @param {Date|string|number} date - 日期
     * @returns {string} ISO字符串
     */
    static toISO(date) {
        if (!date) return '';
        
        const d = new Date(date);
        return isNaN(d.getTime()) ? '' : d.toISOString();
    }

    /**
     * 获取时区偏移量
     * @returns {string} 时区偏移量字符串
     */
    static getTimezoneOffset() {
        const offset = new Date().getTimezoneOffset();
        const hours = Math.floor(Math.abs(offset) / 60);
        const minutes = Math.abs(offset) % 60;
        const sign = offset <= 0 ? '+' : '-';
        
        return `${sign}${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
    }

    /**
     * 格式化文件名安全的日期字符串
     * @param {Date|string|number} date - 日期
     * @returns {string} 文件名安全的日期字符串
     */
    static formatForFilename(date = new Date()) {
        const d = new Date(date);
        if (isNaN(d.getTime())) return 'invalid-date';
        
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hour = String(d.getHours()).padStart(2, '0');
        const minute = String(d.getMinutes()).padStart(2, '0');
        const second = String(d.getSeconds()).padStart(2, '0');
        
        return `${year}${month}${day}_${hour}${minute}${second}`;
    }

    /**
     * 获取本周的开始和结束日期
     * @returns {Object} {start: Date, end: Date}
     */
    static getThisWeek() {
        const now = new Date();
        const dayOfWeek = now.getDay();
        const start = new Date(now);
        start.setDate(now.getDate() - dayOfWeek);
        start.setHours(0, 0, 0, 0);
        
        const end = new Date(start);
        end.setDate(start.getDate() + 6);
        end.setHours(23, 59, 59, 999);
        
        return { start, end };
    }

    /**
     * 获取本月的开始和结束日期
     * @returns {Object} {start: Date, end: Date}
     */
    static getThisMonth() {
        const now = new Date();
        const start = new Date(now.getFullYear(), now.getMonth(), 1);
        const end = new Date(now.getFullYear(), now.getMonth() + 1, 0, 23, 59, 59, 999);
        
        return { start, end };
    }

    /**
     * 验证日期字符串格式
     * @param {string} dateString - 日期字符串
     * @param {string} format - 期望的格式
     * @returns {boolean} 是否符合格式
     */
    static isValidFormat(dateString, format = 'YYYY-MM-DD') {
        if (!dateString) return false;
        
        // 简单的格式验证
        const patterns = {
            'YYYY-MM-DD': /^\d{4}-\d{2}-\d{2}$/,
            'YYYY-MM-DD HH:mm': /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/,
            'YYYY-MM-DD HH:mm:ss': /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/
        };
        
        const pattern = patterns[format];
        if (!pattern) return false;
        
        if (!pattern.test(dateString)) return false;
        
        // 验证日期是否有效
        const date = new Date(dateString);
        return !isNaN(date.getTime());
    }
}

// 导出工具类
export { DateUtils };

// 兼容旧的函数接口
export function formatLogTime(timestamp) {
    return DateUtils.formatLogTime(timestamp);
}

export function formatRelativeTime(date) {
    return DateUtils.formatRelative(date);
}

export function formatDuration(milliseconds) {
    return DateUtils.formatDuration(milliseconds);
}
