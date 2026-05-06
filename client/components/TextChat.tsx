import ChatDisplay from "@/ui/TextChat/ChatDisplay";

export default function TextChat() {
  return (
    <div className="text-black flex flex-col">
      <ChatDisplay />

      <div className="flex flex-row">
        <button className="w-20 h-20 bg-blue-500 text-white rounded-lg">
          Start
        </button>
        <input type="text" />
      </div>
    </div>
  );
}
