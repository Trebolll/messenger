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
    btn._anim = { canvas: cvs, ctx: cvs.getContext('2d'), type, ripples: [], t: 0 };

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

    // — Бегущие λ-частицы вдоль края ———————————————————————————————————
    const perimeter = 2 * (W + H);
    const count     = 6;
    for (let i = 0; i < count; i++) {
      const progress = ((t * 0.18 + i / count + phaseOffset * 0.05) % 1);
      const pos      = progress * perimeter;
      let px, py;

      if (pos < W)             { px = pos;          py = 0; }
      else if (pos < W + H)    { px = W;             py = pos - W; }
      else if (pos < 2 * W + H){ px = W - (pos - W - H); py = H; }
      else                     { px = 0;             py = H - (pos - 2 * W - H); }

      const alpha = 0.15 + 0.2 * Math.abs(Math.sin(t * 2 + i));
      const rotation = t * 2 + i;
      drawMiniLambda(ctx, px, py, 12, alpha, a.type, rotation);
    }

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
  function drawMiniLambda(ctx, x, y, size, alpha, type, rotation = 0) {
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.translate(x, y);
    ctx.rotate(rotation);
    const color = type === 'solid' ? 'rgba(255,255,255,1)' : (getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#3b82f6');
    ctx.strokeStyle = color;
    ctx.lineWidth   = Math.max(0.8, size * 0.1);
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
