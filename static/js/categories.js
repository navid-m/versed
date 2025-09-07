// Category management functionality
class CategoryManager {
    constructor() {
        this.categories = [];
        this.currentCategory = null;
        this.init();
    }

    init() {
        this.loadCategories();
        this.setupEventListeners();
    }

    async loadCategories() {
        try {
            const response = await fetch('/api/categories');
            if (response.ok) {
                const data = await response.json();
                this.categories = data.categories;
                this.renderCategories();
            }
        } catch (error) {
            console.error('Error loading categories:', error);
        }
    }

    renderCategories() {
        const container = document.getElementById('categoryContainer');
        if (!container) return;

        const categoryButtons = this.categories.map(category => `
            <button
                class="category-btn inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium transition-colors ${this.currentCategory === category.id ? 'bg-gray-900 dark:bg-gray-700 text-white' : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}"
                data-category-id="${category.id}"
                data-category-name="${category.name}"
            >
                ${category.name}
                <span class="ml-2 opacity-50 hover:opacity-100" onclick="categoryManager.showCategoryMenu(event, ${category.id})">
                    <i class="fas fa-ellipsis-h text-xs"></i>
                </span>
            </button>
        `).join('');

        const allButton = `
            <button
                class="category-btn inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-gray-900 dark:bg-gray-700 text-white"
                data-category-id="all"
            >
                All
            </button>
        `;

        const addButton = `
            <button
                class="category-btn inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600"
                onclick="categoryManager.showAddCategoryModal()"
            >
                <i class="fas fa-plus mr-1"></i>
                Add Category
            </button>
        `;

        container.innerHTML = allButton + categoryButtons + addButton;
        this.attachCategoryListeners();
    }

    attachCategoryListeners() {
        document.querySelectorAll('.category-btn[data-category-id]').forEach(btn => {
            btn.addEventListener('click', (e) => {
                if (e.target.closest('.fa-ellipsis-h')) return; // Don't trigger category selection for menu

                const categoryId = btn.getAttribute('data-category-id');
                this.selectCategory(categoryId);
            });
        });
    }

    selectCategory(categoryId) {
        this.currentCategory = categoryId;
        this.renderCategories();
        this.loadCategoryFeeds(categoryId);
    }

    async loadCategoryFeeds(categoryId) {
        console.log('Loading feeds for category:', categoryId);
        try {
            let url;
            if (categoryId === 'all') {
                url = '/api/feeds';
                console.log('Loading all feeds from:', url);
            } else {
                url = `/api/categories/${categoryId}/feeds`;
                console.log('Loading category feeds from:', url);
            }

            const response = await fetch(url);
            console.log('API response status:', response.status);
            if (response.ok) {
                const data = await response.json();
                console.log('API response data:', data);
                this.renderCategoryFeeds(data.feeds || [], categoryId);
            } else {
                console.error('Failed to load category feeds:', response.status);
                this.renderCategoryFeeds([], categoryId);
            }
        } catch (error) {
            console.error('Error loading category feeds:', error);
            this.renderCategoryFeeds([], categoryId);
        }
    }

