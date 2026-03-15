// Контактная информация для сотрудничества
const CONTACT_INFO = {
  email:    'trebolll@yandex.ru',
  telegram: 'https://t.me/treboll777',
};

function openContactModal() {
  const overlay = document.getElementById('contact-modal-overlay');
  const box     = document.getElementById('contact-modal-box');
  overlay.style.display = 'flex';
  requestAnimationFrame(() => requestAnimationFrame(() => {
    box.style.transform = 'translateY(0)';
    box.style.opacity   = '1';
  }));
}

function closeContactModal() {
  const box = document.getElementById('contact-modal-box');
  box.style.transform = 'translateY(20px)';
  box.style.opacity   = '0';
  setTimeout(() => {
    document.getElementById('contact-modal-overlay').style.display = 'none';
  }, 300);
}

document.addEventListener('keydown', e => { if (e.key === 'Escape') closeContactModal(); });

document.addEventListener('DOMContentLoaded', () => {
  const emailEl = document.getElementById('contact-email');
  const tgEl    = document.getElementById('contact-telegram');
  if (emailEl) { emailEl.href = 'mailto:' + CONTACT_INFO.email; emailEl.textContent = CONTACT_INFO.email; }
  if (tgEl)    { tgEl.href = CONTACT_INFO.telegram; tgEl.textContent = CONTACT_INFO.telegram.replace('https://', ''); }
});