import VideoDisplay from "@/components/ui/VideoChat/VideoDisplay";

export default function VideoChat({
  localStream,
  remoteStream: strangerStream,
}: {
  localStream: MediaStream | undefined;
  remoteStream: MediaStream | undefined;
}) {
  return (
    <div className="flex flex-row md:flex-col w-full h-full gap-2">
      <div className="flex-1 min-w-0 min-h-0" id="stranger">
        <VideoDisplay videoSource={strangerStream} />
      </div>
      <div className="flex-1 min-w-0 min-h-0" id="me">
        <VideoDisplay videoSource={localStream} />
      </div>
    </div>
  );
}
