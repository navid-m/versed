document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.querySelector('input[type="text"][placeholder="Search content..."]');
    const postsContainer = document.querySelector('.space-y-3');
    const originalContent = postsContainer.innerHTML;
    let searchTimeout;

    function renderSearchResults(results) {
        if (results.length === 0) {
            postsContainer.innerHTML = `
                <div class="text-center py-12">
                    <div class="w-16 h-16 mx-auto mb-4 text-gray-400 dark:text-gray-600">
                        <i class="fas fa-search text-4xl"></i>
                    </div>
                    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                        No results found
                    </h3>
                    <p class="text-gray-500 dark:text-gray-400">
                        Try searching for something else.
                    </p>
                </div>
            `;
            return;
        }

        let html = '';
        results.forEach(item => {
            html += `
                <article class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md dark:hover:shadow-xl transition-shadow">
                    <div class="p-4">
                        <div class="flex items-start space-x-3">
                            <div class="flex flex-col items-center space-y-1 flex-shrink-0">
                                <button class="p-1 text-orange-500 hover:text-orange-600 transition-colors" data-feed-id="${item.id}" data-vote-type="upvote">
                                    <i class="fas fa-chevron-up text-sm"></i>
                                </button>
                                <span class="text-xs font-medium text-orange-500">${item.score || 0}</span>
                                <button class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors" data-feed-id="${item.id}" data-vote-type="downvote">
                                    <i class="fas fa-chevron-down text-sm"></i>
                                </button>
                            </div>

                            <div class="flex-1 min-w-0">
                                <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 hover:text-gray-700 dark:hover:text-gray-300 transition-colors line-clamp-2">
                                    <a href="${item.url}" target="_blank" class="hover:underline">
                                        ${item.title}
                                    </a>
                                </h2>

                                <div class="relative mb-4 modern-description">
                                    <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed line-height-6 font-medium tracking-wide line-clamp-3 bg-gradient-to-br from-gray-50/80 to-white/50 dark:from-gray-800/60 dark:to-gray-700/40 backdrop-blur-sm rounded-lg px-4 py-3 border-l-4 border-blue-500/30 dark:border-blue-400/40 shadow-sm">
                                        <p>${item.description || "No description available"}</p>
                                    </div>
                                    <div class="absolute inset-0 bg-gradient-to-r from-blue-50/20 to-indigo-50/20 dark:from-blue-900/10 dark:to-indigo-900/10 rounded-lg blur-xl transform scale-105 opacity-60"></div>
                                </div>

                                <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-3">
                                    <span class="flex items-center">
                                        <i class="far fa-user mr-1"></i>
                                        ${item.author || "Unknown author"}
                                    </span>
                                    <span class="flex items-center">
                                        <i class="far fa-clock mr-1"></i>
                                        ${new Date(item.published_at).toLocaleDateString()}
                                    </span>
                                    <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
                                        ${item.source_name || "Unknown source"}
                                    </span>
                                </div>
                            </div>
                            ${document.querySelector('[data-email]') ? `
                            <div class="flex flex-col items-center space-y-1 flex-shrink-0 ml-3">
                                <button class="p-1 text-gray-400 hover:text-blue-500 transition-colors save-button" data-feed-id="${item.id}" data-action="save" title="Save to reading list">
                                    <i class="far fa-bookmark text-sm"></i>
                                </button>
                            </div>
                            ` : ''}
                        </div>
                    </div>
                </article>
            `;
        });

        postsContainer.innerHTML = html;

        attachEventListeners();
    }

    function debounce(func) {
        let timeoutId;
        return function(...args) {
            const context = this;
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => {
                func.apply(context, args);
            }, 1000);
        };
    }

    function attachEventListeners() {
        const voteButtons = document.querySelectorAll('[data-vote-type]');
        voteButtons.forEach(button => {
            button.addEventListener('click', async function() {
                const feedId = this.getAttribute('data-feed-id');
                const voteType = this.getAttribute('data-vote-type');

                try {
                    const response = await fetch('/api/vote', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            feed_id: feedId,
                            vote_type: voteType,
                        }),
                    });

                    if (!response.ok) {
                        throw new Error('Failed to submit vote');
                    }

                    const data = await response.json();
                    const scoreElement = this.parentElement.querySelector('.text-xs.font-medium.text-orange-500');
                    if (scoreElement) {
                        scoreElement.textContent = data.new_score;
                    }
                } catch (error) {
                    console.error('Error submitting vote:', error);
                }
            });
        });

        if (document.querySelector('[data-email]')) {
            const saveButtons = document.querySelectorAll('.save-button:not([data-listener-attached])');
            saveButtons.forEach(button => {
                button.setAttribute("data-listener-attached", "true");
                const feedId = button.getAttribute('data-feed-id');
                checkSaveStatus(button, feedId);
                button.addEventListener('click', debounce(async function() {
                    const action = this.getAttribute('data-action');
                    this.disabled = true;
                    const icon = this.querySelector('i');
                    const originalIconClass = icon.className;
                    icon.className = 'fas fa-spinner fa-spin text-sm';
                    
                    try {
                        let response;
                        if (action === 'save') {
                            response = await fetch('/api/reading-list/save', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json',
                                },
                                body: JSON.stringify({
                                    item_id: feedId,
                                }),
                            });
                        } else {
                            response = await fetch('/api/reading-list/remove', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json',
                                },
                                body: JSON.stringify({
                                    item_id: feedId,
                                }),
                            });
                        }

                        if (!response.ok) {
                            throw new Error('Failed to update reading list');
                        }
                        toggleSaveButton(this);
                    } catch (error) {
                        console.error('Error updating reading list:', error);
                        icon.className = originalIconClass;
                    } finally {
                        this.disabled = false;
                    }
                }))
            });
        }
    }
    async function checkSaveStatus(button, feedId) {
        try {
            const response = await fetch(`/api/reading-list/check/${feedId}`);
            if (response.ok) {
                const data = await response.json();
                setSaveButtonState(button, data.saved);
            }
        } catch (error) {
            console.error('Error checking save status:', error);
        }
    }
    
    function setSaveButtonState(button, isSaved) {
        const icon = button.querySelector('i');

        if (isSaved) {
            button.setAttribute('data-action', 'unsave');
            icon.className = 'fas fa-bookmark text-sm';
            button.className = 'p-1 text-blue-500 hover:text-red-500 transition-colors save-button';
            button.setAttribute('title', 'Remove from reading list');
        } else {
            button.setAttribute('data-action', 'save');
            icon.className = 'far fa-bookmark text-sm';
            button.className = 'p-1 text-gray-400 hover:text-blue-500 transition-colors save-button';
            button.setAttribute('title', 'Save to reading list');
        }
    }
    function toggleSaveButton(button) {
        const icon = button.querySelector('i');
        const currentAction = button.getAttribute('data-action');

        if (currentAction === 'save') {
            button.setAttribute('data-action', 'unsave');
            icon.className = 'fas fa-bookmark text-sm';
            button.className = 'p-1 text-blue-500 hover:text-red-500 transition-colors save-button';
            button.setAttribute('title', 'Remove from reading list');
        } else {
            button.setAttribute('data-action', 'save');
            icon.className = 'far fa-bookmark text-sm';
            button.className = 'p-1 text-gray-400 hover:text-blue-500 transition-colors save-button';
            button.setAttribute('title', 'Save to reading list');
        }
    }

    searchInput.addEventListener('input', function() {
        const query = this.value.trim();

        clearTimeout(searchTimeout);

        if (query === '') {
            postsContainer.innerHTML = originalContent;
            attachEventListeners();
            return;
        }

        searchTimeout = setTimeout(async () => {
            try {
                const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
                if (!response.ok) {
                    throw new Error('Search failed');
                }

                const data = await response.json();
                renderSearchResults(data.items);
            } catch (error) {
                console.error('Search error:', error);
                postsContainer.innerHTML = `
                    <div class="text-center py-12">
                        <div class="w-16 h-16 mx-auto mb-4 text-red-400">
                            <i class="fas fa-exclamation-triangle text-4xl"></i>
                        </div>
                        <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                            Search error
                        </h3>
                        <p class="text-gray-500 dark:text-gray-400">
                            Please try again later.
                        </p>
                    </div>
                `;
            }
        }, 300);
    });

    searchInput.addEventListener('search', function() {
        if (this.value === '') {
            postsContainer.innerHTML = originalContent;
            attachEventListeners();
        }
    });

    attachEventListeners();
});
