document.getElementById('copy-btn').addEventListener('click', async () => {
    const commandText = document.querySelector('.command').textContent;
    const toast = document.getElementById('copy-toast');
    try {
        await navigator.clipboard.writeText(commandText);
        toast.textContent = 'Copied to clipboard!';
        toast.classList.add('show');
    } catch (err) {
        console.error('Failed to copy text: ', err);
        toast.textContent = 'Copy failed — select the command manually';
        toast.classList.add('show');
    } finally {
        setTimeout(() => {
            toast.classList.remove('show');
        }, 2000);
    }
});

fetch('https://api.github.com/repos/VallabhPatil3078/orbit/releases/latest')
    .then(res => res.json())
    .then(data => {
        const badge = document.getElementById('version-badge');
        if (data.tag_name) {
            badge.textContent = `${data.tag_name} is now live!`;
        } else {
            badge.textContent = 'In active development';
        }
    })
    .catch(() => {
        const badge = document.getElementById('version-badge');
        badge.textContent = 'In active development';
    });
