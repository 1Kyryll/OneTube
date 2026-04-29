export type VideoStatus = "uploading" | "processing" | "ready" | "failed"; 

export interface VideoDTO {
    id: string;
    uploader_id: string;
    uploader_username: string;
    uploader_avatar_url?: string;
    title: string;
    description: string;
    duration_seconds?: number;
    visibility: "public" | "unlisted" | "private";
    status: VideoStatus;
    view_count: number;
    thumbnail_url?: string;
    created_at: string;
    published_at?: string;
}

export interface CreateVideoResponse {
    video_id: string;
    upload_url: string;
    storage_key: string;
    expires_in_seconds: number;
}

export interface WatchResponse {
    video: VideoDTO;
    master_playlist_url: string;
}

export interface FeedPage {
    items: VideoDTO[];
    next_cursor?: { published_at: string; id: string };
}