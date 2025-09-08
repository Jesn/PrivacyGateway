/**
 * 状态管理器测试
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { StateManager } from '@utils/StateManager.js';

describe('StateManager', () => {
  let stateManager;

  beforeEach(() => {
    stateManager = new StateManager();
  });

  describe('初始化', () => {
    it('应该正确初始化默认状态', () => {
      expect(stateManager.getState('auth.isLoggedIn')).toBe(false);
      expect(stateManager.getState('ui.currentTab')).toBe('overview');
      expect(stateManager.getState('data.configs')).toEqual([]);
      expect(stateManager.getState('settings.theme')).toBe('light');
    });

    it('应该初始化空的订阅者和中间件', () => {
      expect(stateManager.subscribers.size).toBe(0);
      expect(stateManager.middleware).toEqual([]);
      expect(stateManager.history).toEqual([]);
    });
  });

  describe('状态获取和设置', () => {
    it('应该正确获取状态', () => {
      // 获取完整状态
      const fullState = stateManager.getState();
      expect(fullState).toHaveProperty('auth');
      expect(fullState).toHaveProperty('ui');
      expect(fullState).toHaveProperty('data');
      expect(fullState).toHaveProperty('settings');

      // 获取嵌套状态
      expect(stateManager.getState('auth.isLoggedIn')).toBe(false);
      expect(stateManager.getState('ui.currentTab')).toBe('overview');
    });

    it('应该正确设置状态', () => {
      stateManager.setState('auth.isLoggedIn', true);
      expect(stateManager.getState('auth.isLoggedIn')).toBe(true);

      stateManager.setState('ui.currentTab', 'settings');
      expect(stateManager.getState('ui.currentTab')).toBe('settings');
    });

    it('应该支持设置嵌套对象', () => {
      stateManager.setState('data.user', { id: 1, name: 'John' });
      expect(stateManager.getState('data.user.id')).toBe(1);
      expect(stateManager.getState('data.user.name')).toBe('John');
    });

    it('应该返回undefined对于不存在的路径', () => {
      expect(stateManager.getState('nonexistent.path')).toBeUndefined();
    });
  });

  describe('状态订阅', () => {
    it('应该正确订阅状态变化', () => {
      const callback = vi.fn();
      const unsubscribe = stateManager.subscribe('auth.isLoggedIn', callback);

      stateManager.setState('auth.isLoggedIn', true);
      expect(callback).toHaveBeenCalledWith(true, 'auth.isLoggedIn');

      // 测试取消订阅
      unsubscribe();
      stateManager.setState('auth.isLoggedIn', false);
      expect(callback).toHaveBeenCalledTimes(1);
    });

    it('应该支持父路径订阅', () => {
      const callback = vi.fn();
      stateManager.subscribe('auth', callback);

      stateManager.setState('auth.isLoggedIn', true);
      expect(callback).toHaveBeenCalledWith(
        expect.objectContaining({ isLoggedIn: true }),
        'auth'
      );
    });

    it('应该支持全局订阅', () => {
      const callback = vi.fn();
      stateManager.subscribe('*', callback);

      stateManager.setState('auth.isLoggedIn', true);
      expect(callback).toHaveBeenCalledWith(
        expect.objectContaining({
          auth: expect.objectContaining({ isLoggedIn: true })
        }),
        '*'
      );
    });

    it('应该处理订阅回调中的错误', () => {
      const errorCallback = vi.fn(() => {
        throw new Error('Callback error');
      });
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      stateManager.subscribe('auth.isLoggedIn', errorCallback);
      stateManager.setState('auth.isLoggedIn', true);

      expect(consoleSpy).toHaveBeenCalledWith(
        '状态订阅回调执行错误:',
        expect.any(Error)
      );

      consoleSpy.mockRestore();
    });
  });

  describe('中间件', () => {
    it('应该正确应用中间件', () => {
      const middleware = vi.fn((action) => {
        if (action.path === 'auth.isLoggedIn' && action.value === true) {
          return { ...action, value: false }; // 修改值
        }
        return action;
      });

      stateManager.addMiddleware(middleware);
      stateManager.setState('auth.isLoggedIn', true);

      expect(middleware).toHaveBeenCalled();
      expect(stateManager.getState('auth.isLoggedIn')).toBe(false);
    });

    it('应该支持中间件阻止状态更新', () => {
      const middleware = vi.fn(() => null); // 返回null阻止更新

      stateManager.addMiddleware(middleware);
      stateManager.setState('auth.isLoggedIn', true);

      expect(stateManager.getState('auth.isLoggedIn')).toBe(false); // 保持原值
    });
  });

  describe('历史记录', () => {
    it('应该记录状态变化历史', () => {
      stateManager.setState('auth.isLoggedIn', true);
      stateManager.setState('ui.currentTab', 'settings');

      const history = stateManager.getHistory();
      expect(history).toHaveLength(2);
      expect(history[0]).toMatchObject({
        path: 'auth.isLoggedIn',
        oldValue: false,
        newValue: true
      });
      expect(history[1]).toMatchObject({
        path: 'ui.currentTab',
        oldValue: 'overview',
        newValue: 'settings'
      });
    });

    it('应该限制历史记录大小', () => {
      // 设置较小的历史记录限制进行测试
      stateManager.maxHistorySize = 3;

      for (let i = 0; i < 5; i++) {
        stateManager.setState('test.counter', i);
      }

      const history = stateManager.getHistory();
      expect(history).toHaveLength(3);
      expect(history[0].newValue).toBe(2); // 最早的记录
      expect(history[2].newValue).toBe(4); // 最新的记录
    });

    it('应该正确清除历史记录', () => {
      stateManager.setState('auth.isLoggedIn', true);
      stateManager.clearHistory();

      expect(stateManager.getHistory()).toHaveLength(0);
    });
  });

  describe('批量更新', () => {
    it('应该支持批量状态更新', () => {
      const callback = vi.fn();
      stateManager.subscribe('auth', callback);

      stateManager.batchUpdate({
        'auth.isLoggedIn': true,
        'auth.user': { id: 1, name: 'John' },
        'ui.currentTab': 'profile'
      });

      expect(stateManager.getState('auth.isLoggedIn')).toBe(true);
      expect(stateManager.getState('auth.user.name')).toBe('John');
      expect(stateManager.getState('ui.currentTab')).toBe('profile');

      // 批量更新应该只触发一次通知
      expect(callback).toHaveBeenCalledTimes(1);
    });

    it('应该支持静默批量更新', () => {
      const callback = vi.fn();
      stateManager.subscribe('auth.isLoggedIn', callback);

      stateManager.batchUpdate({
        'auth.isLoggedIn': true
      }, { silent: true });

      expect(stateManager.getState('auth.isLoggedIn')).toBe(true);
      expect(callback).not.toHaveBeenCalled();
    });
  });

  describe('状态重置', () => {
    it('应该重置指定路径的状态', () => {
      stateManager.setState('auth.isLoggedIn', true);
      stateManager.setState('auth.user', { id: 1 });

      stateManager.reset('auth.isLoggedIn');

      expect(stateManager.getState('auth.isLoggedIn')).toBe(false);
      expect(stateManager.getState('auth.user')).toEqual({ id: 1 }); // 其他状态不变
    });

    it('应该重置全部状态', () => {
      stateManager.setState('auth.isLoggedIn', true);
      stateManager.setState('ui.currentTab', 'settings');

      const callback = vi.fn();
      stateManager.subscribe('*', callback);

      stateManager.reset();

      expect(stateManager.getState('auth.isLoggedIn')).toBe(false);
      expect(stateManager.getState('ui.currentTab')).toBe('overview');
      expect(callback).toHaveBeenCalled();
    });
  });

  describe('状态持久化', () => {
    beforeEach(() => {
      // 清理localStorage
      localStorage.clear();
    });

    it('应该持久化指定状态到localStorage', () => {
      stateManager.setState('settings.theme', 'dark');
      stateManager.setState('settings.language', 'en');

      stateManager.persist(['settings']);

      const stored = JSON.parse(localStorage.getItem('app-state'));
      expect(stored.settings).toEqual({
        theme: 'dark',
        language: 'en',
        pageSize: 20
      });
    });

    it('应该从localStorage恢复状态', () => {
      const testState = {
        settings: { theme: 'dark', language: 'en' }
      };
      localStorage.setItem('app-state', JSON.stringify(testState));

      stateManager.restore(['settings']);

      expect(stateManager.getState('settings.theme')).toBe('dark');
      expect(stateManager.getState('settings.language')).toBe('en');
    });

    it('应该处理localStorage错误', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      // 模拟localStorage错误
      vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
        throw new Error('Storage error');
      });

      stateManager.persist(['settings']);

      expect(consoleSpy).toHaveBeenCalledWith(
        '状态持久化失败:',
        expect.any(Error)
      );

      consoleSpy.mockRestore();
    });
  });

  describe('计算属性', () => {
    it('应该支持计算属性', () => {
      stateManager.computed('fullName', (state) => {
        const user = state.data.user || {};
        return `${user.firstName || ''} ${user.lastName || ''}`.trim();
      }, ['data.user']);

      stateManager.setState('data.user', { firstName: 'John', lastName: 'Doe' });

      expect(stateManager.getState('computed.fullName')).toBe('John Doe');
    });

    it('应该在依赖变化时重新计算', () => {
      const getter = vi.fn((state) => {
        const user = state.data.user || {};
        return `${user.firstName || ''} ${user.lastName || ''}`.trim();
      });

      stateManager.computed('fullName', getter, ['data.user']);

      stateManager.setState('data.user', { firstName: 'John', lastName: 'Doe' });
      stateManager.setState('data.user.firstName', 'Jane');

      expect(getter).toHaveBeenCalledTimes(3); // 初始 + 两次更新
      expect(stateManager.getState('computed.fullName')).toBe('Jane Doe');
    });
  });

  describe('动作', () => {
    it('应该支持定义动作', () => {
      stateManager.action('login', (stateManager, user) => {
        stateManager.setState('auth.isLoggedIn', true);
        stateManager.setState('auth.user', user);
      });

      const user = { id: 1, name: 'John' };
      stateManager.login(user);

      expect(stateManager.getState('auth.isLoggedIn')).toBe(true);
      expect(stateManager.getState('auth.user')).toEqual(user);
    });
  });
});
