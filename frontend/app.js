/* ═══ State ═══ */
let activeTransfers = [];
let targetPeerAddress = "";
let knownFiles = new Set();
let currentAccount = null;
let allHistory = [];
let allFiles = [];

function getFileType(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext)) return 'image';
    if (['mp4', 'mkv', 'avi', 'mov', 'webm'].includes(ext)) return 'video';
    if (['pdf', 'doc', 'docx', 'txt', 'rtf', 'xls', 'xlsx'].includes(ext)) return 'document';
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return 'archive';
    return 'other';
}

/* ═══ Notifications ═══ */
let pairRequests = [];
let transferRequests = [];

function toggleNotifications() {
    const dropdown = document.getElementById('notifications-dropdown');
    dropdown.style.display = dropdown.style.display === 'none' ? 'flex' : 'none';
}

async function fetchNotifications() {
    if (!currentAccount) return;
    try {
        const prRes = await fetch('/api/pair-requests');
        const trRes = await fetch('/api/transfer-requests');
        if (prRes.ok) pairRequests = await prRes.json() || [];
        if (trRes.ok) transferRequests = await trRes.json() || [];
        
        renderNotifications();
    } catch (e) { console.log("Bildirim alınamadı", e); }
}

function renderNotifications() {
    const list = document.getElementById('notifications-list');
    const badge = document.getElementById('notification-badge');
    const total = pairRequests.length + transferRequests.length;
    
    if (total > 0) {
        badge.style.display = 'flex';
        badge.textContent = total;
    } else {
        badge.style.display = 'none';
    }

    if (total === 0) {
        list.innerHTML = `<div class="empty-state" style="padding: 1.5rem 1rem;"><p style="font-size: 0.9rem;">Bildirim yok</p></div>`;
        return;
    }

    let html = '';
    
    pairRequests.forEach(req => {
        html += `
        <div class="notification-item">
            <p><strong>${req.fromName}</strong> sizinle eşleşmek istiyor.</p>
            <div class="notification-actions">
                <button class="btn-accept" onclick="acceptPair('${req.id}')">Kabul Et</button>
                <button class="btn-reject" onclick="rejectPair('${req.id}')">Reddet</button>
            </div>
        </div>`;
    });

    transferRequests.forEach(req => {
        html += `
        <div class="notification-item">
            <p><strong>${req.fromName}</strong> size <strong>${req.filename}</strong> dosyasını göndermek istiyor.</p>
            <div class="notification-actions">
                <button class="btn-accept" onclick="acceptTransfer('${req.id}')">İndir</button>
                <button class="btn-reject" onclick="rejectTransfer('${req.id}')">Reddet</button>
            </div>
        </div>`;
    });

    list.innerHTML = html;
}

