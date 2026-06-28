window.updatePassword = async function (currentPassword, newPassword, confirmPassword) {
    if (newPassword !== confirmPassword) {
        showSnackbar('Passwords do not match')
        return
    }

    const ok = await window.apiPost('{{ $.Paths.BackendUsersChangeOwnPassword }}', {current_password: currentPassword, new_password: newPassword})
    if (ok) reloadPageAndShowSnackbar('Password was updated successfully')
}

window.setPassword = async function (newPassword, confirmPassword) {
    if (newPassword !== confirmPassword) {
        showSnackbar('Passwords do not match')
        return
    }

    const ok = await window.apiPost('{{ $.Paths.BackendUsersSetOwnPassword }}', {new_password: newPassword})
    if (ok) reloadPageAndShowSnackbar('Password was set successfully')
}
