// users.js
document.addEventListener('DOMContentLoaded', function () {
    // Fetch and display registered users with their IP addresses
    function fetchUsers() {
        fetch('/api/admin/users')
            .then(response => response.json())
            .then(data => {
                const usersList = document.getElementById('usersList');
                const noUsers = document.getElementById('noUsers');

                if (data.users && data.users.length > 0) {
                    usersList.innerHTML = '';
                    data.users.forEach(user => {
                        const userElement = document.createElement('div');
                        userElement.className = 'bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm p-4 mb-4';
                        userElement.innerHTML = `
                            <div class="flex justify-between items-center">
                                <div>
                                    <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">${user.username}</h3>
                                    <p class="text-gray-600 dark:text-gray-400">${user.email}</p>
                                    <p class="text-gray-600 dark:text-gray-400">IP: ${user.ip_address || 'N/A'}</p>
                                </div>
                            </div>
                        `;
                        usersList.appendChild(userElement);
                    });
                    noUsers.style.display = 'none';
                } else {
                    usersList.innerHTML = '';
                    noUsers.style.display = 'block';
                }
            })
            .catch(error => {
                console.error('Error fetching users:', error);
            });
    }

    // Initial fetch
    fetchUsers();

    // Set up periodic refresh
    setInterval(fetchUsers, 30000); // Refresh every 30 seconds
});