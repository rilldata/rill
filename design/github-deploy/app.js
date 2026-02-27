/* =========================================
   DataSync — Shared Interactivity
   ========================================= */

document.addEventListener('DOMContentLoaded', () => {
  setActiveNav();
  initDropdowns();
  initModals();
  initTabs();
  initCommitForm();
  initSelectAll();
  initBranchActions();
});

/* Active nav link */
function setActiveNav() {
  const page = location.pathname.split('/').pop() || 'index.html';
  document.querySelectorAll('.nav-link[href]').forEach(link => {
    if (link.getAttribute('href') === page) {
      link.classList.add('active');
    }
  });
}

/* ── Dropdowns ── */
function initDropdowns() {
  document.querySelectorAll('[data-dd]').forEach(trigger => {
    const menu = document.querySelector(trigger.dataset.dd);
    if (!menu) return;
    trigger.addEventListener('click', e => {
      e.stopPropagation();
      const isOpen = menu.classList.contains('open');
      closeAllDropdowns();
      if (!isOpen) menu.classList.add('open');
    });
  });

  document.addEventListener('click', closeAllDropdowns);
  document.addEventListener('keydown', e => { if (e.key === 'Escape') closeAllDropdowns(); });
}

function closeAllDropdowns() {
  document.querySelectorAll('.dropdown-menu.open').forEach(m => m.classList.remove('open'));
}

/* ── Modals ── */
function initModals() {
  document.querySelectorAll('[data-modal]').forEach(btn => {
    const id = btn.dataset.modal;
    btn.addEventListener('click', () => openModal(id));
  });

  document.querySelectorAll('[data-modal-close]').forEach(btn => {
    btn.addEventListener('click', () => {
      btn.closest('.modal-overlay')?.classList.remove('open');
    });
  });

  document.querySelectorAll('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', e => {
      if (e.target === overlay) overlay.classList.remove('open');
    });
  });

  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal-overlay.open').forEach(m => m.classList.remove('open'));
    }
  });
}

function openModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.add('open');
}

function closeModal(id) {
  const el = document.getElementById(id);
  if (el) el.classList.remove('open');
}

/* ── Tabs ── */
function initTabs() {
  document.querySelectorAll('.tabs').forEach(tabBar => {
    const tabs = tabBar.querySelectorAll('.tab[data-tab]');
    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const group = tab.dataset.group || tabBar.dataset.group;
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');

        if (group) {
          document.querySelectorAll(`[data-panel][data-group="${group}"]`).forEach(panel => {
            panel.style.display = panel.dataset.panel === tab.dataset.tab ? '' : 'none';
          });
        }
      });
    });
  });
}

