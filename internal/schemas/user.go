package schemas

type UserCreate struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserCreateResponse struct {
	Login string `json:"login"`
}
