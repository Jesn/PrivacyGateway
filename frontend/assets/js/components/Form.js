/**
 * 表单组件
 * 支持表单验证、数据绑定、自动提交等功能
 */

import { Component } from './Component.js';
import { ValidationUtils } from '../utils/validation.js';
import { DOMUtils } from '../utils/dom.js';

class Form extends Component {
    getDefaultOptions() {
        return {
            fields: [],
            data: {},
            rules: {},
            layout: 'vertical', // vertical, horizontal, inline
            labelWidth: '100px',
            showRequired: true,
            validateOnChange: true,
            validateOnBlur: true,
            submitOnEnter: true,
            resetOnSubmit: false,
            className: '',
            onSubmit: null,
            onValidate: null,
            onFieldChange: null
        };
    }

    getInitialState() {
        return {
            data: { ...this.options.data },
            errors: {},
            touched: new Set(),
            submitting: false,
            valid: true
        };
    }

    render() {
        const { layout, className } = this.options;
        
        this.container.innerHTML = `
            <form class="form form-${layout} ${className}" data-role="form">
                ${this.renderFields()}
                ${this.renderActions()}
            </form>
        `;
    }

    renderFields() {
        const { fields } = this.options;
        
        return fields.map(field => this.renderField(field)).join('');
    }

    renderField(field) {
        const { layout, labelWidth, showRequired } = this.options;
        const { data, errors } = this.state;
        const value = data[field.name] || '';
        const error = errors[field.name];
        const isRequired = this.isFieldRequired(field.name);
        
        const fieldClass = [
            'form-field',
            `field-${field.type}`,
            error ? 'has-error' : '',
            field.className || ''
        ].filter(Boolean).join(' ');
        
        return `
            <div class="${fieldClass}" data-field="${field.name}">
                ${this.renderLabel(field, isRequired, showRequired)}
                <div class="field-control" ${layout === 'horizontal' ? `style="margin-left: ${labelWidth}"` : ''}>
                    ${this.renderInput(field, value)}
                    ${error ? `<div class="field-error">${error}</div>` : ''}
                    ${field.help ? `<div class="field-help">${field.help}</div>` : ''}
                </div>
            </div>
        `;
    }

    renderLabel(field, isRequired, showRequired) {
        const { layout, labelWidth } = this.options;
        
        if (!field.label) return '';
        
        const labelStyle = layout === 'horizontal' ? `style="width: ${labelWidth}"` : '';
        const requiredMark = isRequired && showRequired ? '<span class="required">*</span>' : '';
        
        return `
            <label class="field-label" for="${field.name}" ${labelStyle}>
                ${field.label}${requiredMark}
            </label>
        `;
    }

    renderInput(field, value) {
        const commonAttrs = `
            name="${field.name}"
            id="${field.name}"
            data-field="${field.name}"
            ${field.placeholder ? `placeholder="${field.placeholder}"` : ''}
            ${field.disabled ? 'disabled' : ''}
            ${field.readonly ? 'readonly' : ''}
            ${field.required ? 'required' : ''}
        `;

        switch (field.type) {
            case 'text':
            case 'email':
            case 'url':
            case 'password':
                return `<input type="${field.type}" class="form-input" value="${value}" ${commonAttrs}>`;
            
            case 'number':
                return `<input type="number" class="form-input" value="${value}" 
                        ${field.min !== undefined ? `min="${field.min}"` : ''}
                        ${field.max !== undefined ? `max="${field.max}"` : ''}
                        ${field.step !== undefined ? `step="${field.step}"` : ''}
                        ${commonAttrs}>`;
            
            case 'textarea':
                return `<textarea class="form-textarea" rows="${field.rows || 3}" ${commonAttrs}>${value}</textarea>`;
            
            case 'select':
                return this.renderSelect(field, value, commonAttrs);
            
            case 'radio':
                return this.renderRadio(field, value, commonAttrs);
            
            case 'checkbox':
                return this.renderCheckbox(field, value, commonAttrs);
            
            case 'switch':
                return this.renderSwitch(field, value, commonAttrs);
            
            case 'file':
                return `<input type="file" class="form-file" 
                        ${field.accept ? `accept="${field.accept}"` : ''}
                        ${field.multiple ? 'multiple' : ''}
                        ${commonAttrs}>`;
            
            case 'hidden':
                return `<input type="hidden" value="${value}" ${commonAttrs}>`;
            
            default:
                return `<input type="text" class="form-input" value="${value}" ${commonAttrs}>`;
        }
    }

