/**
 * 基础组件测试
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Component } from '@components/Component.js';

// 创建测试组件类
class TestComponent extends Component {
  getDefaultOptions() {
    return {
      title: 'Test Component',
      visible: true
    };
  }

  getInitialState() {
    return {
      count: 0,
      message: 'Hello'
    };
  }

  render() {
    const { title } = this.options;
    const { count, message } = this.state;
    
    this.container.innerHTML = `
      <div class="test-component">
        <h1>${title}</h1>
        <p>${message}</p>
        <span class="count">${count}</span>
        <button class="increment">+</button>
        <button class="decrement">-</button>
      </div>
    `;
  }

  bindEvents() {
    this.delegate('.increment', 'click', () => {
      this.setState({ count: this.state.count + 1 });
    });

    this.delegate('.decrement', 'click', () => {
      this.setState({ count: this.state.count - 1 });
    });
  }
}

describe('Component', () => {
  let container;
  let component;

  beforeEach(() => {
    container = createMockElement('div', { id: 'test-container' });
    document.body.appendChild(container);
  });

  describe('构造函数', () => {
    it('应该正确初始化组件', () => {
      component = new TestComponent(container, { title: 'Custom Title' });
      
      expect(component.container).toBe(container);
      expect(component.options.title).toBe('Custom Title');
      expect(component.state.count).toBe(0);
      expect(component.id).toBeDefined();
      expect(component.isDestroyed).toBe(false);
    });

    it('应该支持选择器字符串作为容器', () => {
      component = new TestComponent('#test-container');
      expect(component.container).toBe(container);
    });

    it('应该在容器不存在时抛出错误', () => {
      expect(() => {
        new TestComponent('#non-existent');
      }).toThrow('组件容器不能为空');
    });

    it('应该合并默认选项和传入选项', () => {
      component = new TestComponent(container, { title: 'Custom', newOption: 'value' });
      
      expect(component.options.title).toBe('Custom');
      expect(component.options.visible).toBe(true); // 默认选项
      expect(component.options.newOption).toBe('value');
    });
  });

  describe('生命周期', () => {
    it('应该按正确顺序调用生命周期方法', () => {
      const beforeMountSpy = vi.spyOn(TestComponent.prototype, 'beforeMount');
      const mountSpy = vi.spyOn(TestComponent.prototype, 'mount');
      const afterMountSpy = vi.spyOn(TestComponent.prototype, 'afterMount');
      const renderSpy = vi.spyOn(TestComponent.prototype, 'render');
      const bindEventsSpy = vi.spyOn(TestComponent.prototype, 'bindEvents');

      component = new TestComponent(container);

      expect(beforeMountSpy).toHaveBeenCalled();
      expect(mountSpy).toHaveBeenCalled();
      expect(renderSpy).toHaveBeenCalled();
      expect(bindEventsSpy).toHaveBeenCalled();
      expect(afterMountSpy).toHaveBeenCalled();
    });

    it('应该在挂载后触发mounted事件', () => {
      const eventSpy = vi.fn();
      
      component = new TestComponent(container);
      component.on('mounted', eventSpy);
      
      // 重新挂载以触发事件
      component.afterMount();
      
      expect(eventSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            component: component
          })
        })
      );
    });
  });

  describe('状态管理', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确设置状态', () => {
      component.setState({ count: 5 });
      expect(component.state.count).toBe(5);
    });

    it('应该支持函数式状态更新', () => {
      component.setState(prevState => ({ count: prevState.count + 10 }));
      expect(component.state.count).toBe(10);
    });

    it('应该在状态更新后重新渲染', () => {
      const renderSpy = vi.spyOn(component, 'render');
      component.setState({ count: 5 });
      expect(renderSpy).toHaveBeenCalled();
    });

    it('应该触发状态变化事件', () => {
      const eventSpy = vi.fn();
      component.on('stateChange', eventSpy);
      
      component.setState({ count: 5 });
      
      expect(eventSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            prevState: expect.objectContaining({ count: 0 }),
            newState: expect.objectContaining({ count: 5 })
          })
        })
      );
    });

    it('应该支持状态更新回调', () => {
      const callback = vi.fn();
      component.setState({ count: 5 }, callback);
      expect(callback).toHaveBeenCalled();
    });

    it('应该正确获取状态', () => {
      expect(component.getState('count')).toBe(0);
      expect(component.getState()).toEqual({ count: 0, message: 'Hello' });
    });
  });

  describe('事件处理', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确添加事件监听器', () => {
      const handler = vi.fn();
      component.addEventListener('click', handler);
      
      const event = createMockEvent('click');
      container.dispatchEvent(event);
      
      expect(handler).toHaveBeenCalled();
    });

    it('应该正确移除事件监听器', () => {
      const handler = vi.fn();
      component.addEventListener('click', handler);
      component.removeEventListener('click', handler);
      
      const event = createMockEvent('click');
      container.dispatchEvent(event);
      
      expect(handler).not.toHaveBeenCalled();
    });

    it('应该支持事件委托', () => {
      const handler = vi.fn();
      component.delegate('.increment', 'click', handler);
      
      const button = component.$('.increment');
      mockClick(button);
      
      expect(handler).toHaveBeenCalled();
    });

    it('应该正确触发自定义事件', () => {
      const handler = vi.fn();
      component.on('customEvent', handler);
      
      component.emit('customEvent', { data: 'test' });
      
      expect(handler).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            data: 'test'
          })
        })
      );
    });
  });

  describe('DOM操作', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确查找元素', () => {
      const title = component.$('h1');
      expect(title).toBeTruthy();
      expect(title.textContent).toBe('Test Component');
    });

    it('应该正确查找多个元素', () => {
      const buttons = component.$$('button');
      expect(buttons).toHaveLength(2);
    });

    it('应该正确显示和隐藏组件', () => {
      component.hide();
      expect(container.style.display).toBe('none');
      
      component.show();
      expect(container.style.display).toBe('');
    });
  });

  describe('子组件管理', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确添加子组件', () => {
      const childContainer = createMockElement('div');
      const childComponent = new TestComponent(childContainer);
      
      component.addChild('child', childComponent);
      
      expect(component.getChild('child')).toBe(childComponent);
      expect(childComponent.parent).toBe(component);
    });

    it('应该正确移除子组件', () => {
      const childContainer = createMockElement('div');
      const childComponent = new TestComponent(childContainer);
      const destroySpy = vi.spyOn(childComponent, 'destroy');
      
      component.addChild('child', childComponent);
      component.removeChild('child');
      
      expect(component.getChild('child')).toBeUndefined();
      expect(destroySpy).toHaveBeenCalled();
    });
  });

  describe('组件销毁', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确销毁组件', () => {
      const beforeDestroySpy = vi.spyOn(component, 'beforeDestroy');
      const afterDestroySpy = vi.spyOn(component, 'afterDestroy');
      
      component.destroy();
      
      expect(beforeDestroySpy).toHaveBeenCalled();
      expect(afterDestroySpy).toHaveBeenCalled();
      expect(component.isDestroyed).toBe(true);
      expect(container.innerHTML).toBe('');
    });

    it('应该清理所有事件监听器', () => {
      const handler = vi.fn();
      component.addEventListener('click', handler);
      
      component.destroy();
      
      const event = createMockEvent('click');
      container.dispatchEvent(event);
      
      expect(handler).not.toHaveBeenCalled();
    });

    it('应该销毁所有子组件', () => {
      const childContainer = createMockElement('div');
      const childComponent = new TestComponent(childContainer);
      const destroySpy = vi.spyOn(childComponent, 'destroy');
      
      component.addChild('child', childComponent);
      component.destroy();
      
      expect(destroySpy).toHaveBeenCalled();
    });

    it('应该防止重复销毁', () => {
      const beforeDestroySpy = vi.spyOn(component, 'beforeDestroy');
      
      component.destroy();
      component.destroy(); // 第二次调用
      
      expect(beforeDestroySpy).toHaveBeenCalledTimes(1);
    });
  });

  describe('功能测试', () => {
    beforeEach(() => {
      component = new TestComponent(container);
    });

    it('应该正确处理点击事件', () => {
      const incrementButton = component.$('.increment');
      const decrementButton = component.$('.decrement');
      const countSpan = component.$('.count');
      
      // 测试增加
      mockClick(incrementButton);
      expect(component.state.count).toBe(1);
      expect(countSpan.textContent).toBe('1');
      
      // 测试减少
      mockClick(decrementButton);
      expect(component.state.count).toBe(0);
      expect(countSpan.textContent).toBe('0');
    });

    it('应该在销毁后停止响应状态更新', () => {
      component.destroy();
      
      const renderSpy = vi.spyOn(component, 'render');
      component.setState({ count: 5 });
      
      expect(renderSpy).not.toHaveBeenCalled();
    });
  });
});
