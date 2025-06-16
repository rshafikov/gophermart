package schemas

type Withdraw struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
