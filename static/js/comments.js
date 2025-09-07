// Comments functionality
class CommentsManager {
    constructor() {
        this.init();
    }

    init() {
        // Handle "View comments" button clicks
        document.addEventListener('click', (e) => {
            if (e.target.closest('.view-comments-btn')) {
                e.preventDefault();
                const button = e.target.closest('.view-comments-btn');
                const postId = button.dataset.postId;
                this.viewPostComments(postId);
            }
        });

        // Handle comment form submission
        const submitButton = document.getElementById('submitComment');
        if (submitButton) {
            submitButton.addEventListener('click', () => this.submitComment());
        }

        // Handle edit comment buttons
        document.addEventListener('click', (e) => {
            if (e.target.closest('.edit-comment-btn')) {
                e.preventDefault();
                const button = e.target.closest('.edit-comment-btn');
                const commentId = button.dataset.commentId;
                this.editComment(commentId);
            }
        });

        // Handle delete comment buttons
        document.addEventListener('click', (e) => {
            if (e.target.closest('.delete-comment-btn')) {
                e.preventDefault();
                const button = e.target.closest('.delete-comment-btn');
                const commentId = button.dataset.commentId;
                this.deleteComment(commentId);
            }
        });

        // Handle escape key to cancel editing
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.cancelEdit();
            }
        });
    }

    viewPostComments(postId) {
        // Navigate to the post view page
        window.location.href = `/post/${postId}`;
    }

    async submitComment() {
        const content = document.getElementById('commentContent').value.trim();
        const submitButton = document.getElementById('submitComment');
        const postId = submitButton.dataset.postId;

        if (!content) {
            this.showMessage('Comment cannot be empty', 'error');
            return;
        }

        submitButton.disabled = true;
        submitButton.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>Posting...';

        try {
            const response = await fetch(`/api/posts/${postId}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ content }),
            });

            if (response.ok) {
                const newComment = await response.json();
                this.addCommentToUI(newComment);
                document.getElementById('commentContent').value = '';
                this.showMessage('Comment posted successfully', 'success');
            } else {
                const error = await response.json();
                this.showMessage(error.error || 'Failed to post comment', 'error');
            }
        } catch (error) {
            console.error('Error posting comment:', error);
            this.showMessage('Failed to post comment', 'error');
        } finally {
            submitButton.disabled = false;
            submitButton.innerHTML = '<i class="fas fa-paper-plane mr-2"></i>Post Comment';
        }
    }

    async editComment(commentId) {
        const commentElement = document.querySelector(`[data-comment-id="${commentId}"]`);
        const contentElement = commentElement.querySelector('.text-gray-700.dark\\:text-gray-300');

        if (commentElement.classList.contains('editing')) {
            return; // Already editing
        }

        commentElement.classList.add('editing');
        const originalContent = contentElement.textContent.trim();

        // Replace content with textarea
        contentElement.innerHTML = `
            <textarea class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 resize-vertical text-sm" rows="3">${originalContent}</textarea>
            <div class="flex justify-end space-x-2 mt-2">
                <button class="px-2 py-1 text-xs bg-gray-500 text-white rounded hover:bg-gray-600 save-edit-btn" data-comment-id="${commentId}">Save</button>
                <button class="px-2 py-1 text-xs bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-400 dark:hover:bg-gray-500 cancel-edit-btn">Cancel</button>
            </div>
        `;

        // Focus on textarea
        const textarea = contentElement.querySelector('textarea');
        textarea.focus();
        textarea.setSelectionRange(textarea.value.length, textarea.value.length);

        // Handle save
        contentElement.querySelector('.save-edit-btn').addEventListener('click', () => {
            this.saveCommentEdit(commentId);
        });

        // Handle cancel
        contentElement.querySelector('.cancel-edit-btn').addEventListener('click', () => {
            this.cancelEdit();
        });
    }

    async saveCommentEdit(commentId) {
        const commentElement = document.querySelector(`[data-comment-id="${commentId}"]`);
        const textarea = commentElement.querySelector('textarea');
        const newContent = textarea.value.trim();

        if (!newContent) {
            this.showMessage('Comment cannot be empty', 'error');
            return;
        }

        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ content: newContent }),
            });

            if (response.ok) {
                const updatedComment = await response.json();
                this.updateCommentInUI(updatedComment);
                this.showMessage('Comment updated successfully', 'success');
            } else {
                const error = await response.json();
                this.showMessage(error.error || 'Failed to update comment', 'error');
            }
        } catch (error) {
            console.error('Error updating comment:', error);
            this.showMessage('Failed to update comment', 'error');
        }
    }

    cancelEdit() {
        const editingElement = document.querySelector('.editing');
        if (editingElement) {
            const commentId = editingElement.dataset.commentId;
            this.refreshComment(commentId);
        }
    }

    async deleteComment(commentId) {
        if (!confirm('Are you sure you want to delete this comment?')) {
            return;
        }

        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'DELETE',
            });

            if (response.ok) {
                this.removeCommentFromUI(commentId);
                this.showMessage('Comment deleted successfully', 'success');
            } else {
                const error = await response.json();
                this.showMessage(error.error || 'Failed to delete comment', 'error');
            }
        } catch (error) {
            console.error('Error deleting comment:', error);
            this.showMessage('Failed to delete comment', 'error');
        }
    }

    addCommentToUI(comment) {
        const commentsList = document.getElementById('commentsList');
        if (!commentsList) return;

        // Remove "no comments" message if present
        const noCommentsElement = commentsList.querySelector('.text-center.py-8');
        if (noCommentsElement) {
            noCommentsElement.remove();
        }

        const commentHTML = this.createCommentHTML(comment);
        commentsList.insertAdjacentHTML('afterbegin', commentHTML);

        // Update comment count in header
        this.updateCommentCount();
    }

    updateCommentInUI(comment) {
        const commentElement = document.querySelector(`[data-comment-id="${comment.ID}"]`);
        if (!commentElement) return;

        commentElement.classList.remove('editing');
        const contentElement = commentElement.querySelector('.text-gray-700.dark\\:text-gray-300');
        contentElement.innerHTML = comment.Content;
    }

    removeCommentFromUI(commentId) {
        const commentElement = document.querySelector(`[data-comment-id="${commentId}"]`);
        if (commentElement) {
            commentElement.remove();
            this.updateCommentCount();
        }

        // Add "no comments" message if no comments left
        const commentsList = document.getElementById('commentsList');
        if (commentsList && commentsList.children.length === 0) {
            commentsList.innerHTML = `
                <div class="text-center py-8 text-gray-500 dark:text-gray-400">
                    <i class="fas fa-comments text-2xl mb-2 block"></i>
                    <p>No comments yet. Be the first to share your thoughts!</p>
                </div>
            `;
        }
    }

    async refreshComment(commentId) {
        try {
            const response = await fetch(`/api/comments/${commentId}`);
            if (response.ok) {
                const comment = await response.json();
                this.updateCommentInUI(comment);
            }
        } catch (error) {
            console.error('Error refreshing comment:', error);
        }
    }

    createCommentHTML(comment) {
        const currentUserId = parseInt(document.body.dataset.userId) || 0;
        const isOwner = currentUserId === comment.UserID;

        return `
            <div class="border-l-2 border-gray-200 dark:border-gray-600 pl-4" data-comment-id="${comment.ID}">
                <div class="flex items-start space-x-3">
                    <div class="flex-shrink-0">
                        <div class="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                            <i class="fas fa-user text-gray-600 dark:text-gray-400 text-xs"></i>
                        </div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center space-x-2 mb-1">
                            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">${comment.Username}</span>
                            <span class="text-xs text-gray-500 dark:text-gray-400">${new Date(comment.CreatedAt).toLocaleDateString('en-US', {
                                year: 'numeric',
                                month: 'short',
                                day: 'numeric',
                                hour: 'numeric',
                                minute: '2-digit',
                                hour12: true
                            })}</span>
                        </div>
                        <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                            ${comment.Content}
                        </div>
                        ${isOwner ? `
                        <div class="flex items-center space-x-2 mt-2">
                            <button class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 edit-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-edit mr-1"></i>Edit
                            </button>
                            <button class="text-xs text-red-500 hover:text-red-700 delete-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-trash mr-1"></i>Delete
                            </button>
                        </div>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    }

    updateCommentCount() {
        const commentsList = document.getElementById('commentsList');
        if (!commentsList) return;

        const commentCount = commentsList.querySelectorAll('[data-comment-id]').length;
        const header = document.querySelector('h2');
        if (header) {
            header.textContent = `Comments (${commentCount})`;
        }
    }

    showMessage(message, type = 'info') {
        // Simple notification system - you might want to enhance this
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 px-4 py-2 rounded-md text-white z-50 ${
            type === 'success' ? 'bg-green-500' :
            type === 'error' ? 'bg-red-500' :
            'bg-blue-500'
        }`;
        notification.textContent = message;

        document.body.appendChild(notification);

        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// Initialize comments manager when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new CommentsManager();
});
