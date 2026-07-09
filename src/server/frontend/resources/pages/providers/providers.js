window.createOidcAuthProvider = async () => {
    const form = document.getElementById('oidc-provider-create-form')
    const provider = {
        name: form.providerName.value,
        issuer_domain_path: form.issuerDomainPath.value,
        client_id: form.clientId.value,
        client_secret: form.clientSecret.value,
    }

    const ok = await apiPost('{{ $.Paths.BackendOidcAuthProvidersCreate }}', provider)
    if (ok) reloadPageAndShowSnackbar('Provider created.')
}

window.updateOidcAuthProvider = async (providerId) => {
    const row = document.querySelector(`.oidc-auth-provider-row[data-provider-id="${providerId}"]`)
    const provider = {
        id: Number(providerId),
        name: row.querySelector('.oidc-provider-name-edit').value,
        issuer_domain_path: row.querySelector('.oidc-provider-issuer-domain-path-edit').value,
        client_id: row.querySelector('.oidc-provider-client-id-edit').value,
        client_secret: row.querySelector('.oidc-provider-client-secret-edit').value,
    }

    const ok = await apiPost('{{ $.Paths.BackendOidcAuthProvidersUpdate }}', provider)
    if (ok) reloadPageAndShowSnackbar('Provider saved.')
}

window.testOidcAuthProviderDiscovery = async (button) => {
    const container = button.closest('.label-bar, td')
    const issuerInput = container.querySelector('[name="issuerDomainPath"], .oidc-provider-issuer-domain-path-edit')
    const ok = await apiPost('{{ $.Paths.BackendOidcAuthProvidersTestDiscovery }}', { issuer_domain_path: issuerInput.value })
    if (ok) showSnackbar('Discovery endpoint is available and seems valid.')
}

window.deleteOidcAuthProvider = async (providerId) => {
    const row = document.querySelector(`.oidc-auth-provider-row[data-provider-id="${providerId}"]`)
    const providerName = row.querySelector('.oidc-provider-name-edit').value
    const isConfirmed = await confirmDialog(`Delete provider '${providerName}'? Linked sign-in methods will be removed.`)
    if (!isConfirmed) return

    const ok = await apiPost('{{ $.Paths.BackendOidcAuthProvidersDelete }}', { value: String(providerId) })
    if (ok) reloadPageAndShowSnackbar('Provider deleted.')
}
