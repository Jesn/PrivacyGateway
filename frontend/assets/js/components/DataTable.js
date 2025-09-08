/**
 * 数据表格组件
 * 支持数据渲染、分页、排序、选择、筛选等功能
 */

import { Component } from './Component.js';
import { DOMUtils } from '../utils/dom.js';

class DataTable extends Component {
    getDefaultOptions() {
        return {
            columns: [],
            data: [],
            pagination: true,
            pageSize: 10,
            currentPage: 1,
            sortable: true,
            selectable: false,
            multiSelect: false,
            searchable: false,
            filterable: false,
            loading: false,
            emptyText: '暂无数据',
            rowKey: 'id',
            className: '',
            height: 'auto',
            onRowClick: null,
            onRowSelect: null,
            onSort: null,
            onPageChange: null,
            onSearch: null,
            onFilter: null
        };
    }

    getInitialState() {
        return {
            sortColumn: null,
            sortDirection: 'asc',
            selectedRows: new Set(),
            searchKeyword: '',
            filters: {},
            currentData: [],
            totalPages: 0,
            loading: false
        };
    }

    beforeMount() {
        this.processData();
    }

    render() {
        const { className, height } = this.options;
        
        this.container.innerHTML = `
            <div class="data-table-wrapper ${className}">
                ${this.renderToolbar()}
                <div class="data-table-container" style="height: ${height}">
                    ${this.renderTable()}
                </div>
                ${this.options.pagination ? this.renderPagination() : ''}
            </div>
        `;
    }

    renderToolbar() {
        const { searchable, selectable } = this.options;
        const { selectedRows } = this.state;
        
        if (!searchable && !selectable) return '';
        
        return `
            <div class="data-table-toolbar">
                ${selectable && selectedRows.size > 0 ? this.renderBulkActions() : ''}
                ${searchable ? this.renderSearchBox() : ''}
            </div>
        `;
    }

    renderBulkActions() {
        const { selectedRows } = this.state;
        
        return `
            <div class="bulk-actions">
                <span class="selected-count">已选择 ${selectedRows.size} 项</span>
                <button class="btn btn-sm btn-danger" data-action="bulk-delete">批量删除</button>
                <button class="btn btn-sm btn-secondary" data-action="clear-selection">清除选择</button>
            </div>
        `;
    }

    renderSearchBox() {
        const { searchKeyword } = this.state;
        
        return `
            <div class="search-box">
                <input type="text" 
                       class="form-input" 
                       placeholder="搜索..." 
                       value="${searchKeyword}"
                       data-role="search">
                <button class="btn btn-sm btn-primary" data-action="search">搜索</button>
            </div>
        `;
    }

    renderTable() {
        const { columns, selectable, loading } = this.options;
        const { currentData } = this.state;
        
        if (loading) {
            return this.renderLoading();
        }
        
        return `
            <table class="data-table">
                <thead>
                    <tr>
                        ${selectable ? '<th class="select-column"><input type="checkbox" data-role="select-all"></th>' : ''}
                        ${columns.map(col => this.renderHeaderCell(col)).join('')}
                    </tr>
                </thead>
                <tbody>
                    ${currentData.length > 0 ? 
                        currentData.map(row => this.renderRow(row)).join('') : 
                        this.renderEmptyRow()
                    }
                </tbody>
            </table>
        `;
    }

    renderHeaderCell(column) {
        const { sortable } = this.options;
        const { sortColumn, sortDirection } = this.state;
        const isSorted = sortColumn === column.key;
        const sortClass = isSorted ? `sorted ${sortDirection}` : '';
        
        return `
            <th class="table-header ${sortClass} ${sortable && column.sortable !== false ? 'sortable' : ''}" 
                data-column="${column.key}"
                data-sortable="${sortable && column.sortable !== false}">
                <span class="header-content">
                    ${column.title}
                    ${sortable && column.sortable !== false ? '<span class="sort-indicator"></span>' : ''}
                </span>
            </th>
        `;
    }

    renderRow(row) {
        const { columns, selectable, rowKey } = this.options;
        const { selectedRows } = this.state;
        const rowId = row[rowKey];
        const isSelected = selectedRows.has(rowId);
        
        return `
            <tr class="table-row ${isSelected ? 'selected' : ''}" 
                data-row-id="${rowId}">
                ${selectable ? `
                    <td class="select-column">
                        <input type="checkbox" 
                               data-role="select-row" 
                               data-row-id="${rowId}"
                               ${isSelected ? 'checked' : ''}>
                    </td>
                ` : ''}
                ${columns.map(col => this.renderCell(row, col)).join('')}
            </tr>
        `;
    }