async function acceptPair(id) {
    await fetch('/api/pair-accept', { method: 'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({reqId: id}) });
    showToast("Eşleştirme kabul edildi");
    fetchNotifications(); fetchPeers();
}
async function rejectPair(id) {
    await fetch('/api/pair-reject', { method: 'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({reqId: id}) });
    fetchNotifications();
}
async function acceptTransfer(id) {
    await fetch('/api/transfer-accept', { method: 'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({reqId: id}) });
    showToast("Dosya transferi başlatılıyor...");
    fetchNotifications(); fetchFiles(); fetchHistory();
}
async function rejectTransfer(id) {
    await fetch('/api/transfer-reject', { method: 'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({reqId: id}) });
    fetchNotifications();
}

/* ═══ Toast ═══ */
function showToast(message) {
    const c = document.getElementById('toast-container');
    const t = document.createElement('div');
    t.className = 'toast';
    t.innerHTML = `<strong>🔒</strong> ${message}`;
    c.appendChild(t);
    setTimeout(() => {
        t.style.opacity = '0'; t.style.transform = 'translateX(30px)'; t.style.transition = 'all 0.4s ease';
        setTimeout(() => t.remove(), 400);
    }, 5000);
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

/* ═══ Auth Functions ═══ */
function showLogin() {
    document.getElementById('auth-register').style.display = 'none';
    document.getElementById('auth-login').style.display = 'block';
    document.getElementById('auth-welcome').style.display = 'none';
}

function showRegister() {
    document.getElementById('auth-register').style.display = 'block';
    document.getElementById('auth-login').style.display = 'none';
    document.getElementById('auth-welcome').style.display = 'none';
}

async function doRegister() {
    const email = document.getElementById('reg-email').value.trim();
    const password = document.getElementById('reg-password').value;
    if (!email || !password) return showToast('E-posta ve şifre gerekli');
    if (password.length < 6) return showToast('Şifre en az 6 karakter olmalı');

    try {
        const res = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text);
        }
        const account = await res.json();
        currentAccount = account;

        document.getElementById('welcome-id-text').textContent = account.id;
        document.getElementById('auth-register').style.display = 'none';
        document.getElementById('auth-login').style.display = 'none';
        document.getElementById('auth-welcome').style.display = 'block';
    } catch (err) {
        showToast('Hata: ' + err.message);
    }
}

async function doLogin() {
    const email = document.getElementById('login-email').value.trim();
    const password = document.getElementById('login-password').value;
    if (!email || !password) return showToast('E-posta ve şifre gerekli');

    try {
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text);
        }
        const account = await res.json();
        currentAccount = account;
        enterApp();
    } catch (err) {
        showToast('Hata: ' + err.message);
    }
}

async function doLogout() {
    try {
        await fetch('/api/logout', { method: 'POST' });
        currentAccount = null;
        document.getElementById('app-shell').style.display = 'none';
        document.getElementById('auth-screen').style.display = 'flex';
        showLogin();
        showToast('Çıkış yapıldı');
    } catch (err) {
        showToast('Çıkış yapılamadı');
    }
}

function copyID() {
    const id = document.getElementById('welcome-id-text').textContent;
    navigator.clipboard.writeText(id).then(() => showToast('ID kopyalandı!'));
}

function enterApp() {
    document.getElementById('auth-screen').style.display = 'none';
    document.getElementById('app-shell').style.display = 'flex';
    updateProfileUI();
    initApp();
}

async function checkAuth() {
    try {
        const res = await fetch('/api/profile');
        if (res.ok) {
            currentAccount = await res.json();
            enterApp();
        }
    } catch (e) {
        // Not logged in, show auth screen
    }
}

function updateProfileUI() {
    if (!currentAccount) return;

    const name = currentAccount.displayName || currentAccount.id;
    const initial = name.charAt(0).toUpperCase();
    const color = currentAccount.avatarColor || '#818cf8';

    // Sidebar mini profile
    document.getElementById('profile-mini-name').textContent = name;
    document.getElementById('profile-mini-id').textContent = currentAccount.id;
    const avatar = document.getElementById('profile-avatar');
    if (currentAccount.avatarBase64) {
        avatar.innerHTML = `<img src="${currentAccount.avatarBase64}" class="profile-avatar-img" />`;
        avatar.style.background = 'transparent';
    } else {
        avatar.textContent = initial;
        avatar.style.background = color;
    }

    // Settings profile
    document.getElementById('settings-name').textContent = name;
    document.getElementById('settings-id').textContent = currentAccount.id;
    const avatarLg = document.getElementById('settings-avatar');
    if (currentAccount.avatarBase64) {
        avatarLg.innerHTML = `<img src="${currentAccount.avatarBase64}" class="profile-avatar-img" />`;
        avatarLg.style.background = 'transparent';
    } else {
        avatarLg.textContent = initial;
        avatarLg.style.background = color;
    }
    document.getElementById('profile-name-input').value = currentAccount.displayName || '';
}

async function uploadAvatar(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async function(e) {
        const base64 = e.target.result;
        try {
            const res = await fetch('/api/profile/avatar', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ avatarBase64: base64 })
            });
            if (res.ok) {
                showToast('Profil fotoğrafı güncellendi');
                checkAuth(); // refresh profile
            }
        } catch (err) {
            showToast('Fotoğraf yüklenemedi');
        }
    };
    reader.readAsDataURL(file);
}

async function updateProfile() {
    const name = document.getElementById('profile-name-input').value.trim();
    try {
        const res = await fetch('/api/profile', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ displayName: name })
        });
        if (res.ok) {
            currentAccount = await res.json();
            updateProfileUI();
            showToast('Profil güncellendi!');
        }
    } catch (e) {
        showToast('Profil güncellenemedi');
    }
}

/* ═══ Pairing ═══ */
async function doPair() {
    const id = document.getElementById('pair-id-input').value.trim();
    if (!id) return showToast('Lütfen bir cihaz ID girin');

    try {
        const res = await fetch('/api/pair', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ peerId: id })
        });

        if (!res.ok) {
            const text = await res.text();
            throw new Error(text);
        }

        const data = await res.json();
        showToast(`Eşleştirme başarılı! ${data.peer.name}`);
        document.getElementById('pair-id-input').value = '';

        // Refresh profile to get updated pairedPeers
        const profRes = await fetch('/api/profile');
        if (profRes.ok) currentAccount = await profRes.json();

        // The API now returns { message: "..." } instead of completing immediately.
        showToast("Eşleştirme isteği gönderildi. Karşı tarafın onayı bekleniyor.");
    } catch (err) {
        showToast('Hata: ' + err.message);
    }
}

