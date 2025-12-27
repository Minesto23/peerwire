document.addEventListener('DOMContentLoaded', () => {
    const fileInput = document.getElementById('fileInput');
    const fileLabel = document.getElementById('fileLabel');
    const dropZone = document.getElementById('dropZone');
    const destinationInput = document.getElementById('destination');

    // Default destination to pseudo-home if possible, or leave empty/placeholder
    // Ideally we could fetch a default from server, but for now placeholder is enough.

    // File Input Change
    fileInput.addEventListener('change', (e) => {
        if (e.target.files && e.target.files.length > 0) {
            fileLabel.innerText = e.target.files[0].name;
            dropZone.classList.add('active');
        }
    });

    // Drag and Drop Visuals
    dropZone.addEventListener('dragover', (e) => {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', () => {
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', (e) => {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        // The file input handles the actual file via the overly transparent input, 
        // but if they dropped ON the box, we might need to manually assign if not hit input directly.
        // However, the input covers the whole div essentially.

        if (e.dataTransfer.files.length) {
            fileInput.files = e.dataTransfer.files;
            fileLabel.innerText = e.dataTransfer.files[0].name;
        }
    });

    // Directory Picker Logic
    const browseBtn = document.getElementById('browseBtn');
    const modal = document.getElementById('dirModal');
    const closeModal = document.getElementById('closeModal');
    const cancelBtn = document.getElementById('cancelBtn');
    const selectBtn = document.getElementById('selectBtn');
    const folderList = document.getElementById('folderList');
    const currentPathEl = document.getElementById('currentPath');

    let currentBrowsePath = '';

    function openModal() {
        modal.classList.remove('hidden');
        modal.style.display = 'flex'; // Ensure flex display when visible
        loadDirectory(destinationInput.value || '');
    }

    function closeModalFunc() {
        modal.classList.add('hidden');
        setTimeout(() => {
            modal.style.display = 'none';
        }, 300); // Match transition duration
    }

    browseBtn.addEventListener('click', openModal);
    closeModal.addEventListener('click', closeModalFunc);
    cancelBtn.addEventListener('click', closeModalFunc);

    selectBtn.addEventListener('click', () => {
        if (currentBrowsePath) {
            destinationInput.value = currentBrowsePath;
        }
        closeModalFunc();
    });

    async function loadDirectory(path) {
        try {
            const res = await fetch(`/browse?path=${encodeURIComponent(path)}`);
            if (!res.ok) throw new Error(await res.text());
            const data = await res.json();

            currentBrowsePath = data.current;
            currentPathEl.innerText = data.current;
            folderList.innerHTML = '';

            data.folders.forEach(folder => {
                const item = document.createElement('div');
                item.className = 'folder-item';
                item.innerHTML = `<span class="icon">📁</span><span class="name">${folder.Name}</span>`;
                item.onclick = () => loadDirectory(folder.Path);
                folderList.appendChild(item);
            });
        } catch (err) {
            console.error("Browse Error:", err);
            // Optionally show error in modal
        }
    }

    // Poll Status
    setInterval(() => {
        fetch('/status')
            .then(r => r.json())
            .then(data => {
                const msgEl = document.getElementById('msg');
                const percentEl = document.getElementById('percent');
                const barEl = document.getElementById('bar');
                const statusBadge = document.getElementById('connectionStatus');

                msgEl.innerText = data.Message;
                percentEl.innerText = Math.round(data.Percent) + '%';
                barEl.style.width = data.Percent + '%';

                if (data.Percent >= 100) {
                    barEl.style.background = 'var(--success-color)';
                    statusBadge.style.color = 'var(--success-color)';
                    statusBadge.innerText = 'Completed';
                } else if (data.Running) {
                    statusBadge.style.color = 'var(--accent-color)';
                    statusBadge.innerText = 'Downloading';
                }
            })
            .catch(err => {
                console.error("Status Poll Error:", err);
                document.getElementById('connectionStatus').innerText = 'Disconnected';
                document.getElementById('connectionStatus').style.color = 'var(--error-color)';
            });
    }, 800);
});
