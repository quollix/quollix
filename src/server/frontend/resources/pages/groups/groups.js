window.createGroup = async (groupName) => {
    const name = (groupName || "").trim()
    if (!name) {
        showSnackbar("Please enter a group name.")
        return
    }
    const ok = await apiPost('{{ $.Paths.BackendGroupsCreate }}', { value: name })
    if (ok) window.reloadPageAndShowSnackbar("Group created successfully.")
}

window.deleteGroup = async (groupId, groupName) => {
    const confirmed = await confirmDialog(`Delete group '${groupName}'?`)
    if (!confirmed) return
    const ok = await apiPost('{{ $.Paths.BackendGroupsDelete }}', { value: groupId })
    if (ok) window.reloadPageAndShowSnackbar("Group deleted successfully.")
}

window.manageGroupMembers = (groupId) => {
    window.location.href = `{{ $.Paths.FrontendGroupMembers }}?group-id=${encodeURIComponent(groupId)}`
}

window.manageGroupApps = (groupId) => {
    window.location.href = `{{ $.Paths.FrontendGroupApps }}?group-id=${encodeURIComponent(groupId)}`
}