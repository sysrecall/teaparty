import VideoDisplay from "@/ui/VideoChat/VideoDisplay";

export default function VideoChat() {
  return (
    <div className="flex flex-col">
      <div id="me" className="">
        <VideoDisplay videoSource="localhost" />
      </div>
      <div id="stranger">
        <VideoDisplay videoSource="localhost" />
      </div>
    </div>
  );
}
