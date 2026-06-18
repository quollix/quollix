window.updateAutoMaintenanceSettings = async (checkbox, appId, isOfficial) => {
    const row = checkbox.closest("tr")
    const automaticUpdatesCheckbox = row.querySelector(".auto-update-cell input")
    const automaticBackupsCheckbox = row.querySelector(".auto-backup-cell input")

    const isAutomaticUpdatesCheckbox = checkbox === automaticUpdatesCheckbox
    const isEnablingAutomaticUpdates = isAutomaticUpdatesCheckbox && automaticUpdatesCheckbox.checked

    if (isEnablingAutomaticUpdates && !isOfficial) {
        const message = "Updates will be downloaded and installed automatically from the app store. Since this app is maintained by a third party, enabling automatic updates may introduce untrusted code to your system. Only enable this if you trust the maintainer. Do you want to continue?"
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
