html.classList.add("dark");

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
