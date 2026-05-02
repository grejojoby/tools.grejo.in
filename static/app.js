/* ============================================================
   tools.grejo.in — app.js
   ============================================================ */

'use strict';

/* ---- API ---- */
async function api(endpoint, body) {
  try {
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    return data;
  } catch (err) {
    return { error: err.message };
  }
}

/* ---- Helpers ---- */
function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function setTextOutput(outputEl, errorEl, res) {
  if (res.error) {
    outputEl.value = '';
    outputEl.classList.add('error');
    if (errorEl) { errorEl.textContent = res.error; errorEl.classList.add('visible'); }
  } else {
    outputEl.value = res.output ?? '';
    outputEl.classList.remove('error');
    if (errorEl) { errorEl.textContent = ''; errorEl.classList.remove('visible'); }
  }
}

/* ---- Build number ---- */
fetch('/api/build').then(r => r.json()).then(d => {
  const el = document.getElementById('build-number');
  if (el && d.build) el.textContent = '#' + d.build;
}).catch(() => {});

/* ---- Copy to clipboard ---- */
document.addEventListener('click', (e) => {
  const btn = e.target.closest('.btn-copy');
  if (!btn) return;
  const id = btn.dataset.copy;
  const el = document.getElementById(id);
  if (!el) return;
  const text = el.tagName === 'TEXTAREA' ? el.value : el.textContent;
  navigator.clipboard.writeText(text).then(() => {
    const orig = btn.innerHTML;
    btn.innerHTML = `
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
      Copied!`;
    btn.classList.add('copied');
    setTimeout(() => { btn.innerHTML = orig; btn.classList.remove('copied'); }, 1600);
  });
});

/* ---- Keyboard shortcut: ⌘↵ / Ctrl+↵ ---- */
document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    const panel = document.querySelector('.tool-panel.active');
    if (panel) panel.querySelector('.btn-primary')?.click();
  }
});

/* ---- Theme ---- */
function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);
}

document.getElementById('theme-toggle').addEventListener('click', () => {
  const cur = document.documentElement.getAttribute('data-theme');
  applyTheme(cur === 'dark' ? 'light' : 'dark');
});

/* ---- Navigation ---- */
function closeSidebar() {
  document.getElementById('sidebar').classList.remove('open');
  document.getElementById('sidebar-backdrop').classList.remove('visible');
}

function navigate(toolId) {
  document.querySelectorAll('.tool-panel').forEach((p) => {
    p.classList.toggle('active', p.id === toolId);
  });
  document.querySelectorAll('.nav-item').forEach((a) => {
    a.classList.toggle('active', a.getAttribute('href') === '#' + toolId);
  });
  closeSidebar();
}

document.querySelectorAll('.nav-item').forEach((a) => {
  a.addEventListener('click', (e) => {
    e.preventDefault();
    const id = a.getAttribute('href').slice(1);
    history.pushState(null, '', '#' + id);
    navigate(id);
  });
});

// Mobile hamburger — toggle sidebar + backdrop together
document.getElementById('menu-toggle').addEventListener('click', () => {
  const isOpen = document.getElementById('sidebar').classList.toggle('open');
  document.getElementById('sidebar-backdrop').classList.toggle('visible', isOpen);
});

// Tap backdrop to close sidebar
document.getElementById('sidebar-backdrop').addEventListener('click', closeSidebar);

// Hash-based routing
window.addEventListener('popstate', () => {
  const id = location.hash.slice(1) || 'json-prettify';
  navigate(id);
});

/* ============================================================
   JSON: Prettify
   ============================================================ */
(function () {
  const input = () => document.getElementById('jp-input');
  const errEl = () => document.getElementById('jp-error');

  document.getElementById('jp-submit').addEventListener('click', async () => {
    const res = await api('/api/json/format', { input: input().value, mode: 'pretty' });
    if (res.error) {
      input().classList.add('error');
      errEl().textContent = res.error; errEl().classList.add('visible');
    } else {
      input().value = res.output ?? '';
      const lines = res.output.split('\n').length;
      const lineHeight = 20;
      const padding = 26;
      input().style.height = (lines * lineHeight + padding) + 'px';
      input().classList.remove('error');
      errEl().textContent = ''; errEl().classList.remove('visible');
    }
  });

  document.getElementById('jp-clear').addEventListener('click', () => {
    input().value = '';
    input().classList.remove('error');
    errEl().textContent = ''; errEl().classList.remove('visible');
  });
})();

/* ============================================================
   JSON: Compress
   ============================================================ */
