class CommentsManager {
   constructor() {
      this.currentCommentCount = 0;
      this.init();
   }

   init() {
      this.initializeCommentCount();
      this.setupIndexPageListener();
      document.addEventListener("click", (e) => {
         if (e.target.closest(".view-comments-btn")) {
            e.preventDefault();
            const button = e.target.closest(".view-comments-btn");
            const postId = button.dataset.postId;
            this.viewPostComments(postId);
         }
      });

      const submitButton = document.getElementById("submitComment");
      if (submitButton) {
         submitButton.addEventListener("click", () => this.submitComment());
      }

      document.addEventListener("click", (e) => {
         if (e.target.closest(".edit-comment-btn")) {
            e.preventDefault();
            const button = e.target.closest(".edit-comment-btn");
            const commentId = button.dataset.commentId;
            this.editComment(commentId);
         }
      });

      document.addEventListener("click", (e) => {
         if (e.target.closest(".delete-comment-btn")) {
            e.preventDefault();
            const button = e.target.closest(".delete-comment-btn");
            const commentId = button.dataset.commentId;
            this.deleteComment(commentId);
         }
      });

      // Reply button handlers
      document.addEventListener("click", (e) => {
         if (e.target.closest(".reply-btn")) {
            e.preventDefault();
            const button = e.target.closest(".reply-btn");
            const commentId = button.dataset.commentId;
            this.showReplyForm(commentId);
         }
      });

      document.addEventListener("click", (e) => {
         if (e.target.closest(".cancel-reply-btn")) {
            e.preventDefault();
            const button = e.target.closest(".cancel-reply-btn");
            const commentId = button.dataset.commentId;
            this.cancelReply(commentId);
         }
      });

      document.addEventListener("click", (e) => {
         if (e.target.closest(".submit-reply-btn")) {
            e.preventDefault();
            const button = e.target.closest(".submit-reply-btn");
            const commentId = button.dataset.commentId;
            const parentId = button.dataset.parentId;
            const postId = button.dataset.postId;
            this.submitReply(commentId, parentId, postId);
         }
      });

      document.addEventListener("keydown", (e) => {
         if (e.key === "Escape") {
            this.cancelEdit();
         }
      });
   }

   initializeCommentCount() {
      const commentsList = document.getElementById("commentsList");
      if (commentsList) {
         this.currentCommentCount =
            commentsList.querySelectorAll("[data-comment-id]").length;
         this.updateCommentCountDisplay();
      }

      this.updateCommentButtonsFromStorage();
   }

   updateCommentButtonsFromStorage() {
      const commentButtons = document.querySelectorAll(".view-comments-btn");
      if (commentButtons.length === 0) return;

      console.log("Updating comment buttons from localStorage");

      const storedCounts = JSON.parse(
         localStorage.getItem("commentCounts") || "{}"
      );
      console.log("Stored comment counts:", storedCounts);

      commentButtons.forEach((button) => {
         const postId = button.dataset.postId;
         const storedCount = storedCounts[postId];

         if (storedCount !== undefined) {
            console.log(
               "Updating button for post:",
               postId,
               "to count:",
               storedCount
            );
            button.innerHTML = `
               <i class="far fa-comments mr-1"></i>
               View comments (${storedCount})
            `;
         }
      });
   }

   viewPostComments(postId) {
      window.location.href = `/post/${postId}`;
   }

