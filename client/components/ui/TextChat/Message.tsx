export default function Message({
  sender,
  message,
}: {
  sender: string;
  message: string;
}) {
  const isMe = sender === "me";

  return (
    <div
      className={`flex flex-col gap-1 ${isMe ? "items-end" : "items-start"}`}
    >
      <span className="text-xs text-gray-400 px-1">
        {isMe ? "You" : "Stranger"}
      </span>
      <div
        className={`max-w-[75%] px-4 py-2 rounded-2xl text-sm leading-relaxed break-words ${
          isMe
            ? "bg-blue-500 text-white rounded-br-sm"
            : "bg-white text-gray-800 border border-gray-200 rounded-bl-sm shadow-sm"
        }`}
      >
        {message}
      </div>
    </div>
  );
}
