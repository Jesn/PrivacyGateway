/**
 * 模态框组件
 * 支持多种类型的弹框：确认、表单、信息展示等
 */

import { Component } from './Component.js';
import { DOMUtils } from '../utils/dom.js';

class Modal extends Component {
    getDefaultOptions() {
        return {
            title: '提示',
            content: '',
            type: 'info', // info, confirm, form, custom
            size: 'medium', // small, medium, large
            closable: true,
            maskClosable: true,
            showFooter: true,
            confirmText: '确定',
            cancelText: '取消',
            className: '',
            zIndex: 1000,
            animation: true,
            autoFocus: true,
            destroyOnClose: true,
            onConfirm: null,
            onCancel: null,
            onClose: null
        };
    }

    getInitialState() {
        return {
            visible: false,
            loading: false,
            data: null
        };
    }

    render() {
        if (!this.state.visible) {
            this.container.style.display = 'none';
            return;
        }

        const { title, content, type, size, closable, showFooter, className } = this.options;
        
        this.container.innerHTML = `
            <div class="modal-overlay" data-role="overlay">
                <div class="modal-dialog modal-${size} ${className}" data-role="dialog">
                    ${this.renderHeader()}
                    <div class="modal-body" data-role="body">
                        ${this.renderContent()}
                    </div>
                    ${showFooter ? this.renderFooter() : ''}
                </div>
            </div>
        `;

        this.container.style.display = 'flex';
        this.container.style.zIndex = this.options.zIndex;
        
        // 添加动画类
        if (this.options.animation) {
            setTimeout(() => {
                DOMUtils.addClass(this.container, 'modal-show');
            }, 10);
        }

        // 自动聚焦
        if (this.options.autoFocus) {
            this.autoFocus();
        }
    }

    renderHeader() {
        const { title, closable } = this.options;
        
        return `
            <div class="modal-header">
                <h3 class="modal-title">${title}</h3>
                ${closable ? '<button class="modal-close" data-role="close" aria-label="关闭">&times;</button>' : ''}
            </div>
        `;
    }

    renderContent() {
        const { content, type } = this.options;
        
        switch (type) {
            case 'confirm':
                return this.renderConfirmContent();
            case 'form':
                return this.renderFormContent();
            case 'loading':
                return this.renderLoadingContent();
            default:
                return typeof content === 'string' ? content : '';
        }
    }

    renderConfirmContent() {
        const { content } = this.options;
        const icon = this.getConfirmIcon();
        
        return `
            <div class="modal-confirm">
                <div class="modal-confirm-icon">${icon}</div>
                <div class="modal-confirm-content">${content}</div>
            </div>
        `;
    }

    renderFormContent() {
        const { content } = this.options;
        return `<div class="modal-form">${content}</div>`;
    }

    renderLoadingContent() {
        return `
            <div class="modal-loading">
                <div class="loading-spinner"></div>
                <div class="loading-text">${this.options.content || '加载中...'}</div>
            </div>
        `;
    }

    renderFooter() {
        const { type, confirmText, cancelText } = this.options;
        const { loading } = this.state;
        
        if (type === 'loading') {
            return '';
        }
        
        return `
            <div class="modal-footer">
                ${type === 'confirm' || type === 'form' ? 
                    `<button class="btn btn-secondary" data-role="cancel" ${loading ? 'disabled' : ''}>${cancelText}</button>` : 
                    ''
                }
                <button class="btn btn-primary" data-role="confirm" ${loading ? 'disabled' : ''}>
                    ${loading ? '<span class="loading"></span>' : ''} ${confirmText}
                </button>
            </div>
        `;
    }

    getConfirmIcon() {
        const { confirmType = 'warning' } = this.options;
        
        const icons = {
            warning: '⚠️',
            error: '❌',
            success: '✅',
            info: 'ℹ️',
            question: '❓'
        };
        
        return icons[confirmType] || icons.warning;
    }

    bindEvents() {
        // 关闭按钮
        this.delegate('[data-role="close"]', 'click', this.handleClose.bind(this));
        
        // 确认按钮
        this.delegate('[data-role="confirm"]', 'click', this.handleConfirm.bind(this));
        
        // 取消按钮
        this.delegate('[data-role="cancel"]', 'click', this.handleCancel.bind(this));
        
        // 遮罩点击
        if (this.options.maskClosable) {
            this.delegate('[data-role="overlay"]', 'click', (e) => {
                if (e.target === e.currentTarget) {
                    this.handleClose();
                }
            });
        }
        
        // ESC键关闭
        this.addEventListener('keydown', this.handleKeydown.bind(this), document);
    }

