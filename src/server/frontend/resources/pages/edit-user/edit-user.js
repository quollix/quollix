window.changeUsername = async (userId, newUsername) => {
    const confirmed = await confirmDialog(`Change username to '${newUsername}'? OIDC sign-in usually keeps working, but connected apps may cache profile data or run into username collisions.`)
    if (!confirmed) return
    const ok = await apiPost('{{ $.Paths.BackendUsersChangeUsername }}', { user_id: userId, username: newUsername })
    if (ok) reloadPageAndShowSnackbar("Username updated.")
}

window.changeEmail = async (userId, newEmail) => {
    const confirmed = await confirmDialog(`Change email to '${newEmail}'? OIDC sign-in usually keeps working, but connected apps may cache profile data or run into email collisions.`)
    if (!confirmed) return
    const ok = await apiPost('{{ $.Paths.BackendUsersChangeEmail }}', { user_id: userId, new_email: newEmail })
    if (ok) reloadPageAndShowSnackbar("Email updated.")
}
