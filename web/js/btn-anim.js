// ─── btn-anim.js — анимация кнопок лендинга ───────────────────────────────

(function () {

  // ── Ждём когда лендинг реально виден ────────────────────────────────────
  function init() {
    const btnRegister = document.getElementById('btn-register');
    const btnLogin    = document.getElementById('btn-login');
    if (!btnRegister || !btnLogin) { setTimeout(init, 100); return; }

    setupButton(btnRegister, 'solid');
    setupButton(btnLogin,    'outline');
    startIdleAnimation(btnRegister, btnLogin);
  }

  // ── Canvas-оверлей на каждой кнопке ─────────────────────────────────────
  function setupButton(btn, type) {
    const cvs = document.createElement('canvas');
    cvs.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;pointer-events:none;border-radius:inherit;';
    btn.appendChild(cvs);

    // Создаем хаотичные частицы для каждой кнопки
    const particles = Array.from({ length: 8 }, () => ({
      x: Math.random() * btn.offsetWidth,
      y: Math.random() * btn.offsetHeight,
      vx: (Math.random() - 0.5) * 0.8,
      vy: (Math.random() - 0.5) * 0.8,
      size: 8 + Math.random() * 6,
      phase: Math.random() * Math.PI * 2,
      rotSpeed: (Math.random() - 0.5) * 0.05
    }));

    btn._anim = {
      canvas: cvs,
      ctx: cvs.getContext('2d'),
      type,
      ripples: [],
      particles,
      t: 0
    };

    // Resize canvas px
    function resize() {
      const dpr   = window.devicePixelRatio || 1;
      cvs.width   = btn.offsetWidth  * dpr;
      cvs.height  = btn.offsetHeight * dpr;
      btn._anim.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    resize();
    window.addEventListener('resize', resize);

    // Ripple при клике
    btn.addEventListener('mousedown', e => {
      const rect = btn.getBoundingClientRect();
      btn._anim.ripples.push({
        x     : e.clientX - rect.left,
        y     : e.clientY - rect.top,
        r     : 0,
        maxR  : Math.hypot(btn.offsetWidth, btn.offsetHeight) * 1.1,
        alpha : 0.35,
        born  : performance.now(),
      });
    });

    // Hover — добавляем флаг
    btn.addEventListener('mouseenter', () => { btn._anim.hovered = true;  });
    btn.addEventListener('mouseleave', () => { btn._anim.hovered = false; });
  }

  // ── Главный анимационный цикл для обеих кнопок ──────────────────────────
  function startIdleAnimation(btnR, btnL) {
    let t = 0;

    function frame() {
      t += 0.022;
      drawBtn(btnR, t, 0);
      drawBtn(btnL, t, Math.PI);   // сдвиг фазы для разнообразия
      requestAnimationFrame(frame);
    }

    requestAnimationFrame(frame);
  }

  function drawBtn(btn, t, phaseOffset) {
    const a    = btn._anim;
    if (!a) return;
    const ctx  = a.ctx;
    const W    = btn.offsetWidth;
    const H    = btn.offsetHeight;
    ctx.clearRect(0, 0, W, H);

    // — Хаотично движущиеся λ-частицы ———————————————————————————————————
    a.particles.forEach((p, i) => {
      // 1. Движение с легким "шумом"
      p.x += p.vx + Math.sin(t * 0.8 + p.phase) * 0.4;
      p.y += p.vy + Math.cos(t * 0.8 + p.phase) * 0.4;

      // 2. Отскок от границ
      if (p.x < 0 || p.x > W) p.vx *= -1;
      if (p.y < 0 || p.y > H) p.vy *= -1;

      // 3. Плавное мерцание
      const alpha = 0.2 + 0.25 * Math.sin(t * 1.5 + p.phase);
      const rotation = t * 1.5 + p.phase;

      drawMiniLambda(ctx, p.x, p.y, p.size, alpha, a.type, rotation, a.hovered);

      // 4. Шанс побега частицы (затягивание в черную дыру) — увеличиваем шанс до 0.003
      if (Math.random() < 0.003) {
        const rect = btn.getBoundingClientRect();
        if (window.spawnEscapedParticle) {
          window.spawnEscapedParticle(rect.left + p.x, rect.top + p.y, Math.random() > 0.5);
          // Возвращаем частицу на случайное место в кнопке
          p.x = Math.random() * W;
          p.y = Math.random() * H;
        }
      }
    });

    // — Пульсирующее свечение по центру при hover ————————————————————
    if (a.hovered) {
      const glow   = ctx.createRadialGradient(W / 2, H / 2, 0, W / 2, H / 2, W * 0.6);
      const bright = 0.08 + 0.06 * Math.sin(t * 3 + phaseOffset);
      glow.addColorStop(0, `rgba(59,130,246,${bright})`);
      glow.addColorStop(1, 'rgba(59,130,246,0)');
      ctx.fillStyle = glow;
      ctx.fillRect(0, 0, W, H);
    }

    // — Рябь от кликов —————————————————————————————————————————————————
    const now = performance.now();
    a.ripples = a.ripples.filter(rip => rip.alpha > 0.01);
    a.ripples.forEach(rip => {
      const age   = (now - rip.born) / 600;
      rip.r       = rip.maxR * Math.min(age * 1.4, 1);
      rip.alpha   = 0.35 * Math.max(0, 1 - age);
      ctx.beginPath();
      ctx.arc(rip.x, rip.y, rip.r, 0, Math.PI * 2);
      ctx.strokeStyle = `rgba(255,255,255,${rip.alpha})`;
      ctx.lineWidth   = 2;
      ctx.stroke();
    });
  }

  // ── Маленькая λ ──────────────────────────────────────────────────────────
  function drawMiniLambda(ctx, x, y, size, alpha, type, rotation = 0, hovered = false) {
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.translate(x, y);
    ctx.rotate(rotation);

    // По умолчанию используем акцентный цвет
    let accent = getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#3b82f6';
    let color = accent;

    if (type === 'solid') {
      // Кнопка закрашена акцентным — частицы белые.
      // При наведении становится прозрачной — частицы акцентные.
      color = hovered ? accent : 'rgba(255,255,255,1)';
    } else {
      // Кнопка прозрачная — частицы акцентные.
      // При наведении закрашивается — частицы белые.
      color = hovered ? 'rgba(255,255,255,1)' : accent;
    }

    ctx.strokeStyle = color;
    ctx.lineWidth   = Math.max(1, size * 0.12);
    ctx.lineCap     = 'round';

    ctx.beginPath();
    ctx.moveTo(0, -size * 0.5);
    ctx.lineTo(-size * 0.42, size * 0.5);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(size * 0.08, -size * 0.1);
    ctx.lineTo(size * 0.44, size * 0.5);
    ctx.stroke();

    ctx.restore();
  }

  // ── Запуск ────────────────────────────────────────────────────────────────
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

})();