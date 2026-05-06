export default function VideoDisplay({ videoSource }: { videoSource: string }) {
  return (
    <>
      <video
        src={videoSource}
        className="w-full h-full bg-amber-200 rounded-lg"
      ></video>
    </>
  );
}
