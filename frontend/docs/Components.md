# 组件文档

Privacy Gateway Frontend 的组件系统文档。

## 📋 目录

- [组件系统概述](#组件系统概述)
- [基础组件类](#基础组件类)
- [UI组件](#ui组件)
- [使用指南](#使用指南)
- [最佳实践](#最佳实践)

## 🧩 组件系统概述

Privacy Gateway Frontend 采用基于类的组件系统，提供了一套完整的UI组件库和开发框架。

### 核心特性

- **生命周期管理** - 完整的组件生命周期钩子
- **状态管理** - 响应式状态更新和管理
- **事件系统** - 灵活的事件处理和委托
- **DOM操作** - 封装的DOM查询和操作方法
- **子组件管理** - 层次化的组件树管理

### 架构设计

```
Component (基类)
├── Modal (模态对话框)
├── DataTable (数据表格)
├── Form (表单组件)
└── 自定义组件...
```

## 🏗️ 基础组件类

### Component

所有组件的基类，提供核心功能和生命周期管理。

#### 构造函数

```javascript
constructor(container, options = {})
```

**参数**
- `container` - DOM元素或选择器字符串
- `options` - 组件配置选项

#### 生命周期方法

```javascript
class MyComponent extends Component {
  // 组件挂载前
  beforeMount() {
    console.log('组件即将挂载');
  }
  
  // 组件挂载
  mount() {
    this.render();
    this.bindEvents();
  }
  
  // 组件挂载后
  afterMount() {
    console.log('组件已挂载');
    this.emit('mounted', { component: this });
  }
  
  // 组件销毁前
  beforeDestroy() {
    console.log('组件即将销毁');
  }
  
  // 组件销毁后
  afterDestroy() {
    console.log('组件已销毁');
  }
}
```

#### 状态管理

```javascript
// 设置状态
this.setState({ count: 5 });

// 函数式状态更新
this.setState(prevState => ({ 
  count: prevState.count + 1 
}));

// 获取状态
const count = this.getState('count');
const allState = this.getState();

// 状态更新回调
this.setState({ count: 5 }, () => {
  console.log('状态已更新');
});
```

#### 事件处理

```javascript
// 添加事件监听器
this.addEventListener('click', this.handleClick.bind(this));

// 事件委托
this.delegate('.button', 'click', this.handleButtonClick.bind(this));

// 触发自定义事件
this.emit('customEvent', { data: 'value' });

// 监听自定义事件
this.on('customEvent', (event) => {
  console.log(event.detail.data);
});
```

#### DOM操作

```javascript
// 查找单个元素
const button = this.$('.button');

// 查找多个元素
const buttons = this.$$('.button');

// 显示/隐藏组件
this.show();
this.hide();
```

## 🎨 UI组件

### Modal 组件

模态对话框组件，支持多种类型和自定义配置。

#### 基本用法

```javascript
import { Modal } from '@components/Modal.js';

// 信息对话框
Modal.info({
  title: '提示',
  content: '这是一条信息',
  onConfirm: () => console.log('确认')
});

// 确认对话框
Modal.confirm({
  title: '确认删除',
  content: '确定要删除这个项目吗？',
  confirmText: '删除',
  cancelText: '取消',
  onConfirm: () => {
    // 删除逻辑
    return true; // 返回true关闭对话框
  },
  onCancel: () => {
    console.log('取消删除');
  }
});

// 自定义对话框
Modal.custom({
  title: '自定义内容',
  content: '<div>自定义HTML内容</div>',
  size: 'large',
  showFooter: false
});
```

#### 配置选项

```javascript
const options = {
  title: '对话框标题',           // 标题文本
  content: '对话框内容',         // 内容HTML或文本
  size: 'medium',              // 尺寸: small, medium, large
  type: 'info',                // 类型: info, confirm, custom
  showHeader: true,            // 是否显示头部
  showFooter: true,            // 是否显示底部
  showCloseButton: true,       // 是否显示关闭按钮
  confirmText: '确定',         // 确认按钮文本
  cancelText: '取消',          // 取消按钮文本
  maskClosable: true,          // 点击遮罩是否关闭
  keyboard: true,              // 是否支持键盘ESC关闭
  onConfirm: () => {},         // 确认回调
  onCancel: () => {},          // 取消回调
  onClose: () => {}            // 关闭回调
};
```

### DataTable 组件

数据表格组件，支持排序、分页、搜索等功能。

#### 基本用法

```javascript
import { DataTable } from '@components/DataTable.js';

const table = new DataTable(container, {
  columns: [
    { 
      key: 'id', 
      title: 'ID', 
      sortable: true,
      width: '80px'
    },
    { 
      key: 'name', 
      title: '名称', 
      sortable: true 
    },
    { 
      key: 'status', 
      title: '状态',
      render: (value, row) => {
        return value ? 
          '<span class="text-green-600">启用</span>' : 
          '<span class="text-red-600">禁用</span>';
      }
    },
    {
      key: 'actions',
      title: '操作',
      render: (value, row) => {
        return `
          <button class="btn-edit" data-id="${row.id}">编辑</button>
          <button class="btn-delete" data-id="${row.id}">删除</button>
        `;
      }
    }
  ],
  data: [
    { id: 1, name: '配置1', status: true },
    { id: 2, name: '配置2', status: false }
  ],
  pagination: true,
  pageSize: 10,
  searchable: true,
  selectable: true
});

// 监听行选择
table.on('rowSelect', (event) => {
  console.log('选中的行:', event.detail.selectedRows);
});

// 监听排序
table.on('sort', (event) => {
  console.log('排序:', event.detail.column, event.detail.direction);
});
```

#### 配置选项

```javascript
const options = {
  columns: [],                 // 列配置数组
  data: [],                   // 数据数组
  pagination: false,          // 是否启用分页
  pageSize: 20,              // 每页数量
  searchable: false,         // 是否启用搜索
  selectable: false,         // 是否启用行选择
  sortable: true,            // 是否启用排序
  loading: false,            // 是否显示加载状态
  emptyText: '暂无数据',      // 空数据提示
  height: 'auto',            // 表格高度
  striped: true,             // 是否显示斑马纹
  bordered: true,            // 是否显示边框
  hover: true                // 是否启用悬停效果
};
```

#### 方法

```javascript
// 更新数据
table.setData(newData);

// 获取选中行
const selected = table.getSelectedRows();

// 清空选择
table.clearSelection();

// 刷新表格
table.refresh();

// 设置加载状态
table.setLoading(true);
```

### Form 组件

表单组件，支持多种输入类型和验证。

#### 基本用法

```javascript
import { Form } from '@components/Form.js';

const form = new Form(container, {
  fields: [
    {
      name: 'name',
      label: '名称',
      type: 'text',
      required: true,
      placeholder: '请输入名称'
    },
    {
      name: 'email',
      label: '邮箱',
      type: 'email',
      required: true,
      validation: {
        pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
        message: '请输入有效的邮箱地址'
      }
    },
    {
      name: 'status',
      label: '状态',
      type: 'switch',
      switchLabel: '启用'
    },
    {
      name: 'category',
      label: '分类',
      type: 'select',
      options: [
        { value: 'api', label: 'API' },
        { value: 'web', label: 'Web' }
      ]
    }
  ],
  onSubmit: (data) => {
    console.log('表单数据:', data);
    // 处理表单提交
  }
});

// 设置表单值
form.setValues({
  name: '示例名称',
  email: 'example@email.com'
});

// 获取表单值
const values = form.getValues();

// 验证表单
const isValid = form.validate();
```

#### 字段类型

- `text` - 文本输入框
- `email` - 邮箱输入框
- `password` - 密码输入框
- `number` - 数字输入框
- `textarea` - 多行文本框
- `select` - 下拉选择框
- `checkbox` - 复选框
- `radio` - 单选框
- `switch` - 开关
- `date` - 日期选择器

## 📖 使用指南

### 创建自定义组件

```javascript
import { Component } from '@components/Component.js';

class CustomComponent extends Component {
  getDefaultOptions() {
    return {
      title: '默认标题',
      showHeader: true
    };
  }
  
  getInitialState() {
    return {
      count: 0,
      visible: true
    };
  }
  
  render() {
    const { title, showHeader } = this.options;
    const { count, visible } = this.state;
    
    this.container.innerHTML = `
      <div class="custom-component ${visible ? 'visible' : 'hidden'}">
        ${showHeader ? `<h3>${title}</h3>` : ''}
        <div class="content">
          <p>计数: ${count}</p>
          <button class="increment">增加</button>
          <button class="decrement">减少</button>
        </div>
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

// 使用组件
const component = new CustomComponent('#container', {
  title: '自定义组件',
  showHeader: true
});
```

### 组件通信

```javascript
// 父子组件通信
class ParentComponent extends Component {
  afterMount() {
    // 创建子组件
    const child = new ChildComponent(this.$('.child-container'));
    this.addChild('child', child);
    
    // 监听子组件事件
    child.on('dataChange', (event) => {
      console.log('子组件数据变化:', event.detail);
    });
  }
}

class ChildComponent extends Component {
  handleDataChange(newData) {
    // 触发事件通知父组件
    this.emit('dataChange', { data: newData });
  }
}
```

### 组件管理器

```javascript
import { componentManager } from '@utils/ComponentManager.js';

// 注册组件
componentManager.register('custom-component', CustomComponent);

// 批量注册
componentManager.registerBatch({
  'modal': Modal,
  'data-table': DataTable,
  'form': Form
});

// 创建组件实例
const component = componentManager.create('custom-component', container, options);

// 自动初始化页面中的组件
componentManager.autoInit();

// 销毁所有组件
componentManager.destroyAll();
```

## 💡 最佳实践

### 1. 组件设计原则

- **单一职责** - 每个组件只负责一个功能
- **可复用性** - 设计通用的、可配置的组件
- **可测试性** - 编写易于测试的组件代码
- **性能优化** - 避免不必要的DOM操作和重渲染

### 2. 状态管理

```javascript
// ✅ 好的做法
this.setState({ 
  user: { ...this.state.user, name: 'new name' } 
});

// ❌ 避免直接修改状态
this.state.user.name = 'new name';
```

### 3. 事件处理

```javascript
// ✅ 使用事件委托
this.delegate('.button', 'click', this.handleClick);

// ❌ 避免为每个元素单独绑定事件
this.$$('.button').forEach(btn => {
  btn.addEventListener('click', this.handleClick);
});
```

### 4. 内存管理

```javascript
class MyComponent extends Component {
  beforeDestroy() {
    // 清理定时器
    if (this.timer) {
      clearInterval(this.timer);
    }
    
    // 清理外部事件监听器
    window.removeEventListener('resize', this.handleResize);
    
    // 清理其他资源
    this.cleanup();
  }
}
```

### 5. 错误处理

```javascript
class MyComponent extends Component {
  async loadData() {
    try {
      this.setState({ loading: true });
      const data = await api.getData();
      this.setState({ data, loading: false });
    } catch (error) {
      console.error('数据加载失败:', error);
      this.setState({ 
        error: error.message, 
        loading: false 
      });
      this.emit('error', { error });
    }
  }
}
```

### 6. 性能优化

```javascript
class MyComponent extends Component {
  shouldUpdate(newState, oldState) {
    // 只在必要时重新渲染
    return newState.data !== oldState.data;
  }
  
  render() {
    // 使用文档片段减少DOM操作
    const fragment = document.createDocumentFragment();
    // ... 构建内容
    this.container.appendChild(fragment);
  }
}
```

## 🔧 调试和开发

### 组件调试

```javascript
// 启用组件调试模式
componentManager.debug();

// 查看组件树
console.log(componentManager.getComponentTree());

// 查看组件状态
console.log(component.getState());

// 监听所有组件事件
componentManager.on('*', (event) => {
  console.log('组件事件:', event);
});
```

### 开发工具

在浏览器控制台中可以使用以下全局变量：

- `window.componentManager` - 组件管理器
- `window.stateManager` - 状态管理器
- `window.eventBus` - 事件总线

### 测试组件

```javascript
// 组件单元测试示例
describe('CustomComponent', () => {
  let component;
  let container;
  
  beforeEach(() => {
    container = document.createElement('div');
    component = new CustomComponent(container);
  });
  
  afterEach(() => {
    component.destroy();
  });
  
  it('应该正确初始化', () => {
    expect(component.state.count).toBe(0);
    expect(component.container).toBe(container);
  });
  
  it('应该响应点击事件', () => {
    const button = component.$('.increment');
    button.click();
    expect(component.state.count).toBe(1);
  });
});
```
