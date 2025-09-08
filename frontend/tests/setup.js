/**
 * 测试环境设置文件
 * 配置全局测试环境、模拟对象、工具函数等
 */

import { vi } from 'vitest';

// ===== 全局模拟 =====

// 模拟 localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  length: 0,
  key: vi.fn()
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true
});

// 模拟 sessionStorage
const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  length: 0,
  key: vi.fn()
};

Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
  writable: true
});

// 模拟 fetch
global.fetch = vi.fn();

// 模拟 IntersectionObserver
global.IntersectionObserver = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn()
}));

// 模拟 ResizeObserver
global.ResizeObserver = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn()
}));

// 模拟 MutationObserver
global.MutationObserver = vi.fn(() => ({
  observe: vi.fn(),
  disconnect: vi.fn(),
  takeRecords: vi.fn()
}));

// 模拟 PerformanceObserver
global.PerformanceObserver = vi.fn(() => ({
  observe: vi.fn(),
  disconnect: vi.fn()
}));

// 模拟 performance API
Object.defineProperty(window, 'performance', {
  value: {
    now: vi.fn(() => Date.now()),
    mark: vi.fn(),
    measure: vi.fn(),
    getEntriesByType: vi.fn(() => []),
    getEntriesByName: vi.fn(() => []),
    memory: {
      usedJSHeapSize: 1000000,
      totalJSHeapSize: 2000000,
      jsHeapSizeLimit: 4000000
    }
  },
  writable: true
});

// 模拟 navigator
Object.defineProperty(window, 'navigator', {
  value: {
    userAgent: 'Mozilla/5.0 (Test Environment)',
    language: 'zh-CN',
    platform: 'Test',
    cookieEnabled: true,
    onLine: true
  },
  writable: true
});

// 模拟 location
Object.defineProperty(window, 'location', {
  value: {
    href: 'http://localhost:3000',
    origin: 'http://localhost:3000',
    protocol: 'http:',
    host: 'localhost:3000',
    hostname: 'localhost',
    port: '3000',
    pathname: '/',
    search: '',
    hash: '',
    reload: vi.fn(),
    assign: vi.fn(),
    replace: vi.fn()
  },
  writable: true
});

// ===== 测试工具函数 =====

/**
 * 创建模拟的DOM元素
 * @param {string} tagName - 标签名
 * @param {Object} attributes - 属性对象
 * @param {string} textContent - 文本内容
 * @returns {HTMLElement} 模拟元素
 */
export function createMockElement(tagName = 'div', attributes = {}, textContent = '') {
  const element = document.createElement(tagName);
  
  Object.entries(attributes).forEach(([key, value]) => {
    if (key === 'className') {
      element.className = value;
    } else if (key === 'dataset') {
      Object.entries(value).forEach(([dataKey, dataValue]) => {
        element.dataset[dataKey] = dataValue;
      });
    } else {
      element.setAttribute(key, value);
    }
  });
  
  if (textContent) {
    element.textContent = textContent;
  }
  
  return element;
}

/**
 * 创建模拟的事件对象
 * @param {string} type - 事件类型
 * @param {Object} properties - 事件属性
 * @returns {Event} 模拟事件
 */
export function createMockEvent(type, properties = {}) {
  const event = new Event(type, { bubbles: true, cancelable: true });
  
  Object.entries(properties).forEach(([key, value]) => {
    Object.defineProperty(event, key, { value, writable: true });
  });
  
  return event;
}

/**
 * 等待异步操作完成
 * @param {number} ms - 等待时间（毫秒）
 * @returns {Promise} Promise对象
 */
export function waitFor(ms = 0) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 等待DOM更新
 * @returns {Promise} Promise对象
 */
export function nextTick() {
  return new Promise(resolve => {
    if (typeof window !== 'undefined' && window.requestAnimationFrame) {
      window.requestAnimationFrame(() => {
        window.requestAnimationFrame(resolve);
      });
    } else {
      setTimeout(resolve, 0);
    }
  });
}

/**
 * 模拟用户点击
 * @param {HTMLElement} element - 目标元素
 */
export function mockClick(element) {
  const event = createMockEvent('click', {
    target: element,
    currentTarget: element
  });
  element.dispatchEvent(event);
}

/**
 * 模拟用户输入
 * @param {HTMLElement} element - 输入元素
 * @param {string} value - 输入值
 */
export function mockInput(element, value) {
  element.value = value;
  const event = createMockEvent('input', {
    target: element,
    currentTarget: element
  });
  element.dispatchEvent(event);
}

/**
 * 模拟表单提交
 * @param {HTMLFormElement} form - 表单元素
 */
export function mockSubmit(form) {
  const event = createMockEvent('submit', {
    target: form,
    currentTarget: form
  });
  form.dispatchEvent(event);
}

/**
 * 清理测试环境
 */
export function cleanup() {
  // 清理 DOM
  document.body.innerHTML = '';
  document.head.innerHTML = '';
  
  // 重置模拟对象
  vi.clearAllMocks();
  
  // 清理存储
  localStorageMock.clear();
  sessionStorageMock.clear();
  
  // 清理定时器
  vi.clearAllTimers();
}

// ===== 全局测试钩子 =====

// 每个测试前的清理
beforeEach(() => {
  cleanup();
});

// 每个测试后的清理
afterEach(() => {
  cleanup();
});

// 测试套件结束后的清理
afterAll(() => {
  cleanup();
});

// ===== 全局变量 =====

// 将工具函数挂载到全局，方便测试文件使用
global.createMockElement = createMockElement;
global.createMockEvent = createMockEvent;
global.waitFor = waitFor;
global.nextTick = nextTick;
global.mockClick = mockClick;
global.mockInput = mockInput;
global.mockSubmit = mockSubmit;
global.cleanup = cleanup;
