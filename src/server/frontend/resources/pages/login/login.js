window.login = async (event, username, password) => {
    event.preventDefault()
    const ok = await apiPost('{{ $.Paths.BackendLogin }}', { username, password })
    if (ok) {
        // If login was triggered by an OIDC /authorize request, "next" contains the original authorize URL. Redirect back to it so the OIDC flow continues.
        const nextUrl = new URLSearchParams(window.location.search).get('next')

        // Fallback to normal app landing page for non-OIDC logins.
        window.location.href = nextUrl || '{{ $.Paths.FrontendInstalledApps }}'
    }
}