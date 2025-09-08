/**
 * 日期工具测试
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { DateUtils } from '@utils/date.js';

describe('DateUtils', () => {
  beforeEach(() => {
    // 固定时间为 2023-10-15 12:00:00
    vi.setSystemTime(new Date('2023-10-15T12:00:00.000Z'));
  });

  describe('format', () => {
    it('应该正确格式化日期', () => {
      const date = new Date('2023-10-15T08:30:45.000Z');
      
      expect(DateUtils.format(date, 'YYYY-MM-DD')).toBe('2023-10-15');
      expect(DateUtils.format(date, 'YYYY-MM-DD HH:mm:ss')).toBe('2023-10-15 08:30:45');
      expect(DateUtils.format(date, 'MM/DD/YYYY')).toBe('10/15/2023');
    });

    it('应该处理字符串和时间戳输入', () => {
      const timestamp = new Date('2023-10-15T08:30:45.000Z').getTime();
      const dateString = '2023-10-15T08:30:45.000Z';
      
      expect(DateUtils.format(timestamp, 'YYYY-MM-DD')).toBe('2023-10-15');
      expect(DateUtils.format(dateString, 'YYYY-MM-DD')).toBe('2023-10-15');
    });

    it('应该处理无效输入', () => {
      expect(DateUtils.format(null)).toBe('-');
      expect(DateUtils.format(undefined)).toBe('-');
      expect(DateUtils.format('invalid-date')).toBe('-');
    });
  });

  describe('formatRelative', () => {
    it('应该正确格式化相对时间', () => {
      const now = new Date('2023-10-15T12:00:00.000Z');
      
      // 刚刚
      const justNow = new Date(now.getTime() - 30 * 1000);
      expect(DateUtils.formatRelative(justNow)).toBe('刚刚');
      
      // 分钟前
      const minutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
      expect(DateUtils.formatRelative(minutesAgo)).toBe('5分钟前');
      
      // 小时前
      const hoursAgo = new Date(now.getTime() - 2 * 60 * 60 * 1000);
      expect(DateUtils.formatRelative(hoursAgo)).toBe('2小时前');
      
      // 天前
      const daysAgo = new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000);
      expect(DateUtils.formatRelative(daysAgo)).toBe('3天前');
    });

    it('应该处理超过一周的日期', () => {
      const now = new Date('2023-10-15T12:00:00.000Z');
      const weekAgo = new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000);
      
      const result = DateUtils.formatRelative(weekAgo);
      expect(result).toMatch(/\d{2}-\d{2} \d{2}:\d{2}/);
    });
  });

  describe('formatLogTime', () => {
    it('应该正确格式化日志时间', () => {
      const date = new Date('2023-10-15T08:30:45.000Z');
      const result = DateUtils.formatLogTime(date);
      
      // 结果应该是本地化的时间格式
      expect(result).toMatch(/\d{2}\/\d{2}\/\d{4}/);
    });
  });

  describe('getStartOfToday', () => {
    it('应该返回今天的开始时间', () => {
      const startOfToday = DateUtils.getStartOfToday();
      
      expect(startOfToday.getHours()).toBe(0);
      expect(startOfToday.getMinutes()).toBe(0);
      expect(startOfToday.getSeconds()).toBe(0);
      expect(startOfToday.getMilliseconds()).toBe(0);
    });
  });

  describe('getEndOfToday', () => {
    it('应该返回今天的结束时间', () => {
      const endOfToday = DateUtils.getEndOfToday();
      
      expect(endOfToday.getHours()).toBe(23);
      expect(endOfToday.getMinutes()).toBe(59);
      expect(endOfToday.getSeconds()).toBe(59);
      expect(endOfToday.getMilliseconds()).toBe(999);
    });
  });

  describe('getDaysAgo', () => {
    it('应该返回N天前的日期', () => {
      const threeDaysAgo = DateUtils.getDaysAgo(3);
      const expected = new Date('2023-10-12T12:00:00.000Z');
      
      expect(threeDaysAgo.toDateString()).toBe(expected.toDateString());
    });
  });

  describe('getDaysLater', () => {
    it('应该返回N天后的日期', () => {
      const threeDaysLater = DateUtils.getDaysLater(3);
      const expected = new Date('2023-10-18T12:00:00.000Z');
      
      expect(threeDaysLater.toDateString()).toBe(expected.toDateString());
    });
  });

  describe('isToday', () => {
    it('应该正确判断是否为今天', () => {
      const today = new Date('2023-10-15T08:30:00.000Z');
      const yesterday = new Date('2023-10-14T08:30:00.000Z');
      const tomorrow = new Date('2023-10-16T08:30:00.000Z');
      
      expect(DateUtils.isToday(today)).toBe(true);
      expect(DateUtils.isToday(yesterday)).toBe(false);
      expect(DateUtils.isToday(tomorrow)).toBe(false);
    });

    it('应该处理无效输入', () => {
      expect(DateUtils.isToday(null)).toBe(false);
      expect(DateUtils.isToday('invalid-date')).toBe(false);
    });
  });

  describe('isYesterday', () => {
    it('应该正确判断是否为昨天', () => {
      const yesterday = new Date('2023-10-14T08:30:00.000Z');
      const today = new Date('2023-10-15T08:30:00.000Z');
      const dayBeforeYesterday = new Date('2023-10-13T08:30:00.000Z');
      
      expect(DateUtils.isYesterday(yesterday)).toBe(true);
      expect(DateUtils.isYesterday(today)).toBe(false);
      expect(DateUtils.isYesterday(dayBeforeYesterday)).toBe(false);
    });
  });

  describe('daysBetween', () => {
    it('应该正确计算两个日期之间的天数差', () => {
      const date1 = new Date('2023-10-15T12:00:00.000Z');
      const date2 = new Date('2023-10-18T12:00:00.000Z');
      
      expect(DateUtils.daysBetween(date1, date2)).toBe(3);
      expect(DateUtils.daysBetween(date2, date1)).toBe(3);
    });

    it('应该处理同一天的情况', () => {
      const date1 = new Date('2023-10-15T08:00:00.000Z');
      const date2 = new Date('2023-10-15T20:00:00.000Z');
      
      expect(DateUtils.daysBetween(date1, date2)).toBe(1);
    });

    it('应该处理无效输入', () => {
      expect(DateUtils.daysBetween('invalid', new Date())).toBe(0);
      expect(DateUtils.daysBetween(new Date(), null)).toBe(0);
    });
  });

  describe('formatDuration', () => {
    it('应该正确格式化毫秒', () => {
      expect(DateUtils.formatDuration(500)).toBe('500ms');
      expect(DateUtils.formatDuration(0)).toBe('0ms');
    });

    it('应该正确格式化秒', () => {
      expect(DateUtils.formatDuration(1500)).toBe('1.5s');
      expect(DateUtils.formatDuration(30000)).toBe('30.0s');
    });

    it('应该正确格式化分钟', () => {
      expect(DateUtils.formatDuration(90000)).toBe('1.5m');
      expect(DateUtils.formatDuration(300000)).toBe('5.0m');
    });

    it('应该正确格式化小时', () => {
      expect(DateUtils.formatDuration(3600000)).toBe('1.0h');
      expect(DateUtils.formatDuration(5400000)).toBe('1.5h');
    });

    it('应该处理无效输入', () => {
      expect(DateUtils.formatDuration(null)).toBe('0ms');
      expect(DateUtils.formatDuration(-100)).toBe('0ms');
    });
  });

  describe('formatForFilename', () => {
    it('应该生成文件名安全的日期字符串', () => {
      const date = new Date('2023-10-15T08:30:45.000Z');
      const result = DateUtils.formatForFilename(date);
      
      expect(result).toBe('20231015_083045');
      expect(result).toMatch(/^\d{8}_\d{6}$/);
    });

    it('应该处理无效输入', () => {
      expect(DateUtils.formatForFilename('invalid')).toBe('invalid-date');
    });
  });

  describe('getThisWeek', () => {
    it('应该返回本周的开始和结束日期', () => {
      const { start, end } = DateUtils.getThisWeek();
      
      expect(start.getDay()).toBe(0); // 周日
      expect(end.getDay()).toBe(6);   // 周六
      expect(start.getHours()).toBe(0);
      expect(end.getHours()).toBe(23);
    });
  });

  describe('getThisMonth', () => {
    it('应该返回本月的开始和结束日期', () => {
      const { start, end } = DateUtils.getThisMonth();
      
      expect(start.getDate()).toBe(1);
      expect(start.getHours()).toBe(0);
      expect(end.getMonth()).toBe(start.getMonth());
      expect(end.getHours()).toBe(23);
    });
  });

  describe('isValidFormat', () => {
    it('应该正确验证日期格式', () => {
      expect(DateUtils.isValidFormat('2023-10-15', 'YYYY-MM-DD')).toBe(true);
      expect(DateUtils.isValidFormat('2023-10-15 08:30', 'YYYY-MM-DD HH:mm')).toBe(true);
      expect(DateUtils.isValidFormat('2023-10-15 08:30:45', 'YYYY-MM-DD HH:mm:ss')).toBe(true);
    });

    it('应该拒绝无效格式', () => {
      expect(DateUtils.isValidFormat('15-10-2023', 'YYYY-MM-DD')).toBe(false);
      expect(DateUtils.isValidFormat('2023/10/15', 'YYYY-MM-DD')).toBe(false);
      expect(DateUtils.isValidFormat('invalid-date', 'YYYY-MM-DD')).toBe(false);
    });

    it('应该验证日期有效性', () => {
      expect(DateUtils.isValidFormat('2023-02-30', 'YYYY-MM-DD')).toBe(false);
      expect(DateUtils.isValidFormat('2023-13-01', 'YYYY-MM-DD')).toBe(false);
    });
  });
});
