document.addEventListener("DOMContentLoaded", function () {
   const loginForm = document.querySelector('form[action="/signin"]');

   if (loginForm) {
      loginForm.addEventListener("submit", async function (e) {
         e.preventDefault();

         const formData = new FormData(loginForm);
         const submitButton = loginForm.querySelector('button[type="submit"]');
         const originalButtonText = submitButton.innerHTML;

         try {
            submitButton.disabled = true;
            submitButton.innerHTML =
               '<i class="fas fa-spinner fa-spin mr-2"></i> Signing in...';
            submitButton.classList.add("opacity-75");

            const response = await fetch("/signin", {
               method: "POST",
               body: formData,
               headers: {
                  Accept: "application/json",
               },
            });

            const data = await response.json();

            if (response.ok) {
               window.location.href = "/";
            } else if (data.toast) {
               window.showToast[data.toast.type || "error"](data.toast.message);
            } else {
               window.showToast.error(
                  "An error occurred during login. Please try again."
               );
            }
         } catch (error) {
            console.error("Login error:", error);
            window.showToast.error("An error occurred. Please try again.");
         } finally {
            submitButton.disabled = false;
            submitButton.innerHTML = originalButtonText;
            submitButton.classList.remove("opacity-75");
         }
      });
   }
});
