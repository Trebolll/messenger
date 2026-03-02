// ─── lambda-anim.js — анимация λ на лендинге ─────────────────────────────

(function () {
  const canvas = document.getElementById('lambda-canvas');
  if (!canvas) return;

  const ctx = canvas.getContext('2d');
  let W, H, cx, cy, dpr;
  let t = 0;
  let raf;

  // ── Размеры ────────────────────────────────────────────────────────────
  function resize() {
    dpr = window.devicePixelRatio || 1;
    W = canvas.offsetWidth;
    H = canvas.offsetHeight;
    canvas.width  = W * dpr;
    canvas.height = H * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    cx = W / 2;
    cy = H / 2;
  }

  window.addEventListener('resize', resize);
  resize();

  // ── Цвет из CSS-переменной (поддержка тем) ────────────────────────────
  function accentColor(alpha) {
    return `rgba(59,130,246,${alpha})`;   // blue-500
  }

  // ── Рисуем λ через Path2D с параметрическими деформациями ────────────
  function drawLambda(x, y, size, skewX, skewY, alpha) {
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.translate(x, y);
    ctx.transform(1, skewY, skewX, 1, 0, 0);

    const s  = size;
    const sw = Math.max(1, s * 0.07);   // толщина штриха

    ctx.strokeStyle = accentColor(1);
    ctx.lineWidth   = sw;
    ctx.lineCap     = 'round';
    ctx.lineJoin    = 'round';

    // Левый штрих: сверху-центр → влево-низ
    ctx.beginPath();
    ctx.moveTo(0, -s * 0.5);
    ctx.lineTo(-s * 0.42, s * 0.5);
    ctx.stroke();

    // Правый штрих: чуть правее центра-верх → вправо-низ
    ctx.beginPath();
    ctx.moveTo(s * 0.08, -s * 0.1);
    ctx.lineTo(s * 0.44, s * 0.5);
    ctx.stroke();

    ctx.restore();
  }

  // ── Частицы ───────────────────────────────────────────────────────────
  const PARTICLE_COUNT = 28;
  const particles = Array.from({ length: PARTICLE_COUNT }, (_, i) => ({
    angle : (i / PARTICLE_COUNT) * Math.PI * 2,
    r     : 80 + Math.random() * 60,
    speed : 0.003 + Math.random() * 0.004,
    size  : 1.5 + Math.random() * 2.5,
    phase : Math.random() * Math.PI * 2,
  }));

  // ── Главный цикл ──────────────────────────────────────────────────────
  function draw() {
    ctx.clearRect(0, 0, W, H);
    t += 0.016;

    // — Орбитальные частицы —
    particles.forEach(p => {
      p.angle += p.speed;
      const rx    = p.r * (1 + 0.12 * Math.sin(t * 0.7 + p.phase));
      const ry    = p.r * 0.55 * (1 + 0.12 * Math.cos(t * 0.5 + p.phase));
      const px    = cx + Math.cos(p.angle) * rx;
      const py    = cy + Math.sin(p.angle) * ry;
      const fade  = 0.25 + 0.45 * Math.abs(Math.sin(t * 1.1 + p.phase));
      ctx.beginPath();
      ctx.arc(px, py, p.size, 0, Math.PI * 2);
      ctx.fillStyle = accentColor(fade);
      ctx.fill();
    });

    // — Пульсирующие кольца —
    for (let i = 0; i < 3; i++) {
      const phase   = (t * 0.6) + i * (Math.PI * 2 / 3);
      const radius  = 72 + 22 * Math.sin(phase);
      const ringA   = 0.06 + 0.06 * Math.abs(Math.sin(phase));
      ctx.beginPath();
      ctx.arc(cx, cy, radius, 0, Math.PI * 2);
      ctx.strokeStyle = accentColor(ringA);
      ctx.lineWidth   = 1.5;
      ctx.stroke();
    }

    // — Волновые линии под лямбдой —
    for (let wave = 0; wave < 2; wave++) {
      ctx.beginPath();
      const waveAlpha = 0.07 + wave * 0.04;
      ctx.strokeStyle = accentColor(waveAlpha);
      ctx.lineWidth   = 1;
      for (let px = 0; px <= W; px += 3) {
        const dist = (px - cx) / (W * 0.4);
        const py2  = cy + 55 + wave * 22
                    + 14 * Math.sin(dist * 4 + t * 1.2 + wave * 1.5)
                    * Math.exp(-dist * dist * 0.8);
        px === 0 ? ctx.moveTo(px, py2) : ctx.lineTo(px, py2);
      }
      ctx.stroke();
    }

    // — Главная λ (пульсирует + лёгкий skew) —
    const pulse = 1 + 0.06 * Math.sin(t * 1.8);
    const skewX = 0.04 * Math.sin(t * 0.9);
    const skewY = 0.02 * Math.cos(t * 1.3);
    drawLambda(cx, cy, 110 * pulse, skewX, skewY, 1);

    // — Отражение (призрак снизу) —
    ctx.save();
    ctx.translate(cx, cy + 110 * pulse * 0.6);
    ctx.scale(1, -0.22);
    ctx.globalAlpha = 0.12;
    ctx.filter = 'blur(2px)';
    ctx.restore();
    drawLambda(cx, cy + 68, 110 * pulse * 0.55, skewX, skewY, 0.1);

    // — Вторичная λ (медленно вращается по кругу) —
    const orbitAngle = t * 0.25;
    const orbitR     = 130;
    const ox = cx + Math.cos(orbitAngle) * orbitR;
    const oy = cy + Math.sin(orbitAngle) * orbitR * 0.35;
    drawLambda(ox, oy, 20, 0, 0, 0.13 + 0.07 * Math.sin(t + 1));

    raf = requestAnimationFrame(draw);
  }

  draw();

  // ── Стоп при уходе со страницы ────────────────────────────────────────
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) { cancelAnimationFrame(raf); }
    else { t = 0; draw(); }
  });
})();
