/** Keep embedded quantities and times monospaced without changing prose typography. */
export function NumericCopy({ text }: { text: string }) {
  return text.split(/(\d+(?:[.,:]\d+)*(?:[%×])?)/g).map((part, index) =>
    /\d/.test(part) ? <span key={index} className="num">{part}</span> : part,
  );
}
