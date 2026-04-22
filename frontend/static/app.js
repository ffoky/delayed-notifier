const API = '';
let currentUser = null;
let refreshTimer = null;

function switchTab(tab) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    event.target.classList.add('active');
    document.getElementById('tabEmail').style.display    = tab === 'email'    ? '' : 'none';
    document.getElementById('tabTelegram').style.display = tab === 'telegram' ? '' : 'none';
}

async function doAuth() {
    const msg = document.getElementById('authMsg');
    msg.textContent = '';

    const emailVal    = document.getElementById('authEmail').value.trim();
    const telegramVal = document.getElementById('authTelegram').value.trim();

    const body = {};
    if (emailVal)    body.email       = emailVal;
    if (telegramVal) body.telegram_id = parseInt(telegramVal, 10);

    if (!body.email && !body.telegram_id) {
        msg.className = 'msg err';
        msg.textContent = 'Введите email или Telegram ID';
        return;
    }

    try {
        const res  = await fetch(`${API}/auth`, {
            method:  'POST',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify(body),
        });
        const data = await res.json();
        if (!res.ok) { msg.className = 'msg err'; msg.textContent = data.error || 'Ошибка'; return; }

        currentUser = data;
        localStorage.setItem('user', JSON.stringify(data));
        showApp();
    } catch (e) {
        msg.className = 'msg err';
        msg.textContent = e.message;
    }
}

function showApp() {
    document.getElementById('authScreen').style.display = 'none';
    document.getElementById('appScreen').style.display  = '';

    const label = currentUser.email
        ? `✉ ${currentUser.email}`
        : `✈ ${currentUser.telegram_id}`;
    document.getElementById('userLabel').textContent = label;

    setDefaultPlannedAt();
    loadNotifications();
    refreshTimer = setInterval(loadNotifications, 3000);
}

function logout() {
    localStorage.removeItem('user');
    currentUser = null;
    clearInterval(refreshTimer);
    document.getElementById('authScreen').style.display = '';
    document.getElementById('appScreen').style.display  = 'none';
}

async function createNotification() {
    const msg = document.getElementById('createMsg');
    msg.textContent = '';

    const plannedLocal = document.getElementById('plannedAt').value;
    if (!plannedLocal) { msg.className = 'msg err'; msg.textContent = 'Укажите время'; return; }

    const body = {
        user_id:    currentUser.id,
        channel:    document.getElementById('channel').value,
        text:       document.getElementById('notifText').value.trim(),
        planned_at: new Date(plannedLocal).toISOString(),
    };

    if (!body.text) { msg.className = 'msg err'; msg.textContent = 'Введите текст'; return; }

    try {
        const res  = await fetch(`${API}/notify`, {
            method:  'POST',
            headers: { 'Content-Type': 'application/json' },
            body:    JSON.stringify(body),
        });
        const data = await res.json();
        if (!res.ok) { msg.className = 'msg err'; msg.textContent = data.error || 'Ошибка'; return; }

        msg.className = 'msg ok';
        msg.textContent = `Создано: ${data.id.slice(0, 8)}…`;
        document.getElementById('notifText').value = '';
        setDefaultPlannedAt();
        await loadNotifications();
    } catch (e) {
        msg.className = 'msg err';
        msg.textContent = e.message;
    }
}

async function loadNotifications() {
    try {
        const res  = await fetch(`${API}/notify`);
        const data = await res.json();
        renderTable(Array.isArray(data) ? data : []);
        document.getElementById('refreshInfo').textContent = 'Обновлено ' + new Date().toLocaleTimeString('ru-RU');
    } catch (e) {
        document.getElementById('tableWrap').innerHTML = `<div class="empty">Ошибка: ${e.message}</div>`;
    }
}

function renderTable(rows) {
    if (!rows.length) {
        document.getElementById('tableWrap').innerHTML = '<div class="empty">Уведомлений пока нет</div>';
        return;
    }
    document.getElementById('tableWrap').innerHTML = `
    <table>
      <thead><tr>
        <th>ID</th><th>Канал</th><th>Текст</th><th>Статус</th>
        <th>Запланировано</th><th>Отправлено</th><th>Попытки</th><th></th>
      </tr></thead>
      <tbody>
        ${rows.map(n => `
          <tr>
            <td class="uuid">${n.id.slice(0,8)}…</td>
            <td>${n.channel === 'telegram'
        ? '<span class="channel-tg">✈ telegram</span>'
        : '<span class="channel-em">✉ email</span>'}</td>
            <td class="text-cell" title="${esc(n.text)}">${esc(n.text)}</td>
            <td><span class="badge badge-${n.status}">${n.status}</span></td>
            <td class="time-cell">${fmt(n.planned_at)}</td>
            <td class="time-cell">${fmt(n.sent_at)}</td>
            <td>${n.retries ?? 0}</td>
            <td>${n.status === 'planned' || n.status === 'sending'
        ? `<button class="btn btn-danger" onclick="cancelNotif('${n.id}')">Отменить</button>`
        : ''}</td>
          </tr>`).join('')}
      </tbody>
    </table>`;
}

async function cancelNotif(id) {
    const res = await fetch(`${API}/notify/${id}`, { method: 'DELETE' });
    if (!res.ok) { const d = await res.json(); alert(d.error || 'Ошибка'); return; }
    await loadNotifications();
}

function fmt(iso) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'medium' });
}

function esc(s) {
    return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function setDefaultPlannedAt() {
    const d   = new Date(Date.now() + 5 * 60 * 1000);
    const pad = n => String(n).padStart(2, '0');
    document.getElementById('plannedAt').value =
        `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

(function init() {
    const saved = localStorage.getItem('user');
    if (saved) { currentUser = JSON.parse(saved); showApp(); }
})();