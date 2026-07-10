window.setPassword = async (event, password) => {
    event.preventDefault()
    const token = new URLSearchParams(location.search).get('token')
    const ok = await apiPost('{{ $.Static.Paths.BackendUsersSetPassword }}', { password, token })
    if (ok) {
        document.getElementById('success-msg').style.display = ''
    }
}
