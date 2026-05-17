// AntClaw MQL5/XM Calendar Proxy — v2 with debug
// Deploy to Cloudflare Workers
//
// Usage: https://antclaw-calendar.xxx.workers.dev/mql5

export default {
  async fetch(request) {
    const url = new URL(request.url);
    const path = url.pathname;

    // ---- Debug endpoints ----
    if (path === '/health') {
      return new Response('ok', { status: 200 });
    }

    if (path === '/debug/httpbin') {
      // Verify fetch() works at all
      try {
        const r = await fetch('https://httpbin.org/get');
        return new Response(r.body, { status: r.status,
          headers: { 'Content-Type': 'application/json' }});
      } catch(e) { return new Response('fetch failed: '+e.message, {status:500}); }
    }

    if (path === '/debug/mql5-ip') {
      // What IP does Cloudflare see MQL5 as?
      try {
        const r = await fetch('https://www.mql5.com', { method: 'GET',
          headers: { 'User-Agent': 'Mozilla/5.0' }});
        return new Response('MQL5 status: '+r.status+' / '+r.statusText,
          { status: 200 });
      } catch(e) { return new Response('MQL5 fetch failed: '+e.message, {status:500}); }
    }

    // ---- Production routes ----
    if (path === '/mql5') {
      return proxyMQL5(request);
    }

    if (path === '/xm') {
      return proxyXM(request);
    }

    return new Response('AntClaw Calendar Proxy. Use /mql5 or /xm', { status: 404 });
  }
};

async function proxyMQL5(request) {
  const body = await request.text();

  return fetch('https://www.mql5.com/en/economic-calendar/content', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      'Accept': 'application/json',
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36',
      'Referer': 'https://www.mql5.com/en/economic-calendar',
      'Origin': 'https://www.mql5.com',
      'X-Requested-With': 'XMLHttpRequest',
    },
    body: body,
  }).then(resp => {
    return new Response(resp.body, {
      status: resp.status,
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
        'Access-Control-Allow-Origin': '*',
      },
    });
  }).catch(err => {
    return new Response(JSON.stringify({error: 'mql5 fetch failed', detail: err.message}), {
      status: 502,
      headers: { 'Content-Type': 'application/json' },
    });
  });
}

async function proxyXM(request) {
  const body = await request.text();

  return fetch('https://www.xm.com/economic-calendar', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      'Accept': 'application/json',
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36',
    },
    body: body,
  }).then(resp => {
    return new Response(resp.body, {
      status: resp.status,
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
        'Access-Control-Allow-Origin': '*',
      },
    });
  }).catch(err => {
    return new Response(JSON.stringify({error: 'xm fetch failed', detail: err.message}), {
      status: 502,
      headers: { 'Content-Type': 'application/json' },
    });
  });
}
