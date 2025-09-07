const darkModeToggle = document.getElementById("darkModeToggle");
const toggleSlider = document.getElementById("toggleSlider");
const html = document.documentElement;

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
});

darkModeToggle.addEventListener("click", () => {
   const isDark = html.classList.toggle("dark");
   localStorage.setItem('theme', isDark ? 'dark' : 'light');
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
   const voteButtons = document.querySelectorAll("[data-vote-type]");

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
});
