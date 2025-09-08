const darkModeToggle = document.getElementById("darkModeToggle");
const toggleSlider = document.getElementById("toggleSlider");
const userMenuButton = document.getElementById("userMenuButton");
const userDropdown = document.getElementById("userDropdown");
const html = document.documentElement;

html.classList.add("dark");

function htmlDecode(input) {
   var doc = new DOMParser().parseFromString(input, "text/html");
   return doc.documentElement.textContent;
}

document.addEventListener("DOMContentLoaded", function () {
   const avatarElement = document.querySelector("[data-email]");
   if (avatarElement) {
      const email = avatarElement.getAttribute("data-email");
      if (email && email.length > 0) {
         avatarElement.textContent = email.charAt(0).toUpperCase();
      }
   }
});

darkModeToggle.addEventListener("click", () => {
   html.classList.toggle("dark");
   if (html.classList.contains("dark")) {
      toggleSlider.classList.remove("translate-x-6");
      toggleSlider.classList.add("translate-x-1");
   } else {
      toggleSlider.classList.remove("translate-x-1");
      toggleSlider.classList.add("translate-x-6");
   }
});

if (userMenuButton && userDropdown) {
   userMenuButton.addEventListener("click", (e) => {
      e.stopPropagation();
      userDropdown.classList.toggle("hidden");
   });

   document.addEventListener("click", (e) => {
      if (
         !userMenuButton.contains(e.target) &&
         !userDropdown.contains(e.target)
      ) {
         userDropdown.classList.add("hidden");
      }
   });
}

// Hidden Posts Management
document.addEventListener("DOMContentLoaded", function () {
   const hiddenPostsContainer = document.getElementById("hiddenPostsContainer");
   const refreshHiddenPostsBtn = document.getElementById("refreshHiddenPosts");

   // Load hidden posts on page load
   loadHiddenPosts();

   // Handle refresh button
   if (refreshHiddenPostsBtn) {
      refreshHiddenPostsBtn.addEventListener("click", loadHiddenPosts);
   }

   async function loadHiddenPosts() {
      if (!hiddenPostsContainer) return;

      try {
         const response = await fetch("/api/posts/hidden");
         if (!response.ok) {
            throw new Error("Failed to load hidden posts");
         }

         const data = await response.json();
         renderHiddenPosts(data.hiddenItems);
      } catch (error) {
         console.error("Error loading hidden posts:", error);
         hiddenPostsContainer.innerHTML = `
            <div class="text-center py-8 text-red-500 dark:text-red-400">
               <i class="fas fa-exclamation-triangle text-2xl mb-2"></i>
               <p>Failed to load hidden posts. Please try again.</p>
            </div>
         `;
      }
   }

   function renderHiddenPosts(hiddenItems) {
      if (!hiddenItems || hiddenItems.length === 0) {
         hiddenPostsContainer.innerHTML = `
            <div class="text-center py-8 text-gray-500 dark:text-gray-400">
               <i class="far fa-eye-slash text-2xl mb-2"></i>
               <p>No hidden posts found.</p>
               <p class="text-sm">Posts you hide from your feed will appear here.</p>
            </div>
         `;
         return;
      }

      const postsHTML = hiddenItems.map(item => `
         <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 border border-gray-200 dark:border-gray-600">
            <div class="flex items-start justify-between">
               <div class="flex-1 min-w-0">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2">
                     <a href="${item.URL}" target="_blank" class="hover:text-blue-600 dark:hover:text-blue-400">
                        ${item.Title}
                     </a>
                  </h3>
                  <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-2">
                     <span class="flex items-center">
                        <i class="far fa-user mr-1"></i>
                        ${item.Author || 'Unknown'}
                     </span>
                     <span class="flex items-center">
                        <i class="far fa-clock mr-1"></i>
                        ${new Date(item.PublishedAt).toLocaleDateString()}
                     </span>
                     <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
                        ${item.SourceName || 'Unknown'}
                     </span>
                  </div>
               </div>
               <button class="ml-3 px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-xs font-medium rounded-md transition-colors unhide-btn"
                  data-feed-id="${item.ID}">
                  <i class="fas fa-eye mr-1"></i>
                  Unhide
               </button>
            </div>
         </div>
      `).join('');

      hiddenPostsContainer.innerHTML = postsHTML;

      // Add event listeners for unhide buttons
      const unhideButtons = hiddenPostsContainer.querySelectorAll(".unhide-btn");
      unhideButtons.forEach(button => {
         button.addEventListener("click", async function () {
            const feedId = this.getAttribute("data-feed-id");
            await unhidePost(feedId, this);
         });
      });
   }

   async function unhidePost(feedId, buttonElement) {
      try {
         buttonElement.disabled = true;
         buttonElement.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i>Unhiding...';

         const response = await fetch(`/api/posts/${feedId}/unhide`, {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
         });

         if (!response.ok) {
            throw new Error("Failed to unhide post");
         }

         // Remove the post from the UI with animation
         const postElement = buttonElement.closest(".bg-gray-50, .bg-gray-700");
         postElement.style.transition = "opacity 0.3s ease-out";
         postElement.style.opacity = "0";
         setTimeout(() => {
            postElement.remove();
            // If no more posts, reload to show empty state
            if (hiddenPostsContainer.children.length === 0) {
               loadHiddenPosts();
            }
         }, 300);

      } catch (error) {
         console.error("Error unhiding post:", error);
         buttonElement.disabled = false;
         buttonElement.innerHTML = '<i class="fas fa-eye mr-1"></i>Unhide';
         // Could add error notification here
      }
   }
});
