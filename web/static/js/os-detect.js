// web/static/js/os-detect.js
//
// Static-first enhancement: the page renders complete and usable with JS off.
// This script only adds the "recommended" highlight on the matching downloads
// row and reveals the matching primary CTA variant (hiding the fallback).

(function () {
  "use strict";

  /**
   * Detect the visitor's OS+arch and return our target identifier
   * (matches asset names produced by .github/workflows/release.yml).
   *
   * @param {string} userAgent
   * @returns {"linux-amd64"|"darwin-arm64"|"windows-amd64"|null}
   */
  function detectTarget(userAgent) {
    var ua = (userAgent || "").toLowerCase();
    if (/windows/.test(ua)) return "windows-amd64";
    if (/mac os x|macintosh|iphone|ipad/.test(ua)) return "darwin-arm64";
    if (/linux|x11/.test(ua)) return "linux-amd64";
    return null;
  }

  // Expose for tests.html.
  window.__pzm = { detectTarget: detectTarget };

  document.addEventListener("DOMContentLoaded", function () {
    var target = detectTarget(navigator.userAgent);
    document.body.dataset.target = target || "unknown";
    if (!target) return;

    // Highlight downloads table row.
    var row = document.querySelector('tr[data-target="' + target + '"]');
    if (row) row.classList.add("recommended");

    // Reveal the matching CTA variant; hide the fallback.
    var cta = document.querySelector('.cta[data-target="' + target + '"]');
    if (cta) {
      cta.removeAttribute("hidden");
      var fallback = document.querySelector(".cta-fallback");
      if (fallback) fallback.setAttribute("hidden", "");
    }
    // If no per-target CTA exists (asset missing for that release), the
    // fallback remains visible — handles the "no Linux build for this version"
    // edge case automatically.
  });
})();