    renderCategoryFeeds(feeds, categoryId) {
        console.log('Rendering feeds for category:', categoryId, 'feeds count:', feeds.length);
        const container = document.getElementById('postsContainer');
        if (!container) {
            console.error('postsContainer not found!');
            return;
        }

        if (feeds.length === 0) {
            if (categoryId === 'all') {
                console.log('Category is "all", not rendering add feed button');
                // For 'all' category, show the original feed items
                // This should be handled by the existing feed loading logic
                return;
            } else {
                console.log('Category is empty, showing add feed button');
                // For specific categories with no feeds
                container.innerHTML = `
                    <div class="text-center py-12">
                        <div class="w-16 h-16 mx-auto mb-4 text-gray-400 dark:text-gray-600">
                            <i class="fas fa-folder-open text-4xl"></i>
                        </div>
                        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                            No feeds in this category
                        </h3>
                        <p class="text-gray-500 dark:text-gray-400 mb-4">
                            Add RSS feeds or subreddits to organize your content.
                        </p>
                        <button
                            onclick="categoryManager.showAddFeedModal(${categoryId})"
                            class="inline-flex items-center px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                        >
                            <i class="fas fa-plus mr-2"></i>
                            Add Feed
                        </button>
                    </div>
                `;
            }
        } else {
            console.log('Category has feeds, showing feed list with add button');
            // Render feeds for this category with Add Feed button at the top
            container.innerHTML = `
                <div class="mb-6 text-center">
                    <button
                        onclick="categoryManager.showAddFeedModal(${categoryId})"
                        class="inline-flex items-center px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                    >
                        <i class="fas fa-plus mr-2"></i>
                        Add Feed to Category
                    </button>
                </div>
                <div class="space-y-4">
                    ${feeds.map(feed => `
                        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center space-x-3">
                                    <div class="w-3 h-3 bg-blue-500 rounded-full"></div>
                                    <div>
                                        <h3 class="font-medium text-gray-900 dark:text-gray-100">${feed.name}</h3>
                                        <p class="text-sm text-gray-500 dark:text-gray-400">${feed.url}</p>
                                    </div>
                                </div>
                                <div class="flex items-center space-x-2">
                                    <button
                                        onclick="categoryManager.removeFeedFromCategory(${categoryId}, ${feed.id})"
                                        class="text-red-500 hover:text-red-600 transition-colors"
                                        title="Remove from category"
                                    >
                                        <i class="fas fa-times"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            `;
        }
    }

    showAddCategoryModal() {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Create New Category</h3>
                <form id="addCategoryForm">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Category Name
                        </label>
                        <input
                            type="text"
                            id="categoryName"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="e.g., Technology, News, Hobbies"
                            required
                        >
                    </div>
                    <div class="mb-6">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Description (optional)
                        </label>
                        <textarea
                            id="categoryDescription"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="Describe what this category is for..."
                            rows="3"
                        ></textarea>
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            onclick="this.closest('.fixed').remove()"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            class="px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                        >
                            Create Category
                        </button>
                    </div>
                </form>
            </div>
        `;

        document.body.appendChild(modal);

        document.getElementById('addCategoryForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.createCategory();
        });
    }

