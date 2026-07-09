window.signIn = async (event, username, password) => {
    event.preventDefault()
    const ok = await apiPost('{{ $.Paths.BackendSignIn }}', { username, password })
    if (ok) {
        // If sign-in was triggered by an OIDC /authorize request, "next" contains the original authorize URL. Redirect back to it so the OIDC flow continues.
        const nextUrl = new URLSearchParams(window.location.search).get('next')

        // Fallback to normal app landing page for non-OIDC sign-ins.
        window.location.href = nextUrl || '{{ $.Paths.FrontendInstalledApps }}'
    }
}

window.startOidcSignIn = async (providerId) => {
    const [body, ok] = await apiPostWithBody('{{ $.Paths.BackendOidcSignIn }}', { value: providerId })
    if (!ok || !body?.redirect_url) return
    window.location.href = body.redirect_url
}
