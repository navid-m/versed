class Toast {
   toastContainer: HTMLDivElement | null;
   constructor() {
      this.toastContainer = null;
      this.initializeContainer();
   }

   initializeContainer() {
      this.toastContainer = document.createElement("div");
      this.toastContainer.id = "toast-container";
      this.toastContainer.className = "fixed top-4 right-4 z-50 space-y-2 w-80";
      document.body.appendChild(this.toastContainer);
   }

   show(message, type = "info", duration = 5000) {
      const toast = document.createElement("div");
      const typeClasses = {
         success: "bg-green-500 text-white",
         error: "bg-red-500 text-white",
         info: "bg-blue-500 text-white",
         warning: "bg-yellow-500 text-black",
      };

      toast.className = `p-4 rounded-md shadow-lg transform transition-all duration-300 ease-in-out ${
         typeClasses[type] || typeClasses.info
      } opacity-0 translate-x-8`;
      toast.innerHTML = `
            <div class="flex items-start">
                <div class="flex-1">${message}</div>
                <button class="ml-2 text-current opacity-70 hover:opacity-100 focus:outline-none">
                    &times;
                </button>
            </div>
        `;
      const closeButton = toast.querySelector("button");
      closeButton.onclick = () => this.removeToast(toast);
      this.toastContainer.appendChild(toast);
      void toast.offsetWidth;

      toast.classList.remove("opacity-0", "translate-x-8");
      toast.classList.add("opacity-100", "translate-x-0");

      if (duration > 0) {
         setTimeout(() => {
            this.removeToast(toast);
         }, duration);
      }

      return toast;
   }

   removeToast(toast) {
      toast.classList.remove("opacity-100", "translate-x-0");
      toast.classList.add("opacity-0", "translate-x-8");
      setTimeout(() => {
         if (toast.parentNode === this.toastContainer) {
            this.toastContainer.removeChild(toast);
         }
      }, 300);
   }
}

(window as any).toast = new Toast();

(window as any).showToast = {
   success: (message, duration) =>
      (window as any).toast.show(message, "success", duration),
   error: (message, duration) => (window as any).toast.show(message, "error", duration),
   info: (message, duration) => (window as any).toast.show(message, "info", duration),
   warning: (message, duration) =>
      (window as any).toast.show(message, "warning", duration),
};
