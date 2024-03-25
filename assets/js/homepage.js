document.addEventListener("DOMContentLoaded", function() {
    fetch('/api/username').then(response => response.json())
        .then(data => document.getElementById('username').textContent = data.username)
        .catch(err => console.error('Error fetching username:', err));
});