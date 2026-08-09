window.regenerateAppSecret = async (appId, name) => {
    const isConfirmed = await confirmDialog(`Regenerate '${name}'? This can break deployments. The app must be restarted before the new value applies.`)
    if (!isConfirmed) return

    const ok = await apiPost('{{ $.Static.Paths.BackendAppSecretRegenerate }}', {
        app_id: String(appId),
        name: name,
    })
    if (ok) reloadPageAndShowSnackbar('Secret regenerated successfully.')
}
