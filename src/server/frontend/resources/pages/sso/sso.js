window.copyToClipboard = async (value, label) => {
    await navigator.clipboard.writeText(value)
    showSnackbar(`${label} copied to clipboard.`)
}

window.regenerateOidcCredentials = async (appId, appLabel) => {
    const isConfirmed = await confirmDialog(`Regenerate OpenID Connect credentials for '${appLabel}'? Existing client credentials will stop working.`)
    if (!isConfirmed) return
    const ok = await apiPost('{{ $.Paths.BackendAppsRegenerateOidcCredentials }}', { value: appId })
    if (ok) reloadPageAndShowSnackbar("Credentials regenerated successfully.")
}
