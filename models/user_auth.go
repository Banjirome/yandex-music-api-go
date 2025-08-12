package models

// Расширенные модели для многошаговой авторизации пользователя.

type AuthTypes struct {
	TrackId   string   `json:"track_id"`
	AuthTypes []string `json:"preferred_auth_methods"`
}

type AuthQR struct {
	Status    string `json:"status"`
	TrackId   string `json:"track_id"`
	CsrfToken string `json:"csrf_token"`
}

type AuthQRStatus struct {
	Status             string `json:"status"`
	Code               string `json:"code"`
	TrackId            string `json:"track_id"`
	MagicLinkConfirmed bool   `json:"magic_link_confirmed"`
}

type AuthCaptcha struct {
	Status   string `json:"status"`
	TrackId  string `json:"track_id"`
	ImageUrl string `json:"image_url"`
}

type AuthLetter struct {
	Status  string `json:"status"`
	TrackId string `json:"track_id"`
}

type AuthLetterStatus struct {
	Status             string `json:"status"`
	TrackId            string `json:"track_id"`
	MagicLinkConfirmed bool   `json:"magic_link_confirmed"`
}

type AuthBase struct {
	Status  string `json:"status"`
	TrackId string `json:"track_id"`
}

type LoginInfo struct {
	Id          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}
