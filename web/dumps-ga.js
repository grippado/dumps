(() => {
  const id = "__DUMPS_GA_ID__";
  if (window.__dumpsGa || !/^G-[A-Z0-9]+$/.test(id)) return;
  window.__dumpsGa = true;
  const KEY = "dumps-consent";
  const dump = location.pathname.split("/").filter(Boolean)[0] || "home";
  const stored = () => { try { return localStorage.getItem(KEY); } catch (e) { return null; } };
  const store = (v) => { try { localStorage.setItem(KEY, v); } catch (e) {} };
  window.dataLayer = window.dataLayer || [];
  window.gtag = window.gtag || function () { window.dataLayer.push(arguments); };
  const denied = { analytics_storage: "denied", ad_storage: "denied", ad_user_data: "denied", ad_personalization: "denied" };
  window.gtag("consent", "default", denied);
  if (stored() === "granted") window.gtag("consent", "update", { analytics_storage: "granted" });
  window.gtag("js", new Date());
  window.gtag("config", id, { dump_id: dump });
  const s = document.createElement("script");
  s.async = true;
  s.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(id);
  (document.head || document.documentElement).appendChild(s);
  addEventListener("dumps-share", (e) => {
    window.gtag("event", "share", { method: (e.detail && e.detail.method) || "unknown", content_type: "dump", item_id: dump });
  });
  addEventListener("dumps-clean", () => window.gtag("event", "dumps_clean_mode", { dump_id: dump }));

  if (stored()) return;
  const pt = (document.documentElement.lang || navigator.language || "").toLowerCase().indexOf("pt") === 0;
  const t = pt
    ? { msg: "Usamos o Google Analytics, só depois do seu aceite, para contar visitas e compartilhamentos. Nada de anúncios.", ok: "Aceitar", no: "Recusar", label: "Aviso de cookies" }
    : { msg: "We use Google Analytics, only after you accept, to count visits and shares. No ads.", ok: "Accept", no: "Decline", label: "Cookie notice" };
  const show = () => {
    const host = document.createElement("div");
    const root = host.attachShadow({ mode: "open" });
    root.innerHTML = `<style>
      :host{all:initial}
      .c{position:fixed;left:16px;right:16px;bottom:72px;z-index:2147483646;max-width:420px;margin:0 auto 0 0;padding:14px 16px;border-radius:14px;
        font:14px/1.45 ui-sans-serif,system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;color:#1c1c1e;background:#fff;
        border:1px solid rgba(0,0,0,.12);box-shadow:0 10px 30px rgba(0,0,0,.18);display:flex;flex-direction:column;gap:10px}
      @media (prefers-color-scheme:dark){.c{color:#f2f2f7;background:#1c1c22;border-color:rgba(255,255,255,.14)}}
      .r{display:flex;gap:8px;justify-content:flex-end}
      button{font:600 13px/1 inherit;font-family:inherit;padding:9px 14px;border-radius:999px;cursor:pointer;border:1px solid currentColor;background:transparent;color:inherit}
      button.ok{background:#5b3fd1;border-color:#5b3fd1;color:#fff}
      button:focus-visible{outline:2px solid #a593ff;outline-offset:2px}
      @media print{.c{display:none}}
    </style><div class="c" role="region" aria-label="${t.label}"><span>${t.msg}</span><div class="r"><button class="no" type="button">${t.no}</button><button class="ok" type="button">${t.ok}</button></div></div>`;
    const close = (v) => { store(v); host.remove(); };
    root.querySelector(".ok").addEventListener("click", () => { window.gtag("consent", "update", { analytics_storage: "granted" }); close("granted"); });
    root.querySelector(".no").addEventListener("click", () => close("denied"));
    if (!document.querySelector("dumps-footer")) root.querySelector(".c").style.bottom = "16px";
    document.body.appendChild(host);
  };
  if (document.body) show(); else addEventListener("DOMContentLoaded", show);
})();
