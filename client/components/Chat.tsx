import TextChat from "./TextChat";
import VideoChat from "./VideoChat";

export default function Chat() {
  return (
    <div className="flex w-full h-full">
      <div className="w-1/3 h-full">
        <VideoChat />
      </div>

      <div className="w-2/3 h-full">
        <TextChat />
      </div>
    </div>
  );
}
