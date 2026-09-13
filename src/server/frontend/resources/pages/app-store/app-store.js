document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form')
    const checkbox = document.getElementById('unofficial')
    const maintainerWrap = document.getElementById('maintainer-wrap')
    const maintainerInput = document.querySelector('[name="{{ $.Static.QueryParams.Store.MaintainerName }}"]')
    const appInput = document.querySelector('[name="{{ $.Static.QueryParams.Store.AppName }}"]')

    searchForm.addEventListener('submit', event => {
        if (isValidSearchTerm(maintainerInput?.value || '')) {
            if (isValidSearchTerm(appInput?.value || '')) return
            showInvalidSearchTermSnackbar('{{ $.Static.QueryParams.Store.AppName }}')
        } else {
            showInvalidSearchTermSnackbar('{{ $.Static.QueryParams.Store.MaintainerName }}')
        }
        event.preventDefault()
    })

    if (!checkbox || !maintainerWrap) return

    const maintainerLabel = maintainerWrap.querySelector('label')

    function applyVisibility() {
        const show = checkbox.checked
        maintainerLabel.style.visibility = show ? 'visible' : 'hidden'
        maintainerInput.style.visibility = show ? 'visible' : 'hidden'
        maintainerInput.disabled = !show
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
    const restore = disableInstallButtonForApp(maintainer, app)
    const ok = await installApp(versionId)
    if (!ok) restore()
}

function disableInstallButtonForApp(maintainer, app) {
    const changedButtons = []
    const rows = document.querySelectorAll(`#store-results-body tr.store-result-row[data-maintainer="${CSS.escape(maintainer)}"][data-app="${CSS.escape(app)}"]`)
    for (const row of rows) {
        const installButton = row.querySelector("button.store-install-button")
        if (!installButton) continue
        changedButtons.push({
            button: installButton,
            wasDisabled: installButton.disabled,
            onclick: installButton.getAttribute("onclick"),
        })
        setButtonDisabled(installButton, true)
        installButton.removeAttribute("onclick")
    }
    return () => {
        for (const changedButton of changedButtons) {
            setButtonDisabled(changedButton.button, changedButton.wasDisabled)
            if (changedButton.onclick) changedButton.button.setAttribute("onclick", changedButton.onclick)
        }
    }
}
