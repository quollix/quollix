window.changeAccessPolicy = async (selectionMenu) => {
    const appId = selectionMenu.closest("tr").dataset.appId
    const nextAccessPolicy = selectionMenu.value
    const prevAccessPolicy = selectionMenu.dataset.prevAccessPolicy

    const warningByPolicy = {
        "{{ $.Static.Policies.PublicAccessPolicy }}": "Make this app public? Anyone, including anonymous users, can access it.",
        "{{ $.Static.Policies.AuthenticatedAccessPolicy }}": "Allow all signed-in users? Every authenticated user can access this app.",
        "{{ $.Static.Policies.GroupRestrictedAccessPolicy }}": "Restrict by group? Only users in assigned groups can access this app."
        // admin_only intentionally has no warning as it is the strictest/safest policy
    }

    const warning = warningByPolicy[nextAccessPolicy]
    if (warning) {
        const confirmed = await confirmDialog(warning)
        if (!confirmed) {
            selectionMenu.value = prevAccessPolicy
            return
        }
    }

    const ok = await apiPost('{{ $.Static.Paths.BackendAppsChangeAccessPolicy }}', {
        app_id: appId,
        access_policy: nextAccessPolicy
    })

    if (ok) {
        showSnackbar("Access policy changed successfully.")
        selectionMenu.dataset.prevAccessPolicy = nextAccessPolicy
    } else {
        selectionMenu.value = prevAccessPolicy
    }
}


window.openFromRow = async (appName, appAccessPolicy, publicAccessPolicy, host) => {
    const isPublicAccessPolicy = appAccessPolicy === publicAccessPolicy

    const base = `${location.protocol}//${appName}.${host}/`
    if (isPublicAccessPolicy) {
        window.open(base, '_blank')
    } else {
        const res = await window.doRequest("{{ $.Static.Paths.BackendSecret }}", null)
        const body = await res.text()
        if (!res.ok) {
            showSnackbar(body || 'Request failed')
            return
        }
        const secret = JSON.parse(body)
        window.open(`${base}?quollix-secret=${secret}`, '_blank')
    }
}

function isAppRowRunning(row) {
    const openButton = row?.querySelector(".open-btn")
    return openButton ? !openButton.disabled : false
}

function setAppRowRunningState(row, isRunning) {
    if (!row) return

    const statusCell = row.querySelector(".app-status")
    if (statusCell) {
        statusCell.textContent = isRunning ? "Running" : "Not running"
    }

    const openButton = row.querySelector(".open-btn")
    if (!openButton) return

    openButton.disabled = !isRunning
    openButton.toggleAttribute("aria-disabled", !isRunning)
}

function removeAppRow(row) {
    if (!row) return () => {}

    const parent = row.parentNode
    const nextSibling = row.nextSibling
    row.remove()

    return () => {
        if (!parent) return
        if (nextSibling?.parentNode === parent) {
            parent.insertBefore(row, nextSibling)
        } else {
            parent.appendChild(row)
        }
    }
}

async function doOptimisticAppRequest(path, appId, applyChange) {
    const restore = applyChange()
    const ok = await doNetworkChangedRequest(path, { value: appId })
    if (!ok) restore()
}

window.handleOperationsSelectChange = async (selectionMenu, appId, appName) => {
    const op = selectionMenu.value
    const row = selectionMenu.closest("tr")
    selectionMenu.value = ''

    if (op === 'start') {
        await doOptimisticAppRequest('{{ $.Static.Paths.BackendAppsStart }}', appId, () => {
            const wasRunning = isAppRowRunning(row)
            setAppRowRunningState(row, true)
            return () => setAppRowRunningState(row, wasRunning)
        })
        return
    }
    if (op === 'stop') {
        const isConfirmed = await window.confirmDialog(`Stop '${appName}'? Users will lose access until it is started again.`)
        if (!isConfirmed) return
        await doOptimisticAppRequest('{{ $.Static.Paths.BackendAppsStop }}', appId, () => {
            const wasRunning = isAppRowRunning(row)
            setAppRowRunningState(row, false)
            return () => setAppRowRunningState(row, wasRunning)
        })
        return
    }
    if (op === 'download') {
        await window.downloadFile('{{ $.Static.Paths.BackendAppDownloadFromApplication }}', { value: appId })
        return
    }
    if (op === 'update') {
        await doNetworkChangedRequest('{{ $.Static.Paths.BackendAppsUpdate }}', { value: appId })
        return
    }
    if (op === 'backup') {
        await doNetworkChangedRequest('{{ $.Static.Paths.BackendBackupsCreate }}', { value: appId })
        return
    }
    if (op === 'delete') {
        const isConfirmed = await window.confirmDialog(`Delete '${appName}'? App data will be removed; backups are kept.`)
        if (!isConfirmed) return
        await doOptimisticAppRequest('{{ $.Static.Paths.BackendAppsDelete }}', appId, () => removeAppRow(row))
        return
    }
}

window.reloadAppsIntoStoreMock = async () => {
    const ok = await apiPost('{{ $.Static.Paths.BackendStoreReloadPublishedApps }}', null)
    if (ok) reloadPageAndShowSnackbar('Local store apps reloaded successfully')
}

window.uploadVersionFile = async () => {
    const dialogWarning = 'Upload third-party app? Only continue if you trust the maintainer.'
    const confirmed = await confirmDialog(dialogWarning)
    if (!confirmed) {
        return
    }
    const ok = await selectAndUploadFile('{{ $.Static.Paths.BackendAppUploadToApplication }}')
    if (ok) reloadPageAndShowSnackbar('Upload successful')
}