    renderCell(row, column) {
        const value = this.getCellValue(row, column.key);
        const displayValue = column.render ? column.render(value, row) : value;
        
        return `
            <td class="table-cell" data-column="${column.key}">
                ${displayValue}
            </td>
        `;
    }

    renderEmptyRow() {
        const { columns, selectable, emptyText } = this.options;
        const colSpan = columns.length + (selectable ? 1 : 0);
        
        return `
            <tr class="empty-row">
                <td colspan="${colSpan}" class="empty-cell">
                    <div class="empty-content">
                        <div class="empty-icon">📋</div>
                        <div class="empty-text">${emptyText}</div>
                    </div>
                </td>
            </tr>
        `;
    }

    renderLoading() {
        return `
            <div class="table-loading">
                <div class="loading-spinner"></div>
                <div class="loading-text">加载中...</div>
            </div>
        `;
    }

    renderPagination() {
        const { currentPage } = this.options;
        const { totalPages } = this.state;
        
        if (totalPages <= 1) return '';
        
        return `
            <div class="data-table-pagination">
                <button class="page-btn" 
                        data-action="prev-page" 
                        ${currentPage <= 1 ? 'disabled' : ''}>
                    上一页
                </button>
                
                <span class="page-info">
                    第 ${currentPage} 页，共 ${totalPages} 页
                </span>
                
                <button class="page-btn" 
                        data-action="next-page" 
                        ${currentPage >= totalPages ? 'disabled' : ''}>
                    下一页
                </button>
            </div>
        `;
    }

    bindEvents() {
        // 表头点击排序
        this.delegate('.sortable', 'click', this.handleSort.bind(this));
        
        // 行点击
        this.delegate('.table-row', 'click', this.handleRowClick.bind(this));
        
        // 选择相关
        this.delegate('[data-role="select-all"]', 'change', this.handleSelectAll.bind(this));
        this.delegate('[data-role="select-row"]', 'change', this.handleSelectRow.bind(this));
        
        // 批量操作
        this.delegate('[data-action="bulk-delete"]', 'click', this.handleBulkDelete.bind(this));
        this.delegate('[data-action="clear-selection"]', 'click', this.handleClearSelection.bind(this));
        
        // 搜索
        this.delegate('[data-role="search"]', 'input', this.handleSearch.bind(this));
        this.delegate('[data-action="search"]', 'click', this.handleSearchClick.bind(this));
        
        // 分页
        this.delegate('[data-action="prev-page"]', 'click', this.handlePrevPage.bind(this));
        this.delegate('[data-action="next-page"]', 'click', this.handleNextPage.bind(this));
    }

    handleSort(e) {
        const column = e.currentTarget.dataset.column;
        const { sortColumn, sortDirection } = this.state;
        
        let newDirection = 'asc';
        if (sortColumn === column && sortDirection === 'asc') {
            newDirection = 'desc';
        }
        
        this.setState({
            sortColumn: column,
            sortDirection: newDirection
        });
        
        this.processData();
        
        if (this.options.onSort) {
            this.options.onSort(column, newDirection);
        }
    }

    handleRowClick(e) {
        // 如果点击的是选择框，不触发行点击
        if (e.target.type === 'checkbox') return;
        
        const rowId = e.currentTarget.dataset.rowId;
        const row = this.options.data.find(item => item[this.options.rowKey] == rowId);
        
        if (this.options.onRowClick) {
            this.options.onRowClick(row, rowId);
        }
    }

    handleSelectAll(e) {
        const checked = e.target.checked;
        const { currentData } = this.state;
        const { rowKey } = this.options;
        
        if (checked) {
            const newSelected = new Set(this.state.selectedRows);
            currentData.forEach(row => newSelected.add(row[rowKey]));
            this.setState({ selectedRows: newSelected });
        } else {
            this.setState({ selectedRows: new Set() });
        }
        
        this.updateSelectAllState();
        this.notifySelection();
    }

    handleSelectRow(e) {
        const checked = e.target.checked;
        const rowId = e.target.dataset.rowId;
        const newSelected = new Set(this.state.selectedRows);
        
        if (checked) {
            newSelected.add(rowId);
        } else {
            newSelected.delete(rowId);
        }
        
        this.setState({ selectedRows: newSelected });
        this.updateSelectAllState();
        this.notifySelection();
    }

