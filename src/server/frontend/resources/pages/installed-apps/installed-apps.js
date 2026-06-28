window.changeAccessPolicy = async (selectionMenu) => {
    const appId = selectionMenu.closest("tr").dataset.appId
    const nextAccessPolicy = selectionMenu.value
    const prevAccessPolicy = selectionMenu.dataset.prevAccessPolicy

    const warningByPolicy = {
        "{{ $.Policies.PublicAccessPolicy }}": "Make this app public? Anyone, including anonymous users, can access it.",
        "{{ $.Policies.AuthenticatedAccessPolicy }}": "Allow all signed-in users? Every authenticated user can access this app.",
        "{{ $.Policies.GroupRestrictedAccessPolicy }}": "Restrict by group? Only users in assigned groups can access this app."
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

    const ok = await apiPost('{{ $.Paths.BackendAppsChangeAccessPolicy }}', {
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
        const res = await window.doRequest("{{ $.Paths.BackendSecret }}", null)
        const body = await res.text()
        if (!res.ok) {
            showSnackbar(body || 'Request failed')
            return
        }
        const secret = JSON.parse(body)
        window.open(`${base}?quollix-secret=${secret}`, '_blank')
    }
}

window.handleOperationsSelectChange = async (selectionMenu, appId, appName) => {
    const op = selectionMenu.value
    selectionMenu.value = ''

    const confirm = (msg) => window.confirmDialog(msg)

    if (op === 'start') {
        await doNetworkChangedRequest('{{ $.Paths.BackendAppsStart }}', { value: appId })
        return
    }
    if (op === 'stop') {
        const isConfirmed = await confirm(`Stop '${appName}'? Users will lose access until it is started again.`)
        if (!isConfirmed) return
        await doNetworkChangedRequest('{{ $.Paths.BackendAppsStop }}', { value: appId })
        return
    }
    if (op === 'download') {
        await window.downloadFile('{{ $.Paths.BackendAppDownloadFromApplication }}', { value: appId })
        return
    }
    if (op === 'update') {
        await doNetworkChangedRequest('{{ $.Paths.BackendAppsUpdate }}', { value: appId })
        return
    }
    if (op === 'backup') {
        await doNetworkChangedRequest('{{ $.Paths.BackendBackupsCreate }}', { value: appId })
        return
    }
    if (op === 'delete') {
        const isConfirmed = await confirm(`Delete '${appName}'? App data will be removed; backups are kept.`)
        if (!isConfirmed) return
        await doNetworkChangedRequest('{{ $.Paths.BackendAppsDelete }}', { value: appId })
        return
    }
}

window.reloadAppsIntoStoreMock = async () => {
    const ok = await apiPost('{{ $.Paths.BackendStoreReloadPublishedApps }}', null)
    if (ok) reloadPageAndShowSnackbar('Local store apps reloaded successfully')
}

window.uploadVersionFile = async () => {
    const dialogWarning = 'Upload third-party app? Only continue if you trust the maintainer.'
    const confirmed = await confirmDialog(dialogWarning)
    if (!confirmed) {
        return
    }
    const ok = await selectAndUploadFile('{{ $.Paths.BackendAppUploadToApplication }}')
    if (ok) reloadPageAndShowSnackbar('Upload successful')
}
