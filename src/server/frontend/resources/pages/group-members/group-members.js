window.addSelectedMembers = async (groupId) => {
    const userIds = window.getSelectedCheckboxValues("nonMembersTableBody", "member-checkbox")
    if (userIds.length === 0) {
        showSnackbar("Select at least one user to add.")
        return
    }

    const ok = await apiPost('{{ $.Paths.BackendGroupsAddUsers }}', {
        group_id: groupId,
        user_ids: userIds
    })
    if (ok) window.reloadPageAndShowSnackbar("Members added successfully.")
}

window.removeSelectedMembers = async (groupId) => {
    const userIds = window.getSelectedCheckboxValues("membersTableBody", "member-checkbox")
    if (userIds.length === 0) {
        showSnackbar("Select at least one user to remove.")
        return
    }

    const ok = await apiPost('{{ $.Paths.BackendGroupsRemoveUsers }}', {
        group_id: groupId,
        user_ids: userIds
    })
    if (ok) window.reloadPageAndShowSnackbar("Members removed successfully.")
}
