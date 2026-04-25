"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { completeUpload, createUploadIntent } from "../actions";

// Browser-only helper — lives in the client component because the server
// action module ("use server") cannot export non-async values or functions
// that touch XMLHttpRequest.
function putToPresignedURL(
  url: string,
  file: File,
  onProgress: (pct: number) => void,
): Promise<void> {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open("PUT", url, true);
        xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");
        xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) onProgress((e.loaded / e.total) * 100);
        };
        xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) resolve();
        else reject(new Error(`Upload failed: ${xhr.status}`));
        };
        xhr.onerror = () => reject(new Error("Network error during upload"));
        xhr.send(file);
    });
}

export function UploadForm() {
    const router = useRouter();
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [visibility, setVisibility] = useState<"public" | "unlisted" | "private">("public");
    const [file, setFile] = useState<File | null>(null);
    const [progress, setProgress] = useState(0);
    const [error, setError] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);

    async function onSubmit(e: React.FormEvent) {
        e.preventDefault();
        if (!file) {
        setError("Choose a file first");
        return;
        }
        setError(null);
        setBusy(true);
        setProgress(0);
        try {
        const intent = await createUploadIntent({
            title: title.trim(),
            description: description.trim(),
            visibility,
            filename: file.name,
            contentType: file.type || "video/mp4",
        });
        await putToPresignedURL(intent.upload_url, file, setProgress);
        const ext = (file.name.split(".").pop() || "mp4").toLowerCase();
        await completeUpload(intent.video_id, ext);
        router.push(`/watch/${intent.video_id}`);
        } catch (err) {
        setError(err instanceof Error ? err.message : String(err));
        setBusy(false);
        }
    }

    return (
        <form onSubmit={onSubmit} style={{ maxWidth: 600 }}>
        <label>Title<br /><input value={title} onChange={(e) => setTitle(e.target.value)} required /></label>
        <br />
        <label>Description<br /><textarea value={description} onChange={(e) => setDescription(e.target.value)} /></label>
        <br />
        <label>Visibility{" "}
            <select value={visibility} onChange={(e) => setVisibility(e.target.value as "public" | "unlisted" | "private")}>
            <option value="public">Public</option>
            <option value="unlisted">Unlisted</option>
            <option value="private">Private</option>
            </select>
        </label>
        <br />
        <input
            type="file"
            accept="video/*"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
        />
        <br />
        <button disabled={busy} type="submit">{busy ? "Uploading..." : "Upload"}</button>
        {busy && <div>Progress: {progress.toFixed(1)}%</div>}
        {error && <div style={{ color: "crimson" }}>{error}</div>}
        </form>
    );
}