/* ═══ Transfers ═══ */
function renderTransfers() {
    const list = document.getElementById('transfer-list');
    const badge = document.getElementById('transfer-count');
    if (!list) return;
    badge.textContent = activeTransfers.length;

    if (activeTransfers.length === 0) {
        list.innerHTML = `<div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            <p>Henüz aktif transfer yok</p><span>Cihaz Eşleştir sekmesinden bir cihaz ekleyip dosya gönderin</span></div>`;
        return;
    }

    list.innerHTML = '';
    activeTransfers.forEach(t => {
        const el = document.createElement('div');
        el.className = 'transfer-item';
        el.innerHTML = `
            <div class="transfer-info"><span class="file-name">${t.filename}</span><span class="transfer-stats">${t.speed} • ${t.progress}%</span></div>
            <div class="progress-bar-container"><div class="progress-bar" style="width: ${t.progress}%"></div></div>
            <div class="transfer-info"><span style="color: var(--text-muted); font-size: 0.78rem;">${t.direction}</span></div>`;
        list.appendChild(el);
    });
}

/* ═══ Peers ═══ */
async function fetchPeers() {
    try {
        const res = await fetch('/api/peers');
        const peers = await res.json();
        const list = document.getElementById('peers-list');
        const badge = document.getElementById('peer-count');
        const statPeers = document.getElementById('stat-peers');
        if (!list) return;

        const count = peers ? peers.length : 0;
        badge.textContent = count;
        if (statPeers) statPeers.textContent = count;

        if (!peers || peers.length === 0) {
            list.innerHTML = `<div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                <p>Henüz eşleşmiş cihaz yok</p><span>Yukarıdaki kutuya arkadaşının ID'sini gir</span></div>`;
            return;
        }

        list.innerHTML = '';
        peers.forEach(peer => {
            const el = document.createElement('div');
            el.className = 'transfer-item';
            el.innerHTML = `
                <div class="transfer-info" style="width: 100%;">
                    <div>
                        <div style="display: flex; align-items: center;">
                            <span class="peer-name">${peer.name || peer.userId}</span>
                            <button class="btn-icon-edit" onclick="editAlias('${peer.userId}', '${peer.name || peer.userId}')" title="İsmi Değiştir">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                            </button>
                        </div>
                        <div class="peer-id">${peer.userId}</div>
                        <div class="peer-address">${peer.address}</div>
                    </div>
                    <button class="btn-primary send-to-peer-btn" data-address="${peer.address}">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
                        Dosya Gönder
                    </button>
                </div>`;
            list.appendChild(el);
        });

        document.querySelectorAll('.send-to-peer-btn').forEach(btn => {
            btn.addEventListener('click', e => {
                targetPeerAddress = e.currentTarget.getAttribute('data-address');
                document.getElementById('file-picker').click();
            });
        });
    } catch (err) {
        console.error('Peer error', err);
    }
}

async function editAlias(peerId, currentName) {
    const newAlias = prompt(`"${currentName}" için yeni bir takma ad belirleyin:\n(Eski haline getirmek için boş bırakın)`, currentName);
    if (newAlias === null) return; // User cancelled

    try {
        const res = await fetch('/api/peers/alias', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ peerId: peerId, alias: newAlias.trim() })
        });
        if (res.ok) {
            showToast('Takma ad başarıyla kaydedildi');
            fetchPeers();
            fetchHistory();
        } else {
            throw new Error(await res.text());
        }
    } catch (err) {
        showToast('Takma ad ayarlanamadı: ' + err.message);
    }
}

/* ═══ Files ═══ */
async function fetchFiles() {
    try {
        const res = await fetch('/api/files');
        allFiles = (await res.json()) || [];
        updateFilePeerFilter();
        renderFiles();
    } catch (err) { console.error("Files error:", err); }
}

function updateFilePeerFilter() {
    const filterSelect = document.getElementById('file-peer-filter');
    const currentValue = filterSelect.value;
    
    const peers = new Set();
    allFiles.forEach(f => {
        if (f.peerName) peers.add(f.peerName);
    });

    filterSelect.innerHTML = '<option value="all">Tüm Kişiler</option>';
    Array.from(peers).sort().forEach(p => {
        const opt = document.createElement('option');
        opt.value = p;
        opt.textContent = p;
        if (p === currentValue) opt.selected = true;
        filterSelect.appendChild(opt);
    });
}

