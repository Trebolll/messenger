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

  let mx = -1000, my = -1000;
  canvas.addEventListener('mousemove', (e) => {
    const rect = canvas.getBoundingClientRect();
    mx = e.clientX - rect.left;
    my = e.clientY - rect.top;
  });
  canvas.addEventListener('mouseleave', () => { mx = -1000; my = -1000; });

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
  const PARTICLE_COUNT = 1500; 

  function createParticle(i, isNew = false) {
    const screenMax = Math.max(W, H);
    // При старте частицы распределяем ближе к видимой зоне
    const r = isNew 
      ? (screenMax * 1.5 + Math.random() * screenMax) 
      : (10 + Math.random() * screenMax * 1.8);
    
    return {
      r: r,
      angle: Math.random() * Math.PI * 2,
      size: 0.25 + Math.random() * 0.8, // Оптимальный размер
      phase: Math.random() * Math.PI * 2,
      isLambda: i % 8 === 0,
      shrinkSpeed: (0.01 + Math.random() * 0.04) * 0.3, 
      tilt: (Math.random() - 0.5) * 0.22,
      pitch: 0.15 + Math.random() * 0.15,
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

  // Экспортируем функцию для добавления внешних частиц (из кнопок)
  window.spawnEscapedParticle = function(pageX, pageY, isLambda) {
    const rect = canvas.getBoundingClientRect();
    const x = (pageX - rect.left) * dpr;
    const y = (pageY - rect.top) * dpr;
    
    // Рассчитываем полярные координаты относительно центра диска
    const dx = x - (W * dpr / 2);
    const dy = (y - (H * dpr / 2)) / 0.28; // учитываем сплюснутость
    
    const r = Math.sqrt(dx * dx + dy * dy) / dpr;
    const angle = Math.atan2(dy, dx);
    
    const tilt = (Math.random() - 0.5) * 0.5;
    particles.push({
      r: r,
      angle: angle,
      size: 0.8 + Math.random() * 1.2,
      phase: Math.random() * Math.PI * 2,
      isLambda: isLambda,
      shrinkSpeed: 0.15 + Math.random() * 0.25, // Плавнее полет
      tilt: tilt,
      pitch: 0.2 + Math.random() * 0.1,
      cosT: Math.cos(tilt), 
      sinT: Math.sin(tilt),
      isEscaped: true 
    });
    
    // Ограничиваем общее число чтобы не тормозило
    if (particles.length > 600) particles.shift();
  };

  function draw() {
    const bodyStyle = getComputedStyle(document.body);
    const bgColor = bodyStyle.getPropertyValue('--bg-main').trim() || '#ffffff';
    const isDark = bgColor.includes('rgb(0,') || bgColor.includes('#0') || bgColor.length > 7 || bgColor === '#1f2937' || bgColor === '#1e1b2e' || bgColor === '#0a1628';

    ctx.globalCompositeOperation = 'source-over';
    ctx.fillStyle = bgColor;
    ctx.globalAlpha = isDark ? 0.3 : 0.5; // Увеличиваем альфу, чтобы шлейф был короче
    ctx.fillRect(0, 0, W * dpr, H * dpr);

    ctx.globalCompositeOperation = isDark ? 'lighter' : 'source-over';
    ctx.globalAlpha = 1.0;
    
    t += 0.0006; // Максимальное замедление времени

    const currentCx = W / 2;
    const currentCy = H / 2;

    // Используем обратный цикл для безопасного удаления частиц (splice)
    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i];
      
      // 1. Динамика вращения (еще плавнее переход скоростей)
      const gravityFactor = 1.6 / Math.pow(Math.max(20, p.r), 0.9);
      p.angle += gravityFactor;
      
      // 2. Гравитационное ускорение падения (делаем мягче)
      const fallSpeed = p.shrinkSpeed * (1 + 120 / (Math.max(0.1, p.r) + 2)); 
      p.r -= fallSpeed;

      // 3. Сингулярность: пропадают мгновенно и гарантированно
      if (p.r < 12 || isNaN(p.r)) {
        if (p.isEscaped) {
          particles.splice(i, 1);
        } else {
          Object.assign(p, createParticle(i, true));
          p.cosT = Math.cos(p.tilt);
          p.sinT = Math.sin(p.tilt);
        }
        continue; // Критично: не рисовать частицу в этом кадре после перемещения/удаления
      }

      const cosA = Math.cos(p.angle);
      const sinA = Math.sin(p.angle);
      const rx = p.r * (1 + 0.01 * Math.sin(t * 1.5 + p.phase));
      const ry = rx * p.pitch; 
      const lx = cosA * rx;
      const ly = sinA * ry;

      let px = currentCx + (lx * p.cosT - ly * p.sinT);
      let py = currentCy + (lx * p.sinT + ly * p.cosT);

      const doppler = 1.0 - (sinA * 0.4); 
      const viewAlpha = Math.min(1, (1800 - p.r) / 700); 
      const distRatio = 1 - (Math.min(p.r, 1200) / 1200);
      let alpha = (0.15 + 0.7 * distRatio) * doppler * viewAlpha;
      let size = p.size * (1.0 + distRatio * 0.4);

      // Более агрессивное уменьшение и исчезновение при приближении к сингулярности
      if (p.r < 100) {
        const factor = Math.max(0, (p.r - 12) / 88);
        size *= factor;
        alpha *= factor;
      }

      if (sinA < 0) alpha *= 0.75;

      const dx = px - mx;
      const dy = py - my;
      const mouseDist = Math.sqrt(dx * dx + dy * dy);
      if (mouseDist < 80 && alpha > 0.02) {
        const glow = Math.pow(1 - mouseDist / 80, 2) * Math.min(1, alpha * 4); 
        ctx.save();
        const coreGrad = ctx.createRadialGradient(px, py, 0, px, py, size * 2);
        coreGrad.addColorStop(0, `rgba(255,255,255,${glow})`); 
        coreGrad.addColorStop(0.5, accentColor(glow));
        coreGrad.addColorStop(1, 'transparent');
        ctx.beginPath();
        ctx.arc(px, py, size * 2.5, 0, Math.PI * 2);
        ctx.fillStyle = coreGrad;
        ctx.fill();

        const auraGrad = ctx.createRadialGradient(px, py, size * 2, px, py, size * 15);
        auraGrad.addColorStop(0, accentColor(glow * 0.4));
        auraGrad.addColorStop(1, accentColor(0));
        ctx.beginPath();
        ctx.arc(px, py, size * 15, 0, Math.PI * 2);
        ctx.fillStyle = auraGrad;
        if (isDark) ctx.globalCompositeOperation = 'lighter';
        ctx.fill();

        if (glow > 0.3) {
          ctx.beginPath();
          ctx.strokeStyle = isDark ? `rgba(255,255,255,${glow * 0.3})` : accentColor(glow * 0.4);
          ctx.lineWidth = 0.5;
          ctx.moveTo(px, py - size * 12 * glow);
          ctx.lineTo(px, py + size * 12 * glow);
          ctx.moveTo(px - size * 12 * glow, py);
          ctx.lineTo(px + size * 12 * glow, py);
          ctx.stroke();
        }
        ctx.restore();
      }

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
    }

    raf = requestAnimationFrame(draw);
  }

  draw();

  // ── Стоп при уходе со страницы ────────────────────────────────────────
  document.addEventListener('visibilitychange', () => {
    if (document.hidden) { cancelAnimationFrame(raf); }
    else { t = 0; draw(); }
  });
})();
