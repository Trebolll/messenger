// ─── News Feed — Wikipedia + Hacker News APIs ──────────────────────────────
// Все источники полностью открытые, без ключей, отдают полный контент статей

let _currentFeed = 'politics';

// ── Утилиты ────────────────────────────────────────────────────────────────

// Fisher-Yates shuffle
function shuffle(arr) {
  const a = [...arr];
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

// Случайное целое от min до max включительно
function randInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

// ── Wikipedia REST API helpers ─────────────────────────────────────────────

async function fetchWikiSummary(title) {
  const url = `https://en.wikipedia.org/api/rest_v1/page/summary/${encodeURIComponent(title)}`;
  const res = await fetch(url, { headers: { 'Accept': 'application/json' } });
  if (!res.ok) throw new Error('wiki summary error');
  return res.json();
}

async function fetchWikiMobileHtml(title) {
  const url = `https://en.wikipedia.org/api/rest_v1/page/mobile-html/${encodeURIComponent(title)}`;
  const res = await fetch(url, { headers: { 'Accept': 'text/html' } });
  if (!res.ok) throw new Error('wiki html error');
  return res.text();
}

async function fetchWikiFeatured() {
  const d     = new Date();
  const year  = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day   = String(d.getDate()).padStart(2, '0');
  const url   = `https://api.wikimedia.org/feed/v1/wikipedia/en/featured/${year}/${month}/${day}`;
  const res   = await fetch(url, { headers: { 'Accept': 'application/json' } });
  if (!res.ok) throw new Error('wiki featured error');
  return res.json();
}

// Случайные Wikipedia статьи через Special:Random API
async function fetchWikiRandom(count = 8) {
  const url = `https://en.wikipedia.org/api/rest_v1/page/random/summary`;
  const results = await Promise.allSettled(
      Array.from({ length: count }, () =>
          fetch(url, { headers: { 'Accept': 'application/json' } }).then(r => r.json())
      )
  );
  return results
      .filter(r => r.status === 'fulfilled' && r.value?.title)
      .map(r => r.value);
}

// Исторические события за случайный день этого месяца
async function fetchWikiOnThisDay() {
  const d     = new Date();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  // Берём случайный день в диапазоне 1–28 чтобы точно существовал
  const day   = String(randInt(1, 28)).padStart(2, '0');
  const url   = `https://api.wikimedia.org/feed/v1/wikipedia/en/onthisday/all/${month}/${day}`;
  try {
    const res  = await fetch(url, { headers: { 'Accept': 'application/json' } });
    if (!res.ok) return [];
    const data = await res.json();
    return shuffle([
      ...(data.events   || []),
      ...(data.births   || []),
      ...(data.deaths   || []),
    ]).slice(0, 10);
  } catch { return []; }
}

// ── Hacker News API ────────────────────────────────────────────────────────

// Берём 500 id и каждый раз выбираем случайные 20
async function fetchHNStories(type = 'topstories') {
  const ids     = await fetch(`https://hacker-news.firebaseio.com/v0/${type}.json`).then(r => r.json());
  const poolSize = Math.min(ids.length, type === 'newstories' ? 200 : 500);
  const picked  = shuffle(ids.slice(0, poolSize)).slice(0, 20);
  const stories = await Promise.all(
      picked.map(id =>
          fetch(`https://hacker-news.firebaseio.com/v0/item/${id}.json`).then(r => r.json()).catch(() => null)
      )
  );
  return stories.filter(s => s && s.url && s.title);
}

// ── Загрузка по категории ──────────────────────────────────────────────────

async function loadNews(feedKey) {
  const skeleton   = document.getElementById('news-skeleton');
  const cards      = document.getElementById('news-cards');
  const errorEl    = document.getElementById('news-error');
  const refreshBtn = document.getElementById('news-refresh-btn');
  const overlay    = document.getElementById('news-loading-overlay');

  // Показываем оверлей со спиннером
  overlay && overlay.classList.remove('hidden');
  skeleton && skeleton.classList.add('hidden');
  cards.classList.add('hidden');
  errorEl.classList.add('hidden');
  refreshBtn && refreshBtn.classList.add('spinning');

  try {
    let items = [];

    if (feedKey === 'politics') {
      const politicsPages = shuffle([
        'United_Nations','European_Union','NATO','United_States_Congress',
        'Democracy','Geopolitics','Diplomacy','International_relations',
        'United_Nations_Security_Council','Foreign_policy','Political_party',
        'Election','Human_rights','Sovereignty','War','Treaty',
      ]).slice(0, 9);
      const [searches, randoms] = await Promise.all([
        Promise.all(politicsPages.map(p =>
            fetch(`https://en.wikipedia.org/api/rest_v1/page/summary/${encodeURIComponent(p)}`).then(r => r.json()).catch(() => null)
        )),
        fetchWikiRandom(4),
      ]);
      shuffle([...searches.filter(Boolean), ...randoms]).forEach(art => {
        if (!art?.title || !art.extract) return;
        items.push({
          _type: 'wiki', _title_key: art.title,
          title: art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: art.extract || '', thumbnail: art.thumbnail?.source || '',
          _source: 'Wikipedia · Политика', pubDate: new Date().toISOString(),
        });
      });
      items = shuffle(items);

    } else if (feedKey === 'science') {
      // Параллельно: featured + случайные статьи + ask HN
      const [featured, randoms, ask] = await Promise.all([
        fetchWikiFeatured(),
        fetchWikiRandom(5),
        fetchHNStories('askstories'),
      ]);

      const image     = featured.image;
      const onthisday = shuffle(featured.onthisday || []).slice(0, 5);

      if (image) {
        items.push({
          _type:      'wiki',
          _title_key: image.title,
          title:      image.description?.text || image.title,
          description: image.description?.text || '',
          thumbnail:  image.thumbnail?.source || image.image?.source || '',
          _source:    'Wikipedia · Изображение дня',
          pubDate:    new Date().toISOString(),
        });
      }

      randoms.forEach(art => {
        if (!art.extract) return;
        items.push({
          _type:      'wiki',
          _title_key: art.title,
          title:      art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: art.extract || '',
          thumbnail:  art.thumbnail?.source || '',
          _source:    'Wikipedia · Случайная статья',
          pubDate:    new Date().toISOString(),
        });
      });

      onthisday.forEach(e => {
        const art = e.pages?.[0];
        if (!art) return;
        items.push({
          _type:      'wiki',
          _title_key: art.title,
          title:      art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: e.text || '',
          thumbnail:  art.thumbnail?.source || '',
          _source:    `Wikipedia · ${e.year} год`,
          pubDate:    new Date().toISOString(),
        });
      });

      shuffle(ask).slice(0, 6).forEach(s => {
        items.push({
          _type:      'hn',
          _hn_id:     s.id,
          title:      s.title,
          description: s.text?.replace(/<[^>]+>/g, '').slice(0, 160) || `${s.score} очков`,
          thumbnail:  '',
          link:       `https://news.ycombinator.com/item?id=${s.id}`,
          _source:    'Hacker News · Ask HN',
          pubDate:    new Date(s.time * 1000).toISOString(),
        });
      });

      items = shuffle(items);

    } else if (feedKey === 'history') {
      const days = await Promise.all([fetchWikiOnThisDay(), fetchWikiOnThisDay()]);
      const randoms = await fetchWikiRandom(5);
      shuffle([...days[0], ...days[1]]).forEach(e => {
        const art = e.pages?.[0];
        if (!art) return;
        items.push({
          _type:      'wiki',
          _title_key: art.title,
          title:      art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: e.text || '',
          thumbnail:  art.thumbnail?.source || '',
          _source:    `История · ${e.year} год`,
          pubDate:    new Date().toISOString(),
        });
      });
      randoms.forEach(art => {
        if (!art.extract) return;
        items.push({
          _type: 'wiki', _title_key: art.title,
          title: art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: art.extract || '', thumbnail: art.thumbnail?.source || '',
          _source: 'Wikipedia · История', pubDate: new Date().toISOString(),
        });
      });
      items = shuffle(items);

    } else if (feedKey === 'space') {
      const spacePages = shuffle([
        'Space_exploration','NASA','Mars','Black_hole','James_Webb_Space_Telescope',
        'SpaceX','International_Space_Station','Moon','Milky_Way','Neutron_star',
        'Exoplanet','Hubble_Space_Telescope','Big_Bang','Dark_matter','Rocket',
      ]).slice(0, 9);
      const [searches, randoms] = await Promise.all([
        Promise.all(spacePages.map(p =>
            fetch(`https://en.wikipedia.org/api/rest_v1/page/summary/${encodeURIComponent(p)}`).then(r => r.json()).catch(() => null)
        )),
        fetchWikiRandom(4),
      ]);
      shuffle([...searches.filter(Boolean), ...randoms]).forEach(art => {
        if (!art?.title) return;
        items.push({
          _type: 'wiki', _title_key: art.title,
          title: art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: art.extract || '', thumbnail: art.thumbnail?.source || '',
          _source: 'Wikipedia · Космос', pubDate: new Date().toISOString(),
        });
      });
      items = shuffle(items);

    } else if (feedKey === 'cars') {
      const carPages = shuffle([
        'Automobile','Formula_One','Ferrari','Porsche','Tesla,_Inc.',
        'Bugatti','Lamborghini','Electric_vehicle','Internal_combustion_engine',
        'Mercedes-Benz','BMW','Audi','Le_Mans_24_Hours','Rallying',
        'McLaren','Volkswagen','Toyota','Hyundai','Chevrolet',
      ]).slice(0, 10);
      const [searches, randoms] = await Promise.all([
        Promise.all(carPages.map(p =>
            fetch(`https://en.wikipedia.org/api/rest_v1/page/summary/${encodeURIComponent(p)}`).then(r => r.json()).catch(() => null)
        )),
        fetchWikiRandom(4),
      ]);
      shuffle([...searches.filter(Boolean), ...randoms]).forEach(art => {
        if (!art?.title) return;
        items.push({
          _type: 'wiki', _title_key: art.title,
          title: art.displaytitle?.replace(/<[^>]+>/g, '') || art.title,
          description: art.extract || '', thumbnail: art.thumbnail?.source || '',
          _source: 'Wikipedia · Автомобили', pubDate: new Date().toISOString(),
        });
      });
      items = shuffle(items);
    }

    if (!items.length) throw new Error('no items');

    overlay && overlay.classList.add('hidden');
    skeleton && skeleton.classList.add('hidden');
    cards.classList.remove('hidden');
    renderNewsCards(items);

  } catch (e) {
    console.error('News error:', e);
    overlay && overlay.classList.add('hidden');
    skeleton && skeleton.classList.add('hidden');
    errorEl.classList.remove('hidden');
  } finally {
    refreshBtn && refreshBtn.classList.remove('spinning');
  }
}

function extractDomain(url) {
  try { return new URL(url).hostname.replace('www.', ''); } catch { return ''; }
}

function timeAgo(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  if (isNaN(d)) return '';
  const diff = (Date.now() - d) / 1000;
  if (diff < 60)    return 'только что';
  if (diff < 3600)  return `${Math.floor(diff / 60)} мин назад`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} ч назад`;
  return `${Math.floor(diff / 86400)} д назад`;
}

// Глобальное хранилище карточек — без JSON в onclick
let _newsItems = [];

function renderNewsCards(items) {
  _newsItems = items;
  const grid = document.getElementById('news-cards');
  grid.innerHTML = items.map((item, i) => {
    const title = item.title || '';
    const desc  = (item.description || '').slice(0, 160);
    const date  = timeAgo(item.pubDate);
    const img   = item.thumbnail || '';
    const src   = (item._source || '').replace(/</g, '&lt;');

    const imgHtml = img
        ? `<img class="news-card-img" src="${img}" alt="" loading="lazy" onerror="this.style.display='none';this.nextElementSibling.style.display='flex'">`
        : '';
    const placeholder = `<div class="news-card-img-placeholder" style="${img ? 'display:none' : ''}">${item._type === 'hn' ? '🔸' : '📖'}</div>`;

    return `
        <div class="news-card" onclick="openArticle(${i})" style="animation-delay:${Math.min(i * 35, 500)}ms">
            ${imgHtml}${placeholder}
            <div class="news-card-body">
                <div class="news-card-source">${src}</div>
                <div class="news-card-title">${title}</div>
                ${desc ? `<div class="news-card-desc">${desc}</div>` : ''}
                <div class="news-card-footer">
                    <span class="news-card-date">${date}</span>
                    <span class="news-card-arrow">
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M7 17L17 7M7 7h10v10"/>
                        </svg>
                    </span>
                </div>
            </div>
        </div>`;
  }).join('');
}

function setNewsFeed(feedKey, btn) {
  document.querySelectorAll('.news-tab').forEach(t => t.classList.remove('active'));
  btn.classList.add('active');
  _currentFeed = feedKey;
  loadNews(feedKey); // всегда перезагружаем — каждый клик = новые карточки
}

function refreshNews() { loadNews(_currentFeed); }

document.addEventListener('DOMContentLoaded', () => { /* wiki панель закрыта по умолчанию */ });

// ─── Article Modal ──────────────────────────────────────────────────────────

function openArticle(idx) {
  const item = _newsItems[idx];
  if (!item) return;

  const overlay  = document.getElementById('article-overlay');
  const loading  = document.getElementById('article-loading');
  const error    = document.getElementById('article-error');
  const content  = document.getElementById('article-content');
  const srcBadge = document.getElementById('article-source');
  const dateEl   = document.getElementById('article-date');
  const extLink  = document.getElementById('article-open-link');
  const errLink  = document.getElementById('article-error-link');

  loading.classList.remove('hidden');
  error.classList.add('hidden');
  content.classList.add('hidden');
  overlay.classList.remove('hidden');
  document.body.style.overflow = 'hidden';

  srcBadge.textContent = item._source || '';
  dateEl.textContent   = timeAgo(item.pubDate);

  const extUrl = item._type === 'wiki'
      ? `https://en.wikipedia.org/wiki/${encodeURIComponent(item._title_key || item.title)}`
      : (item.link || item._hn_url || '#');

  extLink.href = extUrl;
  errLink.href = extUrl;

  if (item._type === 'wiki') {
    fetchWikiArticle(item);
  } else {
    fetchHNArticle(item);
  }
}

