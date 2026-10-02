// Farahdin share image: draws a 1080×1350 (4:5) card on a canvas with the
// browser's Canvas API (no library) and offers it through the Web Share API
// or as a download. Content comes from the data-card-* attributes rendered by
// components.ShareBar; every image is same-origin, so the canvas stays
// exportable. Business data is never computed here.
(function () {
  "use strict";

  var W = 1080, H = 1350, PAD = 96;
  var BG = "#231d32", GOLD = "#bd9c49", TEXT = "#e9e4f3", MUTED = "#a9a2bb";
  var SERIF = '"EB Garamond", Garamond, Georgia, serif';
  var SMALLCAPS = '"AGaramond SC", ' + SERIF;

  function supported() {
    return !!(window.HTMLCanvasElement && HTMLCanvasElement.prototype.toBlob && window.Promise);
  }

  // --- loading -----------------------------------------------------------------

  function load(src) {
    return new Promise(function (resolve, reject) {
      var img = new Image();
      img.decoding = "async";
      img.onload = function () { resolve(img); };
      img.onerror = reject;
      img.src = src;
    });
  }

  function maybe(src) {
    return src ? load(src).catch(function () { return null; }) : Promise.resolve(null);
  }

  function list(value) {
    return (value || "").split(" ").filter(Boolean);
  }

  function fonts() {
    if (!document.fonts || !document.fonts.load) return Promise.resolve();
    return Promise.all(["96px Javassoul", '56px "AGaramond SC"', '36px "EB Garamond"'].map(function (f) {
      return document.fonts.load(f).catch(function () {});
    }));
  }

  // The Matrix Destiny chart is the inline <svg id="matrix"> of the result sheet.
  function chart(bar) {
    var sheet = bar.closest("[data-result-sheet]") || document;
    var svg = sheet.querySelector("svg#matrix");
    if (!svg) return Promise.resolve(null);
    var clone = svg.cloneNode(true);
    clone.setAttribute("xmlns", "http://www.w3.org/2000/svg");
    clone.setAttribute("width", "720");
    clone.setAttribute("height", "660");
    clone.setAttribute("font-family", "Arial, Helvetica, sans-serif");
    clone.removeAttribute("class");
    var title = clone.querySelector("title");
    if (title) title.remove();
    var xml = new XMLSerializer().serializeToString(clone);
    return maybe("data:image/svg+xml;charset=utf-8," + encodeURIComponent(xml));
  }

  function assets(bar) {
    var d = bar.dataset;
    return Promise.all([
      fonts(),
      maybe(d.cardBg),
      Promise.all(list(d.cardImages).map(maybe)),
      Promise.all(list(d.cardIcons).map(maybe)),
      maybe(d.cardDecor),
      d.cardChart !== undefined ? chart(bar) : Promise.resolve(null),
      Number(d.cardLove) > 0 ? maybe(d.cardHeart) : Promise.resolve(null),
      Number(d.cardLove) > 0 ? maybe(d.cardNoheart) : Promise.resolve(null)
    ]).then(function (r) {
      return {
        bg: r[1], cards: r[2].filter(Boolean), icons: r[3].filter(Boolean), decor: r[4],
        chart: r[5], heart: r[6], noheart: r[7]
      };
    });
  }

  // --- drawing -----------------------------------------------------------------

  function roundRect(x, l, t, w, h, r) {
    x.beginPath();
    x.moveTo(l + r, t);
    x.arcTo(l + w, t, l + w, t + h, r);
    x.arcTo(l + w, t + h, l, t + h, r);
    x.arcTo(l, t + h, l, t, r);
    x.arcTo(l, t, l + w, t, r);
    x.closePath();
  }

  // fit returns a font string no wider than maxWidth for text (down to min px).
  function fit(x, text, family, size, maxWidth, min) {
    for (; size > (min || 24); size -= 2) {
      x.font = size + "px " + family;
      if (x.measureText(text).width <= maxWidth) break;
    }
    return x.font;
  }

  // row draws images side by side at one height (the frames line up, as on
  // the result sheet) and returns the height used.
  function row(x, imgs, top, height, gap) {
    var widths = imgs.map(function (i) { return i.naturalWidth * height / i.naturalHeight; });
    var total = widths.reduce(function (a, b) { return a + b; }, 0) + gap * (imgs.length - 1);
    var scale = Math.min(1, (W - 2 * PAD) / total);
    var left = (W - total * scale) / 2;
    imgs.forEach(function (img, i) {
      x.drawImage(img, left, top, widths[i] * scale, height * scale);
      left += (widths[i] + gap) * scale;
    });
    return height * scale;
  }

  function wrap(x, text, maxWidth) {
    var words = text.split(/\s+/).filter(Boolean), lines = [], line = "";
    words.forEach(function (w) {
      var test = line ? line + " " + w : w;
      if (x.measureText(test).width > maxWidth && line) {
        lines.push(line);
        line = w;
      } else {
        line = test;
      }
    });
    if (line) lines.push(line);
    return lines;
  }

  // body draws paragraphs from top to bottom limit, ending with "…" when cut.
  function body(x, text, top, bottom) {
    var size = 36, lh = 52, gap = 18, maxWidth = W - 2 * PAD;
    x.font = size + "px " + SERIF;
    x.fillStyle = TEXT;
    x.textAlign = "left";
    var y = top + size, paras = (text || "").split("\n");
    for (var p = 0; p < paras.length; p++) {
      var lines = wrap(x, paras[p], maxWidth);
      for (var i = 0; i < lines.length; i++) {
        var last = y + lh > bottom;
        var line = lines[i];
        if (last && (i < lines.length - 1 || p < paras.length - 1)) {
          while (line && x.measureText(line + "…").width > maxWidth) line = line.replace(/\s*\S+$/, "");
          line = line.replace(/[\s.,;:!?-]+$/, "");
          x.fillText(line + "…", PAD, y);
          return;
        }
        x.fillText(line, PAD, y);
        y += lh;
        if (y > bottom) return;
      }
      y += gap;
    }
  }

  function draw(d, a) {
    var canvas = document.createElement("canvas");
    canvas.width = W;
    canvas.height = H;
    var x = canvas.getContext("2d");

    // Background: theme colour, the Farahdin photo faded in at the top, a soft glow.
    x.fillStyle = BG;
    x.fillRect(0, 0, W, H);
    if (a.bg) {
      var bh = W * a.bg.naturalHeight / a.bg.naturalWidth;
      x.globalAlpha = 0.2;
      x.drawImage(a.bg, 0, 0, W, bh);
      x.globalAlpha = 1;
      var fade = x.createLinearGradient(0, bh * 0.3, 0, bh);
      fade.addColorStop(0, "rgba(35,29,50,0)");
      fade.addColorStop(1, BG);
      x.fillStyle = fade;
      x.fillRect(0, 0, W, bh + 1);
    }
    var glow = x.createRadialGradient(W / 2, H * 0.42, 40, W / 2, H * 0.42, W * 0.75);
    glow.addColorStop(0, "rgba(189,156,73,0.12)");
    glow.addColorStop(1, "rgba(189,156,73,0)");
    x.fillStyle = glow;
    x.fillRect(0, 0, W, H);

    // Frame.
    x.strokeStyle = "rgba(189,156,73,0.6)";
    x.lineWidth = 2;
    roundRect(x, 36, 36, W - 72, H - 72, 28);
    x.stroke();
    x.strokeStyle = "rgba(189,156,73,0.22)";
    x.lineWidth = 1;
    roundRect(x, 50, 50, W - 100, H - 100, 20);
    x.stroke();

    // Brand and heading.
    x.textAlign = "center";
    x.fillStyle = GOLD;
    x.font = "96px Javassoul, " + SERIF;
    x.fillText("Farahdin", W / 2, 176);
    var y = 176 + 84;
    x.fillStyle = "#ffffff";
    fit(x, d.cardHeading, SMALLCAPS, 56, W - 2 * PAD, 34);
    x.fillText(d.cardHeading, W / 2, y);
    if (d.cardHighlight) {
      y += 56;
      x.fillStyle = GOLD;
      fit(x, d.cardHighlight, SERIF, 36, W - 2 * PAD, 24);
      x.fillText(d.cardHighlight, W / 2, y);
    }
    y += 40;

    // Media: tarot cards, matrix chart, zodiac signs or the feature illustration.
    var used = 0;
    if (a.cards.length) used = row(x, a.cards, y, 400, 32);
    else if (a.chart) used = row(x, [a.chart], y, 470, 0);
    else if (a.icons.length) used = row(x, a.icons, y, 180, 56);
    else if (a.decor) used = row(x, [a.decor], y, 220, 0);
    if (used) y += used + 28;

    var love = Number(d.cardLove);
    if (love > 0 && a.heart) {
      var hearts = [];
      for (var i = 1; i <= 5; i++) hearts.push(i <= love ? a.heart : (a.noheart || a.heart));
      y += row(x, hearts, y, 72, 16) + 28;
    }

    // Footer.
    var footer = H - 176;
    x.strokeStyle = "rgba(189,156,73,0.45)";
    x.lineWidth = 1.5;
    x.beginPath();
    x.moveTo(PAD, footer);
    x.lineTo(W - PAD, footer);
    x.stroke();
    x.textAlign = "center";
    x.fillStyle = GOLD;
    fit(x, d.cardCta, SMALLCAPS, 40, W - 2 * PAD, 26);
    x.fillText(d.cardCta, W / 2, footer + 66);
    x.fillStyle = MUTED;
    x.font = "30px " + SERIF;
    x.fillText(d.cardHost, W / 2, footer + 112);

    // Reading text in the space that is left.
    body(x, d.cardBody, y, footer - 60);
    return canvas;
  }

  // --- share panel -----------------------------------------------------------

  function slug(s) {
    return (s || "hasil").toLowerCase().normalize("NFKD").replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "hasil";
  }

  // render draws the card once per panel and remembers the file for sharing.
  function render(slot) {
    if (slot._card) return slot._card;
    var panel = slot.querySelector("[data-share-panel]");
    var preview = panel.querySelector("[data-share-preview]");
    var loading = panel.querySelector("[data-share-loading]");
    var download = panel.querySelector("[data-share-download]");
    var name = "farahdin-" + slug(slot.dataset.cardHeading) + ".jpg";
    slot._card = assets(slot).then(function (a) {
      var canvas = draw(slot.dataset, a);
      var url = canvas.toDataURL("image/jpeg", 0.9);
      preview.src = url;
      preview.hidden = false;
      loading.hidden = true;
      download.href = url;
      download.download = name;
      download.hidden = false;
      return new Promise(function (resolve) {
        canvas.toBlob(function (blob) {
          resolve(blob && window.File ? new File([blob], name, { type: "image/jpeg" }) : null);
        }, "image/jpeg", 0.9);
      });
    }).catch(function (err) {
      loading.hidden = true;
      panel.querySelector("[data-share-error]").hidden = false;
      slot._card = null;
      throw err;
    });
    return slot._card;
  }

  // Opening the panel prepares the image, so tapping Instagram/TikTok can open
  // the share sheet straight away (browsers require a fresh user gesture).
  document.addEventListener("toggle", function (e) {
    var panel = e.target;
    if (!panel.matches || !panel.matches("[data-share-panel]") || e.newState !== "open") return;
    var native = panel.querySelector("[data-share-native]");
    var copy = panel.querySelector("[data-share-copy]");
    if (native && navigator.share) native.hidden = false;
    if (copy && navigator.clipboard && window.isSecureContext) copy.hidden = false;
    if (!supported()) return;
    var slot = panel.closest("[data-share]");
    panel.querySelector("[data-share-image-area]").hidden = false;
    panel.querySelectorAll("[data-share-app]").forEach(function (b) { b.hidden = false; });
    render(slot).then(function (file) { slot._file = file; }).catch(function () {});
  }, true);

  function canShareFile(file) {
    return !!(file && navigator.canShare && navigator.share && navigator.canShare({ files: [file] }));
  }

  document.addEventListener("click", function (e) {
    var app = e.target.closest("[data-share-app]");
    if (!app) return;
    var slot = app.closest("[data-share]");
    var panel = slot.querySelector("[data-share-panel]");
    var hint = panel.querySelector("[data-share-hint]");
    hint.hidden = true;

    // Instagram and TikTok have no web share links: the only way in from a
    // website is sharing the image file through the system share sheet,
    // where the user picks the app. Elsewhere the image is downloaded.
    if (canShareFile(slot._file)) {
      navigator.share({ files: [slot._file] }).catch(function () { /* cancelled */ });
      return;
    }
    render(slot).then(function (file) {
      slot._file = file;
      if (canShareFile(file)) {
        navigator.share({ files: [file] }).catch(function () {});
        return;
      }
      panel.querySelector("[data-share-download]").click();
      hint.hidden = false;
    }).catch(function () {});
  });
})();
