// Farahdin browser enhancements. Everything here is optional: pages work
// without JavaScript. Keep it small; business rules live on the server.
(function () {
  "use strict";

  var isID = document.documentElement.lang === "id";
  function t(id, en) { return isID ? id : en; }

  // --- Service worker: registration and "new version" banner ---------------
  if ("serviceWorker" in navigator) {
    window.addEventListener("load", function () {
      navigator.serviceWorker.register("/sw.js").then(function (reg) {
        function offerUpdate(worker) {
          var banner = document.getElementById("update-banner");
          if (!banner || !navigator.serviceWorker.controller) return;
          banner.hidden = false;
          banner.querySelector("[data-sw-update]").addEventListener("click", function () {
            worker.postMessage("SKIP_WAITING");
          }, { once: true });
        }
        if (reg.waiting) offerUpdate(reg.waiting);
        reg.addEventListener("updatefound", function () {
          var worker = reg.installing;
          if (!worker) return;
          worker.addEventListener("statechange", function () {
            if (worker.state === "installed") offerUpdate(worker);
          });
        });
      }).catch(function () { /* the app works without a service worker */ });

      var reloading = false;
      navigator.serviceWorker.addEventListener("controllerchange", function () {
        if (reloading) return;
        reloading = true;
        window.location.reload();
      });
    });
  }

  // --- Online / offline banner ----------------------------------------------
  function updateNetwork() {
    var el = document.getElementById("network-status");
    if (el) el.hidden = navigator.onLine;
  }
  window.addEventListener("online", updateNetwork);
  window.addEventListener("offline", updateNetwork);
  document.addEventListener("DOMContentLoaded", updateNetwork);

  // --- Greeting by the device clock (ports getGreetingTime) ------------------
  document.addEventListener("DOMContentLoaded", function () {
    var el = document.querySelector("[data-greeting]");
    if (!el) return;
    var h = new Date().getHours();
    var key = h >= 5 && h < 12 ? "morning" : h >= 12 && h < 17 ? "afternoon" : h >= 17 && h < 21 ? "evening" : "night";
    el.textContent = el.dataset[key];
  });

  // --- Back links: behave like router.back() when we came from this app ----
  document.addEventListener("click", function (e) {
    var back = e.target.closest("a[data-back]");
    if (!back) return;
    var ref = document.referrer;
    if (ref && new URL(ref).origin === location.origin && history.length > 1) {
      e.preventDefault();
      history.back();
    }
  });

  // --- Result sheets ----------------------------------------------------------
  var lastFocus = null;

  function openSheet(sheet) {
    document.body.classList.add("sheet-open");
    var title = sheet.querySelector("#result-title");
    if (title) {
      title.setAttribute("tabindex", "-1");
      title.focus({ preventScroll: true });
    }
  }

  function closeSheet(sheet) {
    // Tarot: start over with a freshly shuffled deck (replace, so Back skips the old spread).
    if (sheet.dataset.reloadOnClose) {
      window.location.replace(sheet.dataset.reloadOnClose);
      return;
    }
    sheet.remove();
    if (!document.querySelector("[data-result-sheet]")) document.body.classList.remove("sheet-open");
    if (lastFocus && document.contains(lastFocus)) lastFocus.focus();
  }

  document.addEventListener("click", function (e) {
    var close = e.target.closest("[data-close-sheet]");
    if (!close) return;
    var sheet = close.closest("[data-result-sheet]");
    if (!sheet) return;
    e.preventDefault();
    closeSheet(sheet);
  });

  document.addEventListener("keydown", function (e) {
    if (e.key !== "Escape") return;
    var sheets = document.querySelectorAll("[data-result-sheet]");
    if (sheets.length) closeSheet(sheets[sheets.length - 1]);
  });

  document.addEventListener("htmx:beforeRequest", function () {
    lastFocus = document.activeElement;
  });

  document.addEventListener("htmx:afterSettle", function (e) {
    var sheet = e.target.querySelector && e.target.querySelector("[data-result-sheet]");
    if (sheet) openSheet(sheet);
  });

  // A sheet rendered by the server on a no-HTMX submit.
  document.addEventListener("DOMContentLoaded", function () {
    var sheet = document.querySelector("[data-result-sheet]");
    if (sheet) openSheet(sheet);
  });

  // --- Network failures during HTMX requests ---------------------------------
  document.addEventListener("htmx:sendError", function (e) {
    var target = e.detail.target;
    if (!target) return;
    var div = document.createElement("div");
    div.setAttribute("role", "alert");
    div.className = "my-3 rounded-lg border border-red-400/60 bg-red-950/40 p-3 text-sm text-red-200";
    div.textContent = t("Tidak ada koneksi. Periksa internet kamu lalu coba lagi.",
      "No connection. Check your internet and try again.");
    target.replaceChildren(div);
  });

  // --- Close the language dropdown when clicking elsewhere ------------------
  document.addEventListener("click", function (e) {
    document.querySelectorAll("details[data-dropdown][open]").forEach(function (d) {
      if (!d.contains(e.target)) d.removeAttribute("open");
    });
  });
})();
