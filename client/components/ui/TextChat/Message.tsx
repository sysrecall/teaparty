export default function Message({
  sender,
  message,
}: {
  sender: string;
  message: string;
}) {
  return (
    <div>
      <span>{sender}</span>: {message}
    </div>
  );
}
