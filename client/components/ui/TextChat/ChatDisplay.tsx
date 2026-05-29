"use client";

import { useEffect, useRef } from "react";
import Message from "./Message";

type ChatDisplayProp = {
  messages: {
    sender: string;
    message: string;
  }[];
};

export default function ChatDisplay({ messages }: ChatDisplayProp) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div className="flex flex-col gap-2 h-full bg-zinc-100 rounded-lg px-3 py-4 overflow-y-auto">
      {messages.length === 0 && (
        <p className="text-center text-gray-400 text-sm mt-auto self-center">
          Say hello!
        </p>
      )}
      {messages.map((message, i) => (
        <Message key={i} sender={message.sender} message={message.message} />
      ))}
      <div ref={bottomRef} />
    </div>
  );
}
