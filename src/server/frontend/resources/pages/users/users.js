window.createUser = async (username, email) => {
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersInviteUser }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('User invited successfully.')
}

window.createUserViaEmail = async (username, email) => {
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersInviteUserViaEmail }}', { username: username, email: email })
  if (ok) reloadPageAndShowSnackbar('Invitation email sent successfully.')
}

window.updateInviteViaEmailButtonState = () => {
  const emailInput = document.getElementById('create-user-email-input')
  const inviteViaEmailButton = document.getElementById('create-user-via-email-button')
  if (!emailInput || !inviteViaEmailButton) return
  inviteViaEmailButton.disabled = emailInput.value === '' || !emailInput.checkValidity()
}

window.deleteUser = async (userId, username) => {
  const confirmed = await confirmDialog(`Delete user '${username}'? The account will lose access to Quollix.`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersDelete }}', { value: userId })
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
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersSetEnabled }}', { user_id: userId, is_enabled: isEnabled })
  if (ok) {
    reloadPageAndShowSnackbar(`User ${isEnabled ? 'enabled' : 'disabled'} successfully.`)
    return
  }
  checkbox.checked = !isEnabled
}

window.setUserRole = async (userId, username, select) => {
  const previousRole = select.dataset.prevRole
  const nextRole = select.value
  const isAdmin = select.value === 'admin'
  const action = isAdmin ? 'promote' : 'demote'
  const role = isAdmin ? 'admin' : 'regular user'
  const confirmed = await confirmDialog(`Are you sure you want to ${action} '${username}' to ${role}?`)
  if (!confirmed) {
    select.value = previousRole
    return
  }
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersSetRole }}', { user_id: userId, is_admin: isAdmin })
  if (ok) {
    select.dataset.prevRole = nextRole
    reloadPageAndShowSnackbar(`'${username}' role updated.`)
    return
  }
  select.value = previousRole
}

window.copyToClipboard = async (link, host) => {
  navigator.clipboard.writeText(link);
  showSnackbar("Link copied to clipboard.")
}

window.resetPassword = async (userId, username) => {
  const confirmed = await confirmDialog(`Reset password of user '${username}'?`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersResetPassword }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("Password reset successful")
}

window.sendPasswordResetEmail = async (userId, username) => {
  const confirmed = await confirmDialog(`Send password reset email to user '${username}'?`)
  if (!confirmed) return
  const ok = await apiPost('{{ $.Static.Paths.BackendUsersResetPasswordViaEmail }}', { value: userId })
  if (ok) reloadPageAndShowSnackbar("Password reset email sent successfully.")
}
