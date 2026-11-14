import { getMessages } from "@/lib/messages";

export default function Page() {
  return <h1 className="text-title">{getMessages("common")("nav.check")}</h1>;
}
