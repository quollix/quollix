window.saveBaseDomain = async (baseDomain) => {
    const confirmed = await confirmDialog(`Change base domain to '${baseDomain}'? Apps may need a restart; certificates and DNS may need updates.`)
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendSettingsBaseDomainSave }}', { value: baseDomain })
    if (ok) window.reloadPageAndShowSnackbar('Base domain saved successfully.')
}

window.uploadCertificate = async () => {
    const confirmed = await confirmDialog(`Replace TLS certificate? The current certificate will be overwritten.`)
    if (!confirmed) return

    const ok = await selectAndUploadFile('{{ $.Paths.BackendSettingsCertificateUpload }}')
    if (ok) showSnackbar('Certificate uploaded successfully.')
}

window.downloadCertificate = async function() {
    const confirmed = await confirmDialog(`Download certificate with TLS private key? Anyone with this file can impersonate this instance.`)
    if (!confirmed) return

    await window.downloadFile('{{ $.Paths.BackendSettingsCertificateDownload }}', null)
    showSnackbar('Certificate downloaded.')
}

window.resetCertificate = async function() {
    const confirmed = await confirmDialog(`Replace TLS certificate with a self-signed certificate? Browsers and clients may stop trusting this instance.`)
    if (!confirmed) return

    const ok = await window.doNetworkChangedRequest('{{ $.Paths.BackendSettingsCertificateReset }}', {})
    if (ok) showSnackbar('Certificate has been reset to a self-signed certificate.')
}

window.startDnsChallenge = async function() {
    const confirmed = await confirmDialog(`Request a Let's Encrypt certificate? You must be able to create DNS TXT records for the base domain.`)
    if (!confirmed) return

    const response = await window.doRequest('{{ $.Paths.BackendSettingsStartDns01CertificateChallenge }}', null)

    if (!response.ok) {
        const message = await response.text()
        showSnackbar(message || 'Request failed')
        return
    }

    const data = await response.json()
    const recordName = data.record_name
    const wildcardKeyAuth = data.wildcard_key_auth

    const resultContainer = document.getElementById('dns01ChallengeResult')
    const recordNameElement = document.getElementById('dns01ChallengeRecordName')
    const recordValueElement = document.getElementById('dns01ChallengeRecordValue')

    recordNameElement.value = recordName || ""
    recordValueElement.value = wildcardKeyAuth || ""
    resultContainer.style.display = 'block'
}

window.copyDns01NameToClipboard = function() {
    return window.copyTextareaToClipboard('dns01ChallengeRecordName')
}

window.copyDns01ValueToClipboard = function() {
    return window.copyTextareaToClipboard('dns01ChallengeRecordValue')
}

window.copyTextareaToClipboard = async function(textareaId) {
    const textareaElement = document.getElementById(textareaId)
    if (!textareaElement) {
        showSnackbar('Copy failed. Please copy manually.')
        console.error(`element with id "${textareaId}" not found.`)
        return
    }

    const textToCopy = textareaElement.value || ""

    try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            await navigator.clipboard.writeText(textToCopy)
            showSnackbar('Copied to clipboard.')
            return
        }

        // Fallback for environments where Clipboard API is unavailable (e.g. non-HTTPS, embedded webviews, older browsers).
        // document.execCommand('copy') is deprecated but still the only widely-supported option outside secure contexts.

        textareaElement.focus()
        textareaElement.select()
        const isSuccessful = document.execCommand('copy')
        showSnackbar(isSuccessful ? 'Copied to clipboard.' : 'Copy failed. Please copy manually.')
    } catch (error) {
        showSnackbar('Copy failed. Please copy manually.')
        console.error(error)
    }
}

window.getKnownHosts = async function(backupServerHost, port) {
    const backupServerForm = document.getElementById('backupServerForm')
    const backupServerKnownHosts = backupServerForm?.elements?.backupServerKnownHosts

    const [response, ok] = await window.doNetworkChangedRequestWithBody('{{ $.Paths.BackendSettingsGetSshKnownHosts }}', {
        host: backupServerHost,
        port: String(port),
    })

    if (ok) {
        const knownHosts = response?.value
        backupServerKnownHosts.value = knownHosts || ''
    }
}

window.testConnection = async function(
    host,
    port,
    user,
    password,
    knownHosts
) {
    const sshConnectionTestRequest = {
        host: host,
        port: port,
        user: user,
        password: password,
        known_hosts: knownHosts,
    }

    const ok = await window.doNetworkChangedRequest('{{ $.Paths.BackendSettingsSshTestAccess }}', sshConnectionTestRequest)
    if (ok) showSnackbar('Connection test successful.')
}

window.save = async function(
    host,
    port,
    user,
    password,
    knownHosts,
    encryptionPassword,
    isEnabled
) {
    const backupServer = {
        is_enabled: isEnabled,
        host: host,
        port: port,
        user: user,
        password: password,
        known_hosts: knownHosts,
        encryption_password: encryptionPassword,
    }

    const ok = await window.doNetworkChangedRequest('{{ $.Paths.BackendSettingsSshSave }}', backupServer)
    if (ok) window.showSnackbar('Backup server settings saved.')
}

