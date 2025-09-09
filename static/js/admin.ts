class AdminPanel {
   constructor() {
      this.init();
   }

   init() {
      this.loadBannedIPs();
      this.loadSubverses();
      this.setupEventListeners();
   }

   setupEventListeners() {
      const banIPBtn = document.getElementById("banIPBtn");
      if (banIPBtn) {
         banIPBtn.addEventListener("click", () => this.banIP());
      }

      const createSubverseBtn = document.getElementById("createSubverseBtn");
      if (createSubverseBtn) {
         createSubverseBtn.addEventListener("click", () =>
            this.createSubverse()
         );
      }
   }

   async createSubverse() {
      const subverseName = (document.getElementById("subverseName") as HTMLInputElement).value.trim();
      const createSubverseBtn = document.getElementById("createSubverseBtn");

      if (!subverseName) {
         this.showMessage("Enter a subverse name", "error");
         return;
      }

      (createSubverseBtn as HTMLButtonElement).disabled = true;
      createSubverseBtn.innerHTML =
         '<i class="fas fa-spinner fa-spin mr-2"></i>Creating...';

      try {
         const response = await fetch("/api/admin/subverses", {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ name: subverseName }),
         });

         if (response.ok) {
            this.showMessage("Subverse created successfully", "success");
            (document.getElementById("subverseName") as HTMLInputElement).value = "";
            this.loadSubverses();
         } else {
            const error = await response.json();
            this.showMessage(
               error.error || "Failed to create subverse",
               "error"
            );
         }
      } catch (error) {
         console.error("Error creating subverse:", error);
         this.showMessage("Failed to create subverse", "error");
      } finally {
         (createSubverseBtn as HTMLButtonElement).disabled = false;
         createSubverseBtn.innerHTML =
            '<i class="fas fa-plus mr-2"></i>Create Subverse';
      }
   }

   async loadSubverses() {
      try {
         const response = await fetch("/api/subverses");
         if (response.ok) {
            const data = await response.json();
            this.renderSubverses(data.subverses || []);
         } else {
            console.error("Failed to load subverses");
            this.renderSubverses([]);
         }
      } catch (error) {
         console.error("Error loading subverses:", error);
         this.renderSubverses([]);
      }
   }

   renderSubverses(subverses) {
      const container = document.getElementById("subversesList");
      const noSubverses = document.getElementById("noSubverses");

      if (!container) return;

      if (subverses.length === 0) {
         container.innerHTML = "";
         if (noSubverses) noSubverses.style.display = "block";
         return;
      }

      if (noSubverses) noSubverses.style.display = "none";

      container.innerHTML = subverses
         .map(
            (subverse) => `
            <div class="border border-gray-200 dark:border-gray-600 rounded-lg p-4" data-subverse-id="${
               subverse.id
            }">
                <div class="flex items-center justify-between">
                    <div>
                        <div class="flex items-center space-x-3">
                            <i class="fas fa-folder text-purple-500"></i>
                            <span class="font-medium text-gray-900 dark:text-gray-100">/s/${
                               subverse.name
                            }</span>
                        </div>
                        <div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
                            <div>Created: ${new Date(
                               subverse.created_at
                            ).toLocaleString()}</div>
                        </div>
                    </div>
                </div>
            </div>
        `
         )
         .join("");
   }

   async banIP() {
      const ipAddress = (document.getElementById("banIP") as HTMLInputElement).value.trim();
      const reason = (document.getElementById("banReason") as HTMLInputElement).value.trim();
      const banIPBtn = document.getElementById("banIPBtn");

      if (!ipAddress) {
         this.showMessage("Enter an IP address", "error");
         return;
      }

      const ipRegex =
         /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
      if (!ipRegex.test(ipAddress)) {
         this.showMessage("Enter a valid IP address", "error");
         return;
      }

      (banIPBtn as HTMLButtonElement).disabled = true;
      banIPBtn.innerHTML =
         '<i class="fas fa-spinner fa-spin mr-2"></i>Banning...';

      try {
         const response = await fetch("/api/admin/ban-ip", {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ ipAddress, reason }),
         });

         if (response.ok) {
            this.showMessage("IP address banned successfully", "success");
            (document.getElementById("banIP") as HTMLInputElement).value = "";
            (document.getElementById("banReason") as HTMLInputElement).value = "";
            this.loadBannedIPs();
         } else {
            const error = await response.json();
            this.showMessage(
               error.error || "Failed to ban IP address",
               "error"
            );
         }
      } catch (error) {
         console.error("Error banning IP:", error);
         this.showMessage("Failed to ban IP address", "error");
      } finally {
         (banIPBtn as HTMLButtonElement).disabled = false;
         banIPBtn.innerHTML = '<i class="fas fa-ban mr-2"></i>Ban IP Address';
      }
   }

   async loadBannedIPs() {
      try {
         const response = await fetch("/api/admin/banned-ips");
         if (response.ok) {
            const data = await response.json();
            this.renderBannedIPs(data.bannedIPs || []);
         } else {
            console.error("Failed to load banned IPs");
            this.renderBannedIPs([]);
         }
      } catch (error) {
         console.error("Error loading banned IPs:", error);
         this.renderBannedIPs([]);
      }
   }

   renderBannedIPs(bannedIPs: any[]) {
      const container = document.getElementById("bannedIPsList");
      const noBannedIPs = document.getElementById("noBannedIPs");

      if (!container) return;

      if (bannedIPs.length === 0) {
         container.innerHTML = "";
         if (noBannedIPs) noBannedIPs.style.display = "block";
         return;
      }

      if (noBannedIPs) noBannedIPs.style.display = "none";

      container.innerHTML = bannedIPs
         .map(
            (bannedIP) => `
            <div class="border border-gray-200 dark:border-gray-600 rounded-lg p-4" data-ip-id="${
               bannedIP.ID
            }">
                <div class="flex items-center justify-between">
                    <div>
                        <div class="flex items-center space-x-3">
                            <i class="fas fa-ban text-red-500"></i>
                            <span class="font-medium text-gray-900 dark:text-gray-100">${
                               bannedIP.IPAddress
                            }</span>
                            ${
                               bannedIP.IsActive
                                  ? '<span class="px-2 py-1 text-xs bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200 rounded-full">Active</span>'
                                  : '<span class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200 rounded-full">Inactive</span>'
                            }
                        </div>
                        <div class="mt-2 text-sm text-gray-600 dark:text-gray-400">
                            <div>Banned: ${new Date(
                               bannedIP.BannedAt
                            ).toLocaleString()}</div>
                            ${
                               bannedIP.Reason
                                  ? `<div>Reason: ${bannedIP.Reason}</div>`
                                  : ""
                            }
                            ${
                               bannedIP.UnbannedAt
                                  ? `<div>Unbanned: ${new Date(
                                       bannedIP.UnbannedAt
                                    ).toLocaleString()}</div>`
                                  : ""
                            }
                        </div>
                    </div>
                    <div class="flex items-center space-x-2">
                        ${
                           bannedIP.IsActive
                              ? `<button class="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700 transition-colors unban-btn" data-ip="${bannedIP.IPAddress}">
                                <i class="fas fa-check mr-1"></i>Unban
                            </button>`
                              : ""
                        }
                    </div>
                </div>
            </div>
        `
         )
         .join("");

      container.querySelectorAll(".unban-btn").forEach((button) => {
         button.addEventListener("click", (e) => {
            const ipAddress = ((e.target as HTMLButtonElement).closest(".unban-btn") as HTMLButtonElement).dataset.ip;
            this.unbanIP(ipAddress);
         });
      });
   }

   async unbanIP(ipAddress: string) {
      if (!confirm(`Are you sure you want to unban IP address ${ipAddress}?`)) {
         return;
      }

      try {
         const response = await fetch("/api/admin/unban-ip", {
            method: "POST",
            headers: {
               "Content-Type": "application/json",
            },
            body: JSON.stringify({ ipAddress }),
         });

         if (response.ok) {
            this.showMessage("IP address unbanned successfully", "success");
            this.loadBannedIPs();
         } else {
            const error = await response.json();
            this.showMessage(
               error.error || "Failed to unban IP address",
               "error"
            );
         }
      } catch (error) {
         console.error("Error unbanning IP:", error);
         this.showMessage("Failed to unban IP address", "error");
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
}

document.addEventListener("DOMContentLoaded", () => {
   if (window.location.pathname === "/admin") {
      new AdminPanel();
   }
});
