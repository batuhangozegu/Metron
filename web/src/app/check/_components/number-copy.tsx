/** Preserve mono numerals inside translated prose without changing word typography. */
export function NumberCopy({ text }: { text: string }) {
  return text.split(/([−+]?\d[\d,.:%]*)/g).map((part, i) => /\d/.test(part) ? <span key={i} className="num">{part}</span> : part);
}
