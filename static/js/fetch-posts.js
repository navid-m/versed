document.addEventListener("DOMContentLoaded", function () {
   const loadMoreButton = document.querySelector(".text-center.py-6 button");
   let currentPage = 1;

   loadMoreButton.addEventListener("click", async function () {
      try {
         const response = await fetch(`/api/feeds?page=${currentPage + 1}`);
         if (!response.ok) {
            throw new Error("Failed to fetch more posts");
         }
         const data = await response.json();
         if (data.items && data.items.length > 0) {
            const postsContainer = document.querySelector(".space-y-3");
            data.items.forEach((item) => {
               const article = document.createElement("article");
               article.className =
                  "bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md dark:hover:shadow-xl transition-shadow";
               article.innerHTML = `
            <div class="p-4">
              <div class="flex items-start space-x-3">
                <div class="flex flex-col items-center space-y-1 flex-shrink-0">
                  <button class="p-1 text-orange-500 hover:text-orange-600 transition-colors" data-feed-id="${
                     item.id
                  }" data-vote-type="upvote">
                     <i class="fas fa-chevron-up text-sm"></i>
                  </button>
                  <span class="text-xs font-medium text-orange-500">${
                     item.score || 0
                  }</span>
                  <button class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors" data-feed-id="${
                     item.id
                  }" data-vote-type="downvote">
                     <i class="fas fa-chevron-down text-sm"></i>
                  </button>
                </div>

                <div class="flex-1 min-w-0">
                  <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2 hover:text-gray-700 dark:hover:text-gray-300 transition-colors line-clamp-2">
                    <a href="${
                       item.url
                    }" target="_blank" class="hover:underline">${item.title}</a>
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
                    <button class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors view-comments-btn" data-post-id="${item.id}">
                      <i class="far fa-comments mr-1"></i>
                      View comments (${item.comments_count || 0})
                    </button>
                  </div>
                </div>
                <div class="flex flex-col items-center space-y-1 flex-shrink-0 ml-3">
                  <button class="p-1 text-gray-400 hover:text-blue-500 transition-colors save-button" data-feed-id="${
                     item.id
                  }" data-action="save" title="Save to reading list">
                    <i class="far fa-bookmark text-sm"></i>
                  </button>
                  <button class="p-1 text-gray-400 hover:text-red-500 transition-colors hide-button" data-feed-id="${
                     item.id
                  }" data-action="hide" title="Hide post">
                    <i class="far fa-eye-slash text-sm"></i>
                  </button>
                </div>
              </div>
            </div>
          `;
               postsContainer.appendChild(article);
            });

            const newSaveButtons = postsContainer.querySelectorAll(
               ".save-button:not([data-listener-attached])"
            );
            newSaveButtons.forEach((button) => {
               button.setAttribute("data-listener-attached", "true");
               const feedId = button.getAttribute("data-feed-id");

               checkSaveStatus(button, feedId);

               button.addEventListener("click", handleSaveClick);
            });
            const newVoteButtons = postsContainer.querySelectorAll(
               "[data-vote-type]:not([data-vote-listener-attached])"
            );
            newVoteButtons.forEach((button) => {
               button.setAttribute("data-vote-listener-attached", "true");
               button.addEventListener("click", handleVoteClick);
            });
            const newHideButtons = postsContainer.querySelectorAll(
               ".hide-button:not([data-hide-listener-attached])"
            );
            newHideButtons.forEach((button) => {
               button.setAttribute("data-hide-listener-attached", "true");
               button.addEventListener("click", handleHideClick);
            });

            currentPage++;
         } else {
            loadMoreButton.textContent = "No more posts to load";
            loadMoreButton.disabled = true;
         }
      } catch (error) {
         console.error("Error loading more posts:", error);
         loadMoreButton.textContent = "Error loading posts";
      }
   });
});

