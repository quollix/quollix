window.filterTableByDataName = (filterInputId, tableBodyId) => {
    const filterValue = (document.getElementById(filterInputId).value || "").trim().toLowerCase()
    const tableBody = document.getElementById(tableBodyId)

    for (const row of tableBody.querySelectorAll("tr")) {
        const rowName = (row.getAttribute("data-name") || "").toLowerCase()
        row.style.display = (!filterValue || rowName.includes(filterValue)) ? "" : "none"
    }
}

window.toggleAllCheckboxes = (tableBodyId, checkboxClass, shouldBeChecked) => {
    const tableBody = document.getElementById(tableBodyId)
    for (const checkbox of tableBody.querySelectorAll(`input.${checkboxClass}`)) {
        checkbox.checked = shouldBeChecked
    }
}

window.getSelectedCheckboxValues = (tableBodyId, checkboxClass) => {
    const tableBody = document.getElementById(tableBodyId)
    const selected = tableBody.querySelectorAll(`input.${checkboxClass}:checked`)
    const values = []
    for (const checkbox of selected) values.push(checkbox.value)
    return values
}