window.saveMaintenanceConfig = async function(maintenanceWindowStartHourString, ianaTimezone) {
    const ok = await apiPost('{{ $.Paths.BackendMaintenanceConfigsSave }}', {
        iana_timezone: ianaTimezone,
        maintenance_window_start_hour: Number(maintenanceWindowStartHourString),
    })
    if (ok) reloadPageAndShowSnackbar("Maintenance settings saved.")
}

window.saveRetentionPolicy = async function() {
    const confirmed = await confirmDialog(
        `Save retention policy? Backups outside the new policy may be deleted during the next maintenance run.`
    )
    if (!confirmed) return

    const keepPreUpdateString = document.getElementById('retention-keep-pre-update')?.value
    const keepDailyString = document.getElementById('retention-keep-daily')?.value
    const keepWeeklyString = document.getElementById('retention-keep-weekly')?.value
    const keepMonthlyString = document.getElementById('retention-keep-monthly')?.value
    const keepYearlyString = document.getElementById('retention-keep-yearly')?.value

    const ok = await apiPost('{{ $.Paths.BackendMaintenanceRetentionPolicySave }}', {
        keep_pre_update: Number(keepPreUpdateString),
        keep_daily: Number(keepDailyString),
        keep_weekly: Number(keepWeeklyString),
        keep_monthly: Number(keepMonthlyString),
        keep_yearly: Number(keepYearlyString),
    })

    if (ok) showSnackbar('Retention policy saved.')
}


window.resetBackupServerConfigs = async function() {
    const confirmed = await confirmDialog(
        `Reset backup server settings? Existing backups may become inaccessible.`
    )
    if (!confirmed) return

    const ok = await window.apiPost('{{ $.Paths.BackendSettingsSshConfigsReset }}', null)
    if (ok) window.reloadPageAndShowSnackbar('Backup server settings have been reset.')
}

window.runMaintenanceJobsNow = async function() {
    const confirmed = await confirmDialog(
        `Run maintenance now? Apps may be temporarily unavailable.`
    )
    if (!confirmed) return

    const ok = await window.doNetworkChangedRequest('{{ $.Paths.BackendMaintenanceTriggerMaintenanceJob }}', {})
    if (ok) showSnackbar('Maintenance jobs started.')
}

window.startCertificateOperationIndicatorPolling = function() {
    const containerElement = document.getElementById('certificate-operation-indicator')
    const spinnerElement = document.getElementById('certificate-operation-spinner')
    const iconElement = document.getElementById('certificate-operation-icon')
    const textElement = document.getElementById('certificate-operation-text')
    if (!containerElement || !spinnerElement || !iconElement || !textElement) return

    const render = (state, operationText) => {
        const normalizedState = (state || '').toLowerCase()
        const text = operationText || ''

        if (normalizedState === 'running') {
            containerElement.style.display = 'flex'
            spinnerElement.style.display = 'inline-block'
            iconElement.style.display = 'none'
            iconElement.className = 'mdi'
            textElement.textContent = text
            return
        }

        if (normalizedState === 'success') {
            containerElement.style.display = 'flex'
            spinnerElement.style.display = 'none'
            iconElement.style.display = 'inline-block'
            iconElement.className = 'mdi mdi-check-circle-outline'
            textElement.textContent = text
            return
        }

        if (normalizedState === 'error') {
            containerElement.style.display = 'flex'
            spinnerElement.style.display = 'none'
            iconElement.style.display = 'inline-block'
            iconElement.className = 'mdi mdi-close-circle-outline'
            textElement.textContent = text
            return
        }

        containerElement.style.display = 'none'
        spinnerElement.style.display = 'none'
        iconElement.style.display = 'none'
        iconElement.className = 'mdi'
        textElement.textContent = ''
    }

    const pollOnce = async () => {
        try {
            const response = await window.doRequest('{{$.Paths.BackendSettingsCertificateOperationStatus}}')
            if (!response.ok) {
                render('idle', '')
            } else {
                const data = await response.json()
                render(data.state, data.current_operation)
            }
        } catch (error) {
            console.log(error)
            render('idle', '')
        }

        window.setTimeout(pollOnce, 1000)
    }

    pollOnce()
}

window.addEventListener('load', () => {
    window.startCertificateOperationIndicatorPolling()
})

window.purgeBackupServer = async function(
    host,
    port,
    user,
    password,
    knownHosts,
) {
    const confirmed = await confirmDialog(
        `Purge backup repository? All backups in the repository will be permanently deleted.`
    )
    if (!confirmed) return

    const backupServer = {
        host: host,
        port: port,
        user: user,
        password: password,
        known_hosts: knownHosts,
    }

    const ok = await window.apiPost('{{ $.Paths.BackendBackupsPurgeBackupServer }}', backupServer)
    if (ok) window.showSnackbar('Backup server has been purged.')
}
