window.createUser = async (username, email) => {
  const ok = await apiPost('{{ $.Paths.BackendUsersInviteUser }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('User invited successfully.')
}

window.createUserViaEmail = async (username, email) => {
  const ok = await apiPost('{{ $.Paths.BackendUsersInviteUserViaEmail }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('Invitation email sent successfully.')
}

window.deleteUser = async (userId, username) => {
  const confirmed = await confirmDialog(`Delete user '${username}'?`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Paths.BackendUsersDelete }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("User deleted successfully.")
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
