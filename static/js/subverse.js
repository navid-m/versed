class SubverseManager {
   constructor(subverseName) {
      this.subverseName = subverseName;
      this.init();
   }

   init() {
      this.setupEventListeners();
      this.loadPosts();
   }

   setupEventListeners() {
      const createPostBtn = document.querySelector(
         '[onclick="showCreatePostModal()"]'
      );
      if (createPostBtn) {
         createPostBtn.addEventListener("click", () =>
            this.showCreatePostModal()
         );
      }

      document.addEventListener("click", (e) => {
         if (
            e.target.classList.contains("modal-overlay") ||
            e.target.classList.contains("close-modal")
         ) {
            this.hideCreatePostModal();
         }
      });

      document.addEventListener("change", (e) => {
         if (e.target.name === "postType") {
            this.togglePostFields(e.target.value);
         }
      });

      const createPostForm = document.getElementById("createPostForm");
      if (createPostForm) {
         createPostForm.addEventListener("submit", (e) =>
            this.handleCreatePost(e)
         );
      }

      document.addEventListener("click", (e) => {
         if (e.target.closest("[data-vote-type]")) {
            const button = e.target.closest("[data-vote-type]");
            const postId = button.dataset.postId;
            const voteType = button.dataset.voteType;
            this.handleVote(postId, voteType, button);
         }
      });
   }

   async loadPosts() {
      try {
         const response = await fetch(`/s/${this.subverseName}/posts`);
         if (response.ok) {
            const data = await response.json();
            this.renderPosts(data.Posts || []);
         } else {
            console.error("Failed to load posts");
            this.renderPosts([]);
         }
      } catch (error) {
         console.error("Error loading posts:", error);
         this.renderPosts([]);
      }
   }

   renderPosts(posts) {
      const container = document.getElementById("postsContainer");
      if (!container) return;

      if (posts.length === 0) {
         container.innerHTML = `
            <div class="text-center py-12">
               <div class="w-16 h-16 mx-auto mb-4 text-gray-400 dark:text-gray-600">
                  <i class="fas fa-folder-open text-4xl"></i>
               </div>
               <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                  No posts in /s/${this.subverseName}
               </h3>
               <p class="text-gray-500 dark:text-gray-400">
                  Be the first to create a post!
               </p>
            </div>
         `;
         return;
      }

      container.innerHTML = posts.map((post) => this.renderPost(post)).join("");
   }

   renderPost(post) {
      const createdAt = new Date(post.created_at).toLocaleDateString("en-US", {
         year: "numeric",
         month: "short",
         day: "numeric",
         hour: "numeric",
         minute: "2-digit",
      });

      return `
         <article class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md dark:hover:shadow-xl transition-shadow">
            <div class="p-4">
               <div class="flex items-start space-x-3">
                  <div class="flex flex-col items-center space-y-1 flex-shrink-0">
                     <button class="p-1 text-orange-500 hover:text-orange-600 transition-colors vote-btn"
                        data-post-id="${post.id}" data-vote-type="upvote">
                        <i class="fas fa-chevron-up text-sm"></i>
                     </button>
                     <span class="text-xs font-medium text-orange-500 score" data-post-id="${
                        post.id
                     }">
                        ${post.score || 0}
                     </span>
                     <button class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors vote-btn"
                        data-post-id="${post.id}" data-vote-type="downvote">
                        <i class="fas fa-chevron-down text-sm"></i>
                     </button>
                  </div>

                  <div class="flex-1 min-w-0">
                     <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 hover:text-gray-700 dark:hover:text-gray-300 transition-colors line-clamp-2">
                        <a href="/posts/${post.id}" class="hover:underline">
                           ${post.title}
                        </a>
                     </h2>

                     ${
                        post.post_type === "link"
                           ? `
                     <div class="mb-4">
                        <a href="${post.url}" target="_blank" class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline">
                           ${post.url}
                        </a>
                     </div>
                     `
                           : ""
                     }

                     ${
                        post.content
                           ? `
                     <div class="relative mb-4 modern-description">
                        <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed line-height-6 font-medium tracking-wide line-clamp-3 bg-gradient-to-br from-gray-50/80 to-white/50 dark:from-gray-800/60 dark:to-gray-700/40 backdrop-blur-sm rounded-lg px-4 py-3 border-l-4 border-red-500/30 dark:border-red-400/40 shadow-sm">
                           <p>${this.escapeHtml(post.content)}</p>
                        </div>
                        <div class="absolute inset-0 bg-gradient-to-r from-red-50/20 to-pink-50/20 dark:from-red-900/10 dark:to-gray-900/10 rounded-lg blur-xl transform scale-105 opacity-60"></div>
                     </div>
                     `
                           : ""
                     }

                     <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-3">
                        <span class="flex items-center">
                           <i class="far fa-user mr-1"></i>
                           ${post.username}
                        </span>
                        <span class="flex items-center">
                           <i class="far fa-clock mr-1"></i>
                           ${createdAt}
                        </span>
                        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200">
                           ${post.post_type}
                        </span>
                     </div>
                  </div>
               </div>
            </div>
         </article>
      `;
   }

   showCreatePostModal() {
      const modal = document.getElementById("createPostModal");
      if (modal) {
         modal.classList.remove("hidden");
      } else {
         this.createPostModal();
      }
      const createPostForm = document.getElementById("createPostForm");
      if (createPostForm) {
         createPostForm.addEventListener("submit", (e) =>
            this.handleCreatePost(e)
         );
      }
   }

   hideCreatePostModal() {
      const modal = document.getElementById("createPostModal");
      if (modal) {
         modal.classList.add("hidden");
      }
   }

   createPostModal() {
      const modal = document.createElement("div");
      modal.id = "createPostModal";
      modal.className =
         "fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 modal-overlay";
      modal.innerHTML = `
         <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-2xl mx-4">
            <div class="flex justify-between items-center mb-4">
               <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100">Create Post</h2>
               <button class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 close-modal">
                  <i class="fas fa-times"></i>
               </button>
            </div>

            <form id="createPostForm">
               <div class="mb-4">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Title</label>
                  <input type="text" name="title" required
                     class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-red-500">
               </div>

               <div class="mb-4">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Post Type</label>
                  <div class="flex space-x-4">
                     <label class="flex items-center">
                        <input type="radio" name="postType" value="text" checked class="mr-2">
                        <span class="text-gray-700 dark:text-gray-300">Text</span>
                     </label>
                     <label class="flex items-center">
                        <input type="radio" name="postType" value="link" class="mr-2">
                        <span class="text-gray-700 dark:text-gray-300">Link</span>
                     </label>
                  </div>
               </div>

               <div id="contentField" class="mb-4">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Content</label>
                  <textarea name="content" rows="6"
                     class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-red-500"></textarea>
               </div>

               <div id="urlField" class="mb-4 hidden">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">URL</label>
                  <input type="url" name="url"
                     class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-red-500"
                     placeholder="https://example.com">
               </div>

               <div class="flex justify-end space-x-3">
                  <button type="button" class="px-4 py-2 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 close-modal">
                     Cancel
                  </button>
                  <button type="submit" class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50" id="submitPostBtn">
                     <i class="fas fa-paper-plane mr-2"></i>
                     Post
                  </button>
               </div>
            </form>
         </div>
      `;

      document.body.appendChild(modal);
   }

   togglePostFields(postType) {
      const contentField = document.getElementById("contentField");
      const urlField = document.getElementById("urlField");

      if (postType === "text") {
         contentField.classList.remove("hidden");
         urlField.classList.add("hidden");
      } else {
         contentField.classList.add("hidden");
         urlField.classList.remove("hidden");
      }
   }

   async handleCreatePost(e) {
      e.preventDefault();

      const formData = new FormData(e.target);
      const postData = {
         title: formData.get("title").trim(),
         post_type: formData.get("postType"),
         content: formData.get("content")?.trim() || "",
         url: formData.get("url")?.trim() || "",
      };

      console.log("Creating post with data:", postData);

      if (!postData.title) {
         this.showMessage("Title is required", "error");
         return;
      }

      if (postData.post_type === "text" && !postData.content) {
         this.showMessage("Content is required for text posts", "error");
         return;
      }

      if (postData.post_type === "link" && !postData.url) {
         this.showMessage("URL is required for link posts", "error");
         return;
      }

      const submitBtn = document.getElementById("submitPostBtn");
      submitBtn.disabled = true;
      submitBtn.innerHTML =
         '<i class="fas fa-spinner fa-spin mr-2"></i>Posting...';

      try {
         const response = await fetch(`/s/${this.subverseName}/posts`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify(postData),
         });

         if (response.ok) {
            this.showMessage("Post created successfully!", "success");
            this.hideCreatePostModal();
            this.loadPosts();
            e.target.reset();
         } else {
            const error = await response.json();
            this.showMessage(error.error || "Failed to create post", "error");
         }
      } catch (error) {
         console.error("Error creating post:", error);
         this.showMessage("Failed to create post", "error");
      } finally {
         submitBtn.disabled = false;
         submitBtn.innerHTML = '<i class="fas fa-paper-plane mr-2"></i>Post';
      }
   }

   async handleVote(postId, voteType, button) {
      if (button.disabled) return;
      button.disabled = true;

      const scoreElement = document.querySelector(
         `[data-post-id="${postId}"].score`
      );

      if (!scoreElement) {
         button.disabled = false;
         return;
      }

      const originalScore = parseInt(scoreElement.textContent) || 0;

      try {
         const response = await fetch(`/api/posts/${postId}/vote`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ vote_type: voteType }),
         });

         if (response.ok) {
            const data = await response.json();
            scoreElement.textContent = data.score;
            this.updateVoteButtonStates(postId, voteType);
         } else if (response.status === 401) {
            this.showMessage("You must be logged in to vote", "error");
            scoreElement.textContent = originalScore;
         } else {
            const error = await response.json();
            this.showMessage(error.error || "Failed to vote", "error");
            scoreElement.textContent = originalScore;
         }
      } catch (error) {
         console.error("Error voting:", error);
         this.showMessage("Failed to vote. Try again.", "error");
         scoreElement.textContent = originalScore;
      } finally {
         button.disabled = false;
      }
   }

   updateVoteButtonStates(postId, voteType) {
      const upvoteBtn = document.querySelector(
         `[data-post-id="${postId}"][data-vote-type="upvote"]`
      );
      const downvoteBtn = document.querySelector(
         `[data-post-id="${postId}"][data-vote-type="downvote"]`
      );

      if (!upvoteBtn || !downvoteBtn) return;

      upvoteBtn.classList.remove("text-orange-600", "text-gray-400");
      upvoteBtn.classList.add("text-orange-500", "hover:text-orange-600");

      downvoteBtn.classList.remove("text-red-600", "text-gray-400");
      downvoteBtn.classList.add(
         "text-gray-400",
         "hover:text-gray-600",
         "dark:hover:text-gray-300"
      );

      if (voteType === "upvote") {
         upvoteBtn.classList.remove("text-orange-500", "hover:text-orange-600");
         upvoteBtn.classList.add("text-orange-600");
      } else if (voteType === "downvote") {
         downvoteBtn.classList.remove(
            "text-gray-400",
            "hover:text-gray-600",
            "dark:hover:text-gray-300"
         );
         downvoteBtn.classList.add("text-red-600");
      }
   }

   escapeHtml(text) {
      const div = document.createElement("div");
      div.textContent = text;
      return div.innerHTML;
   }

   showMessage(message, type = "info") {
      const notification = document.createElement("div");
      notification.className = `fixed top-4 right-4 px-4 py-2 rounded-md text-white z-50 ${
         type === "success"
            ? "bg-green-500"
            : type === "error"
            ? "bg-red-500"
            : "bg-blue-500"
      }`;
      notification.textContent = message;

      document.body.appendChild(notification);

      setTimeout(() => {
         notification.remove();
      }, 3000);
   }
}

document.addEventListener("DOMContentLoaded", () => {
   const subverseName = window.location.pathname.split("/")[2];
   if (subverseName) {
      new SubverseManager(subverseName);
   }
});
