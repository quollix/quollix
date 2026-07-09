window.createUser = async (username, email) => {
  const ok = await apiPost('{{ $.Paths.BackendUsersInviteUser }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('User invited successfully.')
}

window.createUserViaEmail = async (username, email) => {
  const ok = await apiPost('{{ $.Paths.BackendUsersInviteUserViaEmail }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('Invitation email sent successfully.')
}

window.deleteUser = async (userId, username) => {
  const confirmed = await confirmDialog(`Delete user '${username}'? The account will lose access to Quollix.`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Paths.BackendUsersDelete }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("User deleted successfully.")
}

window.setUserEnabled = async (userId, username, checkbox) => {
  const isEnabled = checkbox.checked
  if (!isEnabled) {
    const confirmed = await confirmDialog(`Disable user '${username}'? Existing sessions will be removed and sign-in attempts will be denied.`)
    if (!confirmed) {
      checkbox.checked = true
      return
    }
  }
  const ok = await apiPost('{{ $.Paths.BackendUsersSetEnabled }}', { user_id: userId, is_enabled: isEnabled })
  if (ok) {
    return
  }
  checkbox.checked = !isEnabled
}

window.copyToClipboard = async (link, host) => {
  navigator.clipboard.writeText(link);
  showSnackbar("Link copied to clipboard.")
}

window.resetPassword = async (userId, username) => {
  const confirmed = await confirmDialog(`Reset password of user '${username}'?`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Paths.BackendUsersResetPassword }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("Password reset successful")
}

window.sendPasswordResetEmail = async (userId, username) => {
  const confirmed = await confirmDialog(`Send password reset email to user '${username}'?`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Paths.BackendUsersResetPasswordViaEmail }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("Password reset email sent successfully.")
}
