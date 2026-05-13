import ChatDisplay from "@/ui/TextChat/ChatDisplay";

export default function TextChat() {
  return (
    <div className="flex flex-col w-full h-full text-black">
      <div className="flex-1 overflow-y-auto pb-2 pl-2">
        <ChatDisplay />
      </div>

      <div className="flex items-center gap-2 pl-2">
        <button className="w-28 h-20 bg-blue-500 text-white rounded-lg">
          Start
        </button>
        <input className="flex-1 h-20 rounded px-3 border border-gray-300" />
      </div>
    </div>
  );
}
