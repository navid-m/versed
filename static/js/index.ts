console.log("=== index.js file loaded ===");

const darkModeToggle = document.getElementById("darkModeToggle");
const toggleSlider = document.getElementById("toggleSlider");
const html = document.documentElement;

function htmlDecode(input: string) {
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

   const descriptionContainers =
      document.querySelectorAll("[data-description]");
   descriptionContainers.forEach((container) => {
      const htmlContent = container.getAttribute("data-description");
      if (htmlContent) {
         const tempDiv = document.createElement("div");
         tempDiv.innerHTML = htmlContent;
         container.innerHTML = htmlDecode(
            tempDiv.innerHTML + container.querySelector(".absolute").outerHTML
         );
      }
   });

   checkAdminStatus();
});

darkModeToggle.addEventListener("click", () => {
   const isDark = html.classList.toggle("dark");
   localStorage.setItem("theme", isDark ? "dark" : "light");
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
         !userMenuButton.contains(e.target as Node) &&
         !userDropdown.contains(e.target as Node)
      ) {
         userDropdown.classList.add("hidden");
      }
   });
}

const listViewBtn = document.getElementById("listViewBtn");
const gridViewBtn = document.getElementById("gridViewBtn");
const postsContainer = document.getElementById("postsContainer");

if (listViewBtn && gridViewBtn && postsContainer) {
   listViewBtn.addEventListener("click", () => {
      postsContainer.className = "space-y-3";
      listViewBtn.className =
         "flex items-center px-3 py-1.5 rounded-md text-xs font-medium transition-colors bg-gray-900 dark:bg-gray-600 text-white";
      gridViewBtn.className =
         "flex items-center px-3 py-1.5 rounded-md text-xs font-medium transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200";
   });

   gridViewBtn.addEventListener("click", () => {
      postsContainer.className =
         "grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4";
      gridViewBtn.className =
         "flex items-center px-3 py-1.5 rounded-md text-xs font-medium transition-colors bg-gray-900 dark:bg-gray-600 text-white";
      listViewBtn.className =
         "flex items-center px-3 py-1.5 rounded-md text-xs font-medium transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200";
   });
}
document.addEventListener("DOMContentLoaded", function () {
   console.log("=== index.js DOMContentLoaded fired ===");

   const voteButtons = document.querySelectorAll("[data-vote-type]");

   console.log(`Found ${voteButtons.length} vote buttons`);
   voteButtons.forEach((button) => {
      button.addEventListener("click", async function () {
         const feedId = this.getAttribute("data-feed-id");
         const voteType = this.getAttribute("data-vote-type");

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
            const scoreElement = this.parentElement.querySelector(
               ".text-xs.font-medium.text-orange-500"
            );
            if (scoreElement) {
               scoreElement.textContent = data.new_score;
            }
         } catch (error) {
            console.error("Error submitting vote:", error);
         }
      });
   });

   console.log("=== Looking for hide buttons ===");
   const hideButtons = document.querySelectorAll(".hide-button");
   console.log(`Found ${hideButtons.length} hide buttons`);

   hideButtons.forEach((button, index) => {
      console.log(
         `Setting up hide button ${index} with feedId: ${button.getAttribute(
            "data-feed-id"
         )}`
      );
      if (!button.hasAttribute("data-hide-listener-attached")) {
         button.setAttribute("data-hide-listener-attached", "true");
         button.addEventListener("click", async function () {
            console.log("Hide button clicked!");
            const feedId = this.getAttribute("data-feed-id");
            const action = this.getAttribute("data-action");
            console.log(`Hiding post ${feedId}, action: ${action}`);

            const article = this.closest("article");
            if (article) {
               console.log("Found article, hiding it");
               article.style.transition = "opacity 0.3s ease-out";
               article.style.opacity = "0";
               setTimeout(() => {
                  article.style.display = "none";
               }, 300);
            } else {
               console.log("No article found");
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
                  console.log(`Making API call to unhide post ${feedId}`);
                  response = await fetch(`/api/posts/${feedId}/unhide`, {
                     method: "POST",
                     headers: {
                        "Content-Type": "application/json",
                     },
                  });
               }

               console.log(`Response status: ${response.status}`);
               if (!response.ok) {
                  console.error(
                     `API call failed with status ${response.status}`
                  );
                  const errorText = await response.text();
                  console.error(`Error response: ${errorText}`);
                  throw new Error(`Failed to ${action} post`);
               }

               const data = await response.json();
               console.log(`Post ${feedId} ${action}d successfully`, data);
            } catch (error) {
               console.error(`Error ${action}ing post:`, error);
               if (article) {
                  article.style.display = "";
                  article.style.opacity = "1";
               }
               alert(`Failed to ${action} post. Try again.`);
            }
         });
      } else {
         console.log(`Hide button ${index} already has listener attached`);
      }
   });

   console.log("=== index.js setup complete ===");
});

async function checkAdminStatus() {
   const adminButton = document.querySelector(".admin-button");
   const adminDivider = document.querySelector(".admin-divider");

   if (!adminButton || !adminDivider) {
      console.log(
         "Admin button or divider not found, skipping admin status check"
      );
      return;
   }

   try {
      console.log("Checking admin status...");
      const response = await fetch("/api/user/status");

      if (response.ok) {
         const data = await response.json();
         console.log("Admin status response:", data);

         if (data.isAdmin) {
            console.log("User is admin, showing admin button");
            adminButton.classList.remove("hidden");
            adminDivider.classList.remove("hidden");
         } else {
            console.log("User is not admin, hiding admin button");
            adminButton.classList.add("hidden");
            adminDivider.classList.add("hidden");
         }
      } else {
         console.log("Failed to fetch ASTATUS");
      }
   } catch (error) {
      console.error("Error checking ASTATUS:", error);
   }
}
