window.showBackups = function(maintainer, app) {
  const params = new URLSearchParams();
  params.set('maintainer', maintainer);
  params.set('app', app);
  window.location.href = '{{ $.Paths.FrontendListBackups }}?' + params.toString();
};

const backedUpAppsLoadingMessage = document.getElementById('backed-up-apps-loading-message');
const backedUpAppsContainer = document.getElementById('backed-up-apps-container');

function renderBackedUpApps(apps) {
  backedUpAppsLoadingMessage?.remove();
  backedUpAppsContainer.replaceChildren();

  if (apps.length === 0) {
    const emptyMessage = document.createElement('p');
    emptyMessage.id = 'backed-up-apps-empty-message';
    emptyMessage.textContent = 'No apps with backups were found yet.';
    backedUpAppsContainer.appendChild(emptyMessage);

    const hint = document.createElement('p');
    hint.textContent = 'You can create backups manually from the Installed apps page, or let the maintenance agent create them automatically if configured.';
    backedUpAppsContainer.appendChild(hint);
    return;
  }

  const table = document.createElement('table');
  const thead = document.createElement('thead');
  const headerRow = document.createElement('tr');
  for (const label of ['Maintainer', 'App', 'Action']) {
    const th = document.createElement('th');
    th.textContent = label;
    headerRow.appendChild(th);
  }
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement('tbody');
  for (const app of apps) {
    const row = document.createElement('tr');
    row.className = 'backed-up-app-row';
    row.dataset.maintainer = app.maintainer;
    row.dataset.appName = app.app_name;

    const maintainerCell = document.createElement('td');
    maintainerCell.className = 'backed-up-app-maintainer-cell';
    maintainerCell.textContent = app.maintainer;
    row.appendChild(maintainerCell);

    const appNameCell = document.createElement('td');
    appNameCell.className = 'backed-up-app-name-cell';
    appNameCell.textContent = app.app_name;
    row.appendChild(appNameCell);

    const actionCell = document.createElement('td');
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'lean icon-btn backed-up-app-list-backups-button';
    button.title = 'List backups';
    button.setAttribute('aria-label', `List backups for ${app.maintainer}/${app.app_name}`);
    button.innerHTML = '<i class="mdi mdi-format-list-bulleted" aria-hidden="true"></i>';
    button.addEventListener('click', () => showBackups(app.maintainer, app.app_name));
    actionCell.appendChild(button);
    row.appendChild(actionCell);

    tbody.appendChild(row);
  }
  table.appendChild(tbody);
  backedUpAppsContainer.appendChild(table);
}

if (backedUpAppsContainer) {
  window.loadAsyncPageData(
    () => '{{ $.Paths.BackendBackedUpAppsPage }}',
    (data) => renderBackedUpApps(data.apps || []),
  );
}