/* ── Commit form ── */
function initCommitForm() {
  const msgInput = document.getElementById('commit-msg');
  const counter  = document.getElementById('msg-counter');

  if (msgInput && counter) {
    msgInput.addEventListener('input', () => {
      const len = msgInput.value.length;
      counter.textContent = `${len}/72`;
      counter.style.color = len > 72 ? 'var(--red)' : len > 50 ? 'var(--orange)' : 'var(--text-muted)';
    });
  }

  const form = document.getElementById('deploy-form');
  if (form) {
    form.addEventListener('submit', e => {
      e.preventDefault();
      const btn = form.querySelector('[data-submit-btn]');
      if (!btn) return;

      const checked = document.querySelectorAll('.file-cb:checked').length;
      if (checked === 0) {
        showToast('Select at least one file to commit.', 'warning');
        return;
      }
      if (!msgInput?.value.trim()) {
        showToast('Please enter a commit message.', 'warning');
        msgInput?.focus();
        return;
      }

      const label = btn.textContent;
      btn.disabled = true;
      btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg> Pushing…`;

      setTimeout(() => {
        btn.disabled = false;
        btn.textContent = label;
        // Reset checkboxes
        document.querySelectorAll('.file-cb').forEach(cb => cb.checked = false);
        const sa = document.getElementById('select-all-cb');
        if (sa) sa.checked = false;
        updateStagedCount();
        // Generate a fake short commit hash and show preview panel
        const hash = Math.random().toString(16).slice(2, 9);
        if (typeof showPreviewPanel === 'function') {
          showPreviewPanel(hash);
        } else {
          showToast(`${checked} file${checked > 1 ? 's' : ''} committed and pushed to origin.`, 'success');
        }
      }, 1800);
    });
  }
}

/* ── Select-all files ── */
function initSelectAll() {
  const sa = document.getElementById('select-all-cb');
  const cbs = () => document.querySelectorAll('.file-cb');

  if (sa) {
    sa.addEventListener('change', () => {
      cbs().forEach(cb => cb.checked = sa.checked);
      updateStagedCount();
    });
  }

  document.addEventListener('change', e => {
    if (e.target.classList.contains('file-cb')) {
      const all = [...cbs()];
      if (sa) {
        sa.checked = all.every(c => c.checked);
        sa.indeterminate = !sa.checked && all.some(c => c.checked);
      }
      updateStagedCount();
    }
  });
}

function updateStagedCount() {
  const n = document.querySelectorAll('.file-cb:checked').length;
  const el = document.getElementById('staged-count');
  if (el) el.textContent = n;
}

/* ── Branch switch ── */
function initBranchActions() {
  document.querySelectorAll('[data-checkout]').forEach(btn => {
    btn.addEventListener('click', () => {
      const name = btn.dataset.checkout;
      switchBranch(name);
    });
  });
}

function switchBranch(name) {
  // Update current-branch displays
  document.querySelectorAll('.current-branch-label').forEach(el => el.textContent = name);

  // Highlight row
  document.querySelectorAll('.branch-row').forEach(row => {
    row.classList.toggle('current', row.dataset.branch === name);
  });

  showToast(`Switched to branch "${name}"`, 'success');
}

/* ── Toast notifications ── */
function showToast(message, type = 'info', duration = 3500) {
  const icons = {
    success: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>`,
    error:   `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    info:    `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>`,
    warning: `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>`,
  };

  const colorMap = { success: 'var(--green)', error: 'var(--red)', info: 'var(--blue)', warning: 'var(--orange)' };

  let container = document.querySelector('.toast-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    document.body.appendChild(container);
  }

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `<span style="color:${colorMap[type]};display:flex">${icons[type]}</span><span>${message}</span>`;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.transition = 'opacity 0.25s, transform 0.25s';
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(40px)';
    setTimeout(() => toast.remove(), 250);
  }, duration);
}

/* ── Fake pull/sync button ── */
document.addEventListener('click', e => {
  const btn = e.target.closest('[data-action]');
  if (!btn) return;
  const action = btn.dataset.action;

  if (action === 'sync') {
    btn.disabled = true;
    const orig = btn.innerHTML;
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg> Syncing…`;
    setTimeout(() => {
      btn.disabled = false;
      btn.innerHTML = orig;
      showToast('Repository synced with remote.', 'success');
    }, 1400);
  }

  if (action === 'pull') {
    showToast('Already up to date with origin.', 'info');
  }

  if (action === 'create-branch') {
    const input = document.getElementById('new-branch-name');
    const name = input?.value.trim();
    if (!name) { showToast('Branch name cannot be empty.', 'warning'); return; }
    showToast(`Branch "${name}" created from main.`, 'success');
    closeModal('modal-new-branch');
    if (input) input.value = '';
  }

  if (action === 'delete-branch') {
    const row = btn.closest('.branch-row');
    if (!row) return;
    const name = row.querySelector('.branch-name')?.textContent.trim();
    if (row.classList.contains('current')) {
      showToast('Cannot delete the active branch.', 'error');
      return;
    }
    row.style.transition = 'opacity 0.2s';
    row.style.opacity = '0';
    setTimeout(() => row.remove(), 220);
    showToast(`Branch "${name}" deleted.`, 'info');
  }
});

/* Spin animation */
const style = document.createElement('style');
style.textContent = `.spin { animation: spin 0.8s linear infinite; } @keyframes spin { to { transform: rotate(360deg); } }`;
document.head.appendChild(style);
