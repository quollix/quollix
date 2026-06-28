window.updateAutoMaintenanceSettings = async (checkbox, appId, isOfficial) => {
    const row = checkbox.closest("tr")
    const automaticUpdatesCheckbox = row.querySelector(".auto-update-cell input")
    const automaticBackupsCheckbox = row.querySelector(".auto-backup-cell input")

    const isAutomaticUpdatesCheckbox = checkbox === automaticUpdatesCheckbox
    const isEnablingAutomaticUpdates = isAutomaticUpdatesCheckbox && automaticUpdatesCheckbox.checked

    if (isEnablingAutomaticUpdates && !isOfficial) {
        const message = "Enable automatic updates? Future third-party app versions will be installed without review."
        const confirmed = await window.confirmDialog(message)
        if (!confirmed) {
            automaticUpdatesCheckbox.checked = false
            return
        }
    }

    await apiPost('{{$.Paths.BackendAppAutomaticMaintenanceSettings}}', {
        app_id: appId,
        automatic_updates_enabled: automaticUpdatesCheckbox.checked,
        automatic_backups_enabled: automaticBackupsCheckbox.checked
    })
}