document.addEventListener("DOMContentLoaded", function () {
   const saveButtons = document.querySelectorAll(".save-button");

   saveButtons.forEach((button) => {
      if (!button.hasAttribute("data-listener-attached")) {
         button.setAttribute("data-listener-attached", "true");
         const feedId = button.getAttribute("data-feed-id");
         checkSaveStatus(button, feedId);
         button.addEventListener("click", handleSaveClick);
      }
   });
});

async function handleSaveClick(event) {
   const button = event.currentTarget;
   const feedId = button.getAttribute("data-feed-id");
   const action = button.getAttribute("data-action");

   try {
      let response;
      if (action === "save") {
         response = await fetch("/api/reading-list/save", {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({
               item_id: feedId,
            }),
         });
      } else {
         response = await fetch("/api/reading-list/remove", {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({
               item_id: feedId,
            }),
         });
      }

      if (!response.ok) {
         throw new Error("Failed to update reading list");
      }

      const result = await response.json();

      if (result.success) {
         if (action === "save") {
            setSaveButtonState(button, true);
         } else {
            setSaveButtonState(button, false);
         }
      }
   } catch (error) {
      console.error("Error updating reading list:", error);
   }
}

async function checkSaveStatus(button, feedId) {
   try {
      const response = await fetch("/api/reading-list/check/" + feedId);
      if (response.ok) {
         const data = await response.json();
         setSaveButtonState(button, data.saved);
      }
   } catch (error) {
      console.error("Error checking save status:", error);
   }
}

function setSaveButtonState(button, isSaved) {
   const icon = button.querySelector("i");

   if (isSaved) {
      button.setAttribute("data-action", "unsave");
      icon.className = "fas fa-bookmark text-sm";
      button.className =
         "p-1 text-blue-500 hover:text-red-500 transition-colors save-button";
      button.setAttribute("title", "Remove from reading list");
   } else {
      button.setAttribute("data-action", "save");
      icon.className = "far fa-bookmark text-sm";
      button.className =
         "p-1 text-gray-400 hover:text-blue-500 transition-colors save-button";
      button.setAttribute("title", "Save to reading list");
   }
}

// Voting handler
async function handleVoteClick(event) {
   const button = event.currentTarget;
   const feedId = button.getAttribute("data-feed-id");
   const voteType = button.getAttribute("data-vote-type");

   try {
      const response = await fetch("/api/vote", {
         method: "POST",
         headers: {
            "Content-Type": "application/json",
         },
         body: JSON.stringify({
            feed_id: feedId,
            vote_type: voteType,
         }),
      });

      if (!response.ok) {
         throw new Error("Failed to submit vote");
      }

      const data = await response.json();
      const scoreElement = button.parentElement.querySelector(
         ".text-xs.font-medium.text-orange-500"
      );
      if (scoreElement) {
         scoreElement.textContent = data.new_score;
      }
   } catch (error) {
      console.error("Error submitting vote:", error);
   }
}

// Hide handler
async function handleHideClick(event) {
   const button = event.currentTarget;
   const feedId = button.getAttribute("data-feed-id");
   const action = button.getAttribute("data-action");

   console.log(`Hiding post ${feedId}`);

   // Immediately hide the post
   const article = button.closest("article");
   if (article) {
      article.style.transition = "opacity 0.3s ease-out";
      article.style.opacity = "0";
      setTimeout(() => {
         article.style.display = "none";
      }, 300);
   }

   try {
      let response;
      if (action === "hide") {
         console.log(`Making API call to hide post ${feedId}`);
         response = await fetch(`/api/posts/${feedId}/hide`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
         });
      } else {
         response = await fetch(`/api/posts/${feedId}/unhide`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
         });
      }

      if (!response.ok) {
         console.error(`API call failed with status ${response.status}`);
         throw new Error(`Failed to ${action} post`);
      }

      const data = await response.json();
      console.log(`Post ${feedId} ${action}d successfully`, data);

   } catch (error) {
      console.error(`Error ${action}ing post:`, error);

      // Restore the post if API call failed
      if (article) {
         article.style.display = "";
         article.style.opacity = "1";
      }

      // Show error message
      alert(`Failed to ${action} post. Please try again.`);
   }
}
