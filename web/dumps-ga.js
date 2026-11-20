(() => {
  const id = "__DUMPS_GA_ID__";
  if (window.__dumpsGa || !/^G-[A-Z0-9]+$/.test(id)) return;
  window.__dumpsGa = true;
  const dump = location.pathname.split("/").filter(Boolean)[0] || "home";
  window.dataLayer = window.dataLayer || [];
  window.gtag = window.gtag || function () { window.dataLayer.push(arguments); };
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
})();
