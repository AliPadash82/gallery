fetch('/gallery/photos_retrieve')
.then(response => response.json())
.then(photos => {
    const gallery = document.querySelector('.gallery');
    photos.forEach(photo => {
        const imgContainer = document.createElement('div');
        const title = document.createElement('p');
        title.classList.add('title');
        imgContainer.classList.add('photo');
        const img = document.createElement('img');
        if(photo == undefined || photo.filePath == undefined) { 
            img.src = "";
            img.alt = "undefined"
        } else {
            img.src = '/' + photo.filePath.replace(/\\/g, '/');
            img.alt = photo.title;
        }
        title.textContent = photo.title;
        imgContainer.appendChild(title);
        imgContainer.appendChild(img);
        gallery.appendChild(imgContainer);
    });
})
.catch(error => console.error('Error loading gallery:', error));