(function () {
  const input  = () => document.getElementById('jc-input');
  const output = () => document.getElementById('jc-output');
  const errEl  = () => document.getElementById('jc-error');

  document.getElementById('jc-submit').addEventListener('click', async () => {
    const res = await api('/api/json/format', { input: input().value, mode: 'compact' });
    setTextOutput(output(), errEl(), res);
  });

  document.getElementById('jc-clear').addEventListener('click', () => {
    input().value = ''; output().value = '';
    errEl().textContent = ''; errEl().classList.remove('visible');
    output().classList.remove('error');
  });
})();

/* ============================================================
   JSON: Validate
   ============================================================ */
(function () {
  const input   = () => document.getElementById('jv-input');
  const result  = () => document.getElementById('jv-result');
  const title   = () => document.getElementById('jv-title');
  const msg     = () => document.getElementById('jv-msg');
  const iconEl  = () => document.getElementById('jv-icon');

  const checkIcon = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>`;
  const xIcon     = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`;

  document.getElementById('jv-submit').addEventListener('click', async () => {
    const res = await api('/api/json/validate', { input: input().value });
    const el = result();

    if (res.error && res.valid === undefined) {
      // network / request error
      el.className = 'validate-result visible invalid';
      iconEl().innerHTML = xIcon;
      title().textContent = 'Request Error';
      msg().textContent = res.error;
      return;
    }

    if (res.valid) {
      el.className = 'validate-result visible valid';
      iconEl().innerHTML = checkIcon;
      title().textContent = 'Valid JSON';
      msg().textContent = '';
    } else {
      el.className = 'validate-result visible invalid';
      iconEl().innerHTML = xIcon;
      title().textContent = 'Invalid JSON';
      msg().textContent = res.error || '';
    }
  });

  document.getElementById('jv-clear').addEventListener('click', () => {
    input().value = '';
    result().className = 'validate-result';
  });
})();

/* ============================================================
   JSON: Compare
   ============================================================ */
(function () {
  const inputA  = () => document.getElementById('jcmp-a');
  const inputB  = () => document.getElementById('jcmp-b');
  const errA    = () => document.getElementById('jcmp-error-a');
  const errB    = () => document.getElementById('jcmp-error-b');
  const result  = () => document.getElementById('jcmp-result');

  function clearErrors() {
    [errA(), errB()].forEach((e) => { e.textContent = ''; e.classList.remove('visible'); });
    [inputA(), inputB()].forEach((t) => t.classList.remove('error'));
  }

  // Convert flat annotated lines into {left, right} row pairs for side-by-side display.
  function buildSideBySideRows(lines) {
    const rows = [];
    let i = 0;
    while (i < lines.length) {
      const { type } = lines[i];
      if (type === 'unchanged') {
        rows.push({ left: lines[i], right: lines[i] });
        i++;
      } else if (type === 'changed') {
        // Collect the whole changed block then the whole changed_new block and zip them.
        const chg = [], chgNew = [];
        while (i < lines.length && lines[i].type === 'changed')     chg.push(lines[i++]);
        while (i < lines.length && lines[i].type === 'changed_new') chgNew.push(lines[i++]);
        const max = Math.max(chg.length, chgNew.length);
        for (let j = 0; j < max; j++) rows.push({ left: chg[j] || null, right: chgNew[j] || null });
      } else {
        // removed / added — collect consecutively then zip.
        const rem = [], add = [];
        while (i < lines.length && (lines[i].type === 'removed' || lines[i].type === 'added')) {
          lines[i].type === 'removed' ? rem.push(lines[i++]) : add.push(lines[i++]);
        }
        const max = Math.max(rem.length, add.length);
        for (let j = 0; j < max; j++) rows.push({ left: rem[j] || null, right: add[j] || null });
      }
    }
    return rows;
  }

  function lineSpan(line) {
    return line
      ? `<span class="diff-line ${escHtml(line.type)}">${escHtml(line.text)}</span>`
      : `<span class="diff-line blank"> </span>`;
  }

  function renderDiff(data) {
    const el = result();
    if (data.error) {
      if (data.error.includes('JSON A')) {
        errA().textContent = data.error; errA().classList.add('visible');
        inputA().classList.add('error');
      } else if (data.error.includes('JSON B')) {
        errB().textContent = data.error; errB().classList.add('visible');
        inputB().classList.add('error');
      }
      el.className = 'compare-result';
      return;
    }

    el.classList.add('visible');

    if (data.equal) {
      el.innerHTML = `
        <div class="compare-equal">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
          Documents are identical
        </div>`;
      return;
    }

    const rows = buildSideBySideRows(data.lines || []);
    const leftHtml  = rows.map((r) => lineSpan(r.left)).join('');
    const rightHtml = rows.map((r) => lineSpan(r.right)).join('');

    el.innerHTML = `
      <div class="diff-compare">
        <div class="diff-compare-pane">
          <div class="diff-compare-label">JSON A</div>
          <pre class="diff-view">${leftHtml}</pre>
        </div>
        <div class="diff-compare-pane">
          <div class="diff-compare-label">JSON B</div>
          <pre class="diff-view">${rightHtml}</pre>
        </div>
      </div>`;
  }

  document.getElementById('jcmp-submit').addEventListener('click', async () => {
    clearErrors();
    result().className = 'compare-result';
    const res = await api('/api/json/compare', { a: inputA().value, b: inputB().value });
    renderDiff(res);
  });

  document.getElementById('jcmp-clear').addEventListener('click', () => {
    inputA().value = ''; inputB().value = '';
    clearErrors();
    result().className = 'compare-result';
    result().innerHTML = '';
  });
})();

