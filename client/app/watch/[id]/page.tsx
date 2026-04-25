"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { VideoPlayer } from "../../../features/video/components/VideoPlayer";
import { getWatch, recordView } from "../../../features/video/actions";
import type { VideoDTO } from "../../../features/video/types";

type WatchState =
  | { kind: "loading" }
  | { kind: "processing" }
  | { kind: "ready"; video: VideoDTO; masterURL: string }
  | { kind: "error"; message: string };

export default function WatchPage() {
    const params = useParams<{ id: string }>();
    const id = params.id;
    const [state, setState] = useState<WatchState>({ kind: "loading" });

    useEffect(() => {
        let cancelled = false;
        let timer: ReturnType<typeof setTimeout> | undefined;

        async function poll() {
            try {
                const res = await getWatch(id);
                if (cancelled) return;
                if ("master_playlist_url" in res) {
                    setState({ kind: "ready", video: res.video, masterURL: res.master_playlist_url });
                } else {
                    setState({ kind: "processing" });
                    timer = setTimeout(poll, 3000);
                }
            } catch (err) {
                if (cancelled) return;
                setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
            }
        }
        poll();
        return () => {
            cancelled = true;
            if (timer) clearTimeout(timer);
        };
    }, [id]);

    if (state.kind === "loading") 
        return <main style={{ padding: 24 }}>Loading...</main>;
    if (state.kind === "processing")
        return <main style={{ padding: 24 }}>Video is still transcoding. Retrying every 3s.</main>;
    if (state.kind === "error")
        return <main style={{ padding: 24, color: "crimson" }}>Error: {state.message}</main>;

    const { video, masterURL } = state;
    return (
        <main style={{ padding: 24 }}>
            <h1>{video.title}</h1>
            <p>
                by <strong>@{video.uploader_username}</strong> · {video.view_count} views
            </p>
            <VideoPlayer
                src={masterURL}
                poster={video.thumbnail_url}
                onFirstPlay={() => {
                void recordView(video.id);
                }}
            />
            {video.description && <p style={{ marginTop: 16, whiteSpace: "pre-wrap" }}>{video.description}</p>}
        </main>
    );
}