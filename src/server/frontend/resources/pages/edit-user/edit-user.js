window.changeUsername = async (userId, newUsername) => {
    const confirmed = await confirmDialog(`Change username to '${newUsername}'? This may cause inconsistencies in connected OpenID client apps using the old username.`)
    if (!confirmed) return
    const ok = await apiPost('{{ $.Paths.BackendUsersChangeUsername }}', { user_id: userId, username: newUsername })
    if (ok) reloadPageAndShowSnackbar("Username updated.")
}

window.changeEmail = async (userId, newEmail) => {
    const confirmed = await confirmDialog(`Change email to '${newEmail}'? Connected OpenID client apps may still reference the previous email.`)
    if (!confirmed) return
    const ok = await apiPost('{{ $.Paths.BackendUsersChangeEmail }}', { user_id: userId, new_email: newEmail })
    if (ok) reloadPageAndShowSnackbar("Email updated.")
}