/* ============================================================
   JSON: Sort Keys
   ============================================================ */
(function () {
  const input  = () => document.getElementById('jsk-input');
  const output = () => document.getElementById('jsk-output');
  const errEl  = () => document.getElementById('jsk-error');

  document.getElementById('jsk-submit').addEventListener('click', async () => {
    const res = await api('/api/json/sort-keys', { input: input().value });
    setTextOutput(output(), errEl(), res);
  });

  document.getElementById('jsk-clear').addEventListener('click', () => {
    input().value = ''; output().value = '';
    errEl().textContent = ''; errEl().classList.remove('visible');
    output().classList.remove('error');
  });
})();

/* ============================================================
   URL: Parse
   ============================================================ */
(function () {
  const input  = () => document.getElementById('up-input');
  const errEl  = () => document.getElementById('up-error');
  const result = () => document.getElementById('up-result');

  function row(label, value, mono = true) {
    const isEmpty = value === '' || value === null || value === undefined;
    const td = isEmpty
      ? `<td class="empty">—</td>`
      : `<td${mono ? '' : ' style="font-family:inherit"'}>${escHtml(value)}</td>`;
    return `<tr><th>${escHtml(label)}</th>${td}</tr>`;
  }

  function renderParsed(data) {
    if (data.error) {
      errEl().textContent = data.error; errEl().classList.add('visible');
      result().className = 'parse-result';
      return;
    }
    errEl().textContent = ''; errEl().classList.remove('visible');

    const queryRows = (data.query || []).map((p) =>
      `<tr>
        <th><span class="parse-badge key">param</span>${escHtml(p.key)}</th>
        <td>${escHtml(p.value)}</td>
      </tr>`
    ).join('');

    const fragRows = (data.fragment_query || []).map((p) =>
      `<tr>
        <th><span class="parse-badge key">param</span>${escHtml(p.key)}</th>
        <td>${escHtml(p.value)}</td>
      </tr>`
    ).join('');

    const querySection = queryRows
      ? `<tr><td colspan="2" class="parse-section-label">Query Parameters</td></tr>${queryRows}`
      : '';

    const fragSection = fragRows
      ? `<tr><td colspan="2" class="parse-section-label">Fragment Parameters</td></tr>${fragRows}`
      : '';

    const authSection = (data.username || data.password)
      ? `<tr><td colspan="2" class="parse-section-label">Authentication</td></tr>
         ${row('Username', data.username)}
         ${row('Password', data.password)}`
      : '';

    result().className = 'parse-result visible';
    result().innerHTML = `
      <table class="parse-table">
        <tbody>
          ${row('Scheme',    data.scheme)}
          ${row('Host',      data.host)}
          ${row('Port',      data.port)}
          ${row('Path',      data.path)}
          ${row('Raw Query', data.raw_query)}
          ${row('Fragment',  data.fragment)}
          ${authSection}
          ${querySection}
          ${fragSection}
        </tbody>
      </table>`;
  }

  document.getElementById('up-submit').addEventListener('click', async () => {
    const res = await api('/api/url/parse', { input: input().value });
    renderParsed(res);
  });

  input().addEventListener('keydown', (e) => {
    if (e.key === 'Enter') document.getElementById('up-submit').click();
  });

  document.getElementById('up-clear').addEventListener('click', () => {
    input().value = '';
    errEl().textContent = ''; errEl().classList.remove('visible');
    result().className = 'parse-result';
    result().innerHTML = '';
  });
})();

