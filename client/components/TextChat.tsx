import ChatDisplay from "@/components/ui/TextChat/ChatDisplay";
import { Message, Status } from "./Chat";

type TextChatProps = {
  messages: Message[];
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>;
  status: Status;
  connect: () => void;
  skip: () => void;
  dataChannel: RTCDataChannel | null;
};

export default function TextChat({
  messages,
  setMessages,
  status,
  connect: handleConnect,
  skip: handleSkip,
  dataChannel,
}: TextChatProps) {
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
      {/* Scrollable message area */}
      <div className="flex-1 overflow-y-auto pb-2 pl-2 min-h-0">
        <ChatDisplay messages={messages} />
      </div>

      <div className="flex items-center gap-2 pl-2 py-2 flex-shrink-0">
        {status === "idle" && (
          <button
            className="w-20 h-14 md:w-28 md:h-20 bg-blue-500 text-white rounded-lg text-sm md:text-base flex-shrink-0"
            onClick={handleConnect}
          >
            Start
          </button>
        )}

        {status === "waiting" && (
          <button
            disabled
            className="w-20 h-14 md:w-28 md:h-20 bg-blue-400 text-white rounded-lg text-sm md:text-base flex-shrink-0 cursor-not-allowed"
          >
            Searching…
          </button>
        )}

        {status === "chatting" && (
          <button
            className="w-20 h-14 md:w-28 md:h-20 bg-blue-500 text-white rounded-lg text-sm md:text-base flex-shrink-0"
            onClick={handleSkip}
          >
            Skip
          </button>
        )}

        <form onSubmit={handleSubmit} className="flex flex-1 gap-2 pr-2">
          <input
            id="chatInput"
            name="message"
            placeholder="Type a message…"
            className="flex-1 h-14 md:h-20 rounded-lg px-3 border border-gray-300 text-sm md:text-base focus:outline-none focus:ring-2 focus:ring-blue-400"
          />
          <button
            type="submit"
            className="w-16 h-14 md:w-28 md:h-20 bg-blue-500 text-white rounded-lg text-sm md:text-base flex-shrink-0"
          >
            Send
          </button>
        </form>
      </div>
    </div>
  );
}
