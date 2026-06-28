document.addEventListener('DOMContentLoaded', () => {
    const filterInput = document.getElementById('version-filter')
    const rows = [...document.querySelectorAll('.version-row')]

    filterInput.addEventListener('input', () => {
        const q = filterInput.value.toLowerCase()
        rows.forEach((row) => {
            const name = row.dataset.versionName.toLowerCase()
            row.style.display = name.includes(q) ? '' : 'none'
        })
    })
})
