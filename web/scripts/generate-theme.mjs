import { readFileSync, writeFileSync } from "node:fs";

const tokens = JSON.parse(readFileSync(new URL("../../design/tokens.json", import.meta.url), "utf8"));
const kebab = (name) => name.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`);
const colors = Object.keys(tokens.color.dark);
const aliases = {
  background: "bg", foreground: "text", card: "surface", "card-foreground": "text",
  popover: "surface", "popover-foreground": "text", primary: "text", "primary-foreground": "inverse",
  secondary: "raised", "secondary-foreground": "text", "muted-foreground": "muted",
  accent: "raised", "accent-foreground": "text", destructive: "loss", border: "line", input: "line", ring: "subtle",
  sidebar: "surface", "sidebar-foreground": "text", "sidebar-primary": "text", "sidebar-primary-foreground": "inverse",
  "sidebar-accent": "raised", "sidebar-accent-foreground": "text", "sidebar-border": "line", "sidebar-ring": "subtle",
  "chart-1": "text", "chart-2": "muted", "chart-3": "subtle", "chart-4": "line", "chart-5": "raised",
};
let css = `/* Generated from design/tokens.json by scripts/generate-theme.mjs. */
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn/tailwind.css";

@custom-variant dark (&:where(:root:not([data-theme="light"]), :root:not([data-theme="light"]) *));

`;
for (const [theme, palette] of Object.entries(tokens.color)) {
  css += `${theme === "dark" ? ":root" : '[data-theme="light"]'} {\n  color-scheme: ${theme};\n`;
  css += Object.entries(palette).map(([key, value]) => `  --${kebab(key)}: ${value};`).join("\n");
  // Aliases must be redeclared on theme scopes so nested light previews also resolve correctly.
  css += "\n" + Object.entries(aliases).map(([key, value]) => `  --${key}: var(--${value});`).join("\n");
  css += "\n}\n\n";
}
css += "@theme inline {\n";
css += [...new Set([...colors.map(kebab), ...Object.keys(aliases)])].map((key) => `  --color-${key}: var(--${key});`).join("\n");
css += `
  --font-sans: var(--font-inter), sans-serif;
  --font-heading: var(--font-inter), sans-serif;
  --font-mono: var(--font-jetbrains-mono), monospace;
  --spacing: ${tokens.space[0]}px;
  --spacing-control: ${tokens.size.control}px;
  --spacing-card: ${tokens.size.cardPadding}px;
  --spacing-page: ${tokens.size.pagePadding}px;
  --container-page: ${tokens.size.pageMaxWidth}px;
`;
for (const [name, value] of Object.entries(tokens.radius)) css += `  --radius-${name}: ${value}px;\n`;
css += `  --radius-sm: ${tokens.radius.tag}px;
  --radius-md: ${tokens.radius.control}px;
  --radius-lg: ${tokens.radius.control}px;
  --radius-xl: ${tokens.radius.card}px;
`;
for (const [name, type] of Object.entries(tokens.type)) {
  const key = kebab(name);
  css += `  --text-${key}: ${type.size}px;\n  --text-${key}--line-height: ${type.lineHeight}px;\n  --text-${key}--font-weight: ${type.weight};\n  --text-${key}--letter-spacing: ${type.letterSpacing}em;\n`;
}
css += `}

@utility num {
  font-family: var(--font-jetbrains-mono), ui-monospace, monospace;
  font-variant-numeric: tabular-nums;
}

@layer base {
  * { border-color: var(--line); }
  html { font-family: var(--font-inter), system-ui, sans-serif; }
  body { margin: 0; background: var(--bg); color: var(--text); font-size: ${tokens.type.body.size}px; line-height: ${tokens.type.body.lineHeight}px; }
  button, input, select, textarea { min-height: ${tokens.size.control}px; }
  button, [data-slot="button"], [data-slot="toggle"], [data-slot="toggle-group-item"] { min-width: ${tokens.size.control}px; }
  [data-slot="select-item"] { min-height: ${tokens.size.control}px; }
  button { cursor: pointer; }
  button:disabled { cursor: not-allowed; }
  input[inputmode="decimal"], input[inputmode="numeric"], input[type="number"] { font-family: var(--font-jetbrains-mono), monospace; font-variant-numeric: tabular-nums; }
  :focus-visible { outline: 2px solid var(--text); outline-offset: 3px; }
}

/* Visual meter geometry; all colors resolve through semantic theme tokens. */
.measure-ticks { height: 8px; margin-bottom: 6px; border-right: 1px solid var(--subtle); background: repeating-linear-gradient(90deg, var(--subtle) 0 1px, transparent 1px 10%); }
.measure-track { position: relative; height: 8px; border-radius: 2px; background: var(--raised); }
.measure-solid, .measure-added { position: absolute; top: 0; bottom: 0; border-radius: 2px; }
.measure-solid { left: 0; background: currentColor; }
.measure-added { background: repeating-linear-gradient(135deg, currentColor 0 2px, transparent 2px 5px); }
.measure-mark { position: absolute; top: -4px; bottom: -4px; width: 2px; }
.measure[data-size="sm"] .measure-ticks { height: 4px; margin-bottom: 3px; }
.measure[data-size="sm"] .measure-track { height: 6px; }
.measure[data-size="sm"] .measure-mark { top: -2px; bottom: -2px; }
.theme-dark-icon { display: none; }
[data-theme="light"] .theme-dark-icon { display: block; }
[data-theme="light"] .theme-light-icon { display: none; }
`;
writeFileSync(new URL("../src/app/globals.css", import.meta.url), css);