function renderFiles() {
    const list = document.getElementById('files-list');
    const badge = document.getElementById('file-count');
    const statSynced = document.getElementById('stat-synced');
    const peerFilter = document.getElementById('file-peer-filter').value;
    const typeFilter = document.getElementById('file-type-filter').value;
    if (!list) return;

    let filtered = allFiles;
    if (peerFilter !== 'all') {
        filtered = filtered.filter(f => (f.peerName || 'Bilinmeyen Kaynak') === peerFilter);
    }
    if (typeFilter !== 'all') {
        filtered = filtered.filter(f => getFileType(f.name) === typeFilter);
    }

    if (!filtered || filtered.length === 0) {
        list.innerHTML = `<div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
            <p>Klasörde dosya yok</p><span>Filtreye uygun dosya bulunamadı veya henüz transfer yapılmadı.</span></div>`;
        badge.textContent = '0';
        if (statSynced) statSynced.textContent = '0 KB';
        return;
    }

    // Sort newest to oldest based on modTime
    filtered.sort((a, b) => b.modTime.localeCompare(a.modTime));

    badge.textContent = filtered.length;
    let totalBytes = 0;
    list.innerHTML = '';

    let currentPeer = "";

    filtered.forEach(file => {
        totalBytes += file.size;
        const pName = file.peerName || 'Bilinmeyen Kaynak';
        
        if (peerFilter === 'all' && pName !== currentPeer) {
            currentPeer = pName;
            const header = document.createElement('div');
            header.className = 'history-group-header';
            header.textContent = currentPeer;
            list.appendChild(header);
        }

        const el = document.createElement('div');
        el.className = 'transfer-item';
        el.innerHTML = `
            <div class="transfer-info" style="width: 100%;">
                <div><span class="file-name">${file.name}</span><div class="file-meta">${formatSize(file.size)} • ${file.modTime}</div></div>
                <a href="/api/download?file=${encodeURIComponent(file.name)}" class="btn-primary btn-download">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                    İndir
                </a>
            </div>`;
        list.appendChild(el);
    });

    if (statSynced) statSynced.textContent = formatSize(totalBytes);
}

/* ═══ History ═══ */
async function fetchHistory() {
    try {
        const res = await fetch('/api/history');
        allHistory = (await res.json()) || [];
        updateHistoryPeerFilter();
        renderHistory();
    } catch (err) { console.error("History error:", err); }
}

function updateHistoryPeerFilter() {
    const filterSelect = document.getElementById('history-peer-filter');
    const currentValue = filterSelect.value;
    
    // Get unique peer names
    const peers = new Set();
    allHistory.forEach(h => {
        peers.add(h.peerName || h.peerId);
    });

    filterSelect.innerHTML = '<option value="all">Tüm Kişiler</option>';
    Array.from(peers).sort().forEach(p => {
        const opt = document.createElement('option');
        opt.value = p;
        opt.textContent = p;
        if (p === currentValue) opt.selected = true;
        filterSelect.appendChild(opt);
    });
}

function renderHistory() {
    const list = document.getElementById('history-list');
    const badge = document.getElementById('history-count');
    const peerFilter = document.getElementById('history-peer-filter').value;
    const typeFilter = document.getElementById('history-type-filter').value;
    if (!list) return;

    let filtered = allHistory;
    
    if (peerFilter !== 'all') {
        filtered = filtered.filter(h => (h.peerName || h.peerId) === peerFilter);
    }
    if (typeFilter !== 'all') {
        filtered = filtered.filter(h => getFileType(h.filename) === typeFilter);
    }

    // Sort newest to oldest based on timestamp string
    filtered.sort((a, b) => b.timestamp.localeCompare(a.timestamp));

    badge.textContent = filtered.length;

    if (!filtered || filtered.length === 0) {
        list.innerHTML = `<div class="empty-state">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <p>Sonuç Bulunamadı</p><span>Filtreye uygun transfer kaydı yok.</span></div>`;
        return;
    }

    list.innerHTML = '';
    
    // Group by peer Name if peerFilter is 'all'
    let currentPeer = "";
    
    filtered.forEach(h => {
        const pName = h.peerName || h.peerId;
        if (peerFilter === 'all' && pName !== currentPeer) {
            currentPeer = pName;
            const header = document.createElement('div');
            header.className = 'history-group-header';
            header.textContent = currentPeer;
            list.appendChild(header);
        }

        const el = document.createElement('div');
        el.className = 'transfer-item';
        const dirClass = h.direction === 'sent' ? 'history-sent' : 'history-received';
        const dirLabel = h.direction === 'sent' ? '↑ Gönderildi' : '↓ Alındı';
        el.innerHTML = `
            <div class="transfer-info" style="width: 100%;">
                <div>
                    <span class="file-name">${h.filename}</span>
                    <div class="file-meta">${pName} • ${h.timestamp}</div>
                </div>
                <span class="history-direction ${dirClass}">${dirLabel}</span>
            </div>`;
        list.appendChild(el);
    });
}

