window.filterTableByDataName = (filterInputId, tableBodyId) => {
    const filterValue = (document.getElementById(filterInputId).value || "").trim().toLowerCase()
    const tableBody = document.getElementById(tableBodyId)

    for (const row of tableBody.querySelectorAll("tr")) {
        const rowName = (row.getAttribute("data-name") || "").toLowerCase()
        row.style.display = (!filterValue || rowName.includes(filterValue)) ? "" : "none"
    }
}
