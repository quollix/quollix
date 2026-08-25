function setInstallButtonDisabled(row, disabled) {
    row.dataset.canInstall = disabled ? 'false' : 'true'

    const installButton = row.querySelector('button.version-install-button')
    setButtonDisabled(installButton, disabled)
}

function disableVersionAndOlderRows(row) {
    const changedRows = []
    let currentRow = row

    while (currentRow) {
        const wasInstallable = currentRow.dataset.canInstall === 'true'
        changedRows.push({ row: currentRow, wasInstallable })
        setInstallButtonDisabled(currentRow, true)
        currentRow = currentRow.nextElementSibling
    }

    return () => {
        for (const changedRow of changedRows) {
            setInstallButtonDisabled(changedRow.row, !changedRow.wasInstallable)
        }
    }
}

window.installVersionFromVersionsPage = async (button, versionId) => {
    const row = button.closest('.version-row')
    const restore = row ? disableVersionAndOlderRows(row) : () => {}

    const ok = await installApp(versionId)

    if (!ok) {
        restore()
        return
    }
}
