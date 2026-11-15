window.saveAlertSettings = async function(form) {
    const settings = {
        cpu_usage_enabled: form.elements.cpu_enabled.checked,
        memory_usage_enabled: form.elements.memory_enabled.checked,
        disk_usage_enabled: form.elements.disk_enabled.checked,
        certificate_expiry_enabled: form.elements.certificate_enabled.checked,
        quollix_errors_enabled: form.elements["quollix-errors_enabled"].checked,
        host_usage_trigger_percent: Number.parseInt(form.elements.host_usage_trigger_percent.value, 10),
        host_usage_recovery_percent: Number.parseInt(form.elements.host_usage_recovery_percent.value, 10),
        host_usage_pending_minutes: Number.parseInt(form.elements.host_usage_pending_minutes.value, 10),
    }

    const ok = await window.apiPost('{{$.Static.Paths.BackendSettingsAlertsSave}}', settings)
    if (ok) window.reloadPageAndShowSnackbar('Alert settings saved.')
}
