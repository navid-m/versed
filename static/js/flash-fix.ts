(function () {
   const theme = localStorage.getItem("theme");
   const html = document.documentElement;
   if (
      theme === "dark" ||
      (!theme && window.matchMedia("(prefers-color-scheme: dark)").matches)
   ) {
      html.classList.add("dark");
   }
})();
