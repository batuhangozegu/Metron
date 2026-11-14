import { getMessages } from "@/lib/messages";

export default function Page() {
  return <h1 className="text-title">{getMessages("common")("nav.portfolio")}</h1>;
}