    updateSelectAllState() {
        const selectAllCheckbox = this.$('[data-role="select-all"]');
        if (!selectAllCheckbox) return;
        
        const { currentData } = this.state;
        const { rowKey } = this.options;
        const selectedCount = currentData.filter(row => 
            this.state.selectedRows.has(row[rowKey])
        ).length;
        
        selectAllCheckbox.checked = selectedCount === currentData.length && currentData.length > 0;
        selectAllCheckbox.indeterminate = selectedCount > 0 && selectedCount < currentData.length;
    }

    notifySelection() {
        if (this.options.onRowSelect) {
            const selectedData = this.options.data.filter(row => 
                this.state.selectedRows.has(row[this.options.rowKey])
            );
            this.options.onRowSelect(selectedData, Array.from(this.state.selectedRows));
        }
    }

    getCellValue(row, key) {
        return key.split('.').reduce((obj, k) => obj && obj[k], row);
    }

    processData() {
        let { data } = this.options;
        const { searchKeyword, sortColumn, sortDirection } = this.state;
        
        // 搜索过滤
        if (searchKeyword) {
            data = this.filterData(data, searchKeyword);
        }
        
        // 排序
        if (sortColumn) {
            data = this.sortData(data, sortColumn, sortDirection);
        }
        
        // 分页
        const { pagination, pageSize, currentPage } = this.options;
        let currentData = data;
        let totalPages = 1;
        
        if (pagination) {
            totalPages = Math.ceil(data.length / pageSize);
            const startIndex = (currentPage - 1) * pageSize;
            currentData = data.slice(startIndex, startIndex + pageSize);
        }
        
        this.setState({ currentData, totalPages });
    }

    filterData(data, keyword) {
        const { columns } = this.options;
        const searchableColumns = columns.filter(col => col.searchable !== false);
        
        return data.filter(row => {
            return searchableColumns.some(col => {
                const value = this.getCellValue(row, col.key);
                return String(value).toLowerCase().includes(keyword.toLowerCase());
            });
        });
    }

    sortData(data, column, direction) {
        return [...data].sort((a, b) => {
            const aVal = this.getCellValue(a, column);
            const bVal = this.getCellValue(b, column);
            
            let result = 0;
            if (aVal < bVal) result = -1;
            else if (aVal > bVal) result = 1;
            
            return direction === 'desc' ? -result : result;
        });
    }

    // 公共方法
    
    /**
     * 设置数据
     * @param {Array} data - 新数据
     */
    setData(data) {
        this.options.data = data;
        this.processData();
        this.render();
    }

    /**
     * 获取选中的行
     * @returns {Array} 选中的数据
     */
    getSelectedRows() {
        return this.options.data.filter(row => 
            this.state.selectedRows.has(row[this.options.rowKey])
        );
    }

    /**
     * 清除选择
     */
    clearSelection() {
        this.setState({ selectedRows: new Set() });
        this.render();
    }

    /**
     * 设置加载状态
     * @param {boolean} loading - 是否加载中
     */
    setLoading(loading) {
        this.options.loading = loading;
        this.render();
    }

    /**
     * 跳转到指定页
     * @param {number} page - 页码
     */
    goToPage(page) {
        const { totalPages } = this.state;
        if (page < 1 || page > totalPages) return;
        
        this.options.currentPage = page;
        this.processData();
        this.render();
        
        if (this.options.onPageChange) {
            this.options.onPageChange(page);
        }
    }

    handlePrevPage() {
        this.goToPage(this.options.currentPage - 1);
    }

    handleNextPage() {
        this.goToPage(this.options.currentPage + 1);
    }

    handleSearch(e) {
        const keyword = e.target.value;
        this.setState({ searchKeyword: keyword });
        
        // 防抖处理
        clearTimeout(this.searchTimer);
        this.searchTimer = setTimeout(() => {
            this.processData();
            this.render();
            
            if (this.options.onSearch) {
                this.options.onSearch(keyword);
            }
        }, 300);
    }

    handleSearchClick() {
        const searchInput = this.$('[data-role="search"]');
        if (searchInput) {
            this.handleSearch({ target: searchInput });
        }
    }

    handleBulkDelete() {
        const selectedData = this.getSelectedRows();
        this.emit('bulkDelete', { data: selectedData, ids: Array.from(this.state.selectedRows) });
    }

    handleClearSelection() {
        this.clearSelection();
    }
}

export { DataTable };
