/**
 * Neural AI Button — neural-ai-btn.js
 * При нажатии кнопка «расширяется» — нейросеть вырастает из кнопки наружу.
 * 30 узлов, цвет под тему через --accent-color-rgb.
 */
(function () {
  const canvas = document.getElementById('neural-canvas');
  const ctx = canvas.getContext('2d');
  let active = false;
  let animFrame = null;
  let nodes = [];
  let time = 0;
  let lastTime = 0;
  let expandProgress = 0; // 0 → 1, анимация расширения
  let themeRGB = '59,130,246';

  function getThemeRGB() {
    let rgb = getComputedStyle(document.body).getPropertyValue('--accent-color-rgb');
    if (!rgb || !rgb.trim()) {
      rgb = getComputedStyle(document.documentElement).getPropertyValue('--accent-color-rgb');
    }
    return (rgb || '').trim().replace(/\s*,\s*/g, ',') || '59,130,246';
  }

  function resize() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
  }
  resize();
  window.addEventListener('resize', () => {
    resize();
    if (active) createNodes();
  });

  function getBtnCenter() {
    const btn  = document.getElementById('ai-btn');
    const rect = btn.getBoundingClientRect();
    return {
      cx: rect.left + rect.width  / 2,
      cy: rect.top  + rect.height / 2,
    };
  }

  function createNodes() {
    const { cx, cy } = getBtnCenter();
    nodes = [];
    const count = 30;
    for (let i = 0; i < count; i++) {
      const angle  = (Math.PI * 2 / count) * i + (Math.random() - 0.5) * 0.4;
      const targetR = 70 + Math.random() * 170;
      nodes.push({
        x:  cx,
        y:  cy,
        tx: cx + Math.cos(angle) * targetR,
        ty: cy + Math.sin(angle) * targetR,
        ox: cx + Math.cos(angle) * targetR,
        oy: cy + Math.sin(angle) * targetR,
        vx: (Math.random() - 0.5) * 0.25,
        vy: (Math.random() - 0.5) * 0.25,
        r:  1.5 + Math.random() * 2.2,
        phase: Math.random() * Math.PI * 2,
      });
    }
    nodes.push({ x: cx, y: cy, r: 5, isCenter: true });
  }

  function ease(t) {
    return 1 - Math.pow(1 - t, 3);
  }

  function drawFrame(now) {
    const delta = lastTime ? Math.min((now - lastTime) / 1000, 0.05) : 0.016;
    lastTime = now;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    time += 0.013 * (delta / 0.016);

    if (expandProgress < 1) {
      expandProgress = Math.min(expandProgress + delta / 0.6, 1);
    }
    const ep = ease(expandProgress);

    const rgb    = themeRGB;
    const center = nodes[nodes.length - 1];
    const { cx, cy } = getBtnCenter();
    center.x = cx;
    center.y = cy;

    for (const n of nodes) {
      if (n.isCenter) continue;
      if (expandProgress >= 1) {
        n.x += n.vx + Math.sin(time + n.phase) * 0.28;
        n.y += n.vy + Math.cos(time + n.phase) * 0.28;
        n.x += (n.ox - n.x) * 0.009;
        n.y += (n.oy - n.y) * 0.009;
      } else {
        const baseX = cx + (n.ox - cx) * ep;
        const baseY = cy + (n.oy - cy) * ep;
        n.x = baseX + Math.sin(time + n.phase) * 4 * ep;
        n.y = baseY + Math.cos(time + n.phase) * 4 * ep;
      }
    }

    // Связи центр → узлы
    for (let i = 0; i < nodes.length - 1; i++) {
      const a = nodes[i];
      const dCenter = Math.hypot(a.x - cx, a.y - cy);
      if (dCenter < 260) {
        const alpha = (1 - dCenter / 260) * 0.55 * ep;
        const grad  = ctx.createLinearGradient(cx, cy, a.x, a.y);
        grad.addColorStop(0, `rgba(${rgb},${alpha})`);
        grad.addColorStop(1, `rgba(${rgb},0)`);
        ctx.beginPath();
        ctx.moveTo(cx, cy);
        ctx.lineTo(a.x, a.y);
        ctx.strokeStyle = grad;
        ctx.lineWidth   = 0.9;
        ctx.stroke();
      }
    }

    // Связи между узлами
    for (let i = 0; i < nodes.length - 1; i++) {
      for (let j = i + 1; j < nodes.length - 1; j++) {
        const a = nodes[i], b = nodes[j];
        const dist = Math.hypot(a.x - b.x, a.y - b.y);
        if (dist < 95) {
          ctx.beginPath();
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.strokeStyle = `rgba(${rgb},${(1 - dist / 95) * 0.25 * ep})`;
          ctx.lineWidth   = 0.5;
          ctx.stroke();
        }
      }
    }

    // Узлы
    for (const n of nodes) {
      if (n.isCenter) continue;
      const pulse = (0.7 + Math.sin(time * 2.2 + (n.phase || 0)) * 0.3) * ep;
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r * pulse, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(${rgb},0.88)`;
      ctx.fill();
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.r * pulse * 3.5, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(${rgb},0.07)`;
      ctx.fill();
    }

    animFrame = requestAnimationFrame(drawFrame);
  }

  let aiLastResult = null;

  window.toggleAiPanel = function () {
    const panel  = document.getElementById('ai-panel');
    const btn    = document.getElementById('ai-btn');
    const isOpen = !panel.classList.contains('hidden');

    panel.classList.toggle('hidden', isOpen);
    active = !isOpen;
    btn.classList.toggle('active', active);
    canvas.classList.toggle('active', active);

    if (active) {
      themeRGB      = getThemeRGB();
      time          = 0;
      lastTime      = 0;
      expandProgress = 0;
      createNodes();
      animFrame = requestAnimationFrame(drawFrame);
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