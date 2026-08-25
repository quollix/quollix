document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form')
    const checkbox = document.getElementById('unofficial')
    const maintainerWrap = document.getElementById('maintainer-wrap')

    searchForm.addEventListener('submit', event => {
        if (isValidSearchTerm(document.getElementById('maintainer-input')?.value || '')) {
            if (isValidSearchTerm(document.getElementById('app-input')?.value || '')) return
            showInvalidSearchTermSnackbar('app_name')
        } else {
            showInvalidSearchTermSnackbar('maintainer_name')
        }
        event.preventDefault()
    })

    if (!checkbox || !maintainerWrap) return

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

function isValidSearchTerm(value) {
    return /^[a-z0-9]{0,20}$/.test(value)
}

function showInvalidSearchTermSnackbar(fieldName) {
    showSnackbar(`Invalid input. The content of the field ${fieldName} must be at most 20 characters long. Allowed symbols are: a-z0-9.`)
}

window.goToVersions = async (maintainer, app) => {
    const params = new URLSearchParams({ maintainer, app })
    window.location.href = `{{ $.Static.Paths.FrontendVersions }}?${params.toString()}`
}

window.installAppFromStore = async (maintainer, app, versionId) => {
    const ok = await installApp(versionId)
    if (!ok) return
    disableInstallButtonForApp(maintainer, app)
}

function disableInstallButtonForApp(maintainer, app) {
    const rows = document.querySelectorAll(`#store-results-body tr.store-result-row[data-maintainer="${CSS.escape(maintainer)}"][data-app="${CSS.escape(app)}"]`)
    for (const row of rows) {
        const installButton = row.querySelector("button.store-install-button")
        if (!installButton) continue
        setButtonDisabled(installButton, true)
        installButton.removeAttribute("onclick")
    }
}
