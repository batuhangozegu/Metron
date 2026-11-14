export const THEME_STORAGE_KEY = "metron-theme";

/** Runs synchronously in head, before themed content can paint. Storage may be unavailable. */
export const themeBootstrap = `(()=>{let theme='dark';try{if(localStorage.getItem('${THEME_STORAGE_KEY}')==='light')theme='light'}catch{}document.documentElement.dataset.theme=theme})()`;
