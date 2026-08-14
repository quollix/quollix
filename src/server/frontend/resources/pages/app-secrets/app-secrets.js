window.regenerateAppSecret = async (appId, name) => {
    const isConfirmed = await confirmDialog(`Regenerate '${name}'? This can break deployments. The app must be restarted before the new value applies.`)
    if (!isConfirmed) return

    const ok = await apiPost('{{ $.Static.Paths.BackendAppSecretRegenerate }}', {
        app_id: String(appId),
        name: name,
    })
    if (ok) reloadPageAndShowSnackbar('Secret regenerated successfully.')
}

window.updateAppSecret = async (appId, name) => {
    const row = document.querySelector(`.app-secret-row[data-secret-name="${name}"]`)
    const ok = await apiPost('{{ $.Static.Paths.BackendAppSecretUpdate }}', {
        app_id: String(appId),
        name: name,
        value: row.querySelector('.app-secret-value-edit').value,
    })
    if (ok) showSnackbar('Secret saved.')
}

window.deleteUnusedAppSecret = async (appId, name) => {
    const isConfirmed = await confirmDialog(`Delete unused secret '${name}'?`)
    if (!isConfirmed) return

    const ok = await apiPost('/api/apps/secrets/delete', {
        app_id: String(appId),
        name: name,
    })
    if (ok) reloadPageAndShowSnackbar('Secret deleted successfully.')
}

window.copyAppSecretValue = async (name) => {
    const row = document.querySelector(`.app-secret-row[data-secret-name="${name}"]`)
    await copyToClipboard(row.querySelector('.app-secret-value-edit').value, name)
}
