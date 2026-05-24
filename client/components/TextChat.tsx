import ChatDisplay from "@/components/ui/TextChat/ChatDisplay";
import { useEffect, useState } from "react";
import { Status } from "./Chat";

type Message = {
  sender: string;
  message: string;
};

export default function TextChat({
  status,
  connect: handleConnect,
  skip: handleSkip,
  dataChannel,
}: {
  status: Status;
  connect: () => void;
  skip: () => void;
  dataChannel: RTCDataChannel | null;
}) {
  const [messages, setMessages] = useState<Message[]>([]);

  if (dataChannel) {
    dataChannel.onmessage = (event) => {
      setMessages((val) => [
        ...val,
        { sender: "stranger", message: event.data },
      ]);
    };
  }

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!dataChannel) return;

    const form = e.currentTarget;
    const formData = new FormData(form);

    const message = formData.get("message")?.toString().trim();

    if (!message) return;

    dataChannel.send(message);
    setMessages((val) => [...val, { sender: "me", message }]);

    form.reset();
  };

  return (
    <div className="flex flex-col w-full h-full text-black">
      <div className="flex-1 overflow-y-auto pb-2 pl-2">
        <ChatDisplay messages={messages} />
      </div>

      <div className="flex items-center gap-2 pl-2">
        {status == "idle" && (
          <button
            className="w-28 h-20 bg-blue-500 text-white rounded-lg"
            onClick={handleConnect}
          >
            Start
          </button>
        )}

        {status == "waiting" && (
          <button
            disabled={true}
            className="w-28 h-20 bg-blue-500 text-white rounded-lg"
          >
            Searching
          </button>
        )}

        {status == "chatting" && (
          <button
            className="w-28 h-20 bg-blue-500 text-white rounded-lg"
            onClick={handleSkip}
          >
            Skip
          </button>
        )}

        <form onSubmit={handleSubmit}>
          <button
            type="submit"
            className="w-28 h-20 bg-blue-500 text-white rounded-lg"
          >
            Send
          </button>
          <input
            id="chatInput"
            name="message"
            className="flex-1 h-20 rounded px-3 border border-gray-300"
          />
        </form>
      </div>
    </div>
  );
}
