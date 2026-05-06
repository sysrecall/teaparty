import TextChat from "./TextChat";
import VideoChat from "./VideoChat";

export default function Chat() {
  return (
    <div className="flex flex-row w-full">
      <VideoChat />
      <TextChat />
    </div>
  );
}
