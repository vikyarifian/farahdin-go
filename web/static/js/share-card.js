// Farahdin share image: draws a 1080×1920 (9:16, phone/story-sized) card on a canvas with the
// browser's Canvas API (no library) and offers it through the Web Share API
// or as a download. Content comes from the data-card-* attributes rendered by
// components.ShareBar; every image is same-origin, so the canvas stays
// exportable. Business data is never computed here.
(function () {
  "use strict";

  var W = 1080, H = 1920, PAD = 96;
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

  // rowHeight is the height a row of images takes at the given height
  // (scaled down when the row is wider than the content width).
  function rowScale(imgs, height, gap) {
    var total = imgs.reduce(function (sum, i) { return sum + i.naturalWidth * height / i.naturalHeight; }, 0) + gap * (imgs.length - 1);
    return { scale: Math.min(1, (W - 2 * PAD) / total), total: total };
  }

  // row draws images side by side at one height (the frames line up, as on
  // the result sheet) and returns the height used.
  function row(x, imgs, top, height, gap) {
    var r = rowScale(imgs, height, gap);
    var left = (W - r.total * r.scale) / 2;
    imgs.forEach(function (img) {
      var w = img.naturalWidth * height / img.naturalHeight;
      x.drawImage(img, left, top, w * r.scale, height * r.scale);
      left += (w + gap) * r.scale;
    });
    return height * r.scale;
  }

  // wrap breaks text into lines no wider than maxWidth at the current font.
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

  // sentences splits a paragraph at the spaces that follow . ! ? (with any
  // closing quotes). Dots inside words ("viky.arifian", "1.5") do not split,
  // so joining the parts with a space gives back the original text.
  function sentences(p) {
    return p.split(/(?<=[.!?]["'”’)\]]*)\s+/).map(function (s) { return s.trim(); }).filter(Boolean);
  }

  // layout wraps paragraphs (arrays of sentences) at a font size and returns
  // the lines with their y offsets and the total height.
  function layout(x, paras, size) {
    var lh = Math.round(size * 1.42), gap = Math.round(size * 0.5), maxWidth = W - 2 * PAD;
    x.font = size + "px " + SERIF;
    var lines = [], y = 0;
    paras.forEach(function (p, i) {
      if (i > 0) y += gap;
      wrap(x, p.join(" "), maxWidth).forEach(function (l) {
        lines.push({ text: l, y: y });
        y += lh;
      });
    });
    return { lines: lines, height: y, size: size };
  }

  // fitText shows all of the text when it fits in avail: the font shrinks
  // from 40px to 24px. If even 24px is too big, it keeps as many whole
  // sentences as fit, so the text never stops in the middle of a sentence.
  function fitText(x, text, avail) {
    var paras = (text || "").split("\n").map(sentences).filter(function (p) { return p.length; });
    if (!paras.length) return null;
    for (var size = 40; size >= 24; size -= 2) {
      var l = layout(x, paras, size);
      if (l.height <= avail) return l;
    }
    var kept = [];
    outer:
    for (var p = 0; p < paras.length; p++) {
      for (var s = 0; s < paras[p].length; s++) {
        var trial = kept.map(function (q) { return q.slice(); });
        if (s === 0) trial.push([paras[p][s]]); else trial[trial.length - 1].push(paras[p][s]);
        if (layout(x, trial, 24).height > avail) break outer;
        kept = trial;
      }
    }
    if (!kept.length) kept = [[paras[0][0]]]; // a single huge sentence: shown as far as it goes
    return layout(x, kept, 24);
  }

  function drawText(x, fit, top, bottom) {
    x.font = fit.size + "px " + SERIF;
    x.fillStyle = TEXT;
    x.textAlign = "left";
    fit.lines.forEach(function (l) {
      var baseline = top + l.y + fit.size;
      if (baseline <= bottom) x.fillText(l.text, PAD, baseline);
    });
  }

  // Vertical rhythm of the content block (heights include the gap below).
  var BRAND_H = 110, HEADING_H = 84, HIGHLIGHT_H = 56, MEDIA_GAP = 36, HEARTS_H = 60;
  var FOOTER_H = 56 + 116; // gap above the rule + rule, call to action and address

  function draw(d, a) {
    var canvas = document.createElement("canvas");
    canvas.width = W;
    canvas.height = H;
    var x = canvas.getContext("2d");

    // Background: theme colour, then — clipped to the inner frame — the
    // Farahdin photo centred between the top and the middle of the card,
    // faded out at both ends, and a soft glow.
    x.fillStyle = BG;
    x.fillRect(0, 0, W, H);
    x.save();
    roundRect(x, 52, 52, W - 104, H - 104, 18);
    x.clip();
    if (a.bg) {
      var bh = W * a.bg.naturalHeight / a.bg.naturalWidth;
      var top = Math.round(H * 0.32 - bh / 2);
      x.globalAlpha = 0.28;
      x.drawImage(a.bg, 0, top, W, bh);
      x.globalAlpha = 1;
      var fade = x.createLinearGradient(0, top, 0, top + bh);
      fade.addColorStop(0, BG);
      fade.addColorStop(0.18, "rgba(35,29,50,0)");
      fade.addColorStop(0.6, "rgba(35,29,50,0)");
      fade.addColorStop(1, BG);
      x.fillStyle = fade;
      x.fillRect(0, top - 1, W, bh + 2);
    }
    var glow = x.createRadialGradient(W / 2, H * 0.45, 40, W / 2, H * 0.45, W * 0.8);
    glow.addColorStop(0, "rgba(189,156,73,0.12)");
    glow.addColorStop(1, "rgba(189,156,73,0)");
    x.fillStyle = glow;
    x.fillRect(0, 0, W, H);
    x.restore();

    // Frame.
    x.strokeStyle = "rgba(189,156,73,0.6)";
    x.lineWidth = 2;
    roundRect(x, 36, 36, W - 72, H - 72, 28);
    x.stroke();
    x.strokeStyle = "rgba(189,156,73,0.22)";
    x.lineWidth = 1;
    roundRect(x, 50, 50, W - 100, H - 100, 20);
    x.stroke();

    // Content block (brand, heading, media, text and the footer line with
    // the call to action): measure it, then centre it inside the frame.
    var areaTop = 110, areaBottom = H - 110;
    var media = null;
    if (a.cards.length) media = { imgs: a.cards, h: 340, gap: 28 };
    else if (a.chart) media = { imgs: [a.chart], h: 410, gap: 0 };
    else if (a.icons.length) media = { imgs: a.icons, h: 150, gap: 48 };
    else if (a.decor) media = { imgs: [a.decor], h: 190, gap: 0 };
    var mediaH = media ? media.h * rowScale(media.imgs, media.h, media.gap).scale : 0;
    var love = Number(d.cardLove);
    var hearts = [];
    if (love > 0 && a.heart) {
      for (var i = 1; i <= 5; i++) hearts.push(i <= love ? a.heart : (a.noheart || a.heart));
    }
    var heartsH = hearts.length ? HEARTS_H * rowScale(hearts, HEARTS_H, 14).scale : 0;

    var head = BRAND_H + HEADING_H + (d.cardHighlight ? HIGHLIGHT_H : 0);
    var mid = (mediaH ? MEDIA_GAP + mediaH : 0) + (heartsH ? MEDIA_GAP + heartsH : 0) + MEDIA_GAP;
    var text = fitText(x, d.cardBody, areaBottom - areaTop - head - mid - FOOTER_H);
    var textH = text ? text.height : 0;
    var total = head + mid + textH + FOOTER_H;
    var y = areaTop + Math.max(0, Math.round((areaBottom - areaTop - total) / 2));

    x.textAlign = "center";
    x.fillStyle = GOLD;
    x.font = "96px Javassoul, " + SERIF;
    x.fillText("Farahdin", W / 2, y + 96);
    y += BRAND_H;
    x.fillStyle = "#ffffff";
    fit(x, d.cardHeading, SMALLCAPS, 56, W - 2 * PAD, 34);
    x.fillText(d.cardHeading, W / 2, y + 56);
    y += HEADING_H;
    if (d.cardHighlight) {
      x.fillStyle = GOLD;
      fit(x, d.cardHighlight, SERIF, 36, W - 2 * PAD, 24);
      x.fillText(d.cardHighlight, W / 2, y + 36);
      y += HIGHLIGHT_H;
    }
    if (media) {
      y += MEDIA_GAP;
      y += row(x, media.imgs, y, media.h, media.gap);
    }
    if (heartsH) {
      y += MEDIA_GAP;
      y += row(x, hearts, y, HEARTS_H, 14);
    }
    y += MEDIA_GAP;
    if (text) drawText(x, text, y, areaBottom - FOOTER_H);
    y += textH;

    // Footer text: a short gold rule, the call to action and the address,
    // right after the reading (not pinned to the bottom of the card).
    y += 56;
    x.strokeStyle = "rgba(189,156,73,0.5)";
    x.lineWidth = 1.5;
    x.beginPath();
    x.moveTo(W / 2 - 120, y);
    x.lineTo(W / 2 + 120, y);
    x.stroke();
    x.textAlign = "center";
    x.fillStyle = GOLD;
    fit(x, d.cardCta, SMALLCAPS, 40, W - 2 * PAD, 26);
    x.fillText(d.cardCta, W / 2, y + 62);
    x.fillStyle = MUTED;
    x.font = "30px " + SERIF;
    x.fillText(d.cardHost, W / 2, y + 106);
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