    async createCategory() {
        const name = document.getElementById('categoryName').value.trim();
        const description = document.getElementById('categoryDescription').value.trim();

        if (!name) return;

        try {
            const response = await fetch('/api/categories', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, description })
            });

            if (response.ok) {
                document.querySelector('.fixed').remove();
                await this.loadCategories();
            } else {
                alert('Failed to create category');
            }
        } catch (error) {
            console.error('Error creating category:', error);
            alert('Error creating category');
        }
    }

    showAddFeedModal(categoryId) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Add Feed to Category</h3>
                <form id="addFeedForm">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Feed Type
                        </label>
                        <select
                            id="feedType"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                        >
                            <option value="reddit">Reddit Subreddit</option>
                            <option value="rss">RSS Feed</option>
                        </select>
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Feed URL
                        </label>
                        <input
                            type="url"
                            id="feedUrl"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="https://www.reddit.com/r/programming/ or RSS URL"
                            required
                        >
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Display Name
                        </label>
                        <input
                            type="text"
                            id="feedName"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="Programming, TechCrunch, etc."
                            required
                        >
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            onclick="this.closest('.fixed').remove()"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            class="px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                        >
                            Add Feed
                        </button>
                    </div>
                </form>
            </div>
        `;

        document.body.appendChild(modal);

        document.getElementById('addFeedForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.addFeedToCategory(categoryId);
        });
    }

    async addFeedToCategory(categoryId) {
        const feedType = document.getElementById('feedType').value;
        const feedUrl = document.getElementById('feedUrl').value.trim();
        const feedName = document.getElementById('feedName').value.trim();

        if (!feedUrl || !feedName) return;

        try {
            const response = await fetch(`/api/categories/${categoryId}/feeds`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    type: feedType,
                    url: feedUrl,
                    name: feedName
                })
            });

            if (response.ok) {
                document.querySelector('.fixed').remove();
                this.loadCategoryFeeds(categoryId);
            } else {
                alert('Failed to add feed to category');
            }
        } catch (error) {
            console.error('Error adding feed to category:', error);
            alert('Error adding feed to category');
        }
    }

    async removeFeedFromCategory(categoryId, feedSourceId) {
        if (!confirm('Remove this feed from the category?')) return;

        try {
            const response = await fetch(`/api/categories/${categoryId}/feeds/${feedSourceId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.loadCategoryFeeds(categoryId);
            } else {
                alert('Failed to remove feed from category');
            }
        } catch (error) {
            console.error('Error removing feed from category:', error);
            alert('Error removing feed from category');
        }
    }

    showCategoryMenu(event, categoryId) {
        event.stopPropagation();

        // Remove any existing menus
        document.querySelectorAll('.category-menu').forEach(menu => menu.remove());

        const menu = document.createElement('div');
        menu.className = 'category-menu absolute bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 mt-1';
        menu.innerHTML = `
            <div class="py-1">
                <button onclick="categoryManager.editCategory(${categoryId})" class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                    <i class="fas fa-edit mr-2"></i>Edit
                </button>
                <button onclick="categoryManager.deleteCategory(${categoryId})" class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900">
                    <i class="fas fa-trash mr-2"></i>Delete
                </button>
            </div>
        `;

        event.target.closest('.category-btn').appendChild(menu);

        // Close menu when clicking outside
        setTimeout(() => {
            document.addEventListener('click', function closeMenu() {
                menu.remove();
                document.removeEventListener('click', closeMenu);
            });
        }, 1);
    }

    async editCategory(categoryId) {
        // Find the category in our local array
        const category = this.categories.find(cat => cat.id === categoryId);
        if (!category) {
            alert('Category not found');
            return;
        }

        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">Edit Category</h3>
                <form id="editCategoryForm">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Category Name
                        </label>
                        <input
                            type="text"
                            id="editCategoryName"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="e.g., Technology, News, Hobbies"
                            value="${category.name}"
                            required
                        >
                    </div>
                    <div class="mb-6">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                            Description (optional)
                        </label>
                        <textarea
                            id="editCategoryDescription"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                            placeholder="Describe what this category is for..."
                            rows="3"
                        >${category.description || ''}</textarea>
                    </div>
                    <div class="flex justify-end space-x-3">
                        <button
                            type="button"
                            onclick="this.closest('.fixed').remove()"
                            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            class="px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                        >
                            Update Category
                        </button>
                    </div>
                </form>
            </div>
        `;

        document.body.appendChild(modal);

        document.getElementById('editCategoryForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            await this.updateCategory(categoryId);
        });
    }

    async updateCategory(categoryId) {
        const name = document.getElementById('editCategoryName').value.trim();
        const description = document.getElementById('editCategoryDescription').value.trim();

        if (!name) return;

        try {
            const response = await fetch(`/api/categories/${categoryId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, description })
            });

            if (response.ok) {
                document.querySelector('.fixed').remove();
                await this.loadCategories();
            } else {
                alert('Failed to update category');
            }
        } catch (error) {
            console.error('Error updating category:', error);
            alert('Error updating category');
        }
    }

    async deleteCategory(categoryId) {
        if (!confirm('Delete this category? All feed associations will be removed.')) return;

        try {
            const response = await fetch(`/api/categories/${categoryId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.currentCategory = 'all';
                await this.loadCategories();
            } else {
                alert('Failed to delete category');
            }
        } catch (error) {
            console.error('Error deleting category:', error);
            alert('Error deleting category');
        }
    }

    setupEventListeners() {
        // Add category container if it doesn't exist
        if (!document.getElementById('categoryContainer')) {
            const container = document.querySelector('.flex.flex-wrap.gap-2');
            if (container) {
                container.id = 'categoryContainer';
            }
        }
    }
}

// Initialize category manager
const categoryManager = new CategoryManager();
