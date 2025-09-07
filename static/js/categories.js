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
            console.log('Loading categories from API...');
            const response = await fetch('/api/categories');
            console.log('Categories API response status:', response.status);

            if (!response.ok) {
                console.error('Failed to load categories:', response.status);
                return;
            }

            const data = await response.json();
            console.log('Categories API response data:', data);
            console.log('Categories array:', data.categories);

            if (data.categories && Array.isArray(data.categories)) {
                this.categories = data.categories;
                console.log('Loaded', this.categories.length, 'categories');
                this.categories.forEach((cat, index) => {
                    console.log(`Category ${index}: id=${cat.id} (type: ${typeof cat.id}), name="${cat.name}"`);
                });
                this.renderCategories();
            } else {
                console.error('Invalid categories data structure:', data);
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
        console.log('=== selectCategory called ===');
        console.log('categoryId:', categoryId, '(type:', typeof categoryId, ')');

        // For 'all' category, navigate directly
        if (categoryId === 'all') {
            console.log('Navigating to all category');
            window.location.href = '/';
            return;
        }

        // Check if categories are loaded
        console.log('Current categories array:', this.categories);
        console.log('Categories length:', this.categories ? this.categories.length : 'undefined');

        if (!this.categories || this.categories.length === 0) {
            console.log('Categories not loaded yet, loading first...');
            this.loadCategories().then(() => {
                console.log('Categories loaded, retrying selectCategory');
                this.selectCategory(categoryId); // Retry after loading
            }).catch(error => {
                console.error('Failed to load categories:', error);
            });
            return;
        }

        // Find the category to get its name
        console.log('Looking for category with id:', categoryId);
        const categoryIdNum = parseInt(categoryId, 10);
        console.log('Converted categoryId to number:', categoryIdNum);

        const category = this.categories.find(cat => {
            console.log(`Comparing cat.id=${cat.id} (type: ${typeof cat.id}) with categoryIdNum=${categoryIdNum} (type: ${typeof categoryIdNum})`);
            return cat.id === categoryIdNum;
        });

        console.log('Found category:', category);

        if (!category) {
            console.error('Category not found after loading:', categoryId);
            console.log('Available categories:');
            this.categories.forEach((cat, index) => {
                console.log(`  ${index}: id=${cat.id}, name="${cat.name}"`);
            });
            return;
        }

        // Navigate to the URL route
        const username = this.getUsername();
        console.log('Username for navigation:', username);

        if (username) {
            const categorySlug = category.name.toLowerCase().replace(/\s+/g, '-');
            const url = `/u/${username}/c/${categorySlug}`;
            console.log('Navigating to:', url);
            window.location.href = url;
        } else {
            console.error('Username not found for navigation');
        }
    }

    getUsername() {
        // Try to get username from various sources
        const usernameElement = document.querySelector('[data-username]');
        if (usernameElement) {
            return usernameElement.getAttribute('data-username');
        }

        // Try to get from localStorage if stored
        const storedUsername = localStorage.getItem('verse_username');
        if (storedUsername) {
            return storedUsername;
        }

        // Extract from current URL if we're already on a user route
        const urlMatch = window.location.pathname.match(/^\/u\/([^\/]+)\//);
        if (urlMatch) {
            return urlMatch[1];
        }

        return null;
    }

    async loadCategoryFeeds(categoryId) {
        console.log('Loading feeds for category:', categoryId);
        try {
            // First check if the category has any feeds associated with it
            let feedsResponse = await fetch(`/api/categories/${categoryId}/feeds`);
            if (!feedsResponse.ok) {
                console.error('Failed to get category feeds');
                this.renderCategoryFeedItems([], categoryId, 0);
                return;
            }

            let feedsData = await feedsResponse.json();
            let feedsCount = feedsData.feeds ? feedsData.feeds.length : 0;
            console.log('Category has', feedsCount, 'feeds');

            // Now get the feed items
            let url;
            if (categoryId === 'all') {
                url = '/api/feeds';
                console.log('Loading all feeds from:', url);
            } else {
                url = `/api/categories/${categoryId}/items`;
                console.log('Loading category feed items from:', url);
            }

            const response = await fetch(url);
            console.log('API response status:', response.status);
            if (response.ok) {
                const data = await response.json();
                console.log('API response data:', data);
                this.renderCategoryFeedItems(data.items || [], categoryId, feedsCount);
            } else {
                console.error('Failed to load category feeds:', response.status);
                this.renderCategoryFeedItems([], categoryId, feedsCount);
            }
        } catch (error) {
            console.error('Error loading category feeds:', error);
            this.renderCategoryFeedItems([], categoryId, 0);
        }
    }

    renderCategoryFeedItems(feedItems, categoryId, feedsCount) {
        console.log('Rendering feed items for category:', categoryId, 'items count:', feedItems.length, 'feeds count:', feedsCount);
        const container = document.getElementById('postsContainer');
        if (!container) {
            console.error('postsContainer not found!');
            return;
        }

        if (feedItems.length === 0) {
            if (categoryId === 'all') {
                console.log('Category is "all", not rendering empty state');
                return;
            } else if (feedsCount > 0) {
                // Category has feeds but no feed items yet
                console.log('Category has feeds but no feed items yet');
                container.innerHTML = `
                    <div class="text-center py-12">
                        <div class="w-16 h-16 mx-auto mb-4 text-blue-400 dark:text-blue-600">
                            <i class="fas fa-spinner fa-spin text-4xl"></i>
                        </div>
                        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                            Loading feed items...
                        </h3>
                        <p class="text-gray-500 dark:text-gray-400 mb-4">
                            This category has ${feedsCount} feed(s) but no items have been loaded yet.
                        </p>
                        <p class="text-sm text-gray-400 dark:text-gray-500">
                            Items will appear once the feeds are processed.
                        </p>
                        <div class="mt-6">
                            <button
                                onclick="categoryManager.showAddFeedModal(${categoryId})"
                                class="inline-flex items-center px-4 py-2 bg-gray-900 dark:bg-gray-700 text-white rounded-lg hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors"
                            >
                                <i class="fas fa-plus mr-2"></i>
                                Manage Feeds
                            </button>
                        </div>
                    </div>
                `;
            } else {
                // Category has no feeds at all
                console.log('Category has no feeds, showing add feed button');
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
            console.log('Category has feed items, rendering them');
            // Render feed items in the same format as the main feed
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
                <div class="space-y-3">
                    ${feedItems.map(item => `
                        <article class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md dark:hover:shadow-xl transition-shadow">
                            <div class="p-4">
                                <div class="flex items-start space-x-3">
                                    <div class="flex flex-col items-center space-y-1 flex-shrink-0">
                                        <button class="p-1 text-orange-500 hover:text-orange-600 transition-colors" data-feed-id="${item.ID}" data-vote-type="upvote">
                                            <i class="fas fa-chevron-up text-sm"></i>
                                        </button>
                                        <span class="text-xs font-medium text-orange-500">${item.Score}</span>
                                        <button class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors" data-feed-id="${item.ID}" data-vote-type="downvote">
                                            <i class="fas fa-chevron-down text-sm"></i>
                                        </button>
                                    </div>

                                    <div class="flex-1 min-w-0">
                                        <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 hover:text-gray-700 dark:hover:text-gray-300 transition-colors line-clamp-2">
                                            <a href="${item.URL}" target="_blank" class="hover:underline">
                                                ${item.Title}
                                            </a>
                                        </h2>

                                        <div class="relative mb-4 modern-description">
                                            <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed line-height-6 font-medium tracking-wide line-clamp-3 bg-gradient-to-br from-gray-50/80 to-white/50 dark:from-gray-800/60 dark:to-gray-700/40 backdrop-blur-sm rounded-lg px-4 py-3 border-l-4 border-blue-500/30 dark:border-blue-400/40 shadow-sm">
                                                <p>${item.Description || 'No description available.'}</p>
                                            </div>
                                            <div class="absolute inset-0 bg-gradient-to-r from-blue-50/20 to-indigo-50/20 dark:from-blue-900/10 dark:to-indigo-900/10 rounded-lg blur-xl transform scale-105 opacity-60"></div>
                                        </div>

                                        <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-3">
                                            <span class="flex items-center">
                                                <i class="far fa-user mr-1"></i>
                                                ${item.Author || 'Unknown'}
                                            </span>
                                            <span class="flex items-center">
                                                <i class="far fa-clock mr-1"></i>
                                                ${new Date(item.PublishedAt).toLocaleDateString()}
                                            </span>
                                            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
                                                ${item.SourceName || 'Unknown Source'}
                                            </span>
                                        </div>
                                    </div>

                                    <div class="flex flex-col items-center space-y-1 flex-shrink-0 ml-3">
                                        <button class="p-1 text-gray-400 hover:text-blue-500 transition-colors save-button" data-feed-id="${item.ID}" data-action="save" title="Save to reading list">
                                            <i class="far fa-bookmark text-sm"></i>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </article>
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
            const response = await fetch(`/api/categories/${categoryId}/feeds/create`, {
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
                <button onclick="categoryManager.showAddFeedModal(${categoryId})" class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                    <i class="fas fa-plus mr-2"></i>Amend items
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

// Only initialize if we're not on a category URL route
if (!window.location.pathname.match(/^\/u\/[^\/]+\/c\/[^\/]+$/)) {
    console.log('Initializing category manager for dynamic loading');
    categoryManager.init();
} else {
    console.log('On category URL route - skipping dynamic initialization');
}
