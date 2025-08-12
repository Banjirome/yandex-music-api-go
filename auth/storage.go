package auth

import "net/url"

// Storage хранит данные авторизации и сетевые настройки.
type Storage struct {
	Token        string
	DeviceID     string
	Proxy        *url.URL
	User         *User
	IsAuthorized bool
	AuthToken    *AuthToken
	AccessToken  *AccessToken
}

// New создаёт новое хранилище с опциональным токеном.
func New(token string) *Storage {
	return &Storage{Token: token, DeviceID: "go-sdk", User: &User{}}
}

// SetProxy настраивает прокси.
func (s *Storage) SetProxy(u *url.URL) { s.Proxy = u }

// User представляет текущего авторизованного пользователя.
type User struct {
	Uid   string
	Login string
}

// AuthToken хранит csrf и track для многошаговой авторизации.
type AuthToken struct {
	CsrfToken string
	TrackId   string
}

// AccessToken хранит полученный OAuth music token.
type AccessToken struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
}

// SetUid устанавливает идентификатор пользователя.
func (s *Storage) SetUid(uid string) {
	if s.User == nil {
		s.User = &User{}
	}
	s.User.Uid = uid
}

// SetLogin устанавливает логин пользователя.
func (s *Storage) SetLogin(login string) {
	if s.User == nil {
		s.User = &User{}
	}
	s.User.Login = login
}
