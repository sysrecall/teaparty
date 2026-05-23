import Message from "./Message";

type ChatDisplayProp = {
  messages: {
    sender: string;
    message: string;
  }[];
};

export default function ChatDisplay({ messages }: ChatDisplayProp) {
  return (
    <div className="pt-4 pl-4 h-full flex flex-col bg-zinc-100 rounded-lg start gap-4">
      {messages.map((message, i) => (
        <Message key={i} sender={message.sender} message={message.message} />
      ))}
    </div>
  );
}
