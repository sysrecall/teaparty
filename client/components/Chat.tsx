"use client";

import { useRef, useState } from "react";
import TextChat from "./TextChat";
import VideoChat from "./VideoChat";

const MIN_WIDTH = 800;
const MIN_HEIGHT = 600;

const CONSTRAINTS = {
  // audio: { echoCancellation: true },
  video: {
    width: { ideal: MIN_WIDTH },
    height: { ideal: MIN_HEIGHT },
  },
};

// fallback stun only
const ICE_SERVERS = [{ urls: "stun:stun.l.google.com:19302" }];

async function openCamera(constraints: MediaStreamConstraints | undefined) {
  return await navigator.mediaDevices.getUserMedia(constraints);
}

export type Status = "idle" | "waiting" | "chatting";

type TurnServer =
  | {
      urls: string;
    }
  | {
      urls: string;
      username: string;
      credentials: string;
    };

export type Message = {
  sender: string;
  message: string;
};

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws";

export default function Chat() {
  const socketRef = useRef<WebSocket>(null);
  const iceServers = useRef<TurnServer[]>(ICE_SERVERS);
  const [localStream, setLocalStream] = useState<MediaStream>();
  const [remoteStream, setRemoteStream] = useState<MediaStream>();
  const peerConnectionRef = useRef<RTCPeerConnection>(null);
  const [status, setStatus] = useState<Status>("idle");
  const [dataChannel, setDataChannel] = useState<RTCDataChannel | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);

  function findPeer(
    socket: WebSocket,
    pc: RTCPeerConnection,
    stream: MediaStream,
  ) {
    console.log("finding peers");

    const pendingCandidates: RTCIceCandidate[] = [];

    pc.ondatachannel = (event) => {
      setDataChannel(event.channel);
    };

    stream.getTracks().forEach((track) => pc.addTrack(track, stream));

    pc.ontrack = (event) => {
      const [rs] = event.streams;
      setRemoteStream(rs);
      setStatus("chatting");
    };

    socket.onopen = () => {
      console.log("Connected to the server");
    };

    socket.onmessage = async (event) => {
      console.log("Message from the server:", event.data);

      const pc = peerConnectionRef.current;

      if (!pc) {
        return;
      }

      const data = JSON.parse(event.data);

      switch (data.type) {
        case "match":
          pc.onicecandidate = (event) => {
            if (event.candidate) {
              const message = JSON.stringify({
                type: "candidate",
                message: event.candidate,
              });

              socket.send(message);
            }
          };

          // only send offer if this is the offerer
          if (data.message === "offerer") {
            setDataChannel(pc.createDataChannel("text-chat"));

            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            const message = JSON.stringify({
              type: "offer",
              message: offer,
            });

            socket.send(message);
          }

          break;

        case "servers":
          iceServers.current = JSON.parse(data.message);

          pc.setConfiguration({
            iceServers: iceServers.current,
          });

          break;

        case "offer":
          await pc.setRemoteDescription(
            new RTCSessionDescription(data.message),
          );
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);

          const message = JSON.stringify({
            type: "answer",
            message: answer,
          });

          socket.send(message);

          break;

        case "answer":
          await pc.setRemoteDescription(
            new RTCSessionDescription(data.message),
          );

          pendingCandidates.forEach(async (candidate) => {
            await pc.addIceCandidate(candidate);
          });

          pendingCandidates.length = 0;

          break;

        case "skip":
          peerConnectionRef.current?.close();

          if (socketRef.current) {
            socketRef.current.onclose = null;
            socketRef.current.close();
          }

          const newSocket = new WebSocket(WS_URL);
          socketRef.current = newSocket;

          const newPc = new RTCPeerConnection({
            iceServers: iceServers.current,
          });

          peerConnectionRef.current = newPc;
          setRemoteStream(undefined);
          setDataChannel(null);
          setStatus("waiting");
          setMessages([]);
          findPeer(newSocket, newPc, stream);

          break;

        case "candidate":
          const candidate = new RTCIceCandidate(data.message);
          if (pc.remoteDescription) {
            await pc.addIceCandidate(candidate);
          } else {
            pendingCandidates.push(candidate);
          }
          break;

        default:
          console.error("Invalid message type arrived from the server");
      }
    };

    socket.onerror = (error) => {
      console.error("Websocket Error:", error);
    };

    socket.onclose = () => {
      console.log("Disconnected from the server");
      setStatus("idle");
    };
  }

  async function connect() {
    console.log("Connecting");

    let stream: MediaStream;
    try {
      stream = await openCamera(CONSTRAINTS);
    } catch (err) {
      console.error("Camera error:", err);
      return;
    }

    setLocalStream(stream);

    console.log("local stream set");

    setStatus("waiting");

    const socket = new WebSocket(WS_URL);
    socketRef.current = socket;

    const pc = new RTCPeerConnection({ iceServers: iceServers.current });
    peerConnectionRef.current = pc;

    findPeer(socket, pc, stream);
  }

  async function skip() {
    peerConnectionRef.current?.close();
    socketRef.current?.close();

    const socket = new WebSocket(WS_URL);
    socketRef.current = socket;

    const pc = new RTCPeerConnection({ iceServers: iceServers.current });
    peerConnectionRef.current = pc;

    setRemoteStream(undefined);
    setDataChannel(null);
    setStatus("waiting");
    setMessages([]);

    findPeer(socket, pc, localStream!);
  }

  return (
    <div className="flex w-full h-full">
      <div className="w-1/3 h-full">
        <VideoChat localStream={localStream} remoteStream={remoteStream} />
      </div>

      <div className="w-2/3 h-full">
        <TextChat
          messages={messages}
          setMessages={setMessages}
          skip={skip}
          status={status}
          connect={connect}
          dataChannel={dataChannel}
        />
      </div>
    </div>
  );
}
