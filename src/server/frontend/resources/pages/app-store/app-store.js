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
                    "Show unofficial apps? Only install apps from maintainers you trust."
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
    window.location.href = `{{ $.Static.Paths.FrontendVersions }}?${params.toString()}`
}

window.installAppFromStore = async (maintainer, app, version) => {
    const ok = await installApp(maintainer, app, version)
    if (!ok) return
    disableInstallButtonsForApp(app)
}

function disableInstallButtonsForApp(app) {
    const rows = document.querySelectorAll(`#store-results-body tr.store-result-row[data-app="${CSS.escape(app)}"]`)
    for (const row of rows) {
        const installButton = row.querySelector("button.store-install-button")
        if (!installButton) continue
        installButton.disabled = true
        installButton.setAttribute("aria-disabled", "true")
        installButton.setAttribute("title", "Already installed")
        installButton.setAttribute("aria-label", "Already installed")
        installButton.removeAttribute("onclick")
    }
}

window.downloadVersionFromAppStore = async (maintainer, app, version) => {
    await downloadFile('{{ $.Static.Paths.BackendStoreVersionsDownload }}', {
        Maintainer: maintainer,
        AppName: app,
        VersionName: version
    })
}
