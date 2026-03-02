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
    const computed = getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#3b82f6';
    if (computed.startsWith('#')) {
      const r = parseInt(computed.slice(1, 3), 16);
      const g = parseInt(computed.slice(3, 5), 16);
      const b = parseInt(computed.slice(5, 7), 16);
      return `rgba(${r},${g},${b},${alpha})`;
    }
    return computed.replace('rgb', 'rgba').replace(')', `,${alpha})`);
  }

  // ── Рисуем λ ──────────────────────────────────────────────────────────
  function drawLambda(x, y, size, angle, alpha) {
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.translate(x, y);
    ctx.rotate(angle);

    const s  = size;
    const sw = Math.max(0.8, s * 0.08);

    ctx.strokeStyle = accentColor(1);
    ctx.lineWidth   = sw;
    ctx.lineCap     = 'round';

    ctx.beginPath();
    ctx.moveTo(0, -s * 0.5);
    ctx.lineTo(-s * 0.4, s * 0.5);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(s * 0.1, -s * 0.1);
    ctx.lineTo(s * 0.4, s * 0.5);
    ctx.stroke();

    ctx.restore();
  }

  // ── Частицы (Физически корректный аккреционный диск) ───────────────────
  const PARTICLE_COUNT = 400; 

  function createParticle(i, isNew = false) {
    const spawnRadius = Math.max(W, H) * 0.8;
    const r = isNew ? (spawnRadius + Math.random() * spawnRadius) : (35 + Math.random() * spawnRadius * 1.5);
    
    return {
      r: r,
      angle: Math.random() * Math.PI * 2,
      size: 0.4 + Math.random() * 1.8,
      phase: Math.random() * Math.PI * 2,
      isLambda: i % 7 === 0,
      shrinkSpeed: (0.02 + Math.random() * 0.08) * 0.4,
      // Индивидуальные параметры орбиты для каждой частицы
      tilt: (Math.random() - 0.5) * 0.3,    // Наклон орбиты (угол)
      pitch: 0.2 + Math.random() * 0.12,   // Сплюснутость орбиты (перспектива)
      cosT: 0, 
      sinT: 0
    };
  }

  // Предрассчитываем тригонометрию для наклона
  const particles = Array.from({ length: PARTICLE_COUNT }, (v, i) => {
    const p = createParticle(i);
    p.cosT = Math.cos(p.tilt);
    p.sinT = Math.sin(p.tilt);
    return p;
  });

  function draw() {
    const bodyStyle = getComputedStyle(document.body);
    const bgColor = bodyStyle.getPropertyValue('--bg-main').trim() || '#ffffff';
    const isDark = bgColor.includes('rgb(0,') || bgColor.includes('#0') || bgColor.length > 7 || bgColor === '#1f2937' || bgColor === '#1e1b2e' || bgColor === '#0a1628';

    ctx.globalCompositeOperation = 'source-over';
    ctx.fillStyle = bgColor;
    ctx.globalAlpha = isDark ? 0.15 : 0.3; 
    ctx.fillRect(0, 0, W * dpr, H * dpr);

    ctx.globalCompositeOperation = isDark ? 'lighter' : 'source-over';
    ctx.globalAlpha = 1.0;
    
    t += 0.0012;

    const currentCx = W / 2;
    const currentCy = H / 2;

    particles.forEach((p, i) => {
      // 1. Динамика вращения (законы Кеплера): чем ближе, тем быстрее
      // Увеличил влияние массы в центре для "эффекта гравитации"
      const gravityFactor = 2.2 / Math.pow(p.r, 0.95);
      p.angle += gravityFactor;
      
      // 2. Гравитационное ускорение падения: чем ближе к центру, тем сильнее тянет
      const fallSpeed = p.shrinkSpeed * (1 + 150 / (p.r + 10));
      p.r -= fallSpeed;

      // 3. Сингулярность: пропадают почти в самой точке (радиус 5 вместо 26)
      if (p.r < 5) {
        Object.assign(p, createParticle(i, true));
        p.cosT = Math.cos(p.tilt);
        p.sinT = Math.sin(p.tilt);
      }

      const cosA = Math.cos(p.angle);
      const sinA = Math.sin(p.angle);

      // Локальные координаты на плоскости орбиты частицы
      const rx = p.r * (1 + 0.01 * Math.sin(t * 1.5 + p.phase));
      const ry = rx * p.pitch; 

      const lx = cosA * rx;
      const ly = sinA * ry;

      // Поворачиваем локальные координаты согласно наклону (tilt) этой конкретной частицы
      let px = currentCx + (lx * p.cosT - ly * p.sinT);
      let py = currentCy + (lx * p.sinT + ly * p.cosT);

      const doppler = 1.0 - (sinA * 0.4); 
      const viewAlpha = Math.min(1, (1200 - p.r) / 500); 
      const distRatio = 1 - (Math.min(p.r, 800) / 800);
      let alpha = (0.02 + 0.85 * distRatio) * doppler * viewAlpha;
      let size = p.size * (0.6 + distRatio * 1.4);

      // Более мягкое скрытие задней части
      if (sinA < 0) alpha *= 0.5;

      if (p.isLambda) {
        const s = size * 6;
        const a = alpha * (0.7 + 0.3 * Math.sin(t * 2 + p.phase));
        drawLambda(px, py, s, p.angle, a);
      } else {
        ctx.beginPath();
        ctx.arc(px, py, size, 0, Math.PI * 2);
        ctx.fillStyle = accentColor(alpha);
        ctx.fill();
      }

      // Линзирование только для тех, кто реально за объектом
      if (sinA < -0.2) {
        const lensY = (Math.abs(sinA) * p.r * 0.6);
        const lensAlpha = alpha * 0.35 * (1 - lensY / 200);
        if (lensAlpha > 0) {
           ctx.beginPath();
           ctx.arc(currentCx + cosA * rx * 0.8, currentCy - 28 - lensY * 0.4, size * 0.6, 0, Math.PI * 2);
           ctx.fillStyle = accentColor(lensAlpha);
           ctx.fill();
        }
      }
    });

    raf = requestAnimationFrame(draw);
  }

  draw();

  // ── Стоп при уходе со страницы ────────────────────────────────────────
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) { cancelAnimationFrame(raf); }
    else { t = 0; draw(); }
  });
})();
