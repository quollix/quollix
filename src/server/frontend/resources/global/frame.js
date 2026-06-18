window.logout = async (event) => {
  event.preventDefault();
  const ok = await window.apiPost('{{ $.Paths.BackendUsersLogout }}', null);
  if (ok) window.location.href = '{{ $.Paths.FrontendInstalledApps }}';
}

window.reloadFrontend = async () => {
  const ok = await window.apiPost('{{ $.Paths.BackendReloadFrontendTemplatesFromFileSystem }}', null);
  if (ok) window.reloadPageAndShowSnackbar("Frontend templates reloaded successfully.");
}

window.resetTestState = async () => {
  const confirmed = await window.confirmDialog("Reset test state to snapshot? This removes current test data.")
  if (!confirmed) return

  const ok = await window.apiPost('{{ $.Paths.BackendResetTestState }}', null)
  if (ok) window.reloadPageAndShowSnackbar("Test state reset successfully.")
}

window.updateOperationIndicator = async function() {
  const requestIndicatorElement = document.getElementById('request-indicator')
  const currentOperationElement = document.getElementById('current-operation')
  if (!requestIndicatorElement || !currentOperationElement) return

  try {
    const response = await window.doRequest('{{$.Paths.BackendAppOperationInfo}}')
    if (!response.ok) return
    const data = await response.json()

    if (data.is_ongoing) {
      requestIndicatorElement.style.display = 'inline-block'
      currentOperationElement.textContent = (data.operations || []).join(', ')
    } else {
      requestIndicatorElement.style.display = 'none'
      currentOperationElement.textContent = ''
    }

    const finishedAppOperations = data.app_operations_finished || []
    const shouldReloadForFinishedOperation =
      window.location.pathname === '{{ $.Paths.FrontendInstalledApps }}' ||
      window.location.pathname === '{{ $.Paths.FrontendListBackups }}'

    if (finishedAppOperations.length > 0 && shouldReloadForFinishedOperation) {
      window.location.reload()
      return
    }
  } catch (error) {
    console.log(error)
    requestIndicatorElement.style.display = 'none'
    currentOperationElement.textContent = ''
  }
}

window.startOperationIndicatorPolling = function() {
  const pollOnce = async () => {
    await window.updateOperationIndicator()
    window.setTimeout(pollOnce, 1000)
  }
  pollOnce()
}
