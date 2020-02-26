package middleware

type Access int

const (
	Admin Access = iota + 1
	Authenticated
	Anonymous
)
