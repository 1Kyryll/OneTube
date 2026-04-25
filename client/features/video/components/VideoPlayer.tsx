"use client"; 

import { useEffect, useRef, useState } from "react";
import Hls from "hls.js"

interface Props {
    src: string; 
    poster?: string; 
    onFirstPlay?: () => void; 
}

export function VideoPlayer({ src, poster, onFirstPlay }: Props) {
    const ref = useRef<HTMLVideoElement | null>(null);
    const [levels, setLevels] = useState<{ height: number; index: number }[]>([]);
    const [currentLevel, setCurrentLevel] = useState<number>(-1);
    const hlsRef = useRef<Hls | null>(null);
    const firedFirstPlayRef = useRef(false);

    useEffect(() => {
        const video = ref.current;
        if (!video) return;

        // Prefer hls.js — it's the only path that exposes level switching.
        if (Hls.isSupported()) {
            const hls = new Hls({ enableWorker: true });
            hlsRef.current = hls;

            hls.on(Hls.Events.MANIFEST_PARSED, (_evt, data) => {
                console.log("MANIFEST_PARSED levels:", data.levels);
                setLevels(
                    data.levels
                        .map((l, i) => ({ height: l.height ?? 0, index: i }))
                        .sort((a, b) => a.height - b.height)
                );
            });

            hls.on(Hls.Events.ERROR, (_evt, data) => {
                console.error("hls error", data.type, data.details, data);
            });

            hls.attachMedia(video);
            hls.loadSource(src);

            return () => {
                hls.destroy();
                hlsRef.current = null;
            };
        }

        // Fallback: native HLS (iOS Safari). No level switching available.
        if (video.canPlayType("application/vnd.apple.mpegurl")) {
            video.src = src;
            return;
        }

        video.src = src;
    }, [src])

    function handlePlay() {
        console.log(levels);
        if (!firedFirstPlayRef.current) {
            firedFirstPlayRef.current = true; 
            onFirstPlay?.(); 
        }
    }

    function handleQualityChange(idx: number) {
        setCurrentLevel(idx); 
        if (hlsRef.current) hlsRef.current.currentLevel = idx; 
    }

    return (
        <div>
            <video ref={ref} controls poster={poster} style={{ width: "100%", maxHeight: "70vh" }} onPlay={handlePlay} />
            {levels.length > 1 && (
                <div style={{ marginTop: 8 }}>
                    <label>
                        Quality:{" "}
                        <select
                            value={currentLevel}
                            onChange={(e) => handleQualityChange(Number(e.target.value))}
                        >
                            <option value={-1}>Auto</option>
                            {levels.map((l) => (
                                <option key={l.index} value={l.index}>
                                    {l.height}p
                                </option>
                            ))}
                        </select>
                    </label>
                </div>
            )}
        </div>
    ); 
}