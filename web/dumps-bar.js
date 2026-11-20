(() => {
  const space = { top: 0, bottom: 0 };
  const hosts = { top: null, bottom: null };
  let clean = false;
  const reserve = () => {
    let el = document.getElementById("dumps-chrome-space");
    if (!el) {
      el = document.createElement("style");
      el.id = "dumps-chrome-space";
      (document.head || document.documentElement).appendChild(el);
    }
    el.textContent = `html{transition:padding-top .22s ease,padding-bottom .22s ease;padding-top:${space.top}px!important;padding-bottom:${space.bottom}px!important;scroll-padding-top:${space.top}px}@media print{html{padding:0!important}}`;
  };
  const track = (host, side) => {
    hosts[side] = host;
    const update = () => { space[side] = clean ? 0 : Math.ceil(host.getBoundingClientRect().height); reserve(); };
    host._dumpsUpdate = update;
    update();
    if ("ResizeObserver" in window) new ResizeObserver(update).observe(host);
    else addEventListener("resize", update);
  };

  const ICON = '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path d="M12 2 3 6.5 12 11l9-4.5L12 2Z" fill="currentColor" opacity=".95"/><path d="m3 11.5 9 4.5 9-4.5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" opacity=".7"/><path d="m3 16.5 9 4.5 9-4.5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" opacity=".45"/></svg>';
  const SHARE = '<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M12 3v12M7 8l5-5 5 5M5 13v6a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>';
  const WA = '<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path fill="currentColor" d="M12 2a10 10 0 0 0-8.6 15.1L2 22l5-1.3A10 10 0 1 0 12 2Zm0 18.2a8.2 8.2 0 0 1-4.2-1.2l-.3-.2-3 .8.8-2.9-.2-.3A8.2 8.2 0 1 1 12 20.2Zm4.5-6.1c-.2-.1-1.5-.7-1.7-.8-.2-.1-.4-.1-.6.1l-.8 1c-.1.2-.3.2-.5.1a6.7 6.7 0 0 1-3.3-2.9c-.3-.4.2-.4.7-1.4.1-.2 0-.3 0-.4l-.8-1.8c-.2-.5-.4-.4-.6-.4h-.5a.9.9 0 0 0-.7.3 2.8 2.8 0 0 0-.9 2.1 4.9 4.9 0 0 0 1 2.6 11.2 11.2 0 0 0 4.3 3.8c1.6.7 2.2.7 3 .6a2.5 2.5 0 0 0 1.7-1.2 2 2 0 0 0 .1-1.2c0-.1-.2-.2-.4-.3Z"/></svg>';
  const TG = '<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path fill="currentColor" d="M21.9 4.3 18.7 19.4c-.2 1-.9 1.3-1.7.8l-4.8-3.5-2.3 2.2c-.3.3-.5.5-1 .5l.3-4.9 8.9-8c.4-.3-.1-.5-.6-.2L6.5 13.2l-4.7-1.5c-1-.3-1-1 .2-1.5l18.4-7.1c.9-.3 1.6.2 1.5 1.2Z"/></svg>';
  const EYE_OFF = '<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M3 3l18 18M10.6 5.1A10.5 10.5 0 0 1 12 5c5 0 9 4.5 10 7a12.6 12.6 0 0 1-3.1 4.2M6.3 6.3A12.7 12.7 0 0 0 2 12c1 2.5 5 7 10 7a10 10 0 0 0 4.3-1M9.9 9.9a3 3 0 0 0 4.2 4.2" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>';

  const setClean = (on) => {
    clean = on;
    for (const side of ["top", "bottom"]) {
      const h = hosts[side];
      if (!h) continue;
      if (on) h.setAttribute("data-clean", ""); else h.removeAttribute("data-clean");
      if (h._dumpsUpdate) h._dumpsUpdate();
    }
  };

  const notice = (msg) => {
    const host = document.createElement("div");
    const sr = host.attachShadow({ mode: "open" });
    sr.innerHTML = `<style>
      :host{all:initial;position:fixed;top:16px;right:16px;z-index:2147483647}
      .t{box-sizing:border-box;max-width:min(360px,calc(100vw - 32px));padding:12px 14px;border-radius:12px;
        background:rgba(28,28,30,.92);color:#f2f2f7;box-shadow:0 12px 32px rgba(0,0,0,.25);
        -webkit-backdrop-filter:blur(12px);backdrop-filter:blur(12px);
        font:500 14px/1.4 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,system-ui,sans-serif;
        opacity:0;transform:translateY(-8px);transition:opacity .2s,transform .2s}
      .t{display:flex;align-items:flex-start;gap:10px}
      .t.show{opacity:1;transform:none}
      .x{all:unset;cursor:pointer;flex:none;width:22px;height:22px;margin:-2px -4px 0 0;border-radius:6px;display:grid;place-items:center;color:#c7c7cc;font-size:18px;line-height:1}
      .x:hover{background:rgba(255,255,255,.12);color:#fff}.x:focus-visible{outline:2px solid #a593ff}
      b{font-weight:650}
      kbd{font:600 12px/1 ui-monospace,SFMono-Regular,Menlo,monospace;padding:2px 6px;border-radius:5px;background:rgba(255,255,255,.14)}
      @media (pointer:coarse){.k{display:none}}
    </style><div class="t" role="status" aria-live="polite"><span class="m"></span><button class="x" aria-label="Fechar aviso">×</button></div>`;
    const t = sr.querySelector(".t");
    sr.querySelector(".m").innerHTML = msg;
    (document.body || document.documentElement).appendChild(host);
    requestAnimationFrame(() => t.classList.add("show"));
    let gone = false;
    const dismiss = () => {
      if (gone) return;
      gone = true;
      t.classList.remove("show");
      setTimeout(() => host.remove(), 250);
    };
    sr.querySelector(".x").addEventListener("click", dismiss);
    setTimeout(dismiss, 6000);
    return dismiss;
  };

  let dismissNotice = null;
  const enterClean = () => {
    setClean(true);
    dispatchEvent(new CustomEvent("dumps-clean"));
    if (dismissNotice) dismissNotice();
    dismissNotice = notice('<b>Modo limpo.</b> Recarregue a página para ver a barra de novo<span class="k">, ou aperte <kbd>Esc</kbd></span>.');
  };
  const leaveClean = () => {
    setClean(false);
    if (dismissNotice) { dismissNotice(); dismissNotice = null; }
  };

  const TIP_COOKIE = "dumps_tip";
  const tipSeen = () => document.cookie.split(";").some((c) => c.trim().startsWith(TIP_COOKIE + "="));
  const markTip = () => {
    const secure = location.protocol === "https:" ? "; Secure" : "";
    document.cookie = `${TIP_COOKIE}=1; Max-Age=86400; Path=/; SameSite=Lax${secure}`;
  };
  const typing = (el) => !!el && (el.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(el.tagName));

  const LINK = '<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M10 14a5 5 0 0 0 7.1 0l3-3a5 5 0 0 0-7.1-7.1l-1.4 1.4M14 10a5 5 0 0 0-7.1 0l-3 3a5 5 0 0 0 7.1 7.1l1.4-1.4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>';

  const CSS = `
    :host{all:initial;display:block;position:fixed;top:0;left:0;right:0;width:100%;z-index:2147483647;
      --bg:rgba(255,255,255,.82);--fg:#1c1c1e;--mut:#6e6e73;--line:rgba(0,0,0,.08);--acc:#5b3fd1;--accfg:#fff;--menu:#fff;--hover:rgba(0,0,0,.05)}
    @media (prefers-color-scheme:dark){:host{--bg:rgba(22,22,26,.82);--fg:#f2f2f7;--mut:#98989f;--line:rgba(255,255,255,.1);--acc:#a593ff;--accfg:#16161a;--menu:#232328;--hover:rgba(255,255,255,.07)}}
    .bar{box-sizing:border-box;height:52px;display:grid;grid-template-columns:minmax(0,1fr) minmax(0,2fr) minmax(0,1fr);align-items:center;gap:12px;
      padding:0 max(16px,env(safe-area-inset-left)) 0 max(16px,env(safe-area-inset-left));
      background:var(--bg);color:var(--fg);border-bottom:1px solid var(--line);
      -webkit-backdrop-filter:saturate(180%) blur(14px);backdrop-filter:saturate(180%) blur(14px);
      font:500 14px/1.2 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,system-ui,sans-serif;-webkit-font-smoothing:antialiased}
    .brand{display:flex;align-items:center;gap:8px;min-width:0;color:var(--fg);text-decoration:none}
    .brand svg{color:var(--acc);flex:none}
    .brand span{white-space:nowrap;overflow:hidden;text-overflow:ellipsis;letter-spacing:-.01em}
    .brand b{font-weight:650}.brand i{font-style:normal;color:var(--mut)}
    .title{text-align:center;font-weight:600;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;min-width:0}
    :host{transition:transform .22s ease,opacity .22s ease,visibility 0s}
    :host([data-clean]){transform:translateY(-100%);opacity:0;pointer-events:none;visibility:hidden;transition:transform .22s ease,opacity .22s ease,visibility 0s .22s}
    .right{display:flex;justify-content:flex-end;align-items:center;gap:6px;position:relative}
    .clean{display:inline-flex;align-items:center;justify-content:center;width:34px;height:34px;border-radius:999px;color:var(--mut)}
    .clean:hover{background:var(--hover);color:var(--fg)}
    .clean:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
    .cw{position:relative;display:inline-flex}
    .tip{position:absolute;top:46px;right:-6px;box-sizing:border-box;width:max-content;max-width:min(280px,calc(100vw - 24px));
      display:flex;align-items:flex-start;gap:8px;padding:10px 10px 10px 12px;border-radius:12px;background:var(--fg);color:var(--bg);
      font:500 13px/1.4 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,system-ui,sans-serif;box-shadow:0 10px 28px rgba(0,0,0,.22);
      opacity:0;transform:translateY(-6px);transition:opacity .2s,transform .2s;pointer-events:none}
    .tip:before{content:"";position:absolute;top:-5px;right:18px;width:10px;height:10px;background:inherit;transform:rotate(45deg);border-radius:2px}
    .tip.show{opacity:1;transform:none;pointer-events:auto}
    .tip kbd{font:600 12px/1 ui-monospace,SFMono-Regular,Menlo,monospace;padding:2px 6px;border-radius:5px;background:color-mix(in srgb,var(--bg) 22%,transparent)}
    .tip .x{all:unset;cursor:pointer;flex:none;width:20px;height:20px;border-radius:6px;display:grid;place-items:center;font-size:16px;line-height:1;opacity:.75}
    .tip .x:hover{opacity:1;background:color-mix(in srgb,var(--bg) 18%,transparent)}
    button{all:unset;box-sizing:border-box;cursor:pointer;-webkit-tap-highlight-color:transparent}
    .share{display:inline-flex;align-items:center;gap:6px;height:34px;padding:0 14px;border-radius:999px;background:var(--acc);color:var(--accfg);font-weight:600;font-size:14px}
    .share:hover{filter:brightness(1.08)}.share:active{transform:scale(.97)}
    .share:focus-visible,.item:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
    .menu{position:absolute;top:44px;right:0;min-width:220px;padding:6px;border-radius:14px;background:var(--menu);color:var(--fg);
      border:1px solid var(--line);box-shadow:0 12px 32px rgba(0,0,0,.18),0 2px 6px rgba(0,0,0,.08);
      opacity:0;transform:translateY(-4px) scale(.98);transform-origin:top right;pointer-events:none;transition:opacity .14s,transform .14s}
    .menu.open{opacity:1;transform:none;pointer-events:auto}
    .item{display:flex;align-items:center;gap:10px;width:100%;padding:10px 12px;border-radius:9px;font-size:14px;font-weight:500}
    .item:hover{background:var(--hover)}
    .wa svg{color:#25d366}.tg svg{color:#2aabee}.cp svg{color:var(--mut)}
    .toast{position:fixed;left:50%;bottom:24px;transform:translate(-50%,12px);padding:10px 16px;border-radius:999px;background:var(--fg);color:var(--bg);
      font:600 14px/1 -apple-system,BlinkMacSystemFont,system-ui,sans-serif;opacity:0;pointer-events:none;transition:opacity .18s,transform .18s}
    .toast.show{opacity:1;transform:translate(-50%,0)}
    @media (max-width:560px){.bar{grid-template-columns:auto minmax(0,1fr) auto;gap:10px}.brand span{display:none}.share .label{display:none}.share{width:36px;padding:0;justify-content:center}}
    @media print{:host{display:none}}
  `;

  class DumpsBar extends HTMLElement {
    connectedCallback() {
      if (this.shadowRoot) return;
      const root = this.attachShadow({ mode: "open" });
      const title = this.getAttribute("dump-title") || document.title || "dump";
      const url = this.getAttribute("dump-url") || location.href;
      root.innerHTML = `<style>${CSS}</style>
        <div class="bar" role="banner">
          <div class="brand">${ICON}<span><b></b><i></i></span></div>
          <div class="title"></div>
          <div class="right">
            <span class="cw"><button class="clean" aria-label="Modo limpo: esconder barra e rodapé (Esc)" title="Modo limpo (Esc)">${EYE_OFF}</button>
              <span class="tip" role="note"><span class="tm"></span><button class="x" aria-label="Fechar dica">×</button></span></span>
            <button class="share" aria-haspopup="menu" aria-expanded="false">${SHARE}<span class="label">Compartilhar</span></button>
            <div class="menu" role="menu">
              <button class="item wa" role="menuitem" data-k="whatsapp">${WA}WhatsApp</button>
              <button class="item tg" role="menuitem" data-k="telegram">${TG}Telegram</button>
              <button class="item cp" role="menuitem" data-k="copy">${LINK}Copiar link</button>
            </div>
          </div>
        </div>
        <div class="toast" role="status" aria-live="polite"></div>`;
      const host = location.hostname || "dumps";
      const dot = host.indexOf(".");
      root.querySelector(".brand b").textContent = dot > 0 ? host.slice(0, dot) : host;
      root.querySelector(".brand i").textContent = dot > 0 ? host.slice(dot) : "";
      track(this, "top");
      root.querySelector(".title").textContent = title;
      root.querySelector(".title").title = title;
      const btn = root.querySelector(".share");
      const menu = root.querySelector(".menu");
      const toast = root.querySelector(".toast");
      const setOpen = (open) => { menu.classList.toggle("open", open); btn.setAttribute("aria-expanded", String(open)); };
      const say = (msg) => { toast.textContent = msg; toast.classList.add("show"); clearTimeout(this._t); this._t = setTimeout(() => toast.classList.remove("show"), 1800); };
      const emit = (method) => this.dispatchEvent(new CustomEvent("dumps-share", { bubbles: true, composed: true, detail: { method, url } }));
      const touch = matchMedia("(pointer: coarse)").matches;
      const tip = root.querySelector(".tip");
      const hideTip = () => tip.classList.remove("show");
      tip.querySelector(".x").addEventListener("click", (e) => { e.stopPropagation(); hideTip(); });
      if (!tipSeen()) {
        markTip();
        root.querySelector(".tm").innerHTML = touch
          ? "<b>Dica:</b> toque no olho para o modo limpo, sem barra e sem rodapé."
          : "<b>Dica:</b> aperte <kbd>Esc</kbd> para o modo limpo, sem barra e sem rodapé.";
        setTimeout(() => tip.classList.add("show"), 700);
        setTimeout(hideTip, 7700);
      }
      root.querySelector(".clean").addEventListener("click", (e) => {
        e.stopPropagation();
        setOpen(false);
        hideTip();
        enterClean();
      });

      btn.addEventListener("click", async (e) => {
        e.stopPropagation();
        if (touch && navigator.share) {
          try { await navigator.share({ title, url }); emit("native"); return; }
          catch (err) { if (err && err.name === "AbortError") return; }
        }
        setOpen(!menu.classList.contains("open"));
      });
      menu.addEventListener("click", async (e) => {
        const k = e.target.closest("[data-k]")?.dataset.k;
        if (!k) return;
        setOpen(false);
        const text = `${title} ${url}`;
        if (k === "whatsapp") window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, "_blank", "noopener");
        if (k === "telegram") window.open(`https://t.me/share/url?url=${encodeURIComponent(url)}&text=${encodeURIComponent(title)}`, "_blank", "noopener");
        if (k === "copy") {
          try { await navigator.clipboard.writeText(url); }
          catch { const t = document.createElement("textarea"); t.value = url; document.body.appendChild(t); t.select(); document.execCommand("copy"); t.remove(); }
          say("Link copiado ✓");
        }
        emit(k);
      });
      document.addEventListener("click", () => setOpen(false));
      document.addEventListener("keydown", (e) => {
        if (e.key !== "Escape" || e.repeat || e.isComposing || e.defaultPrevented || typing(document.activeElement)) return;
        if (document.querySelector('dialog[open], [aria-modal="true"]')) return;
        if (menu.classList.contains("open")) { setOpen(false); return; }
        hideTip();
        if (clean) leaveClean(); else enterClean();
      });
    }
  }
  if (!customElements.get("dumps-bar")) customElements.define("dumps-bar", DumpsBar);

  const FOOTER_CSS = `
    :host{all:initial;display:block;position:fixed;left:0;right:0;bottom:0;width:100%;z-index:2147483646;transition:transform .22s ease,opacity .22s ease,visibility 0s;
      --fg:#1c1c1e;--mut:#6e6e73;--line:rgba(0,0,0,.08);--acc:#5b3fd1;--bg:#fafafa}
    @media (prefers-color-scheme:dark){:host{--fg:#f2f2f7;--mut:#98989f;--line:rgba(255,255,255,.1);--acc:#a593ff;--bg:#16161a}}
    footer{box-sizing:border-box;display:flex;align-items:center;justify-content:space-between;gap:12px 24px;flex-wrap:wrap;
      padding:14px max(16px,env(safe-area-inset-left));border-top:1px solid var(--line);background:var(--bg);color:var(--mut);
      font:400 14px/1.4 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,system-ui,sans-serif;-webkit-font-smoothing:antialiased}
    a{color:var(--fg);text-decoration:none;font-weight:500}
    a:hover{color:var(--acc);text-decoration:underline;text-underline-offset:3px}
    a:focus-visible{outline:2px solid var(--acc);outline-offset:2px;border-radius:4px}
    .url{min-width:0;overflow-wrap:anywhere}
    .by{white-space:nowrap}
    :host([data-clean]){transform:translateY(100%);opacity:0;pointer-events:none;visibility:hidden;transition:transform .22s ease,opacity .22s ease,visibility 0s .22s}
    @media print{:host{position:static}footer{background:none;border:0}}
  `;

  const safeUrl = (u) => { try { const x = new URL(u, location.href); return /^https?:$/.test(x.protocol) ? x.href : null; } catch { return null; } };

  class DumpsFooter extends HTMLElement {
    connectedCallback() {
      if (this.shadowRoot) return;
      const root = this.attachShadow({ mode: "open" });
      const url = this.getAttribute("dump-url") || location.href;
      let authors = [];
      try { authors = JSON.parse(this.getAttribute("authors") || "[]"); } catch { authors = []; }
      track(this, "bottom");
      root.innerHTML = `<style>${FOOTER_CSS}</style><footer role="contentinfo"><a class="url"></a><span class="by"></span></footer>`;
      const link = root.querySelector(".url");
      link.href = safeUrl(url) || location.href;
      link.textContent = url;
      const by = root.querySelector(".by");
      const list = authors.filter((a) => a && a.name);
      if (!list.length) { by.remove(); }
      if (list.length) by.append("por ");
      if (list.length) list.forEach((a, i) => {
        if (i) by.append(i === list.length - 1 ? " e " : ", ");
        const href = a.url && safeUrl(a.url);
        const el = document.createElement(href ? "a" : "span");
        if (href) { el.href = href; el.target = "_blank"; el.rel = "noopener"; }
        el.textContent = a.name;
        by.append(el);
      });
    }
  }
  if (!customElements.get("dumps-footer")) customElements.define("dumps-footer", DumpsFooter);
})();
