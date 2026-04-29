"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { cursorToString, getFeed } from "../actions";
import type { FeedPage, VideoDTO } from "../types";
import Image from "next/image";

export function FeedGrid() {
    const [items, setItems] = useState<VideoDTO[]>([]);
    const [cursor, setCursor] = useState<string | undefined>(undefined);
    const [loading, setLoading] = useState(false);
    const [done, setDone] = useState(false);

    async function loadMore(reset = false) {
        if (loading || done) return;
        setLoading(true);
        try {
            const page: FeedPage = await getFeed(reset ? undefined : cursor);
            setItems((prev) => (reset ? page.items : [...prev, ...page.items]));
            const next = await cursorToString(page.next_cursor);
            setCursor(next);
            if (!next) setDone(true);
        } catch {
            setDone(true);
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        void loadMore(true);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const isEmpty = !loading && items.length === 0;

    return (
        <div>
            {isEmpty ? (
                <div
                    style={{
                        padding: "64px 16px",
                        textAlign: "center",
                        color: "#888",
                        fontSize: 16,
                    }}
                >
                    No videos to show
                </div>
            ) : (
                <div
                    style={{
                        display: "grid",
                        gridTemplateColumns: "repeat(auto-fill, minmax(240px, 1fr))",
                        gap: 16,
                    }}
                >
                    {items.map((v) => (
                        <Link
                            key={v.id}
                            href={`/watch/${v.id}`}
                            style={{ textDecoration: "none", color: "inherit" }}
                        >
                            <div>
                                <div
                                    style={{
                                        position: "relative",
                                        width: "100%",
                                        aspectRatio: "16/9",
                                        background: "#1a1a1a",
                                        borderRadius: 8,
                                        overflow: "hidden",
                                    }}
                                >
                                    {v.thumbnail_url && (
                                        <Image
                                            src={v.thumbnail_url}
                                            alt={v.title}
                                            fill
                                            unoptimized
                                            sizes="(max-width: 600px) 100vw, 240px"
                                            style={{ objectFit: "cover" }}
                                        />
                                    )}
                                </div>
                                <div style={{ marginTop: 8, fontWeight: 600, color: "#fff" }}>
                                    {v.title}
                                </div>
                                <div
                                    style={{
                                        display: "flex",
                                        alignItems: "center",
                                        gap: 6,
                                        fontSize: 13,
                                        color: "#9aa3b2",
                                    }}
                                >
                                    {v.uploader_avatar_url ? (
                                        <Image
                                            src={v.uploader_avatar_url}
                                            alt=""
                                            width={20}
                                            height={20}
                                            unoptimized
                                            style={{
                                                borderRadius: "50%",
                                                objectFit: "cover",
                                                background: "#222",
                                            }}
                                        />
                                    ) : null}
                                    <span>
                                        @{v.uploader_username} · {v.view_count} views
                                    </span>
                                </div>
                            </div>
                        </Link>
                    ))}
                </div>
            )}
            {!done && !isEmpty && (
                <button
                    onClick={() => void loadMore(false)}
                    disabled={loading}
                    style={{
                        marginTop: 16,
                        padding: "8px 16px",
                        background: "#1f1f1f",
                        color: "#fff",
                        border: "1px solid #333",
                        borderRadius: 4,
                        cursor: "pointer",
                    }}
                >
                    {loading ? "Loading..." : "Load more"}
                </button>
            )}
        </div>
    );
}