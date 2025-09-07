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