/* ============================================================
   URL: Encode
   ============================================================ */
(function () {
  const input  = () => document.getElementById('ue-input');
  const output = () => document.getElementById('ue-output');

  document.getElementById('ue-submit').addEventListener('click', async () => {
    const res = await api('/api/url/encode', { input: input().value });
    setTextOutput(output(), null, res);
  });

  document.getElementById('ue-clear').addEventListener('click', () => {
    input().value = ''; output().value = '';
    output().classList.remove('error');
  });
})();

/* ============================================================
   URL: Decode
   ============================================================ */
(function () {
  const input  = () => document.getElementById('ud-input');
  const output = () => document.getElementById('ud-output');
  const errEl  = () => document.getElementById('ud-error');

  document.getElementById('ud-submit').addEventListener('click', async () => {
    const res = await api('/api/url/decode', { input: input().value });
    setTextOutput(output(), errEl(), res);
  });

  document.getElementById('ud-clear').addEventListener('click', () => {
    input().value = ''; output().value = '';
    errEl().textContent = ''; errEl().classList.remove('visible');
    output().classList.remove('error');
  });
})();

/* ============================================================
   Cipher: Caesar
   ============================================================ */
(function () {
  let mode = 'encode';

  const input  = () => document.getElementById('cc-input');
  const output = () => document.getElementById('cc-output');
  const keyEl  = () => document.getElementById('cc-key');
  const shiftEl = () => document.getElementById('cc-shift');

  // Mode toggle
  document.getElementById('cc-mode-group').addEventListener('click', (e) => {
    const btn = e.target.closest('.toggle-btn');
    if (!btn) return;
    mode = btn.dataset.value;
    document.querySelectorAll('#cc-mode-group .toggle-btn').forEach((b) => {
      b.classList.toggle('active', b.dataset.value === mode);
    });
  });

  // Key input: allow only letters (plus delete/backspace/arrows)
  keyEl().addEventListener('keypress', (e) => {
    if (!/[a-zA-Z]/.test(e.key)) e.preventDefault();
  });

  document.getElementById('cc-submit').addEventListener('click', async () => {
    const res = await api('/api/cipher/caesar', {
      input: input().value,
      key:   keyEl().value,
      shift: parseInt(shiftEl().value, 10) || 0,
      mode,
    });
    setTextOutput(output(), null, res);
  });

  document.getElementById('cc-clear').addEventListener('click', () => {
    input().value = ''; output().value = '';
    output().classList.remove('error');
  });
})();

/* ============================================================
   Textarea pair resize sync
   When either textarea in an .io-grid is resized by the user,
   the other textarea in the same grid snaps to the same height.
   ============================================================ */
(function syncTextareaResize() {
  document.querySelectorAll('.io-grid').forEach((grid) => {
    const textareas = Array.from(grid.querySelectorAll('.io-textarea'));
    if (textareas.length < 2) return;

    let syncing = false;
    textareas.forEach((ta) => {
      new ResizeObserver(() => {
        if (syncing) return;
        syncing = true;
        const h = ta.offsetHeight;
        textareas.forEach((other) => {
          if (other !== ta) other.style.height = h + 'px';
        });
        syncing = false;
      }).observe(ta);
    });
  });
})();

/* ============================================================
   UUID: Generate v4
   ============================================================ */
(function () {
  const count  = () => document.getElementById('uv-count');
  const output = () => document.getElementById('uv-output');
  const errEl  = () => document.getElementById('uv-error');

  document.getElementById('uv-submit').addEventListener('click', async () => {
    const n = parseInt(count().value, 10) || 1;
    const res = await api('/api/uuid/v4', { count: n });
    if (res.error) {
      output().value = '';
      errEl().textContent = res.error;
      errEl().classList.add('visible');
    } else {
      output().value = res.uuids.join('\n');
      errEl().textContent = '';
      errEl().classList.remove('visible');
    }
  });

  document.getElementById('uv-clear').addEventListener('click', () => {
    output().value = '';
    count().value = '1';
    errEl().textContent = '';
    errEl().classList.remove('visible');
  });
})();

/* ============================================================
   Bootstrap
   ============================================================ */
(function init() {
  // Restore theme
  const saved = localStorage.getItem('theme');
  if (saved) applyTheme(saved);
  else if (window.matchMedia('(prefers-color-scheme: dark)').matches) applyTheme('dark');

  // Route to correct tool
  const toolId = location.hash.slice(1) || 'json-prettify';
  navigate(toolId);
})();
