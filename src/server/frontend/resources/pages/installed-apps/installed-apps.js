window.changeAccessPolicy = async (selectionMenu) => {
  const appId = selectionMenu.closest("tr").dataset.appId
  const nextAccessPolicy = selectionMenu.value
  const prevAccessPolicy = selectionMenu.dataset.prevAccessPolicy

  const warningByPolicy = {
    "{{ $.Policies.PublicAccessPolicy }}": "This will make the app accessible to anyone, including anonymous users. Are you sure?",
    "{{ $.Policies.AuthenticatedAccessPolicy }}": "This will make the app accessible to all authenticated users. Are you sure?",
    "{{ $.Policies.GroupRestrictedAccessPolicy }}": "This will make the app accessible to all users in groups with access to this app. Are you sure?"
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
    const isConfirmed = await confirm(`Are you sure you want to STOP '${appName}'?`)
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
    const isConfirmed = await confirm(`Are you sure you want to DELETE '${appName}'? Backups of this app will not be affected.`)
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
  const dialogWarning = 'Upload this app file only if you trust the maintainer. Third-party apps can contain malicious or unsafe content, steal data, damage existing app data, abuse system resources, or otherwise compromise this device. Do you want to continue?'
  const confirmed = await confirmDialog(dialogWarning)
  if (!confirmed) {
    return
  }
  const ok = await selectAndUploadFile('{{ $.Paths.BackendAppUploadToApplication }}')
  if (ok) reloadPageAndShowSnackbar('Upload successful')
}
