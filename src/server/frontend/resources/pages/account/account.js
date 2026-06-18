window.updatePassword = async function (currentPassword, newPassword) {
    const ok = await window.apiPost('{{ $.Paths.BackendUsersChangeOwnPassword }}', {current_password: currentPassword, new_password: newPassword})
    if (ok) reloadPageAndShowSnackbar('Password was updated successfully')
}
