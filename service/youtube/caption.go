package youtube

type Caption struct {
	Events []CaptionEvent `json:"events"`
}

type CaptionEvent struct {
	TStartMs int              `json:"tStartMs"`
	Segments []CaptionSegment `json:"segs,omitempty"`
}

type CaptionSegment struct {
	UTF8      string `json:"utf8"`
	TOffsetMs int    `json:"tOffsetMs"`
	AcAsrConf int    `json:"acAsrConf"`
}

// -----

type PlayerRequest struct {
	Context PlayerRequestContext `json:"context"`
	VideoID string               `json:"videoId"`
}

type PlayerRequestContext struct {
	Client PlayerRequestContextClient `json:"client"`
}

type PlayerRequestContextClient struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

type PlayerResponse struct {
	Captions PlayerResponseCaptions `json:"captions"`
}

type PlayerResponseCaptions struct {
	PlayerCaptionsTracklistRenderer PlayerResponseCaptionsTracks `json:"playerCaptionsTracklistRenderer"`
}

type PlayerResponseCaptionsTracks struct {
	CaptionTracks []CaptionTrack `json:"captionTracks"`
}

type CaptionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Name         struct {
		SimpleText string `json:"simpleText"`
	} `json:"name"`
	Kind string `json:"kind"`
}
