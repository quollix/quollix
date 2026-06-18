document.addEventListener('DOMContentLoaded', () => {
    const checkbox = document.getElementById('unofficial')
    const maintainerWrap = document.getElementById('maintainer-wrap')
    const maintainerLabel = maintainerWrap.querySelector('label')
    const maintainerInput = maintainerWrap.querySelector('input')

    function applyVisibility() {
        const show = checkbox.checked
        maintainerLabel.style.visibility = show ? 'visible' : 'hidden'
        maintainerInput.style.visibility = show ? 'visible' : 'hidden'
    }

    async function onUnofficialChange() {
        if (checkbox.checked) {
            const isConfirmed = await window.confirmDialog(
                    "Unofficial apps are created by third parties and may be unsafe or even malicious. Only proceed if you trust the source. You are responsible for anything you install. Do you want to continue?"
            )
            if (!isConfirmed) {
                checkbox.checked = false
                checkbox.blur()
            }
        }
        applyVisibility()
    }

    checkbox.addEventListener('change', () => void onUnofficialChange())
    applyVisibility()
})

window.goToVersions = async (maintainer, app) => {
    const params = new URLSearchParams({ maintainer, app })
    window.location.href = `{{ $.Paths.FrontendVersions }}?${params.toString()}`
}

window.downloadVersionFromAppStore = async (maintainer, app, version) => {
    await downloadFile('{{ $.Paths.BackendStoreVersionsDownload }}', {
        Maintainer: maintainer,
        AppName: app,
        VersionName: version
    })
}
