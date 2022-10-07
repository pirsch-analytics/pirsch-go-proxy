(function () {
    "use strict";

    // respect Do-Not-Track
    if (navigator.doNotTrack === "1" || localStorage.getItem("disable_pirsch")) {
        return;
    }

    const script = document.querySelector("#pirschjs");

    // include pages
    try {
        const include = script.getAttribute("data-include");
        const paths = include ? include.split(",") : [];

        if (paths.length) {
            let found = false;

            for (let i = 0; i < paths.length; i++) {
                if (new RegExp(paths[i]).test(location.pathname)) {
                    found = true;
                    break;
                }
            }

            if (!found) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
    }

    // exclude pages
    try {
        const exclude = script.getAttribute("data-exclude");
        const paths = exclude ? exclude.split(",") : [];

        for (let i = 0; i < paths.length; i++) {
            if (new RegExp(paths[i]).test(location.pathname)) {
                return;
            }
        }
    } catch (e) {
        console.error(e);
    }

    // register hit function
    const endpoint = script.getAttribute("data-endpoint") || "/pirsch/hit";
    const disableQueryParams = script.hasAttribute("data-disable-query");
    const disableReferrer = script.hasAttribute("data-disable-referrer");
    const disableResolution = script.hasAttribute("data-disable-resolution");

    function hit() {
        const url = endpoint +
            "?nc=" + new Date().getTime() +
            "&url=" + encodeURIComponent(disableQueryParams ? (location.href.includes('?') ? location.href.split('?')[0] : location.href).substring(0, 1800) : location.href.substring(0, 1800)) +
            "&t=" + encodeURIComponent(document.title) +
            "&ref=" + (disableReferrer ? '' : encodeURIComponent(document.referrer)) +
            "&w=" + (disableResolution ? '' : screen.width) +
            "&h=" + (disableResolution ? '' : screen.height);
        const req = new XMLHttpRequest();
        req.open("GET", url);
        req.send();
    }

    if (history.pushState) {
        const pushState = history["pushState"];

        history.pushState = function () {
            pushState.apply(this, arguments);
            hit();
        }

        window.addEventListener("popstate", hit);
    }

    if (!document.body) {
        window.addEventListener("DOMContentLoaded", hit);
    } else {
        hit();
    }
})();
