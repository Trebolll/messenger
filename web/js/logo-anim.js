// ─── logo-anim.js — анимация логотипа (виртуальные частицы λ) ──────────────
(function () {
  const canvas = document.getElementById('logo-canvas');
  if (!canvas) return;

  const ctx = canvas.getContext('2d');
  let w, h, cx, cy, dpr;
  let t = 0;
  let state = 'single'; // 'single', 'splitting', 'orbiting', 'merging'
  let stateTimer = 0;
  let flashAlpha = 0;
  let flashParticles = [];
  let flashRings = [];

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
    const a = isNaN(alpha) ? 0 : alpha;
    const computed = getComputedStyle(document.body).getPropertyValue('--accent-color').trim() || '#3b82f6';
    if (computed.startsWith('#')) {
      const r = parseInt(computed.slice(1, 3), 16);
      const g = parseInt(computed.slice(3, 5), 16);
      const b = parseInt(computed.slice(5, 7), 16);
      return `rgba(${r},${g},${b},${a})`;
    }
    return computed.replace('rgb', 'rgba').replace(')', `,${a})`);
  }

  function drawLambda(x, y, size, angle, alpha, scaleX = 1) {
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
    ctx.lineJoin = 'round'; // Добавим мягкости углам

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

  function update() {
    t += 0.016;
    stateTimer += 0.016;

    if (state === 'single') {
      if (stateTimer > 8 + Math.random() * 5) {
        state = 'splitting';
        stateTimer = 0;
      }
    } else if (state === 'splitting') {
      if (stateTimer > 1) {
        state = 'orbiting';
        stateTimer = 0;
      }
    } else if (state === 'orbiting') {
      if (stateTimer > 4) { // Увеличили время до 4с для разгона
        state = 'merging';
        stateTimer = 0;
      }
    } else if (state === 'merging') {
      if (stateTimer > 1.2) {
        state = 'single';
        stateTimer = 0;
        flashAlpha = 1.5; 
        
        // Генерируем ударные кольца
        flashRings.push({ radius: 10, speed: 2, life: 1.0 });
        flashRings.push({ radius: 5, speed: 4, life: 0.7 });
      }
    }

    if (flashAlpha > 0) flashAlpha -= 0.035; // Чуть медленнее затухание

    // Обновляем частицы вспышки
    flashParticles = flashParticles.filter(p => p.life > 0);
    flashParticles.forEach(p => {
      p.x += p.vx;
      p.y += p.vy;
      p.life -= p.decay;
    });

    // Обновляем кольца
    flashRings = flashRings.filter(r => r.life > 0);
    flashRings.forEach(r => {
      r.radius += r.speed;
      r.life -= 0.02;
    });
  }

  function draw() {
    if (!cx || !cy) return requestAnimationFrame(draw);
    ctx.clearRect(0, 0, w, h);
    update();

    const baseSize = 34;

    if (state === 'single') {
      const appearance = Math.min(1, stateTimer / 1.5); 
      const scale = 0.8 + 0.2 * appearance;
      
      ctx.save();
      ctx.translate(cx, cy);
      ctx.scale(scale, scale);
      drawLambda(0, 0, baseSize, Math.sin(t * 0.5) * 0.1, appearance);
      ctx.restore();
    } 
    else if (state === 'splitting') {
      const p = stateTimer; // 0 to 1
      const offset = p * 15;
      drawLambda(cx - offset, cy, baseSize, -p * 0.2, 1 - p * 0.2);
      drawLambda(cx + offset, cy, baseSize, p * 0.2, 1 - p * 0.2);
      if (p > 0.5) {
        const p2 = (p - 0.5) * 2;
        drawLambda(cx, cy - offset * p2, baseSize * 0.7, 0, (1 - p) * 0.8);
      }
    } 
    else if (state === 'orbiting') {
      const progress = stateTimer / 4; 
      const accel = Math.pow(progress, 2.8); 
      const speed = 5 + accel * 45; 
      const angle = stateTimer * speed;
      const radius = 14 * (1 - progress * 0.7); 
      
      // 3 объекта в танце
      for (let i = 0; i < 3; i++) {
        const a = angle + (i * Math.PI * 2 / 3);
        const x = cx + Math.cos(a) * radius;
        const y = cy + Math.sin(a) * radius;
        const shake = progress > 0.8 ? (Math.random() - 0.5) * progress * 5 : 0;
        drawLambda(x + shake, y + shake, baseSize * 0.7, a * 0.5, 0.7 + progress * 0.3);
      }
    } else if (state === 'merging') {
      const p = 1 - stateTimer / 1.2; 
      const offset = p * 15;
      const wobble = Math.sin(stateTimer * 40) * 2 * (1 - p); 
      
      for (let i = 0; i < 3; i++) {
        const a = (i * Math.PI * 2 / 3);
        const ox = Math.cos(a) * offset;
        const oy = Math.sin(a) * offset;
        drawLambda(cx + ox + (i===0?wobble:0), cy + oy, baseSize * 0.8, 0, 0.8 + (1 - p) * 0.2);
      }
      
      // Дополнительные энергетические линии при слиянии
      if (p < 0.3) {
        ctx.beginPath();
        ctx.strokeStyle = accentColor(1 - p / 0.3);
        ctx.lineWidth = 0.5;
        ctx.arc(cx, cy, baseSize * (1 - p / 0.3), 0, Math.PI * 2);
        ctx.stroke();
      }
    }

    // Постоянное блестящее мерцание (subtle sparkle)
    const sparkle = 0.5 + 0.5 * Math.sin(t * 10);
    const alphaSparkle = 0.4 + 0.6 * Math.random() * (state === 'merging' ? 1.5 : 1);
    
    // Блеск/Вспышка (ультра-плавный многослойный градиент без границ)
    if (flashAlpha > 0 || state === 'merging' || (state === 'orbiting' && stateTimer > 1)) {
      ctx.save();
      ctx.globalCompositeOperation = 'lighter';
      
      let glowStr = flashAlpha;
      if (state === 'merging') glowStr = Math.max(glowStr, 0.15);
      if (state === 'orbiting') {
        const progress = stateTimer / 4;
        const accel = Math.pow(progress, 3); // Резкий рост свечения к концу разгона
        glowStr = accel * 0.45; // Свечение нарастает от 0 до 0.45
      }

      if (glowStr > 0.01) {
        const radius = baseSize * 5.5; // Оптимальный радиус для 400px холста
        const grad = ctx.createRadialGradient(cx, cy, 0, cx, cy, radius);
        
        grad.addColorStop(0, `rgba(255,255,255,${glowStr * 1.0})`);
        grad.addColorStop(0.1, accentColor(glowStr * 0.8));
        grad.addColorStop(0.2, accentColor(glowStr * 0.4));
        grad.addColorStop(0.5, accentColor(glowStr * 0.1));
        grad.addColorStop(0.8, accentColor(glowStr * 0.02));
        grad.addColorStop(1, 'rgba(0,0,0,0)');
        
        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.arc(cx, cy, radius, 0, Math.PI * 2);
        ctx.fill();
      }
      ctx.restore();
    }

    // Отрисовка ударных колец
    flashRings.forEach(r => {
      ctx.beginPath();
      ctx.strokeStyle = accentColor(r.life * 0.5);
      ctx.lineWidth = 2;
      ctx.arc(cx, cy, r.radius, 0, Math.PI * 2);
      ctx.stroke();
    });

    requestAnimationFrame(draw);
  }

  draw();
})();
