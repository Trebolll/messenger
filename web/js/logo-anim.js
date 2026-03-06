// ─── logo-anim.js — с фазой gathering (сведение в точку) ───
(function () {
  const canvas = document.getElementById('logo-canvas');
  if (!canvas) return;

  const ctx = canvas.getContext('2d');
  let w, h, cx, cy, dpr;
  let t = 0;
  let lastTime = 0;
  let state = 'single';        // single, splitting, orbiting, gathering, merging
  let stateTimer = 0;
  let flashAlpha = 0;
  let flashRings = [];

  // ========== НАСТРОЙКИ ==========
  const ENABLE_TRIPLE_SPLIT = true;
  const ENABLE_PARTICLES = false;    // пока отключим для чистоты
  const ENABLE_TRAILS = true;
  const ENABLE_MICRO = true;
  const ENABLE_GLOW = true;
  // ===============================

  const MAX_PARTICLES = 30;
  const TRAIL_LENGTH = 4;
  const MAX_MICRO = 5;
  const GLOW_RADIUS_FACTOR = 5;

  let particles = [];
  let trails = [];
  let microLambdas = [];
  let trailFrameSkip = 0;

  function resize() {
    dpr = window.devicePixelRatio || 1;
    const rect = canvas.parentNode.getBoundingClientRect();
    w = rect.width;
    h = rect.height;
    cx = w / 2;
    cy = h / 2;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  }
  window.addEventListener('resize', resize);
  resize();

  function accentColor(alpha) {
    const a = isNaN(alpha) ? 0 : Math.min(1, Math.max(0, alpha));
    const computed = getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#3b82f6';
    if (computed.startsWith('#')) {
      const r = parseInt(computed.slice(1, 3), 16);
      const g = parseInt(computed.slice(3, 5), 16);
      const b = parseInt(computed.slice(5, 7), 16);
      return `rgba(${r},${g},${b},${a})`;
    }
    return computed.replace('rgb', 'rgba').replace(')', `,${a})`);
  }

  function drawLambda(x, y, size, angle, alpha, scaleX = 1, fill = false) {
    ctx.save();
    ctx.globalAlpha = alpha;
    ctx.translate(x, y);
    ctx.rotate(angle);
    ctx.scale(scaleX, 1);

    const s = size;
    const sw = Math.max(1.5, s * 0.12);
    ctx.strokeStyle = accentColor(1);
    ctx.lineWidth = sw;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    ctx.beginPath();
    ctx.moveTo(0, -s * 0.5);
    ctx.lineTo(-s * 0.4, s * 0.5);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(s * 0.1, -s * 0.1);
    ctx.lineTo(s * 0.4, s * 0.5);
    ctx.stroke();

    if (fill && ENABLE_GLOW) {
      ctx.globalAlpha = alpha * 0.1;
      ctx.fillStyle = accentColor(1);
      ctx.fill();
    }
    ctx.restore();
  }

  // ---- ЧАСТИЦЫ (заглушки) ----
  function spawnParticles(count, x, y, baseSpeed = 2, sizeScale = 1) { /* не используется */ }
  function updateParticles(delta) {}
  function drawParticles() {}

  // ---- СЛЕДЫ ----
  function updateTrails(positions) {
    if (!ENABLE_TRAILS) return;
    if (!trails.length) for (let i = 0; i < 3; i++) trails[i] = [];
    for (let i = 0; i < 3; i++) {
      if (!positions[i]) continue;
      trails[i].push({ x: positions[i].x, y: positions[i].y, life: 1.0 });
      if (trails[i].length > TRAIL_LENGTH) trails[i].shift();
    }
    trails.forEach(t => t.forEach(p => p.life *= 0.9));
  }

  function drawTrails() {
    if (!ENABLE_TRAILS) return;
    trails.forEach((t, idx) => {
      for (let i = 0; i < t.length; i++) {
        const p = t[i];
        const alpha = p.life * 0.2 * (i / t.length);
        ctx.save();
        ctx.globalAlpha = alpha;
        ctx.fillStyle = accentColor(alpha);
        ctx.beginPath();
        ctx.arc(p.x, p.y, 1 + idx, 0, Math.PI * 2);
        ctx.fill();
        ctx.restore();
      }
    });
  }

  // ---- МИКРО-ЛЯМБДЫ ----
  function spawnMicroLambda(x, y, angle) {
    if (!ENABLE_MICRO || microLambdas.length >= MAX_MICRO) return;
    microLambdas.push({
      x, y,
      baseX: x, baseY: y,
      angle: angle,
      life: 1.0,
      decay: 0.02,
      offset: Math.random() * Math.PI * 2,
    });
  }

  function updateMicroLambdas(delta) {
    if (!ENABLE_MICRO) return;
    microLambdas = microLambdas.filter(m => m.life > 0);
    microLambdas.forEach(m => {
      m.life -= m.decay * delta * 60;
      const dist = 12 * (1 - m.life);
      m.x = m.baseX + Math.cos(t * 3 + m.offset) * dist;
      m.y = m.baseY + Math.sin(t * 3 + m.offset) * dist;
    });
  }

  function drawMicroLambdas() {
    if (!ENABLE_MICRO) return;
    microLambdas.forEach(m => {
      drawLambda(m.x, m.y, 6, m.angle + t, m.life * 0.3);
    });
  }

  // ---- ОСНОВНОЙ UPDATE ----
  function update(delta) {
    t += delta;
    stateTimer += delta;

    if (state === 'single') {
      if (stateTimer > 8 + Math.random() * 4) {
        state = 'splitting';
        stateTimer = 0;
        flashAlpha = 0.9;
      }
    } else if (state === 'splitting') {
      if (stateTimer > 1.4) {
        state = 'orbiting';
        stateTimer = 0;
        trails = [];
        microLambdas = [];
        flashRings.push({ radius: 5, speed: 2, life: 0.8 });
      }
      // микро-лямбды в splitting почти не появляются
    } else if (state === 'orbiting') {
      if (stateTimer > 4.0) {  // чуть короче, чтобы gathering был заметнее
        state = 'gathering';
        stateTimer = 0;
        // сбросим трейлы, они в gathering не нужны
        trails = [];
        flashAlpha = 0.5;
      } else {
        if (ENABLE_MICRO && Math.random() < 0.02) {
          const progress = stateTimer / 4.0;
          const angle = t * 4;
          const r = 25 + Math.sin(t)*6;
          const x = cx + Math.cos(angle) * r;
          const y = cy + Math.sin(angle) * r;
          spawnMicroLambda(x, y, angle);
        }
      }
    } else if (state === 'gathering') {
      // Длительность gathering — 1 секунда, за это они сходятся в центр
      if (stateTimer > 1.0) {
        state = 'merging';
        stateTimer = 0;
        flashAlpha = 1.2;
        flashRings.push({ radius: 8, speed: 3, life: 1.0 });
      } else {
        // Можно добавить микро-лямбды, но очень редко
        if (ENABLE_MICRO && Math.random() < 0.01) {
          spawnMicroLambda(cx + (Math.random()-0.5)*20, cy + (Math.random()-0.5)*20, Math.random()*Math.PI*2);
        }
      }
    } else if (state === 'merging') {
      if (stateTimer > 1.6) {
        state = 'single';
        stateTimer = 0;
        flashAlpha = 1.5;
        flashRings.push({ radius: 10, speed: 2, life: 1.0 });
        flashRings.push({ radius: 5, speed: 4, life: 0.7 });
      }
    }

    if (flashAlpha > 0) flashAlpha -= 0.025;

    flashRings = flashRings.filter(r => r.life > 0);
    flashRings.forEach(r => {
      r.radius += r.speed * delta * 60;
      r.life -= 0.02 * delta * 60;
    });

    updateMicroLambdas(delta);
  }

  // ---- ОТРИСОВКА ----
  function draw(now) {
    if (!cx || !cy) return requestAnimationFrame(draw);
    const delta = lastTime ? Math.min((now - lastTime) / 1000, 0.05) : 0.016;
    lastTime = now;
    ctx.clearRect(0, 0, w, h);
    update(delta);

    const baseSize = 34;

    // SINGLE
    if (state === 'single') {
      const appearance = Math.min(1, stateTimer / 1.5);
      const scale = 0.8 + 0.2 * appearance;
      ctx.save();
      ctx.translate(cx, cy);
      ctx.scale(scale, scale);
      drawLambda(0, 0, baseSize, Math.sin(t * 0.5) * 0.1, appearance, 1, true);
      ctx.restore();
    }

    // SPLITTING
    else if (state === 'splitting') {
      const p = Math.min(1, stateTimer / 1.4);
      const offset = p * 22;

      drawLambda(cx - offset, cy, baseSize, -p * 0.3, 1 - p * 0.2);
      drawLambda(cx + offset, cy, baseSize, p * 0.3, 1 - p * 0.2);
      drawLambda(cx, cy - offset * 0.9, baseSize * 0.85, 0, 1 - p * 0.2);
    }

    // ORBITING
    else if (state === 'orbiting') {
      const progress = Math.min(1, stateTimer / 4.0);
      const accel = Math.pow(progress, 2.5);
      const speed = 4 + accel * 40;
      const angle = stateTimer * speed;
      const radius = 15 * (1 - progress * 0.6);

      const positions = [];
      for (let i = 0; i < 3; i++) {
        const a = angle + (i * Math.PI * 2 / 3);
        let x = cx + Math.cos(a) * radius;
        let y = cy + Math.sin(a) * radius;
        if (progress > 0.7) {
          x += (Math.random() - 0.5) * 2;
          y += (Math.random() - 0.5) * 2;
        }
        drawLambda(x, y, baseSize * 0.7, a * 0.5 + Math.sin(angle)*0.1, 0.7 + progress * 0.3, 1, true);
        positions.push({ x, y });
      }

      trailFrameSkip = (trailFrameSkip + 1) % 2;
      if (trailFrameSkip === 0) updateTrails(positions);
      drawTrails();

      if (ENABLE_MICRO) {
        for (let j = 0; j < 2; j++) {
          const orbAngle = angle * 1.2 + j * 3;
          const px = cx + Math.cos(orbAngle) * (radius + 15);
          const py = cy + Math.sin(orbAngle) * (radius + 15);
          drawLambda(px, py, 6, orbAngle, 0.15);
        }
      }
    }

    // GATHERING — новая фаза: три лямбды съезжаются в центр
    else if (state === 'gathering') {
      const p = Math.min(1, stateTimer / 1.0); // 0 → 1
      // Начинаем с радиуса, который был в конце orbiting (примерно 6-8), и уменьшаем до 0
      const startRadius = 8;
      const currentRadius = startRadius * (1 - p);
      // Угол продолжаем вращать, но медленнее
      const angle = t * 8; // небольшое вращение

      for (let i = 0; i < 3; i++) {
        const a = angle + (i * Math.PI * 2 / 3);
        // Позиция: от периферии к центру
        const x = cx + Math.cos(a) * currentRadius;
        const y = cy + Math.sin(a) * currentRadius;
        // Размер немного уменьшается к центру
        const size = baseSize * 0.7 * (0.9 + 0.1 * Math.sin(t*10 + i));
        // Прозрачность почти полная, но можно сделать пульсирующей
        const alpha = 0.9 + 0.1 * Math.sin(t*15 + i);
        drawLambda(x, y, size, a * 0.3, alpha, 1, true);
      }

      // Добавим свечение, усиливающееся к концу сбора
      if (ENABLE_GLOW) {
        ctx.save();
        ctx.globalCompositeOperation = 'lighter';
        const glow = 0.2 + p * 0.3;
        const radius = baseSize * 4;
        const grad = ctx.createRadialGradient(cx, cy, 0, cx, cy, radius);
        grad.addColorStop(0, `rgba(255,255,255,${glow})`);
        grad.addColorStop(0.5, accentColor(glow * 0.5));
        grad.addColorStop(1, 'rgba(0,0,0,0)');
        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.arc(cx, cy, radius, 0, 2 * Math.PI);
        ctx.fill();
        ctx.restore();
      }
    }

    // MERGING (немного модифицирован: теперь они стартуют почти из центра)
    else if (state === 'merging') {
      const p = 1 - Math.min(1, stateTimer / 1.6); // 1 → 0
      // В начале merging (p=1) они все в центре, поэтому offset мал
      const offset = p * 10; // небольшой разброс для эффекта "выдоха"
      const wobble = Math.sin(stateTimer * 40) * 2 * (1 - p);

      for (let i = 0; i < 3; i++) {
        const a = (i * Math.PI * 2 / 3) + Math.sin(t * 5) * 0.1;
        const ox = Math.cos(a) * offset;
        const oy = Math.sin(a) * offset;
        const x = cx + ox + (i === 1 ? wobble : 0);
        const y = cy + oy + (i === 2 ? wobble : 0);
        drawLambda(x, y, baseSize * 0.8, a * 0.3, 1, 1, true);
      }

      if (p < 0.4) {
        ctx.beginPath();
        ctx.strokeStyle = accentColor(0.4);
        ctx.lineWidth = 2;
        ctx.arc(cx, cy, baseSize * (1.3 + (1 - p) * 2), 0, 2 * Math.PI);
        ctx.stroke();
      }
    }

    // ОБЩЕЕ СВЕЧЕНИЕ (простое)
    if (ENABLE_GLOW && (flashAlpha > 0.02 || state === 'merging' || state === 'gathering')) {
      ctx.save();
      ctx.globalCompositeOperation = 'lighter';
      let glow = flashAlpha;
      if (state === 'merging') glow = Math.max(glow, 0.2);
      if (state === 'gathering') glow = Math.max(glow, 0.15);
      if (glow > 0.01) {
        const radius = baseSize * GLOW_RADIUS_FACTOR;
        const grad = ctx.createRadialGradient(cx, cy, 0, cx, cy, radius);
        grad.addColorStop(0, `rgba(255,255,255,${glow})`);
        grad.addColorStop(0.3, accentColor(glow * 0.5));
        grad.addColorStop(0.8, 'rgba(0,0,0,0)');
        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.arc(cx, cy, radius, 0, 2 * Math.PI);
        ctx.fill();
      }
      ctx.restore();
    }

    // КОЛЬЦА
    flashRings.forEach(r => {
      ctx.beginPath();
      ctx.strokeStyle = accentColor(r.life * 0.5);
      ctx.lineWidth = 1.5 * r.life;
      ctx.arc(cx, cy, r.radius, 0, 2 * Math.PI);
      ctx.stroke();
    });

    drawMicroLambdas();

    requestAnimationFrame(draw);
  }

  draw();
})();