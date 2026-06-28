window.testEmailConnection = async function (
    smtpHost,
    smtpPort,
    fromEmailAddress,
    emailAccountUsername,
    emailAccountPassword,
    isEnabled
) {
    const emailConfig = {
        smtp_host: smtpHost,
        smtp_port: String(smtpPort),
        from_email_address: fromEmailAddress,
        email_account_username: emailAccountUsername,
        email_account_password: emailAccountPassword,
        is_enabled: isEnabled,
    }

    const ok = await window.apiPost('{{ $.Paths.BackendEmailTestConnection }}', emailConfig)
    if (ok) showSnackbar('Email connection test successful.')
}

window.saveEmailConfig = async function (
    smtpHost,
    smtpPort,
    fromEmailAddress,
    emailAccountUsername,
    emailAccountPassword,
    isEnabled
) {
    const emailConfig = {
        smtp_host: smtpHost,
        smtp_port: String(smtpPort),
        from_email_address: fromEmailAddress,
        email_account_username: emailAccountUsername,
        email_account_password: emailAccountPassword,
        is_enabled: isEnabled,
    }

    const ok = await window.apiPost('{{ $.Paths.BackendEmailSaveConfig }}', emailConfig)
    if (ok) showSnackbar('Email settings saved.')
}

window.sendTestEmail = async function (
    smtpHost,
    smtpPort,
    fromEmailAddress,
    emailAccountUsername,
    emailAccountPassword,
    toEmail
) {
    const requestBody = {
        email_config: {
            smtp_host: smtpHost,
            smtp_port: String(smtpPort),
            from_email_address: fromEmailAddress,
            email_account_username: emailAccountUsername,
            email_account_password: emailAccountPassword,
            is_enabled: true,
        },
        to_email: (toEmail || '').trim(),
    }

    const ok = await window.apiPost('{{ $.Paths.BackendEmailSendTestEmail }}', requestBody)
    if (ok) showSnackbar('Test email sent.')
}

window.resetEmailConfigs = async function () {
    const confirmed = await confirmDialog(
        `Reset email settings? Email notifications may stop working until reconfigured.`
    )
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendEmailResetConfig }}', {})
    if (ok) window.reloadPageAndShowSnackbar('Email settings have been reset.')
}

window.saveOidcEmailExposure = async function (checkbox) {
    const exposeRealEmail = checkbox.checked
    if (exposeRealEmail) {
        const confirmed = await confirmDialog(
            `Expose real email addresses to apps? Connected apps may receive, store, or show users' real email addresses.`
        )
        if (!confirmed) {
            checkbox.checked = false
            return
        }
    }

    const ok = await window.apiPost('{{ $.Paths.BackendEmailSaveOidcEmailExposure }}', {
        value: exposeRealEmail,
    })
    if (ok) {
        showSnackbar('OIDC email exposure setting saved.')
        return
    }
    checkbox.checked = !exposeRealEmail
}

window.saveInvitationEmailTemplate = async function (template) {
    const ok = await window.apiPost('{{ $.Paths.BackendEmailSaveInvitationTemplate }}', { template })
    if (ok) showSnackbar('Invitation email template saved.')
}

window.resetInvitationEmailTemplate = async function () {
    const confirmed = await confirmDialog(
        `Reset invitation email template? The default text will replace the current template.`
    )
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendEmailResetInvitationTemplate }}', null)
    if (ok) window.reloadPageAndShowSnackbar('Invitation email template has been reset.')
}