/* ═══ File Polling ═══ */
async function initFiles() {
    try {
        const res = await fetch('/api/files');
        const files = await res.json();
        if (files) files.forEach(f => knownFiles.add(f.name));
        fetchFiles();
    } catch (e) {}
}

async function pollFiles() {
    try {
        const res = await fetch('/api/files');
        const files = await res.json();
        if (files) {
            files.forEach(f => {
                if (!knownFiles.has(f.name)) {
                    knownFiles.add(f.name);
                    showToast(`Yeni dosya alındı: <b>${f.name}</b>`);
                }
            });
        }
        fetchFiles();
    } catch (e) {}
}

/* ═══ Init ═══ */
function initApp() {
    renderTransfers();
    initFiles();
    fetchPeers();
    fetchHistory();
    setInterval(fetchPeers, 3000);
    setInterval(pollFiles, 2000);
    setInterval(fetchHistory, 5000);
    setInterval(fetchNotifications, 2000);
}

document.addEventListener('DOMContentLoaded', () => {
    // Navigation
    const navItems = document.querySelectorAll('.nav-item');
    const views = document.querySelectorAll('.view-section');
    const pageTitle = document.getElementById('page-title');
    const pageSub = document.getElementById('page-subtitle');

    const subtitles = {
        'dashboard-view': 'Uçtan uca şifreli dosya senkronizasyonu',
        'folders-view': 'Senkronize edilmiş dosyaları görüntüle ve indir',
        'peers-view': 'ID ile cihaz eşleştir ve dosya gönder',
        'history-view': 'Gönderilen ve alınan dosyaların kaydı',
        'settings-view': 'Profil bilgilerinizi yönetin'
    };

    navItems.forEach(item => {
        item.addEventListener('click', e => {
            e.preventDefault();
            navItems.forEach(n => n.classList.remove('active'));
            views.forEach(v => v.classList.remove('active'));
            item.classList.add('active');
            const targetId = item.getAttribute('data-target');
            document.getElementById(targetId).classList.add('active');
            pageTitle.textContent = item.querySelector('span').textContent;
            if (pageSub) pageSub.textContent = subtitles[targetId] || '';
        });
    });

    // Profile mini card click → settings
    const profileMini = document.getElementById('profile-mini');
    if (profileMini) {
        profileMini.addEventListener('click', () => {
            document.querySelector('[data-target="settings-view"]').click();
        });
    }

    // File picker
    const filePicker = document.getElementById('file-picker');
    if (filePicker) {
        filePicker.addEventListener('change', async e => {
            const file = e.target.files[0];
            if (!file || !targetPeerAddress) return;

            try {
                const formData = new FormData();
                formData.append('file', file);

                const uploadRes = await fetch('/api/upload', { method: 'POST', body: formData });
                if (!uploadRes.ok) throw new Error('Yükleme başarısız');
                const uploadData = await uploadRes.json();

                document.querySelector('[data-target="dashboard-view"]').click();

                activeTransfers.push({
                    id: Date.now(), filename: uploadData.filename,
                    progress: 15, speed: 'Hazırlanıyor...', direction: 'Şifreleniyor...'
                });
                renderTransfers();
                knownFiles.add(uploadData.filename);

                const syncRes = await fetch('/api/sync', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ filename: uploadData.filename, targetPeer: targetPeerAddress })
                });

                if (!syncRes.ok) throw new Error('İstek gönderilemedi');

                // İstek gönderildi, listeden "gönderildi" aşamasını kaldırıp mesaj gösteriyoruz.
                const idx = activeTransfers.findIndex(t => t.filename === uploadData.filename);
                if (idx !== -1) {
                    activeTransfers.splice(idx, 1);
                }
                renderTransfers();
                showToast(`${uploadData.filename} için karşı tarafa onay isteği gönderildi.`);
                fetchHistory();
            } catch (err) {
                showToast(`Hata: ${err.message}`);
            } finally {
                filePicker.value = '';
                targetPeerAddress = '';
            }
        });
    }

    // Check if already logged in
    checkAuth();
});
