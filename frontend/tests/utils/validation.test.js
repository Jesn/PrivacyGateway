/**
 * 数据验证工具测试
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { ValidationUtils } from '@utils/validation.js';

describe('ValidationUtils', () => {
  describe('isEmpty', () => {
    it('应该正确判断空值', () => {
      expect(ValidationUtils.isEmpty(null)).toBe(true);
      expect(ValidationUtils.isEmpty(undefined)).toBe(true);
      expect(ValidationUtils.isEmpty('')).toBe(true);
      expect(ValidationUtils.isEmpty('   ')).toBe(true);
      expect(ValidationUtils.isEmpty([])).toBe(true);
      expect(ValidationUtils.isEmpty({})).toBe(true);
    });

    it('应该正确判断非空值', () => {
      expect(ValidationUtils.isEmpty('hello')).toBe(false);
      expect(ValidationUtils.isEmpty(0)).toBe(false);
      expect(ValidationUtils.isEmpty(false)).toBe(false);
      expect(ValidationUtils.isEmpty([1, 2, 3])).toBe(false);
      expect(ValidationUtils.isEmpty({ a: 1 })).toBe(false);
    });
  });

  describe('isValidLength', () => {
    it('应该正确验证字符串长度', () => {
      expect(ValidationUtils.isValidLength('hello', 1, 10)).toBe(true);
      expect(ValidationUtils.isValidLength('hello', 5, 5)).toBe(true);
      expect(ValidationUtils.isValidLength('hello', 6, 10)).toBe(false);
      expect(ValidationUtils.isValidLength('hello', 1, 4)).toBe(false);
    });

    it('应该处理边界情况', () => {
      expect(ValidationUtils.isValidLength('', 0, 5)).toBe(true);
      expect(ValidationUtils.isValidLength('   ', 0, 5)).toBe(true);
      expect(ValidationUtils.isValidLength(123, 1, 5)).toBe(false);
    });
  });

  describe('isValidURL', () => {
    it('应该正确验证有效的URL', () => {
      expect(ValidationUtils.isValidURL('https://example.com')).toBe(true);
      expect(ValidationUtils.isValidURL('http://localhost:3000')).toBe(true);
      expect(ValidationUtils.isValidURL('ftp://files.example.com')).toBe(true);
    });

    it('应该正确识别无效的URL', () => {
      expect(ValidationUtils.isValidURL('not-a-url')).toBe(false);
      expect(ValidationUtils.isValidURL('http://')).toBe(false);
      expect(ValidationUtils.isValidURL('')).toBe(false);
      expect(ValidationUtils.isValidURL(null)).toBe(false);
    });
  });

  describe('isValidHTTPURL', () => {
    it('应该只接受HTTP/HTTPS协议', () => {
      expect(ValidationUtils.isValidHTTPURL('https://example.com')).toBe(true);
      expect(ValidationUtils.isValidHTTPURL('http://example.com')).toBe(true);
      expect(ValidationUtils.isValidHTTPURL('ftp://example.com')).toBe(false);
      expect(ValidationUtils.isValidHTTPURL('file:///path/to/file')).toBe(false);
    });
  });

  describe('isValidEmail', () => {
    it('应该正确验证邮箱格式', () => {
      expect(ValidationUtils.isValidEmail('user@example.com')).toBe(true);
      expect(ValidationUtils.isValidEmail('test.email+tag@domain.co.uk')).toBe(true);
      expect(ValidationUtils.isValidEmail('user123@test-domain.org')).toBe(true);
    });

    it('应该拒绝无效的邮箱格式', () => {
      expect(ValidationUtils.isValidEmail('invalid-email')).toBe(false);
      expect(ValidationUtils.isValidEmail('user@')).toBe(false);
      expect(ValidationUtils.isValidEmail('@domain.com')).toBe(false);
      expect(ValidationUtils.isValidEmail('user@domain')).toBe(false);
      expect(ValidationUtils.isValidEmail('')).toBe(false);
    });
  });

  describe('isValidIP', () => {
    it('应该正确验证IPv4地址', () => {
      expect(ValidationUtils.isValidIP('192.168.1.1')).toBe(true);
      expect(ValidationUtils.isValidIP('127.0.0.1')).toBe(true);
      expect(ValidationUtils.isValidIP('255.255.255.255')).toBe(true);
      expect(ValidationUtils.isValidIP('0.0.0.0')).toBe(true);
    });

    it('应该拒绝无效的IPv4地址', () => {
      expect(ValidationUtils.isValidIP('256.1.1.1')).toBe(false);
      expect(ValidationUtils.isValidIP('192.168.1')).toBe(false);
      expect(ValidationUtils.isValidIP('192.168.1.1.1')).toBe(false);
      expect(ValidationUtils.isValidIP('not-an-ip')).toBe(false);
    });

    it('应该正确验证IPv6地址', () => {
      expect(ValidationUtils.isValidIP('2001:0db8:85a3:0000:0000:8a2e:0370:7334')).toBe(true);
      expect(ValidationUtils.isValidIP('2001:db8:85a3::8a2e:370:7334')).toBe(false); // 简化格式暂不支持
    });
  });

  describe('isValidPort', () => {
    it('应该正确验证端口号', () => {
      expect(ValidationUtils.isValidPort(80)).toBe(true);
      expect(ValidationUtils.isValidPort('443')).toBe(true);
      expect(ValidationUtils.isValidPort(65535)).toBe(true);
      expect(ValidationUtils.isValidPort(1)).toBe(true);
    });

    it('应该拒绝无效的端口号', () => {
      expect(ValidationUtils.isValidPort(0)).toBe(false);
      expect(ValidationUtils.isValidPort(65536)).toBe(false);
      expect(ValidationUtils.isValidPort(-1)).toBe(false);
      expect(ValidationUtils.isValidPort('not-a-port')).toBe(false);
    });
  });

  describe('isValidJSON', () => {
    it('应该正确验证JSON格式', () => {
      expect(ValidationUtils.isValidJSON('{"key": "value"}')).toBe(true);
      expect(ValidationUtils.isValidJSON('[1, 2, 3]')).toBe(true);
      expect(ValidationUtils.isValidJSON('"string"')).toBe(true);
      expect(ValidationUtils.isValidJSON('123')).toBe(true);
      expect(ValidationUtils.isValidJSON('true')).toBe(true);
    });

    it('应该拒绝无效的JSON格式', () => {
      expect(ValidationUtils.isValidJSON('{key: "value"}')).toBe(false);
      expect(ValidationUtils.isValidJSON('{"key": value}')).toBe(false);
      expect(ValidationUtils.isValidJSON('invalid json')).toBe(false);
      expect(ValidationUtils.isValidJSON('')).toBe(false);
    });
  });

  describe('validateForm', () => {
    let formData;
    let rules;

    beforeEach(() => {
      formData = {
        name: 'John Doe',
        email: 'john@example.com',
        age: 25,
        website: 'https://johndoe.com'
      };

      rules = {
        name: {
          required: true,
          minLength: 2,
          maxLength: 50
        },
        email: {
          required: true,
          type: 'email'
        },
        age: {
          required: true,
          min: 18,
          max: 100
        },
        website: {
          type: 'url'
        }
      };
    });

    it('应该通过有效数据的验证', () => {
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(true);
      expect(Object.keys(result.errors)).toHaveLength(0);
    });

    it('应该检测必填字段缺失', () => {
      delete formData.name;
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(false);
      expect(result.errors.name).toBeDefined();
    });

    it('应该检测字段长度错误', () => {
      formData.name = 'A';
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(false);
      expect(result.errors.name).toBeDefined();
    });

    it('应该检测邮箱格式错误', () => {
      formData.email = 'invalid-email';
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(false);
      expect(result.errors.email).toBeDefined();
    });

    it('应该检测数值范围错误', () => {
      formData.age = 15;
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(false);
      expect(result.errors.age).toBeDefined();
    });

    it('应该支持自定义验证函数', () => {
      rules.name.validator = (value) => {
        return value.includes('John') ? true : '名称必须包含John';
      };

      formData.name = 'Jane Doe';
      const result = ValidationUtils.validateForm(formData, rules);
      expect(result.isValid).toBe(false);
      expect(result.errors.name).toContain('名称必须包含John');
    });
  });

  describe('sanitizeInput', () => {
    it('应该清理字符串输入', () => {
      const input = {
        name: '  John Doe  ',
        email: ' john@example.com ',
        tags: ['  tag1  ', '  tag2  ', '']
      };

      const result = ValidationUtils.sanitizeInput(input);
      expect(result.name).toBe('John Doe');
      expect(result.email).toBe('john@example.com');
      expect(result.tags).toEqual(['tag1', 'tag2']);
    });

    it('应该处理非对象输入', () => {
      expect(ValidationUtils.sanitizeInput(null)).toBe(null);
      expect(ValidationUtils.sanitizeInput('string')).toBe('string');
      expect(ValidationUtils.sanitizeInput(123)).toBe(123);
    });
  });
});
