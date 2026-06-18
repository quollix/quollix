window.grantSelectedApps = async (groupId) => {
    const appNames = window.getSelectedCheckboxValues("noAccessGrantedTableBody", "app-checkbox")
    if (appNames.length === 0) {
        showSnackbar("Select at least one app to grant access.")
        return
    }

    const ok = await apiPost('{{ $.Paths.BackendGroupsGrantGroupAccessToApps }}', {
        group_id: groupId,
        app_names: appNames
    })
    if (ok) window.reloadPageAndShowSnackbar("Access granted successfully.")
}

window.revokeSelectedApps = async (groupId) => {
    const appNames = window.getSelectedCheckboxValues("accessGrantedTableBody", "app-checkbox")
    if (appNames.length === 0) {
        showSnackbar("Select at least one app to revoke access.")
        return
    }

    const ok = await apiPost('{{ $.Paths.BackendGroupsRevokeGroupAccessToApps }}', {
        group_id: groupId,
        app_names: appNames
    })
    if (ok) window.reloadPageAndShowSnackbar("Access revoked successfully.")
}
