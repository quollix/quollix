window.restoreBackup = async (backupId, maintainer, app, version, backupCreationDate) => {
  const confirmed = await confirmDialog(`Restore backup for ${maintainer}/${app} version ${version} from ${backupCreationDate}? Current app data will be overwritten.`)
  if (!confirmed) return

  if (app === 'postgres') {
    const ok = await doNetworkChangedRequest('{{ $.Paths.BackendBackupsRestore }}', {
      backup_id: backupId,
    })

    if (ok) {
      // Restoring an postgres backup usually invalidates the current admin session cookie, so he need to be redirected to the sign-in page. Also, the restore operation aborts prematurely due to the NETWORK_CHANGED error, so we have to add a delay before redirecting.
      showSnackbar('Backup restored successfully. Redirecting to sign-in...')
      setTimeout(() => window.location.href = '{{ $.Paths.FrontendSignIn }}', 3000)
    }
    return
  }

  await doNetworkChangedRequest('{{ $.Paths.BackendBackupsRestore }}', {
    backup_id: backupId,
  })
}

window.deleteBackup = async (backupId, maintainer, app, version) => {
  const confirmed = await confirmDialog(`Delete backup for ${maintainer}/${app} version ${version}? This cannot be undone.`)
  if (!confirmed) return

  await doNetworkChangedRequest('{{ $.Paths.BackendBackupsDelete }}', {
    backup_ids: [backupId],
  })
}

const backupsPageMaintainer = document.getElementById('backups-page-maintainer')?.textContent?.trim() || '';
const backupsPageApp = document.getElementById('backups-page-app-name')?.textContent?.trim() || '';
const backupsLoadingMessage = document.getElementById('backups-loading-message');
const backupsContainer = document.getElementById('backups-container');

function buildBackupsPageDataUrl(isFirstLoad) {
  const params = new URLSearchParams();
  params.set('maintainer', backupsPageMaintainer);
  params.set('app', backupsPageApp);
  if (isFirstLoad) params.set('reload', 'true');
  return '{{ $.Paths.BackendBackupsPage }}?' + params.toString();
}

function renderBackups(backups) {
  backupsLoadingMessage?.remove();
  backupsContainer.replaceChildren();

  if (backups.length === 0) {
    const emptyMessage = document.createElement('p');
    emptyMessage.id = 'backups-empty-message';
    emptyMessage.textContent = 'No backups were found for this app.';
    backupsContainer.appendChild(emptyMessage);
    return;
  }

  const table = document.createElement('table');
  const thead = document.createElement('thead');
  const headerRow = document.createElement('tr');
  for (const label of ['App Version', 'Description', 'Backup Creation Date', 'Quollix Version', 'Actions']) {
    const th = document.createElement('th');
    th.textContent = label;
    headerRow.appendChild(th);
  }
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement('tbody');
  for (const backup of backups) {
    const row = document.createElement('tr');
    row.className = 'backup-row';
    row.dataset.backupId = backup.backup_id;

    const versionCell = document.createElement('td');
    versionCell.className = 'backup-version-name-cell';
    versionCell.textContent = backup.version_name;
    row.appendChild(versionCell);

    const descriptionCell = document.createElement('td');
    descriptionCell.className = 'backup-description-cell';
    descriptionCell.textContent = backup.description;
    row.appendChild(descriptionCell);

    const creationDateCell = document.createElement('td');
    creationDateCell.className = 'backup-creation-date-cell';
    creationDateCell.textContent = backup.backup_creation_date;
    row.appendChild(creationDateCell);

    const createdWithVersionCell = document.createElement('td');
    createdWithVersionCell.className = 'backup-created-with-app-version-cell';
    createdWithVersionCell.textContent = backup.created_with_application_version;
    row.appendChild(createdWithVersionCell);

    const actionsCell = document.createElement('td');
    const actionsRow = document.createElement('div');
    actionsRow.className = 'actions-row';

    const restoreButton = document.createElement('button');
    restoreButton.type = 'button';
    restoreButton.className = 'lean icon-btn backup-restore-button';
    restoreButton.title = 'Restore backup';
    restoreButton.setAttribute('aria-label', `Restore backup ${backup.version_name}`);
    restoreButton.innerHTML = '<i class="mdi mdi-restore" aria-hidden="true"></i>';
    restoreButton.addEventListener('click', () => restoreBackup(
      backup.backup_id,
      backupsPageMaintainer,
      backupsPageApp,
      backup.version_name,
      backup.backup_creation_date,
    ));
    actionsRow.appendChild(restoreButton);

    const deleteButton = document.createElement('button');
    deleteButton.type = 'button';
    deleteButton.className = 'lean icon-btn alert backup-delete-button';
    deleteButton.title = 'Delete backup';
    deleteButton.setAttribute('aria-label', `Delete backup ${backup.version_name}`);
    deleteButton.innerHTML = '<i class="mdi mdi-delete-outline" aria-hidden="true"></i>';
    deleteButton.addEventListener('click', () => deleteBackup(
      backup.backup_id,
      backupsPageMaintainer,
      backupsPageApp,
      backup.version_name,
    ));
    actionsRow.appendChild(deleteButton);

    actionsCell.appendChild(actionsRow);
    row.appendChild(actionsCell);
    tbody.appendChild(row);
  }

  table.appendChild(tbody);
  backupsContainer.appendChild(table);
}

if (backupsContainer) {
  window.loadAsyncPageData(
    buildBackupsPageDataUrl,
    (data) => renderBackups(data.backups || []),
  );
}