    handleClose() {
        if (this.options.onClose) {
            const result = this.options.onClose();
            if (result === false) return;
        }
        
        this.close();
    }

    async handleConfirm() {
        if (this.state.loading) return;
        
        try {
            if (this.options.onConfirm) {
                this.setState({ loading: true });
                
                const result = await this.options.onConfirm(this.state.data);
                
                if (result !== false) {
                    this.close();
                }
            } else {
                this.close();
            }
        } catch (error) {
            console.error('确认操作失败:', error);
        } finally {
            this.setState({ loading: false });
        }
    }

    handleCancel() {
        if (this.options.onCancel) {
            const result = this.options.onCancel();
            if (result === false) return;
        }
        
        this.close();
    }

    handleKeydown(e) {
        if (e.key === 'Escape' && this.options.closable) {
            this.handleClose();
        }
    }

    autoFocus() {
        setTimeout(() => {
            const focusElement = this.$('[autofocus]') || 
                                this.$('input, textarea, select') || 
                                this.$('[data-role="confirm"]');
            
            if (focusElement) {
                focusElement.focus();
            }
        }, 100);
    }

    /**
     * 显示模态框
     * @param {Object} data - 传递给模态框的数据
     */
    show(data = null) {
        this.setState({ visible: true, data });
        document.body.style.overflow = 'hidden';
        this.emit('show', { modal: this, data });
    }

    /**
     * 隐藏模态框
     */
    hide() {
        this.setState({ visible: false });
        document.body.style.overflow = '';
        this.emit('hide', { modal: this });
    }

    /**
     * 关闭模态框
     */
    close() {
        if (this.options.animation) {
            DOMUtils.removeClass(this.container, 'modal-show');
            setTimeout(() => {
                this.hide();
                if (this.options.destroyOnClose) {
                    this.destroy();
                }
            }, 300);
        } else {
            this.hide();
            if (this.options.destroyOnClose) {
                this.destroy();
            }
        }
        
        this.emit('close', { modal: this });
    }

    /**
     * 设置加载状态
     * @param {boolean} loading - 是否加载中
     */
    setLoading(loading) {
        this.setState({ loading });
    }

    /**
     * 更新内容
     * @param {string} content - 新内容
     */
    setContent(content) {
        this.options.content = content;
        const bodyEl = this.$('[data-role="body"]');
        if (bodyEl) {
            bodyEl.innerHTML = this.renderContent();
        }
    }

    /**
     * 更新标题
     * @param {string} title - 新标题
     */
    setTitle(title) {
        this.options.title = title;
        const titleEl = this.$('.modal-title');
        if (titleEl) {
            titleEl.textContent = title;
        }
    }

    beforeDestroy() {
        super.beforeDestroy();
        document.body.style.overflow = '';

        // 移除容器元素
        if (this.container && this.container.parentNode) {
            this.container.parentNode.removeChild(this.container);
        }
    }

    // 静态方法：快速创建各种类型的模态框
    
    /**
     * 显示确认对话框
     * @param {Object} options - 选项
     * @returns {Promise} Promise对象
     */
    static confirm(options) {
        return new Promise((resolve, reject) => {
            const container = DOMUtils.create('div', { className: 'modal' });
            document.body.appendChild(container);
            
            const modal = new Modal(container, {
                type: 'confirm',
                confirmType: 'question',
                ...options,
                onConfirm: async () => {
                    try {
                        if (options.onConfirm) {
                            await options.onConfirm();
                        }
                        resolve(true);
                    } catch (error) {
                        reject(error);
                    }
                },
                onCancel: () => {
                    resolve(false);
                },
                destroyOnClose: true
            });
            
            modal.show();
        });
    }

    /**
     * 显示警告对话框
     * @param {Object} options - 选项
     * @returns {Promise} Promise对象
     */
    static alert(options) {
        return new Promise((resolve) => {
            const container = DOMUtils.create('div', { className: 'modal' });
            document.body.appendChild(container);
            
            const modal = new Modal(container, {
                type: 'confirm',
                confirmType: 'warning',
                showFooter: true,
                ...options,
                onConfirm: () => {
                    resolve(true);
                },
                destroyOnClose: true
            });
            
            modal.show();
        });
    }

    /**
     * 显示信息对话框
     * @param {Object} options - 选项
     * @returns {Modal} Modal实例
     */
    static info(options) {
        const container = DOMUtils.create('div', { className: 'modal' });
        document.body.appendChild(container);
        
        const modal = new Modal(container, {
            type: 'info',
            ...options,
            destroyOnClose: true
        });
        
        modal.show();
        return modal;
    }
}

export { Modal };