    renderSelect(field, value, commonAttrs) {
        const options = field.options || [];
        
        return `
            <select class="form-select" ${commonAttrs}>
                ${field.placeholder ? `<option value="">${field.placeholder}</option>` : ''}
                ${options.map(option => {
                    const optionValue = typeof option === 'object' ? option.value : option;
                    const optionLabel = typeof option === 'object' ? option.label : option;
                    const selected = value == optionValue ? 'selected' : '';
                    return `<option value="${optionValue}" ${selected}>${optionLabel}</option>`;
                }).join('')}
            </select>
        `;
    }

    renderRadio(field, value, commonAttrs) {
        const options = field.options || [];
        
        return options.map(option => {
            const optionValue = typeof option === 'object' ? option.value : option;
            const optionLabel = typeof option === 'object' ? option.label : option;
            const checked = value == optionValue ? 'checked' : '';
            
            return `
                <label class="radio-option">
                    <input type="radio" value="${optionValue}" ${checked} ${commonAttrs}>
                    <span class="radio-label">${optionLabel}</span>
                </label>
            `;
        }).join('');
    }

    renderCheckbox(field, value, commonAttrs) {
        if (field.options) {
            // 多选框组
            const selectedValues = Array.isArray(value) ? value : [];
            
            return field.options.map(option => {
                const optionValue = typeof option === 'object' ? option.value : option;
                const optionLabel = typeof option === 'object' ? option.label : option;
                const checked = selectedValues.includes(optionValue) ? 'checked' : '';
                
                return `
                    <label class="checkbox-option">
                        <input type="checkbox" value="${optionValue}" ${checked} ${commonAttrs}>
                        <span class="checkbox-label">${optionLabel}</span>
                    </label>
                `;
            }).join('');
        } else {
            // 单个复选框
            const checked = value ? 'checked' : '';
            return `
                <label class="checkbox-option">
                    <input type="checkbox" ${checked} ${commonAttrs}>
                    <span class="checkbox-label">${field.checkboxLabel || field.label}</span>
                </label>
            `;
        }
    }

    renderSwitch(field, value, commonAttrs) {
        const checked = value ? 'checked' : '';
        
        return `
            <label class="switch">
                <input type="checkbox" ${checked} ${commonAttrs}>
                <span class="switch-slider"></span>
                ${field.switchLabel ? `<span class="switch-label">${field.switchLabel}</span>` : ''}
            </label>
        `;
    }

    renderActions() {
        const { actions = [] } = this.options;
        
        if (actions.length === 0) {
            return `
                <div class="form-actions">
                    <button type="submit" class="btn btn-primary">提交</button>
                    <button type="button" class="btn btn-secondary" data-action="reset">重置</button>
                </div>
            `;
        }
        
        return `
            <div class="form-actions">
                ${actions.map(action => `
                    <button type="${action.type || 'button'}" 
                            class="btn ${action.className || 'btn-primary'}"
                            data-action="${action.action || ''}"
                            ${action.disabled ? 'disabled' : ''}>
                        ${action.text}
                    </button>
                `).join('')}
            </div>
        `;
    }

    bindEvents() {
        // 表单提交
        this.addEventListener('submit', this.handleSubmit.bind(this));
        
        // 字段变化
        this.delegate('input, select, textarea', 'input', this.handleFieldChange.bind(this));
        this.delegate('input, select, textarea', 'change', this.handleFieldChange.bind(this));
        
        // 字段失焦验证
        if (this.options.validateOnBlur) {
            this.delegate('input, select, textarea', 'blur', this.handleFieldBlur.bind(this));
        }
        
        // 回车提交
        if (this.options.submitOnEnter) {
            this.delegate('input', 'keydown', this.handleKeydown.bind(this));
        }
        
        // 动作按钮
        this.delegate('[data-action]', 'click', this.handleAction.bind(this));
    }

    handleSubmit(e) {
        e.preventDefault();
        
        if (this.state.submitting) return;
        
        const isValid = this.validate();
        if (!isValid) return;
        
        this.setState({ submitting: true });
        
        if (this.options.onSubmit) {
            Promise.resolve(this.options.onSubmit(this.state.data))
                .then(() => {
                    if (this.options.resetOnSubmit) {
                        this.reset();
                    }
                })
                .catch(error => {
                    console.error('表单提交失败:', error);
                })
                .finally(() => {
                    this.setState({ submitting: false });
                });
        } else {
            this.setState({ submitting: false });
        }
    }

