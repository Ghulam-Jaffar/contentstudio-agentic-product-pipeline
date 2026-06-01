package api

// PostThumbs holds the chosen post-level thumbnail and any carousel child thumbnails.
type PostThumbs struct {
	PostID       string       // normalized {pageID}_{postID}
	PostThumbURL string       // best image for the post (largest by area, fallback: full_picture)
	Children     []ChildThumb // one per carousel subattachment (if any)
}

type ChildThumb struct {
	MediaID   string // attachments.subattachments.data[*].target.id (photo/video object id)
	ThumbURL  string // media.image.src (ephemeral)
	MediaType string // photo|video|...
	Type      string // attachment type if you care
}
