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
        `This will reset all email settings. Email credentials may be removed and notifications may stop working until reconfigured. Do you want to continue?`
    )
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendEmailResetConfig }}', {})
    if (ok) window.reloadPageAndShowSnackbar('Email settings have been reset.')
}

window.saveInvitationEmailTemplate = async function (template) {
    const ok = await window.apiPost('{{ $.Paths.BackendEmailSaveInvitationTemplate }}', { template })
    if (ok) showSnackbar('Invitation email template saved.')
}

window.resetInvitationEmailTemplate = async function () {
    const confirmed = await confirmDialog(
        `This will reset the invitation email template to the default text. Do you want to continue?`
    )
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendEmailResetInvitationTemplate }}', null)
    if (ok) window.reloadPageAndShowSnackbar('Invitation email template has been reset.')
}
