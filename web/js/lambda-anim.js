// ─── lambda-anim.js — анимация λ на лендинге ─────────────────────────────

(function () {
  const canvas = document.getElementById('lambda-canvas');
  if (!canvas) return;

  const ctx = canvas.getContext('2d', { alpha: false }); // alpha:false — браузер не тратит время на композитинг прозрачности
  let W, H, cx, cy, dpr;
  let t = 0;
  let raf;
  let lastTime = 0;

  // ── Кэш цвета — обновляем только при смене темы, не каждый кадр ──────
  let _accentR = 59, _accentG = 130, _accentB = 246;
  let _bgColor  = '#ffffff';
  let _isDark   = false;
  let _colorCacheFrame = -999; // номер кадра последнего обновления

  function refreshColorCache() {
    const style   = getComputedStyle(document.body);
    const accent  = style.getPropertyValue('--accent-color').trim() || '#3b82f6';
    _bgColor      = style.getPropertyValue('--bg-main').trim()      || '#ffffff';
    _isDark       = _bgColor === '#1f2937' || _bgColor === '#1e1b2e' ||
        _bgColor === '#0a1628' || _bgColor.startsWith('#0') ||
        _bgColor.includes('rgb(0,');
    if (accent.startsWith('#')) {
      _accentR = parseInt(accent.slice(1, 3), 16);
      _accentG = parseInt(accent.slice(3, 5), 16);
      _accentB = parseInt(accent.slice(5, 7), 16);
    }
  }
  refreshColorCache(); // один раз при старте

  // Обновляем кэш при смене темы (MutationObserver на class body)
  new MutationObserver(refreshColorCache).observe(document.body, { attributes: true, attributeFilter: ['class'] });

  function accentColor(alpha) {
    return `rgba(${_accentR},${_accentG},${_accentB},${alpha})`;
  }

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

  // ── Частицы ───────────────────────────────────────────────────────────
  const PARTICLE_COUNT = 1500;
  // Используем TypedArrays — быстрее обычных объектов для числовых данных
  const pR          = new Float32Array(PARTICLE_COUNT);
  const pAngle      = new Float32Array(PARTICLE_COUNT);
  const pSize       = new Float32Array(PARTICLE_COUNT);
  const pPhase      = new Float32Array(PARTICLE_COUNT);
  const pShrink     = new Float32Array(PARTICLE_COUNT);
  const pTilt       = new Float32Array(PARTICLE_COUNT);
  const pPitch      = new Float32Array(PARTICLE_COUNT);
  const pCosT       = new Float32Array(PARTICLE_COUNT);
  const pSinT       = new Float32Array(PARTICLE_COUNT);
  const pIsLambda   = new Uint8Array(PARTICLE_COUNT);

  function initParticle(i, isNew) {
    const screenMax = Math.max(W, H);
    pR[i]        = isNew
        ? (screenMax * 0.4 + Math.random() * screenMax * 0.5) // СПАВН СУЩЕСТВЕННО БЛИЖЕ: диск будет плотнее
        : (10 + Math.random() * screenMax * 1.5);
    pAngle[i]    = Math.random() * Math.PI * 2;
    pSize[i]     = 0.27 + Math.random() * 0.86; // УВЕЛИЧЕНО НА 7% (было 0.25+0.8)
    pPhase[i]    = Math.random() * Math.PI * 2;
    pShrink[i]   = (0.12 + Math.random() * 0.2) * 0.25;  // ЕЩЕ БЫСТРЕЕ: цикл жизни в 2-4 минуты
    pTilt[i]     = (Math.random() - 0.5) * 0.22;
    pPitch[i]    = 0.15 + Math.random() * 0.15;
    pCosT[i]     = Math.cos(pTilt[i]);
    pSinT[i]     = Math.sin(pTilt[i]);
    pIsLambda[i] = (i % 8 === 0) ? 1 : 0;
  }

  for (let i = 0; i < PARTICLE_COUNT; i++) initParticle(i, false);

  // Предвычисленный кэш градиентов для mouse-glow (3 уровня яркости)
  // Пересоздаём только при resize
  let _glowGrad = null;
  let _glowGradPx = -1, _glowGradPy = -1;

  // ── Глобальные константы анимации ─────────────────────────────────────
  const DISK_TILT = 0.33; // Наклон диска относительно горизонта (~18°)
  const MOUSE_DIST_SQ = 80 * 80;

  window.spawnEscapedParticle = function(pageX, pageY, isLambda) {
    const rect  = canvas.getBoundingClientRect();
    const x     = (pageX - rect.left) * dpr;
    const y     = (pageY - rect.top)  * dpr;
    
    // Берем случайную частицу из основного массива и "телепортируем" её
    const i = Math.floor(Math.random() * PARTICLE_COUNT);
    
    // Пересчитываем экранные координаты (CSS px) обратно в r и angle
    const tx = (x - (W * dpr / 2)) / dpr;
    const ty = (y - (H * dpr / 2)) / dpr;
    
    // Учитываем наклон диска (инвертируем сдвиг от DISK_TILT)
    const dy_corr = ty + tx * DISK_TILT;
    const pitch   = 0.15 + Math.random() * 0.15; 
    
    pR[i]        = Math.sqrt(tx * tx + Math.pow(dy_corr / pitch, 2));
    pAngle[i]    = Math.atan2(dy_corr / pitch, tx);
    pSize[i]     = 0.7 + Math.random() * 0.8; // ЕЩЕ БОЛЬШЕ (было 0.32+0.43), чтобы их было четко видно
    pTilt[i]     = 0; 
    pCosT[i]     = 1;
    pSinT[i]     = 0;
    pPitch[i]    = pitch;
    pIsLambda[i] = isLambda ? 1 : 0;
    pShrink[i]   = 0.08 + Math.random() * 0.12; // ГАРАНТИРУЕМ дрейф к центру для "сбежавших"
  };

  // ── Главный цикл ───────────────────────────────────────────────────────

  function draw(now) {
    const delta = lastTime ? Math.min((now - lastTime) / 1000, 0.05) : 0.001;
    const dtFactor = delta / 0.016;
    lastTime = now;

    // Фон
    ctx.globalCompositeOperation = 'source-over';
    ctx.fillStyle   = _bgColor;
    ctx.globalAlpha = _isDark ? 0.3 : 0.5;
    ctx.fillRect(0, 0, W, H);

    ctx.globalCompositeOperation = _isDark ? 'lighter' : 'source-over';
    ctx.globalAlpha = 1.0;

    t += 0.0006 * dtFactor;

    // ── Основной массив частиц ─────────────────────────────────────────
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      // Физика
      const gravityFactor = 1.6 / Math.pow(Math.max(20, pR[i]), 0.9);
      // Завихрение у сингулярности: при r<60 угловая скорость резко растёт
      const swirlBoost = pR[i] < 70          // РУЧКА: радиус начала завихрения (px)
          ? 1 + Math.pow((70 - pR[i]) / 70, 2) * 10.5  // РУЧКА: сила (8=умеренно, 15=агрессивно)
          : 1.0;
      pAngle[i] += gravityFactor * swirlBoost * dtFactor;

      // Притяжение: базовый дрейф + ускорение к центру. 
      // Добавил минимальный порог 0.05, чтобы никто не "зависал"
      const gravPull = 1.0 + (500 / (pR[i] + 10)); 
      const fallSpeed = (pShrink[i] + 0.05) * dtFactor * gravPull;
      pR[i] -= fallSpeed;

      if (pR[i] < 12) {
        initParticle(i, true);
        continue;
      }

      const cosA = Math.cos(pAngle[i]);
      const sinA = Math.sin(pAngle[i]);
      const rx   = pR[i] * (1 + 0.01 * Math.sin(t * 1.5 + pPhase[i]));
      const ry   = rx * pPitch[i];
      const lx   = cosA * rx;
      const ly   = sinA * ry;

      const px = cx + (lx * pCosT[i] - ly * pSinT[i]);
      // Наклон диска — добавляем сдвиг по Y пропорционально X
      const py = cy + (lx * pSinT[i] + ly * pCosT[i]) - (px - cx) * DISK_TILT;

      const r = pR[i];

      // Доплер: частицы летящие "от нас" чуть темнее — но не режем их в ноль
      const doppler   = 0.7 + 0.3 * (1.0 - sinA * 0.4); // диапазон 0.58..1.0 вместо 0.6..1.4
      // Видимость по расстоянию: плавно появляются начиная с r=1800
      const viewAlpha = r > 1800 ? 0 : Math.min(1, (1800 - r) / 600);
      // Яркость по близости к центру: чем ближе — тем ярче
      const distRatio = 1 - (Math.min(r, 1200) / 1200);
      let alpha = (0.12 + 0.75 * distRatio) * doppler * viewAlpha;
      let size  = pSize[i] * (1.0 + distRatio * 0.5);

      // Затухание только в последние пиксели перед сингулярностью
      if (r < 20) {
        const factor = Math.max(0, (r - 12) / 8);
        size  *= factor;
        alpha *= factor;
      }

      // sinA < 0 — обратная сторона диска, чуть темнее но не убиваем
      if (sinA < 0) alpha *= 0.8;

      // Не рисуем только совсем невидимое — физика продолжает работать всегда
      if (alpha < 0.002) continue;

      // Mouse glow
      const ddx = px - mx;
      const ddy = py - my;
      const distSq = ddx * ddx + ddy * ddy;
      if (distSq < MOUSE_DIST_SQ && alpha > 0.01) {
        const mouseDist = Math.sqrt(distSq);
        const proximity = 1 - mouseDist / 80;           // 0..1, 1 = точно под курсором
        const glow = Math.pow(proximity, 1.6);           // более плавная кривая (было ^2)

        if (glow > 0.03) {
          ctx.save();
          ctx.globalCompositeOperation = 'lighter';      // additive — даёт реальную яркость

          // Ядро — ослепительно белое в центре
          const coreR = size * 3.5;
          const coreGrad = ctx.createRadialGradient(px, py, 0, px, py, coreR);
          coreGrad.addColorStop(0,   `rgba(255,255,255,${glow * 0.95})`);
          coreGrad.addColorStop(0.2, `rgba(255,255,255,${glow * 0.6})`);
          coreGrad.addColorStop(0.5, accentColor(glow * 0.8));
          coreGrad.addColorStop(1,   accentColor(0));
          ctx.beginPath();
          ctx.arc(px, py, coreR, 0, Math.PI * 2);
          ctx.fillStyle = coreGrad;
          ctx.fill();

          // Широкая аура вокруг — цвет акцента
          const auraR = size * 28 * glow + 18;
          const auraGrad = ctx.createRadialGradient(px, py, coreR, px, py, auraR);
          auraGrad.addColorStop(0,   accentColor(glow * 0.35));
          auraGrad.addColorStop(0.4, accentColor(glow * 0.12));
          auraGrad.addColorStop(1,   accentColor(0));
          ctx.beginPath();
          ctx.arc(px, py, auraR, 0, Math.PI * 2);
          ctx.fillStyle = auraGrad;
          ctx.fill();

          // Блик-крест (звёздочка) — только когда совсем близко
          if (glow > 0.45) {
            const arm = (size * 18 + 10) * glow;
            ctx.lineWidth = size * 1.2;
            ctx.lineCap = 'round';
            // Горизонталь
            const hGrad = ctx.createLinearGradient(px - arm, py, px + arm, py);
            hGrad.addColorStop(0,   accentColor(0));
            hGrad.addColorStop(0.5, `rgba(255,255,255,${glow * 0.8})`);
            hGrad.addColorStop(1,   accentColor(0));
            ctx.beginPath();
            ctx.moveTo(px - arm, py); ctx.lineTo(px + arm, py);
            ctx.strokeStyle = hGrad;
            ctx.stroke();
            // Вертикаль
            const vGrad = ctx.createLinearGradient(px, py - arm, px, py + arm);
            vGrad.addColorStop(0,   accentColor(0));
            vGrad.addColorStop(0.5, `rgba(255,255,255,${glow * 0.8})`);
            vGrad.addColorStop(1,   accentColor(0));
            ctx.beginPath();
            ctx.moveTo(px, py - arm); ctx.lineTo(px, py + arm);
            ctx.strokeStyle = vGrad;
            ctx.stroke();
          }

          ctx.restore();
        }
      }

      if (pIsLambda[i]) {
        drawLambda(px, py, size * 6, pAngle[i], alpha * (0.7 + 0.3 * Math.sin(t * 2 + pPhase[i])));
      } else {
        ctx.beginPath();
        ctx.arc(px, py, size, 0, Math.PI * 2);
        ctx.fillStyle = accentColor(alpha);
        ctx.fill();
      }

      // линзирование убрано — создавало дублирование диска
    }

    raf = requestAnimationFrame(draw);
  }

  draw();

  document.addEventListener('visibilitychange', () => {
    if (document.hidden) { cancelAnimationFrame(raf); }
    else { lastTime = 0; draw(); }
  });
})();