async function fetchWikiArticle(item) {
  const loading = document.getElementById('article-loading');
  const error   = document.getElementById('article-error');
  const content = document.getElementById('article-content');
  const titleEl = document.getElementById('article-title');
  const textEl  = document.getElementById('article-text');
  const heroImg = document.getElementById('article-hero-img');

  try {
    // Сначала получаем summary (быстро)
    const summary = await fetchWikiSummary(item._title_key || item.title);

    titleEl.textContent = summary.displaytitle?.replace(/<[^>]+>/g, '') || summary.title;

    const img = summary.thumbnail?.source || item.thumbnail || '';
    if (img) {
      heroImg.src = img;
      heroImg.classList.remove('hidden');
      heroImg.onerror = () => heroImg.classList.add('hidden');
    } else {
      heroImg.classList.add('hidden');
    }

    // Показываем summary пока грузится полная статья
    textEl.innerHTML = `<p>${summary.extract_html || summary.extract || ''}</p>`;
    loading.classList.add('hidden');
    content.classList.remove('hidden');

    // Догружаем полный mobile-html
    const html = await fetchWikiMobileHtml(item._title_key || item.title);
    const doc  = new DOMParser().parseFromString(html, 'text/html');

    // Убираем ненужное
    ['figure.mwe-math-element','sup.reference','.hatnote',
      '.navbox','.sistersitebox','.stub','table','.mbox-small',
      '#toc','.mw-editsection','style','script'
    ].forEach(sel => {
      try { doc.querySelectorAll(sel).forEach(e => e.remove()); } catch {}
    });

    // Берём секции
    const sections = doc.querySelectorAll('section');
    const frag = document.createDocumentFragment();

    sections.forEach((sec, idx) => {
      if (idx > 6) return; // первые 6 секций
      const heading = sec.querySelector('h2,h3');
      if (heading) {
        const h = document.createElement('h2');
        h.textContent = heading.textContent;
        frag.appendChild(h);
      }
      sec.querySelectorAll('p').forEach(p => {
        if (p.textContent.trim().length < 40) return;
        const clone = document.createElement('p');
        clone.textContent = p.textContent;
        frag.appendChild(clone);
      });
    });

    if (frag.childNodes.length > 0) {
      textEl.innerHTML = '';
      textEl.appendChild(frag);
    }

  } catch (e) {
    console.error(e);
    loading.classList.add('hidden');
    error.classList.remove('hidden');
  }
}

