window.createOidcClient = async () => {
    const form = document.getElementById('oidc-client-create-form')
    const client = {
        id: "0",
        name: form.clientName.value,
        domain: form.domain.value,
    }

    const ok = await apiPost('{{ $.Paths.BackendOidcRelyingPartiesCreate }}', client)
    if (ok) reloadPageAndShowSnackbar('OIDC client created.')
}

window.updateOidcClient = async (clientRecordId) => {
    const row = document.querySelector(`.oidc-relying-party-row[data-client-record-id="${clientRecordId}"]`)
    const client = {
        id: String(clientRecordId),
        name: row.querySelector('.oidc-client-name-edit').value,
        domain: row.querySelector('.oidc-client-domain-edit').value,
    }

    const ok = await apiPost('{{ $.Paths.BackendOidcRelyingPartiesUpdate }}', client)
    if (ok) reloadPageAndShowSnackbar('OIDC client saved.')
}

window.deleteOidcClient = async (clientRecordId) => {
    const row = document.querySelector(`.oidc-relying-party-row[data-client-record-id="${clientRecordId}"]`)
    const clientName = row.querySelector('.oidc-client-name-edit').value
    const isConfirmed = await confirmDialog(`Delete client '${clientName}'? It will immediately lose access to Quollix as an OIDC provider.`)
    if (!isConfirmed) return

    const ok = await apiPost('{{ $.Paths.BackendOidcRelyingPartiesDelete }}', { value: String(clientRecordId) })
    if (ok) reloadPageAndShowSnackbar('OIDC client deleted.')
}

window.copyToClipboard = async (value, label) => {
    await navigator.clipboard.writeText(value)
    showSnackbar(`${label} copied to clipboard.`)
}

window.regenerateOidcClientCredentials = async (clientRecordId) => {
    const row = document.querySelector(`.oidc-relying-party-row[data-client-record-id="${clientRecordId}"]`)
    const clientName = row.querySelector('.oidc-client-name-edit').value
    const isConfirmed = await confirmDialog(`Regenerate OpenID Connect credentials for '${clientName}'? Existing client credentials will stop working.`)
    if (!isConfirmed) return

    const ok = await apiPost('{{ $.Paths.BackendOidcRelyingPartiesRegenerate }}', { value: String(clientRecordId) })
    if (ok) reloadPageAndShowSnackbar('Credentials regenerated successfully.')
}
