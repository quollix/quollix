window.showSnackbar = function (msg) {
    const s = document.getElementById('snackbar')
    if (!s) return
    s.textContent = msg || 'An error occurred'
    s.setAttribute('data-visible', 'true')
    s.classList.add('show')
    setTimeout(() => {
        s.classList.remove('show')
        s.setAttribute('data-visible', 'false')
    }, 3000)
}

window.confirmDialog = async function (message) {
    const dlg = document.getElementById('confirm-dialog')
    const msg = document.getElementById('confirm-message')
    msg.textContent = message
    dlg.showModal()
    return new Promise(resolve => {
        dlg.addEventListener('close', () => {
            resolve(dlg.returnValue === 'ok')
        }, { once: true })
    })
}

window.doRequest = async function(path, data) {
    return fetch(path, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data),
    });
}

async function executePostRequest(path, data, options = {}) {
    const returnBody = options.returnBody === true
    const networkChangedTolerant = options.networkChangedTolerant === true

    try {
        const res = await window.doRequest(path, data)

        if (!res.ok) {
            const msg = await res.text()
            showSnackbar(msg || 'Request failed')
            return returnBody ? [null, false] : false
        }

        if (!returnBody) return true

        const body = await res.json()
        return [body, true]
    } catch (error) {
        if (networkChangedTolerant) {
            console.log("error:", error, "| visit {{ .Links.Website }} and search for 'NETWORK_CHANGED' for more information")
            return returnBody ? [null, true] : true
        }

        showSnackbar(error?.message || 'Request failed')
        return returnBody ? [null, false] : false
    }
}

window.apiPostWithBody = async function (path, data) {
    return await executePostRequest(path, data, { returnBody: true })
}

window.apiPost = async function (path, data) {
    return await executePostRequest(path, data)
}

window.doNetworkChangedRequest = async function (path, data) {
    return await executePostRequest(path, data, { networkChangedTolerant: true })
}

window.doNetworkChangedRequestWithBody = async function (path, data) {
    return await executePostRequest(path, data, {
        returnBody: true,
        networkChangedTolerant: true,
    })
}

window.loadAsyncPageData = function (buildUrl, renderData) {
    let isFirstLoad = true

    async function load() {
        try {
            const response = await fetch(buildUrl(isFirstLoad))
            isFirstLoad = false
            if (!response.ok) throw new Error(await response.text())

            const data = await response.json()
            if (data.is_running) {
                setTimeout(load, 200)
                return
            }

            renderData(data)
        } catch (error) {
            console.log("error:", error, "| visit {{ .Links.Website }} and search for 'NETWORK_CHANGED' for more information")
            setTimeout(load, 200)
        }
    }

    load()
}

window.installApp = async (maintainer, app, version) => {
    const ok = await apiPost('{{ $.Paths.BackendStoreVersionsInstall }}', {
        Maintainer: maintainer,
        AppName: app,
        VersionName: version
    })
    if (ok) showSnackbar('Installation successful')
}

function base64ToBytes(base64String) {
    const bin = atob(base64String)
    const bytes = new Uint8Array(bin.length)
    for (let index = 0; index < bin.length; index++) bytes[index] = bin.charCodeAt(index)
    return bytes
}

window.downloadFile = async (path, payload) => {
    const [body, ok] = await window.apiPostWithBody(path, payload)
    if (!ok) return
    if (!body) return

    const base64Content = body.content
    if (!base64Content) return

    const fileName = body.file_name
    const bytes = base64ToBytes(base64Content)

    const blob = new Blob([bytes], {})
    const url = URL.createObjectURL(blob)
    const anchorElement = document.createElement('a')
    anchorElement.href = url
    anchorElement.download = fileName
    document.body.appendChild(anchorElement)
    anchorElement.click()
    anchorElement.remove()
    URL.revokeObjectURL(url)
    return true
}

window.selectAndUploadFile = async function (path) {
    const fileInputElement = document.createElement('input')
    fileInputElement.type = 'file'
    fileInputElement.style.display = 'none'
    document.body.appendChild(fileInputElement)

    const selectedFile = await new Promise((resolve) => {
        fileInputElement.onchange = () => resolve(fileInputElement.files && fileInputElement.files[0])
        fileInputElement.click()
    })

    fileInputElement.remove()
    if (!selectedFile) return false

    const arrayBuffer = await selectedFile.arrayBuffer()
    const bytes = new Uint8Array(arrayBuffer)

    let binaryString = ''
    for (let index = 0; index < bytes.length; index++) {
        binaryString += String.fromCharCode(bytes[index])
    }

    const base64Content = btoa(binaryString)
    return await window.apiPost(path, {file_name: selectedFile.name, content: base64Content})
}

// these two functions enable to show a snackbar after a page reload
document.addEventListener('DOMContentLoaded', () => {
    const msg = sessionStorage.getItem('snackbarMessage')
    if (msg) {
        window.showSnackbar(msg)
        sessionStorage.removeItem('snackbarMessage')
    }
})

window.reloadPageAndShowSnackbar = (msg) => {
    sessionStorage.setItem('snackbarMessage', msg)
    location.reload()
}

window.togglePasswordVisibilityById = function (passwordInputId, toggleButtonId) {
    const passwordInput = document.getElementById(passwordInputId)
    const toggleButton = document.getElementById(toggleButtonId)
    if (!passwordInput || !toggleButton) return

    const isHidden = passwordInput.type === 'password'
    passwordInput.type = isHidden ? 'text' : 'password'

    toggleButton.setAttribute('aria-pressed', String(isHidden))
    toggleButton.setAttribute('aria-label', isHidden ? 'Hide password' : 'Show password')

    const icon = toggleButton.querySelector('i')
    if (icon) icon.className = isHidden ? 'mdi mdi-eye-off-outline' : 'mdi mdi-eye-outline'
}

document.querySelectorAll('.local-time').forEach(t => {
    const d = new Date(t.getAttribute('datetime'));
    t.textContent = d.toLocaleString();
});
