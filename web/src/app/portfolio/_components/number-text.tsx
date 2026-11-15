/** Keep numerals in translated prose in the shared mono style. */
export function NumberText({ text }: { text: string }) {
  return <>{text.split(/(\d[\d.,:%]*R?)/g).map((part, index) => /\d/.test(part) ? <span key={index} className="num">{part}</span> : part)}</>;
}
