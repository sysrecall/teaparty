import VideoDisplay from "@/ui/VideoChat/VideoDisplay";

export default function VideoChat({
  localStream,
  strangerStream,
}: {
  localStream: MediaStream | undefined;
  strangerStream: MediaStream | undefined;
}) {
  return (
    <div className="flex flex-col w-full h-full gap-2">
      <div className="flex-1" id="stranger">
        <VideoDisplay videoSource={strangerStream} />
      </div>
      <div className="flex-1" id="me">
        <VideoDisplay videoSource={localStream} />
      </div>
    </div>
  );
}