   async submitComment() {
      const content = document.getElementById("commentContent").value.trim();
      const submitButton = document.getElementById("submitComment");
      const postId = submitButton.dataset.postId;

      if (!content) {
         this.showMessage("Comment cannot be empty", "error");
         return;
      }

      const currentUsername = document.body.dataset.username || "Anonymous";

      submitButton.disabled = true;
      submitButton.innerHTML =
         '<i class="fas fa-spinner fa-spin mr-2"></i>Posting...';

      try {
         const response = await fetch(`/posts/${postId}/comments`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ content }),
         });

         if (response.ok) {
            const serverComment = await response.json();

            const optimisticComment = {
               ID: serverComment.id || Date.now(),
               PostID: postId,
               UserID: parseInt(document.body.dataset.userId) || 0,
               Username: currentUsername,
               Content: content,
               ParentID: null,
               CreatedAt: new Date().toISOString(),
               UpdatedAt: new Date().toISOString(),
               Replies: [],
            };

            this.addCommentToUI(optimisticComment);
            this.currentCommentCount++;
            this.updateCommentCountDisplay();
            this.updateIndexPageCommentCount(postId, this.currentCommentCount);

            document.getElementById("commentContent").value = "";

            this.showMessage("Comment posted successfully", "success");

            if (serverComment.id) {
               setTimeout(() => {
                  this.refreshCommentFromServer(
                     serverComment.id,
                     optimisticComment.ID
                  );
               }, 1000);
            }
         } else {
            const error = await response.json();
            this.showMessage(error.error || "Failed to post comment", "error");
         }
      } catch (error) {
         console.error("Error posting comment:", error);
         this.showMessage("Failed to post comment", "error");
      } finally {
         submitButton.disabled = false;
         submitButton.innerHTML =
            '<i class="fas fa-paper-plane mr-2"></i>Post Comment';
      }
   }

   refreshCommentFromServer(serverId, tempId) {
      fetch(`/api/comments/${serverId}`)
         .then((response) => response.json())
         .then((updatedComment) => {
            const optimisticElement = document.querySelector(
               `[data-comment-id="${tempId}"]`
            );
            if (optimisticElement) {
               optimisticElement.outerHTML =
                  this.createCommentHTML(updatedComment);
            }
         })
         .catch((error) => {
            console.error("Error refreshing comment:", error);
         });
   }

   updateIndexPageCommentCount(postId, newCount) {
      console.log("Updating comment count for post:", postId, "to:", newCount);

      const commentCounts = JSON.parse(
         localStorage.getItem("commentCounts") || "{}"
      );
      commentCounts[postId] = newCount;
      localStorage.setItem("commentCounts", JSON.stringify(commentCounts));

      console.log("Stored comment counts:", commentCounts);
      this.updateLocalCommentCount(postId, newCount);
   }

   setupIndexPageListener() {
      if (this.indexPageListenerSet) return;
      this.indexPageListenerSet = true;

      console.log("Setting up comment count listener");

      window.addEventListener("message", (event) => {
         console.log("Received message:", event.data);

         if (event.data.type === "updateCommentCount") {
            const { postId, newCount } = event.data;
            console.log(
               "Updating comment count for post:",
               postId,
               "to:",
               newCount
            );
            this.updateLocalCommentCount(postId, newCount);
         }
      });
   }

   updateLocalCommentCount(postId, newCount) {
      const commentButtons = document.querySelectorAll(".view-comments-btn");
      console.log("Found comment buttons:", commentButtons.length);

      commentButtons.forEach((button) => {
         const buttonPostId = button.dataset.postId;
         console.log("Checking button with postId:", buttonPostId);

         if (buttonPostId === postId) {
            console.log("Updating button for post:", postId);
            button.innerHTML = `
               <i class="far fa-comments mr-1"></i>
               View comments (${newCount})
            `;
         }
      });
   }

   async editComment(commentId) {
      const commentElement = document.querySelector(
         `[data-comment-id="${commentId}"]`
      );
      const contentElement = commentElement.querySelector(
         ".text-gray-700.dark\\:text-gray-300"
      );

      if (commentElement.classList.contains("editing")) {
         return;
      }

      commentElement.classList.add("editing");
      const originalContent = contentElement.textContent.trim();

      contentElement.innerHTML = `
            <textarea class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 resize-vertical text-sm" rows="3">${originalContent}</textarea>
            <div class="flex justify-end space-x-2 mt-2">
                <button class="px-2 py-1 text-xs bg-gray-500 text-white rounded hover:bg-gray-600 save-edit-btn" data-comment-id="${commentId}">Save</button>
                <button class="px-2 py-1 text-xs bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-400 dark:hover:bg-gray-500 cancel-edit-btn">Cancel</button>
            </div>
        `;

      const textarea = contentElement.querySelector("textarea");
      textarea.focus();
      textarea.setSelectionRange(textarea.value.length, textarea.value.length);
      contentElement
         .querySelector(".save-edit-btn")
         .addEventListener("click", () => {
            this.saveCommentEdit(commentId);
         });

      contentElement
         .querySelector(".cancel-edit-btn")
         .addEventListener("click", () => {
            this.cancelEdit();
         });
   }

   async saveCommentEdit(commentId) {
      const commentElement = document.querySelector(
         `[data-comment-id="${commentId}"]`
      );
      const textarea = commentElement.querySelector("textarea");
      const newContent = textarea.value.trim();

      if (!newContent) {
         this.showMessage("Comment cannot be empty", "error");
         return;
      }

      try {
         const response = await fetch(`/api/comments/${commentId}`, {
            method: "PUT",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ content: newContent }),
         });

         if (response.ok) {
            const updatedComment = await response.json();
            this.updateCommentInUI(updatedComment);
            this.showMessage("Comment updated successfully", "success");
         } else {
            const error = await response.json();
            this.showMessage(
               error.error || "Failed to update comment",
               "error"
            );
         }
      } catch (error) {
         console.error("Error updating comment:", error);
         this.showMessage("Failed to update comment", "error");
      }
   }

   cancelEdit() {
      const editingElement = document.querySelector(".editing");
      if (editingElement) {
         const commentId = editingElement.dataset.commentId;
         this.refreshComment(commentId);
      }
   }

   async deleteComment(commentId) {
      if (!confirm("Are you sure you want to delete this comment?")) {
         return;
      }

      try {
         const response = await fetch(`/api/comments/${commentId}`, {
            method: "DELETE",
         });

         if (response.ok) {
            this.removeCommentFromUI(commentId);
            this.currentCommentCount--;
            this.updateCommentCountDisplay();

            const postId =
               document.querySelector("[data-post-id]")?.dataset.postId;
            if (postId) {
               this.updateIndexPageCommentCount(
                  postId,
                  this.currentCommentCount
               );
            }

            this.showMessage("Comment deleted successfully", "success");
         } else {
            const error = await response.json();
            this.showMessage(
               error.error || "Failed to delete comment",
               "error"
            );
         }
      } catch (error) {
         console.error("Error deleting comment:", error);
         this.showMessage("Failed to delete comment", "error");
      }
   }

   addCommentToUI(comment) {
      const commentsList = document.getElementById("commentsList");
      if (!commentsList) return;

      const noCommentsElement = commentsList.querySelector(".text-center.py-8");
      if (noCommentsElement) {
         noCommentsElement.remove();
      }

      const commentHTML = this.createCommentHTML(comment);
      commentsList.insertAdjacentHTML("afterbegin", commentHTML);
   }

   updateCommentInUI(comment) {
      const commentElement = document.querySelector(
         `[data-comment-id="${comment.ID}"]`
      );
      if (!commentElement) return;

      commentElement.classList.remove("editing");
      const contentElement = commentElement.querySelector(
         ".text-gray-700.dark\\:text-gray-300"
      );
      contentElement.innerHTML = comment.Content;
   }

   removeCommentFromUI(commentId) {
      const commentElement = document.querySelector(
         `[data-comment-id="${commentId}"]`
      );
      if (commentElement) {
         commentElement.remove();
      }

      const commentsList = document.getElementById("commentsList");
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
         console.error("Error refreshing comment:", error);
      }
   }

   createCommentHTML(comment) {
      const currentUserId = parseInt(document.body.dataset.userId) || 0;
      const isOwner = currentUserId === comment.UserID;

      const username = comment.Username || "Anonymous";
      const createdAt = comment.CreatedAt
         ? new Date(comment.CreatedAt)
         : new Date();

      return `
            <div class="border-l-2 border-gray-200 dark:border-gray-600 pl-4" data-comment-id="${
               comment.ID
            }">
                <div class="flex items-start space-x-3">
                    <div class="flex-shrink-0">
                        <div class="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                            <i class="fas fa-user text-gray-600 dark:text-gray-400 text-xs"></i>
                        </div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center space-x-2 mb-1">
                            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">${username}</span>
                            <span class="text-xs text-gray-500 dark:text-gray-400">${createdAt.toLocaleDateString(
                               "en-US",
                               {
                                  year: "numeric",
                                  month: "short",
                                  day: "numeric",
                                  hour: "numeric",
                                  minute: "2-digit",
                                  hour12: true,
                               }
                            )}</span>
                        </div>
                        <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                            ${comment.Content}
                        </div>
                        ${
                           isOwner
                              ? `
                        <div class="flex items-center space-x-2 mt-2">
                            <button class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 edit-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-edit mr-1"></i>Edit
                            </button>
                            <button class="text-xs text-red-500 hover:text-red-700 delete-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-trash mr-1"></i>Delete
                            </button>
                        </div>
                        `
                              : ""
                        }
                    </div>
                </div>
            </div>
        `;
   }

   updateCommentCount() {
      const commentsList = document.getElementById("commentsList");
      if (!commentsList) return;

      const commentCount =
         commentsList.querySelectorAll("[data-comment-id]").length;
      const header = document.querySelector("h2");
      if (header) {
         header.textContent = `Comments (${commentCount})`;
      }
   }

   updateCommentCountDisplay() {
      const header = document.querySelector("h2");
      if (header) {
         header.textContent = `Comments (${this.currentCommentCount})`;
      }
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

   // Reply functionality methods
   showReplyForm(commentId) {
      // Hide any existing reply forms
      document.querySelectorAll(".reply-form").forEach((form) => {
         form.classList.add("hidden");
      });

      // Show the reply form for this comment
      const replyForm = document.querySelector(
         `.reply-form[data-comment-id="${commentId}"]`
      );
      if (replyForm) {
         replyForm.classList.remove("hidden");
         const textarea = replyForm.querySelector("textarea");
         if (textarea) {
            textarea.focus();
         }
      }
   }

   cancelReply(commentId) {
      const replyForm = document.querySelector(
         `.reply-form[data-comment-id="${commentId}"]`
      );
      if (replyForm) {
         replyForm.classList.add("hidden");
         const textarea = replyForm.querySelector("textarea");
         if (textarea) {
            textarea.value = "";
         }
      }
   }

   async submitReply(commentId, parentId, postId) {
      const replyForm = document.querySelector(
         `.reply-form[data-comment-id="${commentId}"]`
      );
      const textarea = replyForm.querySelector("textarea");
      const submitBtn = replyForm.querySelector(".submit-reply-btn");

      const content = textarea.value.trim();

      if (!content) {
         this.showMessage("Reply cannot be empty", "error");
         return;
      }

      submitBtn.disabled = true;
      submitBtn.textContent = "Posting...";

      try {
         const response = await fetch(`/posts/${postId}/comments`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({
               content: content,
               parent_id: parentId || null,
            }),
         });

         if (response.ok) {
            const newReply = await response.json();

            console.log("Server response for new reply:", newReply);

            // Ensure the reply has the correct data
            if (!newReply.Username) {
               newReply.Username =
                  document.body.dataset.username || "Anonymous";
            }
            if (!newReply.CreatedAt) {
               newReply.CreatedAt = new Date().toISOString();
            }
            if (!newReply.Content) {
               console.error("Reply content is missing from server response!");
               newReply.Content = content; // Use the content we just sent
            }

            console.log("Reply object after processing:", newReply);

            // Add the reply to the UI
            this.addReplyToUI(newReply, parentId);

            // Hide the reply form and clear it
            this.cancelReply(commentId);

            // Update comment count
            this.currentCommentCount++;
            this.updateCommentCountDisplay();
            this.updateIndexPageCommentCount(postId, this.currentCommentCount);

            this.showMessage("Reply posted successfully", "success");
         } else {
            const error = await response.json();
            this.showMessage(error.error || "Failed to post reply", "error");
         }
      } catch (error) {
         console.error("Error posting reply:", error);
         this.showMessage("Failed to post reply", "error");
      } finally {
         submitBtn.disabled = false;
         submitBtn.textContent = "Reply";
      }
   }

   addReplyToUI(reply, parentId) {
      console.log("addReplyToUI called with reply:", reply, "parentId:", parentId);

      const parentComment = document.querySelector(
         `[data-comment-id="${parentId}"]`
      );

      console.log("Parent comment found:", parentComment);

      if (!parentComment) {
         console.error("Parent comment not found for parentId:", parentId);
         console.log("Available comment IDs:");
         document.querySelectorAll("[data-comment-id]").forEach(el => {
            console.log("Comment ID:", el.getAttribute("data-comment-id"));
         });
         return;
      }

      // Get the post ID from the body element
      const postId = document.body.dataset.postId || "";
      console.log("Using postId:", postId);

      // Calculate depth based on how many parent comments we have
      let depth = 0;
      let currentElement = parentComment;
      while (currentElement && currentElement !== document.querySelector("#commentsList")) {
         if (currentElement.hasAttribute("data-comment-id")) {
            depth++;
         }
         currentElement = currentElement.parentElement;
      }
      console.log("Calculated depth:", depth);

      // Find the replies container or create one
      let repliesContainer = parentComment.querySelector(".replies-container");
      console.log("Existing replies container:", repliesContainer);

      if (!repliesContainer) {
         repliesContainer = document.createElement("div");
         repliesContainer.className = "replies-container mt-3 space-y-2";
         parentComment.appendChild(repliesContainer);
         console.log("Created new replies container");
      }

      // Create the reply HTML with proper depth
      const replyHTML = this.createReplyHTML(reply, postId, depth);
      console.log("Generated reply HTML:", replyHTML.substring(0, 200) + "...");

      repliesContainer.insertAdjacentHTML("beforeend", replyHTML);
      console.log("Reply HTML added to container");
   }

   createCommentHTML(comment) {
      const currentUserId = parseInt(document.body.dataset.userId) || 0;
      const isOwner = currentUserId === comment.UserID;

      const username = comment.Username || "Anonymous";
      const createdAt = comment.CreatedAt
         ? new Date(comment.CreatedAt)
         : new Date();

      // Check if this is a reply (has parent_id)
      const isReply =
         comment.ParentID !== null && comment.ParentID !== undefined;
      const depthClass = isReply ? "ml-4" : "";
      const avatarSize = isReply ? "w-6 h-6" : "w-8 h-8";
      const avatarIcon = isReply ? "fa-reply" : "fa-user";

      return `
            <div class="border-l-2 border-gray-200 dark:border-gray-600 pl-4 ${depthClass}" data-comment-id="${
         comment.ID
      }">
                <div class="flex items-start space-x-3">
                    <div class="flex-shrink-0">
                        <div class="${avatarSize} bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                            <i class="fas ${avatarIcon} text-gray-600 dark:text-gray-400 text-xs"></i>
                        </div>
                    </div>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center space-x-2 mb-1">
                            <span class="text-sm font-medium text-gray-900 dark:text-gray-100">${username}</span>
                            <span class="text-xs text-gray-500 dark:text-gray-400">${createdAt.toLocaleDateString(
                               "en-US",
                               {
                                  year: "numeric",
                                  month: "short",
                                  day: "numeric",
                                  hour: "numeric",
                                  minute: "2-digit",
                                  hour12: true,
                               }
                            )}</span>
                        </div>
                        <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                            ${comment.Content}
                        </div>
                        <div class="flex items-center space-x-2 mt-2">
                            <button class="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-500 reply-btn" data-comment-id="${
                               comment.ID
                            }" data-parent-id="${comment.ID}" data-post-id="${
         comment.PostID || document.body.dataset.postId
      }">
                                <i class="fas fa-reply mr-1"></i>Reply
                            </button>
                            ${
                               isOwner
                                  ? `
                            <button class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 edit-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-edit mr-1"></i>Edit
                            </button>
                            <button class="text-xs text-red-500 hover:text-red-700 delete-comment-btn" data-comment-id="${comment.ID}">
                                <i class="fas fa-trash mr-1"></i>Delete
                            </button>
                            `
                                  : ""
                            }
                        </div>

                        <!-- Reply Form (hidden by default) -->
                        <div class="reply-form mt-3 hidden" data-comment-id="${
                           comment.ID
                        }">
                            <div class="flex items-start space-x-3">
                                <div class="flex-shrink-0">
                                    <div class="w-6 h-6 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                                        <i class="fas fa-reply text-gray-600 dark:text-gray-400 text-xs"></i>
                                    </div>
                                </div>
                                <div class="flex-1">
                                    <textarea class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 text-sm resize-y" rows="2" placeholder="Write a reply..." data-reply-content="${
                                       comment.ID
                                    }"></textarea>
                                    <div class="flex justify-end space-x-2 mt-2">
                                        <button class="px-3 py-1 text-xs text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 cancel-reply-btn" data-comment-id="${
                                           comment.ID
                                        }">Cancel</button>
                                        <button class="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 submit-reply-btn" data-comment-id="${
                                           comment.ID
                                        }" data-parent-id="${
         comment.ID
      }" data-post-id="${
         comment.PostID || document.body.dataset.postId
      }">Reply</button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Nested Replies -->
                        ${
                           comment.Replies && comment.Replies.length > 0
                              ? `
                        <div class="replies-container mt-3 space-y-2" data-replies-container="${
                           comment.ID
                        }">
                            ${comment.Replies.map((reply) =>
                               this.createReplyHTML(
                                  reply,
                                  comment.PostID || document.body.dataset.postId
                               )
                            ).join("")}
                        </div>
                        `
                              : ""
                        }
                    </div>
                </div>
            </div>
        `;
   }

   createReplyHTML(reply, postId = "", depth = 1) {
      const currentUserId = parseInt(document.body.dataset.userId) || 0;
      const isOwner = currentUserId === reply.UserID;

      const username = reply.Username || "Anonymous";
      const createdAt = reply.CreatedAt
         ? new Date(reply.CreatedAt)
         : new Date();

      // Calculate margin based on depth
      const marginClass = depth > 1 ? `ml-${4 + (depth - 1) * 4}` : "ml-4";

      return `
         <div class="border-l-2 border-gray-200 dark:border-gray-600 pl-4 ${marginClass}" data-comment-id="${
            reply.ID
         }">
            <div class="flex items-start space-x-3">
               <div class="flex-shrink-0">
                  <div class="w-6 h-6 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                     <i class="fas fa-reply text-gray-600 dark:text-gray-400 text-xs"></i>
                  </div>
               </div>
               <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-2 mb-1">
                     <span class="text-sm font-medium text-gray-900 dark:text-gray-100">${username}</span>
                     <span class="text-xs text-gray-500 dark:text-gray-400">${createdAt.toLocaleDateString(
                        "en-US",
                        {
                           year: "numeric",
                           month: "short",
                           day: "numeric",
                           hour: "numeric",
                           minute: "2-digit",
                           hour12: true,
                        }
                     )}</span>
                  </div>
                  <div class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed">
                     ${reply.Content}
                  </div>
                  <div class="flex items-center space-x-2 mt-2">
                     <button class="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-500 reply-btn" data-comment-id="${
                        reply.ID
                     }" data-parent-id="${reply.ID}" data-post-id="${postId}">
                        <i class="fas fa-reply mr-1"></i>Reply
                     </button>
                     ${
                        isOwner
                           ? `
                        <button class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 edit-comment-btn" data-comment-id="${reply.ID}">
                           <i class="fas fa-edit mr-1"></i>Edit
                        </button>
                        <button class="text-xs text-red-500 hover:text-red-700 delete-comment-btn" data-comment-id="${reply.ID}">
                           <i class="fas fa-trash mr-1"></i>Delete
                        </button>
                     `
                           : ""
                     }
                  </div>

                  <!-- Reply Form (hidden by default) -->
                  <div class="reply-form mt-3 hidden" data-comment-id="${
                     reply.ID
                  }">
                     <div class="flex items-start space-x-3">
                        <div class="flex-shrink-0">
                           <div class="w-6 h-6 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
                              <i class="fas fa-reply text-gray-600 dark:text-gray-400 text-xs"></i>
                           </div>
                        </div>
                        <div class="flex-1">
                           <textarea class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 text-sm resize-y" rows="2" placeholder="Write a reply..." data-reply-content="${
                              reply.ID
                           }"></textarea>
                           <div class="flex justify-end space-x-2 mt-2">
                              <button class="px-3 py-1 text-xs text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 cancel-reply-btn" data-comment-id="${
                                 reply.ID
                              }">Cancel</button>
                              <button class="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 submit-reply-btn" data-comment-id="${
                                 reply.ID
                              }" data-parent-id="${
         reply.ID
      }" data-post-id="${postId}">Reply</button>
                           </div>
                        </div>
                     </div>
                  </div>
               </div>
            </div>
         </div>
      `;
   }
}

document.addEventListener("DOMContentLoaded", () => {
   new CommentsManager();
});
