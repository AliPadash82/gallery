var selectedFile;

document.querySelector('.close-button').addEventListener('click', function() {
    document.getElementById('modal').style.display = 'none';
});

document.getElementById('submit-button').addEventListener('click', function() {
    if (selectedFile) 
        document.getElementById('modal').style.display = 'flex'
})

document.getElementById('confirm-button').addEventListener('click', function() {
    var title = document.getElementById('photoTitle').value;
    var description = document.getElementById('description').value;
    if (selectedFile && title) {
        uploadFileToServer('/upload/upload', selectedFile, title, description);
        document.getElementById('modal').style.display = 'none'; // Close modal
    } else {
        console.log('title is missing.');
    }
});

document.getElementById('fileInput').addEventListener('change', function(event) {
    selectedFile = event.target.files[0];
    handleDraggedInput(selectedFile);
});

document.addEventListener('DOMContentLoaded', function() {
    var dropArea = document.getElementById('dropArea');

    // Prevent default drag behaviors
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    // Highlight drop area when item is dragged over it
    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false);
    });

    function highlight() {
        dropArea.classList.add('highlight');
    }

    function unhighlight() {
        dropArea.classList.remove('highlight');
    }

    // Handle dropped files
    dropArea.addEventListener('drop', handleDrop, false);

    function handleDrop(e) {
        var dt = e.dataTransfer;
        var files = dt.files;

        handleFiles(files);
    }

    function handleFiles(files) {
        ([...files]).forEach(uploadFile);
    }

    function uploadFile(file) {
        selectedFile = file;
        console.log('File you dragged:', file);
        handleDraggedInput(selectedFile);
    }
});

function handleDraggedInput(file) {
    var output = document.getElementById('preview');
    if (!output) {
        output = document.createElement('img');
        output.id = 'preview';
        output.style.maxWidth = '100%'; // Ensure the image fits in the container
        output.style.marginTop = '20px';
        document.querySelector('.upload-box').appendChild(output);
    }
    output.src = URL.createObjectURL(file);
    output.onload = function() {
        URL.revokeObjectURL(output.src); // Free memory
    }
}

function uploadFileToServer(path, file, title, description) {
    const csrf = document.querySelector('meta[name="X-CSRF-Token"]').getAttribute('content');
    const formData = new FormData();
    formData.append('file', file);
    formData.append('title', title)
    formData.append('description', description)

    fetch(path, {
        method: 'POST',
        body: formData,
        credentials: "same-origin",
        headers: {
            'X-CSRF-Token': csrf,
        }
    }).then(response => {
        if (response.ok) {
            return response.json().then(data => {
                if (data.redirect) {
                    // If the server indicates a redirect, perform it client-side
                    window.location.href = data.redirect;
                } else {
                    console.log(data); // Handle normal data
                }
            });
        } else if (response.status === 403) { // Check for forbidden status
            return response.json().then(data => {
                if (data.redirect) {
                    // Perform the redirect as instructed by the server
                    window.location.href = data.redirect;
                }
            });
        } else {
            return response.text().then(text => { throw new Error(text) });
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}  