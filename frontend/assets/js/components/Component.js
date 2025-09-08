/**
 * 基础组件类
 * 提供组件生命周期管理、状态管理、事件绑定等基础功能
 */

import { DOMUtils } from '../utils/dom.js';

class Component {
    constructor(container, options = {}) {
        this.container = typeof container === 'string' 
            ? DOMUtils.get(container) 
            : container;
        
        if (!this.container) {
            throw new Error('组件容器不能为空');
        }

        this.options = { ...this.getDefaultOptions(), ...options };
        this.state = { ...this.getInitialState() };
        this.eventListeners = new Map();
        this.childComponents = new Map();
        this.isDestroyed = false;
        
        // 生成唯一ID
        this.id = this.generateId();
        this.container.setAttribute('data-component-id', this.id);
        
        // 初始化组件
        this.init();
    }

    /**
     * 获取默认选项
     * @returns {Object} 默认选项
     */
    getDefaultOptions() {
        return {};
    }

    /**
     * 获取初始状态
     * @returns {Object} 初始状态
     */
    getInitialState() {
        return {};
    }

    /**
     * 生成唯一ID
     * @returns {string} 唯一ID
     */
    generateId() {
        return `component-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    /**
     * 初始化组件
     */
    init() {
        this.beforeMount();
        this.mount();
        this.afterMount();
    }

    /**
     * 挂载前钩子
     */
    beforeMount() {
        // 子类可以重写此方法
    }

    /**
     * 挂载组件
     */
    mount() {
        this.render();
        this.bindEvents();
    }

    /**
     * 挂载后钩子
     */
    afterMount() {
        // 子类可以重写此方法
        this.emit('mounted', { component: this });
    }

    /**
     * 渲染组件
     */
    render() {
        // 子类必须实现此方法
        throw new Error('render方法必须被子类实现');
    }

    /**
     * 绑定事件
     */
    bindEvents() {
        // 子类可以重写此方法
    }

    /**
     * 更新状态
     * @param {Object|Function} newState - 新状态或状态更新函数
     * @param {Function} callback - 状态更新后的回调
     */
    setState(newState, callback) {
        if (this.isDestroyed) return;

        const prevState = { ...this.state };
        
        if (typeof newState === 'function') {
            this.state = { ...this.state, ...newState(this.state) };
        } else {
            this.state = { ...this.state, ...newState };
        }

        // 触发状态变化事件
        this.emit('stateChange', {
            prevState,
            newState: this.state,
            component: this
        });

        // 重新渲染
        this.update();

        // 执行回调
        if (typeof callback === 'function') {
            callback();
        }
    }

    /**
     * 获取状态
     * @param {string} key - 状态键名
     * @returns {*} 状态值
     */
    getState(key) {
        return key ? this.state[key] : this.state;
    }

    /**
     * 更新组件
     */
    update() {
        if (this.isDestroyed) return;

        this.beforeUpdate();
        this.render();
        this.afterUpdate();
    }

    /**
     * 更新前钩子
     */
    beforeUpdate() {
        // 子类可以重写此方法
    }

    /**
     * 更新后钩子
     */
    afterUpdate() {
        // 子类可以重写此方法
        this.emit('updated', { component: this });
    }

    /**
     * 添加事件监听器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     * @param {Element} target - 目标元素，默认为容器
     */
    addEventListener(event, handler, target = this.container) {
        if (!this.eventListeners.has(target)) {
            this.eventListeners.set(target, new Map());
        }
        
        const targetListeners = this.eventListeners.get(target);
        if (!targetListeners.has(event)) {
            targetListeners.set(event, []);
        }
        
        targetListeners.get(event).push(handler);
        target.addEventListener(event, handler);
    }

    /**
     * 移除事件监听器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     * @param {Element} target - 目标元素
     */
    removeEventListener(event, handler, target = this.container) {
        const targetListeners = this.eventListeners.get(target);
        if (!targetListeners) return;
        
        const eventHandlers = targetListeners.get(event);
        if (!eventHandlers) return;
        
        const index = eventHandlers.indexOf(handler);
        if (index > -1) {
            eventHandlers.splice(index, 1);
            target.removeEventListener(event, handler);
        }
    }

    /**
     * 委托事件监听
     * @param {string} selector - 选择器
     * @param {string} event - 事件名
     * @param {Function} handler - 事件处理函数
     */
    delegate(selector, event, handler) {
        const delegateHandler = (e) => {
            const target = e.target.closest(selector);
            if (target && this.container.contains(target)) {
                handler.call(target, e);
            }
        };
        
        this.addEventListener(event, delegateHandler);
    }

    /**
     * 触发自定义事件
     * @param {string} eventName - 事件名
     * @param {*} data - 事件数据
     */
    emit(eventName, data) {
        const event = new CustomEvent(`component:${eventName}`, {
            detail: data,
            bubbles: true
        });
        
        this.container.dispatchEvent(event);
        
        // 同时触发全局事件
        window.dispatchEvent(new CustomEvent(`component:${this.constructor.name}:${eventName}`, {
            detail: { ...data, componentId: this.id }
        }));
    }

    /**
     * 监听自定义事件
     * @param {string} eventName - 事件名
     * @param {Function} handler - 事件处理函数
     */
    on(eventName, handler) {
        this.addEventListener(`component:${eventName}`, handler);
    }

    /**
     * 移除自定义事件监听
     * @param {string} eventName - 事件名
     * @param {Function} handler - 事件处理函数
     */
    off(eventName, handler) {
        this.removeEventListener(`component:${eventName}`, handler);
    }

    /**
     * 添加子组件
     * @param {string} name - 组件名称
     * @param {Component} component - 组件实例
     */
    addChild(name, component) {
        this.childComponents.set(name, component);
        component.parent = this;
    }

    /**
     * 获取子组件
     * @param {string} name - 组件名称
     * @returns {Component} 组件实例
     */
    getChild(name) {
        return this.childComponents.get(name);
    }

    /**
     * 移除子组件
     * @param {string} name - 组件名称
     */
    removeChild(name) {
        const component = this.childComponents.get(name);
        if (component) {
            component.destroy();
            this.childComponents.delete(name);
        }
    }

    /**
     * 查找元素
     * @param {string} selector - 选择器
     * @returns {Element} 元素
     */
    $(selector) {
        return DOMUtils.get(selector, this.container);
    }

    /**
     * 查找多个元素
     * @param {string} selector - 选择器
     * @returns {NodeList} 元素列表
     */
    $$(selector) {
        return DOMUtils.getAll(selector, this.container);
    }

    /**
     * 显示组件
     */
    show() {
        DOMUtils.show(this.container);
        this.emit('show', { component: this });
    }

    /**
     * 隐藏组件
     */
    hide() {
        DOMUtils.hide(this.container);
        this.emit('hide', { component: this });
    }

    /**
     * 切换显示/隐藏
     */
    toggle() {
        DOMUtils.toggle(this.container);
        this.emit('toggle', { component: this });
    }

    /**
     * 销毁组件
     */
    destroy() {
        if (this.isDestroyed) return;

        this.beforeDestroy();
        
        // 销毁所有子组件
        this.childComponents.forEach(component => component.destroy());
        this.childComponents.clear();
        
        // 移除所有事件监听器
        this.eventListeners.forEach((targetListeners, target) => {
            targetListeners.forEach((handlers, event) => {
                handlers.forEach(handler => {
                    target.removeEventListener(event, handler);
                });
            });
        });
        this.eventListeners.clear();
        
        // 清空容器
        if (this.container) {
            this.container.innerHTML = '';
            this.container.removeAttribute('data-component-id');
        }
        
        this.isDestroyed = true;
        this.afterDestroy();
    }

    /**
     * 销毁前钩子
     */
    beforeDestroy() {
        // 子类可以重写此方法
        this.emit('beforeDestroy', { component: this });
    }

    /**
     * 销毁后钩子
     */
    afterDestroy() {
        // 子类可以重写此方法
        this.emit('destroyed', { component: this });
    }
}

export { Component };