    handleFieldChange(e) {
        const field = e.target;
        const name = field.name;
        const value = this.getFieldValue(field);
        
        // 更新数据
        this.setState({
            data: { ...this.state.data, [name]: value },
            touched: new Set([...this.state.touched, name])
        });
        
        // 实时验证
        if (this.options.validateOnChange && this.state.touched.has(name)) {
            this.validateField(name);
        }
        
        // 触发字段变化事件
        if (this.options.onFieldChange) {
            this.options.onFieldChange(name, value, this.state.data);
        }
    }

    handleFieldBlur(e) {
        const name = e.target.name;
        this.setState({
            touched: new Set([...this.state.touched, name])
        });
        
        this.validateField(name);
    }

    handleKeydown(e) {
        if (e.key === 'Enter' && e.target.type !== 'textarea') {
            e.preventDefault();
            this.handleSubmit(e);
        }
    }

    handleAction(e) {
        const action = e.target.dataset.action;
        
        switch (action) {
            case 'reset':
                this.reset();
                break;
            case 'clear':
                this.clear();
                break;
            default:
                this.emit('action', { action, form: this });
        }
    }

    getFieldValue(field) {
        switch (field.type) {
            case 'checkbox':
                if (field.hasAttribute('data-field')) {
                    // 复选框组
                    const checkboxes = this.$$(`input[name="${field.name}"]:checked`);
                    return Array.from(checkboxes).map(cb => cb.value);
                } else {
                    // 单个复选框
                    return field.checked;
                }
            case 'radio':
                const radioChecked = this.$(`input[name="${field.name}"]:checked`);
                return radioChecked ? radioChecked.value : '';
            case 'file':
                return field.files;
            case 'number':
                return field.value ? parseFloat(field.value) : '';
            default:
                return field.value;
        }
    }

    isFieldRequired(fieldName) {
        const rules = this.options.rules[fieldName];
        return rules && rules.required;
    }

    validate() {
        const { data } = this.state;
        const { rules } = this.options;
        
        const result = ValidationUtils.validateForm(data, rules);
        
        this.setState({
            errors: result.errors,
            valid: result.isValid
        });
        
        if (this.options.onValidate) {
            this.options.onValidate(result.isValid, result.errors);
        }
        
        return result.isValid;
    }

    validateField(fieldName) {
        const { data } = this.state;
        const { rules } = this.options;
        
        const fieldRules = rules[fieldName];
        if (!fieldRules) return true;
        
        const result = ValidationUtils.validateForm(
            { [fieldName]: data[fieldName] },
            { [fieldName]: fieldRules }
        );
        
        this.setState({
            errors: {
                ...this.state.errors,
                [fieldName]: result.errors[fieldName]
            }
        });
        
        return result.isValid;
    }

    /**
     * 设置表单数据
     * @param {Object} data - 表单数据
     */
    setData(data) {
        this.setState({ data: { ...this.state.data, ...data } });
        this.render();
    }

    /**
     * 获取表单数据
     * @returns {Object} 表单数据
     */
    getData() {
        return { ...this.state.data };
    }

    /**
     * 重置表单
     */
    reset() {
        this.setState({
            data: { ...this.options.data },
            errors: {},
            touched: new Set(),
            submitting: false,
            valid: true
        });
        this.render();
    }

    /**
     * 清空表单
     */
    clear() {
        const emptyData = {};
        this.options.fields.forEach(field => {
            if (field.type === 'checkbox' && field.options) {
                emptyData[field.name] = [];
            } else {
                emptyData[field.name] = '';
            }
        });
        
        this.setState({
            data: emptyData,
            errors: {},
            touched: new Set()
        });
        this.render();
    }

    /**
     * 设置字段错误
     * @param {string} fieldName - 字段名
     * @param {string} error - 错误信息
     */
    setFieldError(fieldName, error) {
        this.setState({
            errors: {
                ...this.state.errors,
                [fieldName]: error
            }
        });
    }

    /**
     * 清除字段错误
     * @param {string} fieldName - 字段名
     */
    clearFieldError(fieldName) {
        const errors = { ...this.state.errors };
        delete errors[fieldName];
        this.setState({ errors });
    }

    /**
     * 设置提交状态
     * @param {boolean} submitting - 是否提交中
     */
    setSubmitting(submitting) {
        this.setState({ submitting });
    }
}

export { Form };
