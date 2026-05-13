export default function VideoDisplay({ videoSource }: { videoSource: string }) {
  return (
    <>
      <video
        src={videoSource}
        className="w-full h-full object-cover bg-zinc-200 rounded-lg"
      ></video>
    </>
  );
}
