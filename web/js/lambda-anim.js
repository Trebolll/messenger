(function () {
  // ── Глобальный стейт цвета (общий для всех инстансов) ────────────────
  let _accentR = 59, _accentG = 130, _accentB = 246;
  let _bgColor  = '#ffffff';
  let _isDark   = false;

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
  refreshColorCache();
  new MutationObserver(refreshColorCache).observe(document.body, { attributes: true, attributeFilter: ['class'] });

  function accentColor(alpha) {
    return `rgba(${_accentR},${_accentG},${_accentB},${alpha})`;
  }

  const DISK_TILT = 0.33; 
  const MOUSE_DIST_SQ = 100 * 100;

  // ── Класс анимации ───────────────────────────────────────────────────
  class LambdaAnimation {
    constructor(canvasId) {
      this.canvas = document.getElementById(canvasId);
      if (!this.canvas) return;
      this.ctx = this.canvas.getContext('2d', { alpha: false });
      this.W = 0; this.H = 0; this.cx = 0; this.cy = 0; this.dpr = 1;
      this.t = 0; this.lastTime = 0;
      this.mx = -1000; this.my = -1000;
      this.raf = null;

      const PARTICLE_COUNT = 1500;
      this.PARTICLE_COUNT = PARTICLE_COUNT;
      this.pR          = new Float32Array(PARTICLE_COUNT);
      this.pAngle      = new Float32Array(PARTICLE_COUNT);
      this.pSize       = new Float32Array(PARTICLE_COUNT);
      this.pPhase      = new Float32Array(PARTICLE_COUNT);
      this.pShrink     = new Float32Array(PARTICLE_COUNT);
      this.pTilt       = new Float32Array(PARTICLE_COUNT);
      this.pPitch      = new Float32Array(PARTICLE_COUNT);
      this.pCosT       = new Float32Array(PARTICLE_COUNT);
      this.pSinT       = new Float32Array(PARTICLE_COUNT);
      this.pIsLambda   = new Uint8Array(PARTICLE_COUNT);

      this.initEvents();
      this.resize();
      for (let i = 0; i < PARTICLE_COUNT; i++) this.initParticle(i, false);
      
      this.animate = this.animate.bind(this);
      this.raf = requestAnimationFrame(this.animate);

      document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
          if (this.raf) cancelAnimationFrame(this.raf);
        } else {
          this.lastTime = 0;
          this.raf = requestAnimationFrame(this.animate);
        }
      });
    }

    initEvents() {
      window.addEventListener('mousemove', (e) => {
        if (!this.canvas) return;
        const rect = this.canvas.getBoundingClientRect();
        this.mx = e.clientX - rect.left;
        this.my = e.clientY - rect.top;
      });
      window.addEventListener('resize', () => this.resize());
    }

    resize() {
      if (!this.canvas) return;
      this.dpr = window.devicePixelRatio || 1;
      this.W = this.canvas.offsetWidth;
      this.H = this.canvas.offsetHeight;
      this.canvas.width  = this.W * this.dpr;
      this.canvas.height = this.H * this.dpr;
      this.ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
      this.cx = this.W / 2;
      this.cy = this.H / 2;
    }

    initParticle(i, isNew) {
      const screenMax = Math.max(this.W, this.H) || 1000;
      this.pR[i]        = isNew
          ? (screenMax * 0.4 + Math.random() * screenMax * 0.5)
          : (10 + Math.random() * screenMax * 1.5);
      this.pAngle[i]    = Math.random() * Math.PI * 2;
      this.pSize[i]     = 0.27 + Math.random() * 0.86;
      this.pPhase[i]    = Math.random() * Math.PI * 2;
      this.pShrink[i]   = (0.12 + Math.random() * 0.2) * 0.25;
      this.pTilt[i]     = (Math.random() - 0.5) * 0.22;
      this.pPitch[i]    = 0.15 + Math.random() * 0.15;
      this.pCosT[i]     = Math.cos(this.pTilt[i]);
      this.pSinT[i]     = Math.sin(this.pTilt[i]);
      this.pIsLambda[i] = (i % 8 === 0) ? 1 : 0;
    }

    drawLambda(x, y, size, angle, alpha) {
      const {ctx} = this;
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

    animate(now) {
      if (!this.canvas) return;
      
      const delta = this.lastTime ? Math.min((now - this.lastTime) / 1000, 0.05) : 0.001;
      const dtFactor = delta / 0.016;
      this.lastTime = now;
      const {ctx, W, H, cx, cy} = this;

      ctx.globalCompositeOperation = 'source-over';
      ctx.fillStyle   = _bgColor;
      ctx.globalAlpha = _isDark ? 0.3 : 0.5;
      ctx.fillRect(0, 0, W, H);

      ctx.globalCompositeOperation = _isDark ? 'lighter' : 'source-over';
      ctx.globalAlpha = 1.0;
      this.t += 0.0006 * dtFactor;

      for (let i = 0; i < this.PARTICLE_COUNT; i++) {
        const gravityFactor = 1.6 / Math.pow(Math.max(20, this.pR[i]), 0.9);
        const swirlBoost = this.pR[i] < 70 ? 1 + Math.pow((70 - this.pR[i]) / 70, 2) * 10.5 : 1.0;
        this.pAngle[i] += gravityFactor * swirlBoost * dtFactor;
        const gravPull = 1.0 + (500 / (this.pR[i] + 10)); 
        const fallSpeed = (this.pShrink[i] + 0.05) * dtFactor * gravPull;
        this.pR[i] -= fallSpeed;

        if (this.pR[i] < 12) { this.initParticle(i, true); continue; }

        const cosA = Math.cos(this.pAngle[i]);
        const sinA = Math.sin(this.pAngle[i]);
        const rx   = this.pR[i] * (1 + 0.01 * Math.sin(this.t * 1.5 + this.pPhase[i]));
        const ry   = rx * this.pPitch[i];
        const lx   = cosA * rx;
        const ly   = sinA * ry;
        const px = cx + (lx * this.pCosT[i] - ly * this.pSinT[i]);
        const py = cy + (lx * this.pSinT[i] + ly * this.pCosT[i]) - (px - cx) * DISK_TILT;
        const r = this.pR[i];
        const doppler   = 0.7 + 0.3 * (1.0 - sinA * 0.4);
        const viewAlpha = r > 1800 ? 0 : Math.min(1, (1800 - r) / 600);
        const distRatio = 1 - (Math.min(r, 1200) / 1200);
        let alpha = (0.12 + 0.75 * distRatio) * doppler * viewAlpha;
        let size  = this.pSize[i] * (1.0 + distRatio * 0.5);
        if (r < 20) { const factor = Math.max(0, (r - 12) / 8); size *= factor; alpha *= factor; }
        if (sinA < 0) alpha *= 0.8;
        if (alpha < 0.002) continue;

        // Mouse glow
        const ddx = px - this.mx; const ddy = py - this.my;
        const distSq = ddx * ddx + ddy * ddy;
        if (distSq < MOUSE_DIST_SQ && alpha > 0.01) {
          const proximity = 1 - Math.sqrt(distSq) / 100;
          const glow = Math.pow(Math.max(0, proximity), 1.6);
          if (glow > 0.03) {
            ctx.save();
            ctx.globalCompositeOperation = _isDark ? 'lighter' : 'multiply';
            const coreR = size * 3.5;
            const coreGrad = ctx.createRadialGradient(px, py, 0, px, py, coreR);
            const coreAlpha = _isDark ? glow * 0.95 : glow * 0.4;
            const whiteAlpha = _isDark ? glow * 0.6 : glow * 0.1;
            
            coreGrad.addColorStop(0, _isDark ? `rgba(255,255,255,${coreAlpha})` : accentColor(coreAlpha));
            coreGrad.addColorStop(0.2, _isDark ? `rgba(255,255,255,${whiteAlpha})` : accentColor(whiteAlpha));
            coreGrad.addColorStop(0.5, accentColor(glow * 0.8));
            coreGrad.addColorStop(1, accentColor(0));
            ctx.beginPath(); ctx.arc(px, py, coreR, 0, Math.PI * 2);
            ctx.fillStyle = coreGrad; ctx.fill();

            const auraR = size * 28 * glow + 18;
            const auraGrad = ctx.createRadialGradient(px, py, coreR, px, py, auraR);
            auraGrad.addColorStop(0, accentColor(glow * 0.35));
            auraGrad.addColorStop(0.4, accentColor(glow * 0.12));
            auraGrad.addColorStop(1, accentColor(0));
            ctx.beginPath(); ctx.arc(px, py, auraR, 0, Math.PI * 2);
            ctx.fillStyle = auraGrad; ctx.fill();

            if (glow > 0.45) {
              const arm = (size * 18 + 10) * glow;
              ctx.lineWidth = size * 1.2; ctx.lineCap = 'round';
              const lineAlpha = _isDark ? glow * 0.8 : glow * 0.3;
              const hGrad = ctx.createLinearGradient(px - arm, py, px + arm, py);
              hGrad.addColorStop(0, accentColor(0));
              hGrad.addColorStop(0.5, _isDark ? `rgba(255,255,255,${lineAlpha})` : accentColor(lineAlpha));
              hGrad.addColorStop(1, accentColor(0));
              ctx.beginPath(); ctx.moveTo(px - arm, py); ctx.lineTo(px + arm, py);
              ctx.strokeStyle = hGrad; ctx.stroke();

              const vGrad = ctx.createLinearGradient(px, py - arm, px, py + arm);
              vGrad.addColorStop(0, accentColor(0));
              vGrad.addColorStop(0.5, _isDark ? `rgba(255,255,255,${lineAlpha})` : accentColor(lineAlpha));
              vGrad.addColorStop(1, accentColor(0));
              ctx.beginPath(); ctx.moveTo(px, py - arm); ctx.lineTo(px, py + arm);
              ctx.strokeStyle = vGrad; ctx.stroke();
            }
            ctx.restore();
          }
        }

        if (this.pIsLambda[i]) {
          this.drawLambda(px, py, size * 6, this.pAngle[i], alpha * (0.7 + 0.3 * Math.sin(this.t * 2 + this.pPhase[i])));
        } else {
          ctx.beginPath();
          ctx.arc(px, py, size, 0, Math.PI * 2);
          ctx.fillStyle = accentColor(alpha);
          ctx.fill();
        }
      }
      this.raf = requestAnimationFrame(this.animate);
    }

    spawnAt(pageX, pageY, isLambda) {
      const rect = this.canvas.getBoundingClientRect();
      const x = (pageX - rect.left) * this.dpr;
      const y = (pageY - rect.top) * this.dpr;
      const i = Math.floor(Math.random() * this.PARTICLE_COUNT);
      const tx = (x - (this.W * this.dpr / 2)) / this.dpr;
      const ty = (y - (this.H * this.dpr / 2)) / this.dpr;
      const dy_corr = ty + tx * DISK_TILT;
      const pitch = 0.15 + Math.random() * 0.15;
      this.pR[i] = Math.sqrt(tx * tx + Math.pow(dy_corr / pitch, 2));
      this.pAngle[i] = Math.atan2(dy_corr / pitch, tx);
      this.pSize[i] = 0.7 + Math.random() * 0.8;
      this.pTilt[i] = 0; this.pCosT[i] = 1; this.pSinT[i] = 0;
      this.pPitch[i] = pitch; this.pIsLambda[i] = isLambda ? 1 : 0;
      this.pShrink[i] = 0.08 + Math.random() * 0.12;
    }
  }

  // ── Инициализация ────────────────────────────────────────────────────
  const landingAnim = new LambdaAnimation('lambda-canvas');

  window.spawnEscapedParticle = function(pageX, pageY, isLambda) {
    if (landingAnim && landingAnim.canvas) {
      landingAnim.spawnAt(pageX, pageY, isLambda);
    }
  };

})();
