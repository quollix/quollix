window.regenerateOidcCredentials = async (appId, appLabel) => {
    const isConfirmed = await confirmDialog(`Regenerate OpenID Connect credentials for '${appLabel}'? Existing client credentials will stop working.`)
    if (!isConfirmed) return
    const ok = await apiPost('{{ $.Static.Paths.BackendAppsRegenerateOidcCredentials }}', { value: appId })
    if (ok) reloadPageAndShowSnackbar("Credentials regenerated successfully.")
}
