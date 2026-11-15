/** Keep numbers in translated prose in the same mono style as financial figures. */
export function NumberText({ text }: { text: string }) {
  return text.split(/(\d[\d,.:]*\d%?|\d%?)/g).map((part, index) =>
    /^\d/.test(part) ? <span key={index} className="num">{part}</span> : part,
  );
}
