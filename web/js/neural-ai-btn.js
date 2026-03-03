/**
 * Neural AI Button — neural-ai-btn.js
 * Подключать ПОСЛЕ app.js
 */
(function () {
  const canvas = document.getElementById('neural-canvas');
  const ctx = canvas.getContext('2d');
  let active = false;
  let animFrame = null;
  let nodes = [];
  let time = 0;

  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  resize();
  window.addEventListener('resize', resize);

  function createNodes() {
    const btn = document.getElementById('ai-btn');
    const rect = btn.getBoundingClientRect();
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    nodes = [];
    for (let i = 0; i < 30; i++) {
      const angle = (Math.PI * 2 / 30) * i + Math.random() * 0.4;
      const radius = 55 + Math.random() * 150;
      nodes.push({
        x: cx + Math.cos(angle) * radius,
        y: cy + Math.sin(angle) * radius,
        ox: cx + Math.cos(angle) * radius,
        oy: cy + Math.sin(angle) * radius,
        vx: (Math.random() - 0.5) * 0.3,
        vy: (Math.random() - 0.5) * 0.3,
        r: 1.5 + Math.random() * 2,
        phase: Math.random() * Math.PI * 2,
      });
    }
    nodes.push({ x: cx, y: cy, r: 4, isCenter: true });
  }

  function drawFrame() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    time += 0.012;
    const center = nodes[nodes.length - 1];
    const cx = center.x, cy = center.y;

    for (const n of nodes) {
      if (n.isCenter) continue;
      n.x += n.vx + Math.sin(time + n.phase) * 0.25;
      n.y += n.vy + Math.cos(time + n.phase) * 0.25;
      n.x += (n.ox - n.x) * 0.008;
      n.y += (n.oy - n.y) * 0.008;
    }

    for (let i = 0; i < nodes.length - 1; i++) {
      const a = nodes[i];
      const dCenter = Math.hypot(a.x - cx, a.y - cy);
      if (dCenter < 220) {
        const alpha = (1 - dCenter / 220) * 0.55;
        const grad = ctx.createLinearGradient(cx, cy, a.x, a.y);
        grad.addColorStop(0, `rgba(59,130,246,${alpha})`);
        grad.addColorStop(1, 'rgba(37,99,235,0)');
        ctx.beginPath(); ctx.moveTo(cx, cy); ctx.lineTo(a.x, a.y);
        ctx.strokeStyle = grad; ctx.lineWidth = 0.8; ctx.stroke();
      }
      for (let j = i + 1; j < nodes.length - 1; j++) {
        const b = nodes[j];
        const dist = Math.hypot(a.x - b.x, a.y - b.y);
        if (dist < 85) {
          ctx.beginPath(); ctx.moveTo(a.x, a.y); ctx.lineTo(b.x, b.y);
          ctx.strokeStyle = `rgba(59,130,246,${(1 - dist / 85) * 0.2})`;
          ctx.lineWidth = 0.5; ctx.stroke();
        }
      }
    }

    for (const n of nodes) {
      // Центральный узел не рисуем — там кнопка
      if (n.isCenter) continue;
      const pulse = 0.7 + Math.sin(time * 2 + (n.phase || 0)) * 0.3;
      // Точка — синяя
      ctx.beginPath(); ctx.arc(n.x, n.y, n.r * pulse, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(59,130,246,0.85)';
      ctx.fill();
      // Свечение — синее
      ctx.beginPath(); ctx.arc(n.x, n.y, n.r * pulse * 3, 0, Math.PI * 2);
      ctx.fillStyle = 'rgba(59,130,246,0.08)';
      ctx.fill();
    }

    animFrame = requestAnimationFrame(drawFrame);
  }

  let aiLastResult = null;

  window.toggleAiPanel = function () {
    const panel = document.getElementById('ai-panel');
    const btn   = document.getElementById('ai-btn');
    const isOpen = !panel.classList.contains('hidden');

    panel.classList.toggle('hidden', isOpen);
    active = !isOpen;
    btn.classList.toggle('active', active);
    canvas.classList.toggle('active', active);

    if (active) {
      time = 0;
      createNodes();
      drawFrame();
    } else {
      cancelAnimationFrame(animFrame);
      animFrame = null;
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      document.getElementById('ai-result').classList.add('hidden');
      document.getElementById('ai-apply-btn').classList.add('hidden');
      aiLastResult = null;
    }
  };

  window.aiAction = async function (type) {
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
      const data = await app.apiFetch('/api/ai/suggest', {
        method: 'POST',
        body: JSON.stringify({ text, action: type })
      });
      aiLastResult = data.result;
      result.textContent = data.result;
      result.classList.remove('hidden');
      if (!data.is_advice) applyBtn.classList.remove('hidden');
    } catch (err) {
      result.textContent = 'Ошибка: ' + err.message;
      result.classList.remove('hidden');
    } finally {
      loading.classList.add('hidden');
    }
  };

  window.applyAiResult = function () {
    if (!aiLastResult) return;
    document.getElementById('message-input').value = aiLastResult;
    window.toggleAiPanel();
  };

  window.animateSendButton = function () {
    const btn = document.getElementById('send-btn');
    if (!btn) return;
    btn.classList.add('sending');
    setTimeout(() => btn.classList.remove('sending'), 600);
  };
})();