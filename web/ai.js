// ─── ai.js — AI-ассистент ─────────────────────────────────────────────────

let aiLastResult = null;

function toggleAiPanel() {
    const panel = document.getElementById('ai-panel');
    const btn   = document.getElementById('ai-btn');
    panel.classList.toggle('hidden');
    btn.classList.toggle('active', !panel.classList.contains('hidden'));
}

async function aiAction(type) {
    const text = document.getElementById('message-input').value.trim();
    if (!text) return;

    const loading  = document.getElementById('ai-loading');
    const result   = document.getElementById('ai-result');
    const applyBtn = document.getElementById('ai-apply-btn');

    loading.classList.remove('hidden');
    result.classList.add('hidden');
    applyBtn.classList.add('hidden');
    aiLastResult = null;

    try {
        const token    = localStorage.getItem('alpha_token');
        const response = await fetch('/api/ai/suggest', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({ text, action: type }),
        });

        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Ошибка сервера');

        aiLastResult        = data.result;
        result.textContent  = data.result;
        result.classList.remove('hidden');

        if (!data.is_advice) applyBtn.classList.remove('hidden');
    } catch (err) {
        result.textContent = 'Ошибка: ' + err.message;
        result.classList.remove('hidden');
    } finally {
        loading.classList.add('hidden');
    }
}

function applyAiResult() {
    if (!aiLastResult) return;
    document.getElementById('message-input').value = aiLastResult;
    document.getElementById('ai-panel').classList.add('hidden');
    document.getElementById('ai-result').classList.add('hidden');
    document.getElementById('ai-apply-btn').classList.add('hidden');
    document.getElementById('ai-btn').classList.remove('active');
    aiLastResult = null;
}