async function fetchHNArticle(item) {
  const loading = document.getElementById('article-loading');
  const error   = document.getElementById('article-error');
  const content = document.getElementById('article-content');
  const titleEl = document.getElementById('article-title');
  const textEl  = document.getElementById('article-text');
  const heroImg = document.getElementById('article-hero-img');
  const extLink = document.getElementById('article-open-link');

  try {
    titleEl.textContent = item.title;
    heroImg.classList.add('hidden');

    // Получаем детали HN поста
    const story = await fetch(`https://hacker-news.firebaseio.com/v0/item/${item._hn_id}.json`).then(r => r.json());

    let html = '';
    if (story.text) {
      html = `<div>${story.text}</div>`;
    }

    // Топ комментарии
    if (story.kids?.length) {
      const commentIds = story.kids.slice(0, 8);
      const comments   = await Promise.all(
          commentIds.map(id =>
              fetch(`https://hacker-news.firebaseio.com/v0/item/${id}.json`).then(r => r.json()).catch(() => null)
          )
      );

      html += `<h2>Обсуждение (${story.descendants || 0} комментариев)</h2>`;
      comments.filter(c => c && c.text && !c.deleted && !c.dead).forEach(c => {
        const author = c.by || 'anonymous';
        const text   = c.text?.replace(/<[^>]+>/g, '').slice(0, 400) || '';
        html += `<div class="hn-comment"><strong>${author}</strong><p>${text}</p></div>`;
      });
    }

    if (!html) {
      html = `<p>Откройте оригинальный материал по ссылке.</p>`;
    }

    textEl.innerHTML = html;

    // Метаинфо
    const meta = document.createElement('div');
    meta.className = 'hn-meta';
    meta.innerHTML = `
            <a href="${item._hn_url || '#'}" target="_blank">🔗 ${extractDomain(item._hn_url || '') || 'Источник'}</a>
            <span>⬆ ${story.score} очков</span>
            <a href="https://news.ycombinator.com/item?id=${story.id}" target="_blank">💬 HN Discussion</a>
        `;
    textEl.prepend(meta);

    loading.classList.add('hidden');
    content.classList.remove('hidden');

  } catch (e) {
    loading.classList.add('hidden');
    error.classList.remove('hidden');
  }
}

function closeArticleModal(e) {
  if (e && e.target !== document.getElementById('article-overlay')) return;
  document.getElementById('article-overlay').classList.add('hidden');
  document.body.style.overflow = '';
}

document.addEventListener('keydown', e => {
  if (e.key === 'Escape') {
    const overlay = document.getElementById('article-overlay');
    if (overlay && !overlay.classList.contains('hidden')) {
      overlay.classList.add('hidden');
      document.body.style.overflow = '';
    }
  }
});