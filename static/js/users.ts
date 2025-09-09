document.addEventListener("DOMContentLoaded", function () {
   function fetchUsers() {
      console.log("Fetching users from /api/admin/users...");
      fetch("/api/admin/users")
         .then((response) => {
            console.log("Response status:", response.status);
            if (!response.ok) {
               console.error(
                  "Response not OK:",
                  response.status,
                  response.statusText
               );
               throw new Error(
                  `HTTP ${response.status}: ${response.statusText}`
               );
            }
            return response.json();
         })
         .then((data) => {
            console.log("Received data:", data);
            const usersList = document.getElementById("usersList");
            const noUsers = document.getElementById("noUsers");

            if (data.users && data.users.length > 0) {
               console.log(`Found ${data.users.length} users`);
               usersList.innerHTML = "";
               data.users.forEach((user) => {
                  console.log("Processing user:", user);
                  const userElement = document.createElement("div");
                  userElement.className =
                     "bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm p-4 mb-4";
                  userElement.innerHTML = `
                            <div class="flex justify-between items-center">
                                <div>
                                    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">${
                                       user.username
                                    }</h3>
                                    <p class="text-gray-600 dark:text-gray-400">${
                                       user.email
                                    }</p>
                                    <p class="text-gray-600 dark:text-gray-400">IP: ${
                                       user.ip_address || "N/A"
                                    }</p>
                                </div>
                            </div>
                        `;
                  usersList.appendChild(userElement);
               });
               noUsers.style.display = "none";
            } else {
               console.log("No users found or empty response");
               usersList.innerHTML = "";
               noUsers.style.display = "block";
            }
         })
         .catch((error) => {
            console.error("Error fetching users:", error);
            // Show error in UI
            const usersList = document.getElementById("usersList");
            const noUsers = document.getElementById("noUsers");
            usersList.innerHTML =
               '<p class="text-red-500">Error loading users: ' +
               error.message +
               "</p>";
            noUsers.style.display = "none";
         });
   }
   fetchUsers();
});
