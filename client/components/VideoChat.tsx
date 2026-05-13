import VideoDisplay from "@/ui/VideoChat/VideoDisplay";

export default function VideoChat() {
  return (
    <div className="flex flex-col w-full h-full gap-2">
      <div className="flex-1">
        <VideoDisplay videoSource="localhost" />
      </div>
      <div className="flex-1">
        <VideoDisplay videoSource="localhost" />
      </div>
    </div>
  );